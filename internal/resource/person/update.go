package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
	"github.com/dmalch/terraform-provider-familio/internal/tfdate"
)

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuid := plan.UUID.ValueString()

	// Only PUT /basic when a basic field actually changed: that endpoint is
	// optimistic-locked and the event upserts below don't need it, so a
	// parents/date-only edit must not touch it.
	if basicChanged(&plan, &state) {
		// familio's basic update needs the current updatedAt as a lock token.
		current, err := r.client.GetPersonBasic(ctx, uuid)
		if err != nil {
			resp.Diagnostics.AddError("Cannot read familio_person before update", err.Error())
			return
		}
		if _, err := r.client.UpdatePersonBasic(ctx, uuid, basicFromModel(&plan), current.UpdatedAt); err != nil {
			resp.Diagnostics.AddError("Cannot update familio_person", err.Error())
			return
		}
	}

	// The birth event carries both the birth date and the parents; re-POSTing it
	// upserts the person's single birth event in place (a full replace), so a
	// change to either is applied without recreating the person.
	if !plan.BirthDate.Equal(state.BirthDate) || !plan.Parents.Equal(state.Parents) {
		birth, d := tfdate.PartFromObject(ctx, plan.BirthDate)
		resp.Diagnostics.Append(d...)
		parents, dp := parentList(ctx, plan.Parents)
		resp.Diagnostics.Append(dp...)
		if resp.Diagnostics.HasError() {
			return
		}
		if _, err := r.client.CreateEvent(ctx, uuid, familio.BirthEvent(birth, uuid, parents)); err != nil {
			resp.Diagnostics.AddError("Cannot update familio_person parents/birth date", err.Error())
			return
		}
	}

	// The death event is optional: upsert it when a date is set, delete it when
	// the date is cleared.
	if !plan.DeathDate.Equal(state.DeathDate) {
		r.reconcileDeath(ctx, uuid, plan.DeathDate, resp)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Refresh server-computed fields (names normalisation, timestamps, display
	// name) after the writes above. Dates/parents already reflect the plan.
	if basic, err := r.client.GetPersonBasic(ctx, uuid); err != nil {
		resp.Diagnostics.AddWarning("Could not re-read familio_person after update", err.Error())
	} else {
		applyBasicToState(basic, &plan)
	}
	if display, err := r.client.GetPersonDisplay(ctx, uuid); err != nil {
		resp.Diagnostics.AddWarning("Could not read familio_person display name", err.Error())
	} else {
		plan.DisplayName = types.StringValue(display.DisplayName)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// basicChanged reports whether any editable basic person field differs between
// the plan and prior state (i.e. whether a PUT /basic is needed at all).
func basicChanged(plan, state *ResourceModel) bool {
	return !plan.FirstName.Equal(state.FirstName) ||
		!plan.LastName.Equal(state.LastName) ||
		!plan.Patronymic.Equal(state.Patronymic) ||
		!plan.BirthFirstName.Equal(state.BirthFirstName) ||
		!plan.BirthLastName.Equal(state.BirthLastName) ||
		!plan.Gender.Equal(state.Gender) ||
		!plan.Privacy.Equal(state.Privacy)
}

// reconcileDeath upserts the death event when deathDate is set, or deletes the
// existing death event when deathDate is cleared.
func (r *Resource) reconcileDeath(ctx context.Context, uuid string, deathDate types.Object, resp *resource.UpdateResponse) {
	if !deathDate.IsNull() && !deathDate.IsUnknown() {
		part, d := tfdate.PartFromObject(ctx, deathDate)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		if _, err := r.client.CreateEvent(ctx, uuid, familio.DeathEvent(part, uuid)); err != nil {
			resp.Diagnostics.AddError("Cannot update familio_person death date", err.Error())
		}
		return
	}

	// death_date removed → delete the death event if one exists.
	events, err := r.client.GetPersonEvents(ctx, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Cannot read familio_person events before clearing death", err.Error())
		return
	}
	for i := range events {
		if events[i].Type == familio.EventDeath {
			if err := r.client.DeleteEvent(ctx, uuid, events[i].ID()); err != nil {
				resp.Diagnostics.AddError("Cannot delete familio_person death event", err.Error())
			}
			return
		}
	}
}
