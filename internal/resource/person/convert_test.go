package person

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/gomega"

	"github.com/dmalch/go-familio"
)

// birthBlk builds a birth block object for test input (reusing the production
// read-back builder).
func birthBlk(t *testing.T, date *familio.DateRange, place, comment string, parents []string) types.Object {
	t.Helper()
	obj, diags := birthBlockValue(t.Context(), date, place, comment, parents)
	Expect(diags).To(BeEmpty())
	return obj
}

// lifeBlk builds a death/christening block object for test input.
func lifeBlk(date *familio.DateRange, place, comment string) types.Object {
	return lifeEventBlockValue(date, place, comment)
}

func TestBasicFromModelDefaultsPrivacy(t *testing.T) {
	RegisterTestingT(t)
	m := &ResourceModel{
		FirstName: types.StringValue("Иван"),
		LastName:  types.StringValue("Иванов"),
		Gender:    types.StringValue(familio.GenderMale),
		Privacy:   types.StringNull(),
	}
	got := basicFromModel(m)
	Expect(got.Privacy).To(Equal(familio.PrivacyVisibleForAll))
	Expect(got.FirstName).To(Equal("Иван"))
	Expect(got.LastName).To(Equal("Иванов"))
}

func TestEventsFromModel(t *testing.T) {
	t.Run("birth only, with an unknown date, when no blocks are set", func(t *testing.T) {
		RegisterTestingT(t)
		m := &ResourceModel{Gender: types.StringValue(familio.GenderFemale)}

		events, diags := eventsFromModel(t.Context(), m)

		Expect(diags).To(BeEmpty())
		Expect(events).To(HaveLen(1))
		Expect(events[0].Type).To(Equal(familio.EventBirth))
		Expect(events[0].Date.First).To(BeNil())
	})

	t.Run("birth + death events, when both blocks carry a date", func(t *testing.T) {
		RegisterTestingT(t)
		month, day := 1, 15
		m := &ResourceModel{
			Gender: types.StringValue(familio.GenderMale),
			Birth:  birthBlk(t, &familio.DateRange{Year: 1900, Month: &month, Day: &day}, "", "", nil),
			Death:  lifeBlk(&familio.DateRange{Year: 1971}, "", ""),
		}

		events, diags := eventsFromModel(t.Context(), m)

		Expect(diags).To(BeEmpty())
		Expect(events).To(HaveLen(2))
		Expect(events[0].Date.First.Year).To(Equal(1900))
		Expect(*events[0].Date.First.Month).To(Equal(1))
		Expect(*events[0].Date.First.Day).To(Equal(15))
		Expect(events[1].Type).To(Equal(familio.EventDeath))
		Expect(events[1].Date.First.Year).To(Equal(1971))
		Expect(events[1].Date.First.Month).To(BeNil())
	})

	t.Run("a baptism event with an owner participant, when christening is set", func(t *testing.T) {
		RegisterTestingT(t)
		m := &ResourceModel{
			Gender:      types.StringValue(familio.GenderMale),
			Christening: lifeBlk(&familio.DateRange{Year: 1881}, "", ""),
		}

		events, diags := eventsFromModel(t.Context(), m)

		Expect(diags).To(BeEmpty())
		bap := findEvent(events, familio.EventBaptism)
		Expect(bap).ToNot(BeNil())
		Expect(bap.Date.First.Year).To(Equal(1881))
		Expect(bap.Participants).To(HaveLen(1))
		Expect(bap.Participants[0].Role).To(Equal(familio.RoleOwner))
	})

	t.Run("parents on the birth event, when birth.parents is set", func(t *testing.T) {
		RegisterTestingT(t)
		m := &ResourceModel{
			Gender: types.StringValue(familio.GenderMale),
			Birth:  birthBlk(t, nil, "", "", []string{"uuid-dad", "uuid-mom"}),
		}

		events, diags := eventsFromModel(t.Context(), m)

		Expect(diags).To(BeEmpty())
		Expect(events).To(HaveLen(1))
		Expect(childOf(events[0])).To(Equal(familio.SelfRef))
		Expect(events[0].ParentUUIDs()).To(ConsistOf("uuid-dad", "uuid-mom"))
	})

	t.Run("places and comments flow onto their events; a place-only block still records the event", func(t *testing.T) {
		RegisterTestingT(t)
		m := &ResourceModel{
			Gender:      types.StringValue(familio.GenderMale),
			Birth:       birthBlk(t, &familio.DateRange{Year: 1900}, "sett-birth", "born here", nil),
			Death:       lifeBlk(nil, "sett-death", "died abroad"), // place/comment only, no date
			Christening: lifeBlk(&familio.DateRange{Year: 1900}, "sett-bapt", ""),
		}

		events, diags := eventsFromModel(t.Context(), m)

		Expect(diags).To(BeEmpty())
		birth := findEvent(events, familio.EventBirth)
		Expect(birth.SettlementUUID()).To(Equal("sett-birth"))
		Expect(birth.Comment).To(Equal("born here"))

		death := findEvent(events, familio.EventDeath)
		Expect(death).ToNot(BeNil(), "a death event should be created from death.place/comment alone")
		Expect(death.SettlementUUID()).To(Equal("sett-death"))
		Expect(death.Comment).To(Equal("died abroad"))
		Expect(death.Date.First).To(BeNil(), "a place-only death event has an unknown date")

		Expect(findEvent(events, familio.EventBaptism).SettlementUUID()).To(Equal("sett-bapt"))
	})
}

