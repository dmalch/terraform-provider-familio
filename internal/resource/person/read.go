package person

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/go-familio"
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

	// Birth/death dates and parents come from the events sub-resource. This is
	// managed data, so a failed read must be a hard error: silently leaving these
	// null would let the next apply overwrite real familio values with no diff (#4).
	events, err := r.client.GetPersonEvents(ctx, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Error reading familio_person events", err.Error())
		return
	}
	resp.Diagnostics.Append(applyEventsToState(ctx, events, &state)...)

	// Biography is its own managed sub-resource; a failed read is a hard error so
	// we never silently blank a real value and let the next apply overwrite it.
	bio, err := r.client.GetPersonBiography(ctx, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Error reading familio_person biography", err.Error())
		return
	}
	state.Biography = types.StringValue(bio.Text)

	// Refresh the sources block only when it is managed (non-null); an omitted
	// block must stay null so the provider doesn't claim a person's sources.
	if !state.Sources.IsNull() {
		sources, d := r.readSources(ctx, uuid, state.Sources)
		resp.Diagnostics.Append(d...)
		state.Sources = sources
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
