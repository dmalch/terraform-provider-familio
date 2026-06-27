# Familio.org API notes

This file is the source of truth for familio.org's HTTP surface as used by the provider.
The **read** path is reverse-engineered and working; the **write** path is the subject of
the Phase 0.5 discovery spike and is not yet known.

## Auth — TWO-LAYER (cookie bootstraps a JWT bearer) ⚠️ updated by spike

The `t` session cookie is **NOT** accepted by the authed API directly
(`GET /api/v2/profile` with only the cookie → **401** «Требуется авторизация»).
The real API credential is a **JWT bearer token**:

1. Log in in a browser → familio sets the **`t`** session cookie (HttpOnly).
2. familio's Next.js SSR reads `t` and embeds a **JWT** in the page's
   `__NEXT_DATA__` (`...initialState... "token":"eyJ..."`). It's RS256;
   payload `{ "iat", "exp", "roles":["ROLE_USER"], "uuid":"<userId>" }`,
   expiry ≈ **30 days** after issue.
3. The browser client sends `Authorization: Bearer <jwt>` on every `/api/v2/*`
   call. Confirmed: with the bearer, `GET /api/v2/profile` / `/api/v2/tree` /
   `/api/v2/persons/<uuid>` all return **200** (without it → 401).

**Provider implication:** cookie-only auth is insufficient for reads-that-need-login
and for all writes. The client must, from the `t` cookie, fetch a familio.org HTML
page and scrape the `__NEXT_DATA__` `token`, then send it as `Authorization: Bearer`.
(A dedicated token-mint endpoint was not found — `/api/v2/auth/*`, `/api/v2/me`,
`/api/v2/token` all 404. The SSR-embedded token is the source.) The public
`GET /api/v2/persons?settlement=` read needs neither cookie nor bearer.

- Backend is **API-Platform / Hydra** (`Accept/Content-Type: application/ld+json`;
  list responses use `hydra:member`). `coral.familio.org` is the real backend and is
  reachable **directly, server-side** with the bearer (CORS-lock is browser-only);
  `familio.org/api/v2/*` proxies it. No public OpenAPI/Hydra docs (all `…/docs*` 404).
- Cookies seen on the session: `t` (session), plus DataDome (`__ddg*`) anti-bot and
  `cookieConfirmed`/`records_spoiler` (non-auth).

## Read (known, implemented)

### `GET /api/v2/persons?settlement=<uuid>&itemsPerPage=300&page=<n>`

Public, unauthenticated. Returns all persons (catalog-sourced + user-created) linked to a
settlement. familio's Next.js `/api/v2/*` routes proxy the CORS-locked `coral.familio.org`
backend (host discoverable via `GET /api/v2/coral/path`).

Response envelope:

```jsonc
{ "pager": { "page": 1, "itemsPerPage": 300, "totalItems": 19885 },
  "data": [ {
    "uuid": "85781e3b-…", "displayName": "Августа Степановна", "shortDisplayName": "…",
    "catalogKey": "mkzhuravkinotambov", "catalogName": "Метрические книги …",
    "type": "catalogPerson",
    "birthDate": null, "deathDate": null, "hasDeathEvent": false,
    "birthSettlementText": "", "updatedAt": "2025-02-28T…" } ] }
```

