package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const decisionUsage = `usage: flow decision render -file <decision.json> -session-model <model> -reviewers <REVIEWERS>

render prints the ## Decision block for one decision.json, the format brainstorm-planner.md's Decide step prints.
It reads only the file: no store, no network.

Exit codes: 0 printed; 2 usage error, an unreadable file, or a body that is
not a decision, lacks a required field, or carries a malformed visual.
`

// runDecision implements `flow decision render`. Rendering is a pure
// function of the decision JSON and the two flags (renderDecision), so it is
// tested without fakes.
func runDecision(_ context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, decisionUsage)
		return 2
	}
	if args[0] != "render" {
		fmt.Fprintf(stderr, "flow: unknown decision subcommand %q\n", args[0])
		fmt.Fprint(stderr, decisionUsage)
		return 2
	}
	fset := flag.NewFlagSet("flow decision render", flag.ContinueOnError)
	fset.SetOutput(io.Discard)
	file := fset.String("file", "", "")
	model := fset.String("session-model", "", "")
	reviewers := fset.String("reviewers", "", "")
	if err := fset.Parse(args[1:]); err != nil {
		fmt.Fprintf(stderr, "flow: decision render: %v\n", err)
		fmt.Fprint(stderr, decisionUsage)
		return 2
	}
	if fset.NArg() != 0 {
		fmt.Fprintf(stderr, "flow: decision render: unexpected argument %q\n", fset.Arg(0))
		fmt.Fprint(stderr, decisionUsage)
		return 2
	}
	if *file == "" || *model == "" || *reviewers == "" {
		fmt.Fprint(stderr, "flow: decision render: -file, -session-model and -reviewers are required\n")
		fmt.Fprint(stderr, decisionUsage)
		return 2
	}
	body, err := os.ReadFile(*file)
	if err != nil {
		fmt.Fprintf(stderr, "flow: decision render: %v\n", err)
		return 2
	}
	out, err := renderDecision(body, *model, *reviewers)
	if err != nil {
		fmt.Fprintf(stderr, "flow: decision render: %v\n", err)
		return 2
	}
	fmt.Fprint(stdout, out)
	return 0
}

// dPair is a recorded pair: an object {model, effort, reason}, or one of the
// recorded strings ("skipped — inline").
type dPair struct {
	text                  string
	Model, Effort, Reason string
	isPair                bool
}

