package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
)

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	uuid := data.UUID.ValueString()

	regular, err := d.client.GetPersonRegular(ctx, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Error reading familio_person data source", err.Error())
		return
	}
	basic, err := d.client.GetPersonBasic(ctx, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Error reading familio_person basic fields", err.Error())
		return
	}
	events, err := d.client.GetPersonEvents(ctx, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Error reading familio_person events", err.Error())
		return
	}

	data.OwnerID = stringOrNull(regular.OwnerID)
	data.DisplayName = types.StringValue(regular.DisplayName)
	data.FirstName = types.StringValue(basic.FirstName)
	data.LastName = types.StringValue(basic.LastName)
	data.Patronymic = types.StringValue(basic.MiddleName)
	data.BirthFirstName = types.StringValue(basic.BirthFirstName)
	data.BirthLastName = types.StringValue(basic.BirthLastName)
	data.Gender = types.StringValue(basic.Gender)
	data.Privacy = types.StringValue(basic.Privacy)

	data.BirthDate = ownBirthFormatted(events, uuid)
	data.DeathDate = firstFormattedOfType(events, familio.EventDeath)
	data.ChristeningDate = firstFormattedOfType(events, familio.EventBaptism)

	var parents []string
	if birth := familio.OwnBirthEvent(events, uuid); birth != nil {
		parents = birth.ParentUUIDs()
	}
	resp.Diagnostics.Append(setStrings(ctx, &data.Parents, parents)...)
	resp.Diagnostics.Append(setStrings(ctx, &data.Spouses, familio.SpousesOf(events, uuid))...)
	resp.Diagnostics.Append(setStrings(ctx, &data.Children, familio.ChildrenOf(events, uuid))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ownBirthFormatted returns the formatted date of the person's own birth event.
func ownBirthFormatted(events []familio.Event, uuid string) types.String {
	birth := familio.OwnBirthEvent(events, uuid)
	if birth == nil {
		return types.StringNull()
	}
	return stringOrNull(birth.Date.Formatted)
}

// firstFormattedOfType returns the formatted date of the first event of typ.
func firstFormattedOfType(events []familio.Event, typ string) types.String {
	for i := range events {
		if events[i].Type == typ {
			return stringOrNull(events[i].Date.Formatted)
		}
	}
	return types.StringNull()
}

func setStrings(ctx context.Context, dst *types.Set, values []string) diag.Diagnostics {
	set, diags := types.SetValueFrom(ctx, types.StringType, values)
	if !diags.HasError() {
		*dst = set
	}
	return diags
}

func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
