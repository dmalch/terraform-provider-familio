package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
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

	// The birth event carries its date, parents, place and comment; re-POSTing it
	// upserts the person's single birth event in place (a full replace), so a
	// change to any field of the birth block is applied without recreating the
	// person. They are all re-sent together (the replace would clear an omitted
	// one).
	if !plan.Birth.Equal(state.Birth) {
		bdate, bplace, bcomment, parents, d := birthParts(ctx, plan.Birth)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		ev := familio.BirthEvent(bdate, uuid, parents, bplace, bcomment)
		if _, err := r.client.CreateEvent(ctx, uuid, ev); err != nil {
			resp.Diagnostics.AddError("Cannot update familio_person birth event", err.Error())
			return
		}
	}

	// The death event is optional: upsert it when the block carries any info,
	// delete it when the block is cleared.
	if !plan.Death.Equal(state.Death) {
		r.reconcileDeath(ctx, uuid, plan.Death, resp)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// The christening (baptism) event does not upsert (re-POSTing duplicates it),
	// so reconcile by deleting any existing baptism event and recreating it.
	if !plan.Christening.Equal(state.Christening) {
		r.reconcileChristening(ctx, uuid, plan.Christening, resp)
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

// reconcileDeath upserts the death event when the death block carries any info,
// or deletes the existing death event when the block is cleared.
func (r *Resource) reconcileDeath(ctx context.Context, uuid string, deathBlock types.Object, resp *resource.UpdateResponse) {
	date, place, comment, d := lifeEventParts(ctx, deathBlock)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if hasInfo(date, place, comment) {
		ev := familio.DeathEvent(date, uuid, place, comment)
		if _, err := r.client.CreateEvent(ctx, uuid, ev); err != nil {
			resp.Diagnostics.AddError("Cannot update familio_person death event", err.Error())
		}
		return
	}

	// death block cleared → delete the death event if one exists.
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

// reconcileChristening rewrites the person's baptism event: it deletes any
// existing baptism event(s) (the event is repeatable and does not upsert) and,
// when the christening block carries any info, creates a fresh one.
func (r *Resource) reconcileChristening(ctx context.Context, uuid string, christeningBlock types.Object, resp *resource.UpdateResponse) {
	date, place, comment, d := lifeEventParts(ctx, christeningBlock)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	events, err := r.client.GetPersonEvents(ctx, uuid)
	if err != nil {
		resp.Diagnostics.AddError("Cannot read familio_person events before updating christening", err.Error())
		return
	}
	for i := range events {
		if events[i].Type == familio.EventBaptism {
			if err := r.client.DeleteEvent(ctx, uuid, events[i].ID()); err != nil {
				resp.Diagnostics.AddError("Cannot delete familio_person christening event", err.Error())
				return
			}
		}
	}
	if !hasInfo(date, place, comment) {
		return
	}
	ev := familio.BaptismEvent(date, uuid, place, comment)
	if _, err := r.client.CreateEvent(ctx, uuid, ev); err != nil {
		resp.Diagnostics.AddError("Cannot create familio_person christening event", err.Error())
	}
}
