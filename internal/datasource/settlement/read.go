package settlement

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := d.client.GetSettlement(ctx, data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading familio_settlement data source", err.Error())
		return
	}

	data.Name = stringOrNull(s.PrimaryName)
	data.Type = stringOrNull(s.Type)
	data.Status = stringOrNull(s.Status)

	if s.MainGeorequisite != nil {
		data.Region = stringOrNull(s.MainGeorequisite.Level1)
		data.District = stringOrNull(s.MainGeorequisite.Level2)
		if s.MainGeorequisite.Year != 0 {
			data.AsOfYear = types.Int64Value(int64(s.MainGeorequisite.Year))
		} else {
			data.AsOfYear = types.Int64Null()
		}
	} else {
		data.Region = types.StringNull()
		data.District = types.StringNull()
		data.AsOfYear = types.Int64Null()
	}

	if lat, lon, ok := s.Coordinate.LatLon(); ok {
		data.Latitude = types.Float64Value(lat)
		data.Longitude = types.Float64Value(lon)
	} else {
		data.Latitude = types.Float64Null()
		data.Longitude = types.Float64Null()
	}

	names, diags := types.ListValueFrom(ctx, types.StringType, s.AdditionalNames)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.AdditionalNames = names

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
