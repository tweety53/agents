# self-review-kan-1021-flow-improvement-move-the-bash-hook-and-setup

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the three non-`big` findings of `kan-1021-flow-improvement-move-the-bash-hook-and-setup`'s
self-review pass — the fenced citation-check body, the superpowers sub-skills a `flow-*` agent
cannot load, and the parent's in-place citation correction.

**Architecture:** one clause in `skills/flow/review-panel.md`; the sub-skill file path in two
`skills/flow/implement.md` dispatch paragraphs and one `skills/flow/review-panel-fix-round.md`
sentence, pinned by a new `subskill` row in `stats/internal/guard/dispatchparagraphs.go`; one
exception sentence in `skills/flow/review-panel-fix-round.md`'s audit dispositions.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-23-self-review-kan-1021-flow-improvement-move-the-bash-hook-and-setup-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Work in this worktree `/Users/tweety53/Projects/agents-worktrees/kan-1021-flow-improvement-move-the-bash-hook-and-setup`, branch `spectre/kan-1021-flow-improvement-move-the-bash-hook-and-setup`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- Each task's diff is the unified diff its fix step gives, applied with `git apply` from the worktree's root; all of them were applied together and checked in this worktree at `f85971db`. Save each by copying the lines between its fence markers into a scratch file outside the worktree. A diff that does not apply is reported, never hand-merged by guesswork. Each prose sentence lands character for character: `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt` exactly.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

## Review Focus

- Task 2: both REQUIRED SUB-SKILL blocks keep the plugin-cache path on one line — the guard joins a block's lines with a space, so a wrap inside the backticked path breaks its shared phrase.
- Task 3: the exception covers a line-number offset alone; a citation whose content is not on any line of the cited path is still bounced.

- [x] 1. Run the citation check's fenced lines, never an eval of its body
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `scripts/project-get.sh . "review panel citation check" | head -1`: it prints the fence line
    `` ```bash ``.
    <!-- measured: scripts/project-get.sh . "review panel citation check" | od -c @ f85971db → first line is the fence -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at f85971db together with the other two tasks' diffs; check-verbatim-moves.sh exits 0 with this change's verbatim-moves.txt, check-references.sh, check-markdown-integrity.sh and check-installed-citations.sh exit 0
diff --git a/skills/flow/review-panel.md b/skills/flow/review-panel.md
index 09122798..c68cc81b 100644
--- a/skills/flow/review-panel.md
+++ b/skills/flow/review-panel.md
@@ -101,7 +101,9 @@ check-panel-citation-trigger.sh <worktree> <merge-base>
 On exit 0, run `project-get.sh <worktree> "review panel citation check"` (**Guard resolution**,
 `skills/flow-contracts/pipeline.md`; the key itself is canonical in **Project configuration**,
 `skills/flow-contracts/project-configuration.md`). If declared, run its command
-from the apply worktree and capture combined stdout+stderr verbatim to
+from the apply worktree — the lines between the fence markers `project-get.sh` prints around it,
+never an `eval` of the printed body, whose opening `` ```bash `` line runs a bare `bash` that hangs
+on open stdin — and capture combined stdout+stderr verbatim to
 `<abs-worktree>/.superpowers/sdd/citation-check.md`. `project-get.sh` exit 1 is the "key absent —
 skip silently" case just named; exit 2 is reported and the worktree skipped, the same way the
 guard's own exit 2 below is. On exit 1, skip silently — no file written, no CITATION CHECK
