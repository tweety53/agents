package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// decisionRun drives `flow decision <args>` through run, the way the CLI
// dispatches it, with body written to a decision.json when non-empty.
func decisionRun(t *testing.T, body string, args ...string) (int, string, string) {
	t.Helper()
	full := []string{"decision"}
	for _, a := range args {
		if a == "<file>" {
			p := filepath.Join(t.TempDir(), "decision.json")
			if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
			a = p
		}
		full = append(full, a)
	}
	var out, errb bytes.Buffer
	rc := run(context.Background(), full, strings.NewReader(""), &out, &errb)
	return rc, out.String(), errb.String()
}

const decisionPreamble = "planning:  inline, this session (claude-opus-5-5)\n" +
	"reviewers: primary=opus, principles=opus\n" +
	"\n" +
	"## Decision\n" +
	"\n" +
	"| Input | Rule | Value |\n" +
	"|---|---|---|\n"

const decisionSettings = "\n| Setting | Rule | Result |\n|---|---|---|\n"

// TestDecisionRenderGolden pins every rule of brainstorm-planner.md's worked
// `## Decision` format: micro (the three roll rows not consulted, no ↳ row,
// the `default` panel), a small inline run (string implementer and fixer,
// static grouping), a regular free grouping (a full roster, an experimental
// slot joining the second dispatch, an override), a big sdd run with a split
// group (object pairs, the (mechanical: …; override: …) suffix), and a big
// run whose experimental slot was skipped for the bundle cap with mechanical
// groups (no suffix). The small case is the block kan-823 printed.
func TestDecisionRenderGolden(t *testing.T) {
	inputs := func(tasks, files int, red string) string {
		return `"inputs":{"tasks":` + strconv.Itoa(tasks) + `,"files":` + strconv.Itoa(files) + `,"repos":1,"migration":false,"spec":false,"red":` + red + `,"unverified":` + red + `}`
	}
	cases := []struct {
		name, body, want string
	}{
		{"micro", `{"class":"micro","classMechanical":"micro","override":null,` + inputs(1, 1, "false") + `,
			"rolls":{"compact":65,"experimental":0,"bundle":43,"effort":96},"execution":"inline",
			"implementer":"skipped — inline","fixer":"skipped — inline","panel":"default",
			"groups":null,"groups_mechanical":null,"groups_override":null,"groups_reason":null,
			"parent":{"model":"claude-opus-5-5","effort":"unknown"},"overrides":[]}`,
			decisionPreamble +
				"| class | mechanical micro | micro (override: none) |\n" +
				"| inputs | plan-class.sh | tasks 1 · files 1 · repos 1 · migration no · spec no · red no · unverified no |\n" +
				"| roll: compact | 65 < 90 | not consulted — micro |\n" +
				"| roll: experimental | 0 < 30 | not consulted — micro |\n" +
				"| roll: bundle | 43 ≥ 30 | not consulted — micro |\n" +
				"| roll: effort | 96 ≥ 80 | not consulted — micro |\n" +
				decisionSettings +
				"| execution mode | class micro | inline |\n" +
				"| implementer model | — | skipped — inline |\n" +
				"| review panel | class micro | default |\n" +
				"| implementer groups | — | skipped — inline |\n"},
		{"small inline", `{"class":"small","classMechanical":"small","override":null,` + inputs(2, 3, "false") + `,
			"rolls":{"compact":85,"experimental":84,"bundle":17,"effort":40},"execution":"inline",
			"implementer":"skipped — inline","fixer":"skipped — inline",
			"panel":{"compact":true,"rerun":"delta","roster":[{"slot":"primary","experimental":false},{"slot":"principles","experimental":false}],
				"grouping":"static","dispatches":[{"slots":["primary","principles"],"model":"opus","effort":"medium","reason":"git refusals punish a sloppy reading"}],
				"rerun_dispatch":{"model":"opus","effort":"low","reason":"a delta re-run reads a fix against its finding"},"grouping_reason":null},
			"groups":null,"groups_mechanical":null,"groups_override":null,"groups_reason":null}`,
			decisionPreamble +
				"| class | mechanical small | small (override: none) |\n" +
				"| inputs | plan-class.sh | tasks 2 · files 3 · repos 1 · migration no · spec no · red no · unverified no |\n" +
				"| roll: compact | 85 < 90 | compact |\n" +
				"| roll: experimental | 84 ≥ 30 | no slot |\n" +
				"| roll: bundle | 17 < 30 | static grouping |\n" +
				"| roll: effort | 40 < 80 | medium effort |\n" +
				decisionSettings +
				"| execution mode | class small | inline |\n" +
				"| implementer model | — | skipped — inline |\n" +
				"| ↳ fixer | — | skipped — inline |\n" +
				"| review panel | class small | compact · delta rerun |\n" +
				"| ↳ dispatch 1 | opus / medium — git refusals punish a sloppy reading | primary+principles |\n" +
				"| ↳ rerun | opus / low — a delta re-run reads a fix against its finding | every fix-round re-run, one role per dispatch |\n" +
				"| implementer groups | — | skipped — inline |\n"},
		{"regular free grouping", `{"class":"regular","classMechanical":"small","override":"touches the store seam",` + inputs(6, 12, "true") + `,
			"rolls":{"compact":95,"experimental":12,"bundle":60,"effort":85},"execution":"inline",
			"implementer":"skipped — inline","fixer":"skipped — inline",
			"panel":{"compact":false,"rerun":"delta","roster":[{"slot":"primary","experimental":false},{"slot":"principles","experimental":false},
				{"slot":"failure-modes","experimental":false},{"slot":"mutation","experimental":false},
				{"slot":"exp-a","experimental":true,"prompt":"skills/flow/experimental/a.md","description":"first probe"}],
				"grouping":"free","dispatches":[{"slots":["primary","principles"],"model":"opus","effort":"high","reason":"floor"},
				{"slots":["failure-modes","mutation","exp-a"],"model":"opus","effort":"medium","reason":"edge cases"}],
				"rerun_dispatch":{"model":"sonnet","effort":"low","reason":"confirms a delta"},"grouping_reason":"failure paths read together"},
			"groups":null,"groups_mechanical":null,"groups_override":null,"groups_reason":null}`,
			decisionPreamble +
				"| class | mechanical small | regular (override: touches the store seam) |\n" +
				"| inputs | plan-class.sh | tasks 6 · files 12 · repos 1 · migration no · spec no · red yes · unverified yes |\n" +
				"| roll: compact | 95 ≥ 90 | full |\n" +
				"| roll: experimental | 12 < 30 | exp-a |\n" +
				"| roll: bundle | 60 ≥ 30 | free grouping |\n" +
				"| roll: effort | 85 ≥ 80 | free effort |\n" +
				decisionSettings +
				"| execution mode | class regular | inline |\n" +
				"| implementer model | — | skipped — inline |\n" +
				"| ↳ fixer | — | skipped — inline |\n" +
				"| review panel | class regular | full · delta rerun |\n" +
				"| ↳ dispatch 1 | opus / high — floor | primary+principles |\n" +
				"| ↳ dispatch 2 | opus / medium — edge cases | failure-modes+mutation+exp-a |\n" +
				"| ↳ rerun | sonnet / low — confirms a delta | every fix-round re-run, one role per dispatch |\n" +
				"| ↳ grouping | free | failure paths read together |\n" +
				"| implementer groups | — | skipped — inline |\n"},
		{"big sdd split group", `{"class":"big","classMechanical":"big","override":null,` + inputs(26, 43, "false") + `,
			"rolls":{"compact":26,"experimental":38,"bundle":35,"effort":12},"execution":"sdd",
			"implementer":{"model":"opus","effort":"high","reason":"Go guards with git fixtures"},
			"fixer":{"model":"opus","effort":"medium","reason":"fix rounds are narrower"},
			"panel":{"compact":true,"rerun":"delta","roster":[{"slot":"primary","experimental":false},{"slot":"principles","experimental":false}],
				"grouping":"free","dispatches":[{"slots":["primary","principles"],"model":"opus","effort":"high","reason":"wide diff"}],
				"rerun_dispatch":{"model":"sonnet","effort":"low","reason":"confirms a delta"},"grouping_reason":"compact roster is the floor bundle alone"},
			"groups":[{"bundles":["1","2"],"model":"opus","effort":"high","reason":"store seam"},{"bundles":[3],"model":"sonnet","effort":"low","reason":"docs only"}],
			"groups_mechanical":[["1","2",3]],"groups_override":"split docs out","groups_reason":"split: docs task apart"}`,
			decisionPreamble +
				"| class | mechanical big | big (override: none) |\n" +
				"| inputs | plan-class.sh | tasks 26 · files 43 · repos 1 · migration no · spec no · red no · unverified no |\n" +
				"| roll: compact | 26 < 90 | compact |\n" +
				"| roll: experimental | 38 ≥ 30 | no slot |\n" +
				"| roll: bundle | 35 ≥ 30 | free grouping |\n" +
				"| roll: effort | 12 < 80 | medium effort |\n" +
				decisionSettings +
				"| execution mode | class big | sdd |\n" +
				"| implementer model | Go guards with git fixtures | opus/high |\n" +
				"| ↳ fixer | opus / medium — fix rounds are narrower | opus/medium |\n" +
				"| review panel | class big | compact · delta rerun |\n" +
				"| ↳ dispatch 1 | opus / high — wide diff | primary+principles |\n" +
				"| ↳ rerun | sonnet / low — confirms a delta | every fix-round re-run, one role per dispatch |\n" +
				"| ↳ grouping | free | compact roster is the floor bundle alone |\n" +
				"| implementer groups | — | split: docs task apart |\n" +
				"| ↳ group 1, 2 | opus / high — store seam | 1, 2 (mechanical: [1, 2, 3]; override: split docs out) |\n" +
				"| ↳ group 3 | sonnet / low — docs only | 3 (mechanical: [1, 2, 3]; override: split docs out) |\n"},
		{"experimental skipped for the bundle cap", `{"class":"big","classMechanical":"big","override":null,` + inputs(22, 30, "false") + `,
			"rolls":{"compact":71,"experimental":0,"bundle":16,"effort":77},"execution":"sdd",
			"implementer":{"model":"opus","effort":"xhigh","reason":"concurrency seam"},
			"fixer":{"model":"opus","effort":"high","reason":"same seam"},
			"panel":{"compact":true,"rerun":"delta","roster":[{"slot":"primary","experimental":false},{"slot":"principles","experimental":false}],
				"experimental":"skipped — bundle cap",
				"grouping":"static","dispatches":[{"slots":["primary","principles"],"model":"opus","effort":"high","reason":"floor"}],
				"rerun_dispatch":{"model":"opus","effort":"low","reason":"short re-read"},"grouping_reason":null},
			"groups":[{"bundles":[1],"model":"opus","effort":"xhigh","reason":"the seam"}],
			"groups_mechanical":[[1]],"groups_override":null,"groups_reason":"mechanical"}`,
			decisionPreamble +
				"| class | mechanical big | big (override: none) |\n" +
				"| inputs | plan-class.sh | tasks 22 · files 30 · repos 1 · migration no · spec no · red no · unverified no |\n" +
				"| roll: compact | 71 < 90 | compact |\n" +
				"| roll: experimental | 0 < 30 | skipped — bundle cap |\n" +
				"| roll: bundle | 16 < 30 | static grouping |\n" +
				"| roll: effort | 77 < 80 | medium effort |\n" +
				decisionSettings +
				"| execution mode | class big | sdd |\n" +
				"| implementer model | concurrency seam | opus/xhigh |\n" +
				"| ↳ fixer | opus / high — same seam | opus/high |\n" +
				"| review panel | class big | compact · delta rerun · experimental: skipped — bundle cap |\n" +
				"| ↳ dispatch 1 | opus / high — floor | primary+principles |\n" +
				"| ↳ rerun | opus / low — short re-read | every fix-round re-run, one role per dispatch |\n" +
				"| implementer groups | — | mechanical |\n" +
				"| ↳ group 1 | opus / xhigh — the seam | 1 |\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rc, out, errs := decisionRun(t, c.body, "render", "-file", "<file>",
				"-session-model", "claude-opus-5-5", "-reviewers", "primary=opus, principles=opus")
			if rc != 0 || out != c.want || errs != "" {
				t.Fatalf("rc=%d stderr=%q\ngot:\n%s\nwant:\n%s", rc, errs, out, c.want)
			}
		})
	}
}

