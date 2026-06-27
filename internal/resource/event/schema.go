package event

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
	"github.com/dmalch/terraform-provider-familio/internal/tfdate"
)

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A single-subject life-fact event on a familio.org person — the long tail of " +
			"familio's event catalogue (residence, education, occupation, military service, awards, " +
			"emigration, …). Birth, death and christening are managed on familio_person, and marriages " +
			"by familio_marriage; this resource covers the rest. Changing any attribute forces " +
			"replacement (familio has no in-place event edit).",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Description:   "The familio.org event UUID.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"person": schema.StringAttribute{
				Description:   "UUID of the person this event belongs to. Must already exist.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"type": schema.StringAttribute{
				Description: "Event type. One of familio's single-subject fact types, e.g. location " +
					"(residence), profession, education, militaryService, militaryAward, award, " +
					"citizenship, emigration, immigration, arrest, burial, godparent (Восприемник), " +
					"warranter (Поручитель). Per familio's own model, godparent/warranter are recorded " +
					"on the godparent/witness; familio does not link them to the godchild/party, so name " +
					"that person in `comment`. (birth/death/baptism are on familio_person; " +
					"wedding/divorce are relationships and not allowed here.)",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(familio.FactEventTypes...),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"date": tfdate.Block("Event date. May be approximate (circa), bounded (range = "+
				"before/after) or a span (range = between with end_year/…).", true),
			"comment": schema.StringAttribute{
				Description:   "Free-text note (familio has no type-specific fields; details like an award name or occupation go here).",
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},

			"created_at": schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
			"updated_at": schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
		},
	}
}
