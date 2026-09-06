package mapper

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"testing"
)

type baselineFixture struct {
	Version string         `json:"version"`
	Cases   []baselineCase `json:"cases"`
}

type baselineCase struct {
	ID        string                 `json:"id"`
	Query     string                 `json:"query"`
	Config    *MappingConfig         `json:"config,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Mapped    *string                `json:"mapped,omitempty"`
	Discovery *QueryInfo             `json:"discovery,omitempty"`
}

func loadBaselineFixture(t *testing.T) baselineFixture {
	t.Helper()
	payload, err := os.ReadFile("../../testdata/baseline/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture baselineFixture
	if err := json.Unmarshal(payload, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Version != "1" {
		t.Fatalf("fixture version = %q, want 1", fixture.Version)
	}
	return fixture
}

func TestBaselineSurfaceFixtures(t *testing.T) {
	fixture := loadBaselineFixture(t)
	if len(fixture.Cases) == 0 {
		t.Fatal("baseline fixture has no cases")
	}
	for _, testCase := range fixture.Cases {
		t.Run(testCase.ID, func(t *testing.T) {
			m := NewWithConfig(testCase.Config)
			if testCase.Mapped != nil {
				var got string
				var err error
				if testCase.Context != nil {
					got, err = m.MapQueryWithContext(testCase.Query, testCase.Context)
				} else {
					got, err = m.MapQuery(testCase.Query)
				}
				if err != nil {
					t.Fatal(err)
				}
				if got != *testCase.Mapped {
					t.Fatalf("mapped query = %q, want %q", got, *testCase.Mapped)
				}
			}
			if testCase.Discovery != nil {
				got, err := m.DiscoverQuery(testCase.Query)
				if err != nil {
					t.Fatal(err)
				}
				normalizeQueryInfo(got)
				normalizeQueryInfo(testCase.Discovery)
				if !reflect.DeepEqual(got, testCase.Discovery) {
					t.Fatalf("discovery = %#v, want %#v", got, testCase.Discovery)
				}
			}
		})
	}
}

func normalizeQueryInfo(info *QueryInfo) {
	slices := []*[]string{
		&info.DataModels, &info.Datasets, &info.Lookups, &info.Macros,
		&info.Sources, &info.SourceTypes, &info.InputFields,
	}
	for _, values := range slices {
		if *values == nil {
			*values = []string{}
		}
		sort.Strings(*values)
	}
}
