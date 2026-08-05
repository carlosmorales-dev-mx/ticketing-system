// main.go es el ÚNICO lugar del proyecto que conoce a la vez el
// dominio, los casos de uso y TODOS los adaptadores concretos. Es el
// "composition root": aquí se decide qué implementación real recibe
// cada puerto. Ningún otro archivo debería tener este nivel de
// acoplamiento.
package main

import (
	"log"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/usecase"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/pkg/config"

	// TODO: descomentar a medida que se implementen los adaptadores
	// concretos (siguientes pasos: esquema Postgres -> adaptador
	// postgres -> adaptador redis -> adaptador rabbitmq -> router http).
	//
	// postgresadapter "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/out/postgres"
	// redisadapter    "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/out/redis"
	// rabbitadapter   "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/out/rabbitmq"
	// wsadapter       "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/in/websocket"
	// httpadapter     "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/in/http"
)

func main() {
	cfg := config.Load()

	// --- 1. Adaptadores de salida (implementan application/port/out) ---
	// seatRepo := postgresadapter.NewSeatRepository(db)
	// reservationRepo := postgresadapter.NewReservationRepository(db)
	// ticketRepo := postgresadapter.NewTicketRepository(db)
	// reservationCache := redisadapter.NewReservationCache(redisClient)
	// publisher := rabbitadapter.NewEventPublisher(amqpConn)
	// notifier := wsadapter.NewHub()

	// --- 2. Casos de uso (dependen SOLO de interfaces port/out) ---
	// reserveSeat := usecase.NewReserveSeatUseCase(seatRepo, reservationRepo, reservationCache, publisher, notifier)
	// confirmPayment := usecase.NewConfirmPaymentUseCase(reservationRepo, seatRepo, ticketRepo, reservationCache, notifier)
	// releaseExpired := usecase.NewReleaseExpiredReservationUseCase(reservationRepo, seatRepo, notifier)
	// listSeats := usecase.NewListAvailableSeatsUseCase(seatRepo)
	_ = usecase.NewListAvailableSeatsUseCase // evita "unused import" mientras se completa el wiring

	// --- 3. Adaptador de entrada HTTP (depende de application/port/in) ---
	// router := httpadapter.NewRouter(reserveSeat, confirmPayment, listSeats)

	// --- 4. Goroutine que escucha expiraciones de Redis y dispara el caso de uso ---
	// go func() {
	//     expirations, _ := reservationCache.SubscribeExpirations(context.Background())
	//     for reservationID := range expirations {
	//         _ = releaseExpired.Execute(context.Background(), reservationID)
	//     }
	// }()

	log.Printf("Configuración cargada. La API arrancaría en :%s (wiring de adaptadores pendiente)", cfg.APIPort)
}
