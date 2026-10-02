package guard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/stages"
)

// Every case of scripts/test-check-stage-mark-calls.sh at d71a2327, one
// subtest per ok: label. The bash harness ran the guard on a mktemp fixture
// and asserted on its exit and its combined output; here the guard runs
// in-process over a fixture in t.TempDir(). Each harness case runs the guard
// once, shared by the subtests carrying its labels.
//
// The served stage keys are Env.StageKeys, reading internal/stages' table --
// the table `go run ./cmd/flow stage keys` prints -- except in cases 24, 38
// and 39, which run the real `go run` the guard execs in production: case 24
// against this repository, 38 and 39 against a sandbox stats module.

// smcRepoRoot is this checkout's root, the tree the bash guard sat in.
var smcRepoRoot = func() string {
	root, err := filepath.Abs("../../..")
	if err != nil {
		panic(err)
	}
	return root
}()

// smcEnv is the process's environment with FLOW_GUARD_REPO_ROOT set to root,
// as the shim exports it; served keys come from internal/stages unless real.
func smcEnv(root string, real bool) Env {
	lookup := func(k string) (string, bool) {
		if k == "FLOW_GUARD_REPO_ROOT" {
			return root, true
		}
		return os.LookupEnv(k)
	}
	env := Env{Dir: root, LookupEnv: lookup, Getenv: func(k string) string { v, _ := lookup(k); return v }}
	if !real {
		env.StageKeys = func() (string, error) { return strings.Join(stages.Keys(), "\n"), nil }
	}
	return env
}

// smcInOrder is the harness's `case "$OUT" in *a*b*)`: every part present,
// each after the one before.
func smcInOrder(out string, parts ...string) bool {
	for _, p := range parts {
		i := strings.Index(out, p)
		if i < 0 {
			return false
		}
		out = out[i+len(p):]
	}
	return true
}

type smcCheck struct {
	label string
	ok    func(rc int, out, dir string) bool
}

func smcRC(label string, want int) smcCheck {
	return smcCheck{label, func(rc int, _, _ string) bool { return rc == want }}
}

func smcHas(label string, parts ...string) smcCheck {
	return smcCheck{label, func(_ int, out, _ string) bool { return smcInOrder(out, parts...) }}
}

func smcLacks(label, part string) smcCheck {
	return smcCheck{label, func(_ int, out, _ string) bool { return !strings.Contains(out, part) }}
}

// smcFile is new_fixture's FIXTURE_FILE, "<dir>/SKILL.md".
func smcFile(dir string) string { return dir + "/SKILL.md" }

