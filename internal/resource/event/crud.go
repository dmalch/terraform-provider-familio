package event

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/go-familio"
	"github.com/dmalch/terraform-provider-familio/internal/tfdate"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	date, d := tfdate.RangeFromObject(ctx, plan.Date)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	ev := familio.FactEvent(plan.Type.ValueString(), date, plan.Person.ValueString(), plan.Comment.ValueString())
	created, err := r.client.CreateEvent(ctx, plan.Person.ValueString(), ev)
	if err != nil {
		resp.Diagnostics.AddError("Cannot create familio_event", err.Error())
		return
	}

	plan.UUID = types.StringValue(created.ID())
	plan.CreatedAt = types.StringValue(created.CreatedAt)
	plan.UpdatedAt = types.StringValue(created.UpdatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	person := state.Person.ValueString()
	if person == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	events, err := r.client.GetPersonEvents(ctx, person)
	if err != nil {
		if errors.Is(err, familio.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading familio_event", err.Error())
		return
	}

	ev := familio.FindByID(events, state.UUID.ValueString())
	if ev == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Type = types.StringValue(ev.Type)
	state.Date = tfdate.ObjectFromRange(familio.RangeFromEventDate(ev.Date))
	if ev.Comment == "" {
		state.Comment = types.StringNull()
	} else {
		state.Comment = types.StringValue(ev.Comment)
	}
	state.CreatedAt = types.StringValue(ev.CreatedAt)
	state.UpdatedAt = types.StringValue(ev.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update only ever runs for in-place-updatable attributes. Every familio_event
// attribute forces replacement, so this just carries the computed values forward.
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.UUID = state.UUID
	plan.CreatedAt = state.CreatedAt
	plan.UpdatedAt = state.UpdatedAt
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.DeleteEvent(ctx, state.Person.ValueString(), state.UUID.ValueString())
	if err != nil && !errors.Is(err, familio.ErrNotFound) {
		resp.Diagnostics.AddError("Cannot delete familio_event", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}
