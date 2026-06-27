package acceptance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
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
