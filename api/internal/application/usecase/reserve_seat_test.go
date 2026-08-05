package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/usecase"
	usecasein "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/in"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/reservation"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/seat"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

// --- Fakes: implementaciones en memoria de los puertos de salida ---
// Esta es la ventaja concreta de la arquitectura hexagonal: probar el
// caso de uso crítico del sistema (reservar un asiento) sin levantar
// Postgres, Redis ni RabbitMQ.

type fakeSeatRepo struct {
	reserveSucceeds bool
	released        bool
}

func (f *fakeSeatRepo) FindByID(ctx context.Context, id shared.ID) (*seat.Seat, error) {
	return nil, errors.New("no implementado en el fake")
}
func (f *fakeSeatRepo) ListByEvent(ctx context.Context, eventID shared.ID) ([]*seat.Seat, error) {
	return nil, nil
}
func (f *fakeSeatRepo) TryReserve(ctx context.Context, seatID shared.ID) (bool, error) {
	return f.reserveSucceeds, nil
}
func (f *fakeSeatRepo) Release(ctx context.Context, seatID shared.ID) error {
	f.released = true
	return nil
}
func (f *fakeSeatRepo) MarkSold(ctx context.Context, seatID shared.ID) error { return nil }

type fakeReservationRepo struct {
	saveErr error
	saved   *reservation.Reservation
}

func (f *fakeReservationRepo) Save(ctx context.Context, r *reservation.Reservation) error {
	f.saved = r
	return f.saveErr
}
func (f *fakeReservationRepo) FindByID(ctx context.Context, id shared.ID) (*reservation.Reservation, error) {
	return f.saved, nil
}
func (f *fakeReservationRepo) UpdateStatus(ctx context.Context, id shared.ID, status reservation.Status) error {
	return nil
}
func (f *fakeReservationRepo) FindExpiredPending(ctx context.Context, before time.Time) ([]shared.ID, error) {
	return nil, nil
}

type fakeCache struct {
	setCalled bool
}

func (f *fakeCache) SetWithTTL(ctx context.Context, reservationID, seatID shared.ID, ttl time.Duration) error {
	f.setCalled = true
	return nil
}
func (f *fakeCache) Delete(ctx context.Context, reservationID shared.ID) error { return nil }
func (f *fakeCache) SubscribeExpirations(ctx context.Context) (<-chan shared.ID, error) {
	ch := make(chan shared.ID)
	close(ch)
	return ch, nil
}

type fakePublisher struct{}

func (f *fakePublisher) Publish(ctx context.Context, topic string, payload []byte) error { return nil }

type fakeNotifier struct{}

func (f *fakeNotifier) BroadcastSeatUpdate(ctx context.Context, eventID, seatID shared.ID, status string) error {
	return nil
}

// --- Tests ---

func TestReserveSeatUseCase_Execute_SeatAvailable_Succeeds(t *testing.T) {
	seatRepo := &fakeSeatRepo{reserveSucceeds: true}
	resRepo := &fakeReservationRepo{}
	cache := &fakeCache{}

	uc := usecase.NewReserveSeatUseCase(seatRepo, resRepo, cache, &fakePublisher{}, &fakeNotifier{})

	result, err := uc.Execute(context.Background(), usecasein.ReserveSeatCommand{
		EventID: shared.NewID(),
		SeatID:  shared.NewID(),
		UserID:  shared.NewID(),
	})

	if err != nil {
		t.Fatalf("se esperaba éxito, se obtuvo error: %v", err)
	}
	if result.ReservationID.IsZero() {
		t.Fatal("se esperaba un ReservationID válido")
	}
	if result.ExpiresInSec != int(reservation.DefaultTTL.Seconds()) {
		t.Fatalf("se esperaba ExpiresInSec=%d, se obtuvo %d", int(reservation.DefaultTTL.Seconds()), result.ExpiresInSec)
	}
	if !cache.setCalled {
		t.Fatal("se esperaba que se registrara el TTL en el cache")
	}
	if resRepo.saved == nil {
		t.Fatal("se esperaba que la reserva se guardara en el repositorio")
	}
}

// Este es el test que demuestra, en código, la garantía anti-doble-venta:
// si TryReserve (la operación atómica de Postgres) dice que no se pudo
// reservar, el caso de uso DEBE devolver SEAT_ALREADY_RESERVED y no
// debe llegar a guardar ninguna reserva.
func TestReserveSeatUseCase_Execute_SeatAlreadyReserved_ReturnsDomainError(t *testing.T) {
	seatRepo := &fakeSeatRepo{reserveSucceeds: false}
	resRepo := &fakeReservationRepo{}
	cache := &fakeCache{}

	uc := usecase.NewReserveSeatUseCase(seatRepo, resRepo, cache, &fakePublisher{}, &fakeNotifier{})

	_, err := uc.Execute(context.Background(), usecasein.ReserveSeatCommand{
		EventID: shared.NewID(),
		SeatID:  shared.NewID(),
		UserID:  shared.NewID(),
	})

	if err == nil {
		t.Fatal("se esperaba error SEAT_ALREADY_RESERVED, se obtuvo nil")
	}

	var domainErr *shared.DomainError
	if !errors.As(err, &domainErr) {
		t.Fatalf("se esperaba *shared.DomainError, se obtuvo %T: %v", err, err)
	}
	if domainErr.Code != "SEAT_ALREADY_RESERVED" {
		t.Fatalf("se esperaba código SEAT_ALREADY_RESERVED, se obtuvo %s", domainErr.Code)
	}
	if resRepo.saved != nil {
		t.Fatal("no debería haberse guardado ninguna reserva si el asiento no estaba disponible")
	}
	if cache.setCalled {
		t.Fatal("no debería haberse registrado TTL si el asiento no estaba disponible")
	}
}
