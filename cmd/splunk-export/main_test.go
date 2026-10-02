package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/internal/buildinfo"
	"github.com/delgado-jacob/spl-toolkit/internal/splunkexport"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

// Synthetic server with valid origin and denied inventory/configuration still
// provides useful coverage evidence, which must survive the command boundary.
func commandArgs(t *testing.T) []string {
	t.Helper()
	t.Setenv("EXPORT_COMMAND_TOKEN", "synthetic-secret")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/services/server/info" {
			fmt.Fprint(w, `{"entry":[{"content":{"version":"9.4.2","guid":"private-guid"}}]}`)
			return
		}
		http.Error(w, "sensitive failure", 403)
	}))
	t.Cleanup(server.Close)
	return []string{"--management-url", server.URL, "--allow-insecure", "--credential-env", "EXPORT_COMMAND_TOKEN"}
}
func commandSnapshot(t *testing.T, raw []byte) environment.Snapshot {
	t.Helper()
	report, err := environment.ValidateArtifacts(raw, nil)
	if err != nil || report == nil || report.Status == "invalid" {
		t.Fatalf("invalid artifact %v %+v", err, report)
	}
	var snapshot environment.Snapshot
	if err = json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatal(err)
	}
	return snapshot
}
func TestCommandOutputAndReportSeparation(t *testing.T) {
	args := commandArgs(t)
	path := filepath.Join(t.TempDir(), "report.json")
	var stdout, stderr bytes.Buffer
	code := splunkexport.Run(context.Background(), append(args, "--report", path), &stdout, &stderr)
	if code != 3 {
		t.Fatalf("exit %d stderr %s", code, stderr.String())
	}
	snapshot := commandSnapshot(t, stdout.Bytes())
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report splunkexport.ExportReport
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if report.SnapshotDigest != snapshot.Digest || report.ScopeID != snapshot.ScopeID || !report.AllowInsecure || report.Status != "partial" {
		t.Fatal(report)
	}
	if !strings.Contains(stderr.String(), "insecure transport") || strings.Contains(stderr.String(), "sensitive failure") {
		t.Fatal(stderr.String())
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Fatalf("report permissions %v", info.Mode())
	}
	output := filepath.Join(t.TempDir(), "snapshot.json")
	stdout.Reset()
	stderr.Reset()
	if code = splunkexport.Run(context.Background(), append(args, "--output", output), &stdout, &stderr); code != 3 || stdout.Len() != 0 {
		t.Fatalf("file mode stdout %q exit %d", stdout.String(), code)
	}
	raw, err = os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	commandSnapshot(t, raw)
}
func TestCommandPreservesExistingOutputOnFailure(t *testing.T) {
	args := commandArgs(t)
	dir := t.TempDir()
	output := filepath.Join(dir, "snapshot.json")
	if err := os.WriteFile(output, []byte("existing"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EXPORT_COMMAND_TOKEN", "")
	var stdout, stderr bytes.Buffer
	if code := splunkexport.Run(context.Background(), append(args, "--output", output), &stdout, &stderr); code != 2 || stdout.Len() != 0 {
		t.Fatalf("exit %d stdout %q", code, stdout.String())
	}
	raw, _ := os.ReadFile(output)
	if string(raw) != "existing" {
		t.Fatal("existing artifact replaced")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatal("stage leaked")
	}
}
func TestCommandRejectsInputOutputCollision(t *testing.T) {
	for _, mode := range []string{"same", "cleaned", "hardlink", "symlink", "ca", "report"} {
		t.Run(mode, func(t *testing.T) {
			args := commandArgs(t)
			dir := t.TempDir()
			credential := filepath.Join(dir, "credential")
			os.WriteFile(credential, []byte("secret"), 0600)
			output := credential
			args = args[:len(args)-2]
			args = append(args, "--credential-file", credential)
			switch mode {
			case "cleaned":
				output = dir + "/./credential"
			case "hardlink":
				output = filepath.Join(dir, "link")
				if err := os.Link(credential, output); err != nil {
					t.Skip(err)
				}
			case "symlink":
				output = filepath.Join(dir, "link")
				if err := os.Symlink(credential, output); err != nil {
					t.Skip(err)
				}
			case "ca":
				args = commandArgs(t)
				args = append(args, "--ca-file", credential)
			case "report":
				output = filepath.Join(dir, "artifact")
				args = append(args, "--report", output)
			}
			var stdout, stderr bytes.Buffer
			if code := splunkexport.Run(context.Background(), append(args, "--output", output), &stdout, &stderr); code != 2 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "paths conflict") {
				t.Fatalf("collision accepted %d %s", code, stderr.String())
			}
			raw, _ := os.ReadFile(credential)
			if string(raw) != "secret" {
				t.Fatal("credential replaced")
			}
		})
	}
}
func TestCommandHelpNeedsNoCredential(t *testing.T) {
	for _, arg := range []string{"--help", "--version"} {
		var stdout, stderr bytes.Buffer
		if code := splunkexport.Run(context.Background(), []string{arg}, &stdout, &stderr); code != 0 || stderr.Len() != 0 || stdout.Len() == 0 {
			t.Fatalf("%s: exit %d %s", arg, code, stderr.String())
		}
		if arg == "--version" && stdout.String() != buildinfo.Version+"\n" {
			t.Fatal(stdout.String())
		}
	}
}
