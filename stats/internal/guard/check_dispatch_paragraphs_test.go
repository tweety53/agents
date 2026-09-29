package guard

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-check-dispatch-paragraphs.sh at d71a2327, one
// subtest per ok: label. The bash harness pointed the guard at a sandboxed
// fixture tree with CHECK_DISPATCH_PARAGRAPHS_ROOT and asserted on its exit
// code and on glob patterns over its combined output; here the guard runs
// in-process with the same override on Env, over a fixture in t.TempDir(),
// and each pattern `*"a"*"b"*` is an ordered-substring check (dpHas). The
// fixture paragraphs below are the harness's own variables, byte for byte:
// the canonical dispatch paragraphs and, for each, one variant per required
// phrase with that phrase dropped while staying a plausible paragraph.
// Case 37 (MODEL HANDSHAKE absent from brainstorm.md) was removed from the
// harness with that site (kan-488) and has no row.

// dpDoc is a site file's body as the harness wrote it: paragraphs separated
// by one blank line (write_site adds the trailing newline).
func dpDoc(paragraphs ...string) string { return strings.Join(paragraphs, "\n\n") }

// dpCheck is one ok: label: the exit code when want is nil, else want's
// substrings in order in the combined output.
type dpCheck struct {
	label string
	rc    int
	want  []string
}

func dpRC(label string, rc int) dpCheck          { return dpCheck{label: label, rc: rc} }
func dpHas(label string, want ...string) dpCheck { return dpCheck{label: label, want: want} }

// dpCase is one harness case: files maps a skills/flow/ basename to its
// body; setup changes the tree after the files are written; vars replaces
// the default CHECK_DISPATCH_PARAGRAPHS_ROOT=<fixture> environment.
type dpCase struct {
	files  map[string]string
	setup  func(t *testing.T, root string)
	vars   map[string]string
	checks []dpCheck
}

// dpFixture is new_root plus the case's files: every root carries a
// visual-verify.md and a visual-verify-tooling-analysis.md, each with its
// three required blocks at its one dispatch site, which a case may overwrite.
func dpFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	flow := filepath.Join(root, "skills", "flow")
	writeFile(t, filepath.Join(flow, "visual-verify.md"), dpCleanVisualVerify+"\n")
	writeFile(t, filepath.Join(flow, "visual-verify-tooling-analysis.md"), dpCleanVisualVerify+"\n")
	for name, body := range files {
		writeFile(t, filepath.Join(flow, name), body+"\n")
	}
	return root
}

// dpRun runs the registered guard; out is stdout and stderr interleaved, as
// the harness's 2>&1 read them.
func dpRun(t *testing.T, env Env) (int, string, string, string) {
	t.Helper()
	fn := Registry["check-dispatch-paragraphs"]
	if fn == nil {
		t.Fatal("check-dispatch-paragraphs is not registered")
	}
	var out, stdout, stderr bytes.Buffer
	rc := fn(nil, env, io.MultiWriter(&out, &stdout), io.MultiWriter(&out, &stderr))
	return rc, out.String(), stdout.String(), stderr.String()
}

func dpSymlinkReviewPanel(t *testing.T, root string) {
	outside := filepath.Join(t.TempDir(), "outside")
	writeFile(t, outside, dpReviewerBlock+"\n")
	if err := os.Symlink(outside, filepath.Join(root, "skills/flow/review-panel.md")); err != nil {
		t.Fatal(err)
	}
}

func dpDirectoryReviewPanel(t *testing.T, root string) {
	mkdir(t, filepath.Join(root, "skills/flow/review-panel.md"))
}

func dpUnreadableImplement(t *testing.T, root string) {
	if os.Geteuid() == 0 {
		t.Skip("running as root, permission bits are not enforced")
	}
	p := filepath.Join(root, "skills/flow/implement.md")
	if err := os.Chmod(p, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(p, 0o644) })
}

func TestCheckDispatchParagraphs(t *testing.T) {
	t.Parallel()
	for _, c := range dpCases {
		for _, chk := range c.checks {
			t.Run(chk.label, func(t *testing.T) {
				t.Parallel()
				root := dpFixture(t, c.files)
				if c.setup != nil {
					c.setup(t, root)
				}
				vars := c.vars
				if vars == nil {
					vars = map[string]string{"CHECK_DISPATCH_PARAGRAPHS_ROOT": root}
				}
				rc, out, _, _ := dpRun(t, crEnv(t.TempDir(), vars))
				if chk.want == nil {
					if rc != chk.rc {
						t.Fatalf("rc = %d, want %d; out:\n%s", rc, chk.rc, out)
					}
					return
				}
				if !dpInOrder(out, chk.want) {
					t.Fatalf("output lacks %q in order; out:\n%s", chk.want, out)
				}
			})
		}
	}

	// Beyond the harness: the exact lines and their stdout/stderr split,
	// which its 2>&1 substring checks never pinned, and the root sources the
	// shim hands over. Each expected text is what the bash guard printed at
	// d71a2327 on the same fixture.
	clean := dpCases[0].files
	noRange := map[string]string{"review-panel.md": dpCleanReviewPanel,
		"implement.md": dpDoc(dpCleanImplementNoReadonly, dpReadonlyBlock, dpOutputBudgetBlockNoRange)}
	extras := []struct {
		label  string
		files  map[string]string
		setup  func(t *testing.T, root string)
		env    func(root string) Env
		rc     int
		stdout func(root string) string
		stderr func(root string) string
	}{
		{label: "a clean run prints the OK line on stdout alone", files: clean, rc: 0,
			stdout: func(root string) string { return "DISPATCH-PARAGRAPHS-OK: " + root + " — 30 site(s) validated\n" }},
		{label: "a relative root prints its findings and the INVALID line as given, on stdout alone", files: noRange,
			env: func(root string) Env {
				return crEnv(filepath.Dir(root), map[string]string{"CHECK_DISPATCH_PARAGRAPHS_ROOT": filepath.Base(root)})
			},
			rc: 1, stdout: func(root string) string {
				b := filepath.Base(root)
				return b + `/skills/flow/implement.md:88: block carrying "**OUTPUT BUDGET:**" is missing the required phrase: "by line range"` + "\n" +
					"DISPATCH-PARAGRAPHS-INVALID: " + b + " — 1 violation(s)\n"
			}},
		{label: "a refusal prints file:0 on stderr alone, dropping earlier sites' findings",
			files: map[string]string{"review-panel.md": dpDoc("No REPRODUCE, DON'T READ paragraph here at all.", dpVerbatimBlock),
				"implement.md": dpDoc(dpReviewerBlock, dpImplementerBlock)},
			setup: func(t *testing.T, root string) {
				if err := os.Remove(filepath.Join(root, "skills/flow/visual-verify.md")); err != nil {
					t.Fatal(err)
				}
			},
			rc: 2, stderr: func(root string) string {
				return root + "/skills/flow/visual-verify.md:0: does not exist — this is a required dispatch-paragraph site\n"
			}},
		{label: "a NUL byte ends a line's text, as bash's mapfile did",
			files: map[string]string{"review-panel.md": clean["review-panel.md"], "implement.md": clean["implement.md"],
				"visual-verify.md": dpCleanVisualVerify +
					"\n> **TOOLS:** in your first turn\x00 never a wildcard query re-prices your whole context"},
			rc: 1, stdout: func(root string) string {
				f := root + `/skills/flow/visual-verify.md:14: block carrying "**TOOLS:**" is missing the required phrase: `
				return f + `"never a wildcard query"` + "\n" + f + `"re-prices your whole context"` + "\n" +
					"DISPATCH-PARAGRAPHS-INVALID: " + root + " — 2 violation(s)\n"
			}},
		{label: "a block missing a shared phrase never satisfies its variant",
			files: map[string]string{"review-panel.md": clean["review-panel.md"],
				"implement.md": strings.Replace(clean["implement.md"], "test you write MUST exercise the real\n> thing", "test you write MUST do the real\n> thing", 1)},
			rc: 1, stdout: func(root string) string {
				f := root + "/skills/flow/implement.md"
				return f + `:7: block carrying "**REPRODUCE, DON'T READ:**" is missing the required phrase: "exercise the real thing"` + "\n" +
					f + `:0: missing a block that satisfies the implementer variant (the label "**REPRODUCE, DON'T READ:**" plus every shared phrase plus "a fake or a hand-built value")` + "\n" +
					"DISPATCH-PARAGRAPHS-INVALID: " + root + " — 2 violation(s)\n"
			}},
		{label: "with no override the root is FLOW_GUARD_REPO_ROOT", files: clean,
			env: func(root string) Env { return crEnv(root, map[string]string{"FLOW_GUARD_REPO_ROOT": root}) },
			rc:  0, stdout: func(root string) string { return "DISPATCH-PARAGRAPHS-OK: " + root + " — 30 site(s) validated\n" }},
		{label: "with no override and no FLOW_GUARD_REPO_ROOT it exits 2", files: clean,
			env: func(root string) Env { return crEnv(root, nil) },
			rc:  2, stderr: func(string) string {
				return "check-dispatch-paragraphs: FLOW_GUARD_REPO_ROOT is unset — run scripts/check-dispatch-paragraphs.sh, which sets it\n"
			}},
		{label: "a root that is not a directory exits 2", files: clean,
			env: func(root string) Env {
				return crEnv(root, map[string]string{"CHECK_DISPATCH_PARAGRAPHS_ROOT": root + "/missing"})
			},
			rc: 2, stderr: func(root string) string {
				return "check-dispatch-paragraphs: " + root + "/missing is not a readable directory — cannot scan\n"
			}},
	}
	for _, x := range extras {
		t.Run(x.label, func(t *testing.T) {
			t.Parallel()
			root := dpFixture(t, x.files)
			if x.setup != nil {
				x.setup(t, root)
			}
			env := crEnv(t.TempDir(), map[string]string{"CHECK_DISPATCH_PARAGRAPHS_ROOT": root})
			if x.env != nil {
				env = x.env(root)
			}
			rc, _, stdout, stderr := dpRun(t, env)
			wantOut, wantErr := "", ""
			if x.stdout != nil {
				wantOut = x.stdout(root)
			}
			if x.stderr != nil {
				wantErr = x.stderr(root)
			}
			if rc != x.rc || stdout != wantOut || stderr != wantErr {
				t.Fatalf("got rc=%d stdout=%q stderr=%q\nwant rc=%d stdout=%q stderr=%q", rc, stdout, stderr, x.rc, wantOut, wantErr)
			}
		})
	}

	// Beyond the harness: three checks no harness case reached (panel review
	// of task 5), each pinned by running the bash at d71a2327 on the same
	// root and comparing stdout, stderr and exit byte for byte.
	bash := filepath.Join(t.TempDir(), "check-dispatch-paragraphs.sh")
	src, err := exec.Command(fixtureGit, "-C", "../../..", "show", "d71a2327:scripts/check-dispatch-paragraphs.sh").Output()
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, bash, string(src))
	parity := []struct {
		label string
		root  func(t *testing.T) string
	}{
		{"port: an unquoted label line is a one-line block, never continued", func(t *testing.T) string {
			return dpFixture(t, map[string]string{"review-panel.md": clean["review-panel.md"], "implement.md": clean["implement.md"],
				"visual-verify.md": dpDoc(strings.TrimPrefix(dpToolsBlock, "> "), dpCleanVisualVerify)})
		}},
		{"port: a root that exists but is not readable exits 2", func(t *testing.T) string {
			if os.Geteuid() == 0 {
				t.Skip("running as root, permission bits are not enforced")
			}
			root := dpFixture(t, clean)
			if err := os.Chmod(root, 0); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(root, 0o755) })
			return root
		}},
		{"port: a root that is an executable file exits 2", func(t *testing.T) string {
			root := filepath.Join(t.TempDir(), "root")
			writeFile(t, root, "#!/bin/sh\n")
			if err := os.Chmod(root, 0o755); err != nil {
				t.Fatal(err)
			}
			return root
		}},
	}
	for _, x := range parity {
		t.Run(x.label, func(t *testing.T) {
			t.Parallel()
			root := x.root(t)
			rc, _, stdout, stderr := dpRun(t, crEnv(t.TempDir(), map[string]string{"CHECK_DISPATCH_PARAGRAPHS_ROOT": root}))
			cmd := exec.Command("bash", bash)
			cmd.Env = append(os.Environ(), "CHECK_DISPATCH_PARAGRAPHS_ROOT="+root)
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
			if rc != brc || stdout != bout.String() || stderr != berr.String() {
				t.Fatalf("port rc=%d stdout=%q stderr=%q\nbash rc=%d stdout=%q stderr=%q", rc, stdout, stderr, brc, bout.String(), berr.String())
			}
		})
	}
}

