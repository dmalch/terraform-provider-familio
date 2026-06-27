package familio

import (
	"context"
	"net/http"
)

// WeddingEvent builds a marriage event linking two existing persons as spouses.
// Pass nil for an unknown date.
func WeddingEvent(date *DatePart, partnerA, partnerB string) Event {
	return Event{
		Type: EventWedding,
		Date: equalDate(date),
		Participants: []Participant{
			{PersonUUID: partnerA, Role: RoleSpouse},
			{PersonUUID: partnerB, Role: RoleSpouse},
		},
	}
}

// CreateEvent posts a life event anchored on a person
// (POST /api/v2/persons/<uuid>/events). The event is visible on every
// participant's /events; anchor on any participant. Returns the created event
// (with its new uuid).
func (c *Client) CreateEvent(ctx context.Context, personUUID string, ev Event) (*Event, error) {
	req, err := c.newAuthedRequest(ctx, http.MethodPost, "persons/"+personUUID+"/events", nil, ev)
	if err != nil {
		return nil, err
	}
	var out Event
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteEvent removes an event
// (DELETE /api/v2/persons/<uuid>/events/<eventUuid>). personUUID may be any
// participant of the event.
func (c *Client) DeleteEvent(ctx context.Context, personUUID, eventUUID string) error {
	req, err := c.newAuthedRequest(ctx, http.MethodDelete, "persons/"+personUUID+"/events/"+eventUUID, nil, nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// SpouseUUIDs returns the personUuids of an event's spouse participants.
func (e *Event) SpouseUUIDs() []string {
	var out []string
	for _, p := range e.Participants {
		if p.Role == RoleSpouse {
			out = append(out, p.PersonUUID)
		}
	}
	return out
}
