# Self-review context bundle for kan-839-flow-fix-reproducers-pass-vacuously-when-a-fix

found: 6 of 7 sources; skipped: 1 of 7 sources
skipped: change summary (absent)

## .superpowers/sdd/ledgers/kan-839-flow-fix-reproducers-pass-vacuously-when-a-fix.md

# SDD ledger — kan-839-flow-fix-reproducers-pass-vacuously-when-a-fix

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — implementer

- Task: 1
- Role: implementer
- Key: task-1-implementer
- Model: glm-5.3-flash effort=high
- Commit: b4f20171af2eaff21d1c9017daf282b1e8f55246
- Outcome: completed
- Started: 2026-09-27T22:11:19Z
- Tokens: cost unattributed — session never bound

## Dispatch 2 — implementer

- Task: 2
- Role: implementer
- Key: task-2-implementer
- Model: glm-5.3-flash effort=high
- Commit: cb09445ecfc5d2dc7955f8bfdc41433bf7f33478
- Outcome: completed
- Started: 2026-09-27T22:17:00Z
- Tokens: not measured

## Dispatch 3 — implementer

- Task: 3
- Role: implementer
- Key: task-3-implementer
- Model: glm-5.3-flash effort=high
- Commit: 8fe9f895c11e0e5935f5a1956c01b42786e48ff8
- Outcome: completed
- Started: 2026-09-27T22:19:27Z
- Tokens: not measured

## Dispatch 4 — reviewer

- Task: 1
- Role: reviewer
- Key: task-1-reviewer
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: clean
- Started: 2026-09-27T22:23:05Z
- Tokens: not measured

## Dispatch 5 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-27T22:48:37Z
- Tokens: not measured

## Dispatch 6 — verifier

- Task: no task
- Role: verifier
- Key: verify
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-27T23:19:29Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-839-flow-fix-reproducers-pass-vacuously-when-a-fix-panel.md

# Review panel — kan-839-flow-fix-reproducers-pass-vacuously-when-a-fix

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | Minor | stats/internal/guard/panelexitcontract.go:386 | a declared-but-unasserted premise escapes both enforcement points — pcPremiseAudit only checks declarations resolve while the body assertion is prose-only — so a fix renaming the target still produces the vacuous green the proposal says can no longer happen |   |
| F2 | primary | Minor | skills/flow/review-panel.md:977 | the exit-1 one-disposition-per-class enumeration was not extended with the declared-but-unresolvable-premise class this same diff adds — the parent reading exit 1 has no stated disposition for it |   |

findings-total: 2
finding-status: F1 deferred — the design accepted this residual: the guard audits declarations, the body assertion stays an authoring rule
finding-status: F2 deferred — prose enumeration in the guard-section description lacks the new violation class

reproducers-total: 2
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh

## Pass log

### Round 0

- base moved 5 commits with overlap on skills/flow/review-panel.md; operator chose rebase; clean rebase onto 675a55a9, re-check CLEAR
- roster: compact — 12
- diff-size 594 under cap (single worktree)
- docs-only exit 1 — first non-doc path scripts/check-panel-reproducer-exit-contract.sh; resolved roster runs
- no addition this round — the resolved list ran alone
- agents ran: primary+principles one bundle on glm-5.3-flash/high; findings F1 (primary+principles, deduped), F2 (primary) — all Minor, none Critical/Important; deferral default applied, no fix round, no re-run
## spectre/changes/archive/kan-839-flow-fix-reproducers-pass-vacuously-when-a-fix/tasks.md

# kan-839-flow-fix-reproducers-pass-vacuously-when-a-fix

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

No live-verification task: this change touches guard code and contract prose, not a running
service or persistent state.

- [x] 1. Premise audit in the exit-contract guard

**Build:** green
**Files:** `stats/internal/guard/panelexitcontract.go` `stats/internal/guard/check_panel_reproducer_exit_contract_test.go` `scripts/check-panel-reproducer-exit-contract.sh`
**Tests:** `TestPanelExitContractPremiseAudit`
**Regression:** reverting loses the dispatch-time premise audit — a declared-but-unresolvable premise passes the guard silently again, the dispatch-time half of KAN-839.
**Baseline:** before=3 after=4
<!-- measured: grep -c '^func Test' stats/internal/guard/check_panel_reproducer_exit_contract_test.go @ merge-base 4a278320 (the count BEFORE this change) -->
**After:** none

