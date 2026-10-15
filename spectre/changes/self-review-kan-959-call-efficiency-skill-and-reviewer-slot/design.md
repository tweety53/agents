## Context

- Two non-`big` findings from the kan-959 self-review pass, planned as one change per item 1 of
  **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`). Each is carried by
  exactly one task. Both were verified against the tree at `1a422665`:
  - `renderSlotPrompt` (`stats/internal/guard/renderslotprompt.go`) accepts a slot only when it is
    `mutation`, an `exp-<name>` id without a `/`, or a key of the hardcoded `rspSlots` map; every
    other id is refused exit 2 with `no reviewer template for slot`. The template it then reads is
    `<skill-dir>/<id>-reviewer-prompt.md`, so the map duplicates what the skill directory already
    states. `scripts/render-slot-prompt.sh`'s header already describes the template path per slot
    and an unknown slot as exit 2, so it needs no change.
  - `checkVerbatimMoves` (`stats/internal/guard/verbatimmoves.go`) prints the raw sentence after `::`
    on both FAIL lines; `vmAcks` skips a `#`-leading line as a comment and drops one leading `\`. A
    heading copied as printed is therefore a comment, and a sentence opening with `\` copied as
    printed loses its `\` — the same copy contract broken the same way.
  - `stats/internal/store/settings.go`'s `ValidReviewers` is a separate list — the settings store's
    selectable panel slots — and `test-sanity`/`call-efficiency` are the always-on extra dispatch's
    roles, not settings-selected; it is out of scope.
- Both tasks' Go was spiked at `1a422665`: each new test fails without its fix and passes with it;
  removing task 1's `/` check fails task 1's test, and dropping task 2's `\` case fails task 2's test;
  `go vet` and `gofmt -l` clean; the whole `internal/guard` package passes with task 1 applied.
- No finding was dropped.

## Decisions

### A slot is valid when its template file exists

**ID:** slot-from-template
**Status:** active
**Chosen:** drop `rspSlots`; after `mutation`'s own case, a slot is valid when it carries no `/` and is
either `exp-<name>` (template read later from `experimental/<name>.md`, unchanged) or
`<skill-dir>/<id>-reviewer-prompt.md` is a file. Derives validity from the one place that defines a
slot, so a new template renders on any build.
**Considered:** adding the two ids to the map and rebuilding — the next slot hits the same wall;
reading the slot list from `review-panel.md`'s roster — a second source to parse when the template
file is already the definition; dropping the `/` check because the template must exist anyway — an id
like `../<dir>/x` resolves to an existing template outside the skill directory and an output path
outside `.superpowers/sdd/`.

### FAIL lines print the acknowledgement line itself

**ID:** fail-line-is-ack-line
**Status:** active
**Chosen:** a helper `vmAckLine` prefixes one `\` to a sentence opening with `#` or `\`; both FAIL lines
print through it, and the script header states the escape. The text after `::` then copies into
`verbatim-moves.txt` as printed.
**Considered:** escaping `#` only — a sentence opening with `\` breaks the same copy contract the same
way and costs one more condition; changing `vmAcks` to stop treating `#` as a comment — breaks every
existing list's comment lines; documenting the manual escape only — the header already did, and the
"exactly as printed" sentence still contradicted it.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `low` — two small guard changes, each with one test; no concurrency, data model or store
write. Task 1 touches a validation path, but its spike already proves both refusals hold.
**Considered:** `medium` — the diff carries nothing a low-effort read would miss.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** one group, tasks 1–2, one fixer in plan order — the tasks are file-disjoint and small.
**Considered:** one fixer per task — two dispatches for work one fixer finishes inside its context.

## Open questions

## Live check

None — the change edits two guards' argument checking and output text; their Go tests drive each
guard end to end, and no daemon, store or other runtime state changes.
