package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
)

func TestClosureRESTCanonicalAndErrors(t *testing.T) {
	body := []byte(`{"schema_version":1,"document":{"text":"| lookup users user OUTPUT role"},"bundle":{"schema_version":1,"scope_id":"synthetic","collections":[{"kind":"lookup","coverage":"complete"}],"objects":[{"id":"users","kind":"lookup","name":"users","source_id":"lookup-source"}]}}`)
	request, err := closure.DecodeRequest(body)
	if err != nil {
		t.Fatal(err)
	}
	want, err := closure.Evaluate(request)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		body        []byte
		contentType string
		code        int
	}{{body, "application/json", 200}, {[]byte(`{"schema_version":1,"document":{}}`), "application/json", 400}, {body, "text/plain", 400}, {bytes.Repeat([]byte("x"), 8<<20+1), "application/json", 400}} {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/query/closure", bytes.NewReader(item.body))
		req.Header.Set("Content-Type", item.contentType)
		recorder := httptest.NewRecorder()
		NewServer().Handler().ServeHTTP(recorder, req)
		if recorder.Code != item.code {
			t.Fatalf("code=%d want=%d body=%s", recorder.Code, item.code, recorder.Body.String())
		}
		if item.code == 200 {
			var got closure.Report
			if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil || !reflect.DeepEqual(&got, want) {
				t.Fatalf("report parity=%t err=%v", reflect.DeepEqual(&got, want), err)
			}
		} else if !strings.Contains(recorder.Body.String(), "error") {
			t.Fatalf("missing envelope: %s", recorder.Body.String())
		}
	}
}
