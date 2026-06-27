package marriage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/terraform-provider-familio/internal/tfdate"
)

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A marriage between two persons in a familio.org family tree, modelled as the " +
			"wedding event that links the partners. Changing the partners or the marriage date forces " +
			"replacement (event editing is not yet supported).",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Description:   "The underlying wedding-event UUID.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"partners": schema.SetAttribute{
				Description: "UUIDs of the two partner persons. Both must already exist.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeBetween(2, 2),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"marriage_date": tfdate.Block("Marriage date.", true),

			"created_at": schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
			"updated_at": schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
		},
	}
}
