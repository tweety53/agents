# Verification — kan-302-panel-code-review-slot-hangs-on-fork

Evidence for task 4: guards clean, inventory reconciled. All checks below were run against the
worktree after tasks 1-3 landed (commits `79c46c1`, `a3164e8`, `d67bcc2`).

## 1. Lint run

`cd stats && gofmt -w .` ran first (auto-fix step); it reformatted nothing.

Every command in `.myflow/project.md`'s `## lint` section exited clean:

| Command | Exit |
|---|---|
| `scripts/check-vocabulary.sh` | 0 |
| `scripts/check-references.sh` | 0 |
| `scripts/check-plan-provenance.sh` | 0 |
| `scripts/check-task-build-green.sh` | 0 |
| `scripts/check-workspace-isolation.sh` | 0 |
| `scripts/check-uitest-overrides.sh` | 0 |
| `scripts/check-contract-budget.sh` | 0 (`BUDGET-OK: 104 owned Markdown file(s) within budget`) |
| `scripts/check-markdown-integrity.py` | 0 |
| `scripts/check-stage-mark-calls.sh` | 0 |
| `scripts/check-guard-symlinks.sh` | 0 |
| `scripts/check-self-review-report.sh` | 0 |
| `scripts/check-installed-citations.sh` | 0 |
| `scripts/check-installed-rules.sh` | 0 (`INSTALLED-RULES-WORKTREE`: this is a linked worktree, tracks the main checkout) |
| `scripts/check-normative-inventory.sh` | 0 — printed 1261 normative sentences from 104 files to stdout (counts to stderr); no verdict, per its documented contract |
| `cd stats && gofmt -l .` | 0 |
| `cd stats && go vet ./...` | 0 |
| `cd stats/web && npx tsc -b` | 0 |

**One environment prerequisite, not a guard failure:** `go vet ./...` failed on the first run —
`internal/web/embed.go:29:12: pattern all:dist: no matching files found` — because
`stats/internal/web/dist/` did not exist yet in this worktree. This is documented, deliberate
behaviour (`stats/Makefile`'s `test` target comment): the Go build embeds the SPA's build output via
`//go:embed all:dist` and fails loudly at compile time when it is missing, rather than at runtime.
It is not caused by this change (documentation-only, no Go or SPA source touched) and not one of the
lint commands' own failures. Fixed by building the SPA once — `cd stats/web && npm ci && npm run
build` — which populates `internal/web/dist/` (gitignored, per the repository-root `.gitignore`'s
`stats/internal/web/dist/` entry at line 23). `go vet` and the rest of the list above exited clean
after that build; no repository file changed by running it.

## 2. Normative-inventory diff

`scripts/check-normative-inventory.sh` run against the current tree, diffed against task 1's
`normative-baseline.txt` (1258 lines, captured at a clean `79c46c1`) under `LC_ALL=C` (the sort
order the guard itself uses):

```
43a44
> **Every panel slot SHALL carry a 15-minute wall-clock ceiling from its dispatch.**
51a53
> **No panel slot SHALL be dispatched onto a skill or agent that forks its own background agent.**
790a793
> The dispatcher SHALL NOT block indefinitely on a slot's completion notification; it tracks each in-flight slot's elapsed time and enforces the ceiling itself, so a slot emitting output while making no progress is still bounded.
```

Three lines added, none removed, none reworded. Classified against the plan's permitted set:

- **Removed (roster spec's `SHALL`, "The light preset's third slot invokes the harness's code-review
  skill") — not observed in this diff, as expected.** That requirement still lives in
  `openspec/specs/myflow-review-panel-roster/spec.md` unedited; this change's delta only replaces it
  at finish run 2, when the delta syncs into `openspec/specs/`. The state observed here is "not yet
  removed" — the pre-finish-run-2 state.
- **Not added, and never will be: the delta specs' own sentences.** An earlier draft of this section
  claimed `openspec/changes/` is inside the guard's owned corpus and that the delta specs'
  `SHALL`/`SHALL NOT` sentences therefore appeared in `normative-baseline.txt` already, before this
  diff. **That premise is false** — checked directly: `grep -c "wall-clock ceiling\|forks its own
  background agent" normative-baseline.txt` returns `0`. The real cause is structural exclusion.
  `scripts/lib/owned-corpus.sh`'s `OWNED_CORPUS_SCOPE_DIRS` (line 94) scopes the corpus to `skills/`,
  `rules/`, `openspec/specs/`, `commands/`, `commands-claude/`, `.myflow/`, and the `.md`/`.mdc`
  files sitting directly at the repository root; `openspec/changes/` is never scanned, and that
  file's own header records the exclusion as deliberate, precisely so a during-a-change before/after
  inventory comparison stays valid. So a delta spec's `SHALL` sentences are never inventoried at any
  point in this change's life — not before this diff, not after. They enter the inventory only at
  finish run 2, as part of `openspec/specs/`. Found by task 4's per-task reviewer, as finding F1.
- **Added, in `skills/myflow-do/SKILL.md`** — all three diff lines above. Confirmed by location:
  `skills/myflow-do/SKILL.md:529` ("No panel slot SHALL be dispatched onto a skill or agent that
  forks its own background agent."), `:536-538` ("Every panel slot SHALL carry a 15-minute
  wall-clock ceiling from its dispatch." / "The dispatcher SHALL NOT block indefinitely on a slot's
  completion notification..."). These are task 3's new subsection, exactly the bullet the plan names
  for this file.

No difference falls outside the permitted set. Nothing was restored or reworded.

## 3. `openspec validate`

```
openspec validate --changes kan-302-panel-code-review-slot-hangs-on-fork
```

Exit 0. `✓ change/kan-302-panel-code-review-slot-hangs-on-fork` (alongside the repository's two
other open changes, unaffected).

## 4. Budget headroom (measured after tasks 2-3's edits)

| File | Size (bytes) | Budget (bytes) | Headroom |
|---|---|---|---|
| `skills/myflow-do/SKILL.md` | 87,388 | 103,578 | 16,190 |
| `skills/myflow-do/SKILL-rationale.md` | 18,793 | 20,902 | 2,109 |

Both clear their budget; `check-contract-budget.sh` confirms (`BUDGET-OK`). Net growth measured:
SKILL.md +1,832 bytes over its pre-edit baseline (85,556 @ `441ed2b`), SKILL-rationale.md +2,058
bytes over its pre-edit baseline (16,735) — combined growth 3,890 bytes, not the "roughly 2 KB"
`tasks.md`'s global constraint originally estimated. This does not threaten either budget (16,190
and 2,109 bytes of headroom remain). The correction has landed: `proposal.md`'s Impact section
(lines 71-77) now states the measured figures and the 3,890-byte growth against the "roughly 2 KB"
estimate directly, and `tasks.md`'s global-constraints bullet (lines 30-34) likewise records both
the pre-edit and post-edit figures and the reconciled growth.

## 5. Task 2 step 5's grep, re-run

```
grep -rn 'code-review' skills/ rules/ commands/ commands-claude/ openspec/specs/
```

Surviving hits, all expected:

- `skills/README.md:34` — `superpowers:requesting-code-review` (excluded external skill name)
- `skills/myflow-do/SKILL.md:62` — `superpowers:requesting-code-review`
- `skills/myflow-do/SKILL.md:457` — `superpowers:requesting-code-review` (slot 0's dispatch cell)
- `openspec/specs/myflow-review-panel-roster/spec.md:83,87,97,116` — the requirement this change
  removes; `openspec/specs/` is synced by finish run 2, not by this change

No hit outside those two classes. Task 2's cleanup holds after task 3's edits.
