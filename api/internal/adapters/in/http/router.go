package http

import (
	"net/http"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/in/http/handler"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/in/http/middleware"
	wsadapter "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/in/websocket"
	usecasein "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/in"
)

// RouterConfig agrupa lo que el router necesita para construirse.
// allowedOrigins viene de configuración, nunca hardcodeado a "*".
type RouterConfig struct {
	ReserveSeat       handler.ReservationEnqueuer // encola en la sala de espera virtual (RabbitMQ)
	ConfirmPayment    usecasein.ConfirmPaymentUseCase
	CancelReservation usecasein.CancelReservationUseCase
	ListSeats         usecasein.ListAvailableSeatsUseCase
	Hub               *wsadapter.Hub
	AllowedOrigins    map[string]bool
	HealthChecks      map[string]handler.PingFunc
}

// NewRouter usa el enhanced ServeMux de Go 1.22+ (soporta method +
// path patterns nativamente, sin depender de un router de terceros
// para esto). Menos dependencias = menos superficie de ataque.
func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	seatHandler := handler.NewSeatHandler(cfg.ListSeats)
	reservationHandler := handler.NewReservationHandler(cfg.ReserveSeat, cfg.ConfirmPayment, cfg.CancelReservation)

	mux.HandleFunc("GET /health", handler.Health(cfg.HealthChecks))
	mux.HandleFunc("GET /events/{eventID}/seats", seatHandler.ListSeats)
	mux.HandleFunc("POST /reservations", reservationHandler.Reserve)
	mux.HandleFunc("POST /reservations/{reservationID}/confirm", reservationHandler.Confirm)
	mux.HandleFunc("POST /reservations/{reservationID}/cancel", reservationHandler.Cancel)
	mux.HandleFunc("GET /ws/events/{eventID}", func(w http.ResponseWriter, r *http.Request) {
		cfg.Hub.ServeWS(w, r, r.PathValue("eventID"))
	})

	rateLimiter := middleware.NewRateLimiter(15, 30) // 15 req/s sostenidas, ráfaga de 30

	// Orden de la cadena (se ejecuta de arriba hacia abajo):
	// 1. Recover     -> ningún panic tumba el proceso
	// 2. Logging     -> se audita toda petición, incluso las bloqueadas después
	// 3. SecurityHeaders
	// 4. CORS
	// 5. MaxBodyBytes -> corta payloads gigantes antes de parsear JSON
	// 6. RateLimiter  -> lo último antes del negocio
	return middleware.Chain(mux,
		middleware.Recover,
		middleware.Logging,
		middleware.SecurityHeaders,
		middleware.CORS(cfg.AllowedOrigins),
		middleware.MaxBodyBytes(1<<20), // 1 MiB máximo por request
		rateLimiter.Middleware,
	)
}
