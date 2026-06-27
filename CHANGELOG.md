## 0.1.0

FEATURES:

* **New data source:** `familio_settlement_persons` — list every person linked to a
  familio.org settlement, with an optional client-side `catalog_key` filter.
* **New resource:** `familio_person` — full create / read / update / delete + import for a
  family-tree person: names (first / last / patronymic / maiden), gender, privacy, and
  birth/death dates (nested `{year, month, day}` blocks). Imported by person UUID.
* **New resource:** `familio_marriage` — a marriage between two persons, modelled as the
  underlying `wedding` event that links them (an association resource). Full CRUD + import.
  Imported by `"<partner_person_uuid>:<wedding_event_uuid>"`.
* **Authentication:** familio.org session via a raw `cookie` header, a bare `session_token`,
  or `browser` extraction — each with `FAMILIO_COOKIES` / `FAMILIO_SESSION` env fallbacks. The
  session cookie is exchanged for the API's JWT bearer automatically.

NOTES:

* Editing an existing event's date in place is not yet supported, so changing a
  birth/death/marriage date forces resource replacement.
* Unofficial — not affiliated with, endorsed, or sponsored by Familio.
