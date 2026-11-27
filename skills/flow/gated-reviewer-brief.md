# Gated reviewer brief

Every gated reviewer bundle of **4. Execute (SDD + TDD)** (`skills/flow/implement.md`) — an original
bundle or a fix re-review — reads this file in full first: its prompt is a stub
naming this file and one `PASS task-<n>` block per task it reviews.

> **LEAF RULES:** the TOOLS, FOREGROUND BUILDS and NO DELEGATION paragraphs of
> `skills/flow/implementer-brief.md` bind you.

> **READ-ONLY REVIEW:** You review a tree other agents are working in — never mutate it. No
> command that writes: no `git checkout`, `git restore`, `git reset`, `git stash`, `git clean`,
> `git worktree add`, no commit, no index change, and no file edit outside your own report file —
> the panel's mutating slots are the one declared exception, and they work in throwaway copies the
> parent creates, not this tree. Inspect with Read, Grep, and the read-only git forms — `git show`, `git diff`,
> `git log`, `git status`. A mutation you cause is indistinguishable from a defect the next
> implementer inherits, and a restore you perform is a claim nobody can check. A test or probe
> you run is kept bounded (PROOF RUNS, `skills/flow/implementer-brief.md`).

> **GO CHANGES:** When a reviewed diff includes a `*.go` file, read and follow the `go-code-review` skill
> first. A diff with no `.go` file skips this.

> **FRAME DEPARTURES:** a part of a capture that departs from the mockup frame or approved
> prototype it is compared with is an Important finding, unless a `## Decisions` entry of the
> change or a sidecar `#` comment naming the sibling ticket that builds that part excuses it.
> Its fix changes the app until the capture matches the frame, never a sidecar comment that
> names the departure.

> **INDEPENDENT PASSES:** each pass reviews its own task's diff range and the code, never an
> earlier pass's report or conclusions; a defect that sits in two passes' ranges is raised under
> each. Write each pass's report file before beginning the next pass.

> **OPEN GAPS:** Read `## Open gaps` of the
> `<abs-worktree>/.superpowers/sdd/implementer-report-*.md` whose `commit:` names the pass's task.
> Each gap that **CONFIDENCE ONCE GREEN** (`skills/flow/implementer-brief.md`) says is never a gap
> is a finding of that pass: Important, or Minor for a later task's work.

and, for each `PASS task-<n>` block of your prompt:

> **REPORT FILE:** write this pass's report to
> `<abs-worktree>/.superpowers/sdd/reviewer-report-task-<n>.md` before beginning the next pass,
> the bundle's last one as your **last** act — a `## Verdict` section carrying exactly `clean` or
> `fix`, and each finding with its file, line and severity — Critical, Important or Minor as
> `skills/flow/reviewer-calibration.md` defines them. The verdict is `fix`
> only when a Critical or Important finding stands; Minor-only findings are `clean`. A fix round's re-review writes
> `reviewer-report-task-<n>-fix-<k>.md`. Your report files are every pass's report and the bundle's
> findings file, the FINDINGS FILE path your prompt names: one JSON object per line for every
> finding of every pass, carrying `task` (the pass's task id), `severity`, `location`
> (`<file>:<line>`) and `note`, and never a `ref` — written empty when no pass raised one, beside
> the last pass's report. Your final message is one line per pass, in plan order:
> `task <n>: clean|fix — Critical <n> Important <n> Minor <n> — <its report path>`.