- `catalogKey` is `null` for **user-created** profiles (the provider's write target).
- No server-side catalog facet: `catalog=` + `settlement=` ⇒ `totalItems: 0`. Filter
  `catalogKey` client-side.
- `itemsPerPage` ≤ 300 to avoid backend timeouts; page until a short/empty page.

## Read — authed (confirmed by spike, need Bearer)

- `GET /api/v2/profile` → `{ user:{uuid,email,…}, profile:{displayName,firstName,lastName,middleName,gender,…} }` — current account.
- `GET /api/v2/tree` → `{ nodes:[{ nodeId, nodeParams:{ role, parents:[{sex,nodeId}], partners:[…] } }] }` — the tree graph.
- `GET /api/v2/persons/<uuid>` → a single tree person. **This confirms the provider's guessed read endpoint.** A *regular* (user-created) person looks like:
  `{ uuid, type:"regularPerson", displayName, originalDisplayName, shortDisplayName, ownerId, gender, birthPlace, deathPlace, deathSettlementText, photo, biography, isMine, isMe, canBuildTree, privacyType:"invisible", updatedAt, tags, isGrantedToMe }`.
- Frontend routes (from `_buildManifest`): `/persons/new` (create), `/persons/new/simple/[id]`, `/persons/[personId]` (edit), `/my-tree`, `/tree`, `/persons`.

## Write — CONFIRMED by spike (captured + replayed)

All writes need `Authorization: Bearer <jwt>` + `Content-Type: application/ld+json`.

### Create — `POST /api/v2/persons[?owner=<userId>]` → 201

Request body (only the **birth** event is required; death event optional):
```jsonc
{
  "basic": {
    "firstName": "Иван", "lastName": "Иванов", "middleName": "Иванович",
    "birthFirstName": "", "birthLastName": "",          // maiden name at birth
    "gender": "male",                                    // male | female
    "privacy": "visible_for_all"                         // | "invisible" (others TBD)
  },
  "photo": null,
  "events": [
    { "uuid": null, "type": "birth",
      "date": { "calendar":"gregorian", "type":"equal", "first":null, "second":null },
      "participants": [ { "personUuid":"self", "role":"child" } ],
      "settlement": null, "comment": "" }
    // optional second event: { "type":"death", participants:[{personUuid:"self",role:"owner"}], … }
  ],
  "biography": null
}
```
Response `201` → `{ basic:{ uuid, displayName, firstName, lastName, middleName, gender, privacy, createdAt, updatedAt, … }, photo, events:[{ uuid, type, date:{…,formatted}, participants:[{personUuid,role,displayName,gender}], settlement, comment, createdAt, updatedAt }] }`. **New person uuid = `basic.uuid`.**

### Read sub-resources (Bearer)
- `GET /api/v2/persons/<uuid>` → the `regularPerson` view (displayName, gender, birthPlace, privacyType, …).
- `GET /api/v2/persons/<uuid>/basic` → `{ uuid, createdAt, updatedAt, gender, privacy, firstName, lastName, middleName, birthLastName, birthFirstName }` — the edit-form source.
- `GET /api/v2/persons/<uuid>/events` → `[{ uuid, type, date, settlement, comment, participants, … }]`.

### Update — `PUT /api/v2/persons/<uuid>/basic` → 200 (CONFIRMED, replayed) ⭐
Body is **just the basic fields** (no token in the body):
```jsonc
{ "firstName":"…", "lastName":"…", "middleName":"…", "birthFirstName":"", "birthLastName":"",
  "gender":"female", "privacy":"invisible" }
```
**Concurrency gotcha (corrected):** the optimistic-lock token is the **`X-Base-Version` HTTP header**,
NOT a body field — value = the `updatedAt` you last read from `GET /basic`. (An earlier spike note
wrongly guessed a body `timestamp`/`uuid`; those are ignored, and without the header you get 400 «Не
указана дата последнего обновления информации».) A stale header → **409 Conflict**. Response echoes the
new `/basic` with a bumped `updatedAt`. The familio web editor uses the same `{"X-Base-Version": I}`
header for `/basic`, `/biography`, and `/source` (seen in `_app` chunk). ⇒ provider Update: `GET /basic`
→ `PUT` fields with header `X-Base-Version: <updatedAt>`. (Editing the photo also fires
`DELETE /api/v2/persons/<uuid>/photo`.)

### Delete — `DELETE /api/v2/persons/<uuid>` → 204. Confirmed.

### Relationships are EVENTS, not a union resource (CONFIRMED, captured) ⭐
Familio has **no `union` resource**. Kinship is modelled as **events with `participants[]`**:
- **Marriage/partnership** = a `wedding` event with two `spouse` participants:
  `{"type":"wedding","date":{…},"participants":[{"personUuid":"<A>","role":"spouse"},{"personUuid":"<B>","role":"spouse"}],"settlement":null,"comment":""}`
- **Parent↔child** = a `birth` event whose participants are the **one child** plus 0–2
  **gender-agnostic `parent`** participants (CONFIRMED — `father`/`mother` are NOT valid roles,
  rejected «Не определена роль участника»; the parent's father/mother display is inferred from
  that person's own gender, not the role):
  `participants:[{personUuid:"<child>",role:"child"},{personUuid:"<parentA>",role:"parent"},{personUuid:"<parentB>",role:"parent"}]`.
  Roles confirmed: `child`, `parent`, `owner` (death), `spouse` (wedding).
- `"personUuid":"self"` = placeholder for the person being created in the **same** `POST /api/v2/persons` request.

The tree UI's **"+ Муж/Жена/Отец/Мать/Сын/Дочь"** all route to
`POST /api/v2/persons/new/simple/<existingUuid>?role=spouse|parent|child&gender=…` which submits a
normal `POST /api/v2/persons?owner=<userId>` whose `events[]` carries the relating event (wedding with
the existing person as the other `spouse`, or birth with the existing person as parent/child). So
**creating a related person + the link is one atomic person-create.**

⇒ Provider modelling (decided): there is no `familio_union` object, so marriage is the
**`familio_marriage`** association resource — it POSTs a `wedding` event between two existing persons
(`POST /api/v2/persons/<uuid>/events`, confirmed below). Single-subject life facts (birth/death) are
folded into `familio_person`. Parent↔child links (a child's birth event carrying parent participants)
are deferred to future `father`/`mother` attributes on the person that owns that birth event.

### Dates
`{ calendar:"gregorian"|"julian", type:"equal"|"exact"|…, first:{day,month,year,type}|null, second:null }`;
`formatted` is server-computed (`"Неизвестно"` when empty). Validate via
`PUT /api/v2/validate/complex-date` (204 = ok). Surnames: `POST /api/v2/surnames/validate` `{surname}` (204).

### Events sub-resource (CONFIRMED, replayed) ⭐
Standalone life events — including linking two **existing** persons:
- **Create** `POST /api/v2/persons/<personUuid>/events` with a single Event body
  (`{uuid:null, type, date, participants:[…], settlement:null, comment:""}`) → **201**, returns the
  event with its new `uuid` (+ `createdAt`/`updatedAt`). For a marriage: `type:"wedding"`,
  `participants:[{personUuid:A,role:"spouse"},{personUuid:B,role:"spouse"}]`. The date may be set on
  create (e.g. `first:{day,month,year,type:"gregorian"}` → echoes `formatted:"12.05.1875"`). The event
  shows up on **every participant's** `/events`, so anchor read/delete on any participant.
- **Delete** `DELETE /api/v2/persons/<personUuid>/events/<eventUuid>` → **204**.
- **In-place edit via POST-upsert (CONFIRMED, replayed) ⭐** `PUT …/events/<id>` is still blocked
  by an unknown concurrency-token field, but **you never need it**: re-`POST`ing a `birth`/`death`
  event for a person **upserts that person's single event of that type** — it is a **full replace**
  of participants + date, not an append. So:
  - **Edit birth date / add / remove / change parents:** `POST /persons/<child>/events` a `birth`
    event with `[{child, role:child}, {parent…}]` and the desired date. Whatever you send is the new
    state (omit a parent ⇒ removed; omit the date ⇒ date cleared). The birth event count stays 1; the
    old event uuid is replaced.
  - **Edit / set death date:** same, `POST` a `death` event `[{person, role:owner}]` + date.
  - **Remove the death event:** `DELETE …/events/<deathUuid>` → **204** (death is optional). The
    **birth event is mandatory — deleting the sole birth event is `409`** («not found»), so birth is
    upsert-only (no removal).

`POST /api/v2/events` (no person prefix) → **404**; events are strictly a person sub-resource.

## Remaining gaps (minor)

1. **Wedding-event in-place edit** — the marriage resource still uses RequiresReplace for
   partners/date. The birth/death POST-upsert trick edits a *single-subject* event by re-posting the
   whole event; whether re-posting a `wedding` likewise upserts (vs. creating a duplicate event) is
   untested, so `familio_marriage` keeps RequiresReplace for now. (The `PUT …/events/<id>`
   concurrency-token field name remains unknown but is no longer needed for births/deaths.)
2. **Settlement on events** (place of marriage/birth) — `settlement` accepts null; setting a real
   settlement uuid is untested, so not yet exposed.
3. **Token refresh** — JWT lasts ~30 days; the provider re-scrapes `__NEXT_DATA__.token` near expiry.

### Parent↔child — RESOLVED (`familio_person.parents`)
A child's parents are managed as a `parents` set (0–2 person uuids) on `familio_person` via the birth
event's `parent` participants. Create embeds them in the create birth event; add/remove/change and
birth-date edits all go through the **birth-event POST-upsert** above — no person recreation. Modelled
as a gender-agnostic set (mirrors `familio_marriage.partners`) because the API has no father/mother
roles. Read picks the birth event where the person is the `child` (a parent's `/events` also lists
their children's births, where they are the `parent`).

## Capture recipe (browser)

Fill in each row by driving a logged-in browser through the family-tree editor and
capturing the Network panel. For every action record: **method · full URL (REST `/api/v2/*`
proxy? Next.js server action with `Next-Action` header? direct `coral.familio.org`?) ·
request headers (CSRF / `x-*` token beyond the `t` cookie?) · JSON request body · JSON
response (esp. the new uuid)**.

| Action | Method | URL | Auth/headers | Request body | Response |
|---|---|---|---|---|---|
| Create person | ? | ? | ? | ? | ? |
| Read person by uuid | ? | `GET /api/v2/persons/<uuid>` (guess — confirm) | ? | — | ? |
| Update person (name/gender/dates/place) | ? | ? | ? | ? | ? |
| Add parent / child link | ? | ? | ? | ? | ? |
| Delete person | ? | ? | ? | — | ? |
| Create union (marriage) | ? | ? | ? | ? | ? |
| Add spouse to union | ? | ? | ? | ? | ? |
| Add child to union | ? | ? | ? | ? | ? |
| Set marriage / divorce date | ? | ? | ? | ? | ? |
| Delete union | ? | ? | ? | — | ? |

**Key open question:** does the union editor link two *existing* persons directly, or does
it create a person then merge (like Geni)? This decides whether `CreateUnion` is simple or
needs a temp-profile dance.

### Spike method

Drive the user's real logged-in Chrome (`playwright-cli attach --extension`, reuse the `t`
cookie — see the `familio-search` skill and the `playwright_cli_drive_real_chrome` /
`familio_authenticated_access` memories). This performs **real mutations** on the user's
account; run it only with explicit go-ahead.
