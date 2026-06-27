package familio

import "testing"

func TestCookiesFromHeader(t *testing.T) {
	cookies := CookiesFromHeader("t=abc123;  other=xyz ; =skip; bad")
	if len(cookies) != 2 {
		t.Fatalf("expected 2 cookies, got %d: %+v", len(cookies), cookies)
	}
	if cookies[0].Name != "t" || cookies[0].Value != "abc123" {
		t.Errorf("first cookie = %s=%s, want t=abc123", cookies[0].Name, cookies[0].Value)
	}
	if cookies[1].Name != "other" || cookies[1].Value != "xyz" {
		t.Errorf("second cookie = %s=%s, want other=xyz", cookies[1].Name, cookies[1].Value)
	}
}

func TestCookiesFromHeaderEmpty(t *testing.T) {
	if got := CookiesFromHeader("   "); got != nil {
		t.Errorf("expected nil for blank header, got %+v", got)
	}
}

func TestCookieFromSessionToken(t *testing.T) {
	cookies := CookieFromSessionToken("  tok  ")
	if len(cookies) != 1 || cookies[0].Name != "t" || cookies[0].Value != "tok" {
		t.Fatalf("got %+v, want single t=tok cookie", cookies)
	}
	if got := CookieFromSessionToken(""); got != nil {
		t.Errorf("expected nil for empty token, got %+v", got)
	}
}
