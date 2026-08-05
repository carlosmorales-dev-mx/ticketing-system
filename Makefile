.DEFAULT_GOAL := help

COMPOSE       := podman-compose -f deploy/podman-compose.yaml
API_DIR       := api
BINARY        := bin/api

.PHONY: help
help: ## Muestra esta ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ---------------------------------------------------------------------------
# Infraestructura (Postgres, Redis, RabbitMQ, Swagger UI)
# ---------------------------------------------------------------------------

.PHONY: up
up: ## Levanta toda la infraestructura (Postgres, Redis, RabbitMQ, Swagger UI)
	$(COMPOSE) up -d

.PHONY: down
down: ## Para los contenedores (conserva los volúmenes/datos)
	$(COMPOSE) down

.PHONY: reset
reset: ## Para y BORRA volúmenes: reset total de datos (Postgres/Redis/RabbitMQ)
	$(COMPOSE) down -v

.PHONY: ps
ps: ## Estado de los contenedores
	$(COMPOSE) ps

.PHONY: logs
logs: ## Logs en vivo de toda la infraestructura
	$(COMPOSE) logs -f

.PHONY: logs-api-deps
logs-api-deps: ## Logs solo de postgres/redis/rabbitmq (útil mientras corres la API local)
	$(COMPOSE) logs -f postgres redis rabbitmq

# ---------------------------------------------------------------------------
# Servicio Go
# ---------------------------------------------------------------------------

.PHONY: run
run: ## Arranca la API en local (requiere: make up antes)
	cd $(API_DIR) && go run ./cmd/api

.PHONY: build
build: ## Compila el binario de la API
	cd $(API_DIR) && go build -o ../$(BINARY) ./cmd/api

.PHONY: test
test: ## Ejecuta los tests con detector de race conditions (crítico en este proyecto)
	cd $(API_DIR) && go test -race -cover ./...

.PHONY: tidy
tidy: ## Sincroniza go.mod / go.sum
	cd $(API_DIR) && go mod tidy

.PHONY: fmt
fmt: ## Formatea el código Go
	cd $(API_DIR) && gofmt -s -w .

.PHONY: vet
vet: ## Análisis estático (go vet)
	cd $(API_DIR) && go vet ./...

.PHONY: lint
lint: fmt vet ## fmt + vet (añade golangci-lint aquí si lo instalas)

# ---------------------------------------------------------------------------
# Base de datos
# ---------------------------------------------------------------------------

.PHONY: db-shell
db-shell: ## Abre una shell psql dentro del contenedor de Postgres
	podman exec -it ticketing-postgres psql -U ticketing -d ticketing

.PHONY: db-reset
db-reset: ## Reinicia solo la base de datos (borra y reaplica migraciones)
	$(COMPOSE) stop postgres
	podman rm -f ticketing-postgres
	podman volume rm -f deploy_postgres_data
	$(COMPOSE) up -d postgres

# ---------------------------------------------------------------------------
# Documentación (OpenAPI / Swagger)
# ---------------------------------------------------------------------------

.PHONY: docs
docs: ## Abre Swagger UI (requiere: make up)
	@echo "Swagger UI -> http://localhost:8081"
	@echo "RabbitMQ management -> http://localhost:15672"

.PHONY: openapi-validate
openapi-validate: ## Valida que openapi.yaml es sintácticamente correcto
	python3 -c "import yaml; yaml.safe_load(open('$(API_DIR)/openapi.yaml')); print('OpenAPI YAML válido ✓')"

.PHONY: smoke-test
smoke-test: ## Prueba end-to-end: reservas concurrentes + verificación anti-doble-venta (requiere: make run en otra terminal)
	./scripts/smoke_test.sh

.PHONY: up-full
up-full: ## Levanta TODO en contenedores (infra + API + frontend) para una demo de un solo comando
	$(COMPOSE) --profile full up -d --build

.PHONY: down-full
down-full: ## Para los contenedores del profile "full" (api + frontend)
	$(COMPOSE) --profile full down

# ---------------------------------------------------------------------------
# Frontend (Nuxt)
# ---------------------------------------------------------------------------

.PHONY: frontend-install
frontend-install: ## Instala las dependencias del frontend con pnpm
	cd frontend && pnpm install

.PHONY: frontend-dev
frontend-dev: ## Arranca solo el frontend Nuxt (pnpm dev, puerto 3000)
	cd frontend && pnpm dev

# ---------------------------------------------------------------------------
# Atajo para el flujo típico de desarrollo
# ---------------------------------------------------------------------------

.PHONY: dev
dev: up run ## make up + make run (solo backend) en un solo comando

.PHONY: dev-full
dev-full: up ## Infra + API + frontend, TODO en un solo comando/terminal (Ctrl+C detiene ambos)
	@trap 'kill 0' EXIT INT TERM; \
	(cd $(API_DIR) && go run ./cmd/api) & \
	(cd frontend && pnpm dev) & \
	wait
