# Review panel — kan-173-make-the-daemon-s-dependencies-required-so-a

## Pass 0 — the resolved roster

**Roster resolved from the settings store:** `primary`, `principles`, `code-review-low`, `bugbot`.
**No addition this round — the resolved list ran alone.** No operator instruction named a slot the
resolved list did not carry, at this stage's start or since.

**Diff size:** `check-panel-diff-size.sh` exited 0, `under cap`. The panel diff
(`final-review.diff`, `git diff 520e9f6`) measured 1228 lines. No operator question was reached.

**Diff read by every diff-reading slot:** `.superpowers/sdd/final-review.diff` — the whole branch
against merge base `520e9f6`, covering `a26701c`, `5503ad9` and `4da13f7`.

| Slot | Spawned as | Model | Result |
|---|---|---|---|
| Primary | `superpowers:requesting-code-review` | `sonnet` | clean |
| Principles | general-purpose + `principles-reviewer-prompt.md` | `sonnet` | principles-compliant; F1 (Minor) |
| Code review (low) | general-purpose, high-confidence only | `sonnet` | clean |
| Bugbot | **general-purpose substitute** | `sonnet` | 26 mutations, no surviving mutant |

**Bugbot was substituted, not skipped.** The `bugbot` agent type is not offered by this harness
(Claude Code), so the slot ran as a general-purpose subagent carrying Bugbot's brief plus the
mandatory mutation-testing requirement. Its `-model` records `sonnet`, the model actually given,
rather than `unknown (agent-defined)` — that value is correct only for a slot spawned by its own
`subagent_type` with its own agent definition, which a substitute is not.

**Standards passed to the principles slot:** `CLAUDE.md` and `AGENTS.md`, resolved from
`.flow/project.md`'s `## standards`. Principles path:
`~/.claude/skills/flow/engineering-principles.md`.

### Concurrency interference, and the exclusive re-run

All four pass-0 slots were dispatched concurrently into one shared worktree. One of them — the
Bugbot substitute — mutates code by design. Two independent observations recorded it:

- The Principles slot, at its own start, found and reverted an uncommitted mutation
  (`NewDispatchAttributor(nil)` in `stats/internal/harvest/watcher.go`) and removed a
  `cmd/flowd/main.go.bk` backup, believing both to be unrelated leftovers. They were the Bugbot
  substitute's live mutation test and its restore backup.
- The Code review (low) slot observed intermittent full-suite nil-pointer panics in
  `DispatchAttributor.Attribute` that never reproduced in isolation, and attributed them to sandbox
  contention. They match that same live mutation exactly. It correctly declined to raise them as
  findings.

A mutation measurement taken while another process is restoring the file under test is not a
measurement, in either direction. **The Bugbot slot was therefore re-dispatched exclusively**, with
the worktree clean at `4da13f7` and nothing else running against it
(`panel-report-0-bugbot-rerun.md`). Its result supersedes the first pass's and agrees with it in
substance: 15 mutation targets, 13 killed the suite, 2 confirmed dead-or-self-referential rather
than surviving mutants, no finding.

The three reading-based slots' conclusions are unaffected — each read the static
`final-review.diff`, not the live tree — and their verdicts stand as recorded.

**This is a dispatch defect in how this run drove the panel, not a defect in the branch.** Recorded
here rather than smoothed over, because the panel record is the only document that says what
actually reviewed this branch.

### F1 — the one finding

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles | Minor | `stats/cmd/flowd/wiring_test.go:201` | design.md's wiring-test-asserts-real-store decision records deleting the test and a broader reflection test as the alternatives considered, but not the narrower one actually foregone — a single exported Deps() accessor letting the test do a plain type assertion and drop reflect entirely. |

**F1's reproducer was refused, and is recorded unverifiable.** The Principles slot supplied
`cd stats && go test ./cmd/flowd/ -run TestDaemonWiresTheRealStore -v`.
`check-panel-reproducers.sh` exited 1: the command carries a shell metacharacter, and a runnable
reproducer is a bare path optionally followed by plain arguments, never a shell command line. Per
the reproducer contract this is a refusal, not something to rewrite — the line was left exactly as
the slot wrote it, recorded unverifiable, and put to the operator together with the finding itself.
No `run-reproducer.sh` invocation was made for F1.

**F1's resolution was an operator decision, because it was a design question rather than a defect.**
Put to the operator with three options — record the foregone alternative and keep reflection; take
the accessor route instead; or withdraw F1. The operator chose to **record the alternative and keep
reflection**. `design.md`'s `wiring-test-asserts-real-store` decision now carries the
`func (w *Watcher) Deps() Deps` accessor under `**Considered:**`, with the reason it was ruled out
(it reopens the public surface the change shrinks by deleting three `Has*` accessors) and a note
that the `IsValid` guard closes the reflection route's own failure mode.

**No code changed, so no fix round was dispatched and no commit was made.** The fix is an edit to a
planning artifact — `spectre/changes/`, which no task commit ever carries.

```text
fix-mutation: spectre/changes/kan-173-make-the-daemon-s-dependencies-required-so-a/design.md — none — the fix changes no executable behaviour; it records an alternative under a decision's Considered list
fix-mutations-total: 0
```

The rule that a fix's diff must touch a path the finding named does not bite here: F1's location is
where the *technique* it comments on lives, while the defect it names is in the decision record.
Closing it by editing `wiring_test.go` would have meant changing code the operator explicitly chose
not to change.

### Independently settled during this panel

Both Bugbot passes established that the `IsNil()` check in `TestDaemonWiresTheRealStore` is
unreachable: `NewWatcher` panics on a true-nil `deps` before a `*Watcher` carrying one can exist,
and the test's typed-nil `*store.Store` reads as non-nil to both `==` and `reflect.Value.IsNil`
(confirmed by the re-run with a standalone reproducer, and by mutation — removing the check kills no
test). Neither slot raised it as a finding. Put to the operator, who chose to **leave it**: two
harmless defensive lines in a test, and removing them is scope this change was not asked for. The
`IsValid` guard above it is load-bearing and was mutation-proved to fail on a field rename.

## Verification run by the slots

`go build ./...`, `go vet ./...`, `gofmt -l .` — clean. `go test ./... -race -count=1` — all 17
packages green, run fresh at the end of the exclusive Bugbot re-run with the worktree held
exclusively throughout.
