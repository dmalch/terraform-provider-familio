package union

import "github.com/hashicorp/terraform-plugin-framework/types"

// ResourceModel is the familio_union state. Dates are plain strings for now;
// structured {year,month,day,circa} event objects arrive with the write-API
// spike, alongside the real linkage model.
type ResourceModel struct {
	UUID         types.String `tfsdk:"uuid"`
	Partners     types.Set    `tfsdk:"partners"`
	Children     types.Set    `tfsdk:"children"`
	MarriageDate types.String `tfsdk:"marriage_date"`
	DivorceDate  types.String `tfsdk:"divorce_date"`
}
