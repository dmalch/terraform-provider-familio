package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
	"github.com/dmalch/terraform-provider-familio/internal/tfdate"
)

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A person in a familio.org family tree (create, read, update, delete). " +
			"A person's parents are managed here via the parents set; spouses are a separate " +
			"resource — see familio_marriage.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Description:   "The familio.org person UUID.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"first_name": schema.StringAttribute{
				Description:   "Given name (имя).",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"last_name": schema.StringAttribute{
				Description:   "Surname (фамилия). NOTE: familio normalises capitalisation server-side.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"patronymic": schema.StringAttribute{
				Description:   "Patronymic (отчество); familio's middleName.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"birth_first_name": schema.StringAttribute{
				Description:   "Given name at birth (maiden), if different.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"birth_last_name": schema.StringAttribute{
				Description:   "Surname at birth (maiden), if different.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"gender": schema.StringAttribute{
				Description: "Gender. One of: male, female.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(familio.GenderMale, familio.GenderFemale),
				},
			},
			"privacy": schema.StringAttribute{
				Description:   "Privacy. One of: visible_for_all, invisible. Defaults to visible_for_all.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Validators: []validator.String{
					stringvalidator.OneOf(familio.PrivacyVisibleForAll, familio.PrivacyInvisible),
				},
			},
			"birth_date": tfdate.Block("Birth date.", false),
			"death_date": tfdate.Block("Death date. Setting it records a death event; "+
				"removing it deletes that event.", false),
			"christening_date": tfdate.Block("Christening (baptism) date — familio's «Крещение» "+
				"event. Setting it records the event; removing it deletes it. Edited in place.", false),

			"birth_place": placeAttribute("Birth place — familio's «Место рождения». The UUID of " +
				"a familio settlement (the same id familio_settlement_persons / the familio_person " +
				"data source speak). Recorded on the birth event; edited in place."),
			"death_place": placeAttribute("Death place — familio's «Место смерти». A familio " +
				"settlement UUID, recorded on the death event. A death_place set without a death_date " +
				"still records the place (on a death event with an unknown date)."),
			"christening_place": placeAttribute("Christening place — the settlement UUID recorded on " +
				"the «Крещение» (baptism) event."),

			"birth_comment":       commentAttribute("Free-text comment (примечание) on the birth event."),
			"death_comment":       commentAttribute("Free-text comment on the death event."),
			"christening_comment": commentAttribute("Free-text comment on the «Крещение» (baptism) event."),

			"parents": schema.SetAttribute{
				Description: "UUIDs of this person's parents (0–2). familio stores them as " +
					"gender-agnostic participants on this person's birth event, so order does not " +
					"matter and a parent's father/mother role is inferred from their own gender. " +
					"Each parent must already exist. Edited in place.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeBetween(0, 2),
				},
			},

			// Computed, populated from familio.
			"display_name": schema.StringAttribute{Computed: true, Description: "Server-computed full display name."},
			"created_at":   schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
			"updated_at":   schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
		},
	}
}

// placeAttribute builds an optional settlement-UUID attribute for a life event's
// place. The provider sends it to familio as the structured {"uuid": …} the API
// requires (a bare uuid string is rejected), and reads the settlement uuid back.
func placeAttribute(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		Description: desc,
		Optional:    true,
	}
}

// commentAttribute builds an optional free-text comment attribute for a life
// event. The comment rides the same event upsert as the date/place.
func commentAttribute(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		Description: desc,
		Optional:    true,
	}
}
