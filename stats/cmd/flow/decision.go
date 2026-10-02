package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

const decisionUsage = `usage: flow decision render -file <decision.json> -session-model <model> -reviewers <REVIEWERS>

render prints the planning:/reviewers: lines and the ## Decision block for
one decision.json, the format brainstorm-planner.md's Decide step prints.
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

// rule is a pair's rule cell: "<model> / <effort> — <reason>".
func (p dPair) rule() string { return p.Model + " / " + p.Effort + " — " + p.Reason }

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

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// dRoll is a roll's rule cell: the roll against its threshold.
func dRoll(n, threshold int) string {
	if n < threshold {
		return fmt.Sprintf("%d < %d", n, threshold)
	}
	return fmt.Sprintf("%d ≥ %d", n, threshold)
}

// renderDecision is the planning:/reviewers: preamble and the two ## Decision
// tables, one fact per row, every reason in the middle column.
func renderDecision(body []byte, sessionModel, reviewers string) (string, error) {
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
	micro := d.Class == "micro"

	var b strings.Builder
	row := func(cells ...string) { b.WriteString("| " + strings.Join(cells, " | ") + " |\n") }
	fmt.Fprintf(&b, "planning:  inline, this session (%s)\nreviewers: %s\n\n## Decision\n\n", sessionModel, reviewers)

	row("Input", "Rule", "Value")
	b.WriteString("|---|---|---|\n")
	override := "none"
	if d.Override != nil {
		override = *d.Override
	}
	row("class", "mechanical "+d.ClassMechanical, d.Class+" (override: "+override+")")
	in := d.Inputs
	row("inputs", "plan-class.sh", fmt.Sprintf("tasks %d · files %d · repos %d · migration %s · spec %s · red %s · unverified %s",
		in.Tasks, in.Files, in.Repos, yesNo(in.Migration), yesNo(in.Spec), yesNo(in.Red), yesNo(in.Unverified)))
	compact, experimental, bundle := "full", "no slot", "free grouping"
	if d.Rolls.Compact < 90 {
		compact = "compact"
	}
	if d.Rolls.Experimental < 30 {
		experimental = "none available"
		if p.Experimental != "" {
			experimental = p.Experimental
		}
		for _, r := range p.Roster {
			if r.Experimental {
				experimental = r.Slot
			}
		}
	}
	if d.Rolls.Bundle < 30 {
		bundle = "static grouping"
	}
	effort := "free effort"
	if d.Rolls.Effort < 80 {
		effort = "medium effort"
	}
	if micro {
		compact, experimental, bundle, effort = "not consulted — micro", "not consulted — micro", "not consulted — micro", "not consulted — micro"
	}
	row("roll: compact", dRoll(d.Rolls.Compact, 90), compact)
	row("roll: experimental", dRoll(d.Rolls.Experimental, 30), experimental)
	row("roll: bundle", dRoll(d.Rolls.Bundle, 30), bundle)
	row("roll: effort", dRoll(d.Rolls.Effort, 80), effort)

	b.WriteString("\n")
	row("Setting", "Rule", "Result")
	b.WriteString("|---|---|---|\n")
	row("execution mode", "class "+d.Class, d.Execution)
	if d.Implementer.isPair {
		row("implementer model", d.Implementer.Reason, d.Implementer.Model+"/"+d.Implementer.Effort)
	} else {
		row("implementer model", "—", d.Implementer.text)
	}
	if !micro {
		if d.Fixer.isPair {
			row("↳ fixer", d.Fixer.rule(), d.Fixer.Model+"/"+d.Fixer.Effort)
		} else {
			row("↳ fixer", "—", d.Fixer.text)
		}
	}
	if panelDefault != "" {
		row("review panel", "class "+d.Class, panelDefault)
	} else {
		shape := "full"
		if p.Compact {
			shape = "compact"
		}
		value := shape + " · " + p.Rerun + " rerun"
		if p.Experimental == "skipped — bundle cap" {
			value += " · experimental: skipped — bundle cap"
		}
		row("review panel", "class "+d.Class, value)
		for i, disp := range p.Dispatches {
			row(fmt.Sprintf("↳ dispatch %d", i+1), dPair{Model: disp.Model, Effort: disp.Effort, Reason: disp.Reason}.rule(), strings.Join(disp.Slots, "+"))
		}
		row("↳ rerun", p.RerunDispatch.rule(), "every fix-round re-run, one role per dispatch")
		if p.Grouping == "free" {
			row("↳ grouping", "free", p.GroupingReason)
		}
	}
	if d.Groups == nil {
		row("implementer groups", "—", "skipped — inline")
	} else {
		row("implementer groups", "—", d.GroupsReason)
		suffix := ""
		if d.GroupsOverride != nil {
			mech := make([]string, len(d.GroupsMechanical))
			for i, g := range d.GroupsMechanical {
				mech[i] = "[" + dIDs(g) + "]"
			}
			suffix = " (mechanical: " + strings.Join(mech, " · ") + "; override: " + *d.GroupsOverride + ")"
		}
		for _, g := range d.Groups {
			ids := dIDs(g.Bundles)
			row("↳ group "+ids, dPair{Model: g.Model, Effort: g.Effort, Reason: g.Reason}.rule(), ids+suffix)
		}
	}
	if d.Visual != nil {
		row("visual verification", d.Visual.Reason, d.Visual.Verify)
	}
	return b.String(), nil
}
