package event

import "github.com/hashicorp/terraform-plugin-framework/types"

// ResourceModel is the familio_event state: a single-subject fact event on a
// person. date/end_date are nested {year, month, day} blocks; supplying end_date
// makes the event a date range ("between").
type ResourceModel struct {
	UUID      types.String `tfsdk:"uuid"`
	Person    types.String `tfsdk:"person"`
	Type      types.String `tfsdk:"type"`
	Date      types.Object `tfsdk:"date"`
	EndDate   types.Object `tfsdk:"end_date"`
	Comment   types.String `tfsdk:"comment"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}
