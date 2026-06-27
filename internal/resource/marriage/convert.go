package marriage

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
)

// partnerList extracts the partner uuids from the set, sorted so the first is a
// stable anchor for read/delete (the wedding event is reachable via either
// participant's /events).
func partnerList(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	var ids []string
	diags := set.ElementsAs(ctx, &ids, false)
	sort.Strings(ids)
	return ids, diags
}

// findWedding returns the wedding event with the given uuid, or nil.
func findWedding(events []familio.Event, uuid string) *familio.Event {
	for i := range events {
		if events[i].Type == familio.EventWedding && events[i].ID() == uuid {
			return &events[i]
		}
	}
	return nil
}

// commentValue returns the wedding comment to send, or "" when null/unknown.
func commentValue(s types.String) string {
	if s.IsNull() || s.IsUnknown() {
		return ""
	}
	return s.ValueString()
}

// commentOrNull maps a wedding comment back to state, null when empty so an
// omitted comment does not perpetually diff.
func commentOrNull(comment string) types.String {
	if comment == "" {
		return types.StringNull()
	}
	return types.StringValue(comment)
}