func TestCheckStageMarkCalls(t *testing.T) {
	t.Parallel()
	const (
		fence   = "```bash\n"
		unfence = "```\n"
		begin   = "flow stage begin -command '/flow' -stage flow.review-panel -harness <harness> -session-token mf-abc123 <name>\n"
		fastBeg = "flow stage begin -command '/flow-fast' -stage flow.kickoff -harness <harness> -session-token mf-abc123 "
		unknown = "not a key of the served stage-key vocabulary"
	)
	block := func(lines ...string) string { return fence + strings.Join(lines, "") + unfence }
	sandboxGoMod := "module sandboxflow\n\ngo 1.21\n"

	type smcCase struct {
		files  map[string]string         // relative to the fixture dir
		args   func(dir string) []string // nil: {dir}
		root   func(dir string) string   // nil: this checkout
		real   bool                      // exec the real `go run ./cmd/flow stage keys`
		checks []smcCheck
	}
	cases := []smcCase{
		{files: map[string]string{"SKILL.md": block(begin,
			"flow stage end   -command '/flow' -stage flow.review-panel -outcome completed <name>\n")},
			checks: []smcCheck{smcRC("case 1: compliant single-line call passes", 0)}},
		{files: map[string]string{"SKILL.md": block("flow stage begin -command '/flow' -stage flow.review-panel -harness <harness> <name>\n")},
			checks: []smcCheck{smcRC("case 2: missing -session-token is caught", 1),
				smcHas("case 2: finding names the missing flag", "carries no -session-token")}},
		{files: map[string]string{"SKILL.md": block("flow stage begin -command '/flow' -stage flow.review-panel -session-token mf-abc123 <name>\n")},
			checks: []smcCheck{smcRC("case 3: missing -harness is caught", 1),
				smcHas("case 3: finding names the missing flag", "carries no -harness")}},
		{files: map[string]string{"SKILL.md": block("flow stage begin -command '/flow' -stage 'the review panel' <name>\n")},
			checks: []smcCheck{smcRC("case 4: both flags missing is caught", 1),
				smcHas("case 4: both findings reported", "carries no -session-token", "carries no -harness")}},
		{files: map[string]string{"SKILL.md": block(`flow stage begin -command '/flow' -stage flow.review-panel -harness <harness> -session-token "mf-$(date +%s)-$$" <name>` + "\n")},
			checks: []smcCheck{smcRC("case 5: $(...) substitution is caught", 1),
				smcHas("case 5: finding names the substitution shape", "command substitution")}},
		{files: map[string]string{"SKILL.md": block("flow stage begin -command '/flow' -stage flow.review-panel -harness <harness> -session-token \"mf-`date +%s`\" <name>\n")},
			checks: []smcCheck{smcRC("case 6: backtick substitution is caught", 1),
				smcHas("case 6: finding names the backtick", "backtick")}},
		{files: map[string]string{"SKILL.md": block(`flow stage begin -command '/flow' -stage flow.review-panel -harness <harness> -session-token "mf-$RANDOM" <name>` + "\n")},
			checks: []smcCheck{smcRC("case 7: $VAR substitution is caught", 1),
				smcHas("case 7: finding names the shell-variable shape", "shell variable reference")}},
		{files: map[string]string{"SKILL.md": block("flow stage begin -command '/flow' -stage flow.review-panel -harness claude-code -session-token mf-abc123 <name>\n")},
			checks: []smcCheck{smcRC("case 7b: hardcoded harness literal is caught", 1),
				smcHas("case 7b: finding names the hardcoded literal", "hardcoded literal")}},
		{files: map[string]string{"SKILL.md": block("flow stage begin -command '/flow' -stage flow.review-panel -harness zcode -session-token mf-abc123 <name>\n")},
			checks: []smcCheck{smcRC("case 7c: any hardcoded harness literal is caught, not just claude-code", 1)}},
		{files: map[string]string{"SKILL.md": block("flow stage begin -command '/flow-fast' \\\n",
			"  -stage flow.review-panel \\\n", "  -harness <harness> \\\n", "  -session-token mf-continued-001 <name>\n")},
			checks: []smcCheck{smcRC("case 8: compliant multi-line call passes", 0)}},
		{files: map[string]string{"SKILL.md": block(begin,
			"flow stage end   -command '/flow' -stage flow.review-panel -outcome completed <name>\n")},
			checks: []smcCheck{smcRC("case 9: stage end is never checked", 0)}},
		{files: map[string]string{"SKILL.md": "Ordinary prose with no flow stage marks in it.\n"},
			checks: []smcCheck{smcRC("case 10: a file with no stage marks at all is an undeclared coverage violation", 1),
				{"case 10: names the file, the zero count, and that it is undeclared", func(_ int, out, dir string) bool {
					return smcInOrder(out, smcFile(dir), "0 checked", "not declared expected-zero")
				}}}},
		{files: map[string]string{"SKILL.md": block(fastBeg + "<name-or-best-guess>\n")},
			checks: []smcCheck{smcRC("case 11: guessed change name is caught", 1),
				smcHas("case 11: finding names the offending argument", "<name-or-best-guess>")}},
		{files: map[string]string{"SKILL.md": block("flow stage begin -command '/flow-fast' \\\n", "  -stage flow.kickoff \\\n",
			"  -harness <harness> \\\n", "  -session-token mf-abc123 <name-or-best-guess>\n")},
			checks: []smcCheck{smcRC("case 12: guessed change name is caught across continuation lines", 1),
				smcHas("case 12: finding names the offending argument", "<name-or-best-guess>")}},
		{files: map[string]string{"SKILL.md": block(fastBeg + "<name>\n")},
			checks: []smcCheck{smcRC("case 13: compliant <name> argument passes", 0)}},
		{files: map[string]string{"SKILL.md": block("flow stage end -command '/flow-fast' -stage flow.kickoff -outcome completed <name-or-best-guess>\n")},
			checks: []smcCheck{smcRC("case 14: no stage begin call at all is an undeclared coverage violation", 1),
				smcLacks("case 14: stage end with a guessed name is never checked", "names a guess")}},
		{files: map[string]string{"SKILL.md": block(fastBeg + "<name-or-best-guess> \n")},
			checks: []smcCheck{smcRC("case 15: guessed change name with trailing whitespace is caught", 1),
				smcHas("case 15: finding names the offending argument", "<name-or-best-guess>")}},
		{files: map[string]string{"SKILL.md": block(fastBeg + "'<name-or-best-guess>'\n")},
			checks: []smcCheck{smcRC("case 16: single-quoted guessed change name is caught", 1),
				smcHas("case 16: finding names the offending argument", "<name-or-best-guess>")}},
		{files: map[string]string{"SKILL.md": block(fastBeg + "\"<name-or-best-guess>\"\n")},
			checks: []smcCheck{smcRC("case 17: double-quoted guessed change name is caught", 1),
				smcHas("case 17: finding names the offending argument", "<name-or-best-guess>")}},
		{files: map[string]string{"SKILL.md": block(fastBeg + "<name-or-best-guess>  # resolve later\n")},
			checks: []smcCheck{smcRC("case 18: guessed change name behind a trailing comment is caught", 1),
				smcHas("case 18: finding names the offending argument", "<name-or-best-guess>")}},
		{files: map[string]string{"SKILL.md": block("flow stage begin -command '/flow-fast' -stage flow.kickoff -harness <harness> -session-token \"mf-#abc123\" <name>\n")},
			checks: []smcCheck{smcRC("case 19: a # inside a quoted value is not treated as a comment", 0)}},
		{files: map[string]string{"SKILL.md": block(fastBeg + "'<name-or-best guess>'\n")},
			checks: []smcCheck{smcRC("case 20: quoted guess placeholder with an internal space is caught", 1),
				smcHas("case 20: finding names the offending argument", "<name-or-best guess>")}},
		{files: map[string]string{"SKILL.md": block(`flow stage begin -command '/flow-fast' -stage flow.kickoff -harness <harness> -session-token "mf-\"abc#tok" <name-or-best-guess>` + "\n")},
			checks: []smcCheck{smcRC("case 21: guessed change name past an escaped-quote-plus-# session token is caught", 1),
				smcHas("case 21: finding names the offending argument", "<name-or-best-guess>")}},
		{files: map[string]string{"SKILL.md": "flow stage begin -command '/flow-fast' -stage flow.kickoff -harness <harness> -session-token mf-abc123\t<name-or-best-guess>\n"},
			checks: []smcCheck{smcRC("case 22: tab-separated guessed change name is caught", 1),
				smcHas("case 22: finding names the offending argument", "<name-or-best-guess>")}},
		{files: map[string]string{"SKILL.md": block(begin)},
			checks: []smcCheck{smcRC("case 23: a file with one compliant call exits 0", 0),
				{"case 23: the breakdown names the file with its checked count", func(_ int, out, dir string) bool {
					return strings.Contains(out, smcFile(dir)+" 1")
				}}}},
		{args: func(string) []string { return nil }, real: true,
			checks: []smcCheck{smcRC("case 24: the real repository's own tree is clean", 0),
				smcHas("case 24: flow-status's declared zero carries its by-contract reason",
					"skills/flow-status/SKILL.md 0 (declared: read-only status report"),
				smcHas("case 24: an ordinary declared-zero member also reports as declared",
					"skills/flow-contracts/SKILL.md 0 (declared:")}},
		{files: map[string]string{},
			checks: []smcCheck{smcRC("case 25: an empty target is an undeclared coverage violation, not a vanishing pass", 1),
				{"case 25: the violation names the empty target itself", func(_ int, out, dir string) bool {
					return smcInOrder(out, dir, "0 checked", "not declared expected-zero")
				}}}},
		{files: map[string]string{"SKILL.md": "Ordinary prose with no flow stage marks in it.\n",
			"notes.md": "This file is not named SKILL.md or pipeline.md and can never carry a mark.\n"},
			checks: []smcCheck{smcRC("case 26: a SKILL.md stripped of its marks still fires after F1's narrowing", 1),
				smcLacks("case 26: a file that can never carry a mark is excluded from the corpus entirely", "notes.md")}},
		{files: map[string]string{"SKILL.md": block(begin)},
			args: func(dir string) []string { return []string{smcFile(dir), smcFile(dir)} },
			checks: []smcCheck{smcRC("case 27: a rejected coverage_record aborts loudly instead of reporting a contradictory clean verdict", 2),
				smcHas("case 27: the failure names the real cause", "already recorded")}},
		{files: map[string]string{"SKILL.md": block("flow record dispatch -change <name> -task 3 -role implementer -model opus \\\n",
			`  -commit abc1234 -outcome completed -session-token "mf-$(date +%s)"`+"\n")},
			checks: []smcCheck{smcRC("case 28: a substituted session token on `record dispatch` is caught", 1),
				smcHas("case 28: the finding names the substitution shape", "command substitution")}},
		{files: map[string]string{"SKILL.md": block("flow record dispatch -change <name> -task 3 -role implementer -model opus \\\n",
			"  -commit abc1234 -outcome completed -session-token mf-abc123\n")},
			checks: []smcCheck{smcRC("case 29: a literal session token on `record dispatch` passes", 0),
				{"case 29: the breakdown counts the `record dispatch` call site", func(_ int, out, dir string) bool {
					return strings.Contains(out, smcFile(dir)+" 1")
				}}}},
		{files: map[string]string{"SKILL.md": block("flow record dispatch -change <name> -role reviewer -slot Primary -model sonnet \\\n",
			"  -outcome completed\n")},
			checks: []smcCheck{smcRC("case 30: a `record dispatch` with no -session-token is caught", 1),
				smcHas("case 30: the finding names the missing flag", "carries no -session-token")}},
		{files: map[string]string{"brainstorm.md": block("flow record dispatch begin -change <name> -role planner -model opus \\\n",
			`  -key planner-opus -session-token "mf-$(date +%s)"`+"\n")},
			checks: []smcCheck{smcRC("case 31: a substituted token in a phase file (brainstorm.md) is caught", 1),
				{"case 31: finding names the phase file and the substitution shape", func(_ int, out, dir string) bool {
					return smcInOrder(out, dir+"/brainstorm.md", "command substitution")
				}}}},
		{files: map[string]string{"SKILL.md": block("flow stage begin -command '/flow-plan' -stage plan.bogus -harness <harness> -session-token fp-abc123 -jira-key <KEY>\n",
			"flow stage end   -command '/flow-plan' -stage plan.bogus -jira-key <KEY> -outcome staged\n")},
			checks: []smcCheck{smcRC("case 32: an unlisted -stage key is caught", 1),
				smcHas("case 32: the finding names the unlisted key", "plan.bogus", unknown)}},
		{files: map[string]string{"SKILL.md": block("flow stage begin -command '/flow-plan' -stage plan.session -harness <harness> -session-token fp-abc123 -jira-key <KEY>\n")},
			checks: []smcCheck{smcRC("case 33: a listed -stage key passes", 0)}},
		{files: map[string]string{"SKILL.md": block("flow stage begin -command '/flow' -stage 'flow.kickoff' -harness <harness> -session-token mf-abc123 <name>\n",
			`flow stage begin -command '/flow' -stage "flow.brainstorm" -harness <harness> -session-token mf-abc123 <name>`+"\n")},
			checks: []smcCheck{smcRC("case 34: a quoted listed -stage key passes", 0)}},
		{files: map[string]string{"SKILL.md": block("flow stage begin -command '/flow' -stage flow.kickof -harness <harness> -session-token mf-abc123 <name>\n")},
			checks: []smcCheck{smcRC("case 35: a near-miss substring of a listed -stage key is caught", 1),
				smcHas("case 35: the finding names the near-miss key", "flow.kickof", unknown)}},
		{files: map[string]string{"SKILL.md": block("flow stage begin -command '/flow' -stage flow.kickof. -harness <harness> -session-token mf-abc123 <name>\n")},
			checks: []smcCheck{smcRC("case 36: a metacharacter near-miss -stage key is caught", 1),
				smcHas("case 36: the finding names the near-miss key", "flow.kickof.", unknown)}},
		{files: map[string]string{"SKILL.md": block("flow stage begin -command '/flow' -stage flow.kickof. -harness <harness> -session-token mf-abc123 <name>\n")},
			checks: []smcCheck{smcRC("case 37: the unlisted-key finding still fires", 1),
				{"case 37: the finding's message text is pinned whole", func(_ int, out, dir string) bool {
					return strings.Contains(out, smcFile(dir)+":2: -stage flow.kickof. is not a key of the served stage-key vocabulary (flow stage keys) -- a mark under an unknown key is refused by the daemon as a caller mistake and the stage goes unrecorded; use a listed key or add the row first")
				}}}},
		// Cases 38-39 (KAN-533): the served source is the root's own stats
		// module -- here a sandbox root whose cmd/flow serves keys of its own,
		// with no README.md anywhere.
		{files: map[string]string{"stats/go.mod": sandboxGoMod,
			"stats/cmd/flow/main.go": "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"flow.alpha\")\n\tfmt.Println(\"plan.beta\")\n}\n",
			"skills/SKILL.md":        block("flow stage begin -command '/flow' -stage flow.alpha -harness <harness> -session-token mf-abc123 <name>\n")},
			args: func(dir string) []string { return []string{dir + "/skills"} }, root: func(dir string) string { return dir }, real: true,
			checks: []smcCheck{smcRC("case 38: the guard consumes the sandbox's served keys, with no README anywhere", 0)}},
		{files: map[string]string{"stats/go.mod": sandboxGoMod,
			"stats/cmd/flow/main.go": "package main\n\nfunc main() {\n}\n",
			"skills/SKILL.md":        block("flow stage begin -command '/flow' -stage flow.alpha -harness <harness> -session-token mf-abc123 <name>\n")},
			args: func(dir string) []string { return []string{dir + "/skills"} }, root: func(dir string) string { return dir }, real: true,
			checks: []smcCheck{smcRC("case 39: an empty serve is cannot-answer (exit 2), never a pass", 2),
				smcHas("case 39: the cannot-answer line names the served source", "served stage-key source")}},
	}

	for i, c := range cases {
		// The fixture belongs to the parent test, so it outlives the
		// parallel subtest reading this case's one run.
		dir := t.TempDir()
		for rel, body := range c.files {
			writeFile(t, filepath.Join(dir, rel), body)
		}
		args := []string{dir}
		if c.args != nil {
			args = c.args(dir)
		}
		root := smcRepoRoot
		if c.root != nil {
			root = c.root(dir)
		}
		// One parallel subtest runs the case, its checks nested beneath it:
		// a parallel subtest per check held a -parallel slot apiece while
		// the one running the case worked and the rest waited on it.
		t.Run(fmt.Sprint("run ", i), func(t *testing.T) {
			t.Parallel()
			var out bytes.Buffer
			rc := checkStageMarkCalls(args, smcEnv(root, c.real), &out, &out)
			for _, chk := range c.checks {
				t.Run(chk.label, func(t *testing.T) {
					if !chk.ok(rc, out.String(), dir) {
						t.Fatalf("rc=%d out=%s", rc, out.String())
					}
				})
			}
		})
	}

	// Beyond the harness: the findings quote a value with bash's printf %q,
	// which no harness case pinned. smcQuote is checked against the bash on
	// PATH itself, under a UTF-8 and a C locale.
	for _, locale := range []string{"en_US.UTF-8", "C"} {
		t.Run("port: %q matches bash printf %q under LC_ALL="+locale, func(t *testing.T) {
			t.Parallel()
			inputs := []string{"", "claude-code", "<harness>", "mf-$(date +%s)-$$", "mf-`date +%s`", "mf-$RANDOM",
				"#a", "a#", "~a", "a~", "=a", "a=b", "a b!\"$&'()*,;<>?[\\]^`{|}", "+-./:@_%",
				"a\tb", "a\rb", "\x01\x1b\x7f", "a'b\\c\td", "é", "é\t", "\xff", "\a\b\f\v"}
			var script strings.Builder
			script.WriteString(`for s in "$@"; do printf '%q\0' "$s"; done`)
			cmd := exec.Command("bash", append([]string{"-c", script.String(), "bash"}, inputs...)...)
			cmd.Env = append(os.Environ(), "LC_ALL="+locale)
			got, err := cmd.Output()
			if err != nil {
				t.Fatal(err)
			}
			want := strings.Split(strings.TrimSuffix(string(got), "\x00"), "\x00")
			utf8 := locale != "C"
			for i, in := range inputs {
				if q := smcQuote(in, utf8); q != want[i] {
					t.Errorf("smcQuote(%q) = %s, bash = %s", in, q, want[i])
				}
			}
		})
	}

	// Beyond the harness: mechanisms the bash got implicitly from awk, read
	// and grep are explicit code here, so each is pinned by running the bash
	// at d71a2327 and the port over the same files and comparing stdout,
	// stderr and exit byte for byte (panel review of task 3).
	t.Run("port: implicit awk/read/grep mechanisms match the bash at d71a2327", func(t *testing.T) {
		t.Parallel()
		tree := t.TempDir()
		for _, rel := range []string{"scripts/check-stage-mark-calls.sh", "scripts/lib/coverage.sh"} {
			src, err := exec.Command(fixtureGit, "-C", smcRepoRoot, "show", "d71a2327:"+rel).Output()
			if err != nil {
				t.Fatal(err)
			}
			// The pinned bash pipes the key list into `grep -q` under
			// pipefail: bash's printf writes "%s" and "\n" separately, so a
			// grep that matched and exited between them SIGPIPEs printf and
			// the key reads as unknown — under load, about 1 run in 120. A
			// here-string has no writer to kill.
			src = bytes.ReplaceAll(src, []byte(`printf '%s\n' "$STAGE_KEYS" | grep -qxF -- "$stage_key"`), []byte(`grep -qxF -- "$stage_key" <<<"$STAGE_KEYS"`))
			writeFile(t, filepath.Join(tree, rel), string(src))
		}
		if err := os.Symlink(filepath.Join(smcRepoRoot, "stats"), filepath.Join(tree, "stats")); err != nil {
			t.Fatal(err)
		}
		dir := t.TempDir()
		mark := "flow stage begin -command '/flow' -stage flow.review-panel "
		files := map[string]string{
			// the tab trim of IFS=$'\t' read, around a continuation ending in `\  `
			"tabs/SKILL.md": fence + "\t" + mark + "\\  \n\t-harness <harness> -session-token mf-$x\t\n" + unfence,
			// the CR strip, on a CRLF line and on a bare flag the line ends with
			"crlf/SKILL.md": fence + "flow record dispatch begin -change <name> -session-token mf-$x -key k\r\n" +
				mark + "-session-token mf-abc -harness\r\n" + unfence,
			// the tab trim, on a call ending in a bare flag
			"tabend/SKILL.md": fence + mark + "-session-token mf-abc -harness\t\n" + unfence,
			// the `=` form of the presence checks
			"equals/SKILL.md": fence + mark + "-harness=<h> -session-token=x <name>\n" + unfence,
			// POSIX leftmost-longest value extraction
			"quoted/SKILL.md": fence + mark + `-harness <harness> -session-token "a"$x <name>` + "\n" + unfence,
			// `_` opens a shell variable name
			"underscore/SKILL.md": fence + mark + "-harness <harness> -session-token mf-$_x <name>\n" + unfence,
			// the placeholder shape
			"nested/SKILL.md": fence + mark + "-harness <a><b> -session-token mf-abc <name>\n" + unfence,
			// a call still open at end of file
			"open/SKILL.md": fence + mark + "-harness <harness> \\\n-session-token mf-$x \\\n",
			// awk's own error for a file it cannot open
			"unreadable/SKILL.md": fence + mark + "-harness <harness> -session-token mf-abc <name>\n" + unfence,
		}
		for rel, body := range files {
			writeFile(t, filepath.Join(dir, rel), body)
		}
		if err := os.Chmod(filepath.Join(dir, "unreadable/SKILL.md"), 0); err != nil {
			t.Fatal(err)
		}
		lookup := func(k string) (string, bool) {
			if k == "LC_ALL" {
				return "en_US.UTF-8", true
			}
			return smcEnv(tree, false).LookupEnv(k)
		}
		env := smcEnv(tree, false)
		env.LookupEnv = lookup
		env.Getenv = func(k string) string { v, _ := lookup(k); return v }
		var gout, gerr bytes.Buffer
		grc := checkStageMarkCalls([]string{dir}, env, &gout, &gerr)

		cmd := exec.Command("bash", filepath.Join(tree, "scripts/check-stage-mark-calls.sh"), dir)
		cmd.Dir = tree
		cmd.Env = append(os.Environ(), "LC_ALL=en_US.UTF-8")
		var bout, berr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &bout, &berr
		brc := 0
		if err := cmd.Run(); err != nil {
			ee, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatal(err)
			}
			brc = ee.ExitCode()
		}
		// The hardcoded-harness message was restated after d71a2327, when the
		// Cursor and Codex installs were removed; the pinned bash still names
		// them, so its output is brought to the current wording before the
		// byte comparison. Everything else is compared as the bash printed it.
		bash := strings.ReplaceAll(bout.String(),
			"`~/.claude/skills/`, `~/.cursor/skills/` and `~/.codex/skills/` alike",
			"`~/.claude/skills/` and `~/.zcode/skills/` alike")
		if grc != brc || gout.String() != bash || gerr.String() != berr.String() {
			t.Fatalf("port rc=%d stdout=\n%s\nstderr=\n%s\nbash rc=%d stdout=\n%s\nstderr=\n%s",
				grc, gout.String(), gerr.String(), brc, bash, berr.String())
		}
	})
}

