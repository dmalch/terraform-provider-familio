package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeletePerson(ctx, state.UUID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Cannot delete familio_person", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}
