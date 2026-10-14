# self-review-kan-937-step-2-frontend-logic-cardio-api-models-step

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
- Every prose block below is copied verbatim, line breaks aside: this change's `verbatim-moves.txt` acknowledges each new sentence by its exact text, and a reworded sentence fails `scripts/check-verbatim-moves.sh`.
- Go in `stats/` follows the Go skills `CLAUDE.md` names; run `cd stats && gofmt -w .` before `gofmt -l .`.

- [x] 1. A fix run runs the unfinished-work guard on itself before its handoff
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Run `grep -n "check-unfinished-work" skills/flow/verify-and-handoff.md` and confirm no line runs the guard — only the `(the same member **1. Check for unfinished work** … passes check-unfinished-work)` mention under **Close the stage** — the defect.
  - [x] **Step 2: Add the paragraph.** In `skills/flow/verify-and-handoff.md`, **Write `IN_PROGRESS`**, directly after the fenced block holding `flow record dispatch end -change <name> -C <canonical-worktree> -key <key> -session-token <its token> -outcome <completed|aborted>` and before the paragraph opening "Write the state file:", insert this paragraph verbatim, a blank line either side:

```markdown verified:check-verbatim-moves.sh, check-references.sh, check-installed-citations.sh and check-markdown-integrity.py exit 0 with this paragraph applied and this change's verbatim-moves.txt present, spiked at 7ba0c7b6
**A fix run then runs the unfinished-work guard on itself**, once per worktree of the resolved
set, with the arguments **1. Check for unfinished work** (`skills/flow/integrate.md`) passes it —
`check-unfinished-work.sh <worktree> <name> <canonical-worktree>` — so an appended task left
unrecorded or unreviewed stops this run rather than the next integrate. An `OUTSTANDING:` line is
this run's own unfinished work: finish what its breakdown names — a missing dispatch or review is
run and recorded, never recorded after the fact for work that did not run — and run the guard
again; the state is written only once every worktree prints `CLEAR:`. A missing verdict line or an
exit 2 is reported and stops the run at the state it read.
```

  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-installed-citations.sh`, `scripts/check-markdown-integrity.py` — all exit 0. Commit.

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`
**Tests:** **none** — one prose paragraph; the guards in step 3 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): a fix run runs the unfinished-work guard before its handoff
**After:** none

**Decision:** fix-run-self-check

- [x] 2. Signal four requires a reviewer row begun after an appended task's implementer row
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In `stats/internal/guard/check_unfinished_work_test.go`, a scratch case with the `appended` plan and rows `[{"seq":1,"key":"task-7-implementer","role":"implementer","outcome":"completed"}]` prints `CLEAR:` — the defect: an implementer-only appended task passes.
  - [x] **Step 2: Write the failing cases.** In `TestCheckUnfinishedWork`, section `// 2c. Signal four`, replace the three CLEAR cases that follow the first OUTSTANDING case (the "with its implementer row", "taskId" and "+-joined" cases) with the block below — each now carries a later reviewer row — and add the two new cases after them:

