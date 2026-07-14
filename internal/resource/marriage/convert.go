package marriage

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/go-familio"
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

// weddingEvent builds the wedding event to upsert, attaching the optional
// place (settlement uuid) that the WeddingEvent constructor does not take.
func weddingEvent(date *familio.DateRange, partnerA, partnerB string, comment, place types.String) familio.Event {
	event := familio.WeddingEvent(date, partnerA, partnerB, strValue(comment))
	event.Settlement = familio.SettlementRef(strValue(place))
	return event
}

// strValue returns the comment/place value to send, or "" when null/unknown.
func strValue(s types.String) string {
	if s.IsNull() || s.IsUnknown() {
		return ""
	}
	return s.ValueString()
}

// strOrNull maps a wedding comment/place back to state, null when empty so an
// omitted attribute does not perpetually diff.
func strOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
