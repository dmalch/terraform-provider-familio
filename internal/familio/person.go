package familio

import (
	"context"
	"net/http"
)

// PersonInput is the desired state for a tree person, sent to create/update.
// The exact field names/shape will be confirmed by the Phase 0.5 write-API
// discovery spike; this is a placeholder structure.
type PersonInput struct {
	FirstName       string `json:"firstName,omitempty"`
	LastName        string `json:"lastName,omitempty"`
	Patronymic      string `json:"patronymic,omitempty"`
	Gender          string `json:"gender,omitempty"`
	BirthDate       string `json:"birthDate,omitempty"`
	DeathDate       string `json:"deathDate,omitempty"`
	BirthSettlement string `json:"birthSettlement,omitempty"`
	FatherUUID      string `json:"fatherUuid,omitempty"`
	MotherUUID      string `json:"motherUuid,omitempty"`
}

// GetPerson reads a single tree person by UUID.
//
// TBD-confirm (Phase 0.5): this targets GET /api/v2/persons/<uuid> as a best
// guess. If Familio exposes no single-person route, the spike will replace this
// with the confirmed path (or a settlement-scoped lookup). Read paths handle
// ErrNotFound by dropping the resource from state.
func (c *Client) GetPerson(ctx context.Context, uuid string) (*Person, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "persons/"+uuid, nil, nil)
	if err != nil {
		return nil, err
	}
	var p Person
	if err := c.do(req, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// CreatePerson mints a new tree person. Not yet implemented — see API.md.
func (c *Client) CreatePerson(_ context.Context, _ PersonInput) (*Person, error) {
	return nil, ErrWriteNotImplemented
}

// UpdatePerson edits an existing tree person. Not yet implemented — see API.md.
func (c *Client) UpdatePerson(_ context.Context, _ string, _ PersonInput) (*Person, error) {
	return nil, ErrWriteNotImplemented
}

// DeletePerson removes a tree person. Not yet implemented — see API.md.
func (c *Client) DeletePerson(_ context.Context, _ string) error {
	return ErrWriteNotImplemented
}