**Decision:** premise-citation-form
**Decision:** premise-coverage-tolerant-guard
**Decision:** premise-mutation-exempt

**Commit:** fix(guard): premise citations resolved by the exit-contract audit

  - [x] **Step 1: Write the failing test** — append to `check_panel_reproducer_exit_contract_test.go`, using the file's own fixture (`newPCSandbox`, `pcFindings`, `s.repro`, `s.run`, `rrExpect`):

```go verified:authored in-tree for this change
func TestPanelExitContractPremiseAudit(t *testing.T) {
	t.Parallel()
	decl := "# demonstrates: target.txt:2:defect present here\n"

	t.Run("a declared premise that resolves violates nothing", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", "repro.sh"))
		s.repro(t, "repro.sh", decl+"# premise: target.txt:1:line one\nexit 9")
		got, out := s.run(t, "premise-resolves")
		expect(t, got, out, 0, "REPRODUCER-EXIT-CONTRACT-OK")
	})

	t.Run("a declared premise citing a missing file exits 1", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", "repro.sh"))
		s.repro(t, "repro.sh", decl+"# premise: absent.txt:1:no such line\nexit 9")
		got, out := s.run(t, "premise-missing-file")
		expect(t, got, out, 1, "premise")
	})

	t.Run("a malformed premise line exits 1", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", "repro.sh"))
		s.repro(t, "repro.sh", decl+"# premise: target.txt\nexit 9")
		got, out := s.run(t, "premise-malformed")
		expect(t, got, out, 1, "malformed premise")
	})

	t.Run("absence of premise lines violates nothing (tolerant)", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", "repro.sh"))
		s.repro(t, "repro.sh", decl+"exit 9")
		got, out := s.run(t, "premise-tolerant")
		expect(t, got, out, 0, "REPRODUCER-EXIT-CONTRACT-OK")
	})

	t.Run("a mutation-declared reproducer skips the premise audit", func(t *testing.T) {
		t.Parallel()
		s := newPCSandbox(t, pcFindings("F1", "open", "repro.sh"))
		s.repro(t, "repro.sh", "# mutation-reproducer\n# premise: absent.txt:1:no such line\nexit 0")
		got, out := s.run(t, "premise-mutation-exempt")
		expect(t, got, out, 0, "REPRODUCER-EXIT-CONTRACT-OK")
	})
}
```

  - [x] **Step 2: Run it to see it fail**

```bash verified:authored in-tree for this change
cd stats && go test ./internal/guard/ -run TestPanelExitContractPremiseAudit
```

Expected: FAIL — the unresolvable-premise and malformed subtests exit 0 today because `# premise:` lines are inert comments to `pcAudit`.

  - [x] **Step 3: Implement the premise audit** — in `panelexitcontract.go`:

```text verified:read in this worktree at merge-base 4a278320
pcDeclare is "# demonstrates: " (line 40); the audit is pcAudit(ref, tree, reproPath)
(line 357), called once at line 300 (audit := pcAudit(ref, tree, reproPath)).
```

Extract pcAudit's per-citation checks — shape cut on the declare prefix, `/`-prefix and `..` lexical refusal, `filepath.EvalSymlinks` resolved containment under `tree`, `isFile`, line number, line content — into a shared helper `pcResolveCitation(ref, tree, decl, prefix, noun string) []string` whose violation messages name the declaration kind (`noun`). `pcAudit` keeps its zero-demonstrations violation and its message wording unchanged. Add `pcPremiseDeclare = "# premise: "` and `pcPremiseAudit(ref, tree, reproPath) []string` that resolves every premise line through `pcResolveCitation` and returns nothing when there are none — tolerant, never a zero-declaration violation. At the line-300 call site, append `pcPremiseAudit`'s result where `pcAudit`'s is appended, skipped exactly where the demonstrates audit is skipped for mutation-declared reproducers. Violation wording follows pcAudit's, with "premise declaration" in place of "demonstrates declaration" — the step-1 needles (`premise`, `malformed premise`) must match.

  - [x] **Step 4: Run the guard's tests**

```bash verified:authored in-tree for this change
cd stats && go test ./internal/guard/ -run 'TestCheckPanelReproducerExitContract|TestPanelExitContract'
```

