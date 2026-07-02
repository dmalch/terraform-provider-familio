package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a single familio.org person by UUID, including the owning account " +
			"(owner_id) and the person's parents, spouses and children — useful for discovering " +
			"importable tree nodes and telling your own tree from other researchers' profiles.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Description: "The familio.org person UUID to look up.",
				Required:    true,
			},
			"owner_id": schema.StringAttribute{
				Description: "UUID of the account that owns this profile (null for catalog-sourced " +
					"persons). Filter on this to keep only your own tree.",
				Computed: true,
			},
			"display_name":     schema.StringAttribute{Computed: true, Description: "Server-computed full display name."},
			"first_name":       schema.StringAttribute{Computed: true, Description: "Given name (имя)."},
			"last_name":        schema.StringAttribute{Computed: true, Description: "Surname (фамилия)."},
			"patronymic":       schema.StringAttribute{Computed: true, Description: "Patronymic (отчество)."},
			"birth_first_name": schema.StringAttribute{Computed: true, Description: "Given name at birth (maiden)."},
			"birth_last_name":  schema.StringAttribute{Computed: true, Description: "Surname at birth (maiden)."},
			"gender":           schema.StringAttribute{Computed: true, Description: "Gender (male/female)."},
			"privacy":          schema.StringAttribute{Computed: true, Description: "Privacy (visible_for_all/invisible)."},
			"birth_date":       schema.StringAttribute{Computed: true, Description: "Birth date, as familio formats it (null if unknown)."},
			"death_date":       schema.StringAttribute{Computed: true, Description: "Death date, as familio formats it (null if unknown)."},
			"christening_date": schema.StringAttribute{Computed: true, Description: "Christening (baptism) date, as familio formats it (null if unknown)."},
			"parents": schema.SetAttribute{
				Description: "UUIDs of this person's parents (from their birth event).",
				Computed:    true,
				ElementType: types.StringType,
			},
			"spouses": schema.SetAttribute{
				Description: "UUIDs of this person's spouses (from their wedding events).",
				Computed:    true,
				ElementType: types.StringType,
			},
			"children": schema.SetAttribute{
				Description: "UUIDs of this person's children (persons whose birth event lists this person as a parent).",
				Computed:    true,
				ElementType: types.StringType,
			},
			"marriages": schema.ListNestedAttribute{
				Description: "This person's unions — one per spouse — each carrying the wedding-event " +
					"(marriage) uuid. Pair spouse_uuid (or this person's uuid) with marriage_uuid to " +
					"import a familio_marriage (\"<person_uuid>:<marriage_uuid>\").",
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"spouse_uuid":   schema.StringAttribute{Computed: true, Description: "The spouse's person UUID."},
						"marriage_uuid": schema.StringAttribute{Computed: true, Description: "UUID of the wedding event identifying this marriage."},
					},
				},
			},
		},
	}
}

// marriageObjectType is the element type of the marriages list. It must mirror
// the marriages nested attribute above.
func marriageObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"spouse_uuid":   types.StringType,
		"marriage_uuid": types.StringType,
	}}
}
