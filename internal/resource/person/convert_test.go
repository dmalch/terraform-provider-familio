package person

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
	"github.com/dmalch/terraform-provider-familio/internal/tfdate"
)

func TestBasicFromModelDefaultsPrivacy(t *testing.T) {
	m := &ResourceModel{
		FirstName: types.StringValue("Иван"),
		LastName:  types.StringValue("Иванов"),
		Gender:    types.StringValue(familio.GenderMale),
		Privacy:   types.StringNull(),
	}
	got := basicFromModel(m)
	if got.Privacy != familio.PrivacyVisibleForAll {
		t.Errorf("privacy = %q, want default %q", got.Privacy, familio.PrivacyVisibleForAll)
	}
	if got.FirstName != "Иван" || got.LastName != "Иванов" {
		t.Errorf("name not carried: %+v", got)
	}
}

func TestEventsFromModelBirthOnly(t *testing.T) {
	m := &ResourceModel{
		Gender:    types.StringValue(familio.GenderFemale),
		BirthDate: types.ObjectNull(tfdate.AttrTypes),
		DeathDate: types.ObjectNull(tfdate.AttrTypes),
	}
	events, diags := eventsFromModel(context.Background(), m)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if len(events) != 1 || events[0].Type != familio.EventBirth {
		t.Fatalf("want exactly one birth event, got %+v", events)
	}
	if events[0].Date.First != nil {
		t.Errorf("birth date should be unknown (nil First), got %+v", events[0].Date.First)
	}
}

func TestEventsFromModelWithDates(t *testing.T) {
	month := 1
	day := 15
	m := &ResourceModel{
		Gender:    types.StringValue(familio.GenderMale),
		BirthDate: tfdate.Object(&familio.DatePart{Year: 1900, Month: &month, Day: &day}),
		DeathDate: tfdate.Object(&familio.DatePart{Year: 1971}),
	}
	events, diags := eventsFromModel(context.Background(), m)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if len(events) != 2 {
		t.Fatalf("want birth+death events, got %d", len(events))
	}
	b := events[0].Date.First
	if b == nil || b.Year != 1900 || b.Month == nil || *b.Month != 1 || b.Day == nil || *b.Day != 15 {
		t.Errorf("birth date wrong: %+v", b)
	}
	d := events[1].Date.First
	if d == nil || d.Year != 1971 || d.Month != nil || d.Day != nil {
		t.Errorf("death date wrong (year only expected): %+v", d)
	}
}

func TestDateObjectRoundTrip(t *testing.T) {
	month := 6
	obj := tfdate.Object(&familio.DatePart{Year: 1850, Month: &month})
	if obj.IsNull() {
		t.Fatal("object should not be null")
	}
	back, diags := tfdate.PartFromObject(context.Background(), obj)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if back.Year != 1850 || back.Month == nil || *back.Month != 6 || back.Day != nil {
		t.Errorf("round-trip mismatch: %+v", back)
	}
	if !tfdate.Object(nil).IsNull() {
		t.Error("nil DatePart should produce a null object")
	}
}
