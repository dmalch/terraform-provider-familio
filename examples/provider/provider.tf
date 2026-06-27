terraform {
  required_providers {
    familio = {
      source = "dmalch/familio"
    }
  }
}

# Auth via the FAMILIO_COOKIES env var (or set cookie/session_token/browser here).
provider "familio" {}

# List every person linked to a settlement, optionally filtered to one catalog.
data "familio_settlement_persons" "zhuravkino" {
  settlement  = "e0c1a09c-b7ed-4d5c-a22f-3a86db42bbc6"
  catalog_key = "mkzhuravkinotambov"
}

output "person_count" {
  value = length(data.familio_settlement_persons.zhuravkino.persons)
}

# Import/read an existing person (write support pending):
#   terraform import familio_person.example <person-uuid>
resource "familio_person" "example" {
  uuid = "00000000-0000-0000-0000-000000000000"
}
