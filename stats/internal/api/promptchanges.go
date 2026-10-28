package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// promptChangeDTO is one prompt change on the wire: a marker in the
// review-dynamics view, and the body POST /api/v1/prompt-changes accepts.
type promptChangeDTO struct {
	SHA     string    `json:"sha"`
	At      time.Time `json:"at"`
	Kind    string    `json:"kind"`
	Area    string    `json:"area"`
	Slot    *string   `json:"slot"`
	Target  string    `json:"target"`
	Summary string    `json:"summary"`
	CaseID  *string   `json:"caseId"`
}

// promptChangeAreas and promptChangeKinds mirror prompt_changes' CHECK
// constraints, so a bad body is a 400 naming the field rather than a
// store error.
var (
	promptChangeAreas = map[string]bool{"review": true, "impl": true, "vv": true, "fix": true}
	promptChangeKinds = map[string]bool{"tuning": true, "edit": true}
)

func toPromptChangeDTOs(rows []store.PromptChange) []promptChangeDTO {
	out := make([]promptChangeDTO, len(rows))
	for i, c := range rows {
		out[i] = promptChangeDTO{SHA: c.SHA, At: c.At.UTC(), Kind: c.Kind, Area: c.Area, Slot: c.Slot,
			Target: c.Target, Summary: c.Summary, CaseID: c.CaseID}
	}
	return out
}

// postPromptChange serves POST /api/v1/prompt-changes: upserts one prompt
// change (prompt-changes-table) and answers 204. A missing sha, target or
// time, or a kind or area outside the table's set, is 400.
func (h *statsHandler) postPromptChange(w http.ResponseWriter, r *http.Request) {
	var dto promptChangeDTO
	if err := decodeJSONBody(w, r, &dto); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	switch {
	case dto.SHA == "" || dto.Target == "" || dto.At.IsZero():
		writeError(w, http.StatusBadRequest, "prompt change needs sha, target and at")
		return
	case !promptChangeKinds[dto.Kind]:
		writeError(w, http.StatusBadRequest, fmt.Sprintf("unknown kind %q; accepted: tuning, edit", dto.Kind))
		return
	case !promptChangeAreas[dto.Area]:
		writeError(w, http.StatusBadRequest, fmt.Sprintf("unknown area %q; accepted: review, impl, vv, fix", dto.Area))
		return
	}
	in := store.PromptChange{SHA: dto.SHA, At: dto.At, Kind: dto.Kind, Area: dto.Area, Slot: dto.Slot,
		Target: dto.Target, Summary: dto.Summary, CaseID: dto.CaseID}
	if err := h.store.UpsertPromptChange(r.Context(), in); err != nil {
		status, msg := mapStoreError(h.logger, "upsert prompt change", err)
		writeError(w, status, msg)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// reviewDynamicsDTO is the review-dynamics view's rows: one object, not an
// array (KAN-1046).
type reviewDynamicsDTO struct {
	Bucket       string             `json:"bucket"`
	Series       []dynamicsPointDTO `json:"series"`
	Heat         []heatCellDTO      `json:"heat"`
	PreviousHeat []heatCellDTO      `json:"previousHeat"`
	// HeatVersion and PreviousHeatVersion key the heatmap by model version
	// instead of family (model-version-split).
	HeatVersion         []heatCellDTO      `json:"heatVersion"`
	PreviousHeatVersion []heatCellDTO      `json:"previousHeatVersion"`
	Summary             dynamicsSummaryDTO `json:"summary"`
	PreviousSummary     dynamicsSummaryDTO `json:"previousSummary"`
	Markers             []promptChangeDTO  `json:"markers"`
	LastPromptChangeAt  *time.Time         `json:"lastPromptChangeAt"`
}

type dynamicsPointDTO struct {
	Bucket string  `json:"bucket"`
	Metric string  `json:"metric"`
	Series string  `json:"series"`
	Value  float64 `json:"value"`
	Basis  int     `json:"basis"`
}

type heatCellDTO struct {
	Slot       string `json:"slot"`
	Model      string `json:"model"`
	Effort     string `json:"effort"`
	Dispatches int    `json:"dispatches"`
	Critical   int    `json:"critical"`
	Important  int    `json:"important"`
	Minor      int    `json:"minor"`
}

type dynamicsSummaryDTO struct {
	FindingsPerChange          float64  `json:"findingsPerChange"`
	CriticalImportantPerChange float64  `json:"criticalImportantPerChange"`
	MeanFixRounds              float64  `json:"meanFixRounds"`
	ImplFindingsPerTask        *float64 `json:"implFindingsPerTask"`
	VVFirstPassShare           *float64 `json:"vvFirstPassShare"`
	Changes                    int      `json:"changes"`
}

func toReviewDynamicsDTO(d store.ReviewDynamics) reviewDynamicsDTO {
	series := make([]dynamicsPointDTO, len(d.Series))
	for i, p := range d.Series {
		series[i] = dynamicsPointDTO{Bucket: p.Bucket.UTC().Format(time.DateOnly), Metric: p.Metric,
			Series: p.Series, Value: p.Value, Basis: p.Basis}
	}
	heat := func(cells []store.HeatCell) []heatCellDTO {
		out := make([]heatCellDTO, len(cells))
		for i, c := range cells {
			out[i] = heatCellDTO(c)
		}
		return out
	}
	summary := func(s store.DynamicsSummary) dynamicsSummaryDTO { return dynamicsSummaryDTO(s) }
	var last *time.Time
	if d.LastPromptChangeAt != nil {
		t := d.LastPromptChangeAt.UTC()
		last = &t
	}
	return reviewDynamicsDTO{
		Bucket: d.Bucket, Series: series, Heat: heat(d.Heat), PreviousHeat: heat(d.PreviousHeat),
		HeatVersion: heat(d.HeatVersion), PreviousHeatVersion: heat(d.PreviousHeatVersion),
		Summary: summary(d.Summary), PreviousSummary: summary(d.PreviousSummary),
		Markers: toPromptChangeDTOs(d.Markers), LastPromptChangeAt: last,
	}
}
