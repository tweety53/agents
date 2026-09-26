package guard

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// TestGatherDispatchContext ports every case of the retired
// scripts/test-gather-dispatch-context.sh, one subtest per `ok:` label of
// that harness's green run at 0747740. The three cases that drove the bash
// helper's hash-tool fallbacks (32, 33 and 35) are substituted -- see each
// row's comment -- because the Go port hashes with crypto/sha256 and has no
// hash tool to be missing.
//
// Each subtest builds its own repository by copying one template built by
// the parent, runs the guard in-process through Registry, and reads back
// the bundle it wrote. The guard sees a PATH holding only the case's own bin
// directory, so `flow` is absent unless a case stubs it: the real store is
// never consulted.
func TestGatherDispatchContext(t *testing.T) {
	t.Parallel()
	tmpl := gdcTemplate(t)

	type check func(t *testing.T, f *gdcFx, r gdcResult)
	type scen func(t *testing.T, f *gdcFx) gdcResult

	exit := func(want int) check {
		return func(t *testing.T, _ *gdcFx, r gdcResult) {
			if r.rc != want {
				t.Fatalf("exit %d, want %d\n%s", r.rc, want, r.out)
			}
		}
	}
	has := func(subs ...string) check {
		return func(t *testing.T, _ *gdcFx, r gdcResult) {
			for _, s := range subs {
				if !strings.Contains(r.out, s) {
					t.Fatalf("output lacks %q\n%s", s, r.out)
				}
			}
		}
	}
	lacks := func(subs ...string) check {
		return func(t *testing.T, _ *gdcFx, r gdcResult) {
			for _, s := range subs {
				if strings.Contains(r.out, s) {
					t.Fatalf("output carries %q\n%s", s, r.out)
				}
			}
		}
	}
	line := func(l string) check {
		return func(t *testing.T, _ *gdcFx, r gdcResult) {
			if !gdcHasLine(r.out, l) {
				t.Fatalf("output has no line %q\n%s", l, r.out)
			}
		}
	}
	all := func(cs ...check) check {
		return func(t *testing.T, f *gdcFx, r gdcResult) {
			for _, c := range cs {
				c(t, f, r)
			}
		}
	}
	errHas := func(s string) check {
		return func(t *testing.T, _ *gdcFx, r gdcResult) {
			if !strings.Contains(r.err, s) {
				t.Fatalf("stderr lacks %q\n%s", s, r.err)
			}
		}
	}

	base := func(t *testing.T, f *gdcFx) gdcResult { return f.run() }
	// rm removes change-root leaves; "" names the principles file.
	rm := func(leaves ...string) scen {
		return func(t *testing.T, f *gdcFx) gdcResult {
			for _, l := range leaves {
				p := f.principles
				if l != "" {
					p = f.changeRoot + "/" + l
				}
				gdcRemove(t, p)
			}
			return f.run()
		}
	}
	allSources := []string{"proposal.md", "design.md", "tasks.md", ""}
	withProject := func(body string) scen {
		return func(t *testing.T, f *gdcFx) gdcResult {
			writeFile(t, f.repo+"/.flow/project.md", body)
			return f.run()
		}
	}
	withPlan := func(plan string, ids string) scen {
		return func(t *testing.T, f *gdcFx) gdcResult {
			writeFile(t, f.changeRoot+"/tasks.md", plan)
			if ids == "" {
				return f.run()
			}
			return f.run(ids)
		}
	}
	withFlow := func(stub string) scen {
		return func(t *testing.T, f *gdcFx) gdcResult {
			writeExec(t, f.bin+"/flow", stub)
			return f.run()
		}
	}
	principlesAt := func(rel string) scen {
		return func(t *testing.T, f *gdcFx) gdcResult {
			f.principles = f.repo + "/" + rel
			return f.run()
		}
	}
	// secret writes TOP-SECRET-<n> into a directory outside the repository.
	secret := func(t *testing.T, f *gdcFx, n string) string {
		dir := f.base + "/outside" + n
		writeFile(t, dir+"/secret.md", "TOP-SECRET-"+n+"\n")
		return dir
	}

	case12 := func(t *testing.T, f *gdcFx) gdcResult {
		outside := f.base + "/outside12"
		writeFile(t, outside+"/changes/demo/tasks.md", "TOP-SECRET-12\n")
		gdcSymlink(t, outside, f.repo+"/spectre-symlink")
		f.changeRoot = f.repo + "/spectre-symlink/changes/demo"
		return f.run()
	}
	case13 := func(t *testing.T, f *gdcFx) gdcResult {
		writeFile(t, f.base+"/outside13/real-principles.md", "REAL-PRINCIPLES-BODY\n")
		gdcReplaceWithSymlink(t, f.base+"/outside13/real-principles.md", f.principles)
		return f.run()
	}
	case14 := func(t *testing.T, f *gdcFx) gdcResult {
		writeFile(t, f.base+"/realskill14/engineering-principles.md", "ANCESTOR-PRINCIPLES-BODY\n")
		gdcSymlink(t, f.base+"/realskill14", f.base+"/symlinkskill14")
		f.principles = f.base + "/symlinkskill14/engineering-principles.md"
		return f.run()
	}
	case15 := func(t *testing.T, f *gdcFx) gdcResult {
		writeFile(t, f.base+"/outside14/demo/proposal.md", "PROPOSAL-BODY\n")
		f.changeRoot = f.base + "/outside14/demo"
		return f.run()
	}
	case16 := func(t *testing.T, f *gdcFx) gdcResult {
		gdcReplaceWithSymlink(t, secret(t, f, "15")+"/secret.md", f.changeRoot+"/proposal.md")
		return f.run()
	}
	case18 := func(t *testing.T, f *gdcFx) gdcResult {
		gdcReplaceWithSymlink(t, f.changeRoot+"/tasks.md", f.changeRoot+"/design.md")
		return f.run()
	}
	case19 := func(t *testing.T, f *gdcFx) gdcResult {
		f.worktreeArg = f.repo + "/"
		return f.run()
	}
	case20 := func(t *testing.T, f *gdcFx) gdcResult {
		f.worktreeArg, f.dir = ".", f.repo
		return f.run()
	}
	case21 := func(t *testing.T, f *gdcFx) gdcResult {
		gdcReplaceWithSymlink(t, f.changeRoot+"/proposal-loop-b.md", f.changeRoot+"/proposal.md")
		gdcSymlink(t, f.changeRoot+"/proposal.md", f.changeRoot+"/proposal-loop-b.md")
		return f.run()
	}
	missingOutput := func(t *testing.T, f *gdcFx) gdcResult {
		return f.runArgs(f.repo, f.changeRoot, "demo", f.principles)
	}
	named := func(name string) scen {
		return func(t *testing.T, f *gdcFx) gdcResult {
			return f.runArgs(f.repo, f.changeRoot, name, f.principles, f.output)
		}
	}
	twice := func(between func(t *testing.T, f *gdcFx)) scen {
		return func(t *testing.T, f *gdcFx) gdcResult {
			if r := f.run(); r.rc != 0 {
				t.Fatalf("first call exited %d: %s", r.rc, r.out)
			}
			between(t, f)
			return f.run()
		}
	}

	const twoIncidents = `[
  {"occurredAt":"2026-09-03T14:20:00Z","guard":"check-task-commit-fields","symptom":"Baseline revert re-entered mid-flight","recovery":"aborted revert, restored dir from stash^3","minutesLost":55,"change":"kan-423"},
  {"occurredAt":"2025-01-01T00:00:00Z","guard":"check-unfinished-work","symptom":"x","recovery":"y","minutesLost":10,"change":null}
]`
	const twoHazards = `[
  {"id":2,"name":"commit-fields-guard-reverts-head-only","body":"Never stash, revert or reset inside a worktree a guard may read; inspect a timed-out guard call first","applies":"all","active":true,"createdAt":"2026-09-09T12:00:00Z"},
  {"id":1,"name":"commit-fields-guard-single-repo","body":"pass the change own worktree, never a sibling","applies":"single-repo","active":true,"createdAt":"2026-09-09T11:00:00Z"}
]`
	// incidentStub is make_flow_stub_dir's stub; hazardStub is
	// make_flow_hazard_stub_dir's, recording its argv when argv is set.
	incidentStub := func(json string) string {
		return "#!/bin/sh\nif [ \"$1\" = record ] && [ \"$2\" = incidents ]; then\ncat <<'JSON'\n" + json +
			"\nJSON\nexit 0\nfi\necho \"flow: unexpected invocation: $*\" >&2\nexit 1\n"
	}
	hazardStub := func(json, argv string) string {
		rec := ""
		if argv != "" {
			rec = "printf '%s\\n' \"$@\" > \"" + argv + "\"\n"
		}
		return "#!/bin/sh\n" + rec + "if [ \"$1\" = record ] && [ \"$2\" = incidents ]; then\necho '[]'\nexit 0\nfi\n" +
			"if [ \"$1\" = hazards ]; then\ncat <<'JSON'\n" + json + "\nJSON\nexit 0\nfi\n" +
			"echo \"flow: unexpected invocation: $*\" >&2\nexit 1\n"
	}
	shapeCall := func(shape string) scen {
		return func(t *testing.T, f *gdcFx) gdcResult {
			// <base>/argv, named from the stub's own path so every shape
			// case shares one body (writeExec).
			writeExec(t, f.bin+"/flow", hazardStub(twoHazards, "${0%/*}/../argv"))
			args := []string{f.repo, f.changeRoot, "demo", f.principles, f.output, "", ""}
			if shape != "" {
				args = append(args, shape)
			}
			return f.runArgs(args...)
		}
	}
	argvHas := func(want ...string) check {
		return func(t *testing.T, f *gdcFx, r gdcResult) {
			b, _ := os.ReadFile(f.base + "/argv")
			for _, w := range want {
				if !gdcHasLine(string(b), w) {
					t.Fatalf("stub argv lacks the line %q: %q", w, b)
				}
			}
		}
	}

	// The plan fixtures of cases 37-42 and 51.
	plan := "# demo plan\n\n> **Execution:** header line\n\n- [ ] 1. First\n**Files:** `a`\n  - [ ] **Step 1: one**\n\n" +
		"- [ ] 2. Second\n**Files:** `b`\n\n```sh\n- [ ] 9. Not a task\n```\n\n- [ ] 3. Third\n**Files:** `c`\n"
	indentedFencePlan := "# demo plan\n\n> **Execution:** header line\n\n- [ ] 1. First\n**Files:** `a`\n\n" +
		"- [ ] 2. Second\n**Files:** `b`\n  ```sh\n- [ ] 1. Fake\n**Files:** `FAKE-MARKER`\n  ```\n\n- [ ] 3. Third\n**Files:** `c`\n"

	// satellite is new_satellite_pair: the repository's change directory
	// carries only link.md, the canonical plan lives in base/canon, and the
	// peers file points nowhere, so only the argument can reach it.
	satellite := func(t *testing.T, f *gdcFx) {
		for _, l := range []string{"proposal", "design", "tasks"} {
			if err := os.Remove(f.changeRoot + "/" + l + ".md"); err != nil {
				t.Fatal(err)
			}
		}
		writeFile(t, f.changeRoot+"/link.md", "## Part of\n\n`canon:demo`\n")
		f.canon = f.base + "/canon"
		f.canonChange = f.canon + "/spectre/changes/demo"
		writeFile(t, f.canonChange+"/proposal.md", "CANON-PROPOSAL-BODY\n")
		writeFile(t, f.canonChange+"/design.md", "CANON-DESIGN-BODY\n")
		writeFile(t, f.canonChange+"/tasks.md", "CANON-TASKS-BODY\n")
		writeFile(t, f.repo+"/spectre/peers", "canon ../canon-nowhere\n")
	}
	sat := func(ids string, viaArg bool, tweak func(t *testing.T, f *gdcFx)) scen {
		return func(t *testing.T, f *gdcFx) gdcResult {
			satellite(t, f)
			if tweak != nil {
				tweak(t, f)
			}
			canon := ""
			if viaArg {
				canon = f.canon
			}
			return f.runArgs(f.repo, f.changeRoot, "demo", f.principles, f.output, ids, canon)
		}
	}

	rows := []struct {
		label string
		scen  scen
		check check
	}{
		// CASE 1: all sources present.
		{"all sources present: exits 0", base, exit(0)},
		{"all sources present: proposal content in bundle", base, has("PROPOSAL-BODY")},
		{"all sources present: design content in bundle", base, has("DESIGN-BODY")},
		{"all sources present: tasks content in bundle", base, has("TASKS-BODY")},
		{"all sources present: principles content in bundle", base, has("PRINCIPLES-BODY")},
		// CASE 2: design.md absent.
		{"design absent: exits 0", rm("design.md"), exit(0)},
		{"design absent: reported as skipped", rm("design.md"), has("skipped: design.md (absent)")},
		{"design absent: proposal still present", rm("design.md"), has("PROPOSAL-BODY")},
		// CASE 3: every source absent.
		{"every source absent: exits 0", rm(allSources...), exit(0)},
		{"every source absent: proposal.md reported skipped", rm(allSources...), has("skipped: proposal.md (absent)")},
		{"every source absent: design.md reported skipped", rm(allSources...), has("skipped: design.md (absent)")},
		{"every source absent: tasks.md reported skipped", rm(allSources...), has("skipped: tasks.md (absent)")},
		{"every source absent: principles reported skipped", rm(allSources...), func(t *testing.T, f *gdcFx, r gdcResult) {
			has("skipped: "+f.principles+" (absent)")(t, f, r)
		}},
		// CASE 6 and 7: the header.
		{"header carries a generated instant", base, func(t *testing.T, _ *gdcFx, r gdcResult) {
			if !regexp.MustCompile(`generated: \d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ`).MatchString(r.out) {
				t.Fatalf("no generated instant:\n%s", r.out)
			}
		}},
		{"header carries the HEAD sha", base, func(t *testing.T, f *gdcFx, r gdcResult) {
			sha, err := exec.Command("git", "-C", f.repo, "rev-parse", "--short", "HEAD").Output()
			if err != nil {
				t.Fatal(err)
			}
			has("head: "+strings.TrimSpace(string(sha)))(t, f, r)
		}},
		// CASE 8: a standards-shaped file in the change root never leaks.
		{"standards-shaped file present: exits 0", gdcStandards, exit(0)},
		{"standards-shaped file content not in bundle", gdcStandards, lacks("STANDARDS-MARKER-CONTENT")},
		// CASE 9: the required <output-path> argument missing.
		{"missing output-path argument: exits 2", missingOutput, exit(2)},
		{"missing output-path argument: usage line names <output-path>", missingOutput, has("<output-path>")},
		// CASES 10 and 11: the change-name allowlist.
		{"change name with slash: exits 2", named("a/b"), exit(2)},
		{"change name with slash: error is named", named("a/b"), has("gather-dispatch-context:")},
		{"change name with glob metacharacter: exits 2", named("a*b"), exit(2)},
		// CASE 12: <change-root> through a symlinked ancestor.
		{"change-root through symlinked ancestor: exits 2", case12, exit(2)},
		{"change-root through symlinked ancestor: target content not in output", case12, lacks("TOP-SECRET-12")},
		// CASE 13: <principles-path> itself a symlink (the global install).
		{"principles-path is a symlink: exits 0", case13, exit(0)},
		{"principles-path is a symlink: target content read", case13, has("REAL-PRINCIPLES-BODY")},
		{"principles-path is a symlink: not reported skipped", case13, func(t *testing.T, f *gdcFx, r gdcResult) {
			lacks("skipped: "+f.principles)(t, f, r)
		}},
		// CASE 14: <principles-path> through a symlinked ancestor.
		{"principles-path through symlinked ancestor: exits 0", case14, exit(0)},
		{"principles-path through symlinked ancestor: content read", case14, has("ANCESTOR-PRINCIPLES-BODY")},
		// CASE 15: <change-root> outside the worktree.
		{"change-root outside the worktree: exits 2", case15, exit(2)},
		// CASE 16: proposal.md a symlink resolving outside the change root.
		{"symlinked proposal.md outside change-root: exits 0", case16, exit(0)},
		{"symlinked proposal.md outside change-root: target content not in bundle", case16, lacks("TOP-SECRET-15")},
		{"symlinked proposal.md outside change-root: reported refused", case16, has("refused: proposal.md (resolves outside the change directory)")},
		// CASE 18: design.md a symlink to another file inside the change root.
		{"symlinked design.md resolving in-tree: exits 0", case18, exit(0)},
		{"symlinked design.md resolving in-tree: heading present", case18, has("## design.md")},
		{"symlinked design.md resolving in-tree: target content present", case18, has("TASKS-BODY")},
		{"symlinked design.md resolving in-tree: not refused (in-tree symlinks allowed)", case18, lacks("refused: design.md")},
		// CASES 19 and 20: a trailing slash and a relative "." worktree.
		{"worktree with trailing slash: exits 0", case19, exit(0)},
		{"worktree with trailing slash: proposal content present", case19, has("PROPOSAL-BODY")},
		{"worktree as relative path '.': exits 0", case20, exit(0)},
		{"worktree as relative path '.': proposal content present", case20, has("PROPOSAL-BODY")},
		// CASE 21: proposal.md a symlink cycle.
		{"symlink-cycle proposal.md: exits 0", case21, exit(0)},
		{"symlink-cycle proposal.md: no content leaked", case21, lacks("TOP-SECRET")},
		{"symlink-cycle proposal.md: reported skipped (absent), not refused", case21, has("skipped: proposal.md (absent)")},
		{"symlink-cycle proposal.md: not misreported as refused", case21, lacks("refused: proposal.md")},
		// CASES 22-27: the project commands section.
		{"project.md with lint and test: bundle reproduces both sections' fenced blocks verbatim",
			withProject("# Project configuration\n\n## lint\n\n```bash\nscripts/check-lint-marker.sh\n```\n\n## test\n\n```bash\nscripts/check-test-marker.sh\n```\n"),
			all(exit(0), func(t *testing.T, _ *gdcFx, r gdcResult) {
				if !regexp.MustCompile(`(?s)## project commands.*scripts/check-lint-marker\.sh.*scripts/check-test-marker\.sh`).MatchString(r.out) {
					t.Fatalf("project commands not reproduced in order:\n%s", r.out)
				}
			})},
		{"project.md with no lint/test/run: reported skipped, no project commands heading",
			withProject("# Project configuration\n\n## apps\n\nNothing here concerns lint, test or run.\n"),
			all(exit(0), has("skipped: project commands (absent)"), lacks("## project commands"))},
		{"no project.md: reported skipped, not a script error", base, all(exit(0), has("skipped: project commands (absent)"))},
		{"project.md lint section stops at next ## heading: test content not swallowed into lint",
			withProject("# Project configuration\n\n## lint\n\nLINT-ONLY-MARKER\n\n## test\n\nTEST-ONLY-MARKER\n"),
			all(exit(0), has("TEST-ONLY-MARKER"), func(t *testing.T, _ *gdcFx, r gdcResult) {
				lint, _, _ := strings.Cut(r.out[strings.Index(r.out, "### lint\n"):], "### test\n")
				if strings.Contains(lint, "TEST-ONLY-MARKER") {
					t.Fatalf("test content swallowed into lint:\n%s", r.out)
				}
			})},
		{"project.md whitespace-only lint section: treated as absent, no ### lint heading",
			withProject("# Project configuration\n\n## lint\n   \t   \n\n## test\n\nscripts/check-test-marker.sh\n"),
			all(exit(0), has("### test"), lacks("### lint"))},
		{"project.md '## linter' heading: not captured as the lint section",
			withProject("# Project configuration\n\n## linter\n\nscripts/should-not-be-captured.sh\n"),
			all(exit(0), has("skipped: project commands (absent)"), lacks("scripts/should-not-be-captured.sh"))},
		// CASES 28-36: SKIP-WHEN-UNCHANGED.
		{"first call rebuilds and writes the bundle plus its hash sidecar", base,
			all(exit(0), gdcExists(".hash"), errHas("bundle rebuilt — no cached bundle"))},
		{"unchanged second call reuses the bundle untouched", func(t *testing.T, f *gdcFx) gdcResult {
			// The harness slept a second between calls to see an mtime move;
			// back-dating the first bundle's mtime proves the same without it.
			old := time.Unix(1_000_000_000, 0)
			r := twice(func(t *testing.T, f *gdcFx) {
				f.before = gdcRead(t, f.output)
				if err := os.Chtimes(f.output, old, old); err != nil {
					t.Fatal(err)
				}
			})(t, f)
			if fi, err := os.Stat(f.output); err != nil || !fi.ModTime().Equal(old) {
				t.Fatalf("bundle rewritten on an unchanged call: %v %v", fi.ModTime(), err)
			}
			return r
		}, all(exit(0), errHas("bundle unchanged — reusing"), func(t *testing.T, f *gdcFx, _ gdcResult) {
			if got := gdcRead(t, f.output); got != f.before {
				t.Fatalf("content changed on an unchanged call:\n%s\n---\n%s", f.before, got)
			}
		})},
		{"editing tasks.md triggers a rebuild carrying the new content", twice(func(t *testing.T, f *gdcFx) {
			writeFile(t, f.changeRoot+"/tasks.md", "EDITED-TASKS-BODY\n")
		}), all(exit(0), errHas("bundle rebuilt — inputs changed"), has("EDITED-TASKS-BODY"))},
		{"a missing .hash sidecar forces a rebuild", twice(func(t *testing.T, f *gdcFx) {
			gdcRemove(t, f.output+".hash")
		}), all(exit(0), errHas("bundle rebuilt — no cached bundle"))},
		// Substitutes CASE 32 ("no hash tool on PATH rewrites the bundle every
		// call with no stale .hash"): the Go port hashes with crypto/sha256,
		// so a PATH with no shasum, sha256sum or openssl -- this case's bin
		// directory holds nothing at all -- still writes the sidecar and
		// still reuses an unchanged bundle.
		{"no hash tool on PATH: the Go port needs none and still reuses an unchanged bundle",
			twice(func(t *testing.T, f *gdcFx) {
				for _, tool := range []string{"shasum", "sha256sum", "openssl"} {
					if _, ok := lookPath(f.env(), tool); ok {
						t.Fatalf("%s is on the guard's PATH", tool)
					}
				}
			}), all(exit(0), gdcExists(".hash"), errHas("bundle unchanged — reusing"))},
		// Substitutes CASE 33 (the no-hash-tool path's `rm -f` of a stale
		// .hash): with no such path, the property it protected -- a sidecar
		// from an earlier body never reuses a bundle for a later one -- is
		// asserted directly: edit, then restore, rebuilds both times.
		{"a stale .hash never reuses the bundle: an edit and its revert both rebuild",
			func(t *testing.T, f *gdcFx) gdcResult {
				r := twice(func(t *testing.T, f *gdcFx) {
					writeFile(t, f.changeRoot+"/tasks.md", "EDITED-TASKS-BODY-33\n")
				})(t, f)
				if !strings.Contains(r.err, "bundle rebuilt — inputs changed") {
					t.Fatalf("edit did not rebuild: %s", r.err)
				}
				writeFile(t, f.changeRoot+"/tasks.md", "TASKS-BODY\n")
				return f.run()
			}, all(exit(0), errHas("bundle rebuilt — inputs changed"), has("TASKS-BODY"), lacks("EDITED-TASKS-BODY-33"))},
		{"a missing bundle file with an intact .hash still forces a rebuild", twice(func(t *testing.T, f *gdcFx) {
			gdcRemove(t, f.output)
		}), all(exit(0), gdcExists(""), errHas("bundle rebuilt — no cached bundle"))},
		// Substitutes CASE 35 (sha256_hex's openssl-only branch): the sidecar
		// is the bare hex SHA-256 of the bundle's body -- everything after
		// the header's generated:/head: lines, without the file's final
		// newline -- and a second call is an unchanged-skip.
		{"the .hash sidecar is the bare sha256 of the bundle body and gives a stable unchanged-skip",
			twice(func(t *testing.T, f *gdcFx) {
				bundle := gdcRead(t, f.output)
				parts := strings.SplitN(bundle, "\n", 5)
				sum := sha256.Sum256([]byte(strings.TrimSuffix(parts[4], "\n")))
				if got := gdcRead(t, f.output+".hash"); got != hex.EncodeToString(sum[:]) {
					t.Fatalf(".hash = %q, want the body's sha256 %x", got, sum)
				}
			}), all(exit(0), errHas("bundle unchanged — reusing"))},
		{"the written bundle ends with a trailing newline", base, func(t *testing.T, f *gdcFx, _ gdcResult) {
			if b := gdcRead(t, f.output); !strings.HasSuffix(b, "\n") {
				t.Fatalf("bundle does not end with a newline: %q", b)
			}
		}},
		// CASES 37-43: the sixth argument scopes ## tasks.md.
		{"scoped to one id: header plus that task's block only", withPlan(plan, "1"), all(exit(0),
			line("> **Execution:** header line"), line("- [ ] 1. First"), line("  - [ ] **Step 1: one**"),
			lacks("Second", "Third", "Not a task"))},
		{"scoped to two ids given in reverse: document order", withPlan(plan, "3,1"), all(exit(0), lacks("Second"),
			func(t *testing.T, _ *gdcFx, r gdcResult) {
				i1, i3 := strings.Index(r.out, "\n- [ ] 1. First\n"), strings.Index(r.out, "\n- [ ] 3. Third\n")
				if i1 < 0 || i3 < 0 || i1 > i3 {
					t.Fatalf("tasks 1 and 3 missing or out of order:\n%s", r.out)
				}
			})},
		{"a fenced task-line lookalike is not a task: exit 2", withPlan(plan, "9"), all(exit(2), errHas("task 9 not found in tasks.md"))},
		{"an unknown id exits 2 naming it", withPlan(plan, "7"), all(exit(2), errHas("task 7 not found in tasks.md"))},
		{"the five-argument call keeps the whole plan", withPlan(plan, ""), all(exit(0), line("- [ ] 2. Second"), line("- [ ] 9. Not a task"))},
		{"an indented (1-3 space) fence still hides its lookalike task line", withPlan(indentedFencePlan, "1,3"), all(exit(0),
			lacks("FAKE-MARKER", "Fake", "Second"), line("- [ ] 1. First"), line("- [ ] 3. Third"))},
		{"task ids given while tasks.md is absent exits 2 naming it", func(t *testing.T, f *gdcFx) gdcResult {
			gdcRemove(t, f.changeRoot+"/tasks.md")
			return f.run("1")
		}, all(exit(2), errHas("task ids given but tasks.md is absent or refused"))},
		// CASES 44-47: ## incidents.
		{"incidents with two rows: table present with newer row first, counted as found", withFlow(incidentStub(twoIncidents)), all(exit(0),
			func(t *testing.T, _ *gdcFx, r gdcResult) {
				if !regexp.MustCompile(`(?m)^found: 5 source\(s\)`).MatchString(r.out) {
					t.Fatalf("incidents not counted as found:\n%s", r.out)
				}
				if n, o := strings.Index(r.out, "check-task-commit-fields"), strings.Index(r.out, "check-unfinished-work"); n < 0 || o < 0 || n > o {
					t.Fatalf("newer row not first:\n%s", r.out)
				}
			},
			has("## incidents", "| 2026-09-03 | check-task-commit-fields | Baseline revert re-entered mid-flight | aborted revert, restored dir from stash^3 | 55 | kan-423 |"))},
		{"incidents empty: skipped as (none), no section", withFlow(incidentStub("[]")), all(exit(0), line("skipped: incidents (none)"), lacks("## incidents"))},
		{"incidents no flow on PATH: skipped as (flow unavailable), exit 0", base, all(exit(0), line("skipped: incidents (flow unavailable)"), lacks("## incidents"))},
		{"a new incident changes the body hash, so the bundle is rebuilt", func(t *testing.T, f *gdcFx) gdcResult {
			writeExec(t, f.bin+"/flow", incidentStub(`[{"occurredAt":"2026-01-01T00:00:00Z","guard":"g","symptom":"s","recovery":"r","minutesLost":1,"change":null}]`))
			f.run()
			writeExec(t, f.bin+"/flow", incidentStub(`[{"occurredAt":"2026-01-02T00:00:00Z","guard":"g","symptom":"s2","recovery":"r","minutesLost":1,"change":null}]`))
			return f.run()
		}, all(errHas("bundle rebuilt"), has("| s2 |"))},
		// CASES 84-88: ## hazards.
		{"hazards with two rows: section present with one bullet per row, counted as found", withFlow(hazardStub(twoHazards, "")), all(exit(0),
			func(t *testing.T, _ *gdcFx, r gdcResult) {
				if !regexp.MustCompile(`(?m)^found: 5 source\(s\)`).MatchString(r.out) {
					t.Fatalf("hazards not counted as found:\n%s", r.out)
				}
			},
			has("## hazards",
				"- **commit-fields-guard-reverts-head-only (all):** Never stash, revert or reset inside a worktree a guard may read; inspect a timed-out guard call first",
				"- **commit-fields-guard-single-repo (single-repo):** pass the change own worktree, never a sibling"))},
		{"hazards empty: skipped as (none), no section", withFlow(hazardStub("[]", "")), all(exit(0), line("skipped: hazards (none)"), lacks("## hazards"))},
		{"hazards no flow on PATH: skipped as (flow unavailable), exit 0", base, all(exit(0), line("skipped: hazards (flow unavailable)"), lacks("## hazards"))},
		{"eighth argument cross-repo forwarded to flow hazards as -shape cross-repo", shapeCall("cross-repo"), all(exit(0), argvHas("-shape", "cross-repo"))},
		{"shape omitted: stub still receives -shape all (fail-open)", shapeCall(""), all(exit(0), argvHas("-shape", "all"))},
		// CASES 48 and 49: a missing principles path shaped like a skip label.
		{"principles-path with mid-string paren: exits 0", principlesAt("weird (dir)/PRINCIPLES-MISSING.md"), exit(0)},
		{"principles-path with mid-string paren: absent suffix appended", principlesAt("weird (dir)/PRINCIPLES-MISSING.md"), gdcSkippedPrinciples},
		{"principles-path shaped 'notes (v2)': absent suffix appended", principlesAt("notes (v2)"), gdcSkippedPrinciples},
		{"principles-path shaped 'foo(bar)': absent suffix appended", principlesAt("foo(bar)"), gdcSkippedPrinciples},
		// CASES 50-57: the plan lives in one tree.
		{"satellite via argument: canonical plan labeled in, satellite commands and head kept", sat("", true, func(t *testing.T, f *gdcFx) {
			writeFile(t, f.repo+"/.flow/project.md", "# sat project\n\n## lint\n\nSAT-LINT-MARKER\n")
			writeFile(t, f.canon+"/.flow/project.md", "# canon project\n\n## lint\n\nCANON-LINT-MARKER\n")
		}), all(exit(0),
			has("## proposal.md (canonical canon:demo)", "## design.md (canonical canon:demo)", "## tasks.md (canonical canon:demo)",
				"CANON-PROPOSAL-BODY", "CANON-DESIGN-BODY", "CANON-TASKS-BODY", "SAT-LINT-MARKER"),
			lacks("CANON-LINT-MARKER"),
			func(t *testing.T, f *gdcFx, r gdcResult) {
				sha, err := exec.Command("git", "-C", f.repo, "rev-parse", "--short", "HEAD").Output()
				if err != nil {
					t.Fatal(err)
				}
				has("head: "+strings.TrimSpace(string(sha)))(t, f, r)
			})},
		{"satellite scoped: canonical and scoped labels compose", sat("1", true, func(t *testing.T, f *gdcFx) {
			writeFile(t, f.canonChange+"/tasks.md", "# demo plan\n\n> **Execution:** header line\n\n- [ ] 1. First\n**Files:** `a`\n  - [ ] **Step 1: one**\n\n- [ ] 2. Second\n**Files:** `b`\n")
		}), all(exit(0), has("## tasks.md (canonical canon:demo) (scoped to task(s) 1)"), line("- [ ] 1. First"), lacks("Second"))},
		{"satellite via peers: resolves and labels without the argument", func(t *testing.T, f *gdcFx) gdcResult {
			for _, l := range []string{"proposal", "design", "tasks"} {
				gdcRemove(t, f.changeRoot+"/"+l+".md")
			}
			writeFile(t, f.changeRoot+"/link.md", "## Part of\n\n`peerc:demo`\n")
			writeFile(t, f.repo+"/spectre/peers", "peerc ../gather-canon-peer\n")
			peer := f.base + "/gather-canon-peer/spectre/changes/demo"
			writeFile(t, peer+"/proposal.md", "PEER-PROPOSAL-BODY\n")
			writeFile(t, peer+"/design.md", "PEER-DESIGN-BODY\n")
			writeFile(t, peer+"/tasks.md", "PEER-TASKS-BODY\n")
			return f.runArgs(f.repo, f.changeRoot, "demo", f.principles, f.output, "", "")
		}, all(exit(0), has("## tasks.md (canonical peerc:demo)", "PEER-TASKS-BODY"))},
		{"unresolvable satellite: three distinct skips, exit 0, bundle written", sat("", false, nil), all(exit(0), lacks("CANON-"),
			func(t *testing.T, f *gdcFx, r gdcResult) {
				for _, l := range []string{"proposal.md", "design.md", "tasks.md"} {
					line("skipped: "+l+" (satellite plan unresolved — link.md at "+f.changeRoot+"/link.md)")(t, f, r)
				}
			})},
		{"unresolvable satellite with ids: exit 2 naming it", sat("1", false, nil), all(exit(2), errHas("task ids given but tasks.md is absent or refused"))},
		{"plain change with the argument: labels byte-identical, no suffix", func(t *testing.T, f *gdcFx) gdcResult {
			return f.runArgs(f.repo, f.changeRoot, "demo", f.principles, f.output, "", f.repo)
		}, all(exit(0), lacks("(canonical"), line("## proposal.md"), line("## design.md"), line("## tasks.md"))},
		{"a satellite's canonical leaf symlink outside the content dir is refused, exit 0", sat("", true, func(t *testing.T, f *gdcFx) {
			gdcReplaceWithSymlink(t, secret(t, f, "56")+"/secret.md", f.canonChange+"/proposal.md")
		}), all(exit(0), lacks("TOP-SECRET-56"), has("refused: proposal.md (canonical canon:demo) (resolves outside the change directory)"))},
		{"a ## Parts-only link.md is not a satellite: ordinary absent-leaf handling", func(t *testing.T, f *gdcFx) gdcResult {
			gdcRemove(t, f.changeRoot+"/tasks.md")
			writeFile(t, f.changeRoot+"/link.md", "## Parts\n\n`peerz:some-part`\n")
			return f.run()
		}, all(exit(0), line("skipped: tasks.md (absent)"), lacks("satellite plan unresolved"))},
	}

	for _, row := range rows {
		t.Run(row.label, func(t *testing.T) {
			t.Parallel()
			f := gdcNewRepo(t, tmpl)
			r := row.scen(t, f)
			row.check(t, f, r)
		})
	}
}

