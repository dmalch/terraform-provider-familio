package acceptance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// Real, globally-readable catalog entities captured from familio (an archive
// case/дело and a catalog-person record); used as stable source references.
const (
	accCaseUUID          = "58e68fa4-9e58-4f11-84bd-510a2dc015eb" // «Ревизские сказки»
	accCatalogPersonUUID = "0123e5fb-e298-46e7-8779-a9bfa793ca5a"
	accCatalogKey        = "gwarmil"
)

const accSourcePerson = `
resource "familio_person" "subj" {
  first_name = "АкцТест"
  last_name  = "Источников"
  gender     = "male"
  privacy    = "invisible"
}
`

// TestAccSource_basic covers the standalone familio_source resource: an archive
// `case` (with a comment, edited in place) and a `catalog_person` (with a
// catalog_key), plus an idempotent re-plan and import of the case source.
func TestAccSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy:             checkPersonsDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: accSourcePerson + `
resource "familio_source" "archive" {
  person         = familio_person.subj.uuid
  reference_uuid = "` + accCaseUUID + `"
  type           = "case"
  comment        = "Ревизская сказка"
}

resource "familio_source" "catalog" {
  person         = familio_person.subj.uuid
  reference_uuid = "` + accCatalogPersonUUID + `"
  type           = "catalog_person"
  catalog_key    = "` + accCatalogKey + `"
}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_source.archive", tfjsonpath.New("type"), knownvalue.StringExact("case")),
					statecheck.ExpectKnownValue("familio_source.archive", tfjsonpath.New("comment"), knownvalue.StringExact("Ревизская сказка")),
					statecheck.ExpectKnownValue("familio_source.archive", tfjsonpath.New("name"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("familio_source.archive", tfjsonpath.New("requisites"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("familio_source.catalog", tfjsonpath.New("type"), knownvalue.StringExact("catalog_person")),
					statecheck.ExpectKnownValue("familio_source.catalog", tfjsonpath.New("catalog_key"), knownvalue.StringExact(accCatalogKey)),
				},
			},
			{
				// Edit the comment in place (no replacement).
				Config: accSourcePerson + `
resource "familio_source" "archive" {
  person         = familio_person.subj.uuid
  reference_uuid = "` + accCaseUUID + `"
  type           = "case"
  comment        = "Ревизская сказка 1811 г."
}

resource "familio_source" "catalog" {
  person         = familio_person.subj.uuid
  reference_uuid = "` + accCatalogPersonUUID + `"
  type           = "catalog_person"
  catalog_key    = "` + accCatalogKey + `"
}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_source.archive", tfjsonpath.New("comment"), knownvalue.StringExact("Ревизская сказка 1811 г.")),
				},
			},
			{
				ResourceName:      "familio_source.archive",
				ImportState:       true,
				ImportStateVerify: true,
				// Composite ID: "<person_uuid>:<reference_uuid>".
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					r := s.RootModule().Resources["familio_source.archive"]
					return r.Primary.Attributes["person"] + ":" + r.Primary.Attributes["reference_uuid"], nil
				},
			},
		},
	})
}

// TestAccPersonSources_authoritative covers the inline familio_person `sources`
// block as an authoritative set: two sources, then one removed — the provider
// reconciles familio to match (the removed source is deleted), with no permadiff.
func TestAccPersonSources_authoritative(t *testing.T) {
	const two = `
resource "familio_person" "subj" {
  first_name = "АкцТест"
  last_name  = "Блокисточн"
  gender     = "male"
  privacy    = "invisible"
  sources = [
    { reference_uuid = "` + accCaseUUID + `", type = "case", comment = "Ревизская сказка" },
    { reference_uuid = "` + accCatalogPersonUUID + `", type = "catalog_person", catalog_key = "` + accCatalogKey + `" },
  ]
}
`
	const one = `
resource "familio_person" "subj" {
  first_name = "АкцТест"
  last_name  = "Блокисточн"
  gender     = "male"
  privacy    = "invisible"
  sources = [
    { reference_uuid = "` + accCaseUUID + `", type = "case", comment = "Ревизская сказка" },
  ]
}
`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy:             checkPersonsDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: two,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_person.subj", tfjsonpath.New("sources"), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue("familio_person.subj", tfjsonpath.New("sources").AtSliceIndex(0).AtMapKey("type"), knownvalue.StringExact("case")),
				},
			},
			{
				Config: one,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_person.subj", tfjsonpath.New("sources"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue("familio_person.subj", tfjsonpath.New("sources").AtSliceIndex(0).AtMapKey("reference_uuid"), knownvalue.StringExact(accCaseUUID)),
				},
			},
		},
	})
}
