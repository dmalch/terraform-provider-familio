package tag

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/dmalch/go-familio"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateTag(ctx, inputFromModel(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Cannot create familio_tag", err.Error())
		return
	}

	applyTagToState(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read finds the tag by id in the account's tag catalogue: familio exposes no
// GET for a single tag, only the owner-addressed list.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueInt64()
	if id <= 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	tags, err := r.client.ListTags(ctx)
	if err != nil {
		if errors.Is(err, familio.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading familio_tag", err.Error())
		return
	}

	found := findByID(tags, int(id))
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	applyTagToState(found, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update rewrites all three writable fields — familio's PUT replaces them, so
// an emptied description clears it. There is no optimistic-lock token on this
// endpoint, so a concurrent edit is silently overwritten rather than rejected.
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The id is Computed and carried through by UseStateForUnknown, but a plan
	// built from an interrupted apply may still leave it unknown.
	id := plan.ID.ValueInt64()
	if id <= 0 {
		id = state.ID.ValueInt64()
	}

	updated, err := r.client.UpdateTag(ctx, int(id), inputFromModel(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Cannot update familio_tag", err.Error())
		return
	}

	applyTagToState(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete removes the tag from the catalogue. familio unassigns it from every
// person it was attached to as a side effect — including persons this
// configuration does not manage.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteTag(ctx, int(state.ID.ValueInt64()))
	if err != nil && !errors.Is(err, familio.ErrNotFound) {
		resp.Diagnostics.AddError("Cannot delete familio_tag", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}