func TestApplyEventsToState(t *testing.T) {
	const personUUID = "uuid-person"

	t.Run("reads the birth block from the person's OWN birth event, not a child's", func(t *testing.T) {
		RegisterTestingT(t)
		// A person who is also a parent: their /events lists their child's birth
		// event (role parent) before their own (role child). The own birth is 1889.
		events := []familio.Event{
			familio.BirthEvent(&familio.DateRange{Year: 1910}, "uuid-child", []string{personUUID}, "sett-child", "child note"),
			familio.BirthEvent(&familio.DateRange{Year: 1889}, personUUID, []string{"uuid-dad", "uuid-mom"}, "sett-own", "own note"),
			familio.DeathEvent(&familio.DateRange{Year: 1950}, personUUID, "sett-death", ""),
		}
		m := &ResourceModel{UUID: types.StringValue(personUUID)}

		Expect(applyEventsToState(t.Context(), events, m)).To(BeEmpty())

		date, place, comment, parents, diags := birthParts(t.Context(), m.Birth)
		Expect(diags).To(BeEmpty())
		Expect(date.Year).To(Equal(1889))
		Expect(place).To(Equal("sett-own"))
		Expect(comment).To(Equal("own note"))
		Expect(parents).To(ConsistOf("uuid-dad", "uuid-mom"))
	})

	t.Run("reads the death block and leaves an absent christening block null", func(t *testing.T) {
		RegisterTestingT(t)
		events := []familio.Event{
			familio.BirthEvent(&familio.DateRange{Year: 1889}, personUUID, nil, "", ""),
			familio.DeathEvent(&familio.DateRange{Year: 1950}, personUUID, "sett-death", ""),
		}
		m := &ResourceModel{UUID: types.StringValue(personUUID)}

		Expect(applyEventsToState(t.Context(), events, m)).To(BeEmpty())

		_, place, _, diags := lifeEventParts(t.Context(), m.Death)
		Expect(diags).To(BeEmpty())
		Expect(place).To(Equal("sett-death"))
		Expect(m.Christening.IsNull()).To(BeTrue())
	})

	t.Run("a no-information birth event reads back as a null block (matches an omitted block)", func(t *testing.T) {
		RegisterTestingT(t)
		events := []familio.Event{familio.BirthEvent(nil, personUUID, nil, "", "")}
		m := &ResourceModel{UUID: types.StringValue(personUUID)}

		Expect(applyEventsToState(t.Context(), events, m)).To(BeEmpty())
		Expect(m.Birth.IsNull()).To(BeTrue())
	})
}

// findEvent returns the first event of the given type, or nil.
func findEvent(events []familio.Event, typ string) *familio.Event {
	for i := range events {
		if events[i].Type == typ {
			return &events[i]
		}
	}
	return nil
}

// childOf returns the child participant's uuid.
func childOf(e familio.Event) string {
	for _, p := range e.Participants {
		if p.Role == familio.RoleChild {
			return p.PersonUUID
		}
	}
	return ""
}
