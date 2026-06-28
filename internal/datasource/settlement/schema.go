package settlement

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a single familio.org settlement (place) by UUID — the same UUID that " +
			"familio_person's birth/death/christening places and familio_source speak. Returns the " +
			"canonical name, administrative requisites («реквизиты»), classification and coordinates, " +
			"so configs can resolve or validate a settlement reference.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Description: "The familio.org settlement UUID to look up.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Canonical (primary) settlement name.",
			},
			"additional_names": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Alternate/historical names, if any.",
			},
			"region": schema.StringAttribute{
				Computed:    true,
				Description: "Level-1 administrative unit of the main requisite (region/oblast), e.g. «Нижегородская область».",
			},
			"district": schema.StringAttribute{
				Computed:    true,
				Description: "Level-2 administrative unit of the main requisite (district/city), e.g. «город Выкса».",
			},
			"as_of_year": schema.Int64Attribute{
				Computed:    true,
				Description: "Year the main administrative requisite is stated as of.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Settlement kind, e.g. «село», «город», «деревня».",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Settlement status, e.g. «жилой» (inhabited).",
			},
			"latitude": schema.Float64Attribute{
				Computed:    true,
				Description: "Latitude of the settlement (decimal degrees).",
			},
			"longitude": schema.Float64Attribute{
				Computed:    true,
				Description: "Longitude of the settlement (decimal degrees).",
			},
		},
	}
}
