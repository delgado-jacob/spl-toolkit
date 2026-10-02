package splunkexport

import (
	"bytes"
	"context"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExitFor(t *testing.T) {
	for _, tc := range []struct {
		r       *environment.Report
		cleanup bool
		want    int
	}{{nil, false, 2}, {&environment.Report{Status: "invalid"}, true, 2}, {&environment.Report{Status: "partial"}, false, 3}, {&environment.Report{Status: "valid"}, true, 3}, {&environment.Report{Status: "valid"}, false, 0}} {
		if got := exitFor(tc.r, tc.cleanup); got != tc.want {
			t.Fatalf("%+v cleanup=%v: %d", tc.r, tc.cleanup, got)
		}
	}
}

func TestWriteArtifactsCombinedBudgetPreservesDestinations(t *testing.T) {
	o, _ := newExportFixture(t)
	r, err := Export(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	o.Output = filepath.Join(dir, "snapshot.json")
	o.ReportOutput = filepath.Join(dir, "report.json")
	for _, path := range []string{o.Output, o.ReportOutput} {
		if err = os.WriteFile(path, []byte("existing"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	r.Report.Diagnostics = append(r.Report.Diagnostics, Diagnostic{Code: "budget", Message: strings.Repeat("x", assembledLimit)})
	var stdout bytes.Buffer
	if err = writeArtifacts(o, r, &stdout); err == nil {
		t.Fatal("oversized combined artifact published")
	}
	if stdout.Len() != 0 {
		t.Fatal("snapshot emitted before budget check")
	}
	for _, path := range []string{o.Output, o.ReportOutput} {
		raw, _ := os.ReadFile(path)
		if string(raw) != "existing" {
			t.Fatal("existing destination replaced")
		}
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 2 {
		t.Fatal("stage leaked")
	}
}
func TestWriteArtifactsRejectsInvalidSnapshot(t *testing.T) {
	o, _ := newExportFixture(t)
	r, err := Export(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	r.Snapshot.Origin.ProductVersion = ""
	var stdout bytes.Buffer
	if err = writeArtifacts(o, r, &stdout); err == nil || stdout.Len() != 0 {
		t.Fatal("invalid artifact published")
	}
}
func TestWriteArtifactsKeepsSnapshotOnReportCommitFailure(t *testing.T) {
	o, _ := newExportFixture(t)
	r, err := Export(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	o.Output = filepath.Join(dir, "snapshot.json")
	o.ReportOutput = filepath.Join(dir, "report-directory")
	os.Mkdir(o.ReportOutput, 0700)
	var stdout bytes.Buffer
	if err = writeArtifacts(o, r, &stdout); err == nil || stdout.Len() != 0 {
		t.Fatal("report rename failure ignored")
	}
	raw, err := os.ReadFile(o.Output)
	if err != nil {
		t.Fatal(err)
	}
	validation, err := environment.ValidateArtifacts(raw, nil)
	if err != nil || validation.Status == "invalid" {
		t.Fatal("committed snapshot invalid")
	}
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".splunk-export-") {
			t.Fatal("stage leaked")
		}
	}
}
func TestWriteArtifactsStagesBothBeforeCommit(t *testing.T) {
	o, _ := newExportFixture(t)
	r, err := Export(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	o.Output = filepath.Join(dir, "snapshot.json")
	o.ReportOutput = filepath.Join(dir, "absent", "report.json")
	os.WriteFile(o.Output, []byte("existing"), 0600)
	var stdout bytes.Buffer
	if err = writeArtifacts(o, r, &stdout); err == nil {
		t.Fatal("report staging failure ignored")
	}
	raw, _ := os.ReadFile(o.Output)
	if string(raw) != "existing" {
		t.Fatal("snapshot replaced before report staging")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatal("stage leaked")
	}
}

type shortArtifactWriter struct{}

func (shortArtifactWriter) Write(p []byte) (int, error) { return len(p) - 1, nil }
func TestWriteArtifactsRejectsShortStdout(t *testing.T) {
	o, _ := newExportFixture(t)
	r, err := Export(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeArtifacts(o, r, shortArtifactWriter{}); err != io.ErrShortWrite {
		t.Fatalf("short output: %v", err)
	}
}
func TestRunOfflineOutputFailures(t *testing.T) {
	for _, arg := range []string{"--help", "--version"} {
		if got := Run(context.Background(), []string{arg}, shortArtifactWriter{}, io.Discard); got != 2 {
			t.Fatalf("short %s output returned %d", arg, got)
		}
	}
}
