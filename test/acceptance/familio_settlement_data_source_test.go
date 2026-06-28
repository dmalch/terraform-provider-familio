package acceptance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// TestAccSettlementDataSource_basic looks up a known, stable settlement (Нижняя
// Верея) by UUID and asserts the name, administrative requisites, type and
// coordinates are surfaced. Read-only — no resources, no CheckDestroy.
func TestAccSettlementDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "familio_settlement" "verey" {
  uuid = "40d1b180-b739-4ecb-9ee5-ced6fefcd0d8"
}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.familio_settlement.verey", tfjsonpath.New("name"), knownvalue.StringExact("Нижняя Верея")),
					statecheck.ExpectKnownValue("data.familio_settlement.verey", tfjsonpath.New("region"), knownvalue.StringExact("Нижегородская область")),
					statecheck.ExpectKnownValue("data.familio_settlement.verey", tfjsonpath.New("type"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.familio_settlement.verey", tfjsonpath.New("latitude"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.familio_settlement.verey", tfjsonpath.New("longitude"), knownvalue.NotNull()),
				},
			},
		},
	})
}
