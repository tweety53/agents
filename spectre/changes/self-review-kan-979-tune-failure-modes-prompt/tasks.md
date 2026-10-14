# self-review-kan-979-tune-failure-modes-prompt

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
- Go in `stats/` follows the Go skills `CLAUDE.md` names; run `cd stats && gofmt -w .` before `gofmt -l .`.

- [ ] 1. Refuse an off-format `In-run pipeline fix:` narrative line at integrate's commit
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In a scratch copy of the `happy_path` fixture of `scripts/test-integrate-change.sh`, append `- In-run pipeline fix: the renderer could not render an exp- slot` to the change's `narrative.md` after `prepare`, run `commit`, and confirm it exits 0 and archives the line unchanged — the defect.
  - [ ] **Step 2: Write the failing case.** Add case `narrative_format` to `scripts/test-integrate-change.sh` (before `# --- happy_path ---` is fine; anywhere after `fixture()` and `to_land()` are defined), and list it in the header's `Cases:` block. Run `scripts/test-integrate-change.sh` and confirm the `narrative_format` checks FAIL.

```bash unverified:confirm the em dash survives the harness's locale; the stop line is the last stdout line
# --- narrative_format ---
fixture
sut prepare "$WORK/a" demo
[ "$RC" -eq 0 ] || halt prepare
nar="$WORK/a-worktrees/demo/spectre/changes/demo/narrative.md"
printf '\n## integrate run\n\n- In-run pipeline fix: the renderer could not render an exp- slot\n' >>"$nar"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "narrative_format: an off-format in-run fix line stops commit with exit 1" '[ "$RC" -eq 1 ]'
check "narrative_format: the stop names the off-format line" \
  'printf "%s" "$OUT" | tail -1 | grep -q "^STOP: narrative — .*the renderer could not render an exp- slot"'
printf '\n## integrate run\n\n- In-run pipeline fix: 1a2b3c4 — the renderer could not render an exp- slot (blast radius 2 files)\n' >"$nar"
sut commit "$WORK/a" demo "$WORK/a-worktrees/demo=feat(a): do the work" "$WORK/b=feat(b): do the work"
check "narrative_format: a line in the format passes commit" '[ "$RC" -eq 0 ]'
```

  - [ ] **Step 3: Implement.** In `scripts/integrate-change.sh`, at the top of `cmd_commit` right after `find_canonical || die …`, refuse an off-format line in the change's live or archived `narrative.md`; add `STOP: narrative` to the `commit` paragraph of the header comment. The format is the one **Record each one** under **Pipeline defects found mid-run** (`skills/flow-contracts/pipeline.md`) fixes: `In-run pipeline fix: <agents sha | deferred> — <the defect> (blast radius <N> files)`.

```bash unverified:confirm grep -E matches the UTF-8 em dash byte-for-byte on macOS and Linux grep
  local nf bad
  for nf in "$CANON/$SPEC/changes/$NAME/narrative.md" "$CANON/$SPEC/changes/archive/$NAME/narrative.md"; do
    [ -f "$nf" ] || continue
    bad="$(grep -n 'In-run pipeline fix:' "$nf" |
      grep -vE 'In-run pipeline fix: `?([0-9a-f]{7,40}|deferred)`? — .+ \(blast radius [0-9]+ files?\)\.?$' || true)"
    [ -z "$bad" ] || stop narrative "$nf holds in-run fix line(s) off the format 'In-run pipeline fix: <agents sha | deferred> — <the defect> (blast radius <N> files)': ${bad//$'\n'/; } — rewrite each to it, then re-run commit"
  done
```

  - [ ] **Step 4: Name the stop in `skills/flow/integrate.md`.** After the paragraph ending "**`/flow` never passes `--force`** to `spectre archive`.", add this paragraph verbatim (its citation keeps `check-verbatim-moves.sh` clean):

