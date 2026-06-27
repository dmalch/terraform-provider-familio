// Package event implements the familio_event resource: a single-subject "fact"
// event on a person (residence, education, occupation, military service, award,
// …) — the long tail of familio's ~50-type event catalogue. familio stores these
// uniformly as { type, date, comment, participants:[{owner}] } with no
// type-specific fields, so one generic resource covers them all (see
// internal/familio/API.md). It is an association/sub-resource: Create posts the
// event on the person, Read finds it by uuid on the person's event list, Delete
// removes it. Changing any attribute forces replacement, since familio has no
// in-place event edit and these events do not upsert.
package event

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/dmalch/terraform-provider-familio/internal/config"
	"github.com/dmalch/terraform-provider-familio/internal/familio"
)

type Resource struct {
	client *familio.Client
}

func NewEventResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event"
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

// ImportState parses a composite ID "<person_uuid>:<event_uuid>". An event uuid
// alone is not addressable (events are a person sub-resource with no global GET),
// so import needs the person to anchor the read; Read reconciles the rest.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	person, ev, ok := strings.Cut(req.ID, ":")
	if !ok || person == "" || ev == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			`familio_event import ID must be "<person_uuid>:<event_uuid>"`,
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("person"), person)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("uuid"), ev)...)
}
