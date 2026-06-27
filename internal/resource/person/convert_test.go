package person

import (
	"context"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
		BirthDate: tfdate.ObjectFromRange(&familio.DateRange{Year: 1900, Month: &month, Day: &day}),
		DeathDate: tfdate.ObjectFromRange(&familio.DateRange{Year: 1971}),
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

func TestEventsFromModelWithChristening(t *testing.T) {
	m := &ResourceModel{
		Gender:          types.StringValue(familio.GenderMale),
		BirthDate:       types.ObjectNull(tfdate.AttrTypes),
		DeathDate:       types.ObjectNull(tfdate.AttrTypes),
		ChristeningDate: tfdate.ObjectFromRange(&familio.DateRange{Year: 1881}),
	}
	events, diags := eventsFromModel(context.Background(), m)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	var bap *familio.Event
	for i := range events {
		if events[i].Type == familio.EventBaptism {
			bap = &events[i]
		}
	}
	if bap == nil {
		t.Fatalf("want a baptism event, got %+v", events)
	}
	if bap.Date.First == nil || bap.Date.First.Year != 1881 {
		t.Errorf("christening year wrong: %+v", bap.Date.First)
	}
	if len(bap.Participants) != 1 || bap.Participants[0].Role != familio.RoleOwner {
		t.Errorf("christening participant should be the owner, got %+v", bap.Participants)
	}
}

func TestEventsFromModelWithParents(t *testing.T) {
	parents := types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("uuid-dad"), types.StringValue("uuid-mom"),
	})
	m := &ResourceModel{
		Gender:    types.StringValue(familio.GenderMale),
		BirthDate: types.ObjectNull(tfdate.AttrTypes),
		DeathDate: types.ObjectNull(tfdate.AttrTypes),
		Parents:   parents,
	}
	events, diags := eventsFromModel(context.Background(), m)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if len(events) != 1 || events[0].Type != familio.EventBirth {
		t.Fatalf("want one birth event, got %+v", events)
	}
	// child=self plus the two parents.
	var child string
	got := events[0].ParentUUIDs()
	for _, p := range events[0].Participants {
		if p.Role == familio.RoleChild {
			child = p.PersonUUID
		}
	}
	if child != familio.SelfRef {
		t.Errorf("child participant = %q, want %q", child, familio.SelfRef)
	}
	sort.Strings(got)
	if len(got) != 2 || got[0] != "uuid-dad" || got[1] != "uuid-mom" {
		t.Errorf("parent participants = %v, want [uuid-dad uuid-mom]", got)
	}
}

func TestApplyEventsToStateUsesOwnBirthEvent(t *testing.T) {
	const personUUID = "uuid-person"
	// A person who is ALSO a parent: their /events lists their child's birth
	// event (where they are role "parent") BEFORE their own birth event (where
	// they are role "child"). A naive first-birth-event pick would read the
	// child's date (1910) or nil; the person's own birth year is 1889.
	events := []familio.Event{
		familio.BirthEvent(&familio.DateRange{Year: 1910}, "uuid-child", []string{personUUID}),
		familio.BirthEvent(&familio.DateRange{Year: 1889}, personUUID, []string{"uuid-dad", "uuid-mom"}),
	}
	m := &ResourceModel{UUID: types.StringValue(personUUID)}
	applyEventsToState(events, m)

	part, diags := tfdate.RangeFromObject(context.Background(), m.BirthDate)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if part == nil {
		t.Fatal("birth_date read back as null; want the person's own birth year 1889")
	}
	if part.Year != 1889 {
		t.Errorf("birth_date year = %d, want 1889 (own birth event, not the child's)", part.Year)
	}
}

func TestApplyParentsToState(t *testing.T) {
	const childUUID = "uuid-child"
	events := []familio.Event{
		// The child's own birth event (this is what we read parents from).
		familio.BirthEvent(nil, childUUID, []string{"uuid-mom", "uuid-dad"}),
		// A child of this person's — present on /events but with role "parent"
		// for childUUID, so it must be ignored when reading childUUID's parents.
		familio.BirthEvent(nil, "uuid-grandchild", []string{childUUID}),
	}
	m := &ResourceModel{UUID: types.StringValue(childUUID)}
	if diags := applyParentsToState(context.Background(), events, m); diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	var ids []string
	if diags := m.Parents.ElementsAs(context.Background(), &ids, false); diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	sort.Strings(ids)
	if len(ids) != 2 || ids[0] != "uuid-dad" || ids[1] != "uuid-mom" {
		t.Errorf("parents = %v, want [uuid-dad uuid-mom]", ids)
	}

	// No parents ⇒ null (matches an omitted config).
	none := []familio.Event{familio.BirthEvent(nil, childUUID, nil)}
	m2 := &ResourceModel{UUID: types.StringValue(childUUID)}
	applyParentsToState(context.Background(), none, m2)
	if !m2.Parents.IsNull() {
		t.Errorf("0 parents should produce a null set, got %v", m2.Parents)
	}
}

func TestDateObjectRoundTrip(t *testing.T) {
	month := 6
	obj := tfdate.ObjectFromRange(&familio.DateRange{Year: 1850, Month: &month})
	if obj.IsNull() {
		t.Fatal("object should not be null")
	}
	back, diags := tfdate.RangeFromObject(context.Background(), obj)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if back.Year != 1850 || back.Month == nil || *back.Month != 6 || back.Day != nil {
		t.Errorf("round-trip mismatch: %+v", back)
	}
	if !tfdate.ObjectFromRange(nil).IsNull() {
		t.Error("nil DateRange should produce a null object")
	}
}