```markdown verified:check-verbatim-moves.sh and check-references.sh both exit 0 with this paragraph applied, spiked at acc2a4b8
**`STOP: narrative`** → an `In-run pipeline fix:` line in the narrative is off the format **Pipeline defects found mid-run** fixes for it (`skills/flow-contracts/pipeline.md`): rewrite that line to that format and re-run `commit`.
```

  - [ ] **Step 5: Verify.** `scripts/test-integrate-change.sh`, `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-verbatim-moves.sh` — all exit 0. Commit.

**Build:** green
**Files:** `scripts/integrate-change.sh`, `scripts/test-integrate-change.sh`, `skills/flow/integrate.md`
**Tests:** `narrative_format`
**Regression:** `narrative_format` — reverting the commit lets `commit` exit 0 over the off-format line, failing the exit-1 and `STOP: narrative` checks.
**Baseline:** before=17 after=18
<!-- measured: grep -c '^# --- ' scripts/test-integrate-change.sh @ acc2a4b8 -->
**Commit:** feat(integrate-change): refuse an off-format in-run pipeline fix line
**After:** none

**Decision:** narrative-check-in-commit

**Decision:** narrative-format-tolerance

- [ ] 2. Record every printed summary in the store
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -rn "record summary" skills scripts hooks rules` and confirm it prints nothing — the defect.
  - [ ] **Step 2: `skills/flow-contracts/pipeline.md`.** In **Summary and live-stack line, before every handoff**, after the paragraph ending "still running without having to ask.", add this paragraph verbatim (`flow record summary` takes `-change`, `-C` and `-file`, `-` reading stdin, per `runRecordSummary` in `stats/cmd/flow/record.go`):

```markdown verified:check-verbatim-moves.sh and check-references.sh both exit 0 with this paragraph applied, spiked at acc2a4b8
**Every summary a run prints is recorded as it prints** — `flow record summary -change <name> -C <canonical-worktree> -file -`, the summary's Markdown on stdin, the `IN_PROGRESS` handoff's `**Summary:**` bullets included; one row per change, the last write wins — so the self-review bundle's first source is the run's own statement, as **Write `IN_PROGRESS`** prints it (`skills/flow/verify-and-handoff.md`).
```

  - [ ] **Step 3: `skills/flow/verify-and-handoff.md`.** After the paragraph ending "prints as its own rule states, not this block.", add this paragraph verbatim:

```markdown verified:check-verbatim-moves.sh and check-references.sh both exit 0 with this paragraph applied, spiked at acc2a4b8
**The `**Summary:**` bullets are recorded before the block prints**, per **Summary and live-stack line, before every handoff** (`skills/flow-contracts/pipeline.md`).
```

  - [ ] **Step 4: Verify.** `scripts/check-references.sh`, `scripts/check-markdown-integrity.py`, `scripts/check-verbatim-moves.sh`, `scripts/check-installed-citations.sh` — all exit 0. Commit.

**Build:** green
**Files:** `skills/flow-contracts/pipeline.md`, `skills/flow/verify-and-handoff.md`
**Tests:** **none** — two prose sentences; the guards in step 4 cover them.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow-contracts): record every printed summary with flow record summary
**After:** none

**Decision:** summary-record-in-pipeline

- [ ] 3. Report a ticked task with no implementer dispatch row as unfinished work
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `stats/internal/guard/check_unfinished_work_test.go`, a scratch case with plan `- [x] 1. done\n- [x] 7. appended by a fix run\n` and the default dispatch rows prints `CLEAR:` — the defect (kan-979's fix-round task reached integrate with no dispatch row).
  - [ ] **Step 2: Write the failing cases.** In `TestCheckUnfinishedWork`, after the signal-three cases, add (each check call on its own line):

```go unverified:confirm each/demo/uwVerdict/uwReason keep these signatures at the task's parent
	// 2c. Signal four — a ticked task no implementer dispatch key names
	//     (kan-979's fix-round task reached integrate with no rows at all).
	each(demo(func(t *testing.T, f *uwFix) {
		f.plan(t, "- [x] 1. done\n- [x] 7. appended by a fix run\n")
	}),
		uwVerdict("a ticked task with no implementer dispatch row is OUTSTANDING", "OUTSTANDING:"),
		uwReason("the breakdown names the undispatched task", "1 ticked task(s) with no implementer dispatch row (Task 7)"))
	each(demo(func(t *testing.T, f *uwFix) {
		f.plan(t, "- [x] 1. done\n- [x] 2. grouped\n- [x] 3. grouped\n")
		f.setFile(t, "dispatches.json", `[{"key":"task-1-implementer","outcome":"completed"},{"key":"task-2+3-implementer-fix-1","outcome":"completed"}]`+"\n")
	}),
		uwVerdict("a group implementer key covers every task it names", "CLEAR:"))
