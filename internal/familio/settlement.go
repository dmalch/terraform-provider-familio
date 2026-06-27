package familio

// Settlement is a place attached to an event — familio's «Место рождения /
// смерти / etc.». The write contract is a structured object, NOT a bare uuid: a
// bare string is rejected (HTTP 400). The minimal accepted body is
// `{"uuid": "<id>"}`; the server enriches the name and administrative requisites
// («реквизиты»), which come back on read but are not needed by the provider, so
// only the uuid is modelled (extra read fields are ignored). See the "Settlement
// / place on events" section of internal/familio/API.md.
type Settlement struct {
	UUID string `json:"uuid"`
}

// SettlementRef wraps a settlement uuid as an event place. An empty uuid yields
// nil — i.e. no place (which clears any existing place on an upsert).
func SettlementRef(uuid string) *Settlement {
	if uuid == "" {
		return nil
	}
	return &Settlement{UUID: uuid}
}

// SettlementUUID returns the event's place uuid, or "" when no place is set.
func (e *Event) SettlementUUID() string {
	if e.Settlement == nil {
		return ""
	}
	return e.Settlement.UUID
}
