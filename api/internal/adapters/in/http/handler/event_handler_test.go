package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/adapters/in/http/handler"
	usecasein "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/in"
)

type stubReset struct{ calls int }

func (s *stubReset) Execute(ctx context.Context, cmd usecasein.ResetEventCommand) (*usecasein.ResetEventResult, error) {
	s.calls++
	return &usecasein.ResetEventResult{ReleasedSeats: 7}, nil
}

func serve(h *handler.EventHandler, id string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /events/{eventID}/reset", h.Reset)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/events/"+id+"/reset", nil))
	return rec
}

func TestEventHandler_Reset_DisabledByDefault_Returns403AndDoesNothing(t *testing.T) {
	uc := &stubReset{}
	rec := serve(handler.NewEventHandler(uc, false), "123e4567-e89b-12d3-a456-426614174000")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("se esperaba 403, se obtuvo %d", rec.Code)
	}
	if uc.calls != 0 {
		t.Fatal("con la herramienta apagada no debe ejecutarse el caso de uso")
	}
}

func TestEventHandler_Reset_Enabled_Returns200WithCount(t *testing.T) {
	uc := &stubReset{}
	rec := serve(handler.NewEventHandler(uc, true), "123e4567-e89b-12d3-a456-426614174000")
	if rec.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, se obtuvo %d", rec.Code)
	}
	if got := rec.Body.String(); got != "{\"released_seats\":7}\n" {
		t.Fatalf("cuerpo inesperado: %q", got)
	}
}

func TestEventHandler_Reset_InvalidID_Returns400(t *testing.T) {
	rec := serve(handler.NewEventHandler(&stubReset{}, true), "no-es-uuid")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("se esperaba 400, se obtuvo %d", rec.Code)
	}
}
