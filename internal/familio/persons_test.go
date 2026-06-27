package familio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListSettlementPersonsPaginates(t *testing.T) {
	// Two full pages of settlementPageSize, then a short final page.
	total := settlementPageSize*2 + 5
	var gotCookie string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/persons" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		gotCookie = r.Header.Get("Cookie")

		page := 1
		if _, err := fmt.Sscanf(r.URL.Query().Get("page"), "%d", &page); err != nil {
			page = 1
		}

		start := (page - 1) * settlementPageSize
		end := start + settlementPageSize
		if end > total {
			end = total
		}

		var data []Person
		for i := start; i < end; i++ {
			data = append(data, Person{UUID: fmt.Sprintf("u-%d", i)})
		}
		resp := personsPage{Data: data}
		resp.Pager.TotalItems = total
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client, err := NewClient(Options{
		BaseURL:   srv.URL + "/",
		Cookies:   CookiesFromHeader("t=secret"),
		RateLimit: 1000,
	})
	if err != nil {
		t.Fatal(err)
	}

	persons, err := client.ListSettlementPersons(context.Background(), "settlement-uuid")
	if err != nil {
		t.Fatal(err)
	}
	if len(persons) != total {
		t.Errorf("got %d persons, want %d", len(persons), total)
	}
	if gotCookie == "" || !contains(gotCookie, "t=secret") {
		t.Errorf("session cookie not sent; Cookie header = %q", gotCookie)
	}
}

func TestDoReturnsErrNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client, _ := NewClient(Options{BaseURL: srv.URL + "/", RateLimit: 1000})
	_, err := client.GetPerson(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestFlexDateUnmarshal(t *testing.T) {
	cases := []struct {
		name        string
		json        string
		wantPresent bool
		wantValue   string
	}{
		{"null", `{"birthDate":null}`, false, ""},
		{"string", `{"birthDate":"1890"}`, true, "1890"},
		{"object", `{"birthDate":{"type":"equal","calendar":"gregorian","first":null,"second":null,"formatted":"Неизвестно"}}`, true, "Неизвестно"},
		{"missing", `{}`, false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var p Person
			if err := json.Unmarshal([]byte(tc.json), &p); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			v, ok := p.BirthDate.Value()
			if ok != tc.wantPresent {
				t.Errorf("present = %v, want %v", ok, tc.wantPresent)
			}
			if ok && v != tc.wantValue {
				t.Errorf("value = %q, want %q", v, tc.wantValue)
			}
		})
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
