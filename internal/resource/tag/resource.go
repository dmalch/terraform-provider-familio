// Package tag implements the familio_tag resource: one of the account's tags
// («метка») — an entry of the /profile/my-tags catalogue. A tag is a
// coloured label with an optional description; it exists independently of any
// person and is attached to persons through the `tags` attribute on
// familio_person (see internal/resource/person/tags.go).
//
// Two things about this resource are unlike the rest of the provider, both
// coming from the API (see go-familio's API.md › Tags sub-resource):
//   - a tag is identified by a small sequential **integer**, not a uuid;
//   - there is no X-Base-Version optimistic lock on any tags endpoint, so a
//     concurrent edit is last-write-wins rather than a 409.
package tag

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/dmalch/go-familio"
	"github.com/dmalch/terraform-provider-familio/internal/config"
)

type Resource struct {
	client *familio.Client
}

func NewTagResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
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

// ImportState takes the tag's numeric id. Terraform import IDs are strings, so
// the digits are parsed here and rejected early — a uuid (what every other
// resource in this provider imports by) would otherwise fail later with a
// confusing API error.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.Atoi(req.ID)
	if err != nil || id <= 0 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("familio_tag import ID must be the tag's numeric id (e.g. \"2832\"), got %q. "+
				"Unlike other familio resources a tag is keyed by an integer, not a uuid; "+
				"run `familio tags list` to see the account's tag ids.", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), int64(id))...)
}
