package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreatePerson(ctx, inputFromModel(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Cannot create familio_person", err.Error())
		return
	}

	applyToState(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
