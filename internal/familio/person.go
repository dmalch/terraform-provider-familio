package familio

import (
	"context"
	"net/http"
	"net/url"
)

// Gender and privacy literals accepted by familio's /api/v2/persons surface.
const (
	GenderMale   = "male"
	GenderFemale = "female"

	PrivacyVisibleForAll = "visible_for_all"
	PrivacyInvisible     = "invisible"
)

// Event types and the participant roles seen in the tree editor.
const (
	EventBirth   = "birth"
	EventDeath   = "death"
	EventWedding = "wedding"

	RoleChild  = "child"
	RoleOwner  = "owner"
	RoleSpouse = "spouse"

	// SelfRef is the placeholder personUuid for the person being created in the
	// same POST /api/v2/persons request (resolved server-side to the new uuid).
	SelfRef = "self"

	calendarGregorian = "gregorian"
	dateTypeEqual     = "equal"
)

// BasicFields are the editable name/gender/privacy fields of a tree person —
// the "basic" object of POST /persons and the (flat) body of PUT /persons/<id>/basic.
type BasicFields struct {
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	MiddleName     string `json:"middleName"`     // отчество (patronymic)
	BirthFirstName string `json:"birthFirstName"` // maiden given name
	BirthLastName  string `json:"birthLastName"`  // maiden surname
	Gender         string `json:"gender"`
	Privacy        string `json:"privacy"`
}

// DatePart is one endpoint of a complex date. Year is required; Month/Day are
// optional (nil ⇒ unspecified). Type carries the per-part calendar.
type DatePart struct {
	Day   *int   `json:"day"`
	Month *int   `json:"month"`
	Year  int    `json:"year"`
	Type  string `json:"type"`
}

// EventDate is familio's complex-date object. A nil First means "date unknown".
type EventDate struct {
	Calendar string    `json:"calendar"`
	Type     string    `json:"type"`
	First    *DatePart `json:"first"`
	Second   *DatePart `json:"second"`
	// Formatted is server-computed on reads; never sent.
	Formatted string `json:"formatted,omitempty"`
}

// Participant links a person into an event by role.
type Participant struct {
	PersonUUID  string `json:"personUuid"`
	Role        string `json:"role"`
	DisplayName string `json:"displayName,omitempty"`
	Gender      string `json:"gender,omitempty"`
}

// Event is a life event (birth/death/wedding) with its participants.
type Event struct {
	UUID         *string       `json:"uuid"` // null on create; real uuid on read-back
	Type         string        `json:"type"`
	Date         EventDate     `json:"date"`
	Participants []Participant `json:"participants"`
	Settlement   *string       `json:"settlement"`
	Comment      string        `json:"comment"`
	CreatedAt    string        `json:"createdAt,omitempty"`
	UpdatedAt    string        `json:"updatedAt,omitempty"`
}

// ID returns the event's uuid (empty when unset, e.g. a request-side event).
func (e *Event) ID() string {
	if e.UUID == nil {
		return ""
	}
	return *e.UUID
}

