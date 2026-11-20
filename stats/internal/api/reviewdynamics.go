package api

import (
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

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
	Projects            []string           `json:"projects"`
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
	return reviewDynamicsDTO{
		Bucket: d.Bucket, Series: series, Heat: heat(d.Heat), PreviousHeat: heat(d.PreviousHeat),
		HeatVersion: heat(d.HeatVersion), PreviousHeatVersion: heat(d.PreviousHeatVersion),
		Summary: summary(d.Summary), PreviousSummary: summary(d.PreviousSummary),
		Projects: d.Projects,
	}
}
