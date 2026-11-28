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

const decisionUsage = `usage: flow decision render -file <decision.json> [-session-model <model>]

render prints the ## Decision block for one decision.json, the decider's
file (skills/flow/decider-prompt.md). It reads only the file: no store, no
network. -session-model is accepted and ignored.

Exit codes: 0 printed; 2 usage error, an unreadable file, or a body that is
not a decision: groups empty or a group without tasks, a panel bundle of
other than one or two slots, or visual.verify not one of required, skipped,
not configured.
`

// runDecision implements `flow decision render`. Rendering is a pure
// function of the decision JSON (renderDecision), so it is tested without
// fakes.
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
	fset.String("session-model", "", "")
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
	if *file == "" {
		fmt.Fprint(stderr, "flow: decision render: -file is required\n")
		fmt.Fprint(stderr, decisionUsage)
		return 2
	}
	body, err := os.ReadFile(*file)
	if err != nil {
		fmt.Fprintf(stderr, "flow: decision render: %v\n", err)
		return 2
	}
	out, err := renderDecision(body)
	if err != nil {
		fmt.Fprintf(stderr, "flow: decision render: %v\n", err)
		return 2
	}
	fmt.Fprint(stdout, out)
	return 0
}

// dDecision is the decider's decision.json. Any other key -- a decision
// recorded before the decider carried class, execution, bundles and more --
// is ignored.
type dDecision struct {
	Groups []struct {
		Tasks  []string `json:"tasks"`
		Pair   string   `json:"pair"`
		Reason string   `json:"reason"`
	} `json:"groups"`
	Panel []struct {
		Slots  []string `json:"slots"`
		Pair   string   `json:"pair"`
		Reason string   `json:"reason"`
	} `json:"panel"`
	Visual *struct {
		Verify string `json:"verify"`
		Reason string `json:"reason"`
	} `json:"visual"`
	Overrides []json.RawMessage `json:"overrides"`
}

// dVisualVerify is the closed set `visual.verify` takes.
var dVisualVerify = map[string]bool{"required": true, "skipped": true, "not configured": true}

// parseDecision decodes body and checks only what a consumer reads: every
// group names its tasks (an implementer dispatch per group), every panel
// bundle is one or two slots (a reviewer dispatch per bundle), and
// visual.verify is one of its three words.
func parseDecision(body []byte) (dDecision, error) {
	var d dDecision
	if err := json.Unmarshal(body, &d); err != nil {
		return d, fmt.Errorf("not a decision: %v", err)
	}
	if len(d.Groups) == 0 {
		return d, fmt.Errorf("groups is empty: every task belongs to one group")
	}
	for i, g := range d.Groups {
		if len(g.Tasks) == 0 || strings.TrimSpace(strings.Join(g.Tasks, "")) == "" {
			return d, fmt.Errorf("groups[%d] names no tasks", i)
		}
	}
	for i, b := range d.Panel {
		if len(b.Slots) < 1 || len(b.Slots) > 2 {
			return d, fmt.Errorf("panel[%d] holds %d slots, want one or two", i, len(b.Slots))
		}
	}
	if d.Visual == nil || !dVisualVerify[d.Visual.Verify] {
		got := ""
		if d.Visual != nil {
			got = d.Visual.Verify
		}
		return d, fmt.Errorf("visual.verify %q is not one of required, skipped, not configured", got)
	}
	return d, nil
}

// renderDecision is the ## Decision block: one line per choice, each group
// and bundle with its pair, when it carries one, and its reason.
func renderDecision(body []byte) (string, error) {
	d, err := parseDecision(body)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	line := func(label string, parts []string) {
		b.WriteString("- **" + label + ":** " + strings.Join(parts, " · ") + "\n")
	}
	b.WriteString("## Decision\n\n")
	var groups, panel []string
	for _, g := range d.Groups {
		groups = append(groups, strings.Join(g.Tasks, "+")+dPair(g.Pair)+dReason(g.Reason))
	}
	for _, p := range d.Panel {
		panel = append(panel, strings.Join(p.Slots, "+")+dPair(p.Pair)+dReason(p.Reason))
	}
	line("Groups", groups)
	if len(panel) > 0 {
		line("Panel", panel)
	}
	visual := d.Visual.Verify
	if r := strings.TrimSpace(d.Visual.Reason); r != "" {
		visual += " — " + r
	}
	line("Visual verify", []string{visual})
	if len(d.Overrides) > 0 {
		var o []string
		for _, raw := range d.Overrides {
			var s string
			if json.Unmarshal(raw, &s) != nil {
				s = string(raw)
			}
			o = append(o, s)
		}
		line("Overrides", o)
	}
	return b.String(), nil
}

// dReason is " (<reason>)", or "" for none.
func dReason(r string) string {
	if r = strings.TrimSpace(r); r == "" {
		return ""
	}
	return " (" + r + ")"
}

// dPair is " [<pair>]", or "" for none.
func dPair(p string) string {
	if p = strings.TrimSpace(p); p == "" {
		return ""
	}
	return " [" + p + "]"
}
