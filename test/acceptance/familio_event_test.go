package acceptance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// TestAccEvent_basic covers a single-subject fact event (a residence with a date
// range + comment): create, read, idempotent re-plan, and import.
func TestAccEvent_basic(t *testing.T) {
	const person = `
resource "familio_person" "subj" {
  first_name = "АкцТест"
  last_name  = "Событьев"
  gender     = "male"
  privacy    = "invisible"
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy:             checkPersonsDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: person + `
resource "familio_event" "test" {
  person   = familio_person.subj.uuid
  type     = "location"
  date     = { year = 1878 }
  end_date = { year = 1890 }
  comment  = "Москва"
}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_event.test", tfjsonpath.New("type"), knownvalue.StringExact("location")),
					statecheck.ExpectKnownValue("familio_event.test", tfjsonpath.New("date").AtMapKey("year"), knownvalue.Int64Exact(1878)),
					statecheck.ExpectKnownValue("familio_event.test", tfjsonpath.New("end_date").AtMapKey("year"), knownvalue.Int64Exact(1890)),
					statecheck.ExpectKnownValue("familio_event.test", tfjsonpath.New("comment"), knownvalue.StringExact("Москва")),
					statecheck.ExpectKnownValue("familio_event.test", tfjsonpath.New("uuid"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         "familio_event.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "uuid",
				// Composite ID: "<person_uuid>:<event_uuid>".
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					e := s.RootModule().Resources["familio_event.test"]
					return e.Primary.Attributes["person"] + ":" + e.Primary.Attributes["uuid"], nil
				},
			},
		},
	})
}
