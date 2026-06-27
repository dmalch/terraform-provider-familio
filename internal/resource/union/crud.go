package union

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := inputFromModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateUnion(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Cannot create familio_union", err.Error())
		return
	}

	plan.UUID = types.StringValue(created.UUID)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// No union read endpoint is known yet; pass state through unchanged. Once a
	// union enters state (after write support lands), this fetches and reconciles.
	var state ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := inputFromModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.UpdateUnion(ctx, plan.UUID.ValueString(), input); err != nil {
		resp.Diagnostics.AddError("Cannot update familio_union", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteUnion(ctx, state.UUID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Cannot delete familio_union", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func inputFromModel(ctx context.Context, m *ResourceModel) (familio.UnionInput, diag.Diagnostics) {
	var input familio.UnionInput
	var diags diag.Diagnostics

	if !m.Partners.IsNull() && !m.Partners.IsUnknown() {
		diags = append(diags, m.Partners.ElementsAs(ctx, &input.PartnerUUIDs, false)...)
	}
	if !m.Children.IsNull() && !m.Children.IsUnknown() {
		diags = append(diags, m.Children.ElementsAs(ctx, &input.ChildUUIDs, false)...)
	}
	input.MarriageDate = m.MarriageDate.ValueString()
	input.DivorceDate = m.DivorceDate.ValueString()

	return input, diags
}