// gdcFx is one case's sandbox: new_repo's REPO, CHANGE_ROOT, PRINCIPLES and
// OUTPUT_PATH, plus the guard's PATH (bin) and working directory.
type gdcFx struct {
	base, repo, changeRoot, principles, output, bin, dir string
	worktreeArg                                          string // defaults to repo
	canon, canonChange                                   string // set by the satellite cases
	before                                               string // a bundle captured mid-case
}

type gdcResult struct {
	rc       int
	err, out string // out is the bundle (when written) plus stderr, as the harness's OUT
}

func (f *gdcFx) env() Env {
	return Env{Getenv: func(k string) string {
		if k == "PATH" {
			return f.bin
		}
		return ""
	}, Dir: f.dir}
}

func (f *gdcFx) run(extra ...string) gdcResult {
	wt := f.worktreeArg
	if wt == "" {
		wt = f.repo
	}
	return f.runArgs(append([]string{wt, f.changeRoot, "demo", f.principles, f.output}, extra...)...)
}

func (f *gdcFx) runArgs(args ...string) gdcResult {
	fn := Registry["gather-dispatch-context"]
	if fn == nil {
		panic("gather-dispatch-context is not registered")
	}
	var stdout, stderr bytes.Buffer
	rc := fn(args, f.env(), &stdout, &stderr)
	if stdout.Len() != 0 {
		panic("gather-dispatch-context wrote to stdout: " + stdout.String())
	}
	r := gdcResult{rc: rc, err: stderr.String(), out: stderr.String()}
	if b, err := os.ReadFile(f.output); err == nil {
		r.out = string(b) + stderr.String()
	}
	return r
}

