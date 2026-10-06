package usecase

import (
	"context"
	"fmt"
	"hash/fnv"
	"sync"

	usecasein "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/in"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/out"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/reservation"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

// reserveSeatUseCase implementa in.ReserveSeatUseCase.
// Nótese que solo depende de INTERFACES (out.SeatRepository,
// out.ReservationRepository, out.ReservationCache, out.EventPublisher,
// out.RealtimeNotifier), nunca de postgres/redis/rabbitmq concretos.
// Eso es lo que hace este código 100% testeable con mocks/fakes.
type reserveSeatUseCase struct {
	seatRepo         out.SeatRepository
	reservationRepo  out.ReservationRepository
	reservationCache out.ReservationCache
	publisher        out.EventPublisher
	notifier         out.RealtimeNotifier
	userLocks        [userLockStripes]sync.Mutex
}

// userLockStripes: cantidad fija de candados. Un usuario siempre cae en
// el mismo, así que sus reservas se procesan de una en una (los 5
// workers de la sala de espera podrían tomar 4 peticiones suyas a la
// vez y saltarse el límite). Usuarios distintos casi nunca comparten
// candado y, si lo hacen, solo esperan unos milisegundos.
const userLockStripes = 64

func (uc *reserveSeatUseCase) lockUser(userID shared.ID) func() {
	h := fnv.New32a()
	_, _ = h.Write([]byte(userID.String()))
	m := &uc.userLocks[h.Sum32()%userLockStripes]
	m.Lock()
	return m.Unlock
}

func NewReserveSeatUseCase(
	seatRepo out.SeatRepository,
	reservationRepo out.ReservationRepository,
	reservationCache out.ReservationCache,
	publisher out.EventPublisher,
	notifier out.RealtimeNotifier,
) usecasein.ReserveSeatUseCase {
	return &reserveSeatUseCase{
		seatRepo:         seatRepo,
		reservationRepo:  reservationRepo,
		reservationCache: reservationCache,
		publisher:        publisher,
		notifier:         notifier,
	}
}

func (uc *reserveSeatUseCase) Execute(ctx context.Context, cmd usecasein.ReserveSeatCommand) (*usecasein.ReserveSeatResult, error) {
	// 0. Límite por usuario: máximo reservation.MaxPendingPerUser
	//    asientos apartados a la vez. El candado por usuario evita que
	//    varias peticiones simultáneas suyas pasen la comprobación juntas.
	//    Limitación conocida: protege dentro de UNA instancia de la API;
	//    con varias réplicas haría falta un conteo atómico en la BD.
	defer uc.lockUser(cmd.UserID)()

	pending, err := uc.reservationRepo.CountPendingByUser(ctx, cmd.EventID, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("error contando reservas del usuario: %w", err)
	}
	if pending >= reservation.MaxPendingPerUser {
		return nil, shared.NewDomainError(
			"MAX_SEATS_PER_USER",
			fmt.Sprintf("solo puedes apartar hasta %d asientos a la vez", reservation.MaxPendingPerUser),
		)
	}

	// 1. Intento ATÓMICO de reserva contra Postgres. Este es el punto
	//    donde se decide, bajo concurrencia real, quién gana la carrera.
	//    La implementación concreta (adapters/out/postgres) debe usar
	//    un UPDATE condicional o SELECT FOR UPDATE dentro de una
	//    transacción. Si devuelve false, alguien más llegó antes.
	reserved, err := uc.seatRepo.TryReserve(ctx, cmd.SeatID)
	if err != nil {
		return nil, fmt.Errorf("error intentando reservar asiento: %w", err)
	}
	if !reserved {
		return nil, shared.NewDomainError(
			"SEAT_ALREADY_RESERVED",
			"el asiento ya no está disponible, otro usuario lo reservó primero",
		)
	}

	// 2. Se creó la reserva de dominio con su TTL de negocio (10 min).
	res := reservation.NewReservation(cmd.EventID, cmd.SeatID, cmd.UserID)

	if err := uc.reservationRepo.Save(ctx, res); err != nil {
		// Compensación: si no pudimos persistir la reserva, liberamos
		// el asiento que acabábamos de bloquear para no dejarlo huérfano.
		_ = uc.seatRepo.Release(ctx, cmd.SeatID)
		return nil, fmt.Errorf("error guardando la reserva: %w", err)
	}

	// 3. Se materializa el TTL en Redis. Cuando esta clave expire,
	//    un adaptador de entrada (redis-subscriber) disparará
	//    ReleaseExpiredReservationUseCase.
	if err := uc.reservationCache.SetWithTTL(ctx, res.ID(), cmd.SeatID, reservation.DefaultTTL); err != nil {
		_ = uc.seatRepo.Release(ctx, cmd.SeatID)
		_ = uc.reservationRepo.UpdateStatus(ctx, res.ID(), reservation.StatusCancelled)
		return nil, fmt.Errorf("error registrando el TTL en cache: %w", err)
	}

	// 4. Notificaciones best-effort: si fallan, no revertimos la
	//    reserva (ya es válida), solo lo logueamos en la capa superior.
	_ = uc.notifier.BroadcastSeatUpdate(ctx, cmd.EventID, cmd.SeatID, "RESERVED")
	_ = uc.publisher.Publish(ctx, "seat.reserved", []byte(res.ID().String()))

	return &usecasein.ReserveSeatResult{
		ReservationID: res.ID(),
		ExpiresInSec:  int(reservation.DefaultTTL.Seconds()),
	}, nil
}