````

  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.sh` and `scripts/check-installed-citations.sh` exit 0.
  - [x] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentence.
**Regression:** none — no test is declared.
**Commit:** docs(flow): run the citation check's fenced lines, never an eval of its body
**After:** none

**Decision:** citation-fence-prose

- [x] 2. Name the superpowers skill files a flow-* agent reads
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n "Skill" agents/flow-medium.md; grep -n "REQUIRED SUB-SKILL" skills/flow/implement.md`:
    the agent file grants no Skill tool and both paragraphs say "Use"/"invoke" with no file path.
    <!-- measured: the two greps above @ f85971db → no Skill grant; paragraphs at implement.md lines 526 and 529 -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`) — the two paragraphs, the panel-fix
    sentence and the guard row:

````diff verified:applied with git apply to this worktree at f85971db together with the other two tasks' diffs; check-verbatim-moves.sh exits 0 with this change's verbatim-moves.txt, check-references.sh, check-markdown-integrity.sh and check-installed-citations.sh exit 0; with it check-dispatch-paragraphs.sh prints DISPATCH-PARAGRAPHS-OK, 34 sites, and with the implement.md hunk reverted alone it prints DISPATCH-PARAGRAPHS-INVALID, 6 violations
diff --git a/skills/flow/implement.md b/skills/flow/implement.md
index 3e611fd0..20bcb700 100644
--- a/skills/flow/implement.md
+++ b/skills/flow/implement.md
@@ -523,11 +523,13 @@ output — then the green partner, and makes **one** commit for the pair: the pa
 `**Commit:**` subject and `Task-Id: <partner>` trailer. The red task never has a commit of its
 own; `check-task-commit-fields.sh` resolves the pair from either id against that commit.
 
-> **REQUIRED SUB-SKILL:** Use superpowers:test-driven-development — RED-GREEN-REFACTOR for this
-> task. Delete any code written before its test.
+> **REQUIRED SUB-SKILL:** Read and follow superpowers:test-driven-development from its file — on
+> Claude Code the newest `~/.claude/plugins/cache/*/superpowers/*/skills/test-driven-development/SKILL.md`
+> — RED-GREEN-REFACTOR for this task. Delete any code written before its test.
 
-> **REQUIRED SUB-SKILL:** When a test fails for a reason RED-GREEN-REFACTOR did not plan, invoke
-> superpowers:systematic-debugging before writing a fix. An expected RED step needs no invocation.
+> **REQUIRED SUB-SKILL:** When a test fails for a reason RED-GREEN-REFACTOR did not plan, read and
+> follow superpowers:systematic-debugging from its file — on Claude Code the newest `~/.claude/plugins/cache/*/superpowers/*/skills/systematic-debugging/SKILL.md`
+> — before writing a fix. An expected RED step needs no such read.
 
 > **PROVE THE GUARD BITES:** When this task's tests assert on configuration or file content — a
 > guard script, an embedded config, a fixture file — a passing run alone is not evidence: break
diff --git a/skills/flow/review-panel-fix-round.md b/skills/flow/review-panel-fix-round.md
index 61b35457..a8ea198b 100644
--- a/skills/flow/review-panel-fix-round.md
+++ b/skills/flow/review-panel-fix-round.md
@@ -432,8 +436,10 @@ to exactly this shape (**Before closing the stage**, below). Inline
 itself under the same paragraphs, dispatching no subagent, and records the pass with `-role
 panel-fix -model <parent model> -effort <parent effort> -agent-id inline`, `begin` before its first
 edit and `end` once the fix commit lands, as that section's **Records** bullet states. Where a finding is
-confirmed as a real defect, the fix subagent invokes **superpowers:systematic-debugging** before
-writing its fix. **Dispatch it on the decision's `fixer` object** — its own model and effort,
+confirmed as a real defect, the fix subagent reads and follows **superpowers:systematic-debugging**
+from its file before writing its fix — on Claude Code the newest
+`~/.claude/plugins/cache/*/superpowers/*/skills/systematic-debugging/SKILL.md`, since a `flow-*` agent
+carries no Skill tool. **Dispatch it on the decision's `fixer` object** — its own model and effort,
 chosen apart from the implementer's, `subagent_type:
 flow-<effort>` with that `model` passed as the Agent tool's own `model` parameter, and
 `-model`/`-effort` below carry that pair. On harness `zcode` the pair given and recorded is `glm-5.3-flash` / `high` instead (**Harness mapping**, `skills/flow-contracts/model-policy.md`); on harness `muse` the dispatch names no model and no `subagent_type`, and the recorded pair is the session's own. Record
diff --git a/stats/internal/guard/dispatchparagraphs.go b/stats/internal/guard/dispatchparagraphs.go
index beb4740e..5ba512bd 100644
--- a/stats/internal/guard/dispatchparagraphs.go
+++ b/stats/internal/guard/dispatchparagraphs.go
@@ -76,6 +76,9 @@ var dpEntries = map[string]dpEntry{
 		[]string{"never mutate it", "no file edit outside your own report file", "a claim nobody can check"}, nil},
 	"fixround": {"**FIX-ROUND SCOPE:**",
 		[]string{"this is a fix-round re-review", "and nothing else on the branch", "never the module, repository or live-spec suite"}, nil},
+	"subskill": {"**REQUIRED SUB-SKILL:**",
+		[]string{"from its file", "on Claude Code the newest `~/.claude/plugins/cache/*/superpowers/*/skills/"},
+		map[string]string{"tdd": "skills/test-driven-development/SKILL.md`", "debugging": "skills/systematic-debugging/SKILL.md`"}},
 	"proofruns": {"**PROOF RUNS:**",
 		[]string{"size the run count to the flake rate you measured", "never a fixed count per step", "shows the check failing with the mechanism removed"}, nil},
 }