// gdcTemplate is new_repo's git repository, built once per test run and
// copied by every case.
func gdcTemplate(t *testing.T) string {
	dir := t.TempDir() + "/repo"
	gitRun(t, "", "init", "-q", dir)
	gitRun(t, dir, "commit", "-q", "--allow-empty", "-m", "init")
	return dir
}

// gdcNewRepo copies the template and populates new_repo's change directory
// and principles file. base is canonicalised, as the harness's `cd -P`
// canonicalises REPO past macOS's /var -> /private/var alias.
func gdcNewRepo(t *testing.T, tmpl string) *gdcFx {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := &gdcFx{base: base, repo: base + "/repo", bin: base + "/bin", dir: base}
	gdcCopyTree(t, tmpl, f.repo)
	mkdir(t, f.bin)
	f.changeRoot = f.repo + "/spectre/changes/demo"
	writeFile(t, f.changeRoot+"/proposal.md", "PROPOSAL-BODY\n")
	writeFile(t, f.changeRoot+"/design.md", "DESIGN-BODY\n")
	writeFile(t, f.changeRoot+"/tasks.md", "TASKS-BODY\n")
	f.principles = f.repo + "/PRINCIPLES.md"
	writeFile(t, f.principles, "PRINCIPLES-BODY\n")
	f.output = f.repo + "/.superpowers/sdd/dispatch-context.md"
	return f
}

func gdcCopyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := dst + strings.TrimPrefix(p, src)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, fi.Mode().Perm())
	})
	if err != nil {
		t.Fatal(err)
	}
}

func gdcStandards(t *testing.T, f *gdcFx) gdcResult {
	writeFile(t, f.changeRoot+"/standards.md", "STANDARDS-MARKER-CONTENT\n")
	return f.run()
}

func gdcSkippedPrinciples(t *testing.T, f *gdcFx, r gdcResult) {
	if !strings.Contains(r.out, "skipped: "+f.principles+" (absent)") {
		t.Fatalf("no absent skip for %q:\n%s", f.principles, r.out)
	}
}

func gdcExists(suffix string) func(t *testing.T, f *gdcFx, r gdcResult) {
	return func(t *testing.T, f *gdcFx, _ gdcResult) {
		if !isFile(f.output + suffix) {
			t.Fatalf("%s%s was not written", f.output, suffix)
		}
	}
}

func gdcHasLine(s, l string) bool {
	for _, x := range strings.Split(s, "\n") {
		if x == l {
			return true
		}
	}
	return false
}

func gdcRead(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func gdcRemove(t *testing.T, p string) {
	t.Helper()
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
}

func gdcSymlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
}

// gdcReplaceWithSymlink is `rm -f link; ln -s target link`.
func gdcReplaceWithSymlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Remove(link); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	gdcSymlink(t, target, link)
}

// TestGdcDirnameMatchesDirname pins gdcDirname against the real dirname(1).
func TestGdcDirnameMatchesDirname(t *testing.T) {
	t.Parallel()
	for _, in := range []string{"/", "//", "/a", "/a/", "a", "a/b", "a//b/", "/a/b"} {
		want, err := exec.Command("dirname", in).Output()
		if err != nil {
			t.Fatal(err)
		}
		if got := gdcDirname(in); got != strings.TrimSuffix(string(want), "\n") {
			t.Errorf("gdcDirname(%q) = %q, dirname(1) = %q", in, got, want)
		}
	}
}

// TestGatherDispatchWorktreeRoot: worktree '/' gets the bash guard's verdict line.
func TestGatherDispatchWorktreeRoot(t *testing.T) {
	t.Parallel()
	f := gdcNewRepo(t, gdcTemplate(t))
	f.worktreeArg = "/"
	r := f.run()
	want := "gather-dispatch-context: change-root '" + f.changeRoot + "' resolves outside the worktree '/'\n"
	if r.rc != 2 || !strings.Contains(r.err, want) {
		t.Fatalf("rc=%d err=%q, want 2 and %q", r.rc, r.err, want)
	}
}