// dpInOrder is the glob `*"w0"*"w1"*…`: each substring found after the last.
func dpInOrder(s string, want []string) bool {
	for _, w := range want {
		i := strings.Index(s, w)
		if i < 0 {
			return false
		}
		s = s[i+len(w):]
	}
	return true
}

var dpCases = []dpCase{
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlock, dpPixelProbeBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpIndependentBlock, dpDelegationBlock, dpDelegationBlock, dpEntryContextBlock, dpFindingsInputBlock, dpContextBundleFailureBlock, dpOutputBudgetBlock),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpGuardBitesBlock, dpDecideBlock, dpReviewerBlock, dpForegroundBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpDelegationBlock, dpDelegationBlock, dpOutputBudgetBlock, dpReadonlyBlock),
	},
		checks: []dpCheck{
			dpRC("case 1: both sites correct exits 0", 0),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc("No REPRODUCE, DON'T READ paragraph here at all.", dpVerbatimBlock),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock),
	},
		checks: []dpCheck{
			dpRC("case 2: exits 1", 1),
			dpHas("case 2: names review-panel.md", "review-panel.md"),
			dpHas("case 2: names the min-blocks violation", "requires at least 1 block(s)"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlockNoSharedPhrase, dpVerbatimBlock),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock),
	},
		checks: []dpCheck{
			dpRC("case 3: exits 1", 1),
			dpHas("case 3: names the missing phrase", "exercise the real thing"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock),
		"implement.md":    dpReviewerBlock,
	},
		checks: []dpCheck{
			dpRC("case 4: exits 1", 1),
			dpHas("case 4: names implement.md", "implement.md"),
			dpHas("case 4: names the min-blocks violation", "requires at least 2 block(s)"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock),
		"implement.md":    dpDoc(dpReviewerBlock, dpReviewerBlock),
	},
		checks: []dpCheck{
			dpRC("case 5: two same-variant blocks still exits 1", 1),
			dpHas("case 5: names implement.md's missing implementer variant", "implement.md", "implementer"),
		}},
	{files: map[string]string{
		"implement.md": dpDoc(dpReviewerBlock, dpImplementerBlock),
	},
		checks: []dpCheck{
			dpRC("case 6: a missing scoped file exits 2", 2),
			dpHas("case 6: names the missing file", "review-panel.md"),
			dpHas("case 6: reports it through the existence branch", "does not exist"),
		}},
	{files: map[string]string{
		"implement.md": dpDoc(dpReviewerBlock, dpImplementerBlock),
	},
		setup: dpSymlinkReviewPanel,
		checks: []dpCheck{
			dpRC("case 7: a symlinked scoped file exits 2", 2),
			dpHas("case 7: names the symlinked file", "review-panel.md"),
		}},
	{files: map[string]string{
		"review-panel.md": dpReviewerBlock,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock),
	},
		checks: []dpCheck{
			dpRC("case 8: exits 1", 1),
			dpHas("case 8: names review-panel.md and the missing VERBATIM REPORT block", "review-panel.md", "VERBATIM REPORT"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlockNoReviewersOwnReport),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock),
	},
		checks: []dpCheck{
			dpRC("case 9: exits 1", 1),
			dpHas("case 9: names the missing phrase", "the reviewer's own report"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlockNoNeverASourceOfFact),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock),
	},
		checks: []dpCheck{
			dpRC("case 10: exits 1", 1),
			dpHas("case 10: names the missing phrase", "never a source of fact"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlockNoReportWins),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock),
	},
		checks: []dpCheck{
			dpRC("case 11: exits 1", 1),
			dpHas("case 11: names the missing phrase", "the report wins"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock),
	},
		vars: map[string]string{"CHECK_DISPATCH_PARAGRAPHS_ROOT": ""},
		checks: []dpCheck{
			dpRC("case 12: a set-but-empty root exits 2", 2),
			dpHas("case 12: names the set-but-empty refusal", "set but empty"),
		}},
	{files: map[string]string{
		"implement.md": dpDoc(dpReviewerBlock, dpImplementerBlock),
	},
		setup: dpDirectoryReviewPanel,
		checks: []dpCheck{
			dpRC("case 13: a directory at a site path exits 2", 2),
			dpHas("case 13: names the not-a-regular-file refusal", "not a regular file"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock),
	},
		setup: dpUnreadableImplement,
		checks: []dpCheck{
			dpRC("case 14: an unreadable site file exits 2", 2),
			dpHas("case 14: names the not-readable refusal", "is not readable"),
		}},
	{files: map[string]string{
		"review-panel.md": "> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a\n> guard, a real transcript, a real process — please just read the code and trust it, worth less than one you did.\n> **VERBATIM REPORT — THE FACT:** the reviewer's own report, never a source of fact, the report wins, exercise the real thing",
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock),
	},
		checks: []dpCheck{
			dpRC("case 15: a glued deficient block still exits 1", 1),
			dpHas("case 15: names the phrase the glued block did not actually carry", "exercise the real thing"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock),
	},
		checks: []dpCheck{
			dpRC("case 16: exits 1", 1),
			dpHas("case 16: names implement.md and the missing FOREGROUND BUILDS block", "implement.md", "FOREGROUND BUILDS"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlockNoStillExecuting, dpForegroundBlock),
	},
		checks: []dpCheck{
			dpRC("case 17: exits 1", 1),
			dpHas("case 17: names the missing phrase", "still executing in the background"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlockNoRunForeground, dpForegroundBlock),
	},
		checks: []dpCheck{
			dpRC("case 18: exits 1", 1),
			dpHas("case 18: names the missing phrase", "Run it in the foreground"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlockNoPollCompletion, dpForegroundBlock),
	},
		checks: []dpCheck{
			dpRC("case 19: exits 1", 1),
			dpHas("case 19: names the missing phrase", "poll it to completion"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock),
	},
		checks: []dpCheck{
			dpRC("case 20: exits 1", 1),
			dpHas("case 20: names implement.md", "implement.md"),
			dpHas("case 20: names the min-blocks violation at its own threshold", "requires at least 3 block(s) carrying the label \"**FOREGROUND BUILDS:**\", found 1"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock),
	},
		checks: []dpCheck{
			dpRC("case 21: exits 1", 1),
			dpHas("case 21: names review-panel.md", "review-panel.md"),
			dpHas("case 21: names the min-blocks violation at its own threshold", "requires at least 2 block(s) carrying the label \"**FOREGROUND BUILDS:**\", found 1"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock),
	},
		checks: []dpCheck{
			dpRC("case 22: exits 1", 1),
			dpHas("case 22: names implement.md and the missing TARGETED TESTS block", "implement.md", "TARGETED TESTS"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlockNoSelector),
	},
		checks: []dpCheck{
			dpRC("case 23: exits 1", 1),
			dpHas("case 23: names implement.md and the missing phrase", "implement.md", "the build tool", "own selector"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlockNoRedGreen),
	},
		checks: []dpCheck{
			dpRC("case 24: exits 1", 1),
			dpHas("case 24: names implement.md and the missing phrase", "implement.md", "once for RED, once for GREEN"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlockNoSuite),
	},
		checks: []dpCheck{
			dpRC("case 25: exits 1", 1),
			dpHas("case 25: names implement.md and the missing phrase", "implement.md", "module or repository suite mid-task"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock),
	},
		checks: []dpCheck{
			dpRC("case 26: exits 1", 1),
			dpHas("case 26: names review-panel.md and the missing TARGETED TESTS block", "review-panel.md", "TARGETED TESTS"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock),
	},
		checks: []dpCheck{
			dpRC("case 27: exits 1", 1),
			dpHas("case 27: names review-panel.md and the missing MUTATION PROOF block", "review-panel.md", "MUTATION PROOF"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlockNoMutationProved),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock),
	},
		checks: []dpCheck{
			dpRC("case 28: exits 1", 1),
			dpHas("case 28: names the missing phrase", "mutation-proved before you end your turn"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlockNoConfirmAndRestore),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock),
	},
		checks: []dpCheck{
			dpRC("case 29: exits 1", 1),
			dpHas("case 29: names the missing phrase", "confirm an existing test fails, and restore"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlockNoSurvivingMutant),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock),
	},
		checks: []dpCheck{
			dpRC("case 30: exits 1", 1),
			dpHas("case 30: names the missing phrase", "a surviving mutant"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlockNoEditLanded),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock),
	},
		checks: []dpCheck{
			dpRC("case 42: exits 1", 1),
			dpHas("case 42: names the missing phrase", "confirm the edit landed"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlockNoRefusal),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock),
	},
		checks: []dpCheck{
			dpRC("case 43: exits 1", 1),
			dpHas("case 43: names the missing phrase", "a refusal, not a surviving mutant"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlockNoNeverBuys),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock),
	},
		checks: []dpCheck{
			dpRC("case 44: exits 1", 1),
			dpHas("case 44: names the missing phrase", "never buys a test"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlockNoShape),
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock),
	},
		checks: []dpCheck{
			dpRC("case 91: a MUTATION PROOF block that cites the fenced block instead of stating the fix-mutation: shape exits 1", 1),
			dpHas("case 91: names the missing phrase", "missing the required phrase: \"`fix-mutation: <path> — <what was mutated> — <the test that failed>`\""),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, "No TOOLS paragraph here at all, just prose.", dpHandshakeBlock, dpDelegationBlock),
	},
		checks: []dpCheck{
			dpRC("case 31: exits 1", 1),
			dpHas("case 31: names implement.md and TOOLS", "implement.md", "TOOLS"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlockNoFirstTurn),
	},
		checks: []dpCheck{
			dpRC("case 32: exits 1", 1),
			dpHas("case 32: names the dropped phrase", "in your first turn"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlockNoWildcard),
	},
		checks: []dpCheck{
			dpRC("case 33: exits 1", 1),
			dpHas("case 33: names the dropped phrase", "never a wildcard query"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlockNoReprices),
	},
		checks: []dpCheck{
			dpRC("case 34: exits 1", 1),
			dpHas("case 34: names the dropped phrase", "re-prices your whole context"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlock),
	},
		checks: []dpCheck{
			dpRC("case 35: exits 1", 1),
			dpHas("case 35: names implement.md and the missing MODEL HANDSHAKE block", "implement.md", "MODEL HANDSHAKE"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlock, dpToolsBlock, dpToolsBlock),
		"implement.md":    dpCleanImplement,
	},
		checks: []dpCheck{
			dpRC("case 36: exits 1", 1),
			dpHas("case 36: names review-panel.md and the missing MODEL HANDSHAKE block", "review-panel.md", "MODEL HANDSHAKE"),
		}},
	{files: map[string]string{
		"visual-verify.md": "No MODEL HANDSHAKE paragraph here at all, just prose.",
		"review-panel.md":  dpCleanReviewPanel,
		"implement.md":     dpCleanImplement,
	},
		checks: []dpCheck{
			dpRC("case 38: exits 1", 1),
			dpHas("case 38: names visual-verify.md and MODEL HANDSHAKE", "visual-verify.md", "MODEL HANDSHAKE"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlockNoFirstLine),
	},
		checks: []dpCheck{
			dpRC("case 39: exits 1", 1),
			dpHas("case 39: names the dropped phrase", "the first line of your first reply"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlockNoNothingElse),
	},
		checks: []dpCheck{
			dpRC("case 40: exits 1", 1),
			dpHas("case 40: names the dropped phrase", "and nothing else on that line"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlockNoBeforeToolCall),
	},
		checks: []dpCheck{
			dpRC("case 41: exits 1", 1),
			dpHas("case 41: names the dropped phrase", "before any tool call"),
		}},
	{files: map[string]string{
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock),
		"implement.md":    dpCleanImplement,
	},
		checks: []dpCheck{
			dpRC("case 45: exits 1", 1),
			dpHas("case 45: names review-panel.md and the missing INDEPENDENT PASSES block", "review-panel.md", "INDEPENDENT PASSES"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock),
	},
		checks: []dpCheck{
			dpRC("case 46: exits 1", 1),
			dpHas("case 46: names implement.md and the missing NO DELEGATION block", "implement.md", "NO DELEGATION"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpIndependentBlock, dpDelegationBlock, dpDelegationBlockNoNeverCallAgent),
	},
		checks: []dpCheck{
			dpRC("case 47: exits 1", 1),
			dpHas("case 47: names the dropped phrase", "review-panel.md", "missing the required phrase: \"Never call the `Agent` tool\""),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpIndependentBlock, dpDelegationBlock, dpDelegationBlockNoNeverSpawn),
	},
		checks: []dpCheck{
			dpRC("case 48: exits 1", 1),
			dpHas("case 48: names the dropped phrase", "review-panel.md", "missing the required phrase: \"never spawn a subagent\""),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpIndependentBlock, dpDelegationBlock, dpDelegationBlockNoLeaf),
	},
		checks: []dpCheck{
			dpRC("case 49: exits 1", 1),
			dpHas("case 49: names the dropped phrase", "review-panel.md", "missing the required phrase: \"the leaf of this run\""),
		}},
	{files: map[string]string{
		"visual-verify-tooling-analysis.md": "No dispatch paragraphs here at all, just prose.",
		"review-panel.md":                   dpCleanReviewPanel,
		"implement.md":                      dpCleanImplement,
	},
		checks: []dpCheck{
			dpRC("case 50a: no block of each in visual-verify-tooling-analysis.md exits 1", 1),
			dpHas("case 50a: names the TOOLS min-blocks violation at its own threshold", "visual-verify-tooling-analysis.md", "requires at least 1 block(s) carrying the label \"**TOOLS:**\", found 0"),
			dpHas("case 50a: names the MODEL HANDSHAKE min-blocks violation at its own threshold", "visual-verify-tooling-analysis.md", "requires at least 1 block(s) carrying the label \"**MODEL HANDSHAKE:**\", found 0"),
			dpHas("case 50a: names the NO DELEGATION min-blocks violation at its own threshold", "visual-verify-tooling-analysis.md", "requires at least 1 block(s) carrying the label \"**NO DELEGATION:**\", found 0"),
		}},
	{files: map[string]string{
		"visual-verify.md": dpDoc(dpToolsBlock, dpHandshakeBlock),
		"review-panel.md":  dpCleanReviewPanel,
		"implement.md":     dpCleanImplement,
	},
		checks: []dpCheck{
			dpRC("case 50: exits 1", 1),
			dpHas("case 50: names visual-verify.md and the missing NO DELEGATION block", "visual-verify.md", "NO DELEGATION"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpIndependentBlock, dpDelegationBlock),
	},
		checks: []dpCheck{
			dpRC("case 51: exits 1", 1),
			dpHas("case 51: names the min-blocks violation at its own threshold", "review-panel.md", "requires at least 2 block(s) carrying the label \"**NO DELEGATION:**\", found 1"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpReviewerBlock, dpForegroundBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpHandshakeBlock, dpDelegationBlock, dpDelegationBlock),
	},
		checks: []dpCheck{
			dpRC("case 52: exactly one TOOLS and one MODEL HANDSHAKE block in implement.md exits 1", 1),
			dpHas("case 52: names the TOOLS min-blocks violation at its own threshold", "requires at least 2 block(s) carrying the label \"**TOOLS:**\", found 1"),
			dpHas("case 52: names the MODEL HANDSHAKE min-blocks violation at its own threshold", "requires at least 2 block(s) carrying the label \"**MODEL HANDSHAKE:**\", found 1"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpForegroundBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpDelegationBlock, dpDelegationBlock),
	},
		checks: []dpCheck{
			dpRC("case 53: exits 1", 1),
			dpHas("case 53: names the REPRODUCE min-blocks violation at its own threshold", "requires at least 3 block(s) carrying the label \"**REPRODUCE, DON'T READ:**\", found 2"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpReviewerBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpDelegationBlock, dpDelegationBlock),
	},
		checks: []dpCheck{
			dpRC("case 54: exits 1", 1),
			dpHas("case 54: names the FOREGROUND min-blocks violation at its own threshold", "requires at least 3 block(s) carrying the label \"**FOREGROUND BUILDS:**\", found 2"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpReviewerBlock, dpForegroundBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpDelegationBlock),
	},
		checks: []dpCheck{
			dpRC("case 55: exits 1", 1),
			dpHas("case 55: names the NO DELEGATION min-blocks violation at its own threshold", "requires at least 2 block(s) carrying the label \"**NO DELEGATION:**\", found 1"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpIndependentBlock, dpDelegationBlock, dpDelegationBlock),
	},
		checks: []dpCheck{
			dpRC("case 56: exits 1", 1),
			dpHas("case 56: names review-panel.md and the missing PIXEL PROBE block", "review-panel.md", "PIXEL PROBE"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlock, dpPixelProbeBlockNoClass, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpIndependentBlock, dpDelegationBlock, dpDelegationBlock),
	},
		checks: []dpCheck{
			dpRC("case 57: exits 1", 1),
			dpHas("case 57: names the missing phrase", "draw/geometry code"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlock, dpPixelProbeBlockNoPixels, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpIndependentBlock, dpDelegationBlock, dpDelegationBlock),
	},
		checks: []dpCheck{
			dpRC("case 58: exits 1", 1),
			dpHas("case 58: names the missing phrase", "actual rendered pixels or geometry"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlock, dpPixelProbeBlockNoRejection, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpIndependentBlock, dpDelegationBlock, dpDelegationBlock),
	},
		checks: []dpCheck{
			dpRC("case 59: exits 1", 1),
			dpHas("case 59: names the missing phrase", "rejected at the fix step"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpReviewerBlock, dpForegroundBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpDelegationBlock, dpDelegationBlock),
	},
		checks: []dpCheck{
			dpRC("case 60: exits 1", 1),
			dpHas("case 60: names implement.md and the missing PROVE THE GUARD BITES block", "implement.md", "PROVE THE GUARD BITES"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpGuardBitesBlockNoScope, dpReviewerBlock, dpForegroundBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpDelegationBlock, dpDelegationBlock),
	},
		checks: []dpCheck{
			dpRC("case 61: exits 1", 1),
			dpHas("case 61: names the missing phrase", "assert on configuration or file content"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpGuardBitesBlockNoBreak, dpReviewerBlock, dpForegroundBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpDelegationBlock, dpDelegationBlock),
	},
		checks: []dpCheck{
			dpRC("case 62: exits 1", 1),
			dpHas("case 62: names the missing phrase", "break the property the assertion protects"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpGuardBitesBlockNoBrokenState, dpReviewerBlock, dpForegroundBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpDelegationBlock, dpDelegationBlock),
	},
		checks: []dpCheck{
			dpRC("case 63: exits 1", 1),
			dpHas("case 63: names the missing phrase", "against the broken state"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpGuardBitesBlock, dpReviewerBlock, dpForegroundBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpDelegationBlock, dpDelegationBlock),
	},
		checks: []dpCheck{
			dpRC("case 64: exits 1", 1),
			dpHas("case 64: names implement.md and the missing REPORT, DON'T DECIDE block", "implement.md", "REPORT, DON'T DECIDE"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpGuardBitesBlock, dpDecideBlockNoOperator, dpReviewerBlock, dpForegroundBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpDelegationBlock, dpDelegationBlock),
	},
		checks: []dpCheck{
			dpRC("case 65: exits 1", 1),
			dpHas("case 65: names the missing phrase", "only the operator can settle"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpGuardBitesBlock, dpDecideBlockNoReported, dpReviewerBlock, dpForegroundBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpDelegationBlock, dpDelegationBlock),
	},
		checks: []dpCheck{
			dpRC("case 66: exits 1", 1),
			dpHas("case 66: names the missing phrase", "reported, never decided in code"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpGuardBitesBlock, dpDecideBlockNoCarry, dpReviewerBlock, dpForegroundBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpDelegationBlock, dpDelegationBlock),
	},
		checks: []dpCheck{
			dpRC("case 67: exits 1", 1),
			dpHas("case 67: names the missing phrase", "carry the question in your REPORT FILE"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpReviewerBlock, dpImplementerBlock, dpGuardBitesBlock, dpDecideBlockNoBlocked, dpReviewerBlock, dpForegroundBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpDelegationBlock, dpDelegationBlock),
	},
		checks: []dpCheck{
			dpRC("case 68: exits 1", 1),
			dpHas("case 68: names the missing phrase", "report BLOCKED"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlock, dpPixelProbeBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpIndependentBlock, dpDelegationBlock, dpDelegationBlock, dpFindingsInputBlock),
	},
		checks: []dpCheck{
			dpRC("case 69: exits 1", 1),
			dpHas("case 69: names review-panel.md and the missing ENTRY CONTEXT block", "review-panel.md", "ENTRY CONTEXT"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpCleanReviewPanel, dpEntryContextBlockNoNamed),
	},
		checks: []dpCheck{
			dpRC("case 70: exits 1", 1),
			dpHas("case 70: names the missing phrase", "your named entry context"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpCleanReviewPanel, dpEntryContextBlockNoWholeTree),
	},
		checks: []dpCheck{
			dpRC("case 71: exits 1", 1),
			dpHas("case 71: names the missing phrase", "not from a whole-tree exploration"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpCleanReviewPanel, dpEntryContextBlockNoCoverage),
	},
		checks: []dpCheck{
			dpRC("case 72: exits 1", 1),
			dpHas("case 72: names the missing phrase", "the coverage trade"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlock, dpPixelProbeBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpIndependentBlock, dpDelegationBlock, dpDelegationBlock, dpEntryContextBlock),
	},
		checks: []dpCheck{
			dpRC("case 73: exits 1", 1),
			dpHas("case 73: names review-panel.md and the missing FINDINGS ARE INPUT block", "review-panel.md", "FINDINGS ARE INPUT"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpCleanReviewPanel, dpFindingsInputBlockNoInput),
	},
		checks: []dpCheck{
			dpRC("case 74: exits 1", 1),
			dpHas("case 74: names the missing phrase", "input, not orders"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpCleanReviewPanel, dpFindingsInputBlockNoResolve),
	},
		checks: []dpCheck{
			dpRC("case 75: exits 1", 1),
			dpHas("case 75: names the missing phrase", "resolve the defect the finding names"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpCleanReviewPanel, dpFindingsInputBlockNoJustify),
	},
		checks: []dpCheck{
			dpRC("case 76: exits 1", 1),
			dpHas("case 76: names the missing phrase", "records the deviation and justifies it"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpCleanReviewPanel, dpFindingsInputBlockNoSilent),
	},
		checks: []dpCheck{
			dpRC("case 77: exits 1", 1),
			dpHas("case 77: names the missing phrase", "A silent deviation is an unfixed finding"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlock, dpPixelProbeBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpIndependentBlock, dpDelegationBlock, dpDelegationBlock, dpEntryContextBlock, dpFindingsInputBlock),
	},
		checks: []dpCheck{
			dpRC("case 78: exits 1", 1),
			dpHas("case 78: names review-panel.md and the missing CONTEXT BUNDLE FAILURE block", "review-panel.md", "CONTEXT BUNDLE FAILURE"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpCleanReviewPanel, dpContextBundleFailureBlockNoNonzero),
	},
		checks: []dpCheck{
			dpRC("case 79: exits 1", 1),
			dpHas("case 79: names the missing phrase", "exited non-zero, or the bundle file is still"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpCleanReviewPanel, dpContextBundleFailureBlockNoBuildFailed),
	},
		checks: []dpCheck{
			dpRC("case 80: exits 1", 1),
			dpHas("case 80: names the missing phrase", "context bundle: build failed"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpCleanReviewPanel, dpContextBundleFailureBlockNoDispatch),
	},
		checks: []dpCheck{
			dpRC("case 81: exits 1", 1),
			dpHas("case 81: names the missing phrase", "dispatch every slot without the bundle"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpCleanReviewPanel, dpContextBundleFailureBlockNoOverride),
	},
		checks: []dpCheck{
			dpRC("case 82: exits 1", 1),
			dpHas("case 82: names the missing phrase", "context bundle: dispatched without it — operator override"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpCleanReviewPanel, dpContextBundleFailureBlockNoStopped),
	},
		checks: []dpCheck{
			dpRC("case 83: exits 1", 1),
			dpHas("case 83: names the missing phrase", "outcome stopped"),
		}},
	{files: map[string]string{
		"implement.md":    dpDoc(dpCleanImplementNoReadonly, dpReadonlyBlock),
		"review-panel.md": dpCleanReviewPanel,
	},
		checks: []dpCheck{
			dpRC("case 84: exits 1", 1),
			dpHas("case 84: names implement.md and the missing OUTPUT BUDGET block", "implement.md", "OUTPUT BUDGET"),
		}},
	{files: map[string]string{
		"implement.md":    dpCleanImplement,
		"review-panel.md": dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlock, dpPixelProbeBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpIndependentBlock, dpDelegationBlock, dpDelegationBlock, dpEntryContextBlock, dpFindingsInputBlock, dpContextBundleFailureBlock),
	},
		checks: []dpCheck{
			dpRC("case 85: exits 1", 1),
			dpHas("case 85: names review-panel.md and the missing OUTPUT BUDGET block", "review-panel.md", "OUTPUT BUDGET"),
		}},
	{files: map[string]string{
		"implement.md":    dpDoc(dpCleanImplementNoReadonly, dpReadonlyBlock, dpOutputBudgetBlockNoRange),
		"review-panel.md": dpCleanReviewPanel,
	},
		checks: []dpCheck{
			dpRC("case 86: exits 1", 1),
			dpHas("case 86: names the missing phrase", "by line range"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpCleanImplementNoReadonly,
	},
		checks: []dpCheck{
			dpRC("case 87: exits 1", 1),
			dpHas("case 87: names implement.md and the missing READ-ONLY REVIEW block", "implement.md", "READ-ONLY REVIEW"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpCleanImplementNoReadonly, dpReadonlyBlockNoNeverMutate),
	},
		checks: []dpCheck{
			dpRC("case 88: exits 1", 1),
			dpHas("case 88: names the missing phrase", "never mutate it"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpCleanImplementNoReadonly, dpReadonlyBlockNoFileEdit),
	},
		checks: []dpCheck{
			dpRC("case 89: exits 1", 1),
			dpHas("case 89: names the missing phrase", "no file edit outside your own report file"),
		}},
	{files: map[string]string{
		"review-panel.md": dpCleanReviewPanel,
		"implement.md":    dpDoc(dpCleanImplementNoReadonly, dpReadonlyBlockNoClaim),
	},
		checks: []dpCheck{
			dpRC("case 90: exits 1", 1),
			dpHas("case 90: names the missing phrase", "a claim nobody can check"),
		}},
}

