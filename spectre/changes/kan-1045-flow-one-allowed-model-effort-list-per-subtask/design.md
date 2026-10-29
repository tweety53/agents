## Context

KAN-1045 asks for one mechanism stating, per flow role, the allowed model × effort pairs, read by
every dispatch site and guard, with escalations as list entries and never `haiku`/`fable`. The
design was approved in plan mode on 2026-10-29
(`.superpowers/sdd/2026-10-29-kan-1045-flow-one-allowed-model-effort-list-per-subtask-design.md`).

At Agent-tool PreToolUse time the hook sees `tool_input.model`, the effort through
`subagent_type` (`flow-<effort>`, `flow-planner-<effort>`) or `tool_input.effort` (Explore, Plan),
and no role field: the description row key is ambiguous (`Task <x>/<n>` is an implementer or a
gated reviewer bundle; `panel-<r>-merger` matches the reviewer shape), and the store's `-role`
vocabulary (`stats/cmd/flow/record.go` `recordRoles`) is coarser than the pair rules (`reviewer`
covers five rules). Hence an explicit `Role:` line.

The daemon (`stats/internal/api/records.go` `checkDecisionModels`) has no repository path, so it
keeps its own opus/sonnet and no-high checks; the hook is the enforcement point.

## Decisions

### Where the list lives

**ID:** policy-table-in-model-policy
**Status:** active
**Chosen:** a `| Role | Allowed pairs | Requires marker |` table under `## Allowed pairs` in `skills/flow-contracts/model-policy.md`, parsed at runtime by Go — the precedent `loop.go` set reading `skills/flow-self-review/SKILL.md` — operator's pick.
**Considered:** a JSON data file embedded in the daemon with `go:embed` — machine-only, harder to read and edit, and every prose citation would point at a file no reader opens.

### How a dispatch names its role

**ID:** role-line-on-every-dispatch
**Status:** active
**Chosen:** every flow dispatch prompt carries a whole line `Role: <role id>`; the guard resolves the role from it alone — operator's pick.
**Considered:** deriving the role from the description row key and the `Key:` line — no prompt change, but the `Task <x>/<n>` and `panel-<r>-merger` ambiguities need permanent special cases.

### A dispatch the list does not admit

**ID:** refuse-unresolved-role
**Status:** active
**Chosen:** in a flow run, a dispatch with no `Role:` line, a role the table lacks, or a pair outside its role's rows is refused, the refusal listing the role's allowed pairs — operator's pick, as the ticket requires.
**Considered:** allowing an unlisted dispatch with a warning — leaves the scattered-pair failure the ticket exists to end.

### Escalations

**ID:** escalation-as-marked-rows
**Status:** active
**Chosen:** an escalation is a table row whose `Requires marker` cell names the prompt line it needs (`No-progress escalation`, `Operator-approved high`); the hook's `FLOW_HIGH_MARKERS` and `FLOW_XHIGH_TYPES` rules are deleted, since no row allows `xhigh` — operator's pick.
**Considered:** keeping the escalation markers as a separate hook rule — two lists again.

### The guard cannot answer

**ID:** guard-exit-2-denies
**Status:** active
**Chosen:** the hook denies a flow-run dispatch when `check-dispatch-pair` exits 2 (table unreadable, usage error), as `enforce-loop-dispatch.py` already does on any non-zero exit — recommended option taken.
**Considered:** allowing the dispatch when the guard cannot answer — a broken table would silently disable enforcement.

### Daemon-side decision validation

**ID:** daemon-keeps-valid-models
**Status:** active
**Chosen:** `flow record decision` keeps only `store.ValidModels` and the no-high rule on implementer, fixer and groups — recommended option taken; the daemon has no repository path to read the table from.
**Considered:** embedding the table in the daemon — needs the table inside `stats/`, contradicting **policy-table-in-model-policy**.

### Loop rows keyed by role id

**ID:** loop-suffix-maps-to-role
**Status:** active
**Chosen:** `loop.go` maps a loop key's suffix to a role id (`self-review-pass`, `-plan`, `-fix`, `-review-1`, `-review-r`; the same five under `pipeline-fix-`) and checks the pair through the shared parser in-process; its ESCALATION (completed earlier round) and review-1 `**Chosen:**` effort checks stay, narrowing within the table.
**Considered:** requiring a `Role:` line on loop dispatches as well and dropping the suffix map — the loop hook already keys every row by `Key:`, so a second identity could disagree with the first.

## Open questions

## Recommended approach

One parser, `stats/internal/guard/modelpolicy.go`, reads **Allowed pairs**; one guard,
`check-dispatch-pair`, wraps it for the agent-baseline hook; `loop.go`, `planagentpair.go` and
`planclass.go` call the parser in-process. The hook's three hard-coded rules and the self-review
pairs table are deleted, and every dispatch template gains its `Role:` line. A JSON file was
rejected (**policy-table-in-model-policy**), description-derived roles were rejected
(**role-line-on-every-dispatch**), and a daemon-embedded table was rejected
(**daemon-keeps-valid-models**).

## Critical files

- `skills/flow-contracts/model-policy.md` — gains **Allowed pairs**, the canonical table.
- `stats/internal/guard/modelpolicy.go` — the parser and the `check-dispatch-pair` guard.
- `scripts/check-dispatch-pair.sh` — the guard's `flow_guard_exec` shim.
- `stats/internal/guard/loop.go` — loop pair check moves onto the table.
- `stats/internal/guard/planagentpair.go`, `stats/internal/guard/planclass.go` — pairs from the table.
- `hooks/enforce-agent-baseline.py` — calls the guard; `PINNED_ROLES`, `FLOW_HIGH_MARKERS`, `FLOW_XHIGH_TYPES` deleted.
- `skills/flow-self-review/SKILL.md` — its pairs table deleted.
- `skills/flow/brainstorm-planner.md`, `skills/flow/review-panel.md`, `skills/flow/implement.md`, `skills/flow/visual-verify.md`, `skills/flow/visual-verify-tooling-analysis.md`, `skills/flow/gated-review-fix.md`, `skills/flow/review-panel-late-fix.md`, `skills/flow-contracts/pipeline.md`, `skills/flow-plan/SKILL.md`, `skills/flow-fast/SKILL.md` — pair prose becomes citations; dispatch templates gain `Role:` lines.

## Existing code to reuse

- `stats/internal/guard/loop.go` `lpTable` — the markdown-table parsing shape (backticked tokens per cell), generalised into `modelpolicy.go`.
- `scripts/lib/flow-guard.sh` `flow_guard_exec` — the shim every Go guard uses.
- `stats/internal/guard` `Registry` — guard registration by `init()`.
- `hooks/enforce-agent-baseline.py` `in_flow_run`, `ZCODE_PAIR` — flow-run detection and the ZCode harness mapping, kept.
- `stats/internal/repotest/enforce_agent_baseline_hook_test.go` — the Go-runs-Python hook test harness.

## Verification

- `cd stats && make -s web-dist && go vet ./... && gofmt -l . && go test ./internal/guard/... ./internal/repotest/...` green.
- Every `scripts/check-*.sh` guard named in `.flow/project.md`'s `## lint` exits clean.
- `git grep -n -e PINNED_ROLES -e FLOW_HIGH_MARKERS -e FLOW_XHIGH_TYPES -e '| Key suffix |'` prints nothing.

## Live check

None — the hook and guards read repository files and hold no runtime state; the hook tests run the
real hook against the real table. After landing, `./setup.sh global` installs it, and the next
flow-run dispatch with no `Role:` line is refused: that refusal is what working looks like, and a
`Role:`-less dispatch going through is what failure looks like.
