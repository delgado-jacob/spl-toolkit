package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
)

func closureCLIFile(t *testing.T, name string, value any) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestClosureCLIFormatsSourcesAndTransport(t *testing.T) {
	bundle := closure.DefinitionBundle{SchemaVersion: 1, ScopeID: "synthetic", Collections: []closure.Collection{{Kind: "lookup", Coverage: "complete"}}, Objects: []closure.Definition{{ID: "users", Kind: "lookup", Name: "users", SourceID: "lookup-source", Relations: []closure.Relation{}}}}
	path := closureCLIFile(t, "bundle.json", bundle)
	query := "| lookup users user OUTPUT role"
	req, err := closure.DecodeRequest([]byte(`{"schema_version":1,"document":{"text":"| lookup users user OUTPUT role"},"bundle":{"schema_version":1,"scope_id":"synthetic","collections":[{"kind":"lookup","coverage":"complete"}],"objects":[{"id":"users","kind":"lookup","name":"users","source_id":"lookup-source"}]}}`))
	if err != nil {
		t.Fatal(err)
	}
	want, err := closure.Evaluate(req)
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{"inline", "flag", "file", "stdin"} {
		t.Run(source, func(t *testing.T) {
			args := []string{"closure", "--bundle", path, "--format=json"}
			switch source {
			case "inline":
				args = append(args, query)
			case "flag":
				args = append(args, "--query", query)
			case "file":
				input := filepath.Join(t.TempDir(), "query.spl")
				if err := os.WriteFile(input, []byte(query), 0600); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--file", input)
			case "stdin":
				args = append(args, "--stdin")
			}
			var out, errOut bytes.Buffer
			code := runCLIWithInput(args, strings.NewReader(query), &out, &errOut)
			var got closure.Report
			if err := json.Unmarshal(out.Bytes(), &got); err != nil {
				t.Fatalf("decode: %v, stderr=%s", err, &errOut)
			}
			if code != 0 || errOut.Len() != 0 || !reflect.DeepEqual(&got, want) {
				t.Fatalf("code=%d parity=%t stderr=%s", code, reflect.DeepEqual(&got, want), &errOut)
			}
		})
	}
	// Explicit bindings are local input files, but their path is absent from the report.
	digest := sha256.Sum256([]byte(query))
	binding := closure.Binding{DocumentDigest: fmt.Sprintf("sha256:%x", digest), Kind: "lookup", Start: 9, End: 14, ObjectID: "users"}
	bindingsPath := closureCLIFile(t, "bindings.json", []closure.Binding{binding})
	boundRequest := req
	boundRequest.Bindings = []closure.Binding{binding}
	boundWant, err := closure.Evaluate(boundRequest)
	if err != nil {
		t.Fatal(err)
	}
	var boundOut, boundErr bytes.Buffer
	boundCode := runCLIWithInput([]string{"closure", "--bundle", path, "--bindings", bindingsPath, "--query", query, "--format=json"}, strings.NewReader(""), &boundOut, &boundErr)
	var boundGot closure.Report
	if err := json.Unmarshal(boundOut.Bytes(), &boundGot); err != nil {
		t.Fatal(err)
	}
	if boundCode != 0 || boundErr.Len() != 0 || !reflect.DeepEqual(&boundGot, boundWant) {
		t.Fatalf("bound parity=%t code=%d stderr=%s", reflect.DeepEqual(&boundGot, boundWant), boundCode, &boundErr)
	}
	for _, format := range []string{"graph", "bom", "text"} {
		var out, errOut bytes.Buffer
		code := runCLIWithInput([]string{"closure", "--bundle", path, "--query", query, "--format", format}, strings.NewReader(""), &out, &errOut)
		if code != 0 || errOut.Len() != 0 {
			t.Fatalf("format=%s code=%d stderr=%s", format, code, &errOut)
		}
		switch format {
		case "text":
			if out.String() != closure.FormatInventory(want) {
				t.Fatalf("text=%s", &out)
			}
		case "graph":
			var got closure.DependencyGraph
			if err := json.Unmarshal(out.Bytes(), &got); err != nil || !reflect.DeepEqual(got, want.Graph) {
				t.Fatalf("graph=%s err=%v", &out, err)
			}
		case "bom":
			var got []closure.BOMEntry
			if err := json.Unmarshal(out.Bytes(), &got); err != nil || !reflect.DeepEqual(got, want.BOM) {
				t.Fatalf("bom=%s err=%v", &out, err)
			}
		}
	}
}

func TestClosureCLIRejectsRequestAndOutputErrors(t *testing.T) {
	bundle := closureCLIFile(t, "bundle.json", closure.DefinitionBundle{SchemaVersion: 1, ScopeID: "synthetic", Collections: []closure.Collection{}, Objects: []closure.Definition{}})
	cases := [][]string{{"closure"}, {"closure", "--bundle", bundle}, {"closure", "--bundle", bundle, "--stdin", "x"}, {"closure", "--bundle", bundle, "--file=-"}, {"closure", "--bundle", bundle, "--bindings=-", "x"}, {"closure", "--bundle", bundle, "--format=sarif", "x"}, {"closure", "--bundle", bundle, "--output", bundle, "x"}}
	for _, args := range cases {
		var out, errOut bytes.Buffer
		if code := runCLIWithInput(args, strings.NewReader(""), &out, &errOut); code != 2 || out.Len() != 0 || errOut.Len() == 0 {
			t.Errorf("args=%q code=%d stdout=%s stderr=%s", args, code, &out, &errOut)
		}
	}
}