// The harness's fixture paragraphs.
const (
	dpReviewerBlock = "> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a\n" +
		"> guard, a real transcript, a real process — at least one check you make MUST exercise the real\n" +
		"> thing. A claim you did not run is worth less than one you did: a doc comment, a type signature\n" +
		"> and a passing test can each read plausibly and be false. Run it before you accept it, and run it\n" +
		"> before you reject it."
	dpImplementerBlock = "> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a\n" +
		"> guard, a real transcript, a real process — at least one test you write MUST exercise the real\n" +
		"> thing. A test backed by a fake or a hand-built value passes while the real integration is broken:\n" +
		"> the shape you construct by hand is not the shape the real producer emits. Build the value the way\n" +
		"> production builds it, or assert against the real boundary."
	dpReviewerBlockNoSharedPhrase = "> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a\n" +
		"> guard, a real transcript, a real process — at least one check you make MUST do the real\n" +
		"> thing. A claim you did not run is worth less than one you did: a doc comment, a type signature\n" +
		"> and a passing test can each read plausibly and be false. Run it before you accept it, and run it\n" +
		"> before you reject it."
	dpVerbatimBlock = "> **VERBATIM REPORT — THE FACT:** each finding below names the file its slot's report was written\n" +
		"> to. That file is the reviewer's own report, unedited — read it before you act on the finding.\n" +
		"> The structured block is the dispatcher's summary of it: direction on what to work on, and\n" +
		"> never a source of fact. Where the two disagree the report wins. Where the block asserts\n" +
		"> something the report does not, treat it as unchecked and establish it yourself before building\n" +
		"> on it."
	dpVerbatimBlockNoReviewersOwnReport = "> **VERBATIM REPORT — THE FACT:** each finding below names the file its slot's report was written\n" +
		"> to. That file is the report, unedited — read it before you act on the finding.\n" +
		"> The structured block is the dispatcher's summary of it: direction on what to work on, and\n" +
		"> never a source of fact. Where the two disagree the report wins. Where the block asserts\n" +
		"> something the report does not, treat it as unchecked and establish it yourself before building\n" +
		"> on it."
	dpVerbatimBlockNoNeverASourceOfFact = "> **VERBATIM REPORT — THE FACT:** each finding below names the file its slot's report was written\n" +
		"> to. That file is the reviewer's own report, unedited — read it before you act on the finding.\n" +
		"> The structured block is the dispatcher's summary of it: direction on what to work on, and\n" +
		"> is not itself a fact. Where the two disagree the report wins. Where the block asserts\n" +
		"> something the report does not, treat it as unchecked and establish it yourself before building\n" +
		"> on it."
	dpVerbatimBlockNoReportWins = "> **VERBATIM REPORT — THE FACT:** each finding below names the file its slot's report was written\n" +
		"> to. That file is the reviewer's own report, unedited — read it before you act on the finding.\n" +
		"> The structured block is the dispatcher's summary of it: direction on what to work on, and\n" +
		"> never a source of fact. Where the two disagree the reviewer's report is followed. Where the block asserts\n" +
		"> something the report does not, treat it as unchecked and establish it yourself before building\n" +
		"> on it."
	dpForegroundBlock = "> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running\n" +
		"> command still executing in the background. Run it in the foreground, or poll it to\n" +
		"> completion, before you stop."
	dpForegroundBlockNoStillExecuting = "> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running\n" +
		"> command left running in the background. Run it in the foreground, or poll it to\n" +
		"> completion, before you stop."
	dpForegroundBlockNoRunForeground = "> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running\n" +
		"> command still executing in the background. Bring it to the foreground, or poll it to\n" +
		"> completion, before you stop."
	dpForegroundBlockNoPollCompletion = "> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running\n" +
		"> command still executing in the background. Run it in the foreground, or wait for it to\n" +
		"> finish, before you stop."
	dpTargetedBlock = "> **TARGETED TESTS:** Run only the tests this task's `**Tests:**` field names, through the build\n" +
		"> tool's own selector — `--tests '<class>'` for Gradle, `-run '<name>'` for `go test`, `-t\n" +
		"> '<name>'` for vitest — once for RED, once for GREEN, and again only after a source edit. Never\n" +
		"> run the module or repository suite mid-task: the full `## test` list runs once per worktree at\n" +
		"> the last bundle, and again in `flow.verify`. Pipe a test run's output through `tail` so a green\n" +
		"> run costs lines of context, not a build log."
	dpTargetedBlockNoSelector = "> **TARGETED TESTS:** Run only the tests this task's `**Tests:**` field names, through whichever\n" +
		"> selector applies — `--tests '<class>'` for Gradle, `-run '<name>'` for `go test`, `-t\n" +
		"> '<name>'` for vitest — once for RED, once for GREEN, and again only after a source edit. Never\n" +
		"> run the module or repository suite mid-task: the full `## test` list runs once per worktree at\n" +
		"> the last bundle, and again in `flow.verify`. Pipe a test run's output through `tail` so a green\n" +
		"> run costs lines of context, not a build log."
	dpTargetedBlockNoRedGreen = "> **TARGETED TESTS:** Run only the tests this task's `**Tests:**` field names, through the build\n" +
		"> tool's own selector — `--tests '<class>'` for Gradle, `-run '<name>'` for `go test`, `-t\n" +
		"> '<name>'` for vitest — once before the fix and once after, and again only after a source edit.\n" +
		"> Never run the module or repository suite mid-task: the full `## test` list runs once per worktree\n" +
		"> at the last bundle, and again in `flow.verify`. Pipe a test run's output through `tail` so a\n" +
		"> green run costs lines of context, not a build log."
	dpTargetedBlockNoSuite = "> **TARGETED TESTS:** Run only the tests this task's `**Tests:**` field names, through the build\n" +
		"> tool's own selector — `--tests '<class>'` for Gradle, `-run '<name>'` for `go test`, `-t\n" +
		"> '<name>'` for vitest — once for RED, once for GREEN, and again only after a source edit. Never\n" +
		"> run the whole test suite mid-task: the full `## test` list runs once per worktree at\n" +
		"> the last bundle, and again in `flow.verify`. Pipe a test run's output through `tail` so a green\n" +
		"> run costs lines of context, not a build log."
	dpMutationBlock = "> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before you\n" +
		"> end your turn — not only the test cases this round adds. Mutate the mechanism: revert it in a\n" +
		"> scratch tree, or flip the single value it turns on — but before the tests run, confirm the edit\n" +
		"> landed: the target changed where you intended, not nowhere and not somewhere else. An edit that\n" +
		"> never applied is a refusal, not a surviving mutant — redo it with a working mechanism; it never\n" +
		"> buys a test. Then confirm an existing test fails, and restore.\n" +
		"> `mutate-and-verify.sh` mechanizes backup, apply, run, report and restore\n" +
		"> for a mutation expressed as a patch file against one or more test harnesses; which mechanism to\n" +
		"> mutate is your judgment, not the script's. Each mutation alters one mechanism — where a single\n" +
		"> revert would also change state a second check reads, split it into surgical mutations, one per\n" +
		"> mechanism. A mutation no test catches is a surviving mutant: add the test that catches it before\n" +
		"> your turn ends. Record one `fix-mutation:` line per behaviour in your report, plus a\n" +
		"> `fix-mutations-total:` count, in exactly this shape: `fix-mutation: <path> — <what was mutated> —\n" +
		"> <the test that failed>`; `fix-mutation: <path> — none — <reason>` for a behaviour you exempt;\n" +
		"> and one `fix-mutations-total: <n>` line after them. Where\n" +
		"> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather\n" +
		"> than deciding it yourself."
	dpMutationBlockNoMutationProved = "> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before your\n" +
		"> turn ends — not only the test cases this round adds. Mutate the mechanism: revert it in a\n" +
		"> scratch tree, or flip the single value it turns on — but before the tests run, confirm the edit\n" +
		"> landed: the target changed where you intended, not nowhere and not somewhere else. An edit that\n" +
		"> never applied is a refusal, not a surviving mutant — redo it with a working mechanism; it never\n" +
		"> buys a test. Then confirm an existing test fails, and restore.\n" +
		"> `mutate-and-verify.sh` mechanizes backup, apply, run, report and restore\n" +
		"> for a mutation expressed as a patch file against one or more test harnesses; which mechanism to\n" +
		"> mutate is your judgment, not the script's. Each mutation alters one mechanism — where a single\n" +
		"> revert would also change state a second check reads, split it into surgical mutations, one per\n" +
		"> mechanism. A mutation no test catches is a surviving mutant: add the test that catches it before\n" +
		"> your turn ends. Record one `fix-mutation:` line per behaviour in your report, plus a\n" +
		"> `fix-mutations-total:` count, in exactly this shape: `fix-mutation: <path> — <what was mutated> —\n" +
		"> <the test that failed>`; `fix-mutation: <path> — none — <reason>` for a behaviour you exempt;\n" +
		"> and one `fix-mutations-total: <n>` line after them. Where\n" +
		"> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather\n" +
		"> than deciding it yourself."
	dpMutationBlockNoConfirmAndRestore = "> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before you\n" +
		"> end your turn — not only the test cases this round adds. Mutate the mechanism: revert it in a\n" +
		"> scratch tree, or flip the single value it turns on — but before the tests run, confirm the edit\n" +
		"> landed: the target changed where you intended, not nowhere and not somewhere else. An edit that\n" +
		"> never applied is a refusal, not a surviving mutant — redo it with a working mechanism; it never\n" +
		"> buys a test. Then check that an existing test fails, and restore.\n" +
		"> `mutate-and-verify.sh` mechanizes backup, apply, run, report and restore\n" +
		"> for a mutation expressed as a patch file against one or more test harnesses; which mechanism to\n" +
		"> mutate is your judgment, not the script's. Each mutation alters one mechanism — where a single\n" +
		"> revert would also change state a second check reads, split it into surgical mutations, one per\n" +
		"> mechanism. A mutation no test catches is a surviving mutant: add the test that catches it before\n" +
		"> your turn ends. Record one `fix-mutation:` line per behaviour in your report, plus a\n" +
		"> `fix-mutations-total:` count, in exactly this shape: `fix-mutation: <path> — <what was mutated> —\n" +
		"> <the test that failed>`; `fix-mutation: <path> — none — <reason>` for a behaviour you exempt;\n" +
		"> and one `fix-mutations-total: <n>` line after them. Where\n" +
		"> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather\n" +
		"> than deciding it yourself."
	dpMutationBlockNoSurvivingMutant = "> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before you\n" +
		"> end your turn — not only the test cases this round adds. Mutate the mechanism: revert it in a\n" +
		"> scratch tree, or flip the single value it turns on — but before the tests run, confirm the edit\n" +
		"> landed: the target changed where you intended, not nowhere and not somewhere else. An edit that\n" +
		"> never applied is a refusal to redo with a working mechanism; it never buys a test. Then confirm\n" +
		"> an existing test fails, and restore.\n" +
		"> `mutate-and-verify.sh` mechanizes backup, apply, run, report and restore\n" +
		"> for a mutation expressed as a patch file against one or more test harnesses; which mechanism to\n" +
		"> mutate is your judgment, not the script's. Each mutation alters one mechanism — where a single\n" +
		"> revert would also change state a second check reads, split it into surgical mutations, one per\n" +
		"> mechanism. A mutation no test catches is an uncaught mutation: add the test that catches it before\n" +
		"> your turn ends. Record one `fix-mutation:` line per behaviour in your report, plus a\n" +
		"> `fix-mutations-total:` count, in exactly this shape: `fix-mutation: <path> — <what was mutated> —\n" +
		"> <the test that failed>`; `fix-mutation: <path> — none — <reason>` for a behaviour you exempt;\n" +
		"> and one `fix-mutations-total: <n>` line after them. Where\n" +
		"> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather\n" +
		"> than deciding it yourself."
	dpMutationBlockNoEditLanded = "> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before you\n" +
		"> end your turn — not only the test cases this round adds. Mutate the mechanism: revert it in a\n" +
		"> scratch tree, or flip the single value it turns on — but before the tests run, the edit is\n" +
		"> asserted: the target changed where you intended, not nowhere and not somewhere else. An edit that\n" +
		"> never applied is a refusal, not a surviving mutant — redo it with a working mechanism; it never\n" +
		"> buys a test. Then confirm an existing test fails, and restore.\n" +
		"> `mutate-and-verify.sh` mechanizes backup, apply, run, report and restore\n" +
		"> for a mutation expressed as a patch file against one or more test harnesses; which mechanism to\n" +
		"> mutate is your judgment, not the script's. Each mutation alters one mechanism — where a single\n" +
		"> revert would also change state a second check reads, split it into surgical mutations, one per\n" +
		"> mechanism. A mutation no test catches is a surviving mutant: add the test that catches it before\n" +
		"> your turn ends. Record one `fix-mutation:` line per behaviour in your report, plus a\n" +
		"> `fix-mutations-total:` count, in exactly this shape: `fix-mutation: <path> — <what was mutated> —\n" +
		"> <the test that failed>`; `fix-mutation: <path> — none — <reason>` for a behaviour you exempt;\n" +
		"> and one `fix-mutations-total: <n>` line after them. Where\n" +
		"> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather\n" +
		"> than deciding it yourself."
	dpMutationBlockNoRefusal = "> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before you\n" +
		"> end your turn — not only the test cases this round adds. Mutate the mechanism: revert it in a\n" +
		"> scratch tree, or flip the single value it turns on — but before the tests run, confirm the edit\n" +
		"> landed: the target changed where you intended, not nowhere and not somewhere else. An edit that\n" +
		"> never applied is an unapplied edit — redo it with a working mechanism; it never buys a test.\n" +
		"> Then confirm an existing test fails, and restore.\n" +
		"> `mutate-and-verify.sh` mechanizes backup, apply, run, report and restore\n" +
		"> for a mutation expressed as a patch file against one or more test harnesses; which mechanism to\n" +
		"> mutate is your judgment, not the script's. Each mutation alters one mechanism — where a single\n" +
		"> revert would also change state a second check reads, split it into surgical mutations, one per\n" +
		"> mechanism. A mutation no test catches is a surviving mutant: add the test that catches it before\n" +
		"> your turn ends. Record one `fix-mutation:` line per behaviour in your report, plus a\n" +
		"> `fix-mutations-total:` count, in exactly this shape: `fix-mutation: <path> — <what was mutated> —\n" +
		"> <the test that failed>`; `fix-mutation: <path> — none — <reason>` for a behaviour you exempt;\n" +
		"> and one `fix-mutations-total: <n>` line after them. Where\n" +
		"> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather\n" +
		"> than deciding it yourself."
	dpMutationBlockNoNeverBuys = "> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before you\n" +
		"> end your turn — not only the test cases this round adds. Mutate the mechanism: revert it in a\n" +
		"> scratch tree, or flip the single value it turns on — but before the tests run, confirm the edit\n" +
		"> landed: the target changed where you intended, not nowhere and not somewhere else. An edit that\n" +
		"> never applied is a refusal, not a surviving mutant — redo it with a working mechanism before\n" +
		"> moving on. Then confirm an existing test fails, and restore.\n" +
		"> `mutate-and-verify.sh` mechanizes backup, apply, run, report and restore\n" +
		"> for a mutation expressed as a patch file against one or more test harnesses; which mechanism to\n" +
		"> mutate is your judgment, not the script's. Each mutation alters one mechanism — where a single\n" +
		"> revert would also change state a second check reads, split it into surgical mutations, one per\n" +
		"> mechanism. A mutation no test catches is a surviving mutant: add the test that catches it before\n" +
		"> your turn ends. Record one `fix-mutation:` line per behaviour in your report, plus a\n" +
		"> `fix-mutations-total:` count, in exactly this shape: `fix-mutation: <path> — <what was mutated> —\n" +
		"> <the test that failed>`; `fix-mutation: <path> — none — <reason>` for a behaviour you exempt;\n" +
		"> and one `fix-mutations-total: <n>` line after them. Where\n" +
		"> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather\n" +
		"> than deciding it yourself."
	// dpMutationBlockNoShape is the pre-KAN-854 wording: it pointed the fix
	// subagent at "the review-panel contract's fenced block", which its
	// dispatch never carries, instead of stating the fix-mutation: shape.
	dpMutationBlockNoShape = "> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before you\n" +
		"> end your turn — not only the test cases this round adds. Mutate the mechanism: revert it in a\n" +
		"> scratch tree, or flip the single value it turns on — but before the tests run, confirm the edit\n" +
		"> landed: the target changed where you intended, not nowhere and not somewhere else. An edit that\n" +
		"> never applied is a refusal, not a surviving mutant — redo it with a working mechanism; it never\n" +
		"> buys a test. Then confirm an existing test fails, and restore.\n" +
		"> A mutation no test catches is a surviving mutant: add the test that catches it before\n" +
		"> your turn ends. Record one `fix-mutation:` line per behaviour in your report, plus a\n" +
		"> `fix-mutations-total:` count, in the shape the review-panel contract's fenced block gives."
	dpToolsBlock = "> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,\n" +
		"> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,\n" +
		"> before anything else. Never ToolSearch for a tool already listed, and never a wildcard query: a\n" +
		"> schema loaded later changes your tool list and re-prices your whole context at full input rate."
	dpToolsBlockNoFirstTurn = "> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,\n" +
		"> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch before your first\n" +
		"> tool call, before anything else. Never ToolSearch for a tool already listed, and never a\n" +
		"> wildcard query: a schema loaded later changes your tool list and re-prices your whole context at\n" +
		"> full input rate."
	dpToolsBlockNoWildcard = "> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,\n" +
		"> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,\n" +
		"> before anything else. Never ToolSearch for a tool already listed, and never a broad query: a\n" +
		"> schema loaded later changes your tool list and re-prices your whole context at full input rate."
	dpToolsBlockNoReprices = "> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,\n" +
		"> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,\n" +
		"> before anything else. Never ToolSearch for a tool already listed, and never a wildcard query: a\n" +
		"> schema loaded later changes your tool list and costs your whole context again at full input\n" +
		"> rate."
	dpHandshakeBlock = "> **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own\n" +
		"> system prompt>` and nothing else on that line. Answer it before any tool call."
	dpHandshakeBlockNoFirstLine = "> **MODEL HANDSHAKE:** the very first line you send is `Model: <the model named in your own\n" +
		"> system prompt>` and nothing else on that line. Answer it before any tool call."
	dpHandshakeBlockNoNothingElse = "> **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own\n" +
		"> system prompt>` and only that on the line. Answer it before any tool call."
	dpHandshakeBlockNoBeforeToolCall = "> **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own\n" +
		"> system prompt>` and nothing else on that line. Answer it before making any tool call."
	dpIndependentBlock = "> **INDEPENDENT PASSES:** each pass starts from `final-review.diff` and the code, never from an\n" +
		"> earlier pass's report or conclusions. Do not cite, defer to, or skip a defect because an earlier\n" +
		"> pass raised it — if it sits in this pass's angle, raise it again under this pass. Write each\n" +
		"> pass's report file before beginning the next pass."
	dpDelegationBlock = "> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a\n" +
		"> subagent, background agent or helper of any kind — you are the leaf of this run, and any child\n" +
		"> you start is unrecorded and outside the conductor's closed list (**Dispatch sites — the\n" +
		"> conductor's closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and\n" +
		"> fixing are your own Read, Bash and Edit calls."
	dpDelegationBlockNoNeverCallAgent = "> **NO DELEGATION:** Do this work yourself. Do not use the `Agent` tool, and never spawn a\n" +
		"> subagent, background agent or helper of any kind — you are the leaf of this run, and any child\n" +
		"> you start is unrecorded and outside the conductor's closed list (**Dispatch sites — the\n" +
		"> conductor's closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and\n" +
		"> fixing are your own Read, Bash and Edit calls."
	dpDelegationBlockNoNeverSpawn = "> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and start no subagent,\n" +
		"> background agent or helper of any kind — you are the leaf of this run, and any child\n" +
		"> you start is unrecorded and outside the conductor's closed list (**Dispatch sites — the\n" +
		"> conductor's closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and\n" +
		"> fixing are your own Read, Bash and Edit calls."
	dpDelegationBlockNoLeaf = "> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a\n" +
		"> subagent, background agent or helper of any kind — you are the last agent in this chain, and any child\n" +
		"> you start is unrecorded and outside the conductor's closed list (**Dispatch sites — the\n" +
		"> conductor's closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and\n" +
		"> fixing are your own Read, Bash and Edit calls."
	dpPixelProbeBlock = "> **PIXEL PROBE:** every fix you land to draw/geometry code — code that computes what is\n" +
		"> drawn: extents, bounds, gridlines, offsets, paths, positions, sizes — carries a probe assertion\n" +
		"> against the actual rendered pixels or geometry, in the style the reviewers' reproducers\n" +
		"> already use: drive the real draw path and assert on the value it renders or computes, never on\n" +
		"> a hand-derived copy of it. A fix commit for this class of bug that lands no such probe is\n" +
		"> rejected at the fix step — the finding stays open and the round does not close — so the\n" +
		"> regression is caught this round, not discovered by the next one. Where pixels cannot be\n" +
		"> rendered in this environment, the probe asserts on the geometry the draw path actually\n" +
		"> computes — the same numbers the renderer will paint — and your report names why the pixel\n" +
		"> output itself was not asserted."
	dpPixelProbeBlockNoClass = "> **PIXEL PROBE:** every fix you land to draw-and-geometry code — code that computes what is\n" +
		"> drawn: extents, bounds, gridlines, offsets, paths, positions, sizes — carries a probe assertion\n" +
		"> against the actual rendered pixels or geometry, in the style the reviewers' reproducers\n" +
		"> already use: drive the real draw path and assert on the value it renders or computes, never on\n" +
		"> a hand-derived copy of it. A fix commit for this class of bug that lands no such probe is\n" +
		"> rejected at the fix step — the finding stays open and the round does not close — so the\n" +
		"> regression is caught this round, not discovered by the next one. Where pixels cannot be\n" +
		"> rendered in this environment, the probe asserts on the geometry the draw path actually\n" +
		"> computes — the same numbers the renderer will paint — and your report names why the pixel\n" +
		"> output itself was not asserted."
	dpPixelProbeBlockNoPixels = "> **PIXEL PROBE:** every fix you land to draw/geometry code — code that computes what is\n" +
		"> drawn: extents, bounds, gridlines, offsets, paths, positions, sizes — carries a probe assertion\n" +
		"> against the actual rendered output, in the style the reviewers' reproducers\n" +
		"> already use: drive the real draw path and assert on the value it renders or computes, never on\n" +
		"> a hand-derived copy of it. A fix commit for this class of bug that lands no such probe is\n" +
		"> rejected at the fix step — the finding stays open and the round does not close — so the\n" +
		"> regression is caught this round, not discovered by the next one. Where pixels cannot be\n" +
		"> rendered in this environment, the probe asserts on the geometry the draw path actually\n" +
		"> computes — the same numbers the renderer will paint — and your report names why the pixel\n" +
		"> output itself was not asserted."
	dpPixelProbeBlockNoRejection = "> **PIXEL PROBE:** every fix you land to draw/geometry code — code that computes what is\n" +
		"> drawn: extents, bounds, gridlines, offsets, paths, positions, sizes — carries a probe assertion\n" +
		"> against the actual rendered pixels or geometry, in the style the reviewers' reproducers\n" +
		"> already use: drive the real draw path and assert on the value it renders or computes, never on\n" +
		"> a hand-derived copy of it. A fix commit for this class of bug that lands no such probe is\n" +
		"> refused before the round closes — the finding stays open — so the\n" +
		"> regression is caught this round, not discovered by the next one. Where pixels cannot be\n" +
		"> rendered in this environment, the probe asserts on the geometry the draw path actually\n" +
		"> computes — the same numbers the renderer will paint — and your report names why the pixel\n" +
		"> output itself was not asserted."
	dpGuardBitesBlock = "> **PROVE THE GUARD BITES:** When this task's tests assert on configuration or file content — a\n" +
		"> guard script, an embedded config, a fixture file — a passing run alone is not evidence: break\n" +
		"> the property the assertion protects, run the guard or the test against the broken state and\n" +
		"> capture its failure, then restore and capture the pass. Report both runs. \"The pattern is now\n" +
		"> stricter\" is intent, not evidence — the failing run against the broken state is the evidence."
	dpGuardBitesBlockNoScope = "> **PROVE THE GUARD BITES:** When this task's tests check configuration or file contents — a\n" +
		"> guard script, an embedded config, a fixture file — a passing run alone is not evidence: break\n" +
		"> the property the assertion protects, run the guard or the test against the broken state and\n" +
		"> capture its failure, then restore and capture the pass. Report both runs. \"The pattern is now\n" +
		"> stricter\" is intent, not evidence — the failing run against the broken state is the evidence."
	dpGuardBitesBlockNoBreak = "> **PROVE THE GUARD BITES:** When this task's tests assert on configuration or file content — a\n" +
		"> guard script, an embedded config, a fixture file — a passing run alone is not evidence: tighten\n" +
		"> the property the assertion protects, run the guard or the test against the broken state and\n" +
		"> capture its failure, then restore and capture the pass. Report both runs. \"The pattern is now\n" +
		"> stricter\" is intent, not evidence — the failing run against the broken state is the evidence."
	dpGuardBitesBlockNoBrokenState = "> **PROVE THE GUARD BITES:** When this task's tests assert on configuration or file content — a\n" +
		"> guard script, an embedded config, a fixture file — a passing run alone is not evidence: break\n" +
		"> the property the assertion protects, run the guard or the test against the mutated input and\n" +
		"> capture its failure, then restore and capture the pass. Report both runs. \"The pattern is now\n" +
		"> stricter\" is intent, not evidence — the failing run against the mutated input is the evidence."
	dpDecideBlock = "> **REPORT, DON'T DECIDE:** A question only the operator can settle — a spec point no recorded\n" +
		"> decision covers, a recorded decision the plan appears to contradict, a scope the plan does not\n" +
		"> name — is reported, never decided in code. Where the plan can proceed, implement it as written\n" +
		"> and carry the question in your REPORT FILE; where it cannot proceed without the answer, report\n" +
		"> BLOCKED. A decision you make silently is a decision nobody recorded."
	dpDecideBlockNoOperator = "> **REPORT, DON'T DECIDE:** A question the implementer cannot answer alone — a spec point no recorded\n" +
		"> decision covers, a recorded decision the plan appears to contradict, a scope the plan does not\n" +
		"> name — is reported, never decided in code. Where the plan can proceed, implement it as written\n" +
		"> and carry the question in your REPORT FILE; where it cannot proceed without the answer, report\n" +
		"> BLOCKED. A decision you make silently is a decision nobody recorded."
	dpDecideBlockNoReported = "> **REPORT, DON'T DECIDE:** A question only the operator can settle — a spec point no recorded\n" +
		"> decision covers, a recorded decision the plan appears to contradict, a scope the plan does not\n" +
		"> name — is raised with the operator, never settled in code. Where the plan can proceed,\n" +
		"> implement it as written and carry the question in your REPORT FILE; where it cannot proceed\n" +
		"> without the answer, report BLOCKED. A decision you make silently is a decision nobody recorded."
	dpDecideBlockNoCarry = "> **REPORT, DON'T DECIDE:** A question only the operator can settle — a spec point no recorded\n" +
		"> decision covers, a recorded decision the plan appears to contradict, a scope the plan does not\n" +
		"> name — is reported, never decided in code. Where the plan can proceed, implement it as written\n" +
		"> and record the question in your report; where it cannot proceed without the answer, report\n" +
		"> BLOCKED. A decision you make silently is a decision nobody recorded."
	dpDecideBlockNoBlocked = "> **REPORT, DON'T DECIDE:** A question only the operator can settle — a spec point no recorded\n" +
		"> decision covers, a recorded decision the plan appears to contradict, a scope the plan does not\n" +
		"> name — is reported, never decided in code. Where the plan can proceed, implement it as written\n" +
		"> and carry the question in your REPORT FILE; where it cannot proceed without the answer, hand the\n" +
		"> plan back as blocked. A decision you make silently is a decision nobody recorded."
	dpEntryContextBlock = "> **ENTRY CONTEXT:** `final-review.diff` is your entry artifact, and the touched-files list in\n" +
		"> this prompt is your named entry context — begin from those two, not from a whole-tree\n" +
		"> exploration. The list names every path this branch touched, per worktree, with the named\n" +
		"> contracts among them; read a listed file only as far as the diff and a finding require, and step\n" +
		"> outside the list only when a finding cannot otherwise be established. Name the coverage trade in\n" +
		"> your report: which touched files you read in full, which you read only at the diff's hunks, and\n" +
		"> which you deliberately did not read."
	dpEntryContextBlockNoNamed = "> **ENTRY CONTEXT:** `final-review.diff` is your entry artifact, and the touched-files list in\n" +
		"> this prompt is your starting scope — begin from those two, not from a whole-tree\n" +
		"> exploration. The list names every path this branch touched, per worktree, with the named\n" +
		"> contracts among them; read a listed file only as far as the diff and a finding require, and step\n" +
		"> outside the list only when a finding cannot otherwise be established. Name the coverage trade in\n" +
		"> your report: which touched files you read in full, which you read only at the diff's hunks, and\n" +
		"> which you deliberately did not read."
	dpEntryContextBlockNoWholeTree = "> **ENTRY CONTEXT:** `final-review.diff` is your entry artifact, and the touched-files list in\n" +
		"> this prompt is your named entry context — begin from those two rather than wandering the tree.\n" +
		"> The list names every path this branch touched, per worktree, with the named\n" +
		"> contracts among them; read a listed file only as far as the diff and a finding require, and step\n" +
		"> outside the list only when a finding cannot otherwise be established. Name the coverage trade in\n" +
		"> your report: which touched files you read in full, which you read only at the diff's hunks, and\n" +
		"> which you deliberately did not read."
	dpEntryContextBlockNoCoverage = "> **ENTRY CONTEXT:** `final-review.diff` is your entry artifact, and the touched-files list in\n" +
		"> this prompt is your named entry context — begin from those two, not from a whole-tree\n" +
		"> exploration. The list names every path this branch touched, per worktree, with the named\n" +
		"> contracts among them; read a listed file only as far as the diff and a finding require, and step\n" +
		"> outside the list only when a finding cannot otherwise be established. State the read-coverage\n" +
		"> call in your report: which touched files you read in full, which you read only at the diff's\n" +
		"> hunks, and which you deliberately did not read."
	dpFindingsInputBlock = "> **FINDINGS ARE INPUT:** a finding names a defect and suggests a route to fixing it — findings\n" +
		"> are input, not orders. Your obligation is to resolve the defect the finding names; the\n" +
		"> suggested route is the raising slot's proposal, never a binding instruction. Where the literal\n" +
		"> route would break something the code already requires — a required ordering, a declared\n" +
		"> contract, a stated invariant — take the route that resolves the defect without the damage, and\n" +
		"> your report records the deviation and justifies it: what the literal route would have broken,\n" +
		"> and how yours resolves the defect. A silent deviation is an unfixed finding, and so is a\n" +
		"> literal compliance that leaves the defect standing."
	dpFindingsInputBlockNoInput = "> **FINDINGS ARE INPUT:** a finding names a defect and suggests a route to fixing it — the\n" +
		"> finding is the fix's brief. Your obligation is to resolve the defect the finding names; the\n" +
		"> suggested route is the raising slot's proposal, never a binding instruction. Where the literal\n" +
		"> route would break something the code already requires — a required ordering, a declared\n" +
		"> contract, a stated invariant — take the route that resolves the defect without the damage, and\n" +
		"> your report records the deviation and justifies it: what the literal route would have broken,\n" +
		"> and how yours resolves the defect. A silent deviation is an unfixed finding, and so is a\n" +
		"> literal compliance that leaves the defect standing."
	dpFindingsInputBlockNoResolve = "> **FINDINGS ARE INPUT:** a finding names a defect and suggests a route to fixing it — findings\n" +
		"> are input, not orders. Your obligation is to fix what the finding points at; the\n" +
		"> suggested route is the raising slot's proposal, never a binding instruction. Where the literal\n" +
		"> route would break something the code already requires — a required ordering, a declared\n" +
		"> contract, a stated invariant — take the route that resolves the defect without the damage, and\n" +
		"> your report records the deviation and justifies it: what the literal route would have broken,\n" +
		"> and how yours resolves the defect. A silent deviation is an unfixed finding, and so is a\n" +
		"> literal compliance that leaves the defect standing."
	dpFindingsInputBlockNoJustify = "> **FINDINGS ARE INPUT:** a finding names a defect and suggests a route to fixing it — findings\n" +
		"> are input, not orders. Your obligation is to resolve the defect the finding names; the\n" +
		"> suggested route is the raising slot's proposal, never a binding instruction. Where the literal\n" +
		"> route would break something the code already requires — a required ordering, a declared\n" +
		"> contract, a stated invariant — take the route that resolves the defect without the damage, and\n" +
		"> your report notes the detour: what the literal route would have broken,\n" +
		"> and how yours resolves the defect. A silent deviation is an unfixed finding, and so is a\n" +
		"> literal compliance that leaves the defect standing."
	dpFindingsInputBlockNoSilent = "> **FINDINGS ARE INPUT:** a finding names a defect and suggests a route to fixing it — findings\n" +
		"> are input, not orders. Your obligation is to resolve the defect the finding names; the\n" +
		"> suggested route is the raising slot's proposal, never a binding instruction. Where the literal\n" +
		"> route would break something the code already requires — a required ordering, a declared\n" +
		"> contract, a stated invariant — take the route that resolves the defect without the damage, and\n" +
		"> your report records the deviation and justifies it: what the literal route would have broken,\n" +
		"> and how yours resolves the defect. An undeclared detour helps nobody, and so does a\n" +
		"> literal compliance that leaves the defect standing."
	dpContextBundleFailureBlock = "> **CONTEXT BUNDLE FAILURE:** the gather above exited non-zero, or the bundle file is still\n" +
		"> absent when the `test -f` on the rebuild's output path above says so — a failed\n" +
		"> build. Record the cause with `flow record pass -round <round> -note 'context bundle: build\n" +
		"> failed — <the script's stderr>'`, then resolve it, never asked, per **Auto-resolution**\n" +
		"> (`skills/flow-contracts/operator-prompts.md`):\n" +
		"> - **Stop — resolve the build, then re-run the panel** *(recommended; silence defaults here)* —\n" +
		">   closes `flow.review-panel` with `-outcome stopped`\n" +
		"> - **Continue — dispatch every slot without the bundle** — recorded with `flow record pass\n" +
		">   -round <round> -note 'context bundle: dispatched without it — operator override'`, every\n" +
		">   slot's prompt dropping the CONTEXT BUNDLE paragraph and naming the bundle's absence in its\n" +
		">   place, so the reviewers' reading the plan and decision directly is on the record, never a\n" +
		">   silent reduction of their context"
	dpContextBundleFailureBlockNoNonzero = "> **CONTEXT BUNDLE FAILURE:** the gather above failed outright, or the bundle file is\n" +
		"> absent when the `test -f` on the rebuild's output path above says so — a failed\n" +
		"> build. Record the cause with `flow record pass -round <round> -note 'context bundle: build\n" +
		"> failed — <the script's stderr>'`, then resolve it, never asked, per **Auto-resolution**\n" +
		"> (`skills/flow-contracts/operator-prompts.md`):\n" +
		"> - **Stop — resolve the build, then re-run the panel** *(recommended; silence defaults here)* —\n" +
		">   closes `flow.review-panel` with `-outcome stopped`\n" +
		"> - **Continue — dispatch every slot without the bundle** — recorded with `flow record pass\n" +
		">   -round <round> -note 'context bundle: dispatched without it — operator override'`, every\n" +
		">   slot's prompt dropping the CONTEXT BUNDLE paragraph and naming the bundle's absence in its\n" +
		">   place, so the reviewers' reading the plan and decision directly is on the record, never a\n" +
		">   silent reduction of their context"
	dpContextBundleFailureBlockNoBuildFailed = "> **CONTEXT BUNDLE FAILURE:** the gather above exited non-zero, or the bundle file is still\n" +
		"> absent when the `test -f` on the rebuild's output path above says so — a failed\n" +
		"> build. Record the cause with `flow record pass -round <round> -note 'context bundle: gather\n" +
		"> error — <the script's stderr>'`, then resolve it, never asked, per **Auto-resolution**\n" +
		"> (`skills/flow-contracts/operator-prompts.md`):\n" +
		"> - **Stop — resolve the build, then re-run the panel** *(recommended; silence defaults here)* —\n" +
		">   closes `flow.review-panel` with `-outcome stopped`\n" +
		"> - **Continue — dispatch every slot without the bundle** — recorded with `flow record pass\n" +
		">   -round <round> -note 'context bundle: dispatched without it — operator override'`, every\n" +
		">   slot's prompt dropping the CONTEXT BUNDLE paragraph and naming the bundle's absence in its\n" +
		">   place, so the reviewers' reading the plan and decision directly is on the record, never a\n" +
		">   silent reduction of their context"
	dpContextBundleFailureBlockNoDispatch = "> **CONTEXT BUNDLE FAILURE:** the gather above exited non-zero, or the bundle file is still\n" +
		"> absent when the `test -f` on the rebuild's output path above says so — a failed\n" +
		"> build. Record the cause with `flow record pass -round <round> -note 'context bundle: build\n" +
		"> failed — <the script's stderr>'`, then resolve it, never asked, per **Auto-resolution**\n" +
		"> (`skills/flow-contracts/operator-prompts.md`):\n" +
		"> - **Stop — resolve the build, then re-run the panel** *(recommended; silence defaults here)* —\n" +
		">   closes `flow.review-panel` with `-outcome stopped`\n" +
		"> - **Continue — run the panel regardless** — recorded with `flow record pass\n" +
		">   -round <round> -note 'context bundle: dispatched without it — operator override'`, every\n" +
		">   slot's prompt dropping the CONTEXT BUNDLE paragraph and naming the bundle's absence in its\n" +
		">   place, so the reviewers' reading the plan and decision directly is on the record, never a\n" +
		">   silent reduction of their context"
	dpContextBundleFailureBlockNoOverride = "> **CONTEXT BUNDLE FAILURE:** the gather above exited non-zero, or the bundle file is still\n" +
		"> absent when the `test -f` on the rebuild's output path above says so — a failed\n" +
		"> build. Record the cause with `flow record pass -round <round> -note 'context bundle: build\n" +
		"> failed — <the script's stderr>'`, then resolve it, never asked, per **Auto-resolution**\n" +
		"> (`skills/flow-contracts/operator-prompts.md`):\n" +
		"> - **Stop — resolve the build, then re-run the panel** *(recommended; silence defaults here)* —\n" +
		">   closes `flow.review-panel` with `-outcome stopped`\n" +
		"> - **Continue — dispatch every slot without the bundle** — recorded with `flow record pass\n" +
		">   -round <round> -note 'context bundle: skipped — operator said go'`, every\n" +
		">   slot's prompt dropping the CONTEXT BUNDLE paragraph and naming the bundle's absence in its\n" +
		">   place, so the reviewers' reading the plan and decision directly is on the record, never a\n" +
		">   silent reduction of their context"
	dpContextBundleFailureBlockNoStopped = "> **CONTEXT BUNDLE FAILURE:** the gather above exited non-zero, or the bundle file is still\n" +
		"> absent when the `test -f` on the rebuild's output path above says so — a failed\n" +
		"> build. Record the cause with `flow record pass -round <round> -note 'context bundle: build\n" +
		"> failed — <the script's stderr>'`, then resolve it, never asked, per **Auto-resolution**\n" +
		"> (`skills/flow-contracts/operator-prompts.md`):\n" +
		"> - **Stop — resolve the build, then re-run the panel** *(recommended; silence defaults here)* —\n" +
		">   closes `flow.review-panel` as stopped\n" +
		"> - **Continue — dispatch every slot without the bundle** — recorded with `flow record pass\n" +
		">   -round <round> -note 'context bundle: dispatched without it — operator override'`, every\n" +
		">   slot's prompt dropping the CONTEXT BUNDLE paragraph and naming the bundle's absence in its\n" +
		">   place, so the reviewers' reading the plan and decision directly is on the record, never a\n" +
		">   silent reduction of their context"
	dpOutputBudgetBlock = "> **OUTPUT BUDGET:** Every tool result stays in your context, and every later turn re-reads your\n" +
		"> whole context — a large output is paid for again on every turn after it. Read a file over 200\n" +
		"> lines by line range — `grep -n` for the symbol, then `sed -n '<a>,<b>p'` or Read with\n" +
		"> `offset`/`limit` — and never re-read a file already in your context unless you have edited it\n" +
		"> since. Cap every search (`| head -40`) and every build, lint or install run (`| tail -30`), and\n" +
		"> reproduce a failing block from its log rather than printing the whole log. Never print a\n" +
		"> generated file — a lockfile, a snapshot, a bundle, a build artifact."
	dpOutputBudgetBlockNoRange = "> **OUTPUT BUDGET:** Every tool result stays in your context, and every later turn re-reads your\n" +
		"> whole context — a large output is paid for again on every turn after it. Read a file over 200\n" +
		"> lines in pieces — `grep -n` for the symbol, then `sed -n '<a>,<b>p'` or Read with\n" +
		"> `offset`/`limit` — and never re-read a file already in your context unless you have edited it\n" +
		"> since. Cap every search (`| head -40`) and every build, lint or install run (`| tail -30`), and\n" +
		"> reproduce a failing block from its log rather than printing the whole log. Never print a\n" +
		"> generated file — a lockfile, a snapshot, a bundle, a build artifact."
	dpReadonlyBlock = "> **READ-ONLY REVIEW:** You review a tree other agents are working in — never mutate it. No\n" +
		"> command that writes: no `git checkout`, `git restore`, `git reset`, `git stash`, `git clean`,\n" +
		"> no commit, no index change, and no file edit outside your own report file — the panel's\n" +
		"> mutating slots are the one declared exception, and they work in throwaway copies, not this\n" +
		"> tree. Inspect with Read, Grep, and the read-only git forms — `git show`, `git diff`,\n" +
		"> `git log`, `git status`. A mutation you cause is indistinguishable from a defect the next\n" +
		"> implementer inherits, and a restore you perform is a claim nobody can check: gymie KAN-635's\n" +
		"> reviewer ran `git checkout <sha> -- .` mid-review, destroyed uncommitted planning artifacts,\n" +
		"> and reported the tree restored."
	dpReadonlyBlockNoNeverMutate = "> **READ-ONLY REVIEW:** You review a tree other agents are working in — do not rewrite it. No\n" +
		"> command that writes: no `git checkout`, `git restore`, `git reset`, `git stash`, `git clean`,\n" +
		"> no commit, no index change, and no file edit outside your own report file — the panel's\n" +
		"> mutating slots are the one declared exception, and they work in throwaway copies, not this\n" +
		"> tree. Inspect with Read, Grep, and the read-only git forms — `git show`, `git diff`,\n" +
		"> `git log`, `git status`. A mutation you cause is indistinguishable from a defect the next\n" +
		"> implementer inherits, and a restore you perform is a claim nobody can check: gymie KAN-635's\n" +
		"> reviewer ran `git checkout <sha> -- .` mid-review, destroyed uncommitted planning artifacts,\n" +
		"> and reported the tree restored."
	dpReadonlyBlockNoFileEdit = "> **READ-ONLY REVIEW:** You review a tree other agents are working in — never mutate it. No\n" +
		"> command that writes: no `git checkout`, `git restore`, `git reset`, `git stash`, `git clean`,\n" +
		"> no commit, no index change, and no edit of any file but your own report — the panel's\n" +
		"> mutating slots are the one declared exception, and they work in throwaway copies, not this\n" +
		"> tree. Inspect with Read, Grep, and the read-only git forms — `git show`, `git diff`,\n" +
		"> `git log`, `git status`. A mutation you cause is indistinguishable from a defect the next\n" +
		"> implementer inherits, and a restore you perform is a claim nobody can check: gymie KAN-635's\n" +
		"> reviewer ran `git checkout <sha> -- .` mid-review, destroyed uncommitted planning artifacts,\n" +
		"> and reported the tree restored."
	dpReadonlyBlockNoClaim = "> **READ-ONLY REVIEW:** You review a tree other agents are working in — never mutate it. No\n" +
		"> command that writes: no `git checkout`, `git restore`, `git reset`, `git stash`, `git clean`,\n" +
		"> no commit, no index change, and no file edit outside your own report file — the panel's\n" +
		"> mutating slots are the one declared exception, and they work in throwaway copies, not this\n" +
		"> tree. Inspect with Read, Grep, and the read-only git forms — `git show`, `git diff`,\n" +
		"> `git log`, `git status`. A mutation you cause is indistinguishable from a defect the next\n" +
		"> implementer inherits, and a restore you perform is an assertion nothing verifies: gymie\n" +
		"> KAN-635's reviewer ran `git checkout <sha> -- .` mid-review, destroyed uncommitted planning\n" +
		"> artifacts, and reported the tree restored."
)

