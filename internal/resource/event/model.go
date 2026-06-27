package event

import "github.com/hashicorp/terraform-plugin-framework/types"

// ResourceModel is the familio_event state: a single-subject fact event on a
// person. date is a nested date block; a date range is expressed within it via
// range = "between" + end_year/end_month/end_day (see internal/tfdate).
type ResourceModel struct {
	UUID      types.String `tfsdk:"uuid"`
	Person    types.String `tfsdk:"person"`
	Type      types.String `tfsdk:"type"`
	Date      types.Object `tfsdk:"date"`
	Comment   types.String `tfsdk:"comment"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}
