package acceptance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// TestAccTreeDataSource_basic builds a tiny family (dad+mom married, one child)
// and crawls the whole connected component from dad, asserting the crawl reaches
// all three persons and that dad's node carries the child and the spouse (with a
// marriage_uuid — the #23/#24 discoverability path).
func TestAccTreeDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy:             checkPersonsDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: `
resource "familio_person" "dad" {
  first_name = "АкцТест"
  last_name  = "Древов"
  gender     = "male"
  privacy    = "invisible"
  birth      = { date = { year = 1850 } }
}

resource "familio_person" "mom" {
  first_name = "АкцТест"
  last_name  = "Древова"
  gender     = "female"
  privacy    = "invisible"
}

resource "familio_person" "child" {
  first_name = "АкцТест"
  last_name  = "Древов"
  gender     = "male"
  privacy    = "invisible"
  birth      = { parents = [familio_person.dad.uuid, familio_person.mom.uuid] }
}

resource "familio_marriage" "m" {
  partners = [familio_person.dad.uuid, familio_person.mom.uuid]
}

# Crawl the whole component from dad. depends_on so the marriage/child edges
# exist before the crawl reads them.
data "familio_tree" "t" {
  root       = familio_person.dad.uuid
  direction  = "component"
  depends_on = [familio_marriage.m, familio_person.child]
}`,
				ConfigStateChecks: []statecheck.StateCheck{
					// dad + mom + child all reached.
					statecheck.ExpectKnownValue("data.familio_tree.t", tfjsonpath.New("nodes"), knownvalue.ListSizeExact(3)),
					// dad is the first node (root, discovery order): one child, one
					// spouse, and that spouse carries the marriage uuid.
					statecheck.ExpectKnownValue("data.familio_tree.t", tfjsonpath.New("nodes").AtSliceIndex(0).AtMapKey("uuid"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("data.familio_tree.t", tfjsonpath.New("nodes").AtSliceIndex(0).AtMapKey("children"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue("data.familio_tree.t", tfjsonpath.New("nodes").AtSliceIndex(0).AtMapKey("spouses"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue("data.familio_tree.t", tfjsonpath.New("nodes").AtSliceIndex(0).AtMapKey("spouses").AtSliceIndex(0).AtMapKey("marriage_uuid"), knownvalue.NotNull()),
				},
			},
		},
	})
}
