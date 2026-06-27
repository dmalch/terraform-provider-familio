# terraform-provider-familio

A [Terraform](https://www.terraform.io) provider for [familio.org](https://familio.org)
— manage family-tree **persons** and **marriages** as code, in the same spirit as the
`dmalch/genealogy` (Geni) provider.

> **Unofficial.** This project is **not affiliated with, endorsed, or sponsored by
> Familio**. "Familio" is used only to identify the service this provider integrates
> with. No Familio logo or branding is used.

## Status

familio.org has no documented public write API; its endpoints were reverse-engineered
from the tree editor (see [`internal/familio/API.md`](internal/familio/API.md)).

| Capability | Status |
|---|---|
| `familio_settlement_persons` data source (list a settlement's persons) | ✅ works |
| `familio_person` — full CRUD + import (incl. parents) | ✅ works |
| `familio_marriage` — full CRUD + import | ✅ works |
| `familio_event` — life-fact events (residence, education, military, …) | ✅ works |

`familio_marriage` is an association resource: a marriage is the `wedding` event linking
two persons. Birth, death and christening (baptism / «Крещение») are life facts folded into
`familio_person`, and a person's
**parents** are managed there too via the `parents` set (0–2 person UUIDs) — familio stores
them as gender-agnostic participants on the child's birth event, so the role (father/mother)
is inferred from each parent's own gender. Birth/death dates and parents are edited **in
place** on `familio_person`. Editing a **marriage** date or its partners still forces
replacement (event editing there is not yet implemented).

## Using the provider

```hcl
terraform {
  required_providers {
    familio = {
      source  = "dmalch/familio"
      version = "~> 0.1"
    }
  }
}

provider "familio" {
  # credentials via attribute or the FAMILIO_COOKIES env var — see Authentication
}

resource "familio_person" "ivan" {
  first_name = "Иван"
  last_name  = "Иванов"
  gender     = "male"
  birth_date = { year = 1850 }
}

resource "familio_person" "maria" {
  first_name = "Мария"
  last_name  = "Иванова"
  gender     = "female"
}

resource "familio_marriage" "ivan_maria" {
  partners      = [familio_person.ivan.uuid, familio_person.maria.uuid]
  marriage_date = { year = 1875 }
}

# Their child, linked to both parents.
resource "familio_person" "pyotr" {
  first_name = "Пётр"
  last_name  = "Иванов"
  gender     = "male"
  birth_date = { year = 1878 }
  parents    = [familio_person.ivan.uuid, familio_person.maria.uuid]
}

# A life-fact event (here: a residence over a date range).
resource "familio_event" "ivan_residence" {
  person   = familio_person.ivan.uuid
  type     = "location"
  date     = { year = 1878 }
  end_date = { year = 1890 }
  comment  = "Москва"
}
```

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