// TestDecisionRenderRefusals pins exit 2 with nothing on stdout for a usage
// error, an unreadable file, a body that is not a decision, and a decision
// missing a required field.
func TestDecisionRenderRefusals(t *testing.T) {
	good := `{"class":"micro","classMechanical":"micro","override":null,
		"inputs":{"tasks":1,"files":1,"repos":1,"migration":false,"spec":false,"red":false,"unverified":false},
		"rolls":{"compact":65,"experimental":0,"bundle":43,"effort":50},"execution":"inline",
		"implementer":"skipped — inline","fixer":"skipped — inline","panel":"default","groups":null}`
	flags := []string{"-session-model", "m", "-reviewers", "r"}
	cases := []struct {
		name, body string
		args       []string
		stderr     string
	}{
		{"no subcommand", "", nil, decisionUsage},
		{"unknown subcommand", "", []string{"print"}, "flow: unknown decision subcommand \"print\"\n" + decisionUsage},
		{"no -file", "", append([]string{"render"}, flags...), "flow: decision render: -file, -session-model and -reviewers are required\n" + decisionUsage},
		{"no -reviewers", good, []string{"render", "-file", "<file>", "-session-model", "m"}, "flow: decision render: -file, -session-model and -reviewers are required\n" + decisionUsage},
		{"stray argument", good, append([]string{"render", "-file", "<file>", "x"}, flags...), "flow: decision render: unexpected argument \"x\"\n" + decisionUsage},
		{"unreadable file", "", append([]string{"render", "-file", "/nonexistent/decision.json"}, flags...), "flow: decision render: open /nonexistent/decision.json: no such file or directory\n"},
		{"not JSON", "{", append([]string{"render", "-file", "<file>"}, flags...), "flow: decision render: not a decision: unexpected end of JSON input\n"},
		{"missing rolls.bundle", strings.Replace(good, `,"bundle":43`, "", 1), append([]string{"render", "-file", "<file>"}, flags...), "flow: decision render: missing required field \"rolls.bundle\"\n"},
		{"missing groups", strings.Replace(good, `,"groups":null`, "", 1), append([]string{"render", "-file", "<file>"}, flags...), "flow: decision render: missing required field \"groups\"\n"},
		{"missing panel.rerun_dispatch", strings.Replace(good, `"panel":"default"`, `"panel":{"compact":true,"rerun":"delta","grouping":"static","dispatches":[]}`, 1), append([]string{"render", "-file", "<file>"}, flags...), "flow: decision render: missing required field \"panel.rerun_dispatch\"\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rc, out, errs := decisionRun(t, c.body, c.args...)
			if rc != 2 || out != "" || errs != c.stderr {
				t.Fatalf("rc=%d stdout=%q\nstderr=%q\nwant  =%q", rc, out, errs, c.stderr)
			}
		})
	}
}
