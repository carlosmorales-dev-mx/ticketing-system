package middleware

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Recover evita que un panic en cualquier handler tumbe todo el
// proceso: lo captura, lo loguea, y responde 500 sin filtrar detalles
// internos al cliente (el mensaje de error real solo va al log del
// servidor).
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recuperado en %s %s: %v", r.Method, r.URL.Path, err)
				http.Error(w, `{"error":"INTERNAL_ERROR"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Logging registra método, ruta, status y duración de cada request.
// No loguea bodies ni headers para evitar filtrar datos sensibles.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.status, time.Since(start))
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

// Hijack reenvía la capacidad de "hijacking" de la conexión TCP al
// ResponseWriter original. Sin este método, statusWriter (que envuelve
// el ResponseWriter para poder loguear el status code) rompe cualquier
// protocolo que necesite tomar control directo de la conexión —como
// WebSocket— porque gorilla/websocket hace un type assertion a
// http.Hijacker y, sin este método, esa aserción falla.
func (sw *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := sw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("el ResponseWriter subyacente no soporta hijacking")
	}
	return hijacker.Hijack()
}

// MaxBodyBytes limita el tamaño del body de la petición. Sin esto,
// un cliente malicioso podría enviar un payload gigante y agotar
// memoria del servidor (uno de los vectores de DoS más básicos y
// más fáciles de prevenir).
func MaxBodyBytes(max int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, max)
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders añade cabeceras defensivas estándar.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

// CORS restringe qué orígenes pueden llamar a la API desde el
// navegador. allowedOrigins debe venir de configuración (env var),
// nunca hardcodeado a "*" en un entorno con credenciales.
func CORS(allowedOrigins map[string]bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowedOrigins[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimiter implementa un limitador por IP con el algoritmo de
// token bucket (golang.org/x/time/rate). Cada IP tiene su propio
// balde de tokens que se rellena a un ritmo fijo; si se agota,
// responde 429 con la cabecera Retry-After, tal como se documenta
// en el OpenAPI.
//
// Por qué por IP y no global: un límite global permitiría que un
// solo cliente abusivo consuma toda la cuota y bloquee al resto de
// usuarios legítimos.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*rate.Limiter
	rps      rate.Limit
	burst    int
}

func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rps:      rate.Limit(requestsPerSecond),
		burst:    burst,
	}
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rps, rl.burst)
		rl.visitors[ip] = limiter
	}
	return limiter
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			// Retry-After documentado en el 429 del OpenAPI.
			w.Header().Set("Retry-After", "1")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"RATE_LIMITED"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	// Confía en X-Forwarded-For solo si hay un proxy reverso delante
	// (ajustar en despliegue real); en local, usa RemoteAddr.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// Chain compone varios middlewares en orden de ejecución (el primero
// de la lista se ejecuta primero).
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
