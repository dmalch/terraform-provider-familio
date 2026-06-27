// Package familio is the inlined HTTP client for familio.org, used by the
// Terraform provider. It mirrors the cookie-auth model of go-geni's web client
// (a session `t` cookie installed on a jar scoped to https://familio.org/),
// rather than Geni's OAuth flow.
//
// Today only the read path is implemented: the public, paginated
// GET /api/v2/persons?settlement=<uuid> endpoint. The tree-editor mutation
// endpoints (create/update/delete persons and unions) are not yet
// reverse-engineered; every write method returns ErrWriteNotImplemented until
// the Phase 0.5 discovery spike documents them in API.md.
package familio

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

const (
	defaultBaseURL   = "https://familio.org/"
	apiV2Path        = "api/v2/"
	defaultUserAgent = "terraform-provider-familio/0.1 (+https://github.com/dmalch/terraform-provider-familio)"
	defaultRateLimit = 2.0
	defaultTimeout   = 60 * time.Second
)

// Client talks to familio.org's /api/v2 surface with a session cookie.
type Client struct {
	httpClient *http.Client
	baseURL    *url.URL
	userAgent  string
	limiter    *rate.Limiter
}

// Options configures a Client. At least one cookie carrying the `t` session
// token should be supplied for any authenticated (write, or gated read) call;
// the public settlement-persons read works without one.
type Options struct {
	// Cookies carries the familio.org session. NewClient installs them on a
	// cookie jar scoped to BaseURL. Build with CookiesFromHeader (paste from
	// DevTools / $FAMILIO_COOKIES) or CookiesFromBrowser (sweetcookie).
	Cookies []*http.Cookie

	// BaseURL overrides https://familio.org/. Useful for tests; production
	// callers should leave it empty.
	BaseURL string

	// UserAgent is sent on every request. Defaults to defaultUserAgent.
	UserAgent string

	// RateLimit caps outgoing requests in requests-per-second. Defaults to
	// defaultRateLimit if unset or non-positive.
	RateLimit float64

	// HTTPClient overrides the default *http.Client. NewClient always sets its
	// Jar from Cookies and its CheckRedirect to detect login bounces, so the
	// override does not need to.
	HTTPClient *http.Client
}

// NewClient builds a Client from Options.
func NewClient(opts Options) (*Client, error) {
	rawBase := opts.BaseURL
	if rawBase == "" {
		rawBase = defaultBaseURL
	}
	base, err := url.Parse(rawBase)
	if err != nil {
		return nil, fmt.Errorf("familio: invalid base URL %q: %w", rawBase, err)
	}

	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("familio: building cookie jar: %w", err)
	}
	if len(opts.Cookies) > 0 {
		jar.SetCookies(base, opts.Cookies)
	}
	httpClient.Jar = jar

	// A redirect to a login/auth path means the session is not valid. Surface
	// it as ErrNotLoggedIn instead of silently following to an HTML login form.
	httpClient.CheckRedirect = func(req *http.Request, _ []*http.Request) error {
		if isLoginPath(req.URL.Path) {
			return ErrNotLoggedIn
		}
		return nil
	}

	userAgent := opts.UserAgent
	if userAgent == "" {
		userAgent = defaultUserAgent
	}

	rl := opts.RateLimit
	if rl <= 0 {
		rl = defaultRateLimit
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    base,
		userAgent:  userAgent,
		limiter:    rate.NewLimiter(rate.Limit(rl), 1),
	}, nil
}

func isLoginPath(path string) bool {
	path = strings.ToLower(path)
	return strings.Contains(path, "/login") || strings.Contains(path, "/auth/") || strings.Contains(path, "/signin")
}
