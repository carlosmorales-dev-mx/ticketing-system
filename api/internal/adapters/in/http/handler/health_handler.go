package handler

import (
	"context"
	"net/http"
)

// PingFunc comprueba una dependencia externa. El handler de salud no
// conoce Postgres, Redis ni RabbitMQ directamente — solo recibe estas
// funciones ya cerradas sobre la conexión real desde cmd/api/main.go.
// Así este paquete no necesita importar los drivers de infraestructura.
type PingFunc func(ctx context.Context) error

// Health responde 200 si todas las dependencias están vivas, o 503 si
// alguna falla — así un balanceador de carga o un orquestador puede
// saber de verdad si el servicio está listo para recibir tráfico, en
// vez de un "200 ok" ciego que no comprueba nada.
func Health(checks map[string]PingFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := make(map[string]string, len(checks))
		healthy := true

		for name, check := range checks {
			if err := check(r.Context()); err != nil {
				results[name] = "down"
				healthy = false
			} else {
				results[name] = "ok"
			}
		}

		status := "ok"
		code := http.StatusOK
		if !healthy {
			status = "degraded"
			code = http.StatusServiceUnavailable
		}

		writeJSON(w, code, map[string]any{"status": status, "checks": results})
	}
}