// CLEAN_REVIEW_PANEL, CLEAN_IMPLEMENT_NO_READONLY and CLEAN_IMPLEMENT: case
// 1's fully-correct site files, the baseline later cases vary one deficiency
// against.
var dpCleanReviewPanel = dpDoc(dpReviewerBlock, dpVerbatimBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpMutationBlock, dpPixelProbeBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpIndependentBlock, dpDelegationBlock, dpDelegationBlock, dpEntryContextBlock, dpFindingsInputBlock, dpContextBundleFailureBlock, dpOutputBudgetBlock)
var dpCleanImplementNoReadonly = dpDoc(dpReviewerBlock, dpImplementerBlock, dpGuardBitesBlock, dpDecideBlock, dpReviewerBlock, dpForegroundBlock, dpForegroundBlock, dpForegroundBlock, dpTargetedBlock, dpToolsBlock, dpToolsBlock, dpHandshakeBlock, dpHandshakeBlock, dpDelegationBlock, dpDelegationBlock)
var dpCleanImplement = dpDoc(dpCleanImplementNoReadonly, dpReadonlyBlock, dpOutputBudgetBlock)

// dpCleanVisualVerify is one visual-verify site file at its one dispatch
// site -- the verifier in visual-verify.md, the tooling analyst in
// visual-verify-tooling-analysis.md -- carrying its three blocks.
var dpCleanVisualVerify = dpDoc(dpToolsBlock, dpHandshakeBlock, dpDelegationBlock)
