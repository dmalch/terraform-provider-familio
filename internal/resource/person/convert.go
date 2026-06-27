package person

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

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

// eventsFromModel assembles the create events: the mandatory birth event plus,
// when a death_date is set, a death event.
func eventsFromModel(ctx context.Context, m *ResourceModel) ([]familio.Event, diag.Diagnostics) {
	var diags diag.Diagnostics

	birth, d := datePartFromObject(ctx, m.BirthDate)
	diags.Append(d...)
	events := []familio.Event{familio.SelfBirthEvent(birth)}

	if !m.DeathDate.IsNull() && !m.DeathDate.IsUnknown() {
		death, dd := datePartFromObject(ctx, m.DeathDate)
		diags.Append(dd...)
		events = append(events, familio.SelfDeathEvent(death))
	}
	return events, diags
}

// datePartFromObject converts a nested {year,month,day} object into a DatePart
// (nil when the object is null/unknown).
func datePartFromObject(ctx context.Context, obj types.Object) (*familio.DatePart, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return nil, diags
	}
	var dm dateModel
	diags.Append(obj.As(ctx, &dm, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}
	part := &familio.DatePart{Year: int(dm.Year.ValueInt64())}
	if !dm.Month.IsNull() && !dm.Month.IsUnknown() {
		mm := int(dm.Month.ValueInt64())
		part.Month = &mm
	}
	if !dm.Day.IsNull() && !dm.Day.IsUnknown() {
		dd := int(dm.Day.ValueInt64())
		part.Day = &dd
	}
	return part, diags
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

// applyEventsToState sets birth_date/death_date from a read-back events slice.
func applyEventsToState(events []familio.Event, m *ResourceModel) {
	m.BirthDate = dateObject(eventDatePart(events, familio.EventBirth))
	m.DeathDate = dateObject(eventDatePart(events, familio.EventDeath))
}

// eventDatePart returns the date of the first event of the given type, or nil.
func eventDatePart(events []familio.Event, typ string) *familio.DatePart {
	for i := range events {
		if events[i].Type == typ {
			return events[i].Date.First
		}
	}
	return nil
}

// dateObject builds a nested date object from a DatePart (null when nil).
func dateObject(part *familio.DatePart) types.Object {
	if part == nil {
		return types.ObjectNull(dateAttrTypes)
	}
	obj, _ := types.ObjectValue(dateAttrTypes, map[string]attr.Value{
		"year":  types.Int64Value(int64(part.Year)),
		"month": int64PtrValue(part.Month),
		"day":   int64PtrValue(part.Day),
	})
	return obj
}

func int64PtrValue(p *int) types.Int64 {
	if p == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*p))
}
