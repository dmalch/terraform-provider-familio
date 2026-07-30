package tags

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The authenticated account's tag catalogue («Мои метки»). Use it to reference " +
			"tags that already exist on familio without importing them as familio_tag resources: " +
			"`by_name` maps a tag's text to the integer id that familio_person's `tags` attribute " +
			"takes.\n\n" +
			"Tags are a **Familio Plus** feature; on a free account only the tag whose `is_free` " +
			"is true can actually be used, though the others are still listed here.",
		Attributes: map[string]schema.Attribute{
			"by_name": schema.MapAttribute{
				Computed:    true,
				ElementType: types.Int64Type,
				Description: "Tag id keyed by tag text — the lookup this data source exists for, e.g. " +
					"`data.familio_tags.all.by_name[\"Проверить в архиве\"]`. familio keeps tag names " +
					"unique per account, so the mapping is unambiguous.",
			},
			"ids": schema.SetAttribute{
				Computed:    true,
				ElementType: types.Int64Type,
				Description: "Every tag id in the catalogue.",
			},
			"tags": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The full catalogue, ordered as familio returns it.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.Int64Attribute{Computed: true, Description: "The tag's integer id."},
						"name":        schema.StringAttribute{Computed: true, Description: "The tag text («Название»)."},
						"color":       schema.StringAttribute{Computed: true, Description: "Palette colour code, e.g. `mint-mist`."},
						"description": schema.StringAttribute{Computed: true, Description: "Free-text description, null when unset."},
						"hex":         schema.StringAttribute{Computed: true, Description: "The fill the web UI paints `color` with, derived locally."},
						"is_free": schema.BoolAttribute{Computed: true, Description: "Whether the tag is usable " +
							"without a Familio Plus subscription."},
					},
				},
			},
		},
	}
}
