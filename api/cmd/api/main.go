// main.go es el composition root: el único lugar que conoce a la vez
// el dominio, los casos de uso y todos los adaptadores concretos.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/in/amqpconsumer"
	httpadapter "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/in/http"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/in/http/handler"
	wsadapter "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/in/websocket"
	postgresadapter "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/out/postgres"
	rabbitadapter "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/out/rabbitmq"
	redisadapter "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/out/redis"
	usecasein "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/in"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/out"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/usecase"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/pkg/config"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("error fatal: %v", err)
	}
}

func run() error {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// --- 1. Adaptadores de salida ---
	pool, err := postgresadapter.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	log.Println("conectado a postgres")

	redisClient := redisadapter.NewClient(cfg.RedisAddr)
	defer redisClient.Close()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return err
	}
	log.Println("conectado a redis")

	amqpConn, err := rabbitadapter.NewConnection(cfg.RabbitMQURL)
	if err != nil {
		return err
	}
	defer amqpConn.Close()

	publisher, publisherChannel, err := rabbitadapter.NewEventPublisher(amqpConn)
	if err != nil {
		return err
	}
	defer publisherChannel.Close()

	// Canal separado para el cliente RPC de la sala de espera (no
	// compartir canales AMQP entre goroutines concurrentes: no es
	// seguro en la librería amqp091-go).
	rpcChannel, err := amqpConn.Channel()
	if err != nil {
		return err
	}
	defer rpcChannel.Close()
	reservationQueue, err := rabbitadapter.NewReservationRPCClient(rpcChannel)
	if err != nil {
		return err
	}
	log.Println("conectado a rabbitmq")

	seatRepo := postgresadapter.NewSeatRepository(pool)
	reservationRepo := postgresadapter.NewReservationRepository(pool)
	ticketRepo := postgresadapter.NewTicketRepository(pool)
	resetRepo := postgresadapter.NewEventResetRepository(pool)
	reservationCache := redisadapter.NewReservationCache(redisClient)
	hub := wsadapter.NewHub()

	// --- 2. Casos de uso ---
	reserveSeat := usecase.NewReserveSeatUseCase(seatRepo, reservationRepo, reservationCache, publisher, hub)
	confirmPayment := usecase.NewConfirmPaymentUseCase(reservationRepo, seatRepo, ticketRepo, reservationCache, hub)
	releaseExpired := usecase.NewReleaseExpiredReservationUseCase(reservationRepo, seatRepo, hub)
	listSeats := usecase.NewListAvailableSeatsUseCase(seatRepo)
	cancelReservation := usecase.NewCancelReservationUseCase(reservationRepo, seatRepo, reservationCache, hub)

	resetEvent := usecase.NewResetEventUseCase(resetRepo, reservationCache, hub)
	if cfg.EnableMapReset {
		log.Println("AVISO: ENABLE_MAP_RESET=true -> POST /events/{id}/reset está activo (solo desarrollo)")
	}

	// --- 3. Sala de espera virtual: workers que SÍ ejecutan el caso
	//     de uso, consumiendo la cola que alimenta reservationQueue ---
	if err := amqpconsumer.StartWorkers(ctx, amqpConn, reserveSeat, cfg.ReservationWorkers); err != nil {
		return err
	}

	// --- 4. Goroutine: liberar reservas cuando Redis avisa que expiraron ---
	expirations, err := reservationCache.SubscribeExpirations(ctx)
	if err != nil {
		return err
	}
	go func() {
		for reservationID := range expirations {
			reqCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := releaseExpired.Execute(reqCtx, reservationID); err != nil {
				log.Printf("error liberando reserva expirada (evento redis) %s: %v", reservationID, err)
			}
			cancel()
		}
	}()

	// --- 5. Barrido periódico: red de seguridad si Redis pierde el
	//     evento de expiración (reinicio, partición de red, etc.) ---
	go runExpirationSweep(ctx, reservationRepo, releaseExpired, time.Duration(cfg.SweepIntervalSeconds)*time.Second)

	// --- 6. Adaptador de entrada HTTP ---
	router := httpadapter.NewRouter(httpadapter.RouterConfig{
		ReserveSeat:       reservationQueue, // pasa por la sala de espera, no llama reserveSeat directo
		ConfirmPayment:    confirmPayment,
		CancelReservation: cancelReservation,
		ListSeats:         listSeats,
		ResetEvent:        resetEvent,
		EnableMapReset:    cfg.EnableMapReset,
		Hub:               hub,
		AllowedOrigins:    cfg.AllowedOrigins,
		HealthChecks: map[string]handler.PingFunc{
			"postgres": func(ctx context.Context) error { return pool.Ping(ctx) },
			"redis":    func(ctx context.Context) error { return redisClient.Ping(ctx).Err() },
			"rabbitmq": func(ctx context.Context) error {
				if amqpConn.IsClosed() {
					return errAMQPClosed
				}
				return nil
			},
		},
	})

	server := &http.Server{
		Addr:              ":" + cfg.APIPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("API escuchando en :%s", cfg.APIPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		log.Println("señal de apagado recibida, drenando conexiones...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}

var errAMQPClosed = errors.New("la conexión AMQP está cerrada")

// runExpirationSweep es la red de seguridad del sistema de reservas:
// cada `interval`, pregunta a Postgres qué reservas PENDING ya
// vencieron y las libera, por si el evento de expiración de Redis se
// perdió en el camino.
func runExpirationSweep(
	ctx context.Context,
	reservationRepo out.ReservationRepository,
	releaseExpired usecasein.ReleaseExpiredReservationUseCase,
	interval time.Duration,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sweepCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			ids, err := reservationRepo.FindExpiredPending(sweepCtx, time.Now())
			if err != nil {
				log.Printf("error en el barrido de reservas vencidas: %v", err)
				cancel()
				continue
			}
			for _, id := range ids {
				if err := releaseExpired.Execute(sweepCtx, id); err != nil {
					log.Printf("error liberando reserva vencida (barrido) %s: %v", id, err)
				}
			}
			if len(ids) > 0 {
				log.Printf("barrido de expiración: %d reserva(s) liberada(s)", len(ids))
			}
			cancel()
		}
	}
}
