package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/go-familio"
)

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	events, diags := eventsFromModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreatePerson(ctx, familio.CreatePersonInput{
		Basic:  basicFromModel(&plan),
		Events: events,
	})
	if err != nil {
		resp.Diagnostics.AddError("Cannot create familio_person", err.Error())
		return
	}

	applyBasicToState(&created.Basic, &plan)
	resp.Diagnostics.Append(applyEventsToState(ctx, created.Events, &plan)...)
	plan.DisplayName = types.StringValue(created.Basic.DisplayName)

	// Sources are a separate sub-resource; attach them after the person exists.
	if desired, managed, d := desiredSources(ctx, plan.Sources); managed {
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(r.writeSources(ctx, created.Basic.UUID, desired)...)
		if resp.Diagnostics.HasError() {
			return
		}
		sources, d := r.readSources(ctx, created.Basic.UUID, plan.Sources)
		resp.Diagnostics.Append(d...)
		plan.Sources = sources
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
