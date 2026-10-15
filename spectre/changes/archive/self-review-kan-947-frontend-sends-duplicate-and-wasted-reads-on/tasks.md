# self-review-kan-947-frontend-sends-duplicate-and-wasted-reads-on

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer; never write `spectre/changes/`.
- Every prose block below is copied verbatim, line breaks aside: this change's `verbatim-moves.txt` acknowledges each new or reworded sentence by its exact text, and any other wording fails `scripts/check-verbatim-moves.sh`.
- Go in `stats/` follows the Go skills `CLAUDE.md` names; run `cd stats && gofmt -w .` before `gofmt -l .`.

- [x] 1. A test-building reproducer runs with a raised bound
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -rn RUN_REPRODUCER_BOUND_SECONDS skills commands-claude rules` — no hit, the defect: no run-loaded file names the override — and read `stats/internal/guard/runreproducer.go`'s comment above `bound, boundLabel := rrDeadline(` and `rrDeadline` itself: the comment says the override "can only ever shrink the wait" while `rrDeadline` accepts any all-digit value.
  - [x] **Step 2: Correct the comment.** In `stats/internal/guard/runreproducer.go`, replace these five comment lines

```go verified:read at e47bb947 in runReproducer, above the rrDeadline call
	// wall clock per case, and the override is monotone in the safe direction —
	// it can only ever shrink the wait, producing MORE timeouts, never fewer,
	// so no value of it can turn an unverified run into a verified one. Never
	// set either for a normal invocation; the Go tests inject
	// Env.ReproducerBound and Env.ReproducerGrace instead.
```

    with:

```go verified:applied at e47bb947 — go vet ./internal/guard/ and gofmt -l clean
	// wall clock per case. The override is unclamped in both directions: a
	// smaller value produces more timeouts, and a larger one lets a reproducer
	// that builds or runs tests -- a Gradle suite outlives 20 seconds -- reach
	// its verdict, the one normal invocation that sets it
	// (skills/flow/review-panel-fix-round.md). A larger bound only lets a run
	// finish; it never changes what the run's exit code means. The Go tests
	// inject Env.ReproducerBound and Env.ReproducerGrace instead.
```

  - [x] **Step 3: Add the paragraph.** In `skills/flow/review-panel-fix-round.md`, directly after the paragraph ending "(KAN-658) — the same tree the exit-contract guard above already ran this reproducer in." and before the paragraph opening "Read its exit code:", insert as its own paragraph, a blank line either side:

```markdown verified:check-verbatim-moves.sh, check-references.sh, check-markdown-integrity.py, check-dispatch-paragraphs.sh, check-installed-citations.sh and check-normative-inventory.sh exit 0 with every task of this plan applied and this change's verbatim-moves.txt present, spiked at e47bb947
**A reproducer that builds or runs tests outlives the runner's 20-second bound.** Every run of
one — the exit-contract guard above, the per-finding run here and its re-run below — carries
`RUN_REPRODUCER_BOUND_SECONDS=<the seconds that build or test run needs>` in front of the
command, so it reaches a verdict rather than a timeout.
```

  - [x] **Step 4: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/guard/`, then from the repository root `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py` — all exit 0 with `gofmt -l` printing nothing. Commit.

**Build:** green
**Files:** `stats/internal/guard/runreproducer.go`, `skills/flow/review-panel-fix-round.md`
**Tests:** **none** — a comment and one prose paragraph; the guards in step 4 cover them.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): run a test-building reproducer with a raised bound
**After:** none

**Decision:** reproducer-bound-override

- [x] 2. Cross-repo reproducer declarations carry the repository prefix
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Read the WORKTREES paragraph in `skills/flow/review-panel.md` (the quote opening "> **WORKTREES:** this change spans") and confirm it prefixes only a finding's `file:line`, nothing of `# demonstrates:` or `# premise:` — the defect — while `pcResolveCitation` in `stats/internal/guard/panelexitcontract.go` already resolves a `<repo>:` prefix on both.
  - [x] **Step 2: Add the clause.** In that paragraph, replace the line

```markdown verified:read at e47bb947 in review-panel.md's WORKTREES paragraph
> its main checkout's basename (`gymie-frontend:src/Foo.tsx:42`) —, and author that finding's reproducer per the cwd contract the
```

    with these three lines, the paragraph's other lines unchanged:

