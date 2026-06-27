package acceptance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccMarriage_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy:             checkPersonsDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: `
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
