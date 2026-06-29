package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/dmalch/go-familio"
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
			// Life events, each grouping its date, place, comment (and parents for
			// birth) into a nested block. See lifeevent.go.
			"birth":       birthBlock(),
			"death":       deathBlock(),
			"christening": christeningBlock(),
			"sources":     sourcesBlock(),

			// Computed, populated from familio.
			"display_name": schema.StringAttribute{Computed: true, Description: "Server-computed full display name."},
			"created_at":   schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
			"updated_at":   schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
		},
	}
}
