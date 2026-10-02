package splunkexport

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testOptions(t *testing.T, address string) Options {
	t.Helper()
	t.Setenv("EXPORT_TEST_TOKEN", "synthetic-secret")
	return Options{ManagementURL: address, AuthMode: "bearer", CredentialEnv: "EXPORT_TEST_TOKEN", InstanceID: "fixture", RequestTimeout: time.Second, JobTimeout: 2 * time.Second, OverallTimeout: 2 * time.Second, MaxRows: 10}
}
func fixtureCA(t *testing.T, s *httptest.Server) string {
	t.Helper()
	cert, err := x509.ParseCertificate(s.TLS.Certificates[0].Certificate[0])
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "ca.pem")
	if err = os.WriteFile(p, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}), 0600); err != nil {
		t.Fatal(err)
	}
	return p
}
func TestTransportRequiresHTTPSAndTrustedCA(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"ok":true}`) }))
	t.Cleanup(s.Close)
	o := testOptions(t, s.URL)
	c, err := NewClient(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	var out map[string]bool
	if err = c.getJSON(context.Background(), "/services/server/info", nil, &out); err == nil {
		t.Fatal("untrusted certificate accepted")
	}
	o.CAFile = fixtureCA(t, s)
	c2, err := NewClient(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c2.Close)
	if err = c2.getJSON(context.Background(), "/services/server/info", nil, &out); err != nil || !out["ok"] {
		t.Fatalf("trusted CA failed: %v", err)
	}
}
func TestTransportHTTPRequiresExplicitOptIn(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{}`) }))
	t.Cleanup(s.Close)
	o := testOptions(t, s.URL)
	if _, err := NewClient(o); err == nil {
		t.Fatal("HTTP accepted by default")
	}
	o.AllowInsecure = true
	c, err := NewClient(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	var out any
	if err = c.getJSON(context.Background(), "/services/server/info", nil, &out); err != nil {
		t.Fatal(err)
	}
}
func TestTransportUnverifiedTLSRequiresExplicitOptIn(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{}`) }))
	t.Cleanup(s.Close)
	o := testOptions(t, s.URL)
	o.AllowInsecure = true
	c, err := NewClient(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	var out any
	if err = c.getJSON(context.Background(), "/services/server/info", nil, &out); err != nil {
		t.Fatal(err)
	}
}
func TestTransportRefusesRedirects(t *testing.T) {
	received := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { received = true }))
	t.Cleanup(target.Close)
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL, 302) }))
	t.Cleanup(s.Close)
	o := testOptions(t, s.URL)
	o.CAFile = fixtureCA(t, s)
	c, err := NewClient(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	var out any
	if err = c.getJSON(context.Background(), "/services/server/info", nil, &out); err == nil {
		t.Fatal("redirect accepted")
	}
	if received {
		t.Fatal("redirect target contacted")
	}
}
func TestTransportBoundsBodyAndSanitizesErrors(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
	}{{"oversize", strings.Repeat("x", 8*1024*1024+1), 200}, {"unauthorized", "synthetic-secret raw detail", 401}, {"unicode", `{"s":"\ud800"}`, 200}, {"malformed", `synthetic-secret`, 200}} {
		t.Run(tc.name, func(t *testing.T) {
			s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tc.status); fmt.Fprint(w, tc.body) }))
			t.Cleanup(s.Close)
			o := testOptions(t, s.URL)
			o.CAFile = fixtureCA(t, s)
			c, err := NewClient(o)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(c.Close)
			var out any
			err = c.getJSON(context.Background(), "/services/server/info", nil, &out)
			if err == nil {
				t.Fatal("bad response accepted")
			}
			if strings.Contains(err.Error(), "synthetic-secret") || strings.Contains(err.Error(), s.URL) {
				t.Fatal("raw response leaked")
			}
		})
	}
}
func TestTransportAuthModes(t *testing.T) {
	for _, mode := range []string{"bearer", "session"} {
		t.Run(mode, func(t *testing.T) {
			s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				want := "Bearer synthetic-secret"
				if mode == "session" {
					want = "Splunk synthetic-secret"
				}
				if r.Header.Get("Authorization") != want {
					t.Error("wrong auth mode")
				}
				if r.URL.Query().Get("output_mode") != "json" {
					t.Error("missing output mode")
				}
				fmt.Fprint(w, `{}`)
			}))
			t.Cleanup(s.Close)
			o := testOptions(t, s.URL)
			o.AuthMode = mode
			o.CAFile = fixtureCA(t, s)
			c, err := NewClient(o)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(c.Close)
			var out any
			if err = c.getJSON(context.Background(), "/services/server/info", nil, &out); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestTransportOptionsValidation(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"--version"}} {
		if _, err := ParseOptions(args); err != ErrHelp && err != ErrVersion {
			t.Fatalf("special flag: %v", err)
		}
	}
	base := []string{"--management-url", "https://example.test", "--credential-env", "MISSING", "--instance-id", "fixture"}
	o, err := ParseOptions(append(base, "--app", "a", "--app", "a", "--app", "b"))
	if err != nil || len(o.Scope.App.Values) != 2 {
		t.Fatalf("set normalization: %v", err)
	}
	for _, args := range [][]string{{"--request-timeout", "0s"}, {"--max-rows", "9223372036854775807"}, {"--app", "*"}, {"--owner", "-"}, {"--allow-http"}, {"--auth-mode", "bad"}} {
		if _, err = ParseOptions(append(base, args...)); err == nil {
			t.Fatalf("accepted invalid options %v", args)
		}
	}
	for _, url := range []string{"https://user:synthetic-secret@example.test", "https://example.test/services", "https://example.test?secret=synthetic-secret", "https://example.test#fragment"} {
		o := testOptions(t, url)
		if _, err = NewClient(o); err == nil || strings.Contains(err.Error(), "synthetic-secret") {
			t.Fatal("origin validation failed or leaked")
		}
	}
}
func TestTransportWindowValidation(t *testing.T) {
	base := []string{"--management-url", "https://example.test", "--credential-env", "MISSING"}
	o, err := ParseOptions(base)
	if err != nil || o.Window.Mode != "all_retained" {
		t.Fatalf("default window: %+v %v", o.Window, err)
	}
	for _, bounds := range [][]string{{"--earliest", "0", "--latest", "1"}, {"--earliest", "2026-01-01T00:00:00+01:00", "--latest", "2026-01-02T00:00:00Z"}, {"--earliest", "2026-01-02T00:00:00Z", "--latest", "2026-01-01T00:00:00Z"}} {
		if _, err = ParseOptions(append(base, bounds...)); err == nil {
			t.Fatal("bad bounds accepted")
		}
	}
	o, err = ParseOptions(append(base, "--earliest", "2026-01-01T00:00:00.123456789Z", "--latest", "2026-01-02T00:00:00Z"))
	if err != nil || o.Window.Mode != "bounded" {
		t.Fatalf("bounded window: %v", err)
	}
}

func TestTransportOneSidedWindow(t *testing.T) {
	base := []string{"--management-url", "https://example.test", "--credential-env", "MISSING"}
	for _, flag := range []string{"--earliest", "--latest"} {
		o, err := ParseOptions(append(base, flag, "2026-01-01T00:00:00Z"))
		if err != nil || o.Window.Mode != "bounded" {
			t.Fatalf("one-sided window %s: %v", flag, err)
		}
	}
}
func TestTransportCredentialValidation(t *testing.T) {
	o := testOptions(t, "https://example.test")
	t.Setenv(o.CredentialEnv, "synthetic-secret\n")
	if _, err := NewClient(o); err == nil {
		t.Fatal("multiline environment credential accepted")
	}
	o.CredentialEnv = ""
	o.CredentialFile = filepath.Join(t.TempDir(), "credential")
	if err := os.WriteFile(o.CredentialFile, []byte(strings.Repeat("x", 64*1024+1)), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewClient(o); err == nil {
		t.Fatal("oversize file credential accepted")
	}
}
func TestTransportPathEscapesOnce(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/servicesNS/owner/app%20name/saved/searches" {
			t.Errorf("bad escaped path %s", r.URL.EscapedPath())
		}
		fmt.Fprint(w, `{}`)
	}))
	t.Cleanup(s.Close)
	o := testOptions(t, s.URL)
	o.CAFile = fixtureCA(t, s)
	c, err := NewClient(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	var out any
	if err = c.getJSON(context.Background(), "/servicesNS/owner/app%20name/saved/searches", nil, &out); err != nil {
		t.Fatal(err)
	}
}
func TestTransportCanonicalUTCWindow(t *testing.T) {
	o, err := ParseOptions([]string{"--management-url", "https://example.test", "--credential-env", "MISSING", "--earliest", "2026-01-01T00:00:00.000+00:00"})
	if err != nil || o.Window.Earliest != "2026-01-01T00:00:00Z" {
		t.Fatalf("canonical UTC: %+v %v", o.Window, err)
	}
}
func TestTransportStreamingBodyDeadline(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"incomplete":`)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	t.Cleanup(s.Close)
	o := testOptions(t, s.URL)
	o.CAFile = fixtureCA(t, s)
	o.RequestTimeout = 20 * time.Millisecond
	c, err := NewClient(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	var out any
	err = c.getJSON(context.Background(), "/services/server/info", nil, &out)
	if err == nil || err.Error() != "request_timeout" {
		t.Fatalf("streaming timeout: %v", err)
	}
}
func TestTransportYearOneBounds(t *testing.T) {
	base := []string{"--management-url", "https://example.test", "--credential-env", "MISSING"}
	for _, bounds := range [][]string{{"--earliest", "0001-01-01T00:00:00Z", "--latest", "0001-01-01T00:00:00Z"}, {"--earliest", "0001-01-01T00:00:00Z", "--latest", "0000-01-01T00:00:00Z"}} {
		if _, err := ParseOptions(append(base, bounds...)); err == nil {
			t.Fatal("unordered year-one bounds accepted")
		}
	}
}

func TestTransportStrictRFC3339WindowLexemes(t *testing.T) {
	base := []string{"--management-url", "https://example.test", "--credential-env", "MISSING"}
	for _, bound := range []string{"--earliest", "--latest"} {
		t.Run(bound, func(t *testing.T) {
			for _, value := range []string{"2026-01-01T0:00:00Z", "2026-01-01T00:00:00,123Z", "2026-01-01T00:00:00.1234567890Z", "2026-01-01T00:00:00.0000000001+00:00"} {
				t.Run(value, func(t *testing.T) {
					if _, err := ParseOptions(append(base, bound, value)); err == nil {
						t.Fatal("non-contractual timestamp accepted")
					}
				})
			}
			for _, value := range []string{"2026-01-01T00:00:00Z", "2026-01-01T00:00:00+00:00", "2026-01-01T00:00:00.1Z", "2026-01-01T00:00:00.123456789Z", "2026-01-01T00:00:00.123456789+00:00"} {
				t.Run(value, func(t *testing.T) {
					o, err := ParseOptions(append(base, bound, value))
					if err != nil || o.Window.Mode != "bounded" {
						t.Fatalf("contractual UTC timestamp rejected: %v", err)
					}
				})
			}
		})
	}
}
