package union

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A union (marriage/partnership) linking persons in a familio.org family tree. " +
			"NOTE: not yet creatable — the Familio write API is still being reverse-engineered.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Description:   "The familio.org union UUID.",
				Computed:      true,
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"partners": schema.SetAttribute{
				Description: "UUIDs of the partner persons in the union.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"children": schema.SetAttribute{
				Description: "UUIDs of the children of the union.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"marriage_date": schema.StringAttribute{
				Description: "Marriage date.",
				Optional:    true,
			},
			"divorce_date": schema.StringAttribute{
				Description: "Divorce date.",
				Optional:    true,
			},
		},
	}
}
