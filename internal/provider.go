package internal

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/dmalch/go-familio"
	"github.com/dmalch/terraform-provider-familio/internal/config"
	dsperson "github.com/dmalch/terraform-provider-familio/internal/datasource/person"
	dssettlement "github.com/dmalch/terraform-provider-familio/internal/datasource/settlement"
	dssettlementpersons "github.com/dmalch/terraform-provider-familio/internal/datasource/settlementpersons"
	dstree "github.com/dmalch/terraform-provider-familio/internal/datasource/tree"
	"github.com/dmalch/terraform-provider-familio/internal/resource/event"
	"github.com/dmalch/terraform-provider-familio/internal/resource/marriage"
	"github.com/dmalch/terraform-provider-familio/internal/resource/person"
	"github.com/dmalch/terraform-provider-familio/internal/resource/source"
)

const (
	envCookies = "FAMILIO_COOKIES"
	envSession = "FAMILIO_SESSION"
	envBrowser = "FAMILIO_BROWSER"
)

// FamilioProvider is the provider implementation.
type FamilioProvider struct {
	version string
}

// New returns a provider factory stamped with the build version (set via
// goreleaser ldflags in main; "dev" for local builds, "test" in tests).
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &FamilioProvider{version: version}
	}
}

func (p *FamilioProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "familio"
	resp.Version = p.version
}

func (p *FamilioProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage persons, marriages and life events on familio.org. " +
			"Unofficial — not affiliated with, endorsed, or sponsored by Familio.",
		Attributes: map[string]schema.Attribute{
			"cookie": schema.StringAttribute{
				Description: "Raw Cookie header containing the familio.org `t` session cookie " +
					"(e.g. \"t=...; ...\"). Falls back to the FAMILIO_COOKIES env var.",
				Optional:  true,
				Sensitive: true,
			},
			"session_token": schema.StringAttribute{
				Description: "Bare familio.org `t` session token; the provider wraps it as a " +
					"`t=<value>` cookie. Falls back to the FAMILIO_SESSION env var.",
				Optional:  true,
				Sensitive: true,
			},
			"browser": schema.StringAttribute{
				Description: "Extract the familio.org session cookie from a logged-in browser " +
					"instead of supplying it directly. One of: chrome, edge, brave, arc, " +
					"chromium, vivaldi, opera, firefox, safari. Falls back to the FAMILIO_BROWSER " +
					"env var. (macOS may require Full Disk Access for the browser's cookie store.)",
				Optional: true,
			},
		},
	}
}

func (p *FamilioProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg config.FamilioProviderConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cookies, err := resolveCookies(cfg)
	if err != nil {
		resp.Diagnostics.AddError("Unable to resolve familio.org credentials", err.Error())
		return
	}
	// Credentials are optional: the settlement-persons read path is public. A
	// missing session cookie only blocks authenticated (gated read / future
	// write) calls, which surface ErrNotLoggedIn at call time. Warn so the
	// failure is not surprising.
	if len(cookies) == 0 {
		resp.Diagnostics.AddWarning(
			"No familio.org credentials configured",
			"Public reads (e.g. the familio_settlement_persons data source) will work, "+
				"but any authenticated call will fail. Set the `cookie`/`session_token`/"+
				"`browser` attribute or the FAMILIO_COOKIES/FAMILIO_SESSION env var.",
		)
	}

	client, err := familio.NewClient(familio.Options{Cookies: cookies})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create familio.org client", err.Error())
		return
	}

	data := &config.ClientData{Client: client}
	resp.ResourceData = data
	resp.DataSourceData = data
}

// resolveCookies applies the documented precedence: cookie > session_token >
// browser, each with its env-var fallback.
func resolveCookies(cfg config.FamilioProviderConfig) ([]*http.Cookie, error) {
	if header := firstNonEmpty(cfg.Cookie.ValueString(), os.Getenv(envCookies)); header != "" {
		return familio.CookiesFromHeader(header), nil
	}
	if token := firstNonEmpty(cfg.SessionToken.ValueString(), os.Getenv(envSession)); token != "" {
		return familio.CookieFromSessionToken(token), nil
	}
	if browser := firstNonEmpty(cfg.Browser.ValueString(), os.Getenv(envBrowser)); browser != "" {
		return familio.CookiesFromBrowser(browser)
	}
	return nil, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func (p *FamilioProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		person.NewPersonResource,
		marriage.NewMarriageResource,
		event.NewEventResource,
		source.NewSourceResource,
	}
}

func (p *FamilioProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		dssettlementpersons.NewDataSource,
		dsperson.NewDataSource,
		dssettlement.NewDataSource,
		dstree.NewDataSource,
	}
}
