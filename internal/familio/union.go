package familio

import "context"

// Union is a marriage/partnership linking tree persons. Field shape is a
// placeholder until the Phase 0.5 write-API discovery spike confirms it.
type Union struct {
	UUID         string   `json:"uuid"`
	PartnerUUIDs []string `json:"partnerUuids"`
	ChildUUIDs   []string `json:"childUuids"`
	MarriageDate string   `json:"marriageDate,omitempty"`
	DivorceDate  string   `json:"divorceDate,omitempty"`
}

// UnionInput is the desired state for a union, sent to create/update.
type UnionInput struct {
	PartnerUUIDs []string `json:"partnerUuids,omitempty"`
	ChildUUIDs   []string `json:"childUuids,omitempty"`
	MarriageDate string   `json:"marriageDate,omitempty"`
	DivorceDate  string   `json:"divorceDate,omitempty"`
}

// GetUnion reads a single union by UUID. Not yet implemented — see API.md.
func (c *Client) GetUnion(_ context.Context, _ string) (*Union, error) {
	return nil, ErrWriteNotImplemented
}

// CreateUnion mints a new union. Not yet implemented — see API.md. The biggest
// open question for the spike is whether Familio links existing persons directly
// into a union or requires a create-then-merge dance like Geni.
func (c *Client) CreateUnion(_ context.Context, _ UnionInput) (*Union, error) {
	return nil, ErrWriteNotImplemented
}

// UpdateUnion edits an existing union. Not yet implemented — see API.md.
func (c *Client) UpdateUnion(_ context.Context, _ string, _ UnionInput) (*Union, error) {
	return nil, ErrWriteNotImplemented
}

// DeleteUnion removes a union. Not yet implemented — see API.md.
func (c *Client) DeleteUnion(_ context.Context, _ string) error {
	return ErrWriteNotImplemented
}
