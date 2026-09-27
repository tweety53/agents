# Design — kan-843-agents-remove-cursor-codex-harnesses-myflow-era

## Context

Measured on `main` at `b8faae9a` (2026-09-28):

- `setup.sh` modes: `cursor | claude-code | codex | zcode | all | global`. zcode reuses
  `commands-claude/`, copies `AGENTS.md` per project and renders `~/.zcode/AGENTS.md`.
- `scripts/check-vocabulary.sh` guards retired myflow stage names and the old panel roster only.
- `scripts/check-self-review-report.sh` scans every file under `docs/self-review/`; it exempts 17
  named pre-shape reports and accepts `myflow-<angle>` labels, which 14 reports carry.
- Dev store (`flow` DB, read-only query): 205 `changes` rows with `updated_by = 'myflow stage begin
  (synthetic)'`; 16 decisions with bare-array `panel.dispatches` elements and 4 with bare-array
  `groups` elements (of 77); 1 pricing row (`glm-5.3-flash`) with a null 1h rate, its collapsed
  `cache_write_per_mtok` equal to its 5m rate (0 = 0).

## Decisions

### Keep Claude Code and zcode only

**ID:** harnesses-claude-code-and-zcode
**Status:** active
**Chosen:** delete Cursor and Codex support outright — `commands/`, `install_cursor`,
`install_codex`, `install_rules_cursor`, the `cursor`/`codex`/`all` modes, `~/.cursor`/`~/.codex`
installs and the `~/.codex/AGENTS.md` block — and trim every live mention. `AGENTS.md` stays
(zcode). The `-harness` placeholder rule stays, its reason restated for `~/.claude` and `~/.zcode`.
**Considered:** keeping `all` as an alias for claude-code+zcode — rejected, nobody uses it and
`global` already covers both.

### Delete the vocabulary guard

**ID:** delete-vocabulary-guard
**Status:** active
**Chosen:** delete `scripts/check-vocabulary.sh` and `scripts/test-check-vocabulary.sh` and every
reference (lint lists, sibling guards, Go guards, coverage lib, tests).
**Considered:** porting it to Go — rejected, the vocabulary it guards is retired.

### Self-review reports: delete 17, relabel 14

**ID:** self-review-delete-17-relabel-14
**Status:** active
**Chosen:** the guard loses its legacy-label branch and exemption list; the 17 pre-shape reports
are deleted (git history keeps them) and `myflow-<angle>` → `flow-<angle>` in the 14 others.
**Considered:** keeping the exemptions (legacy stays); deleting all 31 reports — the operator chose
this option.

### Stats: migrate legacy rows, then remove the readers

**ID:** stats-legacy-migrate-then-remove
**Status:** active
**Chosen:** one data migration and the code removal together:
- `changes.updated_by` `'myflow stage begin (synthetic)'` → `'flow stage begin (synthetic)'`;
  `stages.SyntheticChangeUpdatedBy` follows.
- `decisions`: every bare-array element of `panel.dispatches` becomes `{"slots": <array>}`, of
  `groups` becomes `{"bundles": <array>}`; the aggregate SQL drops its bare-array branch.
- `pricing`: where the 1h rate is null and the collapsed column equals the 5m rate, set 1h = 5m;
  drop `cache_write_per_mtok`. `PricingRate.CacheWritePerMTok`, its seed values and the
  "nil 1h + collapsed column equals 5m" flat rule go; the flat rule is "1h equals 5m". Every row
  prices exactly as before.
- Untouched: the decision `implementer` string form (still written today as
  `"skipped — inline"`) and generic absent-key handling (`fixer`).
The migration reaches the dev store when the operator next starts `flowd` on the new binary; no
agent restarts it. Earlier migration files are never edited.
**Considered:** leaving the compat readers — the operator chose migrate-and-remove.

### Ticket special cases

**ID:** drop-ticket-special-cases
**Status:** active
**Chosen:** remove code that special-cases a ticket or names a test by ticket: the self-review
exemption list, `test-setup.sh`'s "KAN-369" group (renamed by behaviour), the kan-488 wording in
`check-model-resolution-shell.sh`'s message, the `// KAN-185` comment beside an unchanged value.
**Considered:** stripping all ~3,000 inline `(KAN-NNN)` citations — the operator kept them.

## Open questions
