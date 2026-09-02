# Design — kan-380-flow-self-review-model-overridable-per-project

## Context

Bounded change; no separate spec file. `proposal.md` carries why; this file carries how.
`SELF_REVIEW_MODEL` (**Model resolution**, `skills/flow/SKILL.md`) today reads one store field and
inherits `DEFAULT_MODEL` when it is empty; `PLANNING_MODEL` (KAN-374) already has the three-tier
shape with a project key and a verified opus fallback. No Go file reads `.flow/project.md` keys, so
the new key is read by the skill prose alone.

## Resolution

```bash unverified:confirm `jq -r '.selfReviewModel // empty'` yields the empty string on both "" and a missing key
SELF_REVIEW_MODEL="$(printf '%s' "$SETTINGS_JSON" | jq -r '.selfReviewModel // empty')"
# <project>/.flow/project.md's `## self review model` body, when present and a ValidModels member, wins
[ -z "$SELF_REVIEW_MODEL" ] && SELF_REVIEW_MODEL=fable
```

Order: a per-run session instruction > `## self review model` in `<project>/.flow/project.md` >
the store's `selfReviewModel` > the literal `fable`. An unreachable store resolves to `fable` too,
named a fallback exactly as `DEFAULT_MODEL`'s `sonnet` is. The session override is recorded with
the dispatch it changes and never written back.

## The project key

`## self review model` in `<project>/.flow/project.md`: one `ValidModels` member and nothing else,
trimmed and matched byte for byte, like `## planning model`; a body matching no member is reported
by name (quoting what was found) and dropped. This repository's own `.flow/project.md` sets it to
`fable`.

## The handshake

The archive-phase self-review subagent (**Run self-review**, `skills/flow/archive.md`) opens its
report with `Model: <the model named in its own system prompt>`. A mismatch with
`SELF_REVIEW_MODEL` re-dispatches once on `opus`; a second mismatch proceeds on whatever answered
and names that model in the run's own output. No dispatch record — self-review records none today
and this change adds none. The line is stripped before the five-angle body is used.

## What "empty" means

The store's `selfReviewModel` empty now means "the store's own default, `fable`", the same
meaning `planningModel`'s empty carries — not "inherit `defaultModel`". Every site that says
inherit is reworded, none deleted:

- `skills/flow/SKILL.md` Model resolution (bash block and the `SELF_REVIEW_MODEL` paragraph)
- `skills/flow-settings/SKILL.md` — the printed line and JSON description in **Read current
  settings**, the option "Inherit default model" in **Offer to change each field** becomes "Store
  default (fable)", and the paragraph contrasting it with `planningModel` is dropped since the two
  now agree
- `skills/flow-contracts/finish-contract-run2.md` **Run self-review**'s parenthetical
- `stats/internal/store/settings.go` — `Settings.SelfReviewModel` and `PlanningModel` doc comments
- `stats/cmd/flow/settings.go` — the `-self-review-model` flag help string
- `stats/internal/store/migrations/0016_flow_settings_self_review_model.sql` — header comment only;
  migrations are tracked by filename (`stats/internal/store/migrations.go`), so the file's content
  is not checksummed

No Go logic, no migration, no test change: `ValidateSettings` already accepts empty and every
`ValidModels` member including `fable`.

## Files

`skills/flow/SKILL.md`, `skills/flow/archive.md`, `skills/flow-contracts/project-configuration.md`,
`skills/flow-contracts/finish-contract-run2.md`, `skills/flow-settings/SKILL.md`,
`.flow/project.md`, the three `stats/` comment sites, and `scripts/check-contract-budget.sh` rows
for any of those that outgrow their budget.

## Testing

Prose guards from `.flow/project.md`'s `## lint`: `check-vocabulary.sh`, `check-references.sh`,
`check-contract-budget.sh`, `check-installed-citations.sh`, `check-markdown-integrity.py`,
`check-stage-mark-calls.sh`, `check-dispatch-paragraphs.sh`; `cd stats && go vet ./... && gofmt -l .`
and `go test ./internal/store/... ./cmd/flow/... -race -count=1` to prove the comment edits
compile and change no behaviour.

## Decisions

### Empty `selfReviewModel` means `fable`, not inherit

**ID:** self-review-empty-is-fable
**Status:** active
**Chosen:** empty resolves to the literal `fable`, matching `planningModel` — the only shape in
which the ticket's default is reachable.
**Considered:** keeping inherit with `fable` below `DEFAULT_MODEL` — unreachable, since the store
always supplies a `defaultModel`. Keeping inherit and offering `fable` only through the project key
— the ticket's default would not be a default.

### Opus fallback is verified by a handshake, unrecorded

**ID:** self-review-opus-handshake
**Status:** active
**Chosen:** the self-review subagent names its model first; a mismatch re-dispatches once on
`opus`; no dispatch record, since self-review has none today.
**Considered:** a documented tier with no handshake — the harness silently substitutes models, so
prose alone cannot say what ran. Adding a `self-review` dispatch role — a Go change beyond the
ticket.

### This repository sets the key to `fable`

**ID:** agents-sets-self-review-key
**Status:** active
**Chosen:** `## self review model` with body `fable` in this repo's `.flow/project.md` — a no-op
in resolution that exercises the key on every archive run here.
**Considered:** `opus` — hides the fable/opus fallback path in the one repo that exercises it.
Omitting the key — the ticket names the file.

## Open questions

None recorded.
