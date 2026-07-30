package tag

import "github.com/hashicorp/terraform-plugin-framework/types"

// ResourceModel is the familio_tag state. ID is an Int64 rather than the usual
// uuid string because familio keys tags by a small sequential integer. Hex is
// derived locally from Color; IsFree is server-computed.
type ResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Color       types.String `tfsdk:"color"`
	Description types.String `tfsdk:"description"`
	Hex         types.String `tfsdk:"hex"`
	IsFree      types.Bool   `tfsdk:"is_free"`
}