// TestStageMarkCallsChecksStageMark pins that the guard checks a `flow
// stage mark` line by the `stage begin` rules -- a -session-token, a
// placeholder -harness, no guessed name -- and checks every comma-separated
// -stages key against the served vocabulary.
func TestStageMarkCallsChecksStageMark(t *testing.T) {
	t.Parallel()
	const mark = "flow stage mark -command '/flow-fast' "
	cases := []struct {
		name, line string
		rc         int
		want       string
	}{
		{"compliant", mark + "-stages flow.preflight,flow.unfinished-work-gate -harness <harness> -session-token ff-abc <name>", 0, "clean (1 call site(s) checked)"},
		{"missing token", mark + "-stages flow.preflight -harness <harness> <name>", 1, "`flow stage mark` carries no -session-token"},
		{"hardcoded harness", mark + "-stages flow.preflight -harness claude-code -session-token ff-abc <name>", 1, "-harness claude-code is a hardcoded literal"},
		{"missing harness", mark + "-stages flow.preflight -session-token ff-abc <name>", 1, "`flow stage mark` carries no -harness"},
		{"guessed name", mark + "-stages flow.preflight -harness <harness> -session-token ff-abc <guessed-name>", 1, "`stage mark` names a guess"},
		{"unserved key", mark + "-stages flow.preflight,flow.no-such-stage -harness <harness> -session-token ff-abc <name>", 1, "-stage flow.no-such-stage is not a key of the served stage-key vocabulary"},
		{"no -stages", mark + "-harness <harness> -session-token ff-abc <name>", 1, "`stage mark` carries no -stages key"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			writeFile(t, smcFile(dir), "```bash\n"+c.line+"\n```\n")
			var out bytes.Buffer
			rc := checkStageMarkCalls([]string{dir}, smcEnv(smcRepoRoot, false), &out, &out)
			if rc != c.rc || !strings.Contains(out.String(), c.want) {
				t.Fatalf("rc=%d, want %d; output lacks %q:\n%s", rc, c.rc, c.want, out.String())
			}
		})
	}
}
