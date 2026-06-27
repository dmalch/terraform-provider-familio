package settlementpersons

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "All persons (catalog-sourced and user-created) linked to a familio.org " +
			"settlement, via the public /api/v2/persons endpoint.",
		Attributes: map[string]schema.Attribute{
			"settlement": schema.StringAttribute{
				Description: "The familio.org settlement UUID to list persons for.",
				Required:    true,
			},
			"catalog_key": schema.StringAttribute{
				Description: "Optional client-side filter: keep only persons whose catalogKey " +
					"equals this value (familio has no server-side catalog facet). Omit to " +
					"return all persons for the settlement.",
				Optional: true,
			},
			"persons": schema.ListNestedAttribute{
				Description: "The persons linked to the settlement.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid":                  schema.StringAttribute{Computed: true},
						"display_name":          schema.StringAttribute{Computed: true},
						"short_display_name":    schema.StringAttribute{Computed: true},
						"catalog_key":           schema.StringAttribute{Computed: true},
						"catalog_name":          schema.StringAttribute{Computed: true},
						"type":                  schema.StringAttribute{Computed: true},
						"birth_date":            schema.StringAttribute{Computed: true},
						"death_date":            schema.StringAttribute{Computed: true},
						"has_death_event":       schema.BoolAttribute{Computed: true},
						"birth_settlement_text": schema.StringAttribute{Computed: true},
						"updated_at":            schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}
