package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
	"github.com/dmalch/terraform-provider-familio/internal/tfdate"
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

// eventsFromModel assembles the create events: the mandatory birth event (with
// any parents and the birth place) plus a death/christening event when that
// fact's date OR its place is set (a known place with an unknown date still
// records the event).
func eventsFromModel(ctx context.Context, m *ResourceModel) ([]familio.Event, diag.Diagnostics) {
	var diags diag.Diagnostics

	birth, d := tfdate.RangeFromObject(ctx, m.BirthDate)
	diags.Append(d...)
	parents, dp := parentList(ctx, m.Parents)
	diags.Append(dp...)
	events := []familio.Event{familio.BirthEvent(birth, familio.SelfRef, parents, placeValue(m.BirthPlace), strValue(m.BirthComment))}

	if hasDate(m.DeathDate) || hasPlace(m.DeathPlace) || hasStr(m.DeathComment) {
		death, dd := tfdate.RangeFromObject(ctx, m.DeathDate)
		diags.Append(dd...)
		events = append(events, familio.SelfDeathEvent(death, placeValue(m.DeathPlace), strValue(m.DeathComment)))
	}
	if hasDate(m.ChristeningDate) || hasPlace(m.ChristeningPlace) || hasStr(m.ChristeningComment) {
		bap, db := tfdate.RangeFromObject(ctx, m.ChristeningDate)
		diags.Append(db...)
		events = append(events, familio.SelfBaptismEvent(bap, placeValue(m.ChristeningPlace), strValue(m.ChristeningComment)))
	}
	return events, diags
}

// placeValue is the settlement uuid carried by a *_place attribute, or "" when
// it is null/unknown (i.e. no place). It is an alias of strValue named for the
// call site.
func placeValue(s types.String) string {
	return strValue(s)
}

// strValue returns an optional string attribute's value, or "" when null/unknown.
func strValue(s types.String) string {
	if s.IsNull() || s.IsUnknown() {
		return ""
	}
	return s.ValueString()
}

// hasDate reports whether a nested date object is set (non-null/known).
func hasDate(o types.Object) bool {
	return !o.IsNull() && !o.IsUnknown()
}

// hasPlace reports whether a *_place attribute carries a settlement uuid.
func hasPlace(s types.String) bool {
	return hasStr(s)
}

// hasStr reports whether an optional string attribute carries a non-empty value.
func hasStr(s types.String) bool {
	return !s.IsNull() && !s.IsUnknown() && s.ValueString() != ""
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

// applyEventsToState sets birth/death/christening dates and places from a
// read-back events slice.
func applyEventsToState(events []familio.Event, m *ResourceModel) {
	uuid := m.UUID.ValueString()
	m.BirthDate = tfdate.ObjectFromRange(ownBirthDate(events, uuid))
	m.DeathDate = tfdate.ObjectFromRange(eventDate(events, familio.EventDeath))
	m.ChristeningDate = tfdate.ObjectFromRange(eventDate(events, familio.EventBaptism))
	m.BirthPlace = ownBirthPlace(events, uuid)
	m.DeathPlace = eventPlace(events, familio.EventDeath)
	m.ChristeningPlace = eventPlace(events, familio.EventBaptism)
	m.BirthComment = ownBirthComment(events, uuid)
	m.DeathComment = eventComment(events, familio.EventDeath)
	m.ChristeningComment = eventComment(events, familio.EventBaptism)
}

// ownBirthComment returns the comment on the person's OWN birth event, or null.
func ownBirthComment(events []familio.Event, personUUID string) types.String {
	birth := familio.OwnBirthEvent(events, personUUID)
	if birth == nil {
		return types.StringNull()
	}
	return strOrNull(birth.Comment)
}

// eventComment returns the comment on the first event of the given type, or null.
func eventComment(events []familio.Event, typ string) types.String {
	for i := range events {
		if events[i].Type == typ {
			return strOrNull(events[i].Comment)
		}
	}
	return types.StringNull()
}

// ownBirthPlace returns the settlement uuid of the person's OWN birth event (the
// one where they are the child — mirrors ownBirthDate), or null.
func ownBirthPlace(events []familio.Event, personUUID string) types.String {
	birth := familio.OwnBirthEvent(events, personUUID)
	if birth == nil {
		return types.StringNull()
	}
	return placeOrNull(birth.SettlementUUID())
}

// eventPlace returns the settlement uuid of the first event of the given type,
// or null.
func eventPlace(events []familio.Event, typ string) types.String {
	for i := range events {
		if events[i].Type == typ {
			return placeOrNull(events[i].SettlementUUID())
		}
	}
	return types.StringNull()
}

// placeOrNull maps a settlement uuid to a Terraform string (null when empty), so
// an absent place reads back as null and matches an omitted config.
func placeOrNull(uuid string) types.String {
	return strOrNull(uuid)
}

// strOrNull maps a value to a Terraform string, returning null for "" so an
// absent place/comment reads back as null and matches an omitted config.
func strOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

// applyParentsToState sets the parents set from this person's own birth event
// (the one where they are the child — a parent's /events also lists their
// children's births). 0 parents ⇒ null, matching an omitted config.
func applyParentsToState(ctx context.Context, events []familio.Event, m *ResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	birth := familio.OwnBirthEvent(events, m.UUID.ValueString())
	if birth == nil {
		m.Parents = types.SetNull(types.StringType)
		return diags
	}
	ids := birth.ParentUUIDs()
	if len(ids) == 0 {
		m.Parents = types.SetNull(types.StringType)
		return diags
	}
	set, d := types.SetValueFrom(ctx, types.StringType, ids)
	diags.Append(d...)
	m.Parents = set
	return diags
}

// ownBirthDate returns the date of the person's OWN birth event — the one where
// they are the child. A person who is also a parent has their children's birth
// events on the same /events list (where they hold role parent), so a plain type
// filter (eventDate) could return a child's birth date instead. Mirrors
// applyParentsToState, which reads parents from this same event.
func ownBirthDate(events []familio.Event, personUUID string) *familio.DateRange {
	birth := familio.OwnBirthEvent(events, personUUID)
	if birth == nil {
		return nil
	}
	return familio.RangeFromEventDate(birth.Date)
}

// eventDate returns the date of the first event of the given type, or nil.
func eventDate(events []familio.Event, typ string) *familio.DateRange {
	for i := range events {
		if events[i].Type == typ {
			return familio.RangeFromEventDate(events[i].Date)
		}
	}
	return nil
}