Expected: PASS — the existing demonstrates behavior unchanged, the new subtests green.

  - [x] **Step 5: Update the shim header** — in `skills/flow/scripts/check-panel-reproducer-exit-contract.sh`'s THE INSTRUMENT AUDIT (KAN-606) paragraph, add: the audit also resolves every `# premise: <path>:<line>:<content>` declaration by the same machinery, tolerantly — absence of premise lines violates nothing, a declared-but-unresolvable premise joins exit 1's violation classes, and mutation-declared reproducers skip it as they skip the demonstrates audit (KAN-839).

  - [x] **Step 6: Verify and commit**

```bash verified:authored in-tree for this change
cd stats && go vet ./internal/guard/ && gofmt -l internal/guard
bash -n skills/flow/scripts/check-panel-reproducer-exit-contract.sh
git add stats/internal/guard/panelexitcontract.go stats/internal/guard/check_panel_reproducer_exit_contract_test.go skills/flow/scripts/check-panel-reproducer-exit-contract.sh
git commit -m "fix(guard): premise citations resolved by the exit-contract audit"
```

Correction (2026-09-28): the plan declared the shim at `skills/flow/scripts/check-panel-reproducer-exit-contract.sh`; that path is a check-guard-symlinks rule-1 symlink whose relative target, `scripts/check-panel-reproducer-exit-contract.sh`, is the real tracked file any edit touches — the `**Files:**` field is corrected to the target path, and the commit stages that path.

- [x] 2. Premise authoring rule in the panel contract

**Build:** green
**Files:** `skills/flow/review-panel.md`
**Tests:** none
**Regression:** reverting silences the authoring contract about premises — slots stop declaring and asserting them, the task-1 audit audits nothing, and the re-run refusal path never arms.
**Baseline:** before=0 after=0
<!-- measured: grep -c '^func Test' skills/flow/review-panel.md @ merge-base 4a278320 (the count BEFORE this change) -->
**After:** none

**Decision:** premise-enforcement-split
**Decision:** premise-coverage-tolerant-guard

**Commit:** docs(flow): premise assertions in the reproducer contract

  - [x] **Step 1: Authoring paragraph** — in the "Every slot must supply, per finding, a reproducer" block, directly after the `# demonstrates:` declaration sentences, insert:

```markdown verified:authored in-tree for this change
**A runnable reproducer also declares what its checks read**: one
`# premise: <path>:<line>:<content>` line per file, test/class name or `tasks.md` task id its
checks read, in the same first-10-lines window — at least one for every runnable reproducer;
mutation-declared ones are exempt, the sha pin being their instrument audit. A premise must
resolve in every tree the script runs in, unlike the demonstrates citation, which names where
the defect was and must resolve only against the defect-present tree. The script's body asserts
every declared premise before its real checks run: a missing premise is a loud failure naming
it on stderr, exiting non-zero — never exit 0, the vacuous pass a rename must never produce
(KAN-839). Carry the premise rule on every slot's dispatch prompt.
```

  - [x] **Step 2: Guard-invocation section** — in the exit-contract guard section, directly after the sentence ending "the content appears on that line", insert:

```markdown verified:authored in-tree for this change
The audit also resolves every `# premise:` declaration the same way, tolerantly: absence of
premise lines violates nothing, and a declared-but-unresolvable premise joins exit 1's
violation classes; mutation-declared reproducers skip it as they skip the demonstrates audit.
```

  - [x] **Step 3: Re-run section** — in the fix-verification re-run paragraphs, directly after the sha-pin re-author paragraph, insert:

```markdown verified:authored in-tree for this change
A re-run whose script fails a premise assertion — the loud non-zero the authoring rule
requires — reads *demonstrated*, identical to the dispatch-time verdict, and is refused as
ambiguous by design (KAN-839): the renamed premise voids the reproducer, and the refusal
routes it to re-authoring. That refusal is the fix working, never a runner bug.
```

  - [x] **Step 4: Verify and commit**

```bash verified:authored in-tree for this change
scripts/check-vocabulary.sh && scripts/check-references.sh
git add skills/flow/review-panel.md
git commit -m "docs(flow): premise assertions in the reproducer contract"
```

- [x] 3. Composition test: renamed premise refuses, never passes

**Build:** green
**Files:** `stats/internal/guard/runreproducer_test.go`
**Tests:** `TestRunReproducerRenamedPremiseRefuses`
**Regression:** reverting un-pins the composition this change exists for — the kan-692 shape (a renamed premise passing vacuously) could return without a test noticing, the post-fix half of KAN-839.
**Baseline:** before=5 after=6
<!-- measured: grep -c '^func Test' stats/internal/guard/runreproducer_test.go @ merge-base 4a278320 (the count BEFORE this change) -->
**After:** Task 2

**Decision:** premise-enforcement-split

**Commit:** test(guard): pin the renamed-premise refusal shape

  - [x] **Step 1: Write the pinning test** — append to `runreproducer_test.go`, using the file's own helpers (`rrWorktree`, `rrFixture`, `rrRun`, `rrEnv`, `rrExpect`; the runner is unchanged, so this test passes against it — it pins the seam):

```go verified:authored in-tree for this change
func TestRunReproducerRenamedPremiseRefuses(t *testing.T) {
	t.Parallel()
	wt := rrWorktree(t)
	writeFile(t, filepath.Join(wt, "target.txt"), "line one\ndefect present here\n")
	rrFixture(t, wt, "scripts/checks.sh", `[ -f target.txt ] || { echo "premise: missing target.txt" >&2; exit 7; }
! grep -q "defect present here" target.txt`)
	// Pre-fix: the premise holds, the check finds the defect — demonstrated.
	got, out := rrRun(t, rrEnv(t), wt, "scripts/checks.sh")
	rrExpect(t, got, out, 0, "defect demonstrated")
	// The fix renames the target away; the re-run must refuse, never pass.
	if err := os.Rename(filepath.Join(wt, "target.txt"), filepath.Join(wt, "renamed.txt")); err != nil {
		t.Fatal(err)
	}
got, out = rrRun(t, rrEnv(t), wt, "scripts/checks.sh --pre-fix-verdict demonstrated")
rrExpect(t, got, out, 2, "ambiguous")
}
```

Correction (2026-09-28): `rrRun` takes the runner's argv variadically, so the flags ride as separate arguments — the shipped test calls `rrRun(t, rrEnv(t), wt, "scripts/checks.sh", "--pre-fix-verdict", "demonstrated")`, not the plan snippet's flags-inside-the-command-line form, which the runner reads as reproducer arguments and never sees as `--pre-fix-verdict`. The first RED run captured exactly that: the re-run read demonstrated (exit 0) instead of refusing.

  - [x] **Step 2: Run it**

```bash verified:authored in-tree for this change
cd stats && go test ./internal/guard/ -run TestRunReproducerRenamedPremiseRefuses
```

Expected: PASS — the runner already refuses identical verdicts; the test pins that a premise-asserting script lands there instead of in a vacuous pass.

  - [x] **Step 3: Verify and commit**

```bash verified:authored in-tree for this change
cd stats && go vet ./internal/guard/ && gofmt -l internal/guard
git add stats/internal/guard/runreproducer_test.go
git commit -m "test(guard): pin the renamed-premise refusal shape"
```
## spectre/changes/archive/kan-839-flow-fix-reproducers-pass-vacuously-when-a-fix/design.md

# kan-839-flow-fix-reproducers-pass-vacuously-when-a-fix — design

## Context

A reproducer's checks can enumerate nothing after a fix renames their target and still exit 0 —
the post-fix verification re-run has no premise audit (the exit-contract guard skips non-open
findings by design; its KAN-606 audit covers only the defect-present side). The change gives
reproducers premise declarations, asserted by the script and audited at dispatch, so a renamed
premise lands in the KAN-524 ambiguity refusal instead of a green count. Why now, in
`proposal.md`.

## Mechanism

- **Premise declaration.** Every runnable, non-mutation reproducer carries, in the same
  first-10-lines window as `# demonstrates:`, one `# premise: <path>:<line>:<content>` line per
  file, test/class name or `tasks.md` task id its checks read. Same citation shape and
  resolution machinery as demonstrates; semantics differ: demonstrates names where the defect
  **was** (resolves only against the defect-present tree, pre-fix), a premise names what the
  checks **read** (must resolve in every tree the script runs in).
- **Enforcement point 1 — the script asserts.** Authoring rule: the body asserts every declared
  premise before its real checks run; a missing premise is a loud failure naming it on stderr,
  exiting non-zero — never exit 0. The rule rides every slot's dispatch prompt the way the cwd
  contract does.
- **Enforcement point 2 — the guard audits at dispatch.** `pcAudit`
  (`stats/internal/guard/panelexitcontract.go`) resolves premise citations with the same checks
  demonstrates gets — shape, lexical + resolved containment, file, line, content. Tolerant:
  absence of premise lines never violates; declared-but-unresolvable joins exit 1's violation
  classes.
