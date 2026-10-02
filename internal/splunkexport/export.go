package splunkexport

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/delgado-jacob/spl-toolkit/internal/buildinfo"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

// Export acquires evidence from one origin and returns a detached validated artifact.
// Acquisition gaps are retained as coverage; an unusable origin or artifact is fatal.
func Export(parent context.Context, options Options) (*Result, error) {
	c, err := NewClient(options)
	if err != nil {
		return nil, failure("options_or_credentials_invalid")
	}
	defer c.Close()
	o := c.options
	ctx, cancel := context.WithTimeout(parent, o.OverallTimeout)
	defer cancel()
	start := time.Now().UTC().Format(time.RFC3339Nano)
	var info struct {
		Entries []struct {
			Content struct {
				Version string `json:"version"`
				GUID    string `json:"guid"`
			} `json:"content"`
		} `json:"entry"`
	}
	if err = c.getJSON(ctx, "/services/server/info", url.Values{}, &info); err != nil || len(info.Entries) != 1 || strings.TrimSpace(info.Entries[0].Content.Version) == "" {
		return nil, failure("origin_unavailable")
	}
	instanceID := o.InstanceID
	if guid := strings.TrimSpace(info.Entries[0].Content.GUID); guid != "" {
		instanceID = inventoryHash([]any{"splunk_instance_guid", guid})
	}
	if strings.TrimSpace(instanceID) == "" {
		return nil, failure("origin_identity_unavailable")
	}
	origin := environment.Origin{InstanceID: instanceID, ProductVersion: info.Entries[0].Content.Version, Producer: "spl-toolkit-export", ProducerVersion: buildinfo.Version}
	indexes := c.collectIndexes(ctx, instanceID)
	configuration := c.collectConfiguration(ctx, instanceID)
	if configuration.Err != nil {
		return nil, configuration.Err
	}
	metadata := c.collectMetadata(ctx, instanceID, indexes)
	end := time.Now().UTC().Format(time.RFC3339Nano)
	entropy := make([]byte, 16)
	if _, err = rand.Read(entropy); err != nil {
		return nil, failure("scope_id_failed")
	}
	snapshot := environment.Snapshot{
		SchemaVersion: 2,
		ScopeID:       "capture:" + hex.EncodeToString(entropy),
		CaptureScope:  o.Scope,
		Origin:        origin,
		Capture:       environment.CaptureInterval{Start: start, End: end},
		Capabilities:  []environment.Capability{},
		Collections:   append([]environment.Collection{{Kind: "index", Coverage: indexes.Enumeration.Coverage, Reason: indexes.Enumeration.Reason}}, configuration.Collections...),
		Objects:       append(append(indexes.Objects, configuration.Objects...), metadata.Objects...),
		Observation: &environment.ObservationScope{
			IndexSelection:   o.IndexSelection,
			Enumeration:      indexes.Enumeration,
			Indexes:          indexes.Indexes,
			UnmatchedIndexes: indexes.UnmatchedIndexes,
			Method:           "splunk_metadata",
			Visibility:       "exporting_principal",
			Window:           o.Window,
			TimePrecision:    "bucket_overlap",
			AbsenceMeaning:   "not_observed",
			Captures:         metadata.Captures,
		},
	}
	snapshot.Collections = append(snapshot.Collections, metadata.Collections...)
	prepared, validation, err := environment.PrepareSnapshot(snapshot)
	if err != nil || prepared == nil || exitFor(validation, false) == 2 {
		return nil, failure("artifact_invalid")
	}
	snapshot = prepared.Snapshot()
	snapshot.Digest = validation.SnapshotDigest
	diagnostics := append(append(append([]Diagnostic{}, indexes.Diagnostics...), configuration.Diagnostics...), metadata.Diagnostics...)
	sort.SliceStable(diagnostics, func(i, j int) bool {
		a, b := diagnostics[i], diagnostics[j]
		ak := []string{a.Kind, a.IndexID, a.Datatype, a.Code, a.Severity, a.Message}
		bk := []string{b.Kind, b.IndexID, b.Datatype, b.Code, b.Severity, b.Message}
		for k := range ak {
			if ak[k] != bk[k] {
				return ak[k] < bk[k]
			}
		}
		return false
	})
	code := exitFor(validation, indexes.CleanupFailed || metadata.CleanupFailed)
	status := "complete"
	if code == 3 {
		status = "partial"
	}
	report := ExportReport{
		SchemaVersion:  1,
		Status:         status,
		ScopeID:        snapshot.ScopeID,
		AllowInsecure:  o.AllowInsecure,
		Origin:         snapshot.Origin,
		Capture:        snapshot.Capture,
		SnapshotDigest: snapshot.Digest,
		Observation:    snapshot.Observation,
		Limits: ExportLimits{
			RequestTimeoutSeconds: o.RequestTimeout.Seconds(),
			JobTimeoutSeconds:     o.JobTimeout.Seconds(),
			OverallTimeoutSeconds: o.OverallTimeout.Seconds(),
			MaxRows:               o.MaxRows,
			MaxResponseBytes:      responseLimit,
			MaxArtifactBytes:      assembledLimit,
		},
		Collections: []Acquisition{},
		Diagnostics: diagnostics,
	}
	for _, collection := range snapshot.Collections {
		report.Collections = append(report.Collections, Acquisition{Kind: collection.Kind, Coverage: collection.Coverage, Reason: collection.Reason})
	}
	result := &Result{Snapshot: snapshot, Report: report, ExitCode: code}
	if _, _, err = serializeArtifacts(result); err != nil {
		return nil, err
	}
	return result, nil
}

