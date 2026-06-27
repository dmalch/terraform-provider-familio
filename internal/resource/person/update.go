package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdatePerson(ctx, plan.UUID.ValueString(), inputFromModel(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Cannot update familio_person", err.Error())
		return
	}

	applyToState(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
