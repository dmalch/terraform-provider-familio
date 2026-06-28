package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
)

// basicFromModel builds the basic person fields from the plan, defaulting an
// unset privacy to visible_for_all (familio requires a privacy on create).
func basicFromModel(m *ResourceModel) familio.BasicFields {
	privacy := m.Privacy.ValueString()
	if m.Privacy.IsNull() || m.Privacy.IsUnknown() || privacy == "" {
		privacy = familio.PrivacyVisibleForAll
	}
	return familio.BasicFields{
		FirstName:      m.FirstName.ValueString(),
		LastName:       m.LastName.ValueString(),
		MiddleName:     m.Patronymic.ValueString(),
		BirthFirstName: m.BirthFirstName.ValueString(),
		BirthLastName:  m.BirthLastName.ValueString(),
		Gender:         m.Gender.ValueString(),
		Privacy:        privacy,
	}
}

// eventsFromModel assembles the create events from the life-event blocks: the
// mandatory birth event (date + place + comment + parents) plus a death and/or
// christening event when their block carries any information.
func eventsFromModel(ctx context.Context, m *ResourceModel) ([]familio.Event, diag.Diagnostics) {
	var diags diag.Diagnostics

	bdate, bplace, bcomment, parents, d := birthParts(ctx, m.Birth)
	diags.Append(d...)
	events := []familio.Event{familio.BirthEvent(bdate, familio.SelfRef, parents, bplace, bcomment)}

	ddate, dplace, dcomment, dd := lifeEventParts(ctx, m.Death)
	diags.Append(dd...)
	if hasInfo(ddate, dplace, dcomment) {
		events = append(events, familio.SelfDeathEvent(ddate, dplace, dcomment))
	}

	cdate, cplace, ccomment, cc := lifeEventParts(ctx, m.Christening)
	diags.Append(cc...)
	if hasInfo(cdate, cplace, ccomment) {
		events = append(events, familio.SelfBaptismEvent(cdate, cplace, ccomment))
	}
	return events, diags
}

// parentList converts the parents set into a slice of person uuids (nil when
// null/unknown). An empty or null set both mean "no parents".
func parentList(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	if set.IsNull() || set.IsUnknown() {
		return nil, diags
	}
	var ids []string
	diags.Append(set.ElementsAs(ctx, &ids, false)...)
	return ids, diags
}

// applyBasicToState copies the server's basic record (names/gender/privacy +
// uuid + timestamps) onto the model.
func applyBasicToState(rec *familio.BasicRecord, m *ResourceModel) {
	m.UUID = types.StringValue(rec.UUID)
	m.FirstName = types.StringValue(rec.FirstName)
	m.LastName = types.StringValue(rec.LastName)
	m.Patronymic = types.StringValue(rec.MiddleName)
	m.BirthFirstName = types.StringValue(rec.BirthFirstName)
	m.BirthLastName = types.StringValue(rec.BirthLastName)
	m.Gender = types.StringValue(rec.Gender)
	m.Privacy = types.StringValue(rec.Privacy)
	m.CreatedAt = types.StringValue(rec.CreatedAt)
	m.UpdatedAt = types.StringValue(rec.UpdatedAt)
}

// applyEventsToState sets the birth/death/christening blocks from a read-back
// events slice. The birth block is read from the person's OWN birth event (the
// one where they are the child — a parent's /events also lists their children's
// births, where they hold role parent).
func applyEventsToState(ctx context.Context, events []familio.Event, m *ResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	var bdate *familio.DateRange
	var bplace, bcomment string
	var parents []string
	if birth := familio.OwnBirthEvent(events, m.UUID.ValueString()); birth != nil {
		bdate = familio.RangeFromEventDate(birth.Date)
		bplace = birth.SettlementUUID()
		bcomment = birth.Comment
		parents = birth.ParentUUIDs()
	}
	birthObj, d := birthBlockValue(ctx, bdate, bplace, bcomment, parents)
	diags.Append(d...)
	m.Birth = birthObj

	m.Death = lifeEventFromEvents(events, familio.EventDeath)
	m.Christening = lifeEventFromEvents(events, familio.EventBaptism)
	return diags
}

// lifeEventFromEvents builds a death/christening block from the first event of
// the given type, or null when absent.
func lifeEventFromEvents(events []familio.Event, typ string) types.Object {
	for i := range events {
		if events[i].Type == typ {
			return lifeEventBlockValue(
				familio.RangeFromEventDate(events[i].Date),
				events[i].SettlementUUID(),
				events[i].Comment,
			)
		}
	}
	return types.ObjectNull(lifeEventAttrTypes)
}

// strValue returns an optional string attribute's value, or "" when null/unknown.
func strValue(s types.String) string {
	if s.IsNull() || s.IsUnknown() {
		return ""
	}
	return s.ValueString()
}

// strOrNull maps a value to a Terraform string, returning null for "" so an
// absent place/comment reads back as null and matches an omitted config.
func strOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
