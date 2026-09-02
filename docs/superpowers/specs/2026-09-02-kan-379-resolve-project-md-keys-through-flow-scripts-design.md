# kan-379-resolve-project-md-keys-through-flow-scripts — design

## Context

KAN-379 is the mechanics-to-code half kan-378's research note split off: kan-378 deleted restated
prose and left every `/flow` run still loading three contracts at run time to read one value each.
Measured at `ed2fbf8` (main, kan-378's own `measure.sh` in `after` mode — its after-figures are this
change's before-figures): `project-configuration.md` (45252 bytes) is loaded at four points of a
creating run (`review-panel.md:21`, `verify-and-handoff.md:66`, `integrate.md:133`,
`archive.md:80`); `workspace-isolation.md` (25115 bytes) at `implement.md:99` — only to compute a
workspace id that `flow workspace-id <name>` already derives — and at `verify-and-handoff.md:48`;
`plan-provenance.md` (25221 bytes) by the planner, of which the run-applied part is 5822 bytes.
<!-- measured: wc -c skills/flow-contracts/{project-configuration,workspace-isolation,plan-provenance}.md @ ed2fbf8 -->
<!-- measured: awk per-## section byte sums over skills/flow-contracts/plan-provenance.md @ ed2fbf8 -->

Two corrections to the issue text, found while reading: the principles-slot load is already gone
(kan-378 cut the cite; `principles-reviewer-prompt.md:171-178` carries the entry-form and
containment rule itself), and the provenance tags are enforced by `check-plan-provenance.py`, not
`check-plan-shape.py` (which checks the F1–F6 task-field shape).

The operator widened scope in brainstorming: `pipeline.md`'s mechanics move into `flow` in this
change rather than a follow-up, and the never-read `## planning model` / `## self review model`
keys in `SKILL.md`'s model-resolution block are fixed here too.

No staged research note existed for this change (`docs/superpowers/research/kan-379.md`,
`kan-379-*.md` absent; the issue's `flow-context-thinning.md` was never committed).

## 1. `project-get.sh` and the shared section extractor

`scripts/lib/project-section.sh` defines `project_section <file> <key>`: prints the body of the
first `## <key>` heading in `<file>` — every line after the heading up to, not including, the next
`^## ` line or EOF, `### ` subheadings included — with leading and trailing blank lines removed and
nothing else normalised, after `strip_bom_cat` (`scripts/lib/strip-bom.sh`). Three existing inline
copies of that awk are replaced by a `source` of this file: `gather-dispatch-context.sh`'s
`extract_project_section`, `check-model-keys.sh`'s `extract_section_body` (which keeps its own
backtick-and-whitespace trim as a second stage), and the new script below. Same reasoning as
`scripts/lib/within-root.sh`'s header: one definition, correctable once.

`scripts/project-get.sh <project-root> <key>`:

- Exit `0`: the key is declared once; its body is on stdout.
- Exit `1`: `<project-root>/.flow/project.md` is absent, or declares no `## <key>` heading; one line
  on stderr says which. This is every "optional key absent → skip" case in the phase files.
- Exit `2`: usage (not exactly two arguments), `<project-root>` not a directory, or `## <key>`
  declared more than once — the same ambiguity refusal `check-visual-trigger.sh` makes.

`<key>` is the heading text after `## `, quoted when it carries spaces
(`project-get.sh <root> "default landing route"`). The script derives no repository root from
`$SCRIPT_DIR/..` (`check-guard-symlinks.sh` rule 4); it takes the root as an argument. It is
symlinked into `skills/flow/scripts/` (rule 2) and tested by `scripts/test-project-get.sh` in the
`test-check-visual-trigger.sh` shape: fixture trees under `TMPDIR`, the real script as a
subprocess, every exit code and the BOM, twice-declared, empty-body and fenced-body cases.

## 2. Repointing the load points

| File:line | Today | After |
|---|---|---|
| `review-panel.md:21` | Load `project-configuration.md`; read `## review panel citation check` | `project-get.sh <worktree> "review panel citation check"`; exit 1 is the existing skip-silently path; exit 2 is reported and skipped like the guard's own exit 2 |
| `review-panel.md:288` | `## standards` per `project-configuration.md` | `project-get.sh <worktree> standards` for the entries; the entry-form and containment rule cited from `principles-reviewer-prompt.md`'s `[STANDARDS_PATHS]` step, which already carries it |
| `verify-and-handoff.md:66` | Load `project-configuration.md`; run `## lint`/`## test` | `project-get.sh <worktree> lint` and `project-get.sh <worktree> test`; exit 1 is the existing auto-detect path |
| `verify-and-handoff.md:48` | Load `workspace-isolation.md` unconditionally | Load it only when `prepare-workspace.sh` exits non-zero, or exits 0 with stderr naming a `cache index` row — the two cases whose procedure lives there |
| `verify-and-handoff.md:57` | Fallback when the script cannot be located | Unchanged — loaded only on that path |
| `implement.md:99-104` | Load `workspace-isolation.md`; compute the id by hand | `flow workspace-id <name>`; no contract load |
| `integrate.md:133` | Load `project-configuration.md`; read `## default landing route` | `project-get.sh <main-checkout> "default landing route"`; the byte-for-byte literal match already stated inline stays |
| `archive.md:80-86` | Load `project-configuration.md`; run `remove` with the id re-derived by hand | `project-get.sh <main-checkout> "workspace isolation"`, read the `remove` row of its command table; id from `flow workspace-id <name>` |

Each "**Load `…`**" sentence is deleted, never reworded. `project-configuration.md` stays canonical
for every key's meaning; the phase files cite it by section where they already did and never load
it on the ordinary path.

## 3. `plan-provenance.md` split by audience

Same split kan-372 made for `project-configuration.md`. `plan-provenance.md` keeps what a tagger
applies: **The four tags**, **Tag syntax examples**, **The asymmetry rule**, **The implementer's
duty**, **When a measurement contradicts the plan** (5822 bytes). **The guard's scope, and why it
is narrow**, **The quotation exemption** with its four subsections, and **What the guard does not
do** (19195 bytes) move verbatim to `skills/flow-contracts/plan-provenance-guard.md`, canonical for
what `check-plan-provenance.py` enforces; the guard's own docstring cites the new file.
<!-- measured: awk per-## section byte sums over skills/flow-contracts/plan-provenance.md @ ed2fbf8 -->
Citers repointed by section: `build-green.md:69-73` (**What the guard does not do**),
`check-plan-provenance.py:73,591,904,2634`, `flow-contracts/SKILL.md`'s file index, and
`rules/flow-manual-review.mdc`'s contract table. `brainstorm-planner.md` **D** keeps its "Load
`plan-provenance.md`" line — the file it names is now the small one. Two budget rows in
`check-contract-budget.sh`: the old row lowered to landed size plus a quarter, one new row.

## 4. `pipeline.md` mechanics into `flow`

**Change name resolution → `flow state resolve [-addr url] [-timeout dur] [-C dir]`** in
`stats/cmd/flow/state.go`, beside `state list`, reusing `listStateBoard`,
`fallbackStateListRecords` and `fallback.ProjectKey` (whose second return is the main checkout).
Prints one JSON object:

```json unverified:confirm the field names against state.go once written
{"source":"store","complete":true,
 "candidates":[{"name":"…","state":"IN_PROGRESS","updatedAt":"…","updatedBy":"…"}],
 "unreadable":[]}
```

- `source:"store"`: every board row whose `state` is not `FINISHED`.
- `source:"fallback"`: the union of every readable fallback record's name and every directory name
  directly under `<main-checkout>/spectre/changes/` other than `archive`, minus any name for which
  `<main-checkout>/spectre/changes/archive/<name>/` exists; `unreadable` carries the names of
  fallback files that could not be read or parsed. Reads the changes directory directly instead of
  running `spectre list --json` — same source of truth, no subprocess (decision below).
- Never blocks: the fallback path exits 0 exactly as `state list` does; a read error on the
  changes directory is reported on stderr and the fallback records still print.

`pipeline.md`'s section shrinks to the call, "echo `source`, name every `unreadable` entry", the
three-outcome rule (one match / several → **AskUserQuestion** / zero → the command's own no-change
handling) and the Jira naming line. The sentences explaining store-first, the fallback union and
why the filesystem source still exists move verbatim under `pipeline-rationale.md`'s existing
**Change name resolution** heading. Callers: `brainstorm.md:24` cites the section unchanged;
`skills/flow-status/SKILL.md:32-49` switches `flow state list` to `flow state resolve` (every
column its table prints is carried on `candidates`). Tests in `state_test.go` via `httptest`, plus
a fixture changes tree for the fallback union and the archive exclusion.

**Stage marks — no new code.** `stage.go` already rejects an undocumented key, a missing
`-session-token`/`-harness` and a substituted token; `check-stage-mark-calls.sh` guards the call
sites. The section keeps its caller obligations and the fenced example; every sentence describing
what the CLI or the guard does and why moves verbatim under `pipeline-rationale.md` **Stage
marks**. Predicted 5601 → about 2500 bytes.
<!-- predicted: wc -c on the section after the move; nothing normative is cut -->

**Handoff output — unchanged.** Its content is composed from run outcome (paths, PR, next command);
`handoff-blocks.md` owns the per-state block. There is no deterministic input a CLI could render it
from, so no `flow handoff` is added.

## 5. The model-resolution gap

`SKILL.md`'s `## Model resolution` bash block reads neither project key today — only two comments
say the key wins. The block gains, for each of `planning model` and `self review model`, before its
`fable` fallback line:

```bash unverified:runs under check-model-resolution-shell.sh once the harness cases below exist
PROJECT_PM="$(project-get.sh "${MAIN_CHECKOUT:-.}" 'planning model' 2>/dev/null | tr -d '`' | xargs)"
if [ -n "$PROJECT_PM" ]; then
  if flow settings models | grep -qx -- "$PROJECT_PM"; then PLANNING_MODEL="$PROJECT_PM"
  else echo "⚠ flow: .flow/project.md '## planning model' body '$PROJECT_PM' is not a valid model — dropped" >&2; fi
fi
```

`flow settings models` prints `store.ValidModels` one per line, sorted; `check-model-keys.sh`
cannot serve here because it reads `settings.go` from this repository and is not shipped.
`check-model-resolution-shell.sh` gains four cases per key — present and valid wins over the
store, present and invalid is reported and dropped, absent falls through, present and valid with the store empty wins over the fallback — with its `flow` stub
answering both `settings get` and `settings models`, `project-get.sh` run for real from `scripts/`,
and `MAIN_CHECKOUT` pointed at a fixture.

## Measurement

Load sets as kan-378 defined them, `wc -c` summed per set, tokens as bytes/4. Before is
kan-378's after column, re-run at `ed2fbf8`; after is recorded by the plan's last task.

| Set | Before | After |
|---|---|---|
| A — creating-run parent | 282272 bytes, ~70568 tok | recorded by the measurement task |
| B — per-subagent fixed overhead | 25694 bytes, ~6423 tok | unchanged by this design |
| C — planner | 97362 bytes, ~24340 tok | recorded by the measurement task |
| D — principles slot | 34679 bytes, ~8669 tok | unchanged by this design |
| E — implementer | 43694 bytes, ~10923 tok | unchanged by this design |
<!-- measured: kan-378's measure.sh block (tasks.md task 8), run as measure.sh <repo> after @ ed2fbf8 -->

Set A after drops `project-configuration.md` and `workspace-isolation.md` from the ordinary path
and shrinks `pipeline.md`; set C swaps the 25221-byte `plan-provenance.md` for its 5822-byte run
half. Predicted A about 205000 bytes, C about 78000 bytes.
<!-- predicted: the same measure.sh with the two contracts removed from A and the run half in C, run at the branch tip -->

## Verification

- Every command under `## lint` in `.flow/project.md` exits clean after every task;
  `scripts/run-guard-tests.sh` and `cd stats && go test ./... -race -count=1` pass.
- `scripts/check-references.sh` clean after every repointed section cite;
  `scripts/check-guard-symlinks.sh` clean with the new symlink; `scripts/check-contract-budget.sh`
  clean with the two new rows.
- `scripts/check-normative-inventory.sh` output at `ed2fbf8` is a subset of its output at the
  branch tip — no `SHALL`/`MUST` sentence disappears in the moves.

## Decisions

### Where `project get` lives

**ID:** project-get-is-bash
**Status:** active
**Chosen:** `scripts/project-get.sh` over a shared `scripts/lib/project-section.sh` — three
inline copies of the extractor already exist in `scripts/`, guards are invoked by bare name through
`skills/flow/scripts/` symlinks with no build or install step, and `flow` reads no project file
today.
**Considered:** `flow project get` in Go — rejected: a fourth extractor in a second language, a
rebuild-and-install step for a five-line awk, and the three Bash callers would still carry theirs.

### How the planner gets the tag table

**ID:** plan-provenance-split-by-audience
**Status:** active
**Chosen:** split `plan-provenance.md` into the run-applied half (kept under its name) and
`plan-provenance-guard.md`, per kan-372's `project-configuration-split-resolution-vs-authoring`.
**Considered:** inlining the four-tag list into `brainstorm.md`'s dispatch prompt — rejected: it
restates canonical text, which the repository's non-repetition rule forbids, and drifts silently.

### `pipeline.md` mechanics move now

**ID:** pipeline-mechanics-in-scope
**Status:** active
**Chosen:** include the fourth scope item in this change: change-name resolution becomes
`flow state resolve`, stage-mark explanation moves to the rationale, handoff output stays.
**Considered:** deferring to a follow-up after measuring — rejected by the operator in
brainstorming.

### `state resolve` reads `spectre/changes/` itself

**ID:** state-resolve-reads-changes-dir
**Status:** active
**Chosen:** the fallback union reads directory names under `<main-checkout>/spectre/changes/`
directly.
**Considered:** shelling out to `spectre list --json` as the prose does today — rejected: a
subprocess dependency on a second CLI for a directory listing, and `spectre`'s `id` is that
directory name.

### Handoff output stays prose

**ID:** no-flow-handoff-command
**Status:** active
**Chosen:** leave **Handoff output** in `pipeline.md` unchanged.
**Considered:** a `flow handoff <name>` renderer — rejected: the block's content is run outcome
the CLI never sees (worktree paths, PR URL, what stopped), so the command would print a template
the agent fills in anyway.

### Fix the model-resolution gap here

**ID:** model-keys-read-in-block
**Status:** active
**Chosen:** the `## Model resolution` block reads both keys through `project-get.sh` and
validates membership through a new `flow settings models`; the harness gains four cases per
key.
**Considered:** an open question only — rejected by the operator; validating through
`check-model-keys.sh` — rejected: it reads `settings.go` from this repository and cannot ship.

## Open questions

None recorded — the convergence check closed with no unresolved item.
