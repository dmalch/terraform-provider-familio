package familio

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
)

func TestSettlementRef(t *testing.T) {
	RegisterTestingT(t)
	Expect(SettlementRef("")).To(BeNil(), "empty uuid should yield no settlement")
	Expect(SettlementRef("40d1b180")).To(Equal(&Settlement{UUID: "40d1b180"}))
}

// TestSettlementWriteIsObjectWithUUID locks the confirmed contract: the wire
// settlement must be the structured {"uuid": …} object (a bare string is
// rejected by familio with 400), and only the uuid is sent.
func TestSettlementWriteIsObjectWithUUID(t *testing.T) {
	RegisterTestingT(t)
	b, err := json.Marshal(BirthEvent(nil, "self", nil, "40d1b180", ""))
	Expect(err).ToNot(HaveOccurred())

	var got struct {
		Settlement *struct {
			UUID string `json:"uuid"`
		} `json:"settlement"`
	}
	Expect(json.Unmarshal(b, &got)).To(Succeed())
	Expect(got.Settlement).ToNot(BeNil())
	Expect(got.Settlement.UUID).To(Equal("40d1b180"))

	// No place ⇒ settlement is explicit null (matches familio's create body).
	none, _ := json.Marshal(BirthEvent(nil, "self", nil, "", ""))
	var probe map[string]json.RawMessage
	Expect(json.Unmarshal(none, &probe)).To(Succeed())
	Expect(string(probe["settlement"])).To(Equal("null"))
}

// TestSettlementReadBack decodes the enriched object familio returns
// ({uuid,name,mainGeorequisite}) and confirms the provider recovers the uuid.
func TestSettlementReadBack(t *testing.T) {
	RegisterTestingT(t)
	const body = `{"uuid":null,"type":"birth","date":{"calendar":"gregorian","type":"equal","first":null,"second":null},
		"participants":[],"comment":"",
		"settlement":{"uuid":"40d1b180","name":"Нижняя Верея","mainGeorequisite":{"level1":"Нижегородская область","level2":"город Выкса","year":2019}}}`
	var ev Event
	Expect(json.Unmarshal([]byte(body), &ev)).To(Succeed())
	Expect(ev.SettlementUUID()).To(Equal("40d1b180"))

	// An event with no place reads back as "" (nil settlement).
	var empty Event
	Expect(json.Unmarshal([]byte(`{"type":"birth","settlement":null,"participants":[]}`), &empty)).To(Succeed())
	Expect(empty.SettlementUUID()).To(BeEmpty())
}

func TestEventBuildersCarryPlace(t *testing.T) {
	RegisterTestingT(t)
	birth := BirthEvent(nil, "self", nil, "s1", "")
	death := DeathEvent(nil, "self", "s1", "")
	baptism := BaptismEvent(nil, "self", "s1", "")
	Expect(birth.SettlementUUID()).To(Equal("s1"))
	Expect(death.SettlementUUID()).To(Equal("s1"))
	Expect(baptism.SettlementUUID()).To(Equal("s1"))
}
