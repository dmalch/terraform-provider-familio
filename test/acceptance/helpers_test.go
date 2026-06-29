package acceptance

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/dmalch/go-familio"
	"github.com/dmalch/terraform-provider-familio/internal"
)

// testProtoV6ProviderFactories wires the in-process provider for the test harness.
var testProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"familio": providerserver.NewProtocol6WithError(internal.New("test")()),
}

// testAccPreCheck skips acceptance tests unless familio credentials are present.
// resource.Test additionally requires TF_ACC=1 to run at all.
func testAccPreCheck(t *testing.T) {
	if os.Getenv("FAMILIO_COOKIES") == "" && os.Getenv("FAMILIO_SESSION") == "" {
		t.Skip("set FAMILIO_COOKIES (or FAMILIO_SESSION) to run familio acceptance tests")
	}
}

// newTestClient builds a familio client from the same env credentials the
// provider uses, for out-of-band CheckDestroy assertions.
func newTestClient(t *testing.T) *familio.Client {
	t.Helper()
	var opts familio.Options
	if header := os.Getenv("FAMILIO_COOKIES"); header != "" {
		opts.Cookies = familio.CookiesFromHeader(header)
	} else if tok := os.Getenv("FAMILIO_SESSION"); tok != "" {
		opts.Cookies = familio.CookieFromSessionToken(tok)
	}
	c, err := familio.NewClient(opts)
	if err != nil {
		t.Fatalf("building test client: %v", err)
	}
	return c
}

// checkPersonsDestroyed asserts every familio_person in state is gone from
// familio. Because a marriage's wedding event is anchored on (and cascades
// with) its partners, this also confirms the marriage's event is gone.
func checkPersonsDestroyed(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := newTestClient(t)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "familio_person" {
				continue
			}
			uuid := rs.Primary.Attributes["uuid"]
			_, err := c.GetPersonBasic(context.Background(), uuid)
			if errors.Is(err, familio.ErrNotFound) {
				continue
			}
			if err != nil {
				return fmt.Errorf("checking person %s: %w", uuid, err)
			}
			return fmt.Errorf("person %s still exists after destroy", uuid)
		}
		return nil
	}
}
