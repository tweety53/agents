package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// DecidedPair is the decided pair of the dispatch keyed key on role, read from <wt>/.superpowers/sdd/decision.json,
// and the `decider: <what>` text its row records as -pair-pick: a bundle's
// `panel-<round>-<slots>[-<repo>]` on panel-reviewer. "" when the decision
// names no pair for it or the role's base pairs no longer hold that pair.
func DecidedPair(wt, role, key string, base []string) (pair, pick string) {
	b, err := os.ReadFile(filepath.Join(wt, ".superpowers", "sdd", "decision.json"))
	var d PlanDecision
	if err != nil || json.Unmarshal(b, &d) != nil {
		return "", ""
	}
	if role == BundleRole {
		if m := PanelKey.FindStringSubmatch(key); m != nil {
			for _, p := range d.Panel {
				if s := strings.Join(p.Slots, "+"); m[1] == s || strings.HasPrefix(m[1], s+"-") {
					pair, pick = p.Pair, "decider: bundle "+s
				}
			}
		}
	}
	if pair == "" || !slices.Contains(base, pair) {
		return "", ""
	}
	return pair, pick
}

// BundleRole is the role a decided dispatch runs on: a pass-1 bundle's reviewer. The decider picks
// its pair from that role's base pairs.
const BundleRole = "panel-reviewer"

var PanelKey = regexp.MustCompile(`^panel-[0-9]+-(.+)$`)

type PlanDecision struct {
	Groups []DecisionEntry `json:"groups"`
	Panel  []DecisionEntry `json:"panel"`
	Visual struct {
		Verify string `json:"verify"`
		Reason string `json:"reason"`
	} `json:"visual"`
	Overrides []json.RawMessage `json:"overrides"`
}

type DecisionEntry struct {
	Tasks  []string `json:"tasks,omitempty"`
	Slots  []string `json:"slots,omitempty"`
	Pair   string   `json:"pair,omitempty"`
	Reason string   `json:"reason"`
}
