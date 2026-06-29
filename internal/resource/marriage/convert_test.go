package marriage

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/gomega"

	"github.com/dmalch/go-familio"
)

func TestPartnerListSorted(t *testing.T) {
	RegisterTestingT(t)
	set, diags := types.SetValueFrom(context.Background(), types.StringType, []string{"bbb", "aaa"})
	Expect(diags).To(BeEmpty())

	got, d := partnerList(context.Background(), set)
	Expect(d).To(BeEmpty())
	Expect(got).To(Equal([]string{"aaa", "bbb"}))
}

func TestFindWedding(t *testing.T) {
	RegisterTestingT(t)
	birthID, wedID := "b1", "w1"
	events := []familio.Event{
		{UUID: &birthID, Type: familio.EventBirth},
		{UUID: &wedID, Type: familio.EventWedding, Participants: []familio.Participant{
			{PersonUUID: "A", Role: familio.RoleSpouse},
			{PersonUUID: "B", Role: familio.RoleSpouse},
		}},
	}

	wed := findWedding(events, "w1")
	Expect(wed).ToNot(BeNil())
	Expect(wed.SpouseUUIDs()).To(ConsistOf("A", "B"))
	Expect(findWedding(events, "b1")).To(BeNil(), "a birth event must not match a wedding lookup")
	Expect(findWedding(events, "missing")).To(BeNil())
}
