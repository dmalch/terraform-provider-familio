# terraform-provider-familio

A [Terraform](https://www.terraform.io) provider for [familio.org](https://familio.org)
— manage family-tree **persons** and **unions** as code, in the same spirit as the
`dmalch/genealogy` (Geni) provider.

> **Unofficial.** This project is **not affiliated with, endorsed, or sponsored by
> Familio**. "Familio" is used only to identify the service this provider integrates
> with. No Familio logo or branding is used.

## Status

Early development. familio.org has no documented public write API, so the provider is
being built in stages:

| Capability | Status |
|---|---|
| `familio_settlement_persons` data source (list a settlement's persons) | ✅ works |
| `familio_person` — Read / import | ✅ works |
| `familio_person` — Create / Update / Delete | ⛔ pending write-API discovery |
| `familio_union` — full CRUD | ⛔ pending write-API discovery |

Until the tree-editor mutation endpoints are reverse-engineered (see
[`internal/familio/API.md`](internal/familio/API.md)), every write returns an explicit
"write not yet implemented" diagnostic rather than failing obscurely.

## Authentication

familio.org uses a **browser session cookie** named `t` (not OAuth). Provide it any of
three ways (checked in this order):

```hcl
provider "familio" {
  # 1. Raw Cookie header (env fallback: FAMILIO_COOKIES)
  cookie = "t=...; ..."

  # 2. Bare session token, wrapped as t=<value> (env fallback: FAMILIO_SESSION)
  # session_token = "..."

  # 3. Extract the cookie from a logged-in browser
  # browser = "chrome"   # chrome|edge|brave|arc|chromium|vivaldi|opera|firefox|safari
}
```

## Example

```hcl
data "familio_settlement_persons" "zhuravkino" {
  settlement  = "e0c1a09c-b7ed-4d5c-a22f-3a86db42bbc6"
  catalog_key = "mkzhuravkinotambov" # optional client-side filter
}

output "person_count" {
  value = length(data.familio_settlement_persons.zhuravkino.persons)
}

# Read/import an existing person:
#   terraform import familio_person.example <person-uuid>
resource "familio_person" "example" {
  uuid = "85781e3b-...."
}
```

## Development

```bash
go build ./...
go test ./...
```

The HTTP client is inlined under [`internal/familio/`](internal/familio); it mirrors the
cookie-auth model of `go-geni`'s web client. The provider plumbing follows
`terraform-provider-genealogy` / `terraform-provider-myheritage` conventions.

## License

Apache-2.0. See [`LICENSE`](LICENSE).
