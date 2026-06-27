package familio

import (
	"context"
	"net/http"
)

// WeddingEvent builds a marriage event linking two existing persons as spouses.
// Pass nil for an unknown date.
func WeddingEvent(date *DateRange, partnerA, partnerB string) Event {
	return Event{
		Type: EventWedding,
		Date: EventDateFromRange(date),
		Participants: []Participant{
			{PersonUUID: partnerA, Role: RoleSpouse},
			{PersonUUID: partnerB, Role: RoleSpouse},
		},
	}
}

// FactEventTypes is the set of single-subject "fact" event types a familio_event
// resource may carry (role "owner"). It EXCLUDES the events managed by dedicated
// surfaces — birth/death/baptism (folded into familio_person) and the true
// two-person relationship events wedding/divorce/affiance/nikah — so the two
// never fight over one event.
//
// godparent (Восприемник) and warranter (Поручитель) ARE included: per Familio's
// own model (confirmed by their team in the official group) these are
// single-subject events recorded on the godparent/witness themselves — Familio
// deliberately does NOT link them to the specific godchild/party, so the
// godchild is noted in the comment, exactly as Familio's power users do.
// Source: the editor's event catalogue (internal/familio/API.md).
var FactEventTypes = []string{
	"arrest", "barMitzvah", "batMitzvah", "blessing", "militaryAward", "militaryService",
	"citizenship", "titleOfNobility", "demobilization", "immigration", "naming", "confirmation",
	"concentrationCamp", "location", "award", "education", "circumcision", "condemnation",
	"sentenceServing", "reburial", "militaryRankObtaining", "captured", "ordination", "burial",
	"conscription", "missing", "profession", "convictRehabilitation", "renaming", "resurnaming",
	"crime", "hajj", "exhumation", "emigration", "injury", "travel", "pilgrimage", "collectiveFarm",
	"party", "evacuation", "scienceDegree", "dekulakization", "treatment", "combat",
	"militaryCemetery", "heroicAct", "reference", "godparent", "warranter",
}

// FactEvent builds a single-subject fact event (role "owner") of the given type
// for ownerRef, with a free-text comment. Pass nil for an unknown date.
func FactEvent(eventType string, date *DateRange, ownerRef, comment string) Event {
	return Event{
		Type:         eventType,
		Date:         EventDateFromRange(date),
		Participants: []Participant{{PersonUUID: ownerRef, Role: RoleOwner}},
		Comment:      comment,
	}
}

// FindByID returns the event with the given uuid, or nil.
func FindByID(events []Event, uuid string) *Event {
	for i := range events {
		if events[i].ID() == uuid {
			return &events[i]
		}
	}
	return nil
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

// ParentUUIDs returns the personUuids of an event's parent participants.
func (e *Event) ParentUUIDs() []string {
	var out []string
	for _, p := range e.Participants {
		if p.Role == RoleParent {
			out = append(out, p.PersonUUID)
		}
	}
	return out
}

// hasChild reports whether personUUID is the child participant of the event.
func (e *Event) hasChild(personUUID string) bool {
	for _, p := range e.Participants {
		if p.Role == RoleChild && p.PersonUUID == personUUID {
			return true
		}
	}
	return false
}

// OwnBirthEvent returns the birth event in which personUUID is the child — the
// person's *own* birth. A person who is also a parent has the births of their
// children on their /events too (where they are the parent), so a plain
// type-filter is not enough. Returns nil when there is no such event yet. On a
// create read-back the child participant carries the resolved uuid, so passing
// the new uuid works there too.
func OwnBirthEvent(events []Event, personUUID string) *Event {
	for i := range events {
		if events[i].Type == EventBirth && events[i].hasChild(personUUID) {
			return &events[i]
		}
	}
	return nil
}