- **Post-fix effect — no runner change.** A fix renaming a premise makes the verification
  re-run's script fail loudly (non-zero) → the runner reads "demonstrated" → identical to the
  dispatch-time verdict → the KAN-524 ambiguity refusal (exit 2) fires → the re-author path.
- **Coverage and stated limit.** ≥1 premise line mandatory via the authoring rule for every
  newly authored or repaired reproducer; the guard stays tolerant. Residual: a post-change
  reproducer whose author skips premises entirely passes the guard silently — caught by panel
  re-runs and self-review, not mechanically. Accepted.
- **Mutation exemption.** Mutation-declared reproducers declare no premises; their instrument
  audit is the KAN-568 sha pin.

## Files touched

- `skills/flow/review-panel.md` — authoring rule, dispatch-prompt carry, re-run note, the
  guard-invocation section's audit description.
- `stats/internal/guard/panelexitcontract.go` + `check_panel_reproducer_exit_contract_test.go` — premise label in
  the audit, tolerant mode, tests; the shim `check-panel-reproducer-exit-contract.sh` header's
  instrument-audit paragraph.
- Untouched: `runreproducer.go` (the runner), `prove-reproducer.sh`,
  `check-panel-reproducers.sh` (lexical guard).

## Decisions

### Premise enforcement split: script asserts, guard audits at dispatch

**ID:** premise-enforcement-split
**Status:** active
**Chosen:** the reproducer's own body asserts its premises before its checks, and `pcAudit`
resolves premise citations at dispatch time — two points, no runner change; the post-fix rename
is covered by the existing KAN-524 ambiguity refusal.
**Considered:** runner-side audit before every exec — single choke point and clearest verdicts,
but duplicates the audit the script performs and touches `runreproducer.go` and its tests;
post-fix re-run audit only — smallest diff, but the resolution decision lands in the parent's
attention, the pattern this repo replaces with guards.

### One citation form for premises

**ID:** premise-citation-form
**Status:** active
**Chosen:** `# premise: <path>:<line>:<content>` — the exact demonstrates shape and resolution
machinery; a file, a test/class name and a task id all cite their declaration site uniformly.
**Considered:** split `# premise-file:` + citation forms — lighter to author for the common
case, but two shapes to validate and two violation classes to maintain.

### Coverage: mandatory by authoring rule, tolerant in the guard

**ID:** premise-coverage-tolerant-guard
**Status:** active
**Chosen:** every newly authored or repaired runnable non-mutation reproducer carries at least
one premise line; the guard validates only what is declared, so records predating this change
are never re-bounced.
**Considered:** strict from this change on — every open record's reproducer must carry premises
immediately, re-bouncing unrelated changes' panels mid-flight. The accepted residual is stated
under Mechanism.

### Mutation reproducers exempt

**ID:** premise-mutation-exempt
**Status:** active
**Chosen:** mutation-declared reproducers declare no premises; the KAN-568 sha pin is their
instrument audit.
**Considered:** demanding host-tree premises of a reproducer that builds its own mutated tree —
noise for exactly the convention KAN-568 added.

## Open questions

None.
## spectre/changes/archive/kan-839-flow-fix-reproducers-pass-vacuously-when-a-fix/narrative.md

# kan-839-flow-fix-reproducers-pass-vacuously-when-a-fix — session narrative

## 2026-09-28 — creating run

Resumed at implementation the same day the plan gated; this session implemented, reviewed and
verified. What the run actually hit, in order:

- **Kickoff (previous session)** — the worktree-location guard flagged a stray worktree
  (`.claude/worktrees/agent-ae1ef826791eba5ee`, branch fully merged, tree clean); removed it after
  verifying it disposable, keeping the branch ref.
- **zsh word-splitting bit the first task-1 commit attempt**: an unquoted `$T1FILES` reached `git
  add` as one pathspec and the commit silently didn't land; the fields guard then correctly
  refused the planning commit at HEAD. Retried with explicit paths. The repo's Bash-is-not-the-shell
  hazard in miniature.
- **Two record-grammar corrections at task close**, both transcribed before the guard re-ran: the
  plan declared the shim at its `skills/flow/scripts/` symlink path (the tracked file is the
  target, `scripts/check-panel-reproducer-exit-contract.sh`), and `**Files:**` paths must be
  backtick-quoted tokens, not comma-separated prose.
