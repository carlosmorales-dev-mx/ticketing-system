package event

import (
	"time"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

// Event representa un evento vendible (concierto, partido, obra...).
type Event struct {
	id       shared.ID
	name     string
	venue    string
	startsAt time.Time
}

func NewEvent(name, venue string, startsAt time.Time) (*Event, error) {
	if name == "" {
		return nil, shared.NewDomainError("INVALID_EVENT_NAME", "el nombre del evento no puede estar vacío")
	}
	if startsAt.Before(time.Now()) {
		return nil, shared.NewDomainError("INVALID_EVENT_DATE", "la fecha del evento no puede ser en el pasado")
	}
	return &Event{
		id:       shared.NewID(),
		name:     name,
		venue:    venue,
		startsAt: startsAt,
	}, nil
}

func Reconstruct(id shared.ID, name, venue string, startsAt time.Time) *Event {
	return &Event{id: id, name: name, venue: venue, startsAt: startsAt}
}

func (e *Event) ID() shared.ID        { return e.id }
func (e *Event) Name() string         { return e.name }
func (e *Event) Venue() string        { return e.venue }
func (e *Event) StartsAt() time.Time  { return e.startsAt }
