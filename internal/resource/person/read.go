package person

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
)

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	person, err := r.client.GetPerson(ctx, state.UUID.ValueString())
	if err != nil {
		if errors.Is(err, familio.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading familio_person", err.Error())
		return
	}

	applyToState(person, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