```go verified:applied together with every Go block of this plan at 7ba0c7b6 — go vet, the guard, selfreview, store TestEndDispatch and api tests and test-integrate-change.sh all pass
	each(demo(planned(appended, `[{"seq":1,"key":"task-1-implementer","role":"implementer","outcome":"completed"},{"seq":2,"key":"task-7-implementer","role":"implementer","outcome":"completed"},{"seq":3,"key":"panel-2-primary","role":"reviewer","outcome":"clean"}]`)),
		uwVerdict("an appended task with its implementer row and a later reviewer row is CLEAR", "CLEAR:"))
	each(demo(planned(appended, `[{"seq":1,"key":"impl-7","taskId":"7","role":"implementer","outcome":"completed"},{"seq":2,"key":"panel-2-primary","role":"reviewer","outcome":"clean"}]`)),
		uwVerdict("a row whose taskId names the appended task covers it", "CLEAR:"))
	each(demo(planned(appended, `[{"seq":1,"key":"task-6+7-implementer-fix-1","role":"implementer","outcome":"completed"},{"seq":2,"key":"panel-2-primary","role":"reviewer","outcome":"clean"}]`)),
		uwVerdict("a +-joined key covers the appended task it names", "CLEAR:"))
	each(demo(planned(appended, `[{"seq":1,"key":"task-7-implementer","role":"implementer","outcome":"completed"}]`)),
		uwVerdict("an appended task with an implementer row and no reviewer row is OUTSTANDING", "OUTSTANDING:"),
		uwReason("the breakdown names the unreviewed appended task", "1 ticked appended task(s) with no reviewer row begun after its implementer row (Task 7)"))
	each(demo(planned(appended, `[{"seq":1,"key":"panel-1-primary","role":"reviewer","outcome":"clean"},{"seq":2,"key":"task-7-implementer","role":"implementer","outcome":"completed"}]`)),
		uwVerdict("a reviewer row begun before the implementer row does not review the appended task", "OUTSTANDING:"))
```

    Run `cd stats && go test ./internal/guard/ -run 'TestCheckUnfinishedWork' -count=1` and confirm the two new OUTSTANDING checks and the breakdown check FAIL.
  - [x] **Step 3: Implement.** In `stats/internal/guard/unfinishedwork.go`: add `Role string `+"`"+`json:"role"`+"`"+` to the decoded `dispatches` row struct; replace the body of `if planRC == 0 {` under the `SIGNAL FOUR` comment with the block below; rename `uwUndispatched` to `uwAppendedTicked(plan string) ([]string, bool)`, returning every ticked appended id (drop its `covered` parameter and the `!covered[m[2]]` test, and say so in its doc comment). Extend the `SIGNAL FOUR` comment with one sentence: an appended task also needs a `reviewer` row whose seq is greater than its earliest non-reviewer row's — the fix run's panel names no task, so any reviewer row of the change counts, and the earliest row is the anchor so a Minor-only round's inline `panel-fix` row never demands a re-review.

```go verified:applied together with every Go block of this plan at 7ba0c7b6 — go vet, the guard, selfreview, store TestEndDispatch and api tests and test-integrate-change.sh all pass
		if planRC == 0 {
			first := map[string]int{}
			lastReview := -1
			for _, d := range dispatches {
				if d.Role == "reviewer" {
					lastReview = max(lastReview, d.Seq)
					continue
				}
				ids := []string{d.TaskID}
				if m := uwTaskKey.FindStringSubmatch(d.Key); m != nil {
					ids = append(ids, strings.Split(m[1], "+")...)
				}
				for _, id := range ids {
					if s, seen := first[id]; id != "" && (!seen || d.Seq < s) {
						first[id] = d.Seq
					}
				}
			}
			appended, ok := uwAppendedTicked(primary)
			if !ok {
				return unreadable(primary)
			}
			var missing, unreviewed []string
			for _, id := range appended {
				if s, seen := first[id]; !seen {
					missing = append(missing, id)
				} else if lastReview <= s {
					unreviewed = append(unreviewed, id)
				}
			}
			if len(missing) > 0 {
				reasons = append(reasons, strconv.Itoa(len(missing))+" ticked appended task(s) with no dispatch row (Task "+strings.Join(missing, ", ")+") — record each task's dispatch, and its fix round's pass rows and findings, before landing")
			}
			if len(unreviewed) > 0 {
				reasons = append(reasons, strconv.Itoa(len(unreviewed))+" ticked appended task(s) with no reviewer row begun after its implementer row (Task "+strings.Join(unreviewed, ", ")+") — run its fix round's panel and record its pass rows and findings, before landing")
			}
		}
```

  - [x] **Step 4: Update the script header.** In `scripts/check-unfinished-work.sh`, after the `SIGNAL FOUR` paragraph's last line `# of their own, so they are never judged.`, append these comment lines:

```bash verified:read from scripts/check-unfinished-work.sh lines 56-61 at 7ba0c7b6, where the SIGNAL FOUR paragraph ends
# A named appended task also needs a `reviewer` row whose seq is greater than
# the earliest non-reviewer row naming it: the fix run's panel reviews it
# under keys that name no task, so any reviewer row of the change counts.
```

  - [x] **Step 5: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCheckUnfinishedWork' -count=1 -race`, then `scripts/test-integrate-change.sh` and `scripts/check-guard-symlinks.sh` — all exit 0 with `gofmt -l` printing nothing. Commit.

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`, `scripts/check-unfinished-work.sh`
**Tests:** `an appended task with an implementer row and no reviewer row is OUTSTANDING`, `the breakdown names the unreviewed appended task`, `a reviewer row begun before the implementer row does not review the appended task`
**Regression:** reverting the commit returns CLEAR for an appended task with an implementer row and no later reviewer row — all three checks fail.
**Baseline:** before=61 after=64
<!-- measured: grep -cE 'uw(Verdict|Reason)\("' stats/internal/guard/check_unfinished_work_test.go @ 7ba0c7b6 -->
**Commit:** feat(guard): signal four requires a reviewer row after an appended task's implementer row
**After:** none

**Decision:** reviewer-after-implementer

- [x] 3. State the literal `## Merge order` entry syntax
  - [x] **Step 1: Reproduce before touching anything; do not guess.** In a scratch directory, `git init`, `spectre init`, write `spectre/peers` as `peer ../peer`, `spectre new demo`, write `spectre/changes/demo/link.md` as `## Merge order` / blank / `1. .` / `2. peer`, run `spectre validate demo` and confirm it reports `"1. ." is not a numbered "." or peer entry`; rewrite the entries backticked and confirm that finding is gone — the syntax the doc must state.
  - [x] **Step 2: Add the paragraph.** In `skills/flow/cross-repo-worktrees.md`, after the paragraph ending "A change with one repository runs nothing here.", insert verbatim, a blank line either side:

```markdown verified:check-verbatim-moves.sh, check-references.sh, check-installed-citations.sh and check-markdown-integrity.py exit 0 with this block applied and this change's verbatim-moves.txt present, spiked at 7ba0c7b6; spectre validate refuses `1. .` and accepts the backticked form, spiked the same day
**Each `## Merge order` entry is a numbered line naming its repository in backticks** — `spectre
validate` refuses a bare `1. .` — so a change spanning this repository and one peer writes, per
**Finish contract** (`skills/flow-contracts/finish-contract-run1.md`):
```

    then a blank line and the example as a `~~~markdown` fenced block whose lines are, in order: `## Merge order` at column 0, a blank line, ``1. `.` ``, ``2. `<peer>` `` — the fence written with tildes so it nests in no other fence.

  - [x] **Step 3: Verify.** `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-installed-citations.sh`, `scripts/check-markdown-integrity.py` — all exit 0. Commit.

**Build:** green
**Files:** `skills/flow/cross-repo-worktrees.md`
**Tests:** **none** — one prose paragraph and its example; the guards in step 3 cover it.
**Regression:** none — the task declares no tests
**Baseline:** before=0 after=0
**Commit:** docs(flow): state the literal merge-order entry syntax spectre validate accepts
**After:** none

**Decision:** merge-order-literal

- [x] 4. The self-review bundle logs every merge-order peer's change commits
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Write the test of step 2 first and run it: the peer's commit is missing from the `## git log --stat` section — the defect.
  - [x] **Step 2: Write the failing test.** Add to `stats/internal/selfreview/bundle_test.go`:

```go verified:applied together with every Go block of this plan at 7ba0c7b6 — go vet, the guard, selfreview, store TestEndDispatch and api tests and test-integrate-change.sh all pass
// TestBundleAssemblyLogsEveryMergeOrderPeer pins a cross-repo change's
// satellite commits in the git-log source: each peer the archived link.md's
// ## Merge order names, resolved through spectre/peers relative to the
// canonical repository, contributes its spectre/<name> commits.
func TestBundleAssemblyLogsEveryMergeOrderPeer(t *testing.T) {
	repo := gitRepo(t)
	peer := gitRepo(t)
	runGit(t, peer, "update-ref", "refs/remotes/origin/main", "main")
	runGit(t, peer, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	runGit(t, peer, "checkout", "-b", "spectre/demo")
	peerSHA := commitAll(t, peer, "app.go", "package main\n", "feat(app): the satellite's half")
	runGit(t, peer, "checkout", "main")
	rel, err := filepath.Rel(repo, peer)
	if err != nil {
		t.Fatal(err)
	}
	commitAll(t, repo, "spectre/peers", "front "+rel+"\n", "chore: declare the peer")
	writeArchiveBranch(t, repo, "demo", map[string]string{
		"tasks.md": "# demo tasks\n",
		"link.md":  "## Merge order\n\n1. `.`\n2. `front`\n",
	})

	bundle, err := Bundle("demo", records.Run{Change: "demo"}, "", false, false, []string{repo}, ExecRunner{})
	if err != nil {
		t.Fatalf("Bundle: %v", err)
	}
	gitLog := bundle[strings.Index(bundle, "## git log --stat"):]
	if !strings.Contains(gitLog, "### front") || !strings.Contains(gitLog, "commit "+peerSHA) {
		t.Errorf("git log section missing the peer's change commit %s under ### front:\n%s", peerSHA, gitLog)
	}
}
```

    Run `cd stats && go test ./internal/selfreview/ -run 'TestBundleAssemblyLogsEveryMergeOrderPeer' -count=1` and confirm it FAILS.
  - [x] **Step 3: Implement.** In `stats/internal/selfreview/git.go` add `path/filepath` to the imports, add the function and regexp below after `changeBranchLog`, and in `Bundle` change `gitLog = gitLogSection(g, repo, deriveFinishCommits(g, repo, rev, change))` to `gitLog = gitLogSection(g, repo, deriveFinishCommits(g, repo, rev, change)) + peerLogs(g, repo, rev, change)`. Add one sentence to `Bundle`'s doc comment: the git-log source also carries each merge-order peer's change-branch log under `### <peer>`.

```go verified:applied together with every Go block of this plan at 7ba0c7b6 — go vet, the guard, selfreview, store TestEndDispatch and api tests and test-integrate-change.sh all pass
// mergeOrderEntry is one `## Merge order` line of a link.md: a numbered
// entry naming `.` or a peer in backticks — the syntax spectre validate
// accepts.
var mergeOrderEntry = regexp.MustCompile("^[0-9]+\\. `([^`]+)`$")

// peerLogs renders each satellite repository's change-branch log — a
// cross-repo change's implementation commits outside repo — under a
// `### <peer>` heading, in the archived link.md's `## Merge order`. Peers
// resolve through rev's spectre/peers (`<name> <path>`, a relative path
// joined to repo). A change with no link.md, a peer the file does not
// declare and a branch that resolves nothing all contribute nothing: a
// missing source is never fatal.
func peerLogs(g Runner, repo, rev, name string) string {
	link, err := show(g, repo, rev, archiveDir+"/"+name+"/link.md")
	if err != nil {
		return ""
	}
	peersFile, _ := show(g, repo, rev, "spectre/peers")
	paths := map[string]string{}
	for l := range strings.Lines(string(peersFile)) {
		if f := strings.Fields(l); len(f) == 2 {
			paths[f[0]] = f[1]
		}
	}
	var b strings.Builder
	inOrder := false
	for l := range strings.Lines(string(link)) {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "## ") {
			inOrder = l == "## Merge order"
			continue
		}
		m := mergeOrderEntry.FindStringSubmatch(l)
		if !inOrder || m == nil || m[1] == "." {
			continue
		}
		p, ok := paths[m[1]]
		if !ok {
			continue
		}
		if !filepath.IsAbs(p) {
			p = filepath.Join(repo, p)
		}
		if log := changeBranchLog(g, p, changeBranchPrefix+name); log != "" {
			fmt.Fprintf(&b, "### %s\n\n%s\n\n", m[1], log)
		}
	}
	return b.String()
}
```

  - [x] **Step 4: Update `skills/flow-contracts/finish-contract-run1.md`.** In **Run the self-review pass**, replace the text `` `git log --stat` of every implementation, planning and
archive commit of the change, oldest first — and prints it`` with exactly:

```markdown verified:check-verbatim-moves.sh exits 0 with this replacement applied and this change's verbatim-moves.txt present, spiked at 7ba0c7b6
`git log --stat` of every implementation, planning and
archive commit of the change, oldest first, followed by each peer the archived `link.md`'s
`## Merge order` names, its `spectre/<name>` commits against its base under a `### <peer>`
heading — and prints it
```

  - [x] **Step 5: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/selfreview/ && go test ./internal/selfreview/ -count=1 -race`, then `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py` — all exit 0 with `gofmt -l` printing nothing. Commit.