// BasicRecord is the basic person view. GET /persons/<uuid>/basic returns it
// flat (no displayName); the POST /persons 201 nests it under "basic" and does
// include displayName, so the field is populated on create but empty on a
// /basic read (use GetPersonDisplay for reads).
type BasicRecord struct {
	BasicFields
	UUID        string `json:"uuid"`
	DisplayName string `json:"displayName"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// personDisplay is the slice of the regularPerson view (GET /persons/<uuid>)
// the provider surfaces as the computed display_name.
type personDisplay struct {
	UUID        string `json:"uuid"`
	DisplayName string `json:"displayName"`
}

// personCreateBody is the POST /persons request envelope.
type personCreateBody struct {
	Basic     BasicFields `json:"basic"`
	Photo     *string     `json:"photo"`
	Events    []Event     `json:"events"`
	Biography *string     `json:"biography"`
}

// basicUpdateBody is the (flat) PUT /persons/<uuid>/basic request: the basic
// fields plus the optimistic-lock token and the target uuid.
type basicUpdateBody struct {
	BasicFields
	Timestamp string `json:"timestamp"`
	UUID      string `json:"uuid"`
}

// createResponse is the 201 body of POST /persons.
type createResponse struct {
	Basic  BasicRecord `json:"basic"`
	Events []Event     `json:"events"`
}

// CreatePersonInput is the caller-facing create payload: the basic fields plus
// the life events to attach. Exactly one birth event is required; build them
// with SelfBirthEvent / SelfDeathEvent.
type CreatePersonInput struct {
	Basic  BasicFields
	Events []Event
}

// equalDate wraps an optional DatePart in a single-valued gregorian complex date.
func equalDate(first *DatePart) EventDate {
	if first != nil && first.Type == "" {
		first.Type = calendarGregorian
	}
	return EventDate{Calendar: calendarGregorian, Type: dateTypeEqual, First: first}
}

// SelfBirthEvent builds the mandatory birth event for the person being created.
// Pass nil for an unknown date.
func SelfBirthEvent(date *DatePart) Event {
	return Event{
		Type:         EventBirth,
		Date:         equalDate(date),
		Participants: []Participant{{PersonUUID: SelfRef, Role: RoleChild}},
	}
}

// SelfDeathEvent builds an optional death event for the person being created.
func SelfDeathEvent(date *DatePart) Event {
	return Event{
		Type:         EventDeath,
		Date:         equalDate(date),
		Participants: []Participant{{PersonUUID: SelfRef, Role: RoleOwner}},
	}
}

// CreatePerson mints a new tree person (POST /api/v2/persons), owned by the
// authenticated account. Returns the server's basic record (incl. the new uuid)
// and the events it created.
func (c *Client) CreatePerson(ctx context.Context, in CreatePersonInput) (*createResponse, error) {
	// bearerToken also populates userUUID, used as ?owner=.
	if _, err := c.bearerToken(ctx); err != nil {
		return nil, err
	}
	query := url.Values{}
	if c.userUUID != "" {
		query.Set("owner", c.userUUID)
	}

	body := personCreateBody{Basic: in.Basic, Events: in.Events}
	req, err := c.newAuthedRequest(ctx, http.MethodPost, "persons", query, body)
	if err != nil {
		return nil, err
	}
	var resp createResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPersonBasic reads the editable basic fields (+ timestamps) of a person.
// ErrNotFound ⇒ the person is gone; the Read path drops it from state.
func (c *Client) GetPersonBasic(ctx context.Context, uuid string) (*BasicRecord, error) {
	req, err := c.newAuthedRequest(ctx, http.MethodGet, "persons/"+uuid+"/basic", nil, nil)
	if err != nil {
		return nil, err
	}
	var rec BasicRecord
	if err := c.do(req, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// GetPersonEvents reads a person's life events (for birth/death dates).
func (c *Client) GetPersonEvents(ctx context.Context, uuid string) ([]Event, error) {
	req, err := c.newAuthedRequest(ctx, http.MethodGet, "persons/"+uuid+"/events", nil, nil)
	if err != nil {
		return nil, err
	}
	var events []Event
	if err := c.do(req, &events); err != nil {
		return nil, err
	}
	return events, nil
}

// GetPersonDisplay reads the computed display name from the regularPerson view.
func (c *Client) GetPersonDisplay(ctx context.Context, uuid string) (*personDisplay, error) {
	req, err := c.newAuthedRequest(ctx, http.MethodGet, "persons/"+uuid, nil, nil)
	if err != nil {
		return nil, err
	}
	var d personDisplay
	if err := c.do(req, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// UpdatePersonBasic edits a person's basic fields. timestamp is the optimistic-
// lock token (the updatedAt last read from GetPersonBasic); a stale value is
// rejected with HTTP 400. Returns the refreshed basic record.
func (c *Client) UpdatePersonBasic(ctx context.Context, uuid string, fields BasicFields, timestamp string) (*BasicRecord, error) {
	body := basicUpdateBody{BasicFields: fields, Timestamp: timestamp, UUID: uuid}
	req, err := c.newAuthedRequest(ctx, http.MethodPut, "persons/"+uuid+"/basic", nil, body)
	if err != nil {
		return nil, err
	}
	var rec BasicRecord
	if err := c.do(req, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// DeletePerson removes a tree person (DELETE /api/v2/persons/<uuid>).
func (c *Client) DeletePerson(ctx context.Context, uuid string) error {
	req, err := c.newAuthedRequest(ctx, http.MethodDelete, "persons/"+uuid, nil, nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}
