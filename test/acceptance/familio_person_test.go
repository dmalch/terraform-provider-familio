package acceptance

import (
	"fmt"
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
  first_name  = "АкцТест"
  last_name   = "Персонов"
  gender      = "male"
  privacy     = "invisible"
  birth       = { date = { year = 1850 } }
  christening = { date = { year = 1850, month = 4 } }
  biography   = "Крестьянин села Нижняя Верея."
}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_person.test", tfjsonpath.New("first_name"), knownvalue.StringExact("АкцТест")),
					statecheck.ExpectKnownValue("familio_person.test", tfjsonpath.New("last_name"), knownvalue.StringExact("Персонов")),
					statecheck.ExpectKnownValue("familio_person.test", tfjsonpath.New("gender"), knownvalue.StringExact("male")),
					statecheck.ExpectKnownValue("familio_person.test", tfjsonpath.New("birth").AtMapKey("date").AtMapKey("year"), knownvalue.Int64Exact(1850)),
					statecheck.ExpectKnownValue("familio_person.test", tfjsonpath.New("christening").AtMapKey("date").AtMapKey("month"), knownvalue.Int64Exact(4)),
					statecheck.ExpectKnownValue("familio_person.test", tfjsonpath.New("biography"), knownvalue.StringExact("Крестьянин села Нижняя Верея.")),
					statecheck.ExpectKnownValue("familio_person.test", tfjsonpath.New("uuid"), knownvalue.NotNull()),
				},
			},
			{
				// In-place edit: change a basic field (exercises the X-Base-Version
				// optimistic-lock header), add a death date (event upsert), and edit
				// the biography in place (its own /biography sub-resource version).
				Config: `
resource "familio_person" "test" {
  first_name  = "АкцТестИзм"
  last_name   = "Персонов"
  gender      = "male"
  privacy     = "invisible"
  birth       = { date = { year = 1850 } }
  death       = { date = { year = 1899 } }
  christening = { date = { year = 1851 } }
  biography   = "Крестьянин села Нижняя Верея. Участник Первой мировой войны."
}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("familio_person.test", plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_person.test", tfjsonpath.New("first_name"), knownvalue.StringExact("АкцТестИзм")),
					statecheck.ExpectKnownValue("familio_person.test", tfjsonpath.New("death").AtMapKey("date").AtMapKey("year"), knownvalue.Int64Exact(1899)),
					// christening edited in place (delete + recreate the baptism event).
					statecheck.ExpectKnownValue("familio_person.test", tfjsonpath.New("christening").AtMapKey("date").AtMapKey("year"), knownvalue.Int64Exact(1851)),
					statecheck.ExpectKnownValue("familio_person.test", tfjsonpath.New("biography"), knownvalue.StringExact("Крестьянин села Нижняя Верея. Участник Первой мировой войны.")),
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

// TestAccPerson_approximateDates exercises the #5 date model: an approximate
// (circa → "about") birth, a julian-calendar christening, and a "before" death
// bound. The second step asserts an empty re-plan, proving the dates read back
// without a perpetual diff (RangeFromEventDate round-trips against the live API).
func TestAccPerson_approximateDates(t *testing.T) {
	const config = `
resource "familio_person" "approx" {
  first_name  = "АкцТест"
  last_name   = "Примернов"
  gender      = "male"
  privacy     = "invisible"
  birth       = { date = { year = 1846, circa = true } }
  christening = { date = { year = 1846, calendar = "julian" } }
  death       = { date = { year = 1901, range = "before" } }
}`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy:             checkPersonsDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: config,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_person.approx", tfjsonpath.New("birth").AtMapKey("date").AtMapKey("circa"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("familio_person.approx", tfjsonpath.New("christening").AtMapKey("date").AtMapKey("calendar"), knownvalue.StringExact("julian")),
					statecheck.ExpectKnownValue("familio_person.approx", tfjsonpath.New("death").AtMapKey("date").AtMapKey("range"), knownvalue.StringExact("before")),
				},
			},
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

// TestAccPerson_places exercises #12: birth/death/christening place + comment set
// inside the life-event blocks, to real familio settlement UUIDs. It statechecks
// the values, asserts an empty re-plan (the structured settlement round-trips
// with no permadiff), then edits the birth place in place. Uses real settlements:
// Нижняя Верея and Верхняя Верея (Нижегородская область, город Выкса).
func TestAccPerson_places(t *testing.T) {
	const nizhnyayaVereya = "40d1b180-b739-4ecb-9ee5-ced6fefcd0d8"
	const verkhnyayaVereya = "227e549f-56f3-4844-9d7f-db928cee93fd"
	config := func(birthPlace string) string {
		return fmt.Sprintf(`
resource "familio_person" "place" {
  first_name = "АкцТест"
  last_name  = "Местов"
  gender     = "male"
  privacy    = "invisible"
  birth = {
    date    = { year = 1900 }
    place   = %q
    comment = "Метрическая книга, запись о рождении."
  }
  death       = { date = { year = 1970 }, place = %q }
  christening  = { date = { year = 1900 }, place = %q }
}`, birthPlace, nizhnyayaVereya, nizhnyayaVereya)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy:             checkPersonsDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: config(nizhnyayaVereya),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_person.place", tfjsonpath.New("birth").AtMapKey("place"), knownvalue.StringExact(nizhnyayaVereya)),
					statecheck.ExpectKnownValue("familio_person.place", tfjsonpath.New("death").AtMapKey("place"), knownvalue.StringExact(nizhnyayaVereya)),
					statecheck.ExpectKnownValue("familio_person.place", tfjsonpath.New("christening").AtMapKey("place"), knownvalue.StringExact(nizhnyayaVereya)),
					statecheck.ExpectKnownValue("familio_person.place", tfjsonpath.New("birth").AtMapKey("comment"), knownvalue.StringExact("Метрическая книга, запись о рождении.")),
				},
			},
			{
				// No permadiff: the structured settlement reads back to its uuid.
				Config: config(nizhnyayaVereya),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				// Edit the birth place in place (no resource replacement).
				Config: config(verkhnyayaVereya),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("familio_person.place", plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_person.place", tfjsonpath.New("birth").AtMapKey("place"), knownvalue.StringExact(verkhnyayaVereya)),
				},
			},
		},
	})
}

// TestAccPerson_parents covers parentage (a child with two parents inside the
// birth block) and verifies that changing a parent and editing the birth date
// both apply IN PLACE — i.e. the child is updated, not replaced (which would
// lose its uuid and edges).
func TestAccPerson_parents(t *testing.T) {
	const parents = `
resource "familio_person" "dad" {
  first_name = "АкцТест"
  last_name  = "Отцов"
  gender     = "male"
  privacy    = "invisible"
  birth      = { date = { year = 1860 } }
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
  birth = {
    date    = { year = 1880 }
    parents = [familio_person.dad.uuid, familio_person.mom.uuid]
  }
}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_person.child", tfjsonpath.New("birth").AtMapKey("parents"), knownvalue.SetSizeExact(2)),
					statecheck.ExpectKnownValue("familio_person.child", tfjsonpath.New("birth").AtMapKey("date").AtMapKey("year"), knownvalue.Int64Exact(1880)),
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
  birth = {
    date    = { year = 1881 }
    parents = [familio_person.dad.uuid]
  }
}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("familio_person.child", plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_person.child", tfjsonpath.New("birth").AtMapKey("parents"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue("familio_person.child", tfjsonpath.New("birth").AtMapKey("date").AtMapKey("year"), knownvalue.Int64Exact(1881)),
				},
			},
			{
				// Regression for #4: dad is a parent, so dad's /events also lists the
				// child's birth event. Importing dad must read back dad's OWN birth
				// year (1860), not the child's (1881); ImportStateVerify fails if
				// birth is dropped to null or read from the wrong event.
				ResourceName:                         "familio_person.dad",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "uuid",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources["familio_person.dad"].Primary.Attributes["uuid"], nil
				},
			},
		},
	})
}