**Build:** green
**Files:** `stats/internal/selfreview/git.go`, `stats/internal/selfreview/bundle_test.go`, `skills/flow-contracts/finish-contract-run1.md`
**Tests:** `TestBundleAssemblyLogsEveryMergeOrderPeer`
**Regression:** `TestBundleAssemblyLogsEveryMergeOrderPeer` — reverting the commit drops the peer's section, so the git-log section lacks `### front` and the peer commit.
**Baseline:** before=25 after=26
<!-- measured: grep -c '^func Test' stats/internal/selfreview/bundle_test.go @ 7ba0c7b6 -->
**Commit:** feat(selfreview): log every merge-order peer's change commits in the bundle
**After:** none

**Decision:** peer-logs-in-git-log

- [x] 5. Refuse a zero-width inline dispatch row
  - [x] **Step 1: Reproduce before touching anything; do not guess.** Write the test of step 2 first and run it: `EndDispatch` closes the inline row one second after its begin with no error — the defect.
  - [x] **Step 2: Write the failing test.** Add to `stats/internal/store/records_test.go`, after `TestEndDispatchRejectsInvalidAgentID`:

```go verified:applied together with every Go block of this plan at 7ba0c7b6 — go vet, the guard, selfreview, store TestEndDispatch and api tests and test-integrate-change.sh all pass
// TestEndDispatchRefusesAZeroWidthInlineRow pins the inline window floor:
// an inline row closed seconds after its begin, with no token report, is
// refused and stays open; closed a minute later it closes.
func TestEndDispatchRefusesAZeroWidthInlineRow(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-record-end-inline-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	begin := baseDispatch("implementer", "opus")
	begin.Key = "task-1-implementer"
	begin.SessionToken = "mf-end-inline"
	begin.AgentID = "inline"
	if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", begin); err != nil {
		t.Fatalf("RecordDispatch: %v", err)
	}
	end := records.DispatchEnd{
		SessionToken: "mf-end-inline",
		Key:          "task-1-implementer",
		Outcome:      "completed",
		EndedAt:      begin.StartedAt.Add(time.Second),
	}
	if _, err := st.EndDispatch(ctx, projectKey, "kan-1", end); !errors.Is(err, store.ErrInlineWindowTooShort) {
		t.Fatalf("EndDispatch one second after an inline begin: err = %v, want store.ErrInlineWindowTooShort", err)
	}
	end.EndedAt = begin.StartedAt.Add(time.Minute)
	if _, err := st.EndDispatch(ctx, projectKey, "kan-1", end); err != nil {
		t.Fatalf("EndDispatch a minute after an inline begin: %v", err)
	}
}
```

    Run `cd stats && go test ./internal/store/ -run 'TestEndDispatchRefusesAZeroWidthInlineRow' -count=1` and confirm it FAILS (it fails to compile until the sentinel exists — that is the red state).
  - [x] **Step 3: Implement the refusal.** In `stats/internal/store/records.go`, beside `ErrAgentIDInvalid`, add the sentinel and floor; in `EndDispatch`, directly after the `validateAgentID` check, add the window check. Add one paragraph to `EndDispatch`'s doc comment: an inline row with no token report closing under `minInlineWindow` after its begin is refused — begin and end recorded together after the work leave the harvester a zero-width window to credit.

