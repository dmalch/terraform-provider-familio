package tag

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/go-familio"
)

// inputFromModel builds the familio create/update body from the configuration.
// A null description is sent as "" — familio has no null for it, and the client
// validates the result before the request.
func inputFromModel(m *ResourceModel) familio.TagInput {
	return familio.TagInput{
		Name:        m.Name.ValueString(),
		Color:       m.Color.ValueString(),
		Description: m.Description.ValueString(),
	}
}

// applyTagToState writes a read-back tag into the model. An empty description
// maps back to null so an omitted optional value round-trips without a
// permadiff, mirroring how familio_source handles its optional strings.
func applyTagToState(t *familio.Tag, m *ResourceModel) {
	m.ID = types.Int64Value(int64(t.ID))
	m.Name = types.StringValue(t.Name)
	m.Color = types.StringValue(t.Color)
	m.Description = descriptionOrNull(t.Description)
	m.Hex = types.StringValue(familio.TagColorHex[t.Color])
	m.IsFree = types.BoolValue(t.IsFree)
}

// descriptionOrNull maps a server-empty description to a null attribute.
func descriptionOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

// findByID returns the tag with the given id from a catalogue listing, or nil.
// familio has no GET for a single tag, so every read goes through the account's
// list.
func findByID(tags []familio.Tag, id int) *familio.Tag {
	for i := range tags {
		if tags[i].ID == id {
			return &tags[i]
		}
	}
	return nil
}
