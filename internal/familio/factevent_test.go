package familio

import "testing"

func TestMakeDateEqualVsBetween(t *testing.T) {
	single := MakeDate(&DatePart{Year: 1900}, nil)
	if single.Type != dateTypeEqual || single.Second != nil {
		t.Errorf("single date should be equal with no second, got %+v", single)
	}

	rng := MakeDate(&DatePart{Year: 1900}, &DatePart{Year: 1910})
	if rng.Type != dateTypeBetween || rng.First == nil || rng.Second == nil {
		t.Errorf("range date should be between with both ends, got %+v", rng)
	}
	if rng.Second.Year != 1910 {
		t.Errorf("range end year = %d, want 1910", rng.Second.Year)
	}
}

func TestFactEvent(t *testing.T) {
	ev := FactEvent("location", MakeDate(&DatePart{Year: 1880}, nil), "person-uuid", "Москва")
	if ev.Type != "location" {
		t.Errorf("type = %q, want location", ev.Type)
	}
	if ev.Comment != "Москва" {
		t.Errorf("comment = %q", ev.Comment)
	}
	if len(ev.Participants) != 1 || ev.Participants[0].Role != RoleOwner || ev.Participants[0].PersonUUID != "person-uuid" {
		t.Errorf("participant should be the owner person, got %+v", ev.Participants)
	}
}

func TestFactEventTypesExcludesManagedElsewhere(t *testing.T) {
	excluded := map[string]bool{"birth": true, "death": true, "baptism": true, "wedding": true, "divorce": true}
	for _, ty := range FactEventTypes {
		if excluded[ty] {
			t.Errorf("FactEventTypes must not include %q (managed by a dedicated surface)", ty)
		}
	}
}

func TestFactEventTypesIncludesGodparentAndWarranter(t *testing.T) {
	// Per Familio's own model these are single-subject events on the
	// godparent/witness, so familio_event must accept them.
	present := map[string]bool{}
	for _, ty := range FactEventTypes {
		present[ty] = true
	}
	for _, ty := range []string{"godparent", "warranter"} {
		if !present[ty] {
			t.Errorf("FactEventTypes must include %q", ty)
		}
	}
}