func serializeArtifacts(result *Result) ([]byte, []byte, error) {
	if result == nil {
		return nil, nil, failure("artifact_invalid")
	}
	snapshot, err := json.Marshal(result.Snapshot)
	if err != nil {
		return nil, nil, failure("artifact_invalid")
	}
	validation, err := environment.ValidateArtifacts(snapshot, nil)
	if err != nil || exitFor(validation, false) == 2 || result.Snapshot.Digest == "" || result.Snapshot.Digest != validation.SnapshotDigest || result.Report.SnapshotDigest != validation.SnapshotDigest || result.Report.ScopeID != result.Snapshot.ScopeID {
		return nil, nil, failure("artifact_invalid")
	}
	report, err := json.Marshal(result.Report)
	if err != nil {
		return nil, nil, failure("artifact_invalid")
	}
	if len(snapshot) > assembledLimit-len(report)-2 {
		return nil, nil, failure("artifact_too_large")
	}
	return append(snapshot, '\n'), append(report, '\n'), nil
}

// Run finishes all acquisition and owned-job cleanup before returning its exit code.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	o, err := ParseOptions(args)
	if errors.Is(err, ErrHelp) {
		n, writeErr := io.WriteString(stdout, helpText)
		if writeErr != nil || n != len(helpText) {
			return 2
		}
		return 0
	}
	if errors.Is(err, ErrVersion) {
		version := buildinfo.Version + "\n"
		n, writeErr := io.WriteString(stdout, version)
		if writeErr != nil || n != len(version) {
			return 2
		}
		return 0
	}
	if err != nil {
		io.WriteString(stderr, "spl-toolkit-export: invalid arguments.\n")
		return 2
	}
	if err = validateArtifactPaths(o); err != nil {
		io.WriteString(stderr, "spl-toolkit-export: artifact paths conflict or are unavailable.\n")
		return 2
	}
	if o.AllowInsecure {
		io.WriteString(stderr, "spl-toolkit-export: warning: insecure transport explicitly enabled.\n")
	}
	result, err := Export(ctx, o)
	if err != nil {
		io.WriteString(stderr, "spl-toolkit-export: acquisition or artifact validation failed.\n")
		return 2
	}
	for _, diagnostic := range result.Report.Diagnostics {
		io.WriteString(stderr, "spl-toolkit-export: "+diagnostic.Code+": "+diagnostic.Message+"\n")
	}
	if err = writeArtifacts(o, result, stdout); err != nil {
		io.WriteString(stderr, "spl-toolkit-export: artifact output failed.\n")
		return 2
	}
	return result.ExitCode
}

