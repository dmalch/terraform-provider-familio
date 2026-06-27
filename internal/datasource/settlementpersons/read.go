package settlementpersons

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
)

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	persons, err := d.client.ListSettlementPersons(ctx, data.Settlement.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing settlement persons", err.Error())
		return
	}

	filter := data.CatalogKey.ValueString()
	models := make([]PersonModel, 0, len(persons))
	for _, p := range persons {
		if !data.CatalogKey.IsNull() && catalogKeyOf(p) != filter {
			continue
		}
		models = append(models, toPersonModel(p))
	}

	list, diags := types.ListValueFrom(ctx, personObjectType(), models)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Persons = list

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func toPersonModel(p familio.Person) PersonModel {
	return PersonModel{
		UUID:                types.StringValue(p.UUID),
		DisplayName:         types.StringValue(p.DisplayName),
		ShortDisplayName:    types.StringValue(p.ShortDisplayName),
		CatalogKey:          stringPtrValue(p.CatalogKey),
		CatalogName:         types.StringValue(p.CatalogName),
		Type:                types.StringValue(p.Type),
		BirthDate:           flexDateValue(p.BirthDate),
		DeathDate:           flexDateValue(p.DeathDate),
		HasDeathEvent:       types.BoolValue(p.HasDeathEvent),
		BirthSettlementText: types.StringValue(p.BirthSettlementText),
		UpdatedAt:           types.StringValue(p.UpdatedAt),
	}
}

func catalogKeyOf(p familio.Person) string {
	if p.CatalogKey == nil {
		return ""
	}
	return *p.CatalogKey
}

func stringPtrValue(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

func flexDateValue(d familio.FlexDate) types.String {
	if v, ok := d.Value(); ok {
		return types.StringValue(v)
	}
	return types.StringNull()
}