func (p *dPair) UnmarshalJSON(b []byte) error {
	if json.Unmarshal(b, &p.text) == nil {
		return nil
	}
	var o struct {
		Model  string `json:"model"`
		Effort string `json:"effort"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(b, &o); err != nil {
		return err
	}
	p.Model, p.Effort, p.Reason, p.isPair = o.Model, o.Effort, o.Reason, true
	return nil
}

// pair is a pair as the block prints it: "<model>/<effort>".
func (p dPair) pair() string { return p.Model + "/" + p.Effort }

type dDecision struct {
	Class           string  `json:"class"`
	ClassMechanical string  `json:"classMechanical"`
	Override        *string `json:"override"`
	Inputs          struct {
		Tasks      int  `json:"tasks"`
		Files      int  `json:"files"`
		Repos      int  `json:"repos"`
		Migration  bool `json:"migration"`
		Spec       bool `json:"spec"`
		Red        bool `json:"red"`
		Unverified bool `json:"unverified"`
	} `json:"inputs"`
	Rolls struct {
		Compact      int `json:"compact"`
		Experimental int `json:"experimental"`
		Bundle       int `json:"bundle"`
		Effort       int `json:"effort"`
	} `json:"rolls"`
	Execution   string          `json:"execution"`
	Implementer dPair           `json:"implementer"`
	Fixer       dPair           `json:"fixer"`
	Panel       json.RawMessage `json:"panel"`
	Groups      []struct {
		Bundles []json.RawMessage `json:"bundles"`
		Model   string            `json:"model"`
		Effort  string            `json:"effort"`
		Reason  string            `json:"reason"`
	} `json:"groups"`
	GroupsMechanical [][]json.RawMessage `json:"groups_mechanical"`
	GroupsOverride   *string             `json:"groups_override"`
	GroupsReason     string              `json:"groups_reason"`
	Visual           *dVisual            `json:"visual"`
}

// dVisual is Decide's visual-verification decision. Optional, so a decision
// written before it existed -- or by a caller that never runs
// flow.visual-verify -- still renders; absent reads as `required`
// everywhere it is consulted.
type dVisual struct {
	Verify string `json:"verify"`
	Reason string `json:"reason"`
}

// dVisualVerify is the closed set `visual.verify` takes.
var dVisualVerify = map[string]bool{"required": true, "skipped": true, "not configured": true}

type dPanel struct {
	Compact bool   `json:"compact"`
	Rerun   string `json:"rerun"`
	Roster  []struct {
		Slot         string `json:"slot"`
		Experimental bool   `json:"experimental"`
	} `json:"roster"`
	Experimental string `json:"experimental"`
	Grouping     string `json:"grouping"`
	Dispatches   []struct {
		Slots  []string `json:"slots"`
		Model  string   `json:"model"`
		Effort string   `json:"effort"`
		Reason string   `json:"reason"`
	} `json:"dispatches"`
	RerunDispatch  dPair  `json:"rerun_dispatch"`
	GroupingReason string `json:"grouping_reason"`
}

// dRequired names every key a decision must carry, by dotted path; a key
// under panel is required only when panel is an object, and groups_reason
// only when groups is not null.
var dRequired = []string{
	"class", "classMechanical", "override", "inputs", "rolls", "execution", "implementer", "fixer", "panel", "groups",
	"inputs.tasks", "inputs.files", "inputs.repos", "inputs.migration", "inputs.spec", "inputs.red", "inputs.unverified",
	"rolls.compact", "rolls.experimental", "rolls.bundle", "rolls.effort",
	"panel.compact", "panel.rerun", "panel.grouping", "panel.dispatches", "panel.rerun_dispatch",
	"groups_reason",
}

// dMissing returns the first required key body lacks.
func dMissing(body []byte) (string, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(body, &top); err != nil {
		return "", err
	}
	sub := map[string]map[string]json.RawMessage{}
	for _, k := range []string{"inputs", "rolls", "panel"} {
		var m map[string]json.RawMessage
		if json.Unmarshal(top[k], &m) == nil {
			sub[k] = m
		}
	}
	for _, path := range dRequired {
		parent, key, nested := strings.Cut(path, ".")
		switch {
		case path == "groups_reason":
			if g, ok := top["groups"]; ok && string(g) != "null" {
				if _, ok := top[path]; !ok {
					return path, nil
				}
			}
		case !nested:
			if _, ok := top[path]; !ok {
				return path, nil
			}
		case parent == "panel" && sub["panel"] == nil:
			// a string panel ("default") carries no sub-keys
		default:
			if _, ok := sub[parent][key]; !ok {
				return path, nil
			}
		}
	}
	return "", nil
}

// dScalar is a bundle id as written: a JSON string unquoted, a number as is.
func dScalar(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	return string(raw)
}

func dIDs(raws []json.RawMessage) string {
	ids := make([]string, len(raws))
	for i, r := range raws {
		ids[i] = dScalar(r)
	}
	return strings.Join(ids, ", ")
}

// renderDecision is the ## Decision block: one short line per choice, a
// reason only where a pair departs from the implementer's, or where a value
// was raised, split or skipped.
func renderDecision(body []byte, _, reviewers string) (string, error) {
	missing, err := dMissing(body)
	if err != nil {
		return "", fmt.Errorf("not a decision: %v", err)
	}
	if missing != "" {
		return "", fmt.Errorf("missing required field %q", missing)
	}
	var d dDecision
	if err := json.Unmarshal(body, &d); err != nil {
		return "", fmt.Errorf("not a decision: %v", err)
	}
	if v := d.Visual; v != nil {
		if !dVisualVerify[v.Verify] {
			return "", fmt.Errorf("visual.verify %q is not one of required, skipped, not configured", v.Verify)
		}
		if strings.TrimSpace(v.Reason) == "" {
			return "", fmt.Errorf("visual.reason is required")
		}
	}
	var panelDefault string
	var p dPanel
	if json.Unmarshal(d.Panel, &panelDefault) != nil {
		if err := json.Unmarshal(d.Panel, &p); err != nil {
			return "", fmt.Errorf("not a decision: panel: %v", err)
		}
	}
	var b strings.Builder
	line := func(label, value string) { b.WriteString("- **" + label + ":** " + value + "\n") }
	b.WriteString("## Decision\n\n")

	class := fmt.Sprintf("%s · tasks %d · files %d · repos %d", d.Class, d.Inputs.Tasks, d.Inputs.Files, d.Inputs.Repos)
	if d.Override != nil {
		class += " (raised from " + d.ClassMechanical + ": " + *d.Override + ")"
	}
	line("Class", class)

	exec := d.Execution
	if d.Implementer.isPair {
		exec += " · implementer " + d.Implementer.pair()
	}
	if d.Fixer.isPair {
		exec += " · fixer " + d.Fixer.pair()
	}
	line("Execution", exec)

	if d.Groups != nil {
		parts := make([]string, len(d.Groups))
		for i, g := range d.Groups {
			gp := dPair{Model: g.Model, Effort: g.Effort}
			parts[i] = dRange(g.Bundles) + " " + gp.pair()
			if !d.Implementer.isPair || gp.pair() != d.Implementer.pair() {
				parts[i] += " (" + g.Reason + ")"
			}
		}
		groups := strings.Join(parts, " · ")
		if d.GroupsOverride != nil {
			groups += " — split: " + *d.GroupsOverride
		}
		line("Groups", groups)
	}

	if panelDefault != "" {
		line("Panel", panelDefault+" — "+reviewers)
	} else {
		parts := []string{"full"}
		if p.Compact {
			parts[0] = "compact"
		}
		for _, disp := range p.Dispatches {
			parts = append(parts, strings.Join(disp.Slots, "+")+" "+dPair{Model: disp.Model, Effort: disp.Effort}.pair())
		}
		parts = append(parts, "reruns "+p.RerunDispatch.pair())
		if p.Experimental == "skipped — bundle cap" {
			parts = append(parts, "experimental skipped — bundle cap")
		}
		line("Panel", strings.Join(parts, " · "))
	}

	if v := d.Visual; v != nil {
		visual := v.Verify
		if v.Verify != "required" {
			visual += " — " + v.Reason
		}
		line("Visual verify", visual)
	}
	return b.String(), nil
}

// dRange is a group's bundle ids, a consecutive integer run written a–b.
func dRange(raws []json.RawMessage) string {
	ids := make([]int, len(raws))
	for i, r := range raws {
		n, err := strconv.Atoi(dScalar(r))
		if err != nil || (i > 0 && n != ids[i-1]+1) {
			return dIDs(raws)
		}
		ids[i] = n
	}
	if len(ids) > 1 {
		return fmt.Sprintf("%d–%d", ids[0], ids[len(ids)-1])
	}
	return dIDs(raws)
}