```markdown verified:check-verbatim-moves.sh, check-references.sh, check-markdown-integrity.py, check-dispatch-paragraphs.sh, check-installed-citations.sh and check-normative-inventory.sh exit 0 with every task of this plan applied and this change's verbatim-moves.txt present, spiked at e47bb947
> its main checkout's basename (`gymie-frontend:src/Foo.tsx:42`) —, prefix the same way every
> `# demonstrates:` and `# premise:` path its reproducer cites in a worktree other than the
> canonical one, and author that finding's reproducer per the cwd contract the
```

  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-dispatch-paragraphs.sh` — all exit 0. Commit.

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** **none** — one prose clause; the guards in step 3 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): prefix cross-repo reproducer declarations with the repository name
**After:** none

**Decision:** prefix-reproducer-declarations

- [x] 3. Read no state by its own line, never through a chain
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run `flow state get no-such-change-xyz -C .` as its own call and record the exit code and both streams: exit 1, stdout empty, stderr `flow: no state recorded for <project>/no-such-change-xyz`. Then confirm **Reading the state** in `skills/flow/SKILL.md` reads any exit 1 as no state — the defect.
  - [x] **Step 2: Add the paragraph.** In `skills/flow/SKILL.md`, **Reading the state**, directly after the fenced `flow state get <name-or-best-guess> -C <repo-root>` block and before the bullet opening "- **Exit 1**, or exit 0 with", insert as its own paragraph, a blank line either side:

```markdown verified:check-verbatim-moves.sh, check-references.sh, check-markdown-integrity.py, check-dispatch-paragraphs.sh, check-installed-citations.sh and check-normative-inventory.sh exit 0 with every task of this plan applied and this change's verbatim-moves.txt present, spiked at e47bb947
Run it as its own Bash call, never chained to another command. **No state** is exit 1 with
`flow: no state recorded for <project>/<name>` on stderr; an exit 1 without that line is a failure
to report, never a creating run.
```

  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py` — all exit 0. Commit.

**Build:** green
**Files:** `skills/flow/SKILL.md`
**Tests:** **none** — one prose paragraph; the guards in step 3 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): read no state by its stderr line, never through a chain
**After:** none

**Decision:** no-state-stderr-line

- [x] 4. The visual trigger matches another worktree's diff against the canonical `ui paths`
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Add the test of step 2 first and run it: it fails with `--repo gymie-frontend: exit 2 output "check-visual-trigger: usage: …"` — the defect: no caller can match a frontend worktree's paths against the canonical project's `gymie-frontend/**`.
  - [x] **Step 2: Write the failing test.** In `stats/internal/guard/check_visual_trigger_test.go`, change the `vtUsage` constant's `usage: check-visual-trigger.sh <project root> (changed paths on stdin)` to `usage: check-visual-trigger.sh [--repo <name>] <project root> (changed paths on stdin)`, and append at the end of the file:

```go verified:applied with step 3's code at e47bb947 — fails without the change (exit 2, usage), passes with it; the whole go test ./... passes, go vet and gofmt -l clean
// TestVisualTriggerRepoPrefix pins the kan-947 self-review fix: `--repo
// <name>` matches each changed path as `<name>/<path>`, so a cross-repo
// change's frontend-only diff matches the canonical project's
// `gymie-frontend/**`; without the flag the same diff matches nothing, and
// `--repo` with no name or no root is a usage refusal.
func TestVisualTriggerRepoPrefix(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, root+"/.flow/project.md", vtSection("`gymie-frontend/**`")+"\n")
	run := func(args ...string) (int, string) {
		var out, errb bytes.Buffer
		code := Registry["check-visual-trigger"](args, Env{Getenv: os.Getenv, Dir: t.TempDir(), Stdin: strings.NewReader("./composeApp/src/App.kt\n\n")}, &out, &errb)
		return code, out.String() + errb.String()
	}
	if code, out := run("--repo", "gymie-frontend", root); code != 0 || !strings.Contains(out, vtMatched("gymie-frontend/composeApp/src/App.kt", "gymie-frontend/**")) {
		t.Errorf("--repo gymie-frontend: exit %d output %q; want 0 and a MATCH on the prefixed path", code, out)
	}
	if code, out := run(root); code != 1 {
		t.Errorf("no --repo: exit %d output %q; want 1", code, out)
	}
	for _, args := range [][]string{{"--repo", "", root}, {"--repo", root}} {
		if code, out := run(args...); code != 2 || !strings.Contains(out, "usage") {
			t.Errorf("%q: exit %d output %q; want 2 usage", args, code, out)
		}
	}
}
```

    Run `cd stats && go test ./internal/guard/ -run 'TestVisualTriggerRepoPrefix|TestCheckVisualTrigger' -count=1` and confirm both FAIL.
  - [x] **Step 3: Implement the flag.** In `stats/internal/guard/visualtrigger.go`, `checkVisualTrigger`, replace the opening usage check

```go verified:read at e47bb947 in checkVisualTrigger
	if len(args) != 1 {
		fmt.Fprintf(stderr, "%susage: check-visual-trigger.sh <project root> (changed paths on stdin)\n", vtPrefix)
```

    with:

```go verified:applied with step 2's test at e47bb947 — the test passes, go vet and gofmt -l clean
	// `--repo <name>` matches each changed path as `<name>/<path>`: a
	// cross-repo change's other worktree's diff, matched against the
	// canonical project's `ui paths`, which name that repository's files by
	// its name (`gymie-frontend/**`).
	repo := ""
	if len(args) == 3 && args[0] == "--repo" && args[1] != "" {
		repo, args = args[1]+"/", args[2:]
	}
	if len(args) != 1 {
		fmt.Fprintf(stderr, "%susage: check-visual-trigger.sh [--repo <name>] <project root> (changed paths on stdin)\n", vtPrefix)
```

    and, in the `changed` closure below it, replace

```go verified:read at e47bb947 in checkVisualTrigger's changed closure
		return awkRecords(strings.ReplaceAll(string(in), "\x00", "")), nil
```

    with:

```go verified:applied with step 2's test at e47bb947 — the test passes, go vet and gofmt -l clean
		paths := awkRecords(strings.ReplaceAll(string(in), "\x00", ""))
		if repo != "" {
			for i, p := range paths {
				if p != "" {
					paths[i] = repo + strings.TrimPrefix(p, "./")
				}
			}
		}
		return paths, nil
```

  - [x] **Step 4: Update the script header.** In `scripts/check-visual-trigger.sh`, replace the two header lines

```bash verified:read at e47bb947 in the header's usage line
# Usage: check-visual-trigger.sh <project root>   (changed paths on stdin,
#                                                    one per line)
```

    with:

```bash verified:bash -n scripts/check-visual-trigger.sh clean with these lines applied, spiked at e47bb947
# Usage: check-visual-trigger.sh [--repo <name>] <project root>
#        (changed paths on stdin, one per line)
#
# --repo <name> matches each changed path as `<name>/<path>`: a cross-repo
# change's other worktree's diff, against the canonical project's `ui paths`,
# which name that repository's files by its main checkout's basename
# (`gymie-frontend/**`) — flow.visual-verify's step 2.
```

  - [x] **Step 5: Call it from step 2 of Visual verification.** In `skills/flow/verify-and-handoff.md`, **Visual verification**, step 2, directly after the line "   dot-slash prefix, an absolute glob, a glob with a space); nothing here restates them." and before the paragraph opening "**A `skipped` decision ends the stage at step 2.**", insert, indented three columns as step 2's text is:

````markdown verified:check-verbatim-moves.sh, check-references.sh, check-markdown-integrity.py, check-dispatch-paragraphs.sh, check-installed-citations.sh and check-normative-inventory.sh exit 0 with every task of this plan applied and this change's verbatim-moves.txt present, spiked at e47bb947
   In a run whose resolved set holds more than one worktree, every other worktree's diff is also
   matched against the canonical worktree's `ui paths`, each path prefixed with that worktree's
   repository name — its main checkout's basename, the name a glob such as `gymie-frontend/**`
   spells another repository's files with:

   ```bash
   git -C <worktree> diff --name-only <merge-base>..HEAD | check-visual-trigger.sh --repo <repo-name> <canonical-worktree>
   ```

   Its exit 0 makes the canonical worktree survive this step as the canonical worktree's own exit 0
   does.
````

  - [x] **Step 6: Count it in the full-suite visual run.** In `skills/flow/implement.md`, the paragraph opening "When the change's diff touches a `## visual verification` `ui paths` glob", replace

```markdown verified:read at e47bb947 in implement.md's full-suite visual clause
for any worktree in the resolved set, its merge base from this run's working notes; exit 1 or 2 for
every worktree skips this run)
```

    with:

```markdown verified:check-verbatim-moves.sh, check-references.sh, check-markdown-integrity.py, check-dispatch-paragraphs.sh, check-installed-citations.sh and check-normative-inventory.sh exit 0 with every task of this plan applied and this change's verbatim-moves.txt present, spiked at e47bb947
for any worktree in the resolved set, or the `--repo` run step 2 of **Visual verification**
(`skills/flow/verify-and-handoff.md`) adds exits 0 for any other worktree, its merge base from this
run's working notes; no exit 0 from either skips this run)
```

  - [x] **Step 7: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'VisualTrigger|VisualVerifyDispatched' -count=1`, then from the repository root `bash -n scripts/check-visual-trigger.sh`, `printf 'stats/web/src/App.tsx\n' | scripts/check-visual-trigger.sh .`, `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-guard-symlinks.sh` — all exit 0 with `gofmt -l` printing nothing. Commit.

**Build:** green
**Files:** `stats/internal/guard/visualtrigger.go`, `stats/internal/guard/check_visual_trigger_test.go`, `scripts/check-visual-trigger.sh`, `skills/flow/verify-and-handoff.md`, `skills/flow/implement.md`
**Tests:** `TestVisualTriggerRepoPrefix`
**Regression:** reverting the commit refuses `--repo` as a usage error (exit 2), so a frontend-only cross-repo diff never matches the canonical `gymie-frontend/**`.
**Baseline:** before=2 after=3
<!-- measured: grep -c '^func Test' stats/internal/guard/check_visual_trigger_test.go @ e47bb947 -->
**Commit:** fix(guard): visual trigger matches another worktree's diff against the canonical ui paths
**After:** none

**Decision:** visual-trigger-repo-prefix

- [x] 5. Inline verify skips a worktree with nothing outside its change directory
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Read **Inline verify** in `skills/flow/verify-and-handoff.md` and confirm it runs the lint and test list for every worktree of the resolved set, with no skip for one whose diff is only its change directory — the defect.
  - [x] **Step 2: Add the skip.** In `skills/flow/verify-and-handoff.md`, directly after the heading `### Inline verify` and before the paragraph opening "Resolve the commands `project-get.sh <worktree> lint`", insert, a blank line either side:

````markdown verified:check-verbatim-moves.sh, check-references.sh, check-markdown-integrity.py, check-dispatch-paragraphs.sh, check-installed-citations.sh and check-normative-inventory.sh exit 0 with every task of this plan applied and this change's verbatim-moves.txt present, spiked at e47bb947
**A worktree with nothing outside its change directory runs none of the list below.** In a run
whose resolved set holds more than one worktree, a worktree is skipped when this prints nothing, its
merge base from this run's working notes:

```bash
git -C <worktree> diff --name-only <merge-base>..HEAD ':(exclude)spectre/changes/<name>' ':(exclude)spectre/changes/archive/<name>'
```

The `## Report` names a skipped worktree as `- <worktree> — skipped, nothing outside its change
directory`.
````

  - [x] **Step 3: Record no row for it.** In the same file's **Recording.** paragraph, replace its first line

```markdown verified:read at e47bb947 in verify-and-handoff.md's Recording. paragraph
**Recording.** One `dispatches` row per worktree, `-role verifier -key verify -model <parent
```

    with this line, the rest of the paragraph unchanged:

```markdown verified:check-verbatim-moves.sh, check-references.sh, check-markdown-integrity.py, check-dispatch-paragraphs.sh, check-installed-citations.sh and check-normative-inventory.sh exit 0 with every task of this plan applied and this change's verbatim-moves.txt present, spiked at e47bb947
**Recording.** One `dispatches` row per worktree the list runs for, `-role verifier -key verify -model <parent
```
  - [x] **Step 4: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-guard-symlinks.sh` — all exit 0. Commit.

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** **none** — prose in the verify stage; the guards in step 4 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): inline verify skips a worktree with nothing outside its change directory
**After:** Task 4

**Decision:** verify-skip-change-dir-only
