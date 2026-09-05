# kan-443-redundant-verifier-dispatches-ran-after-an

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** make the rendered SDD ledger name each dispatch's `dispatch_key`, so three `Role:
verifier` rows read as `verify-<worktree>`, `verify-<worktree>` and `visual-verify-<worktree>`
rather than as three of the same thing — the misread behind KAN-443, closed in `proposal.md`.

**Architecture:** one conditional `Fprintf` in `RenderLedger`
(`stats/internal/records/render.go`), placed after the `Slot` line and before `Model`, in the
shape the `Slot` and `Diff base` conditionals already use; one Go test beside
`TestRenderLedgerNamesDiffBase`. No verifier dispatch rule, skill file or spec changes.

**Tech Stack:** Go standard library (`fmt`, `strings`, `testing`).

**Spec:** `spectre/changes/kan-443-redundant-verifier-dispatches-ran-after-an/design.md`

## Global Constraints

- `records.Dispatch.Key` already exists (`stats/internal/records/types.go` line 87) and is
  already scanned from `dispatch_key` by `dispatchColumns` in `stats/internal/store/records.go`;
  no type, store, migration or API change (`ledger-key-line-over-dispatch-gate`).
- The line is conditional on a non-empty key, never `orElse`-defaulted: pre-KAN-288 rows render
  no `Key` line.
- `neutraliseMarkers` wraps the key exactly as it wraps `Slot` and `Diff base`.
- No Jira write of any kind (`evidence-in-proposal-not-jira`).
- `cd stats && gofmt -l .` prints nothing and `go vet ./...` exits 0 before the task is reported
  done; every guard under `## lint` in `.flow/project.md` exits clean.

---

- [x] 1. Render `- Key: <dispatch_key>` in the SDD ledger

**Build:** green
**Files:** `stats/internal/records/render.go`, `stats/internal/records/render_test.go`
**Tests:** `TestRenderLedgerNamesDispatchKey`
**Regression:** `TestRenderLedgerNamesDispatchKey` — reverting this task's commit removes the
`Key` conditional from `RenderLedger`, so the rendered ledger contains no `- Key:
verify-gymie-worktrees` line and the test's first assertion fails; the second assertion (no
`- Key:` line in a keyless dispatch's block) would still pass on the reverted code, which is why
the first exists.
**Baseline:** before=38 after=39
<!-- measured: cd stats && go test ./internal/records/ -count=1 -v 2>&1 | grep -c '^=== RUN' @ main ed5c016 -->
<!-- predicted: 39 after this task's commit — one new top-level test, no subtests; the same command confirms -->
**Commit:** feat(stats): name each dispatch's key in the rendered SDD ledger

  - [x] **Step 1: Write the failing test** at the end of
    `stats/internal/records/render_test.go`, after `TestRenderLedgerNamesDiffBase`:

```go unverified:confirm the exact `- Slot:` / `- Key:` adjacency string matches what render.go emits once step 2 lands
func TestRenderLedgerNamesDispatchKey(t *testing.T) {
	run := records.Run{
		Change: "demo",
		Dispatches: []records.Dispatch{
			{
				Seq: 31, Role: "verifier", Key: "verify-gymie-worktrees", Model: "sonnet",
				Outcome:   "completed",
				StartedAt: time.Date(2026, 9, 5, 9, 7, 8, 0, time.UTC),
			},
			{
				Seq: 2, Role: "implementer", TaskID: "1", Model: "sonnet",
				Outcome:   "completed",
				StartedAt: time.Date(2026, 9, 5, 9, 8, 0, 0, time.UTC),
			},
		},
	}

	out := records.RenderLedger(run)

	if !strings.Contains(out, "- Key: verify-gymie-worktrees\n") {
		t.Errorf("ledger does not name the first dispatch's key:\n%s", out)
	}
	keyless := out[strings.Index(out, "## Dispatch 2 —"):]
	if strings.Contains(keyless, "- Key:") {
		t.Errorf("a dispatch recorded without a key must render no Key line:\n%s", keyless)
	}
}
```

  - [x] **Step 2: Run it and watch it fail** — the first assertion, on the missing line:

```bash verified:command shape from the project's `## test` conventions; run at main ed5c016 against the existing suite
cd stats && go test ./internal/records/ -run TestRenderLedgerNamesDispatchKey -count=1
```

  - [x] **Step 3: Add the conditional** in `RenderLedger`
    (`stats/internal/records/render.go`), directly after the `Slot` conditional and before the
    `- Model:` line:

```go unverified:confirm placement between the Slot conditional and the Model Fprintf in RenderLedger
if strings.TrimSpace(d.Key) != "" {
	fmt.Fprintf(&b, "- Key: %s\n", neutraliseMarkers(d.Key))
}
```

  - [x] **Step 4: Run the package suite, vet and format**:

```bash verified:the project's Go check commands from CLAUDE.md
cd stats && gofmt -w ./internal/records && go test ./internal/records/ -count=1 && go vet ./... && gofmt -l .
```

  - [x] **Step 5: Run the project's guards** from the repository root — every command under
    `## lint` in `.flow/project.md` — and fix any hit before committing.

  - [x] **Step 6: Commit** with the subject in `**Commit:**` above; the commit touches only the
    two paths in `**Files:**`.
