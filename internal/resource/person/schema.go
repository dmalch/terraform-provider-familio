package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
)

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A person in a familio.org family tree (create, read, update, delete). " +
			"Relationships (parents, spouses) are modelled as events and are not yet managed " +
			"by this resource — see the provider documentation.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Description:   "The familio.org person UUID.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"first_name": schema.StringAttribute{
				Description: "Given name (имя).",
				Optional:    true,
				Computed:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "Surname (фамилия). NOTE: familio normalises capitalisation server-side.",
				Optional:    true,
				Computed:    true,
			},
			"patronymic": schema.StringAttribute{
				Description: "Patronymic (отчество); familio's middleName.",
				Optional:    true,
				Computed:    true,
			},
			"birth_first_name": schema.StringAttribute{
				Description: "Given name at birth (maiden), if different.",
				Optional:    true,
				Computed:    true,
			},
			"birth_last_name": schema.StringAttribute{
				Description: "Surname at birth (maiden), if different.",
				Optional:    true,
				Computed:    true,
			},
			"gender": schema.StringAttribute{
				Description: "Gender. One of: male, female.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(familio.GenderMale, familio.GenderFemale),
				},
			},
			"privacy": schema.StringAttribute{
				Description: "Privacy. One of: visible_for_all, invisible. Defaults to visible_for_all.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(familio.PrivacyVisibleForAll, familio.PrivacyInvisible),
				},
			},
			"birth_date": dateBlock("Birth date."),
			"death_date": dateBlock("Death date. Setting it records a death event."),

			// Computed, populated from familio.
			"display_name": schema.StringAttribute{Computed: true, Description: "Server-computed full display name."},
			"created_at":   schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
			"updated_at":   schema.StringAttribute{Computed: true, Description: "Last update timestamp."},
		},
	}
}

// dateBlock builds a nested {year, month, day} date attribute. Changing a date
// forces replacement: editing existing events is not yet implemented, so the
// only way to change a birth/death date today is to recreate the person.
func dateBlock(desc string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: desc + " Changing it forces a new resource (event editing is not yet supported).",
		Optional:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
		Attributes: map[string]schema.Attribute{
			"year": schema.Int64Attribute{
				Description: "Year (e.g. 1900).",
				Required:    true,
			},
			"month": schema.Int64Attribute{
				Description: "Month, 1-12.",
				Optional:    true,
				Validators:  []validator.Int64{int64validator.Between(1, 12)},
			},
			"day": schema.Int64Attribute{
				Description: "Day of month, 1-31.",
				Optional:    true,
				Validators:  []validator.Int64{int64validator.Between(1, 31)},
			},
		},
	}
}
