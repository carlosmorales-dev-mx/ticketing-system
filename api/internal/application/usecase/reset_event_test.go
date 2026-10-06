package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	usecasein "github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/in"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/out"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/usecase"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

type fakeResetRepo struct {
	result *out.EventResetResult
	err    error
}

func (f *fakeResetRepo) ResetEvent(ctx context.Context, eventID shared.ID) (*out.EventResetResult, error) {
	return f.result, f.err
}

type recordingCache struct{ deleted []shared.ID }

func (c *recordingCache) SetWithTTL(ctx context.Context, r, s shared.ID, ttl time.Duration) error {
	return nil
}
func (c *recordingCache) Delete(ctx context.Context, r shared.ID) error {
	c.deleted = append(c.deleted, r)
	return nil
}
func (c *recordingCache) SubscribeExpirations(ctx context.Context) (<-chan shared.ID, error) {
	ch := make(chan shared.ID)
	close(ch)
	return ch, nil
}

type recordingNotifier struct{ updates []string }

func (n *recordingNotifier) BroadcastSeatUpdate(ctx context.Context, eventID, seatID shared.ID, status string) error {
	n.updates = append(n.updates, status)
	return nil
}

func TestResetEventUseCase_Execute_ClearsCacheAndBroadcastsEverySeat(t *testing.T) {
	repo := &fakeResetRepo{result: &out.EventResetResult{
		ReleasedSeatIDs:       []shared.ID{shared.NewID(), shared.NewID(), shared.NewID()},
		PendingReservationIDs: []shared.ID{shared.NewID(), shared.NewID()},
	}}
	cache := &recordingCache{}
	notifier := &recordingNotifier{}

	uc := usecase.NewResetEventUseCase(repo, cache, notifier)
	res, err := uc.Execute(context.Background(), usecasein.ResetEventCommand{EventID: shared.NewID()})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if res.ReleasedSeats != 3 {
		t.Fatalf("se esperaban 3 asientos liberados, se obtuvo %d", res.ReleasedSeats)
	}
	if len(cache.deleted) != 2 {
		t.Fatalf("se esperaba borrar 2 TTL de Redis, se borraron %d", len(cache.deleted))
	}
	if len(notifier.updates) != 3 {
		t.Fatalf("se esperaban 3 avisos por WebSocket, hubo %d", len(notifier.updates))
	}
	for _, s := range notifier.updates {
		if s != "AVAILABLE" {
			t.Fatalf("todos los avisos deben ser AVAILABLE, llegó %q", s)
		}
	}
}

func TestResetEventUseCase_Execute_RepoError_NoSideEffects(t *testing.T) {
	boom := shared.NewDomainError("EVENT_NOT_FOUND", "el evento no existe")
	cache := &recordingCache{}
	notifier := &recordingNotifier{}

	uc := usecase.NewResetEventUseCase(&fakeResetRepo{err: boom}, cache, notifier)
	_, err := uc.Execute(context.Background(), usecasein.ResetEventCommand{EventID: shared.NewID()})

	var de *shared.DomainError
	if !errors.As(err, &de) || de.Code != "EVENT_NOT_FOUND" {
		t.Fatalf("se esperaba EVENT_NOT_FOUND, se obtuvo: %v", err)
	}
	if len(cache.deleted) != 0 || len(notifier.updates) != 0 {
		t.Fatal("si el reinicio falla no debe haber avisos ni borrados en cache")
	}
}
