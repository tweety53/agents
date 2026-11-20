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
> `--trailer "Flow-Key: full-suite-fix-<n>"` and no `Task-Id:`. When the line names MUTATION PROOF, follow
> that paragraph below in its exact words before you end your turn.
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

> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before you
> end your turn — not only the test cases this round adds. The fix commit lands before the first
> flip — a flip restored with `git checkout --` reverts the file to its index state, so an unstaged
> fix dies with the first restore — and a flip's killing-test failure is trusted as proof only once
> the mutation is asserted landed: the mutated bytes, read back from disk, are the ones the flip
> wrote. Mutate the mechanism: revert it in a
> scratch tree, or flip the single value it turns on — but before the tests run, confirm the edit
> landed: the target changed where you intended, not nowhere and not somewhere else. An edit that
> never applied is a refusal, not a surviving mutant — redo it with a working mechanism; it never
> buys a test. Then confirm an existing test fails, and restore. Where your fix adds or
> strengthens a test, the mechanism to mutate is the fixed line itself — flip the line your fix
> changed, the code the added or strengthened test exists to catch — and the test confirmed to
> fail under the flip is that added or strengthened test: one flip per fixed finding, the proof
> that your fix's own test tests what it claims.
> `flow-guard break-and-prove [--clean <command>] <file> (--sed <expr> | --patch <patch>) -- <test-command>` mechanizes a
> single-file flip: it mutates `<file>`, asserts the test fails, restores `<file>` from a
> pre-mutation byte snapshot rather than `git checkout --`, so uncommitted edits on it survive, and
> asserts the restored run passes; a flip across several files or harnesses is one `break-and-prove --patch` call per file and test command. `<test-command>` runs from the repository root, so a runner
> rooted in a subdirectory names it with its own directory flag — `go -C stats test ./internal/guard` in a
> repository whose `go.mod` sits in `<project>/stats/`. A runner with no such flag runs through a
> shell that enters its directory first — `-- sh -c 'cd stats/web && exec npx vitest run <file>'`
> for a vitest suite under `<project>/stats/web/`, since a vitest started at the root finds no config, runs
> outside its jsdom environment and writes its cache into an untracked `<project>/node_modules/` that the
> post-restore check refuses as drift. Each mutation alters one mechanism — where a single
> revert would also change state a second check reads, split it into surgical mutations, one per
> mechanism. A mutation no test catches is a surviving mutant: add the test that catches it before
> your turn ends. Where your fix changed a guard script — a check that exists to fail on a defect
> class — the mechanism to mutate is the fix itself: run the guard against the defect state with
> the fix reverted, where it must report the defect, and with the fix applied, where it must not;
> the line records the measured pre/post observable, not a bare test name. Record one
> `fix-mutation:` line per behaviour in your report, plus a
> `fix-mutations-total:` count, in exactly this shape: `fix-mutation: <path> — <what was mutated> —
> <the test that failed>`; `fix-mutation: <path> — none — <reason>` for a behaviour you exempt; for
> a guard-script fix the third field is the measured `<pre>→<post> <what the observable counts>`
> (`2→0 orphaned temp lists`); and one `fix-mutations-total: <n>` line after them. Where
> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather
> than deciding it yourself.

> **FIX REPORT:** your report — the REPORT FILE your prompt names, opening with the summary head
> of `implementer-brief.md` — names each finding you addressed, the executable
> behaviours your fix changed, the `fix-mutation:` and `fix-mutations-total:` lines this round's
> contract requires. Each fixed finding carries its
> proof, each run's command and the tail of its output: the test run before the fix, the same
> run after it, and the failure that run shows with the fix reverted — the re-running slot checks
> this proof rather than re-deriving it. Each fixed Critical or Important finding also carries one
> line, `regression-test: F<n> — <repo-relative path> — <test name>`, naming the git-tracked suite
> test that fails at the pre-fix code, or `regression-test: F<n> — none — <reason>` where no suite test can exist.
> The dispatcher waits on that file's presence. Your final message repeats the report's
> `fix-mutation:` and `fix-mutations-total:` lines.