- **Task 3's RED was the plan's own snippet being wrong about `rrRun`**: the helper takes the
  runner's argv variadically, so flags passed inside the command-line string never reach
  `--pre-fix-verdict` parsing — the first run demonstrated the runner reading the renamed-premise
  re-run as *demonstrated* instead of refusing. Fixed the call shape, recorded as a dated
  Correction on the task.
- **The base moved mid-panel-entry** (5 commits on origin/main, overlapping
  `skills/flow/review-panel.md`); the operator chose rebase. Clean 7-commit replay onto
  `675a55a9`, no conflicts; the rebased branch needed `--force-with-lease` (the push contract's
  only sanctioned rewrite). The recorded `measured:` comments still name the pre-rebase base as
  the annotation ref — substantively accurate, counts identical at both commits.
- **Panel round 0 (primary+principles, one bundle)**: Minor-only. F1 (deduped
  primary+principles) — a declared-but-unasserted premise escapes both enforcement points, the
  design's accepted residual understating it; F2 — the exit-1 disposition enumeration not
  extended with the new premise class. Both deferred per the Minor-deferral default, entries in
  KNOWN-BUGS.md; no fix round, no re-run.
- **One verify flake**: `TestConcurrentAppendVersusRetirePreservesEveryEntry` (reconcile, a
  package this change never touches) failed once under full-suite load, passed 3× in isolation
  and on the one inline re-run the flake rule allows. No baseline declared; no sweep entry —
  it did not reproduce.
- Main deleted `scripts/test-check-panel-findings-closed.sh` while this change was in flight
  (its coverage moved into Go), which is why the post-rebase guard harness count is 68, not 69.

## 2026-09-28 — integrate run

- Preflight `RUN1`; unfinished-work `CLEAR`; visual-verify `OK` (no UI paths).
- The base-moved check ran against the state map's pre-rebase `4a278320` and reported `MOVED` —
  35 commits, since main gained ~30 more while this change was in panel. The sync rebase onto
  `origin/main` (now `b8faae9a`) hit exactly one conflict: `KNOWN-BUGS.md`, where main's own
  kan-842 deferrals and this change's task-1 deferral appended at the same tail — resolved as a
  union, both sides kept, `--continue` clean.
- Per the resolution rule, the full lint and test lists ran on the rebased tree: the lint list
  shrank by one on main (main deleted `check-contract-budget.sh` and its declaration — the
  KNOWN-BUGS budget ratchet is gone), 58 guard harnesses pass (main's Go-port consolidation),
  `go test ./... -race` 21/21 packages, SPA 170/170.
- Scoped re-verification found no discoverable `scripts/test-<name>.sh` harness for any overlap
  path — main moved that coverage into the Go suite, which the test list above already covers.
- Route: `merge and push`, taken from the project's configured default, not asked.
## git log --stat

commit 44b9e3bc3b7100dc3802ba8ae635dcd66c218de1
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 28 02:30:37 2026 +0300

    chore(spectre): plan

 .../narrative.md                                         | 16 ++++++++++++++++
 1 file changed, 16 insertions(+)

commit 137028d5ba5626b8ae5420dfc86514f1c5881a98
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 28 02:33:18 2026 +0300

    chore(spectre): archive kan-839-flow-fix-reproducers-pass-vacuously-when-a-fix

 .../design.md                                      |  0
 .../ledger.md                                      | 71 ++++++++++++++++++++++
 .../narrative.md                                   |  0
 .../panel.md                                       | 27 ++++++++
 .../proposal.md                                    |  0
 .../tasks.md                                       |  0
 6 files changed, 98 insertions(+)

## Session narrative

Run 2 chained from run 1's merge-and-push in one invocation. The integrate rebase hit one
conflict — KNOWN-BUGS.md, main's own kan-842 deferrals and this change's task-1 deferral
appending at the same tail — resolved as a union; the resolution rule then required the full
lint and test lists on the rebased tree, all green (58 guard harnesses after main's Go-port
consolidation, 21 Go packages, 170 SPA tests). The landing route was the project's configured
`merge and push`, taken without asking; the merge landed locally (`b68ba9ff`) and run 2 archived,
preserved the store renders into the archive commit, and cleaned the apply worktree, both branches
and the workspace on first pass — `COMPLETE` with an empty survivor report.
