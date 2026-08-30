# kan-370-self-review-model

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

> **Relocation:** no

Two tasks: the store/API/client/CLI plumbing for a `selfReviewModel` field (TDD, Go), then the
skill-level wiring that reads it and dispatches self-review as a subagent (doc-only, no tests).
`design.md` is canonical for the decisions; `proposal.md` for why.

**Baseline, measured before any edit:**

- `internal/store/settings_test.go`: 4 test functions.
  <!-- measured: grep -c '^func Test' internal/store/settings_test.go -->
- `internal/api/settings_test.go`: 3 test functions.
  <!-- measured: grep -c '^func Test' internal/api/settings_test.go -->
- `cmd/flow/settings_test.go`: 3 test functions.
  <!-- measured: grep -c '^func Test' cmd/flow/settings_test.go -->
- `internal/client/client_test.go`: 40 test functions (whole file, shared across every client
  method).
  <!-- measured: grep -c '^func Test' internal/client/client_test.go -->
- Latest migration is `0015_flow_settings.sql`; next is `0016`.
  <!-- measured: ls internal/store/migrations/ | tail -1 -->

- [x] 1. `selfReviewModel` end to end: migration, store, API, client, CLI

Add a `SelfReviewModel` field to the settings record at every layer it already exists in, empty
string meaning "inherit `DefaultModel`" — this is the one place `ValidateSettings` accepts empty
for a model field, unlike `DefaultModel`, which must always be a non-empty `ValidModels` member.

**`internal/store/migrations/0016_flow_settings_self_review_model.sql`** — `ALTER TABLE
flow_settings ADD COLUMN self_review_model TEXT NOT NULL DEFAULT ''`. `NOT NULL DEFAULT ''`
rather than nullable, so `GetSettings` never has to distinguish "column is NULL" from "column is
the empty string" — both already mean the same thing (inherit).

**`internal/store/settings.go`**:
- Add `SelfReviewModel string` to `Settings`.
- `ValidateSettings`: validate `SelfReviewModel` only when non-empty (`s.SelfReviewModel != "" &&
  !ValidModels[s.SelfReviewModel]` → `ErrInvalidModel`, same sentinel `DefaultModel` uses — there's
  no need for a second error type, the wrapped value already disambiguates which field failed).
- `PutSettings`: add `self_review_model` to the `INSERT ... ON CONFLICT DO UPDATE` statement and
  its bound args.
- `GetSettings`: select `self_review_model` alongside the existing two columns; the no-rows branch
  leaves `SelfReviewModel` at its zero value (`""`) — that's already the correct "inherit" default,
  so no third package-level default constant is needed the way `DefaultModel`/`DefaultReviewers`
  are.

**`internal/api/settings.go`**: add `SelfReviewModel string `json:"selfReviewModel"`` to
`settingsDTO`; carry it through `toSettingsDTO` and the `put` handler's `store.Settings{...}`
construction.

**`internal/client/client.go`**: add `SelfReviewModel string `json:"selfReviewModel"`` to
`Settings` (the wire-shape copy, same field name/tag as the API DTO).

**`cmd/flow/settings.go`**:
- `runSettingsSet`: add a `-self-review-model` flag, `fset.String("self-review-model", "",
  "self-review's model, e.g. opus; empty inherits -model")` — optional, no required-flags check
  added for it (only `-model`/`-reviewers` stay required, per the existing check).
- Pass it through into the `client.Settings{...}` literal built for `putSettings`.
- Update `settingsUsage`'s `settings set` line to show the new optional flag.

Write/extend tests at each layer, RED before GREEN:
- `internal/store/settings_test.go`: `ValidateSettings` accepts `SelfReviewModel: ""` and a valid
  model, rejects an invalid one; `PutSettings`/`GetSettings` round-trip the field including the
  empty-string case (store tests already exercise a live Postgres per this file's existing
  pattern — extend the existing round-trip test rather than adding a parallel one).
- `internal/api/settings_test.go`: `put` accepts an empty `selfReviewModel`, rejects an invalid
  one with 400; `get` echoes whatever the stub store returns for it.
