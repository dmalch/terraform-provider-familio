package acceptance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// TestAccPersonDataSource_basic builds a tiny family (dad+mom married, with one
// child) and looks dad up via the data source, asserting it surfaces the owning
// account and walks the relationships (one child, one spouse, no parents).
func TestAccPersonDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy:             checkPersonsDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: `
resource "familio_person" "dad" {
  first_name = "АкцТест"
  last_name  = "Источников"
  gender     = "male"
  privacy    = "invisible"
  birth_date = { year = 1850 }
}

resource "familio_person" "mom" {
  first_name = "АкцТест"
  last_name  = "Источникова"
  gender     = "female"
  privacy    = "invisible"
}

resource "familio_person" "child" {
  first_name = "АкцТест"
  last_name  = "Источников"
  gender     = "male"
  privacy    = "invisible"
  parents    = [familio_person.dad.uuid, familio_person.mom.uuid]
}

resource "familio_marriage" "m" {
  partners = [familio_person.dad.uuid, familio_person.mom.uuid]
}

data "familio_person" "dad" {
  uuid = familio_person.dad.uuid
}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.familio_person.dad", tfjsonpath.New("owner_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.familio_person.dad", tfjsonpath.New("birth_date"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.familio_person.dad", tfjsonpath.New("children"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue("data.familio_person.dad", tfjsonpath.New("spouses"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue("data.familio_person.dad", tfjsonpath.New("parents"), knownvalue.SetSizeExact(0)),
				},
			},
		},
	})
}
