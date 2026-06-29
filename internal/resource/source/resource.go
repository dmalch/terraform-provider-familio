// Package source implements the familio_source resource: a source citation
// («Источник») on a person — the long tail of familio's «Источники» tab. A
// source is an immutable reference to a catalogued entity (an archive `case`/
// дело, or a `catalog_person` index record) plus an editable comment (see
// internal/familio/API.md). It is an association/sub-resource: Create attaches
// the source to the person, Read finds it by its reference uuid on the person's
// source list, comment edits in place via PATCH, Delete removes it.
//
// The same source set is also exposed as an authoritative `sources` block on
// familio_person; the two surfaces are mutually exclusive per person — manage a
// person's sources via the inline block OR via standalone familio_source
// resources, never both.
package source

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/dmalch/go-familio"
	"github.com/dmalch/terraform-provider-familio/internal/config"
)

type Resource struct {
	client *familio.Client
}

func NewSourceResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source"
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

// ImportState parses a composite ID "<person_uuid>:<reference_uuid>". A source
// is a person sub-resource keyed by the referenced entity's uuid and has no
// global GET, so import needs the person to anchor the read; Read reconciles the
// rest. catalog_key cannot be recovered (write-only) and stays null on import.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	person, ref, ok := strings.Cut(req.ID, ":")
	if !ok || person == "" || ref == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			`familio_source import ID must be "<person_uuid>:<reference_uuid>"`,
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("person"), person)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("reference_uuid"), ref)...)
}
