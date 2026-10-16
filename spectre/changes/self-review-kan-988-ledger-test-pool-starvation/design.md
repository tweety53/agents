## Context

- Three non-`big` findings from the kan-988 self-review pass, planned as one change per item 1 of
  **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`). Each is carried by
  exactly one task; none is dropped. All were checked against this worktree at `4f64d5f8`:
  - A (Task 1): `runReproducer` (`stats/internal/guard/runreproducer.go`) refuses absolute argv
    tokens only. `proveReproducer` (`stats/internal/guard/provereproducer.go`) copies the
    reproducer into a detached scratch worktree and runs the pre-fix leg there, so a body naming
    the live worktree's absolute path reads the fixed tree from either leg. The false claim lives
    in `scripts/prove-reproducer.sh`'s header; `skills/flow/scripts/prove-reproducer.sh` is a
    symlink to it, so there is one copy to correct. Only `proveReproducer` knows which tree is the
    live one, so the refusal lives there.
  - B (Task 2): `stats/cmd/flow/record.go`'s `-effort` refusal prints the four accepted words and
    nothing else; **Harness mapping** (`skills/flow-contracts/model-policy.md`) states the
    `default` rule.
  - C (Task 3): `ParseAssistantRecords` (`stats/internal/harvest/transcript.go`) keeps the first
    line per `message.id`, and `dropCarriedMessage` (`stats/internal/harvest/watcher.go`) drops
    every line of the message the previous read ended on. Measured over every gymie transcript
    modified in the last two days (10336 message ids, 6979 with several lines): 260 ids carry
    differing `output_tokens` across their lines — e.g. `[4, 135]`, `[3, 3, 601]`,
    `[4, 4, 4, 660]` — and in all 260 the last line holds the largest. Input and cache counts
    repeat verbatim, so only output and thinking need the maximum. A transcript source does hold
    the final count, so no lower-bound label is needed.
- Spiked at plan time: each task's diff applied alone and all three together to this worktree —
  `gofmt -l .` prints nothing, `go vet ./...` and
  `go test ./internal/guard/ ./cmd/flow/ ./internal/harvest/ -count=1` pass, and
  `scripts/check-guard-symlinks.sh` exits 0; each declared test fails with its task's production
  file(s) reverted (Task 1: `got 1`; Task 2: the stderr assertion; Task 3: output `4`, want `135`,
  in both tests, and the straddle test still fails with only `watcher.go` reverted). Every edit was
  reverted.

## Decisions

### prove-reproducer refuses a body naming the live worktree

**ID:** live-path-refusal
**Status:** active
**Chosen:** `proveReproducer` reads the reproducer before the scratch exists and exits 2 when its
body contains `<worktree>`'s absolute path — as given (cleaned) or symlink-resolved — naming the
path and the cause; an unreadable body is left to the copy step's existing refusal. The header
states the refusal in place of the structural claim and lists it under exit 2.
**Considered:** refusing in `runReproducer` — it sees only the tree it runs in, so the pre-fix leg
would check the scratch path the script never names; rewriting the path inside the scratch copy —
it silently changes the reproducer under proof; a path-boundary match instead of a substring —
extra code for a false positive (a sibling directory whose name extends the worktree's) the message
already explains.

### The effort refusal names the default rule

**ID:** effort-refusal-names-default
**Status:** active
**Chosen:** the refusal appends "-- an effort the session does not state is recorded as default,
never guessed".
**Considered:** accepting `unknown` as an alias of `default` — widens the closed set the store and
the model policy both name.

### A message counts its final output

**ID:** harvest-final-output
**Status:** active
**Chosen:** within one read, the kept Record per id carries the maximum output and thinking counts
of its lines; across reads, the carry records the id with the counts already counted, and a later
line whose output exceeds them stays as an output-only record carrying the difference.
**Considered:** taking the last line instead of the first — the same within one read but still
loses the final count when the message straddles two reads, which the code's own comment calls the
common case; a lower-bound label where output is shown — the final count is in the transcript;
re-harvesting stored rows — no recorded ask, and the offsets have moved past them.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `medium` — Task 3 changes token accounting across reads, where a double count or a
dropped delta passes every existing test; a low-effort read could miss it.
**Considered:** `low` — Tasks 1 and 2 alone are a refusal and a message.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** one group, g1: Tasks 1–3 in plan order, one fixer — the tasks share no file and each is
a spiked diff applied with `git apply`.
**Considered:** one group per task — three dispatches for three verified diffs.

## Open questions

## Live check

After this lands and flowd restarts (`make restart` in `stats/`), pick one subagent dispatch
recorded after the restart and compare its stored output tokens with the sum, per message id, of
the largest `output_tokens` among that id's lines in its `agent-<id>.jsonl`. Before: rows such as
kan-988's reviewer output 28 and 178. After: the two figures match. Not working looks like a stored
output equal to the sum of each id's first line while the transcript's maxima are larger.
