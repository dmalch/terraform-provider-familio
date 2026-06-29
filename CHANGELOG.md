## 0.13.0

FEATURES:

* **`familio_person` now manages the person's biography** — the free-text «tab=2» life
  description — via a new optional `biography` attribute. It is set at create time and edited
  **in place** (a `~ update`, never a replacement), backed by familio's `/biography`
  sub-resource (`PUT` with its own `X-Base-Version` optimistic-lock token, distinct from
  `/basic`). Requires `go-familio` v0.2.0.

## 0.12.1

MAINTENANCE:

* **HTTP client extracted to a standalone module.** The reverse-engineered familio.org client,
  previously vendored under `internal/familio/`, now lives in its own repo and Go module
  [`github.com/dmalch/go-familio`](https://github.com/dmalch/go-familio) (pinned in `go.mod`), so
  the same HTTP layer is reusable from CLIs and scripts. No user-facing behavior change: the
  provider's schema, resources, and data sources are unchanged. The Terraform-specific bridges
  (`internal/tfdate`, `internal/tfsource`, `internal/config`) stay in this repo.

## 0.12.0

BUG FIXES:

* **`familio_source` decodes the catalog as an object `{key, hidden}`** rather than a plain string,
  matching the familio.org API response shape.

## 0.11.0

FEATURES:

* **New data source `familio_settlement`** — look up a settlement (place) by UUID and get its
  canonical `name`, administrative requisites (`region`/`district`/`as_of_year`), `type`, `status`,
  `additional_names` and `latitude`/`longitude`. Resolves and validates the settlement UUIDs that
  `familio_person`'s birth/death/christening places and `familio_source` reference, and makes
  configs self-documenting. Backed by `GET /api/v2/settlements/<uuid>`.

## 0.10.0

ENHANCEMENTS:

* **`familio_marriage` now edits the marriage date and comment in place** instead of forcing a
  destroy/recreate ([#15](https://github.com/dmalch/terraform-provider-familio/issues/15)).
  Changing `marriage_date` or `comment` plans as a `~ update`, not a `-/+ replacement`. familio has
  no event edit and wedding events don't upsert, so under the hood the provider rebuilds the wedding
  event (delete + create) — exactly as it already does for a person's christening — which means the
  computed `uuid` (and `created_at`/`updated_at`) is regenerated on such an edit and shows as
  "known after apply". Changing `partners` still forces replacement: the partner pair is the
  marriage's identity, so a different pair is a different marriage.

## 0.9.0

FEATURES:

* **New resource `familio_source`** — manage a person's source citations (the «Источники» tab).
  A source references a catalogued entity (an archival document — `case`/дело — or a
  `catalog_person` index record) via `reference_uuid` + `type` (+ `catalog_key` for
  catalog-person records); the reference is immutable while `comment` edits in place.
  `name`/`requisites`/`years`/`catalog` are read back from familio. Imported by
  `"<person_uuid>:<reference_uuid>"`.
* **`familio_person` gains a `sources` block** — the same citations managed inline as an
  authoritative list (the provider makes familio match it exactly; `sources = []` removes all,
  an omitted block leaves them unmanaged).

NOTES:

* A given person's sources should be managed through **one** surface — the inline
  `familio_person.sources` block **or** standalone `familio_source` resources — not both (like
  `aws_security_group` inline rules vs. `aws_security_group_rule`). `catalog_key` is write-only at
  the familio API, so it is not recovered on import/refresh.

## 0.8.0

BREAKING CHANGES:

* **`familio_person` life events are now nested blocks.** The flat `*_date` / `*_place` /
  `*_comment` attributes and the top-level `parents` set are replaced by three blocks —
  `birth`, `death`, `christening` — each grouping that event's `date`, `place` and `comment`
  (and, for `birth`, `parents`):

  ```hcl
  # before (≤ 0.7.0)                # after (0.8.0)
  birth_date  = { year = 1880 }     birth = {
  birth_place = "<uuid>"              date    = { year = 1880 }
  parents     = [a.uuid, b.uuid]      place   = "<uuid>"
                                      comment = "..."
                                      parents = [a.uuid, b.uuid]
                                    }
  ```

  This mirrors familio's own model (a date, place, comment and parents are all facets of one
  event) and keeps the resource flat-free as more event fields are added. Update configs and
  re-`import` / refresh state. The `date` block's contents (year/month/day, circa, range,
  end_*, calendar) and all behaviour (in-place editing, no permadiff, a place/comment recorded
  even with an unknown date) are unchanged.

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
