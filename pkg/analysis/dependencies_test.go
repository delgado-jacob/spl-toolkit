package analysis

import (
	"reflect"
	"testing"
)

// Catches lost typed catalog roles and accidental field inference from opaque arguments.
func TestCorpusDependencies(t *testing.T) {
	for _, tc := range []struct {
		query string
		want  Dependencies
		refs  []struct{ kind, name, original string }
	}{
		{`index=main source="/var/log/a" sourcetype=syslog`, Dependencies{Indexes: []string{"main"}, Sources: []string{"/var/log/a"}, SourceTypes: []string{"syslog"}}, nil},
		{`| inputlookup users`, Dependencies{Lookups: []string{"users"}}, nil},
		{`| datamodel Web All_Traffic search`, Dependencies{DataModels: []string{"Web"}, Datasets: []string{"Web.All_Traffic"}}, []struct{ kind, name, original string }{{"data_model", "Web", "Web"}, {"dataset", "Web.All_Traffic", "All_Traffic"}}},
		{`| from datamodel:Network_Traffic.All_Traffic`, Dependencies{DataModels: []string{"Network_Traffic"}, Datasets: []string{"Network_Traffic.All_Traffic"}}, []struct{ kind, name, original string }{{"data_model", "Network_Traffic", "Network_Traffic"}, {"dataset", "Network_Traffic.All_Traffic", "Network_Traffic.All_Traffic"}}},
		{`| from savedsearch:Daily`, Dependencies{Datasets: []string{"savedsearch:Daily"}}, nil},
		{`| tstats count FROM datamodel=Authentication.Authentication WHERE nodename=Authentication.Authentication BY user`, Dependencies{DataModels: []string{"Authentication"}, Datasets: []string{"Authentication.Authentication"}}, nil},
		{`| tstats count FROM datamodel=Authentication`, Dependencies{DataModels: []string{"Authentication"}}, nil},
	} {
		t.Run(tc.query, func(t *testing.T) {
			r, _ := Analyze(QueryDocument{Text: tc.query})
			if !r.Coverage.SyntaxComplete {
				t.Fatal(r.Diagnostics)
			}
			got := r.Dependencies
			// Only empty collection representation is normalized in these focused assertions.
			pairs := [][2][]string{{got.Indexes, tc.want.Indexes}, {got.Sources, tc.want.Sources}, {got.SourceTypes, tc.want.SourceTypes}, {got.Datasets, tc.want.Datasets}, {got.Lookups, tc.want.Lookups}, {got.DataModels, tc.want.DataModels}, {got.Macros, tc.want.Macros}}
			for _, pair := range pairs {
				if !reflect.DeepEqual(append([]string{}, pair[0]...), append([]string{}, pair[1]...)) {
					t.Fatalf("dependencies %+v want %+v", got, tc.want)
				}
			}
			for _, want := range tc.refs {
				ref := scopedReference(t, r, "scope-0", want.name, "read")
				if ref.Kind != want.kind || ref.OriginalName != want.original {
					t.Fatal(ref, want)
				}
			}
			if len(tc.want.DataModels) > 0 || len(tc.want.Datasets) > 0 {
				if r.Status != Incomplete {
					t.Fatal(r.Status)
				}
				for _, ref := range r.References {
					if ref.Kind == "field" {
						t.Fatal("opaque argument invented field", ref)
					}
				}
			}
		})
	}
}
func TestCorpusOpaqueDoesNotInventDependencies(t *testing.T) {
	r, _ := Analyze(QueryDocument{Text: `search * | mystery datamodel=Fake index=wrong source="no" | head 1`})
	if r.Status != Incomplete || len(r.References) != 0 {
		t.Fatal(r)
	}
}

func TestCorpusTstatsSearchDependencies(t *testing.T) {
	r, _ := Analyze(QueryDocument{Text: `| tstats count WHERE index=main source="/var/log/a" sourcetype=syslog BY user`})
	if r.Status != Incomplete || !r.Coverage.SyntaxComplete {
		t.Fatal(r)
	}
	if !reflect.DeepEqual(r.Dependencies.Indexes, []string{"main"}) || !reflect.DeepEqual(r.Dependencies.Sources, []string{"/var/log/a"}) || !reflect.DeepEqual(r.Dependencies.SourceTypes, []string{"syslog"}) {
		t.Fatal(r.Dependencies)
	}
	for _, ref := range r.References {
		if ref.Kind == "field" {
			t.Fatal("unmodeled field inference", ref)
		}
	}
}

func TestCorpusDatamodelSearchModeIsNotDataset(t *testing.T) {
	r, _ := Analyze(QueryDocument{Text: `| datamodel Web search`})
	if r.Status != Incomplete || !r.Coverage.SyntaxComplete || len(r.Dependencies.Datasets) != 0 || !reflect.DeepEqual(r.Dependencies.DataModels, []string{"Web"}) {
		t.Fatal(r)
	}
}

// Quoted search values remain patterns; quoted field identifiers do not.
func TestCorpusQuotedDependencyPatterns(t *testing.T) {
	r, _ := Analyze(QueryDocument{Text: `index="ma*" source="/var/log/*"`})
	if len(r.References) != 2 {
		t.Fatal(r)
	}
	for _, ref := range r.References {
		if ref.Resolution != "wildcard" {
			t.Fatal(ref)
		}
	}
}

func TestCorpusQualifiedDependencyBounds(t *testing.T) {
	r, _ := Analyze(QueryDocument{Text: `| from datamodel:模型.数据`})
	if len(r.References) != 2 || r.References[0].OriginalName != "模型" || r.References[1].OriginalName != "模型.数据" || r.References[0].Location.Start.Offset != 17 || r.References[0].Location.End.Offset != 23 || r.References[1].Location.End.Offset != 30 {
		t.Fatal(r)
	}
	for _, q := range []string{`| from datamodel:`, `| from datamodel:.Bad`, `| from datamodel:Model.`, `| from datamodel:Model.Child.More`, `| from "datamodel:Model.Ch\ild"`} {
		r, _ := Analyze(QueryDocument{Text: q})
		if r.Status != Incomplete || len(r.Dependencies.DataModels) != 0 || len(r.Dependencies.Datasets) != 0 || len(r.References) != 0 {
			t.Fatal("invented malformed qualification", r)
		}
	}
}
