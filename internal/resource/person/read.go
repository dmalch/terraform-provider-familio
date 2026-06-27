package person

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
)

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuid := state.UUID.ValueString()

	basic, err := r.client.GetPersonBasic(ctx, uuid)
	if err != nil {
		if errors.Is(err, familio.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading familio_person", err.Error())
		return
	}
	applyBasicToState(basic, &state)

	// Display name lives on the regularPerson view, not /basic.
	if display, err := r.client.GetPersonDisplay(ctx, uuid); err != nil {
		resp.Diagnostics.AddWarning("Could not read familio_person display name", err.Error())
	} else {
		state.DisplayName = types.StringValue(display.DisplayName)
	}

	// Birth/death dates and parents come from the events sub-resource.
	if events, err := r.client.GetPersonEvents(ctx, uuid); err != nil {
		resp.Diagnostics.AddWarning("Could not read familio_person events", err.Error())
	} else {
		applyEventsToState(events, &state)
		resp.Diagnostics.Append(applyParentsToState(ctx, events, &state)...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
