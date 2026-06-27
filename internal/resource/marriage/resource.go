// Package marriage implements the familio_marriage resource: a marriage between
// two persons, which familio models as a "wedding" event with two spouse
// participants (see internal/familio/API.md). It is an association resource —
// Create posts the event, Read finds it on a partner's event list, Delete
// removes it. Changing the partners or marriage date forces replacement, since
// event editing is not yet implemented.
package marriage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/dmalch/terraform-provider-familio/internal/config"
	"github.com/dmalch/terraform-provider-familio/internal/familio"
)

type Resource struct {
	client *familio.Client
}

func NewMarriageResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_marriage"
}

func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*config.ClientData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *config.ClientData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = data.Client
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}
