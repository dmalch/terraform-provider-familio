## 0.7.0

FEATURES:

* **`familio_person` can now record where a life event happened** — three new optional
  attributes take a familio settlement UUID (the same id `familio_settlement_persons` / the
  `familio_person` data source return):
  * `birth_place` — familio's «Место рождения», recorded on the birth event.
  * `death_place` — «Место смерти», on the death event.
  * `christening_place` — on the «Крещение» (baptism) event.

  Places edit **in place** (no resource replacement), ride the same event upsert as the date,
  and read back without a permadiff. A place set without its date still records the event (e.g.
  a known `death_place` with an unknown death date), so a known place is never silently dropped.
  Internally the provider sends familio's required structured settlement object
  (`{"uuid": …}`); a bare UUID string is rejected by the API, which is why this previously could
  not be expressed.

* **Event comments.** familio events carry a free-text comment (примечание); it is now
  exposed everywhere the provider manages an event:
  * `familio_person` — `birth_comment`, `death_comment`, `christening_comment` (edited in
    place alongside the date/place).
  * `familio_marriage` — `comment` on the wedding event (changing it forces replacement, like
    the partners/date, until wedding events can be edited in place).

  (`familio_event` already had `comment`.)

## 0.6.0

FEATURES:

* **New data source `familio_person`** — look up a single person by UUID. It surfaces the
  owning account (`owner_id`, which is absent from the settlement list and previously needed a
  raw API call) plus names, gender, privacy, the formatted birth/death/christening dates, and
  the person's relationships as UUID sets: `parents`, `spouses` and `children`. This makes
  importable tree nodes discoverable declaratively — e.g. walk a person's `parents` to adopt
  ancestors that aren't tagged to a settlement — and lets configs tell your own tree from other
  researchers' profiles or catalog rows by filtering on `owner_id`.

## 0.5.0

FEATURES:

* **Date blocks now express familio's full complex-date model.** Every date attribute
  (`familio_person` birth/death/christening, `familio_marriage.marriage_date`,
  `familio_event.date`) gains optional fields, mirroring the sibling
  `dmalch/terraform-provider-genealogy` date model:
  * `circa` — an approximate date (familio's `about` type), e.g. `{ year = 1846, circa = true }`.
  * `range` — an open bound or span: `before`, `after`, or `between`. `between` takes a
    second endpoint via `end_year`/`end_month`/`end_day`, e.g.
    `{ year = 1846, range = "between", end_year = 1850 }`.
  * `calendar` — `gregorian` (default) or `julian`, for pre-1918 records.

  Plain `{ year = 1846 }` keeps working unchanged. familio carries a single whole-date type,
  so `circa` and `range` are mutually exclusive (and per-endpoint `end_circa` is not
  supported); combining them is rejected at plan time rather than silently writing a wrong date.

BREAKING CHANGES:

* **`familio_event` no longer has a separate `end_date` block.** A date span is now expressed
  within the single `date` block via `range = "between"` and `end_year`/`end_month`/`end_day`
  (e.g. `date = { year = 1878, range = "between", end_year = 1890 }`), so the same date model
  is used across every resource. Update any `familio_event` using `end_date`.

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
