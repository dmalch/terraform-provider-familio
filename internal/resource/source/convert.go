package source

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/go-familio"
	"github.com/dmalch/terraform-provider-familio/internal/tfsource"
)

// refFromModel builds the familio create body from the configuration fields.
func refFromModel(m *ResourceModel) familio.SourceRef {
	return tfsource.Ref(m.ReferenceUUID.ValueString(), m.Type.ValueString(), m.CatalogKey)
}

// applySourceToState writes the server-derived fields of a read-back source into
// the model. The reference fields (person/reference_uuid/type/catalog_key) are
// left as-is — reference_uuid is the path identity and catalog_key is write-only,
// so both are carried from prior plan/state, not from the API.
func applySourceToState(s *familio.Source, m *ResourceModel) {
	m.ReferenceUUID = types.StringValue(s.UUID)
	m.Type = types.StringValue(s.Type)
	m.Comment = strOrNull(s.Comment)
	m.Name = strOrNull(s.Name)
	m.Requisites = strOrNull(s.Requisites)
	m.Years = strOrNull(s.Years)
	m.Catalog = strOrNull(s.Catalog.String())
	m.CreatedAt = types.StringValue(s.CreatedAt)
	m.UpdatedAt = types.StringValue(s.UpdatedAt)
}

// strOrNull maps a server-empty string to a null attribute so an omitted
// optional value round-trips without a permadiff.
func strOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
