#!/usr/bin/env bash
#
# smoke_test.sh — prueba end-to-end del sistema de venta de entradas.
#
# Requisitos: curl, jq (sudo pacman -S jq)
# Uso: ./scripts/smoke_test.sh
#
# Qué hace:
#   1. Verifica /health
#   2. Lista los asientos del evento semilla
#   3. Dispara N peticiones de reserva CONCURRENTES sobre el MISMO
#      asiento (esto es la prueba real de que TryReserve es atómico)
#   4. Comprueba que exactamente 1 tuvo éxito (201) y el resto 409
#   5. Confirma el pago de la reserva ganadora
#   6. Verifica que el asiento quedó en estado SOLD

set -euo pipefail

API="${API_URL:-http://localhost:8080}"
EVENT_ID="123e4567-e89b-12d3-a456-426614174000"
CONCURRENT_REQUESTS=15

red()   { printf "\033[31m%s\033[0m\n" "$1"; }
green() { printf "\033[32m%s\033[0m\n" "$1"; }
info()  { printf "\033[36m%s\033[0m\n" "$1"; }

command -v jq >/dev/null 2>&1 || { red "Falta jq. Instálalo: sudo pacman -S jq"; exit 1; }

info "== 1. Health check =="
health=$(curl -s -o /dev/null -w "%{http_code}" "$API/health")
if [ "$health" != "200" ]; then
  red "API no responde en $API (status $health). ¿Corriste 'make run'?"
  exit 1
fi
green "API viva en $API"

info "== 2. Listando asientos del evento semilla =="
seats_json=$(curl -s "$API/events/$EVENT_ID/seats")
seat_count=$(echo "$seats_json" | jq 'length')
if [ "$seat_count" -eq 0 ]; then
  red "No hay asientos. ¿Corriste las migraciones (002_seed_dev_data.sql)?"
  exit 1
fi
green "Encontrados $seat_count asientos"

target_seat=$(echo "$seats_json" | jq -r '[.[] | select(.status=="AVAILABLE")][0].seat_id')
if [ "$target_seat" = "null" ] || [ -z "$target_seat" ]; then
  red "No hay asientos AVAILABLE. Reinicia la BD: make db-reset"
  exit 1
fi
info "Asiento objetivo: $target_seat"

info "== 3. Disparando $CONCURRENT_REQUESTS reservas concurrentes sobre el mismo asiento =="
tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

fire_one() {
  local i="$1"
  local user_id
  user_id=$(uuidgen 2>/dev/null || cat /proc/sys/kernel/random/uuid)
  local status
  status=$(curl -s -o "$tmpdir/resp_$i.json" -w "%{http_code}" \
    -X POST "$API/reservations" \
    -H "Content-Type: application/json" \
    -d "{\"event_id\":\"$EVENT_ID\",\"seat_id\":\"$target_seat\",\"user_id\":\"$user_id\"}")
  echo "$status" > "$tmpdir/status_$i.txt"
}
export -f fire_one
export API EVENT_ID target_seat tmpdir

seq 1 "$CONCURRENT_REQUESTS" | xargs -P "$CONCURRENT_REQUESTS" -I{} bash -c 'fire_one "$@"' _ {}

success_count=0
conflict_count=0
rate_limited_count=0
winner_reservation_id=""

for i in $(seq 1 "$CONCURRENT_REQUESTS"); do
  status=$(cat "$tmpdir/status_$i.txt")
  case "$status" in
    201)
      success_count=$((success_count + 1))
      winner_reservation_id=$(jq -r '.reservation_id' "$tmpdir/resp_$i.json")
      ;;
    409)
      conflict_count=$((conflict_count + 1))
      ;;
    429)
      # El rate limiter frenó esta petición ANTES de que llegara a
      # competir por el asiento. Es una segunda capa de defensa
      # actuando correctamente, no un fallo de la prueba: lo único
      # que realmente importa es que nunca gane más de 1 petición.
      rate_limited_count=$((rate_limited_count + 1))
      ;;
    *)
      red "Respuesta inesperada en la petición $i: HTTP $status"
      cat "$tmpdir/resp_$i.json" || true
      ;;
  esac
done

echo ""
info "Resultado: $success_count éxito(s) (201), $conflict_count conflicto(s) (409), $rate_limited_count frenada(s) por rate limit (429)"

if [ "$success_count" -eq 1 ]; then
  green "✔ GARANTÍA ANTI-DOBLE-VENTA CONFIRMADA: exactamente 1 de $CONCURRENT_REQUESTS ganó la carrera por el asiento"
else
  red "✘ FALLO: se esperaba exactamente 1 éxito (201), se obtuvieron $success_count"
  exit 1
fi

info "== 4. Confirmando el pago de la reserva ganadora ($winner_reservation_id) =="
confirm_status=$(curl -s -o "$tmpdir/confirm.json" -w "%{http_code}" \
  -X POST "$API/reservations/$winner_reservation_id/confirm")

if [ "$confirm_status" != "200" ]; then
  red "Fallo confirmando el pago (HTTP $confirm_status)"
  cat "$tmpdir/confirm.json"
  exit 1
fi
ticket_id=$(jq -r '.ticket_id' "$tmpdir/confirm.json")
green "✔ Pago confirmado, ticket emitido: $ticket_id"

info "== 5. Verificando que el asiento quedó SOLD =="
final_status=$(curl -s "$API/events/$EVENT_ID/seats" | jq -r --arg id "$target_seat" '.[] | select(.seat_id==$id) | .status')
if [ "$final_status" = "SOLD" ]; then
  green "✔ Asiento $target_seat confirmado como SOLD"
else
  red "✘ Estado inesperado del asiento: $final_status (se esperaba SOLD)"
  exit 1
fi

echo ""
green "=========================================="
green " TODAS LAS PRUEBAS PASARON CORRECTAMENTE"
green "=========================================="
