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
// any parents) plus, when a death_date is set, a death event.
func eventsFromModel(ctx context.Context, m *ResourceModel) ([]familio.Event, diag.Diagnostics) {
	var diags diag.Diagnostics

	birth, d := tfdate.RangeFromObject(ctx, m.BirthDate)
	diags.Append(d...)
	parents, dp := parentList(ctx, m.Parents)
	diags.Append(dp...)
	events := []familio.Event{familio.BirthEvent(birth, familio.SelfRef, parents)}

	if !m.DeathDate.IsNull() && !m.DeathDate.IsUnknown() {
		death, dd := tfdate.RangeFromObject(ctx, m.DeathDate)
		diags.Append(dd...)
		events = append(events, familio.SelfDeathEvent(death))
	}
	if !m.ChristeningDate.IsNull() && !m.ChristeningDate.IsUnknown() {
		bap, db := tfdate.RangeFromObject(ctx, m.ChristeningDate)
		diags.Append(db...)
		events = append(events, familio.SelfBaptismEvent(bap))
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

// applyEventsToState sets birth_date/death_date/christening_date from a read-back
// events slice.
func applyEventsToState(events []familio.Event, m *ResourceModel) {
	m.BirthDate = tfdate.ObjectFromRange(ownBirthDate(events, m.UUID.ValueString()))
	m.DeathDate = tfdate.ObjectFromRange(eventDate(events, familio.EventDeath))
	m.ChristeningDate = tfdate.ObjectFromRange(eventDate(events, familio.EventBaptism))
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
