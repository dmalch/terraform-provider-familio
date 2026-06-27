package acceptance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccPerson_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy:             checkPersonsDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: `
resource "familio_person" "test" {
  first_name = "АкцТест"
  last_name  = "Персонов"
  gender     = "male"
  privacy    = "invisible"
  birth_date = { year = 1850 }
}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_person.test", tfjsonpath.New("first_name"), knownvalue.StringExact("АкцТест")),
					statecheck.ExpectKnownValue("familio_person.test", tfjsonpath.New("last_name"), knownvalue.StringExact("Персонов")),
					statecheck.ExpectKnownValue("familio_person.test", tfjsonpath.New("gender"), knownvalue.StringExact("male")),
					statecheck.ExpectKnownValue("familio_person.test", tfjsonpath.New("birth_date").AtMapKey("year"), knownvalue.Int64Exact(1850)),
					statecheck.ExpectKnownValue("familio_person.test", tfjsonpath.New("uuid"), knownvalue.NotNull()),
				},
			},
			{
				// In-place edit: change a basic field (exercises the X-Base-Version
				// optimistic-lock header) and add a death date (event upsert).
				Config: `
resource "familio_person" "test" {
  first_name = "АкцТестИзм"
  last_name  = "Персонов"
  gender     = "male"
  privacy    = "invisible"
  birth_date = { year = 1850 }
  death_date = { year = 1899 }
}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("familio_person.test", plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_person.test", tfjsonpath.New("first_name"), knownvalue.StringExact("АкцТестИзм")),
					statecheck.ExpectKnownValue("familio_person.test", tfjsonpath.New("death_date").AtMapKey("year"), knownvalue.Int64Exact(1899)),
				},
			},
			{
				ResourceName:                         "familio_person.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "uuid",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources["familio_person.test"].Primary.Attributes["uuid"], nil
				},
			},
		},
	})
}

// TestAccPerson_parents covers parentage (a child with two parents) and verifies
// that changing a parent and editing the birth date both apply IN PLACE — i.e.
// the child is updated, not replaced (which would lose its uuid and edges).
func TestAccPerson_parents(t *testing.T) {
	const parents = `
resource "familio_person" "dad" {
  first_name = "АкцТест"
  last_name  = "Отцов"
  gender     = "male"
  privacy    = "invisible"
}

resource "familio_person" "mom" {
  first_name = "АкцТест"
  last_name  = "Мамова"
  gender     = "female"
  privacy    = "invisible"
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy:             checkPersonsDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: parents + `
resource "familio_person" "child" {
  first_name = "АкцТест"
  last_name  = "Дитятев"
  gender     = "male"
  privacy    = "invisible"
  birth_date = { year = 1880 }
  parents    = [familio_person.dad.uuid, familio_person.mom.uuid]
}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_person.child", tfjsonpath.New("parents"), knownvalue.SetSizeExact(2)),
					statecheck.ExpectKnownValue("familio_person.child", tfjsonpath.New("birth_date").AtMapKey("year"), knownvalue.Int64Exact(1880)),
				},
			},
			{
				// Drop a parent and change the birth date: must be an in-place update.
				Config: parents + `
resource "familio_person" "child" {
  first_name = "АкцТест"
  last_name  = "Дитятев"
  gender     = "male"
  privacy    = "invisible"
  birth_date = { year = 1881 }
  parents    = [familio_person.dad.uuid]
}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("familio_person.child", plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_person.child", tfjsonpath.New("parents"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue("familio_person.child", tfjsonpath.New("birth_date").AtMapKey("year"), knownvalue.Int64Exact(1881)),
				},
			},
		},
	})
}
