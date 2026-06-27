# Familio.org API notes

This file is the source of truth for familio.org's HTTP surface as used by the provider.
The **read** path is reverse-engineered and working; the **write** path is the subject of
the Phase 0.5 discovery spike and is not yet known.

## Auth

- Session cookie **`t`** (not OAuth). Installed on a cookie jar scoped to
  `https://familio.org/`. Sources, in precedence order:
  `cookie` attr / `$FAMILIO_COOKIES` → `session_token` attr / `$FAMILIO_SESSION` →
  `browser` attr (sweetcookie extraction).
- A redirect to a `/login`/`/auth`/`/signin` path ⇒ `ErrNotLoggedIn`.

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

## Write (UNKNOWN — Phase 0.5 spike deliverable)

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
