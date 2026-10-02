package splunkexport

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/delgado-jacob/spl-toolkit/internal/jsoninput"
)

type Diagnostic struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Kind     string `json:"kind,omitempty"`
	IndexID  string `json:"index_id,omitempty"`
	Datatype string `json:"datatype,omitempty"`
	Message  string `json:"message"`
}

// requestFailure deliberately contains no URLs, credentials, bodies, or server messages.
type requestFailure struct {
	code   string
	status int
}

func (e *requestFailure) Error() string { return e.code }
func failure(code string) error         { return &requestFailure{code: code} }

type Client struct {
	origin        *url.URL
	authorization string
	transport     *http.Transport
	http          *http.Client
	options       Options
	ownedMu       sync.Mutex
	owned         map[string]bool
}

func NewClient(o Options) (*Client, error) {
	if o.RequestTimeout == 0 {
		o.RequestTimeout = 30 * time.Second
	}
	if o.JobTimeout == 0 {
		o.JobTimeout = 120 * time.Second
	}
	if o.OverallTimeout == 0 {
		o.OverallTimeout = 10 * time.Minute
	}
	if o.MaxRows == 0 {
		o.MaxRows = 10000
	}
	if o.AuthMode == "" {
		o.AuthMode = "bearer"
	}
	if err := o.validate(); err != nil {
		return nil, err
	}
	o.normalizeWindow()
	token := ""
	if o.CredentialEnv != "" {
		token = os.Getenv(o.CredentialEnv)
	} else {
		f, err := os.Open(o.CredentialFile)
		if err != nil {
			return nil, errors.New("credential file unavailable")
		}
		data, err := io.ReadAll(io.LimitReader(f, 64*1024+1))
		f.Close()
		if err != nil || len(data) > 64*1024 {
			return nil, errors.New("credential file invalid")
		}
		// A credential file may end with one conventional line terminator.
		token = strings.TrimSuffix(strings.TrimSuffix(string(data), "\n"), "\r")
	}
	if strings.TrimSpace(token) == "" || strings.ContainsAny(token, "\r\n\x00") || len(token) > 64*1024 {
		return nil, errors.New("credential must be nonblank and single line")
	}
	for _, b := range []byte(token) {
		if b < 0x20 || b == 0x7f {
			return nil, errors.New("credential contains invalid header characters")
		}
	}
	pool, err := x509.SystemCertPool()
	if err != nil {
		pool = x509.NewCertPool()
	}
	if o.CAFile != "" {
		f, err := os.Open(o.CAFile)
		if err != nil {
			return nil, errors.New("CA file unavailable")
		}
		data, err := io.ReadAll(io.LimitReader(f, responseLimit+1))
		f.Close()
		if err != nil || len(data) > responseLimit || !pool.AppendCertsFromPEM(data) {
			return nil, errors.New("CA file invalid")
		}
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: pool, InsecureSkipVerify: o.AllowInsecure} // Explicit user opt-in covers HTTP and unverified TLS.
	u, _ := url.Parse(o.ManagementURL)
	u.Path = ""
	prefix := "Bearer "
	if o.AuthMode == "session" {
		prefix = "Splunk "
	}
	c := &Client{origin: u, authorization: prefix + token, transport: tr, options: o, owned: map[string]bool{}}
	c.http = &http.Client{Transport: tr, Timeout: o.RequestTimeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return c, nil
}
func (c *Client) Close() { c.transport.CloseIdleConnections() }
func (c *Client) getJSON(ctx context.Context, path string, query url.Values, out any) error {
	return c.requestJSON(ctx, http.MethodGet, path, query, nil, out)
}
func (c *Client) postFormJSON(ctx context.Context, path string, form url.Values, out any) error {
	return c.requestJSON(ctx, http.MethodPost, path, nil, form, out)
}
func (c *Client) requestJSON(ctx context.Context, method, path string, query, form url.Values, out any) error {
	if !strings.HasPrefix(path, "/services/") && !strings.HasPrefix(path, "/servicesNS/") {
		return failure("request_path_invalid")
	}
	if strings.ContainsAny(path, "?#") {
		return failure("request_path_invalid")
	}
	// Path is already escaped by the fixed endpoint constructor. RawPath prevents double escaping.
	decoded, err := url.PathUnescape(path)
	if err != nil {
		return failure("request_path_invalid")
	}
	u := *c.origin
	u.Path = decoded
	u.RawPath = path
	q := url.Values{}
	for k, v := range query {
		q[k] = append([]string{}, v...)
	}
	q.Set("output_mode", "json")
	u.RawQuery = q.Encode()
	var body io.Reader
	if form != nil {
		copyForm := url.Values{}
		for k, v := range form {
			copyForm[k] = append([]string{}, v...)
		}
		copyForm.Set("output_mode", "json")
		body = strings.NewReader(copyForm.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return failure("request_invalid")
	}
	req.Header.Set("Authorization", c.authorization)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	res, err := c.http.Do(req)
	if err != nil {
		var timeout net.Error
		if errors.As(err, &timeout) && timeout.Timeout() {
			return failure("request_timeout")
		}
		return failure("request_failed")
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return &requestFailure{code: "response_status", status: res.StatusCode}
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, responseLimit+1))
	if err != nil {
		var timeout net.Error
		if errors.As(err, &timeout) && timeout.Timeout() {
			return failure("request_timeout")
		}
		return failure("response_read_failed")
	}
	if len(data) > responseLimit {
		return failure("response_too_large")
	}
	if len(data) == 0 && out == nil {
		return nil
	}
	if jsoninput.ValidateUnicode(data) != nil || !json.Valid(data) {
		return failure("response_json_invalid")
	}
	if out != nil && json.Unmarshal(data, out) != nil {
		return failure("response_json_invalid")
	}
	return nil
}
