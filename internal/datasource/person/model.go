package person

import "github.com/hashicorp/terraform-plugin-framework/types"

// Model is the familio_person data source state. uuid is the only input; the
// rest is read from familio. Dates are the server-formatted strings (as on
// familio_settlement_persons); relationships are sets of person uuids, ready to
// feed terraform import or other resources.
type Model struct {
	UUID            types.String `tfsdk:"uuid"`
	OwnerID         types.String `tfsdk:"owner_id"`
	DisplayName     types.String `tfsdk:"display_name"`
	FirstName       types.String `tfsdk:"first_name"`
	LastName        types.String `tfsdk:"last_name"`
	Patronymic      types.String `tfsdk:"patronymic"`
	BirthFirstName  types.String `tfsdk:"birth_first_name"`
	BirthLastName   types.String `tfsdk:"birth_last_name"`
	Gender          types.String `tfsdk:"gender"`
	Privacy         types.String `tfsdk:"privacy"`
	BirthDate       types.String `tfsdk:"birth_date"`
	DeathDate       types.String `tfsdk:"death_date"`
	ChristeningDate types.String `tfsdk:"christening_date"`
	Parents         types.Set    `tfsdk:"parents"`
	Spouses         types.Set    `tfsdk:"spouses"`
	Children        types.Set    `tfsdk:"children"`
}
