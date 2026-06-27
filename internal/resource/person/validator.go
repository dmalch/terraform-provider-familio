package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resource validates its config at plan time.
var _ resource.ResourceWithValidateConfig = (*Resource)(nil)

// ValidateConfig fails early when a person has no name. familio rejects a
// nameless person on create ("Не определено имя персоны"); catching it at plan
// time gives a clear message instead of an opaque API 400.
func (r *Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Unknown values may resolve to a name later (e.g. from another resource).
	if data.FirstName.IsUnknown() || data.LastName.IsUnknown() {
		return
	}
	// ValueString() returns "" for null, so this covers omitted and empty alike.
	if data.FirstName.ValueString() == "" && data.LastName.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("first_name"),
			"Missing name",
			"A familio person needs a name: set at least one of first_name or last_name.",
		)
	}
}
