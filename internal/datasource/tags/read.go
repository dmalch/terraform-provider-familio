package tags

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/go-familio"
)

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogue, err := d.client.ListTags(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading familio_tags data source", err.Error())
		return
	}

	models := make([]TagModel, 0, len(catalogue))
	byName := make(map[string]int64, len(catalogue))
	ids := make([]int64, 0, len(catalogue))
	for _, t := range catalogue {
		models = append(models, TagModel{
			ID:          types.Int64Value(int64(t.ID)),
			Name:        types.StringValue(t.Name),
			Color:       types.StringValue(t.Color),
			Description: stringOrNull(t.Description),
			Hex:         types.StringValue(familio.TagColorHex[t.Color]),
			IsFree:      types.BoolValue(t.IsFree),
		})
		byName[t.Name] = int64(t.ID)
		ids = append(ids, int64(t.ID))
	}

	list, diags := types.ListValueFrom(ctx, tagElemType, models)
	resp.Diagnostics.Append(diags...)
	nameMap, diags := types.MapValueFrom(ctx, types.Int64Type, byName)
	resp.Diagnostics.Append(diags...)
	idSet, diags := types.SetValueFrom(ctx, types.Int64Type, ids)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Tags = list
	data.ByName = nameMap
	data.IDs = idSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// stringOrNull maps a server-empty string to a null attribute, so an unset
// description reads as null rather than "".
func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
