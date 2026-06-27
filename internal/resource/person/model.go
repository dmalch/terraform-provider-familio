package person

import "github.com/hashicorp/terraform-plugin-framework/types"

// ResourceModel is the familio_person state.
//
// first_name / last_name / patronymic / gender / father_uuid / mother_uuid are
// config-only for now: the public read endpoint returns only display-level
// fields, so Read does not populate them. The write-API spike (Phase 0.5) will
// reveal the editable person-detail fields and wire them through Read.
type ResourceModel struct {
	UUID             types.String `tfsdk:"uuid"`
	FirstName        types.String `tfsdk:"first_name"`
	LastName         types.String `tfsdk:"last_name"`
	Patronymic       types.String `tfsdk:"patronymic"`
	Gender           types.String `tfsdk:"gender"`
	BirthSettlement  types.String `tfsdk:"birth_settlement"`
	FatherUUID       types.String `tfsdk:"father_uuid"`
	MotherUUID       types.String `tfsdk:"mother_uuid"`
	DisplayName      types.String `tfsdk:"display_name"`
	ShortDisplayName types.String `tfsdk:"short_display_name"`
	BirthDate        types.String `tfsdk:"birth_date"`
	DeathDate        types.String `tfsdk:"death_date"`
	HasDeathEvent    types.Bool   `tfsdk:"has_death_event"`
	CatalogKey       types.String `tfsdk:"catalog_key"`
	CatalogName      types.String `tfsdk:"catalog_name"`
	Type             types.String `tfsdk:"type"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}
