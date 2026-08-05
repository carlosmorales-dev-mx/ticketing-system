package reservation_test

import (
	"testing"
	"time"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/reservation"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

func TestReservation_NewReservation_SetsDefaultTTL(t *testing.T) {
	before := time.Now()
	r := reservation.NewReservation(shared.NewID(), shared.NewID(), shared.NewID())
	after := time.Now()

	minExpected := before.Add(reservation.DefaultTTL)
	maxExpected := after.Add(reservation.DefaultTTL)

	if r.ExpiresAt().Before(minExpected) || r.ExpiresAt().After(maxExpected) {
		t.Fatalf("ExpiresAt fuera del rango esperado: got=%v want entre %v y %v", r.ExpiresAt(), minExpected, maxExpected)
	}
	if r.Status() != reservation.StatusPending {
		t.Fatalf("se esperaba status PENDING, se obtuvo %s", r.Status())
	}
}

func TestReservation_IsExpired_BeforeTTL_False(t *testing.T) {
	r := reservation.NewReservation(shared.NewID(), shared.NewID(), shared.NewID())

	if r.IsExpired(time.Now()) {
		t.Fatal("una reserva recién creada no debería estar expirada")
	}
}

func TestReservation_IsExpired_AfterTTL_True(t *testing.T) {
	r := reservation.NewReservation(shared.NewID(), shared.NewID(), shared.NewID())

	future := time.Now().Add(reservation.DefaultTTL + time.Minute)
	if !r.IsExpired(future) {
		t.Fatal("una reserva PENDING pasado su TTL debería considerarse expirada")
	}
}

func TestReservation_Confirm_WhenPending_Succeeds(t *testing.T) {
	r := reservation.NewReservation(shared.NewID(), shared.NewID(), shared.NewID())

	if err := r.Confirm(); err != nil {
		t.Fatalf("Confirm falló: %v", err)
	}
	if r.Status() != reservation.StatusConfirmed {
		t.Fatalf("se esperaba status CONFIRMED, se obtuvo %s", r.Status())
	}
}

// Caso de borde clave: no se puede confirmar dos veces la misma
// reserva (protege contra doble-click / doble envío del formulario).
func TestReservation_Confirm_Twice_SecondFails(t *testing.T) {
	r := reservation.NewReservation(shared.NewID(), shared.NewID(), shared.NewID())
	_ = r.Confirm()

	if err := r.Confirm(); err == nil {
		t.Fatal("se esperaba error al confirmar una reserva ya confirmada")
	}
}

func TestReservation_Expire_WhenNotPending_Fails(t *testing.T) {
	r := reservation.NewReservation(shared.NewID(), shared.NewID(), shared.NewID())
	_ = r.Confirm()

	if err := r.Expire(); err == nil {
		t.Fatal("una reserva ya confirmada no debería poder expirar")
	}
}
