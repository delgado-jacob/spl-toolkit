package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
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
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
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

func TestCommandPreservesInputsThroughSymlinkParentTraversal(t *testing.T) {
	for _, input := range []string{"credential", "ca"} {
		t.Run(input, func(t *testing.T) {
			args := commandArgs(t)
			dir := t.TempDir()
			a, b := filepath.Join(dir, "a"), filepath.Join(dir, "b")
			if err := os.MkdirAll(filepath.Join(b, "sub"), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(a, 0700); err != nil {
				t.Fatal(err)
			}
			link := filepath.Join(a, "link")
			if err := os.Symlink(filepath.Join(b, "sub"), link); err != nil {
				t.Skip(err)
			}
			rawParent := link + string(os.PathSeparator) + ".."
			resolvedParent, parentErr := os.Stat(rawParent)
			targetParent, targetErr := os.Stat(b)
			if targetErr != nil {
				t.Fatal(targetErr)
			}
			if parentErr != nil || !os.SameFile(resolvedParent, targetParent) {
				t.Skip("filesystem does not resolve symlink parent traversal to the target parent")
			}
			protected := filepath.Join(b, input)
			inputBytes := []byte("synthetic-input")
			if input == "ca" {
				certificateServer := httptest.NewTLSServer(http.NotFoundHandler())
				t.Cleanup(certificateServer.Close)
				inputBytes = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateServer.Certificate().Raw})
			}
			if err := os.WriteFile(protected, inputBytes, 0600); err != nil {
				t.Fatal(err)
			}
			// Join would clean away the traversal whose filesystem semantics we test.
			output := link + string(os.PathSeparator) + ".." + string(os.PathSeparator) + input
			if input == "credential" {
				args = args[:len(args)-2]
				args = append(args, "--credential-file", protected)
			} else {
				args = append(args, "--ca-file", protected)
			}
			var stdout, stderr bytes.Buffer
			code := splunkexport.Run(context.Background(), append(args, "--output", output), &stdout, &stderr)
			raw, err := os.ReadFile(protected)
			if err != nil || !bytes.Equal(raw, inputBytes) {
				t.Fatalf("%s overwritten via symlink traversal", input)
			}
			if code != 2 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "paths conflict") {
				t.Fatalf("collision not rejected: exit=%d stderr=%s", code, stderr.String())
			}
			for _, parent := range []string{a, b} {
				entries, _ := os.ReadDir(parent)
				for _, entry := range entries {
					if strings.HasPrefix(entry.Name(), ".splunk-export-") {
						t.Fatal("stage leaked")
					}
				}
			}
		})
	}
}
func TestCommandRejectsAbsentDestinationsWithAliasedParents(t *testing.T) {
	args := commandArgs(t)
	dir := t.TempDir()
	real := filepath.Join(dir, "real")
	if err := os.Mkdir(real, 0700); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(dir, "alias")
	if err := os.Symlink(real, alias); err != nil {
		t.Skip(err)
	}
	output, report := filepath.Join(real, "artifact.json"), filepath.Join(alias, "artifact.json")
	var stdout, stderr bytes.Buffer
	code := splunkexport.Run(context.Background(), append(args, "--output", output, "--report", report), &stdout, &stderr)
	if code != 2 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "paths conflict") {
		t.Fatalf("absent aliases accepted: exit=%d stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("artifact published despite alias")
	}
	entries, _ := os.ReadDir(real)
	if len(entries) != 0 {
		t.Fatal("stage leaked")
	}
}
func TestCommandKeepsSnapshotForCaseInsensitiveDestinationAliases(t *testing.T) {
	dir := t.TempDir()
	probe := filepath.Join(dir, "CaseProbe")
	if err := os.WriteFile(probe, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "caseprobe")); err != nil {
		t.Skip("case-sensitive filesystem")
	}
	os.Remove(probe)
	args := commandArgs(t)
	output, report := filepath.Join(dir, "Artifact.json"), filepath.Join(dir, "artifact.json")
	var stdout, stderr bytes.Buffer
	code := splunkexport.Run(context.Background(), append(args, "--output", output, "--report", report), &stdout, &stderr)
	if code != 2 || stdout.Len() != 0 {
		t.Fatalf("case aliases accepted: exit=%d stderr=%s", code, stderr.String())
	}
	if raw, err := os.ReadFile(output); err == nil {
		commandSnapshot(t, raw)
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".splunk-export-") {
			t.Fatal("stage leaked")
		}
	}
}

func TestCommandStagesInFilesystemDestinationDirectory(t *testing.T) {
	args := commandArgs(t)
	dir := t.TempDir()
	a, b := filepath.Join(dir, "a"), filepath.Join(dir, "b")
	if err := os.MkdirAll(filepath.Join(b, "sub"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(a, 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(a, "link")
	if err := os.Symlink(filepath.Join(b, "sub"), link); err != nil {
		t.Skip(err)
	}
	rawParent := link + string(os.PathSeparator) + ".."
	resolvedParent, parentErr := os.Stat(rawParent)
	targetParent, targetErr := os.Stat(b)
	if targetErr != nil {
		t.Fatal(targetErr)
	}
	if parentErr != nil || !os.SameFile(resolvedParent, targetParent) {
		t.Skip("filesystem does not resolve symlink parent traversal to the target parent")
	}
	if err := os.Chmod(a, 0500); err != nil {
		t.Skip(err)
	}
	t.Cleanup(func() { os.Chmod(a, 0700) })
	if probe, err := os.CreateTemp(a, "permission-probe-*"); err == nil {
		probe.Close()
		os.Remove(probe.Name())
		t.Skip("host does not enforce directory permissions")
	}
	output := link + string(os.PathSeparator) + ".." + string(os.PathSeparator) + "snapshot.json"
	var stdout, stderr bytes.Buffer
	code := splunkexport.Run(context.Background(), append(args, "--output", output), &stdout, &stderr)
	if code != 3 || stdout.Len() != 0 {
		t.Fatalf("actual destination publication failed: exit=%d stderr=%s", code, stderr.String())
	}
	raw, err := os.ReadFile(filepath.Join(b, "snapshot.json"))
	if err != nil {
		t.Fatal(err)
	}
	commandSnapshot(t, raw)
	for _, parent := range []string{a, b} {
		entries, _ := os.ReadDir(parent)
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".splunk-export-") {
				t.Fatal("stage leaked")
			}
		}
	}
}
