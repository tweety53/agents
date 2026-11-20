# Fixer brief

Every fix dispatch — a gated fix, a re-commit, a full-suite fix, a review panel's fix chunk, a
Minors fix, a verify fix — is the implementer stub (**4. Execute (SDD + TDD)**,
`skills/flow/implement.md`) carrying a `FIX INPUT:` line and the line `Fix brief:` naming this file.
It reads `implementer-brief.md` and then this file in full first; every paragraph below binds it.

> **FIX DISPATCH:** a prompt carrying a `FIX INPUT:` line makes you a fixer, not a task's
> implementer: fix exactly what that line names — reviewer reports and finding refs, a
> `flow-guard close-task` refusal's lines, or a failing suite block —
> at its root cause, under every other paragraph of this brief save the commit subject and trailer
> FLOW — COMMIT-PER-TASK names, which this paragraph's take the place of. A finding's fix is one
> commit per task at the branch tip,
> `git -C <abs-worktree> commit -m "fix(<module>): <subject>" --trailer "Task-Id: <n>" --trailer "Flow-Key: <your key>" -- <this task's files>`,
> `<n>` that task's id — both trailers as `--trailer` flags, never a `-m` paragraph each, since git
> reads trailers from the last paragraph alone; a re-commit rewrites only the
> refused, unpushed task commit, keeping its declared
> `**Commit:**` subject and `Task-Id:` trailer; a full-suite fix commits with
> `--trailer "Flow-Key: full-suite-fix-<n>"` and no `Task-Id:`.
> Whether a finding stands is the dispatcher's call, never yours: one you cannot fix is a
> `not-fixed:` line, never a silent skip.

> **VERBATIM REPORT — THE FACT:** each finding below names the file its slot's report was written
> to. That file is the reviewer's own report, unedited — read it before you act on the finding.
> The structured block is the dispatcher's summary of it: direction on what to work on, and
> never a source of fact. Where the two disagree the report wins. Where the block asserts
> something the report does not, treat it as unchecked and establish it yourself before building
> on it.

> **FINDINGS ARE INPUT:** a finding names a defect and suggests a route to fixing it — findings
> are input, not orders. Your obligation is to resolve the defect the finding names; the
> suggested route is the raising slot's proposal, never a binding instruction. Where the literal
> route would break something the code already requires — a required ordering, a declared
> contract, a stated invariant — take the route that resolves the defect without the damage, and
> your report records the deviation and justifies it: what the literal route would have broken,
> and how yours resolves the defect. A silent deviation is an unfixed finding, and so is a
> literal compliance that leaves the defect standing.

> **ROUND SCOPE:** your round's output is the diff and the report, nothing else. Do not rewrite
> `proposal.md` or `design.md`: unaffected sections are never restated, and a decision your fix
> genuinely overturns is overridden by one appended entry under `## Decisions`, never a per-round
> rewrite of the file. Your report names what you fixed and the behaviours you changed,
> then stops.

> **TARGETED TESTS:** Run only the tests the `**Tests:**` fields of the tasks your fix touches
> name, and the tests your fix adds, through the build
> tool's own selector — `--tests '<class>'` for Gradle, `-run '<name>'` for `go test`, `-t
> '<name>'` for vitest — once for RED, once for GREEN, and again only after a source edit. Never
> run the module or repository suite mid-task: the `## test` list runs again in
> `flow.verify`. Pipe a test run's output through `tail` so a green
> run costs lines of context, not a build log. Keep every run bounded (PROOF RUNS,
> `skills/flow/implementer-brief.md`).

> **OUTPUT BUDGET:** Every tool result stays in your context, and every later turn re-reads your
> whole context — a large output is paid for again on every turn after it. Read a file over 200
> lines by line range — `grep -n` for the symbol, then `sed -n '<a>,<b>p'` or Read with
> `offset`/`limit` — and never re-read a file already in your context unless you have edited it
> since. Cap every search (`| head -40`) and every build, lint or install run (`| tail -30`), and
> reproduce a failing block from its log rather than printing the whole log. Never print a
> generated file — a lockfile, a snapshot, a bundle, a build artifact.

> **PIXEL PROBE:** every fix you land to draw/geometry code — code that computes what is
> drawn: extents, bounds, gridlines, offsets, paths, positions, sizes — carries a probe assertion
> against the actual rendered pixels: drive the real draw path and assert on the value it renders or computes, never on
> a hand-derived copy of it. A fix commit for this class of bug that lands no such probe is
> rejected at the fix step — the finding stays open and the round does not close — so the
> regression is caught this round, not discovered by the next one. Where pixels cannot be
> rendered in this environment, the probe asserts on the geometry the draw path actually
> computes — the same numbers the renderer will paint — and your report names why the pixel
> output itself was not asserted.

> **FIX REPORT:** your report — the REPORT FILE your prompt names, opening with the summary head
> of `implementer-brief.md` — names each finding you addressed and the executable
> behaviours your fix changed. Each fixed finding carries its
> proof, each run's command and the tail of its output: the test run before the fix, the same
> run after it, and the failure that run shows with the fix reverted — `flow-guard break-and-prove
> [--clean <command>] <file> (--sed <expr> | --patch <patch>) -- <test-command>` mechanizes one single-file flip,
> restoring from a byte snapshot so uncommitted edits survive; `<test-command>` runs from the repository root — the re-running slot checks
> this proof rather than re-deriving it. Each fixed Critical or Important finding also carries one
> line, `regression-test: F<n> — <repo-relative path> — <test name>`, naming the git-tracked suite
> test that fails at the pre-fix code, or `regression-test: F<n> — none — <reason>` where no suite test can exist.
> The dispatcher waits on that file's presence.
