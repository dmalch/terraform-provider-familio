// Package person implements the familio_person resource: full create / read /
// update / delete against familio.org's tree-editor API (see
// internal/familio/API.md). Birth/death dates are managed as nested date blocks
// (life events); changing one forces replacement because event editing is not
// yet wired up. Relationships (parents, spouses) are events too and are handled
// by a separate resource.
package person

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

func NewPersonResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_person"
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
