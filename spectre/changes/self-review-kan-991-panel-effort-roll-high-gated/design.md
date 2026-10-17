## Context

- Five non-`big` findings from the kan-991 self-review pass, planned as one change per item 1 of
  **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`), on the change's
  own branch `spectre/kan-991-panel-effort-roll-high-gated` at `56e6aaa9`. Each is carried by
  exactly one task; none is dropped or disproved. Checked against this worktree at `56e6aaa9`:
  - A1.1 (Task 1): the archived kan-991 `tasks.md` line 13 (step 6) cites the vocabulary guard
    `52f9660b` deleted. The plan-shape logic is Python (`scripts/check-plan-shape.py`), but
    `check-plan-shape.sh`'s `check_one` already runs a Go sibling per file, `flow-guard
    plan-suite-steps` (`stats/internal/guard/plansuitesteps.go`), reusing the shared tasks.md
    grammar (`tcfIterTasks`, `tcfUnfenced`, `tcfParseTask`). The new check is a second such
    sibling, so no Python grows. Blast radius 3.
  - A1.2 (Task 2): `checkPanelDocsOnly` (`stats/internal/guard/paneldocsonly.go`) refuses any
    path `isDocPath` (`planclass.go:59`, `\.mdc?$`) rejects; `isDocPath` is shared with
    plan-class (`planclass.go:230`, `:291`), so the exemption lives in `paneldocsonly.go` alone.
    Blast radius 4.
  - A1.3 (Task 3): the decision JSON's pair-carrying fields are `implementer`, `fixer`, `groups`,
    `panel.dispatches`, `rerun_dispatch` and `extra_dispatch` (**Model and effort**,
    `skills/flow/brainstorm-planner.md`); none records the gated bundle, and
    `taskreviewersingledispatch.go` reads no effort. The prose granting it `high` is the bundle
    bullet, the "`high` reaches …" sentence and the plan-gate question in `brainstorm-planner.md`,
    and the bundle sentence in `implement.md`. `skills/flow/SKILL-rationale.md`'s kan-991 section
    records the gating's reasoning, so it is listed and updated — the rule Task 5 adds. Blast
    radius 3.
  - A1.4 (Task 4): traced in the parent's session transcript — its 13:00:54Z Bash call ran
    `for s in primary principles; do flow record pass -change $N -round 0 -note "x" >/dev/null 2>&1; done`,
    a placeholder loop body with its output silenced, landing pass rows 2165 and 2166 (`flow record
    passes -change kan-991-panel-effort-roll-high-gated`). `runRecordPass`
    (`stats/cmd/flow/record.go`) only requires `-note` non-empty. Blast radius 2.
  - A1.5 (Task 5): no section-D checklist item names rationale appendices. Blast radius 1.
- Spiked at plan time, every edit reverted: the five diffs apply in task order at `56e6aaa9`
  (`git apply`); with all applied `go test ./internal/guard/ ./cmd/flow/` passes, `gofmt -l .` and
  `go vet` are clean, `scripts/test-check-plan-shape.sh` prints `all cases passed`, and
  `check-references.sh`, `check-markdown-integrity.py`, `check-dispatch-paragraphs.sh`,
  `check-installed-citations.sh`, `check-ask-silence.sh` and `check-guard-symlinks.sh` exit 0;
  `check-verbatim-moves.sh` reports 0 violations with this change's `verbatim-moves.txt`.

## Decisions

### Script-path citations are a Go sibling of plan-suite-steps

**ID:** plan-script-paths-go-sibling
**Status:** active
**Chosen:** a new `flow-guard plan-script-paths <tasks.md>` that `check-plan-shape.sh`'s
`check_one` runs beside `plan-suite-steps`, refusing an unfenced inline-span token opening
`scripts/` that is absent from the repository; the repository is the one holding the plan's
`<spec-root>/changes/<name>/`; a path any task's `**Files:**` names passes, as does a token holding
`<`, `*`, `{` or `$`; a `:<line>` suffix and trailing punctuation are stripped; an archived plan, or a
file outside a `changes/<name>/` directory, is not checked. Measured over the 22 newest archived
plans copied into a live change directory: kan-991's real citation, plus three citations in one
plan written on 2026-09-28, the day those scripts were retired — archived, so skipped.
**Considered:** an F14 in `check-plan-shape.py` — grows guard logic in Python against the repo's
new-guard-in-Go convention; checking archived plans too — the no-argument scan reads unlanded
archives, and a path retired after a plan was written would fail history the plan got right;
any backticked path, not only `scripts/` — the finding names script citations, and other trees
(`skills/`, `stats/`) are already reached by `check-references.sh` or name files a task creates.

### verbatim-moves.txt counts as documentation in paneldocsonly.go alone

**ID:** docs-only-verbatim-moves
**Status:** active
**Chosen:** `checkPanelDocsOnly` passes a path under `<spec-root>/changes/` whose basename is
`verbatim-moves.txt`, `<spec-root>` from `specRootLeaf`; `isDocPath` unchanged.
**Considered:** widening `isDocPath` — plan-class shares it and would reclassify plans; any
`verbatim-moves.txt` anywhere — the acknowledgement file only lives under a change directory (the
`/flow-fast` copy under `.superpowers/sdd/` is never committed).

### The gated reviewer bundle never takes high

**ID:** gated-bundle-never-high
**Status:** active
**Chosen:** the gated per-task reviewer bundle runs at the panel roll's effort, never `high`; the
`high` paths of decision `panel-high-operator-gated` (kan-991) reach `panel.dispatches` entries
only, and the plan-gate question names such an entry.
**Considered:** adding a bundle pair to the decision JSON and having the dispatcher read it — a
schema and Go change for a seam the panel already covers; leaving the prose — a `high` the
operator approves would be silently dropped.

### A pass note needs two words

**ID:** pass-note-two-words
**Status:** active
**Chosen:** `flow record pass` refuses, exit 2 before the store is contacted, a `-note` with fewer
than two whitespace-separated tokens. Measured over the dev store's 2168 `panel_passes` rows: 6
one-token notes, all junk (`x` ×4, `probe` ×2), none real.
**Considered:** a denylist of `x`/`probe` — misses the next placeholder; a minimum character count
— a short real note ("docs-only reduction") and a long junk token are not told apart by length.

### Rationale appendices are listed in Files

**ID:** rationale-appendix-in-files
**Status:** active
**Chosen:** one section-D blockquote in `brainstorm-planner.md`, beside the other planning
directives.
**Considered:** a guard matching changed rules to appendix sections — no mechanical link between a
rule sentence and its rationale paragraph exists to check.

### First review effort and fixer grouping

**ID:** first-review-effort
**Status:** active
**Chosen:** first review on `flow-medium` (effort `medium`); one fixer dispatch (key
`self-review-kan-991-panel-effort-roll-high-gated-fix`) carrying Tasks 1–5 in order — each task is
a verified diff applied with `git apply`.
**Considered:** `flow-low` for the first review — Task 1 is a new guard whose tokeniser and root
derivation decide its false-positive surface, judgment a low-effort read may skim; one fixer per
task — five dispatches for five `git apply`s.

## Open questions

## Live check

Task 4 changes a CLI that writes the dev store: `flow record pass -change self-review-kan-991-panel-effort-roll-high-gated -round 0 -note x`
exits 2 naming the one-word note, and `flow record passes -change self-review-kan-991-panel-effort-roll-high-gated`
lists no row for it. Not working looks like `recorded: pass <id>` and a new `x` row.
