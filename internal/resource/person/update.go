package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/go-familio"
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

	// Life events are preserve-on-omit (issue #22): a block the config does not
	// carry (null) is unmanaged and left untouched on familio, and within a
	// managed block an omitted comment/place/parents is kept by merging from the
	// person's current events. Fetch those once when any managed block changed.
	birthChanged := !plan.Birth.IsNull() && !plan.Birth.Equal(state.Birth)
	deathChanged := !plan.Death.IsNull() && !plan.Death.Equal(state.Death)
	christeningChanged := !plan.Christening.IsNull() && !plan.Christening.Equal(state.Christening)

	var events []familio.Event
	if birthChanged || deathChanged || christeningChanged {
		evs, err := r.client.GetPersonEvents(ctx, uuid)
		if err != nil {
			resp.Diagnostics.AddError("Cannot read familio_person events before update", err.Error())
			return
		}
		events = evs
	}

	// The birth event carries its date, parents, place and comment; re-POSTing it
	// upserts the person's single birth event in place (a full replace). Omitted
	// facets are merged from the current birth event so setting a date does not
	// strip the existing comment/parents.
	if birthChanged {
		bdate, bplace, bcomment, parents, d := birthPartsMerged(ctx, plan.Birth, familio.OwnBirthEvent(events, uuid))
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
	// delete it when a managed block is emptied.
	if deathChanged {
		r.reconcileDeath(ctx, uuid, plan.Death, events, resp)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// The christening (baptism) event does not upsert (re-POSTing duplicates it),
	// so reconcile by deleting any existing baptism event and recreating it.
	if christeningChanged {
		r.reconcileChristening(ctx, uuid, plan.Christening, events, resp)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Biography edits in place via its own sub-resource, which carries its own
	// optimistic-lock version (NOT /basic's) — read it fresh, then PUT.
	if biographyChanged(&plan, &state) {
		current, err := r.client.GetPersonBiography(ctx, uuid)
		if err != nil {
			resp.Diagnostics.AddError("Cannot read familio_person biography before update", err.Error())
			return
		}
		updated, err := r.client.UpdatePersonBiography(ctx, uuid, strValue(plan.Biography), current.UpdatedAt)
		if err != nil {
			resp.Diagnostics.AddError("Cannot update familio_person biography", err.Error())
			return
		}
		plan.Biography = types.StringValue(updated.Text)
	}

	// Sources are an authoritative set when the block is present. Reconcile only
	// on change; a null plan block means "unmanaged" and leaves familio untouched.
	if !plan.Sources.Equal(state.Sources) {
		if desired, managed, d := desiredSources(ctx, plan.Sources); managed {
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			resp.Diagnostics.Append(r.writeSources(ctx, uuid, desired)...)
			if resp.Diagnostics.HasError() {
				return
			}
			sources, d := r.readSources(ctx, uuid, plan.Sources)
			resp.Diagnostics.Append(d...)
			plan.Sources = sources
		}
	}

	// Refresh server-computed fields (names normalisation, timestamps, display
	// name) after the writes above.
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

	// Re-read events so managed life-event blocks reflect the server truth
	// (resolving any preserved/merged facets to concrete values in state). Only
	// managed (non-null) blocks are refreshed; unmanaged ones stay null.
	if birthChanged || deathChanged || christeningChanged {
		if fresh, err := r.client.GetPersonEvents(ctx, uuid); err != nil {
			resp.Diagnostics.AddWarning("Could not re-read familio_person events after update", err.Error())
		} else {
			resp.Diagnostics.Append(applyEventsToState(ctx, fresh, &plan)...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
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

// reconcileDeath upserts the death event when the (managed) death block carries
// any info, or deletes the existing death event when the block is emptied.
// Omitted place/comment are merged from the current death event (events).
func (r *Resource) reconcileDeath(ctx context.Context, uuid string, deathBlock types.Object, events []familio.Event, resp *resource.UpdateResponse) {
	date, place, comment, d := lifeEventPartsMerged(ctx, deathBlock, firstEventOfType(events, familio.EventDeath))
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

	// death block emptied → delete the death event if one exists.
	if ev := firstEventOfType(events, familio.EventDeath); ev != nil {
		if err := r.client.DeleteEvent(ctx, uuid, ev.ID()); err != nil {
			resp.Diagnostics.AddError("Cannot delete familio_person death event", err.Error())
		}
	}
}

// reconcileChristening rewrites the person's baptism event: it deletes any
// existing baptism event(s) (the event is repeatable and does not upsert) and,
// when the (managed) christening block carries any info, creates a fresh one.
// Omitted place/comment are merged from the current baptism event (events).
func (r *Resource) reconcileChristening(ctx context.Context, uuid string, christeningBlock types.Object, events []familio.Event, resp *resource.UpdateResponse) {
	date, place, comment, d := lifeEventPartsMerged(ctx, christeningBlock, firstEventOfType(events, familio.EventBaptism))
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
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
