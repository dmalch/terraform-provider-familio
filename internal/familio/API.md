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

### Update — `PUT /api/v2/persons/<uuid>/basic`
Endpoint+method confirmed (returns **400 validation**, not 404/500, when the body is close).
Body = the `/basic` fields (firstName/lastName/middleName/birthFirstName/birthLastName/gender/privacy).
**Optimistic-locking gotcha:** requires a *"дата последнего обновления информации"* (last-update
timestamp) field for concurrency — echoing back `updatedAt` from `GET /basic` was NOT accepted, so
the exact field key is still unknown (the only remaining gap). ⇒ provider Update must: GET `/basic`,
send fields back + the concurrency token. Events (dates/places) are updated separately under
`/api/v2/persons/<uuid>/events…` (shape TBD).

### Delete — `DELETE /api/v2/persons/<uuid>` → 204. Confirmed.

### Dates
`{ calendar:"gregorian"|"julian", type:"equal"|"exact"|…, first:{day,month,year,type}|null, second:null }`;
`formatted` is server-computed (`"Неизвестно"` when empty). Validate via
`PUT /api/v2/validate/complex-date` (204 = ok). Surnames: `POST /api/v2/surnames/validate` `{surname}` (204).

## Remaining gaps (need 1–2 more captures)

1. **Update concurrency field** — the exact key for the last-update timestamp in `PUT /…/basic`.
2. **Union / relationships** — adding a spouse/parent/child link (the tree's "Добавить родственника"
   flow). Not yet captured. Likely a separate endpoint; events already carry `participants[].role`
   (`child`/`owner`/…), so partnerships may be modeled via shared events or a dedicated relatives endpoint.

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