const helpText = `Usage: spl-toolkit-export --management-url URL (--credential-env NAME | --credential-file PATH) [options]

  --auth-mode bearer|session       Authentication mode (default bearer)
  --ca-file PATH                   Additional trusted CA certificates
  --allow-insecure                 Explicitly permit HTTP or unverified TLS
  --instance-id ID                 Opaque identity fallback when GUID is absent
  --namespace VALUE               Repeatable exact configuration scope
  --app VALUE                     Repeatable exact app scope
  --owner VALUE                   Repeatable exact owner scope
  --index VALUE                   Repeatable exact observation index selection
  --earliest TIMESTAMP            Absolute UTC observation lower bound
  --latest TIMESTAMP              Absolute UTC observation upper bound
  --request-timeout DURATION       Request timeout (default 30s)
  --job-timeout DURATION           Job timeout (default 2m)
  --overall-timeout DURATION       Acquisition timeout (default 10m)
  --max-rows NUMBER                Maximum rows per discovery job (default 10000)
  --output PATH                    Snapshot file; otherwise snapshot JSON on stdout
  --report PATH                    Separate export report file
  --help                          Show this help without connecting
  --version                       Show producer version without connecting
`

func validateArtifactPaths(o Options) error {
	type checkedPath struct {
		path, base   string
		info, parent os.FileInfo
		output       bool
	}
	paths := []checkedPath{}
	for _, candidate := range []struct {
		path   string
		output bool
	}{{o.Output, true}, {o.ReportOutput, true}, {o.CredentialFile, false}, {o.CAFile, false}} {
		if candidate.path == "" {
			continue
		}
		absolute, err := filepath.Abs(filepath.Clean(candidate.path))
		if err != nil {
			return err
		}
		// Stat the supplied path: cleaning symlink/.. changes its filesystem target.
		info, err := os.Stat(candidate.path)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		directory, base := filepath.Split(candidate.path)
		if directory == "" {
			directory = "."
		}
		parent, err := os.Stat(directory)
		if err != nil {
			return err
		}
		for _, previous := range paths {
			if (candidate.output || previous.output) && (absolute == previous.path ||
				info != nil && previous.info != nil && os.SameFile(info, previous.info) ||
				base == previous.base && os.SameFile(parent, previous.parent)) {
				return failure("artifact_path_collision")
			}
		}
		paths = append(paths, checkedPath{path: absolute, base: base, info: info, parent: parent, output: candidate.output})
	}
	return nil
}

// Each destination receives a validated private stage in its own directory.
// Cross-directory renames are not a transaction: a committed snapshot remains
// usable if the later report commit fails.
func writeArtifacts(o Options, result *Result, stdout io.Writer) error {
	snapshot, report, err := serializeArtifacts(result)
	if err != nil {
		return err
	}
	if err = validateArtifactPaths(o); err != nil {
		return err
	}
	type staged struct{ path, destination string }
	stages := []staged{}
	defer func() {
		for _, stage := range stages {
			os.Remove(stage.path)
		}
	}()
	for _, artifact := range []struct {
		destination string
		data        []byte
	}{{o.Output, snapshot}, {o.ReportOutput, report}} {
		if artifact.destination == "" {
			continue
		}
		// Split preserves symlink/.. semantics; Dir would clean the path.
		directory, _ := filepath.Split(artifact.destination)
		if directory == "" {
			directory = "."
		}
		f, err := os.CreateTemp(directory, ".splunk-export-*")
		if err != nil {
			return err
		}
		stages = append(stages, staged{f.Name(), artifact.destination})
		n, writeErr := f.Write(artifact.data)
		closeErr := f.Close()
		if writeErr != nil {
			return writeErr
		}
		if n != len(artifact.data) {
			return io.ErrShortWrite
		}
		if closeErr != nil {
			return closeErr
		}
	}
	if o.Output == "" {
		n, err := stdout.Write(snapshot)
		if err != nil {
			return err
		}
		if n != len(snapshot) {
			return io.ErrShortWrite
		}
	}
	for _, stage := range stages {
		// Recheck after each commit so initially absent case-insensitive aliases
		// cannot let the report overwrite the newly committed snapshot.
		if err = validateArtifactPaths(o); err != nil {
			return err
		}
		if err = os.Rename(stage.path, stage.destination); err != nil {
			return err
		}
	}
	return nil
}
