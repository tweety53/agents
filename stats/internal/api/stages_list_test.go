package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListStages(t *testing.T) {
	rec := httptest.NewRecorder()
	listStages(rec, httptest.NewRequest(http.MethodGet, "/api/v1/stages?command=/flow", nil))
	var body struct{ Stages []string }
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || rec.Code != http.StatusOK || len(body.Stages) == 0 || body.Stages[0] != "flow.kickoff" {
		t.Fatalf("GET /api/v1/stages?command=/flow = %d %s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	listStages(rec, httptest.NewRequest(http.MethodGet, "/api/v1/stages", nil))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("no command = %d, want 400", rec.Code)
	}
}
