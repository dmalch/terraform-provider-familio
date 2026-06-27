## 0.4.2

BUG FIXES:

* **Fixed `terraform import familio_person` silently overwriting an existing person's
  birth date.** When a person is also a parent, familio returns their children's birth events
  on the same `/events` list (where the person holds the `parent` role), so the read picked
  the wrong (or an empty) birth event and imported `birth_date` as `null` — the next `apply`
  then re-asserted the config value with no diff, clobbering the real date (e.g. a known
  b.1889 became b.1890). The birth date is now read from the person's *own* birth event (the
  one where they are the `child`), matching how `parents` is already read.
* **A failed read of a person's events is now a hard error instead of a silent warning**, so
  managed dates and parents can no longer drift to `null` when the events sub-resource is
  unreachable.

## 0.4.1

ENHANCEMENTS:

* `familio_event` now accepts the `godparent` (Восприемник) and `warranter` (Поручитель) event
  types. Per familio's own data model (confirmed by their team) these are single-subject events
  recorded on the godparent/witness — familio does not link them to the godchild/party, so name
  that person in `comment`.

## 0.4.0

FEATURES:

* **New resource `familio_event`** — a single-subject life-fact event on a person, covering the
  long tail of familio's ~50-type event catalogue (residence/`location`, `profession`,
  `education`, `militaryService`, `militaryAward`, `award`, `citizenship`, `emigration`,
  `burial`, …). Attributes: `person`, `type` (validated against the catalogue), `date`, optional
  `end_date` (making the event a date range), and a free-text `comment`. Imported by
  `"<person_uuid>:<event_uuid>"`. birth/death/baptism (on `familio_person`), marriages
  (`familio_marriage`) and the two-participant godparent/warranter events are intentionally not
  handled here, so resources never contend for the same event.

NOTES:

* `familio` has no in-place event edit, so changing any `familio_event` attribute forces
  replacement (Terraform deletes and recreates the underlying event).

## 0.3.0

FEATURES:

* **`familio_person` gains `christening_date`** — a person's baptism (familio's «Крещение»)
  event, as a nested `{year, month, day}` block. Set it to record the christening, remove it
  to delete the event. Edited in place (unlike birth/death, a baptism event does not upsert, so
  a change deletes and recreates it).

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
