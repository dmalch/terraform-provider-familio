package familio

import (
	"encoding/json"
	"testing"
)

func TestSettlementRef(t *testing.T) {
	if SettlementRef("") != nil {
		t.Error("empty uuid should yield a nil settlement (no place)")
	}
	s := SettlementRef("40d1b180")
	if s == nil || s.UUID != "40d1b180" {
		t.Errorf("SettlementRef = %+v, want uuid 40d1b180", s)
	}
}

// TestSettlementWriteIsObjectWithUUID locks the confirmed contract: the wire
// settlement must be the structured {"uuid": …} object (a bare string is
// rejected by familio with 400), and only the uuid is sent.
func TestSettlementWriteIsObjectWithUUID(t *testing.T) {
	ev := BirthEvent(nil, "self", nil, "40d1b180", "")
	b, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got struct {
		Settlement *struct {
			UUID string `json:"uuid"`
		} `json:"settlement"`
	}
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Settlement == nil || got.Settlement.UUID != "40d1b180" {
		t.Errorf("settlement on the wire = %s, want {\"uuid\":\"40d1b180\"}", b)
	}

	// No place ⇒ settlement is explicit null (matches familio's create body).
	none, _ := json.Marshal(BirthEvent(nil, "self", nil, "", ""))
	if !json.Valid(none) {
		t.Fatal("invalid json")
	}
	var probe map[string]json.RawMessage
	_ = json.Unmarshal(none, &probe)
	if string(probe["settlement"]) != "null" {
		t.Errorf("settlement = %s, want null when no place", probe["settlement"])
	}
}

// TestSettlementReadBack decodes the enriched object familio returns
// ({uuid,name,mainGeorequisite}) and confirms the provider recovers the uuid.
func TestSettlementReadBack(t *testing.T) {
	const body = `{"uuid":null,"type":"birth","date":{"calendar":"gregorian","type":"equal","first":null,"second":null},
		"participants":[],"comment":"",
		"settlement":{"uuid":"40d1b180","name":"Нижняя Верея","mainGeorequisite":{"level1":"Нижегородская область","level2":"город Выкса","year":2019}}}`
	var ev Event
	if err := json.Unmarshal([]byte(body), &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ev.SettlementUUID() != "40d1b180" {
		t.Errorf("SettlementUUID = %q, want 40d1b180", ev.SettlementUUID())
	}

	// An event with no place reads back as "" (nil settlement).
	var empty Event
	if err := json.Unmarshal([]byte(`{"type":"birth","settlement":null,"participants":[]}`), &empty); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if empty.SettlementUUID() != "" {
		t.Errorf("SettlementUUID = %q, want empty", empty.SettlementUUID())
	}
}

func TestEventBuildersCarryPlace(t *testing.T) {
	cases := []struct {
		name string
		ev   Event
	}{
		{"birth", BirthEvent(nil, "self", nil, "s1", "")},
		{"death", DeathEvent(nil, "self", "s1", "")},
		{"baptism", BaptismEvent(nil, "self", "s1", "")},
	}
	for _, c := range cases {
		if c.ev.SettlementUUID() != "s1" {
			t.Errorf("%s: place = %q, want s1", c.name, c.ev.SettlementUUID())
		}
	}
}