```go verified:applied together with every Go block of this plan at 7ba0c7b6 — go vet, the guard, selfreview, store TestEndDispatch and api tests and test-integrate-change.sh all pass
// ErrInlineWindowTooShort is returned by EndDispatch when an inline row
// (agent id "inline") with no token report would close under
// minInlineWindow after its begin: a begin and end recorded together,
// after the work, leave the harvester a zero-width window and the work's
// cost unmeasured.
var ErrInlineWindowTooShort = errors.New("store: inline dispatch window too short")

// minInlineWindow is the shortest window an inline row may close with.
// One Bash call lands its begin and end milliseconds apart; real inline
// work takes minutes.
const minInlineWindow = 10 * time.Second

// in EndDispatch, after the validateAgentID check:
	if in.Tokens == nil {
		var agentID *string
		var startedAt time.Time
		err := s.pool.QueryRow(ctx, `
			SELECT d.agent_id, d.started_at
			FROM dispatches d JOIN changes c ON c.id = d.change_id
			WHERE c.project_key = $1 AND c.name = $2
			  AND d.session_token = $3 AND d.dispatch_key = $4`,
			projectKey, change, in.SessionToken, in.Key).Scan(&agentID, &startedAt)
		inline := in.AgentID == "inline" || (in.AgentID == "" && agentID != nil && *agentID == "inline")
		if err == nil && inline && in.EndedAt.Sub(startedAt) < minInlineWindow {
			return records.Dispatch{}, fmt.Errorf("%w: %s/%s key %q closed %s after its begin -- record begin when the work starts and end once its commit lands, or pass -tokens", ErrInlineWindowTooShort, projectKey, change, in.Key, in.EndedAt.Sub(startedAt))
		}
	}
```

  - [x] **Step 4: Map the refusal to 409.** In `stats/internal/api/server.go`, after the `case errors.Is(err, store.ErrAgentIDInvalid):` arm, add `case errors.Is(err, store.ErrInlineWindowTooShort):` returning `http.StatusConflict, err.Error()` — a refusal the client reads as `ErrRecordRejected`, never a 5xx it would journal and replay.
  - [x] **Step 5: Extend the rule in `skills/flow/implement.md`.** In the **Records:** bullet, replace `` An inline fix's `dispatch begin`
  is recorded when the parent starts the fix, before its first edit, and its `dispatch end` once
  the fix commit lands — never the two together, which records the fix as taking no time.`` with exactly:

```markdown verified:check-verbatim-moves.sh, check-references.sh and check-markdown-integrity.py exit 0 with this replacement applied and this change's verbatim-moves.txt present, spiked at 7ba0c7b6
Every inline row's `dispatch
  begin` is recorded when the parent starts that work, before its first edit, and its `dispatch
  end` once the work's commit lands — never the two together, which records the work as taking
  no time; `flow record dispatch end` refuses an inline row closed under ten seconds after its
  begin when it carries no `-tokens` report.
```

  - [x] **Step 6: Verify.** `cd stats && gofmt -w . && gofmt -l . && go vet ./internal/store/ ./internal/api/ && go test ./internal/store/ -run 'TestEndDispatch' -count=1 && go test ./internal/api/ -count=1`, then `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-markdown-integrity.py` — all exit 0 with `gofmt -l` printing nothing. Commit.

**Build:** green
**Files:** `stats/internal/store/records.go`, `stats/internal/store/records_test.go`, `stats/internal/api/server.go`, `skills/flow/implement.md`, `stats/internal/reconcile/reconcile.go`, `stats/internal/reconcile/record_test.go`, `spectre/changes/self-review-kan-937-step-2-frontend-logic-cardio-api-models-step/design.md`, `spectre/changes/self-review-kan-937-step-2-frontend-logic-cardio-api-models-step/verbatim-moves.txt`
**Tests:** `TestEndDispatchRefusesAZeroWidthInlineRow`
**Regression:** `TestEndDispatchRefusesAZeroWidthInlineRow` — reverting the commit removes the sentinel, so the test fails to compile; with the check alone removed, the one-second end succeeds and the test fails.
**Baseline:** before=77 after=78
<!-- measured: grep -c '^func Test' stats/internal/store/records_test.go @ 7ba0c7b6 -->
**Commit:** feat(store): refuse a zero-width inline dispatch row
**After:** none

**Decision:** inline-window-floor
