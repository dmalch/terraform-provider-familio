package familio

import (
	"sort"
	"testing"
)

func TestChildrenOf(t *testing.T) {
	const person = "uuid-person"
	events := []Event{
		// The person's OWN birth (they are the child) — must NOT count as a child.
		BirthEvent(nil, person, []string{"uuid-grandpa"}, "", ""),
		// Two children whose birth events list the person as a parent.
		BirthEvent(nil, "uuid-son", []string{person, "uuid-spouse"}, "", ""),
		BirthEvent(nil, "uuid-daughter", []string{person}, "", ""),
		// An unrelated birth (person not a participant) — ignored.
		BirthEvent(nil, "uuid-stranger", []string{"uuid-other"}, "", ""),
	}
	got := ChildrenOf(events, person)
	sort.Strings(got)
	if len(got) != 2 || got[0] != "uuid-daughter" || got[1] != "uuid-son" {
		t.Errorf("children = %v, want [uuid-daughter uuid-son]", got)
	}
}

func TestSpousesOf(t *testing.T) {
	const person = "uuid-person"
	events := []Event{
		WeddingEvent(nil, person, "uuid-wife", ""),
		WeddingEvent(nil, "uuid-second-wife", person, ""),
		// A wedding the person is not part of — ignored.
		WeddingEvent(nil, "uuid-a", "uuid-b", ""),
		// Person's own birth — not a wedding, ignored.
		BirthEvent(nil, person, nil, "", ""),
	}
	got := SpousesOf(events, person)
	sort.Strings(got)
	if len(got) != 2 || got[0] != "uuid-second-wife" || got[1] != "uuid-wife" {
		t.Errorf("spouses = %v, want [uuid-second-wife uuid-wife]", got)
	}
}
