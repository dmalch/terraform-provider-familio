package settlementpersons

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Model is the data source state.
type Model struct {
	Settlement types.String `tfsdk:"settlement"`
	CatalogKey types.String `tfsdk:"catalog_key"`
	Persons    types.List   `tfsdk:"persons"`
}

// PersonModel is one element of the persons list.
type PersonModel struct {
	UUID                types.String `tfsdk:"uuid"`
	DisplayName         types.String `tfsdk:"display_name"`
	ShortDisplayName    types.String `tfsdk:"short_display_name"`
	CatalogKey          types.String `tfsdk:"catalog_key"`
	CatalogName         types.String `tfsdk:"catalog_name"`
	Type                types.String `tfsdk:"type"`
	BirthDate           types.String `tfsdk:"birth_date"`
	DeathDate           types.String `tfsdk:"death_date"`
	HasDeathEvent       types.Bool   `tfsdk:"has_death_event"`
	BirthSettlementText types.String `tfsdk:"birth_settlement_text"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

func personObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"uuid":                  types.StringType,
			"display_name":          types.StringType,
			"short_display_name":    types.StringType,
			"catalog_key":           types.StringType,
			"catalog_name":          types.StringType,
			"type":                  types.StringType,
			"birth_date":            types.StringType,
			"death_date":            types.StringType,
			"has_death_event":       types.BoolType,
			"birth_settlement_text": types.StringType,
			"updated_at":            types.StringType,
		},
	}
}
