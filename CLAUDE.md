# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A Terraform provider for **familio.org** (an unofficial integration — no public write API
exists; endpoints were reverse-engineered from the tree editor). Manages family-tree
**persons**, **marriages**, and **life-fact events** as code. Built on the modern
**terraform-plugin-framework** (not the legacy SDKv2).

`internal/familio/API.md` is the source of truth for the familio.org HTTP surface — read
it before touching anything in `internal/familio/`. It documents the reverse-engineered
endpoints, request/response shapes, and the auth model.

## Commands

```bash
make build         # go build -> bin/terraform-provider-familio
make build-local   # build into a TF filesystem-mirror path (bin/registry.terraform.io/dmalch/familio/<version>/<platform>/) for local `terraform` testing
make test          # go test -v ./...  (unit tests; no network)
make testacc       # TF_ACC=1 acceptance tests against the LIVE API — see warning below
make lint          # runs golangci-lint v2 (installs into bin/ on first run)
make lint-fix      # golangci-lint run --fix
make docs          # regenerate docs/ from schema + examples/ via tfplugindocs

go test -run TestName ./internal/resource/person/   # single test
```

CI (`.github/workflows/ci.yaml`) runs `make build`, `make test`, golangci-lint, **and**
`make docs` with a `git diff --exit-code` check. **After changing any resource/datasource
schema or its examples, run `make docs` and commit `docs/` — CI fails otherwise.**

### Acceptance tests hit the live site

`make testacc` creates and destroys real persons/marriages on the familio.org account whose
session is in `FAMILIO_COOKIES` (or `FAMILIO_SESSION`). They auto-skip when those env vars
are unset. Don't run them casually.

## Architecture

### Layering

- `main.go` → `internal/provider.go` (`FamilioProvider`) registers resources and the data
  source. `New(version)` is the factory; version comes from goreleaser ldflags.
- `Configure` resolves credentials, builds one `*familio.Client`, and hands it to every
  resource/datasource via `*config.ClientData` set on `resp.ResourceData`/`DataSourceData`.
  Each resource's `Configure` type-asserts `req.ProviderData` back to `*config.ClientData`.
- `internal/familio/` — the **HTTP client** for familio.org's `/api/v2` surface. This is
  the only package that talks to the network. Knows nothing about Terraform types.
- `internal/resource/{person,marriage,event}/` and `internal/datasource/settlementpersons/`
  — Terraform-framework adapters. They translate plan/state ↔ `familio` client calls.
- `internal/tfdate/` — shared bridge between familio's complex-date model and the nested
  `{year, month, day}` Terraform attribute. Used by person, marriage, and event.

### Per-resource package convention

Each resource package splits the same way (follow it when adding a resource):
- `resource.go` — boilerplate: `New*Resource`, `Metadata`, `Configure`, `ImportState`.
- `schema.go` — the framework `Schema` (attribute descriptions feed generated docs).
- `model.go` — the `tfsdk`-tagged state struct.
- `convert.go` — pure functions mapping model ↔ `familio` client structs (the part worth
  unit-testing; see `convert_test.go`).
- CRUD lives in either one `crud.go` (marriage, event) or split `create.go`/`read.go`/
  `update.go`/`delete.go` (person). Either is fine.

### Auth is two-layer (the non-obvious part)

The `t` session cookie alone is **rejected** by the authed API (401). familio's Next.js SSR
embeds a short-lived **JWT bearer** in the page's `__NEXT_DATA__`. So the client, from the
`t` cookie, scrapes an HTML page for `"token":"eyJ..."` (`auth.go`), caches it until ~5min
before its JWT `exp`, and sends it as `Authorization: Bearer` on `/api/v2/*` calls. The
JWT's `uuid` claim is the account id, used as `?owner=` on creates. The public
`settlement_persons` read needs neither cookie nor bearer — credentials are optional and a
missing session only warns at configure time.

Credential precedence (each with an env fallback): `cookie` (`FAMILIO_COOKIES`) >
`session_token` (`FAMILIO_SESSION`) > `browser` (extract via sweetcookie from a logged-in
browser). See `resolveCookies` in `provider.go`.

### Domain model (matters when modeling resources)

familio represents relationships and life facts as **events**, not fields:
- A **marriage** *is* the `wedding` event linking two persons. `familio_marriage` is an
  association resource over a partner pair.
- **Birth / death / christening (baptism, «Крещение»)** are single-subject life facts
  folded into `familio_person` as nested date blocks.
- A person's **parents** are gender-agnostic participants on the *child's birth event* —
  father/mother is inferred from each parent's own gender, not stored as a role. Managed via
  the `parents` set (0–2 UUIDs) on `familio_person`.
- `familio_event` covers the long tail of single-subject fact types (`location`/residence,
  `profession`, `education`, `militaryService`, awards, …).

**In-place editing vs. replacement:** the person resource rebuilds its birth/death/
christening events on Update, so those dates and `parents` edit in place. The marriage
resource cannot yet edit its event, so changing a marriage date or its partners forces
replacement — `tfdate.Block(desc, requiresReplace)` encodes this with the `requiresReplace`
flag.

## Conventions

- Russian-language genealogy domain: user-facing names/data are often Cyrillic; keep it.
- Lint config (`.golangci.yml`) is strict and opt-in (`default: none` + an explicit enable
  list including `errcheck`, `errorlint`, `bodyclose`, `noctx`, `forcetypeassert`,
  `godot`). `godot` requires comment sentences to end with a period.
- Errors from the client are wrapped with `%w`; `ErrNotLoggedIn` is returned (via
  `CheckRedirect`) when a request bounces to a login path.
