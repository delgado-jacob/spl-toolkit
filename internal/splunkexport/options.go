// Package splunkexport acquires bounded evidence from one authenticated Splunk origin.
package splunkexport

import (
	"errors"
	"flag"
	"io"
	"math"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

var ErrHelp = errors.New("exporter help requested")
var ErrVersion = errors.New("exporter version requested")

// time.Parse accepts a single-digit hour, comma fractions, and truncates excess
// fraction digits; acquisition inputs must preserve the stricter CLI contract.
var windowTimestampPattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]{1,9})?(Z|[+-]00:00)$`)

type Options struct {
	ManagementURL                                               string
	AllowInsecure                                               bool
	AuthMode, CredentialEnv, CredentialFile, CAFile, InstanceID string
	Scope                                                       environment.CaptureScope
	IndexSelection                                              environment.Selector
	Window                                                      environment.ObservationWindow
	RequestTimeout, JobTimeout, OverallTimeout                  time.Duration
	MaxRows                                                     int
	Output, ReportOutput                                        string
}

const (
	responseLimit         = 8 << 20
	assembledLimit        = 64 << 20
	resultPageSize        = 500
	configurationPageSize = 100
	pollInterval          = 250 * time.Millisecond
	cleanupTimeout        = 10 * time.Second
)

type repeated []string

func (v *repeated) String() string     { return "" }
func (v *repeated) Set(s string) error { *v = append(*v, s); return nil }
func selector(values []string) (environment.Selector, error) {
	if len(values) == 0 {
		all := true
		return environment.Selector{All: &all}, nil
	}
	seen := map[string]bool{}
	out := []string{}
	for _, v := range values {
		if strings.TrimSpace(v) == "" || strings.ContainsAny(v, "\r\n\x00") {
			return environment.Selector{}, errors.New("invalid selector")
		}
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	sort.Strings(out)
	return environment.Selector{Values: out}, nil
}
func ParseOptions(args []string) (Options, error) {
	o := Options{AuthMode: "bearer", RequestTimeout: 30 * time.Second, JobTimeout: 120 * time.Second, OverallTimeout: 10 * time.Minute, MaxRows: 10000}
	f := flag.NewFlagSet("splunk-export", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	var namespaces, apps, owners, indexes repeated
	var help, version bool
	f.StringVar(&o.ManagementURL, "management-url", "", "")
	f.BoolVar(&o.AllowInsecure, "allow-insecure", false, "")
	f.StringVar(&o.AuthMode, "auth-mode", "bearer", "")
	f.StringVar(&o.CredentialEnv, "credential-env", "", "")
	f.StringVar(&o.CredentialFile, "credential-file", "", "")
	f.StringVar(&o.CAFile, "ca-file", "", "")
	f.StringVar(&o.InstanceID, "instance-id", "", "")
	f.Var(&namespaces, "namespace", "")
	f.Var(&apps, "app", "")
	f.Var(&owners, "owner", "")
	f.Var(&indexes, "index", "")
	f.StringVar(&o.Window.Earliest, "earliest", "", "")
	f.StringVar(&o.Window.Latest, "latest", "", "")
	f.DurationVar(&o.RequestTimeout, "request-timeout", o.RequestTimeout, "")
	f.DurationVar(&o.JobTimeout, "job-timeout", o.JobTimeout, "")
	f.DurationVar(&o.OverallTimeout, "overall-timeout", o.OverallTimeout, "")
	f.IntVar(&o.MaxRows, "max-rows", o.MaxRows, "")
	f.StringVar(&o.Output, "output", "", "")
	f.StringVar(&o.ReportOutput, "report", "", "")
	f.BoolVar(&help, "help", false, "")
	f.BoolVar(&version, "version", false, "")
	if err := f.Parse(args); err != nil || f.NArg() != 0 {
		return Options{}, errors.New("invalid exporter arguments")
	}
	if help {
		return Options{}, ErrHelp
	}
	if version {
		return Options{}, ErrVersion
	}
	var err error
	if o.Scope.Namespace, err = selector(namespaces); err != nil {
		return Options{}, err
	}
	if o.Scope.App, err = selector(apps); err != nil {
		return Options{}, err
	}
	if o.Scope.Owner, err = selector(owners); err != nil {
		return Options{}, err
	}
	if o.IndexSelection, err = selector(indexes); err != nil {
		return Options{}, err
	}
	if o.Window.Earliest == "" && o.Window.Latest == "" {
		o.Window.Mode = "all_retained"
	} else {
		o.Window.Mode = "bounded"
	}
	if err = o.validate(); err != nil {
		return Options{}, err
	}
	o.normalizeWindow()
	return o, nil
}
func (o Options) validate() error {
	u, err := url.Parse(o.ManagementURL)
	if err != nil || u == nil || u.Host == "" || u.Opaque != "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Path != "" && u.Path != "/" || u.RawPath != "" || u.Scheme != "https" && !(u.Scheme == "http" && o.AllowInsecure) {
		return errors.New("invalid management origin; verified HTTPS is required unless allow-insecure is enabled")
	}
	if (o.CredentialEnv == "") == (o.CredentialFile == "") {
		return errors.New("exactly one credential source is required")
	}
	if o.AuthMode != "bearer" && o.AuthMode != "session" {
		return errors.New("invalid authentication mode")
	}
	if o.RequestTimeout <= 0 || o.JobTimeout <= 0 || o.OverallTimeout <= 0 || o.MaxRows <= 0 || o.MaxRows == math.MaxInt {
		return errors.New("timeouts and row limit must be positive and representable")
	}
	for _, s := range []environment.Selector{o.Scope.App, o.Scope.Owner} {
		for _, v := range s.Values {
			if v == "*" || v == "-" {
				return errors.New("reserved wildcard selector is not an exact value")
			}
		}
	}
	var earliest, latest time.Time
	for _, bound := range []struct {
		value  string
		target *time.Time
	}{{o.Window.Earliest, &earliest}, {o.Window.Latest, &latest}} {
		if bound.value == "" {
			continue
		}
		if !windowTimestampPattern.MatchString(bound.value) {
			return errors.New("window bounds must be absolute UTC timestamps")
		}
		parsed, err := time.Parse(time.RFC3339Nano, bound.value)
		_, offset := parsed.Zone()
		if err != nil || offset != 0 {
			return errors.New("window bounds must be absolute UTC timestamps")
		}
		*bound.target = parsed
	}
	if o.Window.Earliest != "" && o.Window.Latest != "" && !earliest.Before(latest) {
		return errors.New("window bounds must be in increasing order")
	}
	return nil
}

func (o *Options) normalizeWindow() {
	o.Window.Mode = "all_retained"
	if o.Window.Earliest != "" || o.Window.Latest != "" {
		o.Window.Mode = "bounded"
	}
	for _, bound := range []*string{&o.Window.Earliest, &o.Window.Latest} {
		if *bound != "" {
			parsed, _ := time.Parse(time.RFC3339Nano, *bound)
			*bound = parsed.UTC().Format(time.RFC3339Nano)
		}
	}
}
