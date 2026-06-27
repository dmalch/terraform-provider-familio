## 0.2.0

FEATURES:

* **`familio_person` now manages parents:** a new `parents` set (0–2 person UUIDs) links a
  person to their parents via the child's birth event. familio stores parents as
  gender-agnostic participants, so the set mirrors `familio_marriage.partners` — order does
  not matter and each parent's father/mother role is inferred from their own gender.

ENHANCEMENTS:

* **`familio_person` life-event dates and parents are now edited in place.** Changing
  `birth_date`, `death_date`, or `parents` updates the person without recreating it (the birth
  event is upserted; a removed `death_date` deletes the death event). These no longer force
  replacement.

BUG FIXES:

* **Fixed `familio_person` updates of name / gender / privacy**, which previously failed with
  HTTP 400 «Не указана дата последнего обновления информации». The optimistic-lock token is
  sent in the `X-Base-Version` header (= the last-read `updatedAt`), not a request-body field.
* Added `UseStateForUnknown` plan modifiers to the computed name/privacy attributes, removing
  spurious "known after apply" diffs on unrelated updates.

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
