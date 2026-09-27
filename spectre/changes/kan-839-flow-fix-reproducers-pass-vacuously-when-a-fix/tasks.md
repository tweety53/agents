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
