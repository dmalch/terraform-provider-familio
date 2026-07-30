package acceptance

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/dmalch/go-familio"
)

// TestAccTag_basic covers the tag catalogue resource end to end: create two
// tags, edit one in place (name, colour and description all edit in place — the
// tags endpoints have no immutable fields and no optimistic lock), and confirm
// both are gone after destroy.
func TestAccTag_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy:             checkTagsDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: `
resource "familio_tag" "verify" {
  name        = "АкцТест Проверить"
  color       = "mint-mist"
  description = "Нужен запрос в архив"
}

resource "familio_tag" "plain" {
  name  = "АкцТест Без описания"
  color = "rose-mist"
}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_tag.verify", tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue("familio_tag.verify", tfjsonpath.New("color"),
						knownvalue.StringExact("mint-mist")),
					// hex is derived locally from the palette code.
					statecheck.ExpectKnownValue("familio_tag.verify", tfjsonpath.New("hex"),
						knownvalue.StringExact("#EBFFEB")),
					statecheck.ExpectKnownValue("familio_tag.verify", tfjsonpath.New("description"),
						knownvalue.StringExact("Нужен запрос в архив")),
					// An omitted description must read back as null, not "".
					statecheck.ExpectKnownValue("familio_tag.plain", tfjsonpath.New("description"),
						knownvalue.Null()),
				},
			},
			{
				// Import by the numeric id, not a uuid.
				ResourceName:      "familio_tag.verify",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Every writable field edits in place; the id must survive.
				Config: `
resource "familio_tag" "verify" {
  name  = "АкцТест Проверено"
  color = "ice-blue"
}

resource "familio_tag" "plain" {
  name        = "АкцТест Без описания"
  color       = "rose-mist"
  description = "теперь с описанием"
}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_tag.verify", tfjsonpath.New("name"),
						knownvalue.StringExact("АкцТест Проверено")),
					statecheck.ExpectKnownValue("familio_tag.verify", tfjsonpath.New("hex"),
						knownvalue.StringExact("#EBFEFF")),
					// Clearing the description sends "" and reads back null.
					statecheck.ExpectKnownValue("familio_tag.verify", tfjsonpath.New("description"),
						knownvalue.Null()),
					statecheck.ExpectKnownValue("familio_tag.plain", tfjsonpath.New("description"),
						knownvalue.StringExact("теперь с описанием")),
				},
			},
		},
	})
}

// TestAccPersonTags covers the authoritative `tags` attribute on familio_person:
// assigning, adding, removing down to the empty set, and finally dropping the
// attribute so the tags become unmanaged. The tag itself survives being
// unassigned — only familio_tag's destroy removes it.
func TestAccPersonTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testProtoV6ProviderFactories,
		CheckDestroy:             checkPersonsAndTagsDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: personTagsConfig(`tags = [familio_tag.a.id]`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_person.p", tfjsonpath.New("tags"),
						knownvalue.SetSizeExact(1)),
				},
			},
			{
				// Adding a second id assigns only the missing one.
				Config: personTagsConfig(`tags = [familio_tag.a.id, familio_tag.b.id]`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_person.p", tfjsonpath.New("tags"),
						knownvalue.SetSizeExact(2)),
				},
			},
			{
				// Dropping an id unassigns it without deleting the tag.
				Config: personTagsConfig(`tags = [familio_tag.b.id]`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_person.p", tfjsonpath.New("tags"),
						knownvalue.SetSizeExact(1)),
					// familio_tag.a still exists — it was only detached.
					statecheck.ExpectKnownValue("familio_tag.a", tfjsonpath.New("id"), knownvalue.NotNull()),
				},
			},
			{
				// [] is "remove them all", distinct from omitting the attribute.
				Config: personTagsConfig(`tags = []`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_person.p", tfjsonpath.New("tags"),
						knownvalue.SetSizeExact(0)),
				},
			},
			{
				// Omitting the attribute leaves the tags unmanaged (null), and must
				// not plan a change against the empty set from the previous step.
				Config: personTagsConfig(""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("familio_person.p", tfjsonpath.New("tags"), knownvalue.Null()),
				},
			},
		},
	})
}

// personTagsConfig renders a person with two tags available and the given tags
// attribute line (empty for "attribute omitted").
func personTagsConfig(tagsLine string) string {
	return `
resource "familio_tag" "a" {
  name  = "АкцТест Метка А"
  color = "mint-mist"
}

resource "familio_tag" "b" {
  name  = "АкцТест Метка Б"
  color = "lilac-glow"
}

resource "familio_person" "p" {
  first_name = "АкцТест"
  last_name  = "Меткин"
  gender     = "male"
  privacy    = "invisible"
  birth      = { date = { year = 1880 } }

  ` + tagsLine + `
}`
}

// checkTagsDestroyed asserts every familio_tag in state is gone from the
// account's catalogue. familio has no GET for a single tag, so this reads the
// whole list once.
func checkTagsDestroyed(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := newTestClient(t)
		tags, err := c.ListTags(context.Background())
		if err != nil {
			if errors.Is(err, familio.ErrNotFound) {
				return nil
			}
			return fmt.Errorf("listing tags after destroy: %w", err)
		}
		alive := make(map[int]bool, len(tags))
		for _, tag := range tags {
			alive[tag.ID] = true
		}
		for name, rs := range s.RootModule().Resources {
			if rs.Type != "familio_tag" {
				continue
			}
			id, err := strconv.Atoi(rs.Primary.Attributes["id"])
			if err != nil {
				return fmt.Errorf("%s has a non-numeric id %q", name, rs.Primary.Attributes["id"])
			}
			if alive[id] {
				return fmt.Errorf("tag %d (%s) still exists after destroy", id, name)
			}
		}
		return nil
	}
}

// checkPersonsAndTagsDestroyed runs both destroy checks, since the person-tags
// test creates each.
func checkPersonsAndTagsDestroyed(t *testing.T) resource.TestCheckFunc {
	persons := checkPersonsDestroyed(t)
	tags := checkTagsDestroyed(t)
	return func(s *terraform.State) error {
		if err := persons(s); err != nil {
			return err
		}
		return tags(s)
	}
}
