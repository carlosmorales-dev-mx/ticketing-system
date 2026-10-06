package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	APIPort              string
	DatabaseURL          string
	RedisAddr            string
	RabbitMQURL          string
	ReservationTTLSecs   int
	AllowedOrigins       map[string]bool
	ReservationWorkers   int
	SweepIntervalSeconds int
	// EnableMapReset activa POST /events/{id}/reset (reinicia el mapa de
	// asientos). Solo para desarrollo/demos: por defecto está apagado.
	EnableMapReset bool
}

func Load() Config {
	origins := getEnv("ALLOWED_ORIGINS", "http://localhost:3000")
	allowed := make(map[string]bool)
	for _, o := range strings.Split(origins, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			allowed[o] = true
		}
	}

	return Config{
		APIPort:              getEnv("API_PORT", "8080"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://ticketing:ticketing_dev_password@localhost:5432/ticketing?sslmode=disable"),
		RedisAddr:            getEnv("REDIS_ADDR", "localhost:6379"),
		RabbitMQURL:          getEnv("RABBITMQ_URL", "amqp://ticketing:ticketing_dev_password@localhost:5672/"),
		ReservationTTLSecs:   getEnvInt("RESERVATION_TTL_SECONDS", 600),
		AllowedOrigins:       allowed,
		ReservationWorkers:   getEnvInt("RESERVATION_WORKERS", 5),
		SweepIntervalSeconds: getEnvInt("SWEEP_INTERVAL_SECONDS", 30),
		EnableMapReset:       getEnvBool("ENABLE_MAP_RESET", false),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