@@ -128,6 +131,7 @@ var dpSites = []dpSite{
 	{"readonly", "skills/flow/implement.md", 1, nil},
 	{"fixround", "skills/flow/review-panel-fix-round.md", 1, nil},
 	{"proofruns", "skills/flow/implement.md", 1, nil},
+	{"subskill", "skills/flow/implement.md", 2, []string{"tdd", "debugging"}},
 }
 
 func checkDispatchParagraphs(_ []string, env Env, stdout, stderr io.Writer) int {
````

  - [x] **Step 3: Update `stats/internal/guard/check_dispatch_paragraphs_test.go`.** With the row in,
    seven existing subtests of `TestCheckDispatchParagraphs` fail, every one on an `implement.md`
    fixture that now lacks the two blocks or on the `33 site(s)` OK line.
    <!-- measured: cd stats && go test ./internal/guard/ -run 'TestCheckDispatchParagraphs' -count=1 with Step 2 applied @ f85971db → 7 failing subtests -->
    Add two block constants beside `dpProofRunsBlock`, `dpSubSkillTDDBlock` and
    `dpSubSkillDebuggingBlock`, each the exact text of its new `implement.md` blockquote; add both
    to every clean `implement.md` fixture (`dpCleanImplementNoReadonly` and the case-1 fixture's
    `dpDoc` list among them); change the OK line's `33 site(s)` to `34 site(s)`. Then add three
    cases after case 101, each one `dpRC` line and one `dpHas` line:
    - case 102: a clean `implement.md` with neither block exits 1 and names
      `requires at least 2 block(s) carrying the label \"**REQUIRED SUB-SKILL:**\", found 0`;
    - case 103: the debugging block without `from its file` exits 1 and names
      `missing the required phrase: \"from its file\"`;
    - case 104: the TDD block twice and no debugging block exits 1 and names
      `missing a block that satisfies the debugging variant`.
  - [x] **Step 4: Prove the guard bites.** Save `skills/flow/implement.md`, revert its two
    paragraphs to their `f85971db` text, run `scripts/check-dispatch-paragraphs.sh` and capture its
    `DISPATCH-PARAGRAPHS-INVALID` line, restore the saved file, and capture the OK line.
  - [x] **Step 5: Verify.** `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckDispatchParagraphs' -count=1`
    passes; `cd stats && go vet ./internal/guard && gofmt -l internal/guard` prints nothing;
    `scripts/check-dispatch-paragraphs.sh`, `scripts/check-verbatim-moves.sh`,
    `scripts/check-references.sh`, `scripts/check-markdown-integrity.sh` and
    `scripts/check-installed-citations.sh` exit 0.
  - [x] **Step 6: Commit** the four files with the subject below.

**Build:** green
**Files:** `skills/flow/implement.md`, `skills/flow/review-panel-fix-round.md`, `stats/internal/guard/dispatchparagraphs.go`, `stats/internal/guard/check_dispatch_paragraphs_test.go`
**Tests:** Case 102, Case 103, Case 104
**Regression:** Case 102, Case 103, Case 104 — reverting the commit removes the `subskill` row, and each case exits 0 where it expects 1.
**Baseline:** before=101 after=104
<!-- measured: grep -oE '"case [0-9]+[a-z]?:' stats/internal/guard/check_dispatch_paragraphs_test.go | sort -u | grep -c . @ f85971db -->
**Commit:** fix(flow): name the superpowers skill files a flow-* agent reads
**After:** Task 1

**Decision:** subskill-file-path
**Decision:** subskill-guard-row
**Decision:** flow-fast-and-shim-unchanged

- [x] 3. Let the parent correct a reproducer citation's line number in place
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run
    `grep -n "bounced exactly as the per-finding run's own answer 1" skills/flow/review-panel-fix-round.md`:
    the audit's exit-1 dispositions bounce every unresolving demonstrates citation, with no exception.
    <!-- measured: the grep above @ f85971db → one hit, line 234 -->
  - [x] **Step 2: Apply the fix** (`git apply <the saved diff>`):

````diff verified:applied with git apply to this worktree at f85971db together with the other two tasks' diffs; check-verbatim-moves.sh exits 0 with this change's verbatim-moves.txt, check-references.sh, check-markdown-integrity.sh and check-installed-citations.sh exit 0
diff --git a/skills/flow/review-panel-fix-round.md b/skills/flow/review-panel-fix-round.md
index 61b35457..a8ea198b 100644
--- a/skills/flow/review-panel-fix-round.md
+++ b/skills/flow/review-panel-fix-round.md
@@ -234,7 +234,11 @@ whose script cannot be read to audit at all, are
 bounced exactly as the per-finding run's own answer 1 below,
 once, back to the raising slot; a reproducer the runner refused as unusable is recorded
 **unverifiable** and put to the round's one batched operator ask, exactly as the per-finding run's own answer 2 below, since
-a refused reproducer never ran and so carries no passing output a bounce could carry. Exit 2 — any
+a refused reproducer never ran and so carries no passing output a bounce could carry. An
+unresolving demonstrates citation whose cited path holds the claimed content verbatim on another
+line — the reviewer's line number off, nothing else — is the one exception, never bounced: the
+parent corrects the line number in the reproducer in place, re-runs the audit, and records the
+correction, old line and new, in that round's `flow record pass`. Exit 2 — any
 reproducer the runner could not verdict: a timeout, a surviving process, a plumbing failure, the
 state record absent or unreadable, an ambiguous worktree — stops
 the run, the same as the lexical guard's exit 2.
````

  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`,
    `scripts/check-markdown-integrity.sh` and `scripts/check-installed-citations.sh` exit 0.
  - [x] **Step 4: Commit** the file with the subject below.

**Build:** green
**Files:** `skills/flow/review-panel-fix-round.md`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentence.
**Regression:** none — no test is declared.
**Commit:** docs(flow): let the parent correct a reproducer citation's line number in place
**After:** Task 2

**Decision:** parent-line-correction
