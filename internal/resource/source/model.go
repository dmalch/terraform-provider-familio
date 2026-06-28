package source

import "github.com/hashicorp/terraform-plugin-framework/types"

// ResourceModel is the familio_source state: a source citation attached to a
// person. The reference (reference_uuid + type + catalog_key) is immutable
// (RequiresReplace); comment edits in place. name/requisites/years/catalog and
// the timestamps are server-derived. See internal/tfsource and API.md.
type ResourceModel struct {
	Person        types.String `tfsdk:"person"`
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
