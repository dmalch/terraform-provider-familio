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

func TestAccMarriage_basic(t *testing.T) {
	const couple = `
resource "familio_person" "husband" {
  first_name = "АкцТест"
  last_name  = "Мужев"
  gender     = "male"
  privacy    = "invisible"
}

resource "familio_person" "wife" {
  first_name = "АкцТест"
  last_name  = "Женева"
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
				Config: couple + `
resource "familio_marriage" "test" {
  partners      = [familio_person.husband.uuid, familio_person.wife.uuid]
  marriage_date = { year = 1875, month = 5, day = 12 }
}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_marriage.test", tfjsonpath.New("partners"), knownvalue.SetSizeExact(2)),
					statecheck.ExpectKnownValue("familio_marriage.test", tfjsonpath.New("marriage_date").AtMapKey("year"), knownvalue.Int64Exact(1875)),
					statecheck.ExpectKnownValue("familio_marriage.test", tfjsonpath.New("uuid"), knownvalue.NotNull()),
				},
			},
			{
				// Editing the date and adding a comment + place (issue #30) must be
				// an in-place update (the underlying wedding event is rebuilt), NOT a
				// replacement. Нижняя Верея — the same real settlement the person
				// place tests use.
				Config: couple + `
resource "familio_marriage" "test" {
  partners      = [familio_person.husband.uuid, familio_person.wife.uuid]
  marriage_date = { year = 1876 }
  place         = "40d1b180-b739-4ecb-9ee5-ced6fefcd0d8"
  comment       = "повторное оглашение"
}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("familio_marriage.test", plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_marriage.test", tfjsonpath.New("marriage_date").AtMapKey("year"), knownvalue.Int64Exact(1876)),
					statecheck.ExpectKnownValue("familio_marriage.test", tfjsonpath.New("place"), knownvalue.StringExact("40d1b180-b739-4ecb-9ee5-ced6fefcd0d8")),
					statecheck.ExpectKnownValue("familio_marriage.test", tfjsonpath.New("comment"), knownvalue.StringExact("повторное оглашение")),
				},
			},
			{
				// No permadiff: the structured settlement reads back to its uuid.
				Config: couple + `
resource "familio_marriage" "test" {
  partners      = [familio_person.husband.uuid, familio_person.wife.uuid]
  marriage_date = { year = 1876 }
  place         = "40d1b180-b739-4ecb-9ee5-ced6fefcd0d8"
  comment       = "повторное оглашение"
}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				ResourceName:                         "familio_marriage.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "uuid",
				// Composite ID: "<partner_person_uuid>:<wedding_event_uuid>".
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					m := s.RootModule().Resources["familio_marriage.test"]
					h := s.RootModule().Resources["familio_person.husband"]
					return h.Primary.Attributes["uuid"] + ":" + m.Primary.Attributes["uuid"], nil
				},
			},
		},
	})
}
