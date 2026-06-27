package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuid := plan.UUID.ValueString()

	// familio's update is optimistic-locked: it needs the current updatedAt as
	// a "timestamp" token. Fetch the freshest one before writing.
	current, err := r.client.GetPersonBasic(ctx, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Cannot read familio_person before update", err.Error())
		return
	}

	updated, err := r.client.UpdatePersonBasic(ctx, uuid, basicFromModel(&plan), current.UpdatedAt)
	if err != nil {
		resp.Diagnostics.AddError("Cannot update familio_person", err.Error())
		return
	}
	applyBasicToState(updated, &plan)

	if display, err := r.client.GetPersonDisplay(ctx, uuid); err != nil {
		resp.Diagnostics.AddWarning("Could not read familio_person display name", err.Error())
	} else {
		plan.DisplayName = types.StringValue(display.DisplayName)
	}

	// birth_date / death_date are RequiresReplace, so they are unchanged here;
	// keep the planned (config) values already on `plan`.
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
