package marriage

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/dmalch/terraform-provider-familio/internal/familio"
)

func TestPartnerListSorted(t *testing.T) {
	set, diags := types.SetValueFrom(context.Background(), types.StringType, []string{"bbb", "aaa"})
	if diags.HasError() {
		t.Fatalf("set build: %v", diags)
	}
	got, d := partnerList(context.Background(), set)
	if d.HasError() {
		t.Fatalf("partnerList: %v", d)
	}
	if len(got) != 2 || got[0] != "aaa" || got[1] != "bbb" {
		t.Errorf("want sorted [aaa bbb], got %v", got)
	}
}

func TestFindWedding(t *testing.T) {
	birthID, wedID := "b1", "w1"
	events := []familio.Event{
		{UUID: &birthID, Type: familio.EventBirth},
		{UUID: &wedID, Type: familio.EventWedding, Participants: []familio.Participant{
			{PersonUUID: "A", Role: familio.RoleSpouse},
			{PersonUUID: "B", Role: familio.RoleSpouse},
		}},
	}
	if got := findWedding(events, "w1"); got == nil {
		t.Fatal("expected to find wedding w1")
	} else if s := got.SpouseUUIDs(); len(s) != 2 {
		t.Errorf("want 2 spouses, got %v", s)
	}
	if findWedding(events, "b1") != nil {
		t.Error("birth event must not match a wedding lookup")
	}
	if findWedding(events, "missing") != nil {
		t.Error("unknown uuid should return nil")
	}
}
