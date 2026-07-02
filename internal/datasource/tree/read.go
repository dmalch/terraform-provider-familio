package tree

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/go-familio"
)

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodes, err := d.client.CrawlTree(ctx, data.Root.ValueString(), familio.TreeOptions{
		Direction: data.Direction.ValueString(),
		Surname:   data.Surname.ValueString(),
		Depth:     int(data.Depth.ValueInt64()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error crawling familio_tree", err.Error())
		return
	}

	models := make([]NodeModel, 0, len(nodes))
	for i := range nodes {
		m, diags := toNodeModel(ctx, nodes[i])
		resp.Diagnostics.Append(diags...)
		models = append(models, m)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	list, diags := types.ListValueFrom(ctx, nodeObjectType(), models)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Nodes = list

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func toNodeModel(ctx context.Context, n familio.TreeNode) (NodeModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	parents, d := refList(ctx, n.Parents)
	diags.Append(d...)
	children, d := refList(ctx, n.Children)
	diags.Append(d...)
	spouses, d := spouseList(ctx, n.Spouses)
	diags.Append(d...)

	return NodeModel{
		UUID:     types.StringValue(n.UUID),
		Name:     stringOrNull(n.Name),
		Year:     yearOrNull(n.Year),
		Parents:  parents,
		Spouses:  spouses,
		Children: children,
	}, diags
}

func refList(ctx context.Context, refs []familio.PersonRef) (types.List, diag.Diagnostics) {
	models := make([]RefModel, 0, len(refs))
	for _, r := range refs {
		models = append(models, RefModel{UUID: types.StringValue(r.UUID), Name: stringOrNull(r.Name)})
	}
	return types.ListValueFrom(ctx, refObjectType(), models)
}

func spouseList(ctx context.Context, spouses []familio.Spouse) (types.List, diag.Diagnostics) {
	models := make([]SpouseModel, 0, len(spouses))
	for _, s := range spouses {
		models = append(models, SpouseModel{
			UUID:         types.StringValue(s.UUID),
			Name:         stringOrNull(s.Name),
			MarriageUUID: stringOrNull(s.MarriageUUID),
		})
	}
	return types.ListValueFrom(ctx, spouseObjectType(), models)
}

func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func yearOrNull(y int) types.Int64 {
	if y == 0 {
		return types.Int64Null()
	}
	return types.Int64Value(int64(y))
}
