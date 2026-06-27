package person

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ResourceModel is the familio_person state. It maps to familio's "basic"
// person fields plus the birth/death life events (modelled as nested date
// blocks). Relationships (parents/spouse) are events too and are deferred to a
// dedicated resource — see internal/familio/API.md.
type ResourceModel struct {
	UUID           types.String `tfsdk:"uuid"`
	FirstName      types.String `tfsdk:"first_name"`
	LastName       types.String `tfsdk:"last_name"`
	Patronymic     types.String `tfsdk:"patronymic"`
	BirthFirstName types.String `tfsdk:"birth_first_name"`
	BirthLastName  types.String `tfsdk:"birth_last_name"`
	Gender         types.String `tfsdk:"gender"`
	Privacy        types.String `tfsdk:"privacy"`
	BirthDate      types.Object `tfsdk:"birth_date"`
	DeathDate      types.Object `tfsdk:"death_date"`
	DisplayName    types.String `tfsdk:"display_name"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

// dateModel is one nested birth_date/death_date block.
type dateModel struct {
	Year  types.Int64 `tfsdk:"year"`
	Month types.Int64 `tfsdk:"month"`
	Day   types.Int64 `tfsdk:"day"`
}

// dateAttrTypes is the attr-type map for the nested date object.
var dateAttrTypes = map[string]attr.Type{
	"year":  types.Int64Type,
	"month": types.Int64Type,
	"day":   types.Int64Type,
}
