package person

import "github.com/hashicorp/terraform-plugin-framework/types"

// ResourceModel is the familio_person state. It maps to familio's "basic"
// person fields plus the birth/death/christening life events, each grouped into
// a nested block carrying its date, place, comment (and, for birth, parents).
// Spouses are events too and are managed elsewhere — see the familio_marriage
// resource and internal/familio/API.md.
type ResourceModel struct {
	UUID           types.String `tfsdk:"uuid"`
	FirstName      types.String `tfsdk:"first_name"`
	LastName       types.String `tfsdk:"last_name"`
	Patronymic     types.String `tfsdk:"patronymic"`
	BirthFirstName types.String `tfsdk:"birth_first_name"`
	BirthLastName  types.String `tfsdk:"birth_last_name"`
	Gender         types.String `tfsdk:"gender"`
	Privacy        types.String `tfsdk:"privacy"`
	Birth          types.Object `tfsdk:"birth"`
	Death          types.Object `tfsdk:"death"`
	Christening    types.Object `tfsdk:"christening"`
	Biography      types.String `tfsdk:"biography"`
	Sources        types.List   `tfsdk:"sources"`
	DisplayName    types.String `tfsdk:"display_name"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}
