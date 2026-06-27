package union

import "github.com/hashicorp/terraform-plugin-framework/types"

// ResourceModel is the familio_union state. A union is a marriage/partnership,
// which familio models as a "wedding" event with two spouse participants — so
// uuid is the underlying event's uuid. Children are not part of a union (they
// link to parents through their own birth events) and are managed via the
// person resource; divorce events are not yet modelled.
type ResourceModel struct {
	UUID         types.String `tfsdk:"uuid"`
	Partners     types.Set    `tfsdk:"partners"`
	MarriageDate types.Object `tfsdk:"marriage_date"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}
