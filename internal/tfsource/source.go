// Package tfsource bridges familio's source-citation model and a Terraform
// nested source object, shared by the familio_source resource and the
// familio_person `sources` block. A source is an immutable reference
// (reference_uuid + type + catalog_key) plus an editable comment; name,
// requisites, years and catalog are server-derived (computed). The familio wire
// shape lives behind familio.Source / familio.SourceRef (internal/familio/source.go).
package tfsource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
)

// Model is one source object. The first four fields are configuration (the
// reference is fixed; comment edits in place); the rest are computed read-back.
// catalog_key is WRITE-ONLY at the familio API (never echoed on reads), so it is
// preserved from prior config/state rather than refreshed — see ModelFromSource.
type Model struct {
	ReferenceUUID types.String `tfsdk:"reference_uuid"`
	Type          types.String `tfsdk:"type"`
	CatalogKey    types.String `tfsdk:"catalog_key"`
	Comment       types.String `tfsdk:"comment"`
	Name          types.String `tfsdk:"name"`
	Requisites    types.String `tfsdk:"requisites"`
	Years         types.String `tfsdk:"years"`
	Catalog       types.String `tfsdk:"catalog"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

// AttrTypes is the attr-type map for the nested source object (the element type
// of the person `sources` list).
var AttrTypes = map[string]attr.Type{
	"reference_uuid": types.StringType,
	"type":           types.StringType,
	"catalog_key":    types.StringType,
	"comment":        types.StringType,
	"name":           types.StringType,
	"requisites":     types.StringType,
	"years":          types.StringType,
	"catalog":        types.StringType,
	"created_at":     types.StringType,
	"updated_at":     types.StringType,
}

// Ref builds the familio create body from the configuration fields. An
// unset/empty catalog_key marshals as null (correct for a `case`).
func Ref(referenceUUID, typ string, catalogKey types.String) familio.SourceRef {
	return familio.SourceRef{UUID: referenceUUID, Type: typ, CatalogKey: catalogKeyPtr(catalogKey)}
}

// ModelFromSource maps a read-back source into a Model, carrying the write-only
// catalogKey forward from prior config/state (the API never returns it).
func ModelFromSource(s familio.Source, catalogKey types.String) Model {
	return Model{
		ReferenceUUID: types.StringValue(s.UUID),
		Type:          types.StringValue(s.Type),
		CatalogKey:    catalogKey,
		Comment:       strOrNull(s.Comment),
		Name:          strOrNull(s.Name),
		Requisites:    strOrNull(s.Requisites),
		Years:         strOrNull(s.Years),
		Catalog:       strOrNull(s.Catalog.String()),
		CreatedAt:     types.StringValue(s.CreatedAt),
		UpdatedAt:     types.StringValue(s.UpdatedAt),
	}
}

// ObjectFromModel renders a Model as a Terraform object value.
func ObjectFromModel(ctx context.Context, m Model) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, AttrTypes, m)
}

// ModelFromObject decodes a Terraform object value into a Model.
func ModelFromObject(ctx context.Context, obj types.Object) (Model, diag.Diagnostics) {
	var m Model
	diags := obj.As(ctx, &m, basetypes.ObjectAsOptions{})
	return m, diags
}

// catalogKeyPtr converts the optional catalog_key into the *string the wire
// expects: nil (→ JSON null) when unset.
func catalogKeyPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

// strOrNull maps a server-empty string to a null attribute so an omitted
// optional value round-trips without a permadiff.
func strOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