```

    Run `cd stats && go test ./internal/guard/ -run 'TestCheckUnfinishedWork' -count=1` and confirm the OUTSTANDING and breakdown checks FAIL.
  - [ ] **Step 3: Implement signal four** in `stats/internal/guard/unfinishedwork.go`, inside the `if anchor == worktree {` block after the unclosed-row reason, gated on `planRC == 0`; add a `SIGNAL FOUR` paragraph beside the code's other signal notes saying what it catches (a fix round recorded nothing) and why the implementer row is the proxy (pass rows have no read verb). Collect `d.Key` from the already-decoded rows.

```go unverified:confirm the variable names (dispatches, primary, planRC, reasons) at the task's parent
var (
	uwTickedTask      = regexp.MustCompile(`^- \[[xX]\] ([0-9]+)\. `)
	uwImplementerKeys = regexp.MustCompile(`^task-([0-9]+(?:\+[0-9]+)*)-implementer`)
)

// uwUndispatched returns the ids of plan's ticked column-0 tasks that no
// task-<ids>-implementer key names, in plan order.
func uwUndispatched(plan string, keys []string) ([]string, bool) {
	covered := map[string]bool{}
	for _, k := range keys {
		if m := uwImplementerKeys.FindStringSubmatch(k); m != nil {
			for _, id := range strings.Split(m[1], "+") {
				covered[id] = true
			}
		}
	}
	b, err := os.ReadFile(plan)
	if err != nil {
		return nil, false
	}
	var missing []string
	for _, l := range strings.Split(string(b), "\n") {
		if m := uwTickedTask.FindStringSubmatch(l); m != nil && !covered[m[1]] {
			missing = append(missing, m[1])
		}
	}
	return missing, true
}

// at the call site, after the unclosed-row reason:
		if planRC == 0 {
			missing, ok := uwUndispatched(primary, keys)
			if !ok {
				return unreadable(primary)
			}
			if len(missing) > 0 {
				reasons = append(reasons, strconv.Itoa(len(missing))+" ticked task(s) with no implementer dispatch row (Task "+strings.Join(missing, ", ")+") — record each task's dispatch, and its fix round's pass rows and findings, before landing")
			}
		}
```

  - [ ] **Step 4: Keep the integrate harness's fixture recorded.** `scripts/test-integrate-change.sh`'s stub `flow` answers `record dispatches` with `[]` while its plan ticks task 1, which signal four now reports; change that stub line to answer `[{"key":"task-1-implementer","outcome":"completed"}]`. Give any existing `check_unfinished_work_test.go` case whose expectation signal four changes the covering row in its fixture, never a weaker assertion.
  - [ ] **Step 5: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCheckUnfinishedWork' -count=1 -race`, then `scripts/test-integrate-change.sh` and `scripts/check-guard-symlinks.sh` — all exit 0 with `gofmt -l` printing nothing. Commit.

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`, `scripts/test-integrate-change.sh`
**Tests:** `a ticked task with no implementer dispatch row is OUTSTANDING`, `the breakdown names the undispatched task`, `a group implementer key covers every task it names`
**Regression:** reverting the commit returns CLEAR for a ticked task no row names — the first two checks fail; the third fails if group keys stop covering their ids.
**Baseline:** before=55 after=58
<!-- measured: grep -cE 'uw(Verdict|Reason)\("' stats/internal/guard/check_unfinished_work_test.go @ acc2a4b8 -->
**Commit:** feat(guard): report a ticked task with no implementer dispatch as unfinished work
**After:** Task 1

**Decision:** undispatched-task-signal

**Decision:** group-key-coverage