- `cmd/flow/settings_test.go`: `settings set` accepts a run with `-self-review-model` omitted
  (still succeeds on `-model`/`-reviewers` alone) and with it set; `settings get`'s JSON output
  includes `selfReviewModel`.
- `internal/client/client_test.go`: `GetSettings`/`PutSettings` round-trip `SelfReviewModel`
  through the fake transport those tests already use.

```text unverified:confirm the exact existing round-trip test names before extending them, migration numbering could shift if another change lands 0016 first
```

**Files:** `internal/store/migrations/0016_flow_settings_self_review_model.sql`,
`internal/store/settings.go`, `internal/store/settings_test.go`, `internal/api/settings.go`,
`internal/api/settings_test.go`, `internal/client/client.go`, `internal/client/client_test.go`,
`cmd/flow/settings.go`, `cmd/flow/settings_test.go`
**Tests:** extended `internal/store/settings_test.go`, `internal/api/settings_test.go`,
`cmd/flow/settings_test.go`, `internal/client/client_test.go` (see above; no new test files)
**Regression:** reverting this commit removes `selfReviewModel` from the wire and store shape
entirely — `/flow-settings` (task 2) would have no field to write, and `/flow`'s model resolution
(task 2) would have nothing to read, so task 2 cannot land without this one.
**Baseline:** before=4 after=4 store tests extended, before=3 after=3 api tests extended,
before=3 after=3 cmd/flow tests extended, before=40 after=40 client tests extended (no new test
function count changes — existing tests are extended to cover the new field, per the note above)
<!-- predicted: no net change in `grep -c '^func Test'` per file after task 1 -->
**Commit:** `feat(store): add selfReviewModel to the harness-wide settings record`
**Build:** green

- [x] 2. Wire `selfReviewModel` through `/flow-settings` and `/flow`'s self-review dispatch

Doc-only: no Go code, no tests — this task changes skill prose that a future `/flow` run follows,
not code this repository executes today.

**`skills/flow-settings/SKILL.md`**: after the existing "Default model" question, add a third
question — "Self-review model" — offering `ValidModels` plus an explicit "Inherit default model"
option (maps to `""`), seeded from the current value read in step 1. Extend step 3's `flow
settings set` call to always pass `-self-review-model` (required by the CLI's new flag shape from
task 1 — it's optional there, but this skill always states an explicit value, "keep current" or
otherwise, exactly as it already does for `-model`/`-reviewers`).

**`skills/flow/SKILL.md`** "Model resolution": add `SELF_REVIEW_MODEL="$(printf '%s'
"$SETTINGS_JSON" | jq -r '.selfReviewModel')"` to the resolution block; state that an empty result
resolves to `DEFAULT_MODEL`, same fallback story as an unreachable store.

**`skills/flow/archive.md`** step 9: change the self-review procedure from an inline reasoning
pass to a subagent dispatch. The subagent receives `gather-self-review-context.sh`'s output and
the five-angle table, runs on `SELF_REVIEW_MODEL`, and returns the five angles' findings (each
angle's list, or an explicit none-marker) as its report body — nothing else about its contract.
The main session then runs the existing per-angle filing-ask prompts, report assembly and commit
unchanged, since a subagent cannot drive `AskUserQuestion`.

**`skills/flow-contracts/finish-contract.md`** Run 2 step 9: this file is canonical for the
self-review procedure — add one sentence noting the reasoning pass runs as a subagent on the
resolved self-review model, so `skills/flow/archive.md`'s own step 9 continues to carry only what
is specific to *executing* it, per that file's existing "not a second statement of this rule" note.

**Files:** `skills/flow-settings/SKILL.md`, `skills/flow/SKILL.md`, `skills/flow/archive.md`,
`skills/flow-contracts/finish-contract.md`
**Tests:** none
**Regression:** reverting this commit leaves `selfReviewModel` writable via the CLI (task 1) but
unread by any skill — self-review keeps running inline on the session model, silently ignoring the
new setting.
**Baseline:** before=0 after=0 (no test-count-bearing files touched)
**Commit:** `docs(flow): dispatch self-review on the configured model`
**Build:** green
