package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A person in a familio.org family tree. " +
			"NOTE: creating/updating/deleting is not yet supported (the Familio write " +
			"API is still being reverse-engineered); only Read and import work today.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Description:   "The familio.org person UUID.",
				Computed:      true,
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"first_name": schema.StringAttribute{
				Description: "Given name (имя).",
				Optional:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "Surname (фамилия).",
				Optional:    true,
			},
			"patronymic": schema.StringAttribute{
				Description: "Patronymic (отчество).",
				Optional:    true,
			},
			"gender": schema.StringAttribute{
				Description: "Gender. One of: male, female.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("male", "female"),
				},
			},
			"birth_settlement": schema.StringAttribute{
				Description: "Birth settlement UUID.",
				Optional:    true,
			},
			"father_uuid": schema.StringAttribute{
				Description: "UUID of the father person.",
				Optional:    true,
			},
			"mother_uuid": schema.StringAttribute{
				Description: "UUID of the mother person.",
				Optional:    true,
			},

			// Computed, populated from the read endpoint.
			"display_name":       schema.StringAttribute{Computed: true, Description: "Full display name."},
			"short_display_name": schema.StringAttribute{Computed: true, Description: "Abbreviated display name."},
			"birth_date":         schema.StringAttribute{Computed: true, Description: "Birth date as returned by familio.org."},
			"death_date":         schema.StringAttribute{Computed: true, Description: "Death date as returned by familio.org."},
			"has_death_event":    schema.BoolAttribute{Computed: true, Description: "Whether a death event is recorded."},
			"catalog_key":        schema.StringAttribute{Computed: true, Description: "Source catalog slug (null for user-created persons)."},
			"catalog_name":       schema.StringAttribute{Computed: true, Description: "Source catalog name."},
			"type":               schema.StringAttribute{Computed: true, Description: "Familio person type."},
			"updated_at":         schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
		},
	}
}
