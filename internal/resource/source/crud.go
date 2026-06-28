package source

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	person := plan.Person.ValueString()
	created, err := r.client.CreateSource(ctx, person, refFromModel(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Cannot create familio_source", err.Error())
		return
	}

	// The create body carries no comment, so set it in a follow-up PATCH.
	if !plan.Comment.IsNull() && plan.Comment.ValueString() != "" {
		created, err = r.client.UpdateSourceComment(ctx, person, created.UUID, plan.Comment.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Cannot set familio_source comment", err.Error())
			return
		}
	}

	applySourceToState(created, &plan)
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

	sources, err := r.client.GetPersonSources(ctx, person)
	if err != nil {
		if errors.Is(err, familio.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading familio_source", err.Error())
		return
	}

	src := familio.FindSourceByID(sources, state.ReferenceUUID.ValueString())
	if src == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	applySourceToState(src, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update only ever runs for the in-place-editable comment; the reference fields
// force replacement.
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Comment.Equal(state.Comment) {
		updated, err := r.client.UpdateSourceComment(ctx, plan.Person.ValueString(), plan.ReferenceUUID.ValueString(), plan.Comment.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Cannot update familio_source comment", err.Error())
			return
		}
		applySourceToState(updated, &plan)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.client.DeleteSource(ctx, state.Person.ValueString(), state.ReferenceUUID.ValueString())
	if err != nil && !errors.Is(err, familio.ErrNotFound) {
		resp.Diagnostics.AddError("Cannot delete familio_source", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}
