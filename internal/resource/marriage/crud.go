package marriage

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
	"github.com/dmalch/terraform-provider-familio/internal/tfdate"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	partners, diags := partnerList(ctx, plan.Partners)
	resp.Diagnostics.Append(diags...)
	date, dd := tfdate.PartFromObject(ctx, plan.MarriageDate)
	resp.Diagnostics.Append(dd...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(partners) != 2 {
		resp.Diagnostics.AddError("Invalid familio_marriage", "partners must contain exactly two person UUIDs")
		return
	}

	event := familio.WeddingEvent(date, partners[0], partners[1])
	created, err := r.client.CreateEvent(ctx, partners[0], event)
	if err != nil {
		resp.Diagnostics.AddError("Cannot create familio_marriage", err.Error())
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

	partners, diags := partnerList(ctx, state.Partners)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(partners) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}
	anchor := partners[0]

	events, err := r.client.GetPersonEvents(ctx, anchor)
	if err != nil {
		if errors.Is(err, familio.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading familio_marriage", err.Error())
		return
	}

	event := findWedding(events, state.UUID.ValueString())
	if event == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	spouses := event.SpouseUUIDs()
	partnerSet, d := types.SetValueFrom(ctx, types.StringType, spouses)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Partners = partnerSet
	state.MarriageDate = tfdate.Object(event.Date.First)
	state.CreatedAt = types.StringValue(event.CreatedAt)
	state.UpdatedAt = types.StringValue(event.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update only ever runs for in-place-updatable attributes. partners and
// marriage_date both force replacement, so nothing here calls the API; it just
// carries the computed values forward.
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

	partners, diags := partnerList(ctx, state.Partners)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(partners) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	err := r.client.DeleteEvent(ctx, partners[0], state.UUID.ValueString())
	if err != nil && !errors.Is(err, familio.ErrNotFound) {
		resp.Diagnostics.AddError("Cannot delete familio_marriage", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}
