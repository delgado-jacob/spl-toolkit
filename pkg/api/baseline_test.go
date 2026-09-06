package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

type apiBaselineFixture struct {
	Cases []apiBaselineCase `json:"cases"`
}

type apiBaselineCase struct {
	ID        string                 `json:"id"`
	Query     string                 `json:"query"`
	Config    *mapper.MappingConfig  `json:"config,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Mapped    *string                `json:"mapped,omitempty"`
	Discovery *mapper.QueryInfo      `json:"discovery,omitempty"`
}

func loadAPIBaselineFixture(t *testing.T) apiBaselineFixture {
	t.Helper()
	payload, err := os.ReadFile("../../testdata/baseline/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture apiBaselineFixture
	if err := json.Unmarshal(payload, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func postBaselineJSON(t *testing.T, client *http.Client, url string, request, response interface{}) int {
	t.Helper()
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode
}

func postBaselineMap(client *http.Client, url string, request MapQueryRequest) (int, MapQueryResponse, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return 0, MapQueryResponse{}, err
	}
	resp, err := client.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return 0, MapQueryResponse{}, err
	}
	defer resp.Body.Close()
	var response MapQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return resp.StatusCode, MapQueryResponse{}, err
	}
	return resp.StatusCode, response, nil
}

func TestHTTPBaselineParity(t *testing.T) {
	server := httptest.NewServer(NewServerWithVersion("0.1.1").Handler())
	defer server.Close()
	fixture := loadAPIBaselineFixture(t)

	for _, testCase := range fixture.Cases {
		t.Run(testCase.ID, func(t *testing.T) {
			if testCase.Mapped != nil {
				request := MapQueryRequest{Query: testCase.Query, Context: testCase.Context, Config: testCase.Config}
				var response MapQueryResponse
				status := postBaselineJSON(t, server.Client(), server.URL+"/api/v1/query/map", request, &response)
				if status != http.StatusOK || !response.Success || response.MappedQuery != *testCase.Mapped {
					t.Fatalf("status=%d response=%#v want mapped query %q", status, response, *testCase.Mapped)
				}
			}
			if testCase.Discovery != nil {
				var response DiscoverQueryResponse
				status := postBaselineJSON(t, server.Client(), server.URL+"/api/v1/query/discover", DiscoverQueryRequest{Query: testCase.Query}, &response)
				normalizeAPIQueryInfo(response.QueryInfo)
				normalizeAPIQueryInfo(testCase.Discovery)
				if status != http.StatusOK || !response.Success || !reflect.DeepEqual(response.QueryInfo, testCase.Discovery) {
					t.Fatalf("status=%d response=%#v want=%#v", status, response, testCase.Discovery)
				}
			}
		})
	}
}

func normalizeAPIQueryInfo(info *mapper.QueryInfo) {
	if info == nil {
		return
	}
	slices := []*[]string{&info.DataModels, &info.Datasets, &info.Lookups, &info.Macros, &info.Sources, &info.SourceTypes, &info.InputFields}
	for _, values := range slices {
		if *values == nil {
			*values = []string{}
		}
		sort.Strings(*values)
	}
}

func TestHTTPMapConfigurationsDoNotLeak(t *testing.T) {
	server := httptest.NewServer(NewServerWithVersion("0.1.1").Handler())
	defer server.Close()
	configs := []*mapper.MappingConfig{
		{Version: "1.0", Mappings: []mapper.FieldMapping{{Source: "src_ip", Target: "left_ip"}}},
		{Version: "1.0", Mappings: []mapper.FieldMapping{{Source: "src_ip", Target: "right_ip"}}},
	}
	var wait sync.WaitGroup
	for i := 0; i < 40; i++ {
		for _, config := range configs {
			config := config
			wait.Add(1)
			go func() {
				defer wait.Done()
				want := "search " + config.Mappings[0].Target + "=1"
				status, response, err := postBaselineMap(server.Client(), server.URL+"/api/v1/query/map", MapQueryRequest{Query: "search src_ip=1", Config: config})
				if err != nil {
					t.Errorf("request failed: %v", err)
					return
				}
				if status != http.StatusOK || response.MappedQuery != want {
					t.Errorf("status=%d mapped=%q want=%q", status, response.MappedQuery, want)
				}
			}()
		}
	}
	wait.Wait()
}

func TestHTTPBaselineErrorsAndDefaults(t *testing.T) {
	server := httptest.NewServer(NewServerWithVersion("0.1.1").Handler())
	defer server.Close()

	invalidRegex := MapQueryRequest{Query: "search src_ip=1", Config: &mapper.MappingConfig{Version: "1.0", Rules: []mapper.ConditionalRule{{ID: "bad", Enabled: true, Conditions: []mapper.Condition{{Type: "source", Operator: "regex", Value: "["}}, Mappings: []mapper.FieldMapping{{Source: "src_ip", Target: "source_ip"}}}}}}
	var configError ErrorResponse
	if status := postBaselineJSON(t, server.Client(), server.URL+"/api/v1/query/map", invalidRegex, &configError); status != http.StatusBadRequest {
		t.Fatalf("invalid regex status=%d response=%#v", status, configError)
	}

	var syntaxResponse ValidateQueryResponse
	if status := postBaselineJSON(t, server.Client(), server.URL+"/api/v1/query/validate", ValidateQueryRequest{Query: "|"}, &syntaxResponse); status != http.StatusUnprocessableEntity || syntaxResponse.Valid {
		t.Fatalf("syntax status=%d response=%#v", status, syntaxResponse)
	}

	var fallback MapQueryResponse
	if status := postBaselineJSON(t, server.Client(), server.URL+"/api/v1/query/map", MapQueryRequest{Query: "search src_ip=1"}, &fallback); status != http.StatusOK || fallback.MappedQuery != "search src_ip=1" {
		t.Fatalf("fallback status=%d response=%#v", status, fallback)
	}

	payload := bytes.NewBufferString(`{"mappings":[{"source":"src_ip","target":"source_ip"}]}`)
	resp, err := server.Client().Post(server.URL+"/api/v1/mappings", "application/json", payload)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("admin endpoint status=%d, want 404", resp.StatusCode)
	}
}
