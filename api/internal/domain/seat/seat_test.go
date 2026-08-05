package seat_test

import (
	"testing"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/seat"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

func TestSeat_Reserve_FromAvailable_Succeeds(t *testing.T) {
	s := seat.NewSeat(shared.NewID(), 1, "A")

	if err := s.Reserve(); err != nil {
		t.Fatalf("se esperaba éxito, se obtuvo error: %v", err)
	}
	if s.Status() != seat.StatusReserved {
		t.Fatalf("se esperaba status RESERVED, se obtuvo %s", s.Status())
	}
}

// Este es el test que más importa de todo el proyecto: la regla de
// negocio central (un asiento ya reservado no puede reservarse otra
// vez) vive en el dominio y se puede probar sin tocar Postgres.
func TestSeat_Reserve_AlreadyReserved_Fails(t *testing.T) {
	s := seat.NewSeat(shared.NewID(), 1, "A")
	_ = s.Reserve()

	err := s.Reserve()
	if err == nil {
		t.Fatal("se esperaba error al reservar un asiento ya reservado, se obtuvo nil")
	}

	var domainErr *shared.DomainError
	if ok := asDomainError(err, &domainErr); !ok {
		t.Fatalf("se esperaba un *shared.DomainError, se obtuvo %T", err)
	}
	if domainErr.Code != "SEAT_ALREADY_RESERVED" {
		t.Fatalf("se esperaba código SEAT_ALREADY_RESERVED, se obtuvo %s", domainErr.Code)
	}
}

func TestSeat_Sell_WithoutReserving_Fails(t *testing.T) {
	s := seat.NewSeat(shared.NewID(), 1, "A")

	if err := s.Sell(); err == nil {
		t.Fatal("se esperaba error al vender un asiento no reservado, se obtuvo nil")
	}
}

func TestSeat_FullLifecycle_AvailableToReservedToSold(t *testing.T) {
	s := seat.NewSeat(shared.NewID(), 1, "A")

	if err := s.Reserve(); err != nil {
		t.Fatalf("Reserve falló: %v", err)
	}
	if err := s.Sell(); err != nil {
		t.Fatalf("Sell falló: %v", err)
	}
	if s.Status() != seat.StatusSold {
		t.Fatalf("se esperaba status SOLD, se obtuvo %s", s.Status())
	}
}

func TestSeat_Release_ReturnsToAvailable(t *testing.T) {
	s := seat.NewSeat(shared.NewID(), 1, "A")
	_ = s.Reserve()

	if err := s.Release(); err != nil {
		t.Fatalf("Release falló: %v", err)
	}
	if s.Status() != seat.StatusAvailable {
		t.Fatalf("se esperaba status AVAILABLE, se obtuvo %s", s.Status())
	}
}

func asDomainError(err error, target **shared.DomainError) bool {
	de, ok := err.(*shared.DomainError)
	if ok {
		*target = de
	}
	return ok
}
