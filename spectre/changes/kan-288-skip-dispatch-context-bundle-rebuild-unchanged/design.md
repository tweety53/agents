## Context

`gather-dispatch-context.sh` (`skills/flow/scripts/`) is invoked from three call sites —
`skills/flow/implement.md`'s `flow.sdd-tdd` stage, `skills/flow/review-panel.md`'s stage start, and
its panel-fix round — each redirecting the script's stdout straight to
`<worktree>/.superpowers/sdd/dispatch-context.md`. The rule that it must rebuild every time exists
to protect one invariant: a dispatch must never read a plan a concurrent fix has already made
stale. That invariant does not require regenerating identical bytes; it requires the check to run
every time and the file to reflect the latest state whenever it has actually changed.

## Decisions

### Bake the skip-when-unchanged check into the script itself

**ID:** bake-into-script
**Status:** active
**Chosen:** change `gather-dispatch-context.sh`'s own contract to take the destination path as a
5th argument and write to it internally — the only place that can decide not to write is the thing
that knows what the new content is, so it also needs to know where the old content lives.
**Considered:** a thin wrapper script (`maybe-gather-dispatch-context.sh`) that runs the existing
script to a temp file, diffs it against the current output, and only replaces the file on a
difference. Rejected: it still pays the full generation cost every call (the "150KB, unnecessary"
part of the complaint is unaddressed), and it adds a second script name every caller and every
future reader has to track for no behavior a single script can't provide directly.

### Hash the full bundle body, not just the three change files

**ID:** hash-full-body
**Status:** active
**Chosen:** the comparison hash covers everything the script prints after the header's
`generated:`/`head:` lines — the found/skipped/refused census and every `## <label>` section
(proposal.md, design.md, tasks.md, the principles file, project commands). This is exactly the
content a dispatched subagent reads, so "unchanged" is correct by construction, and a source going
from absent to present changes the hash automatically because both the census line and the
concatenated content differ.
**Considered:** hashing only proposal.md + design.md + tasks.md, leaving the principles file and
`project.md`'s lint/test/run sections out of the comparison. Rejected: a genuine edit to either
mid-run would silently fail to trigger a rebuild, which is exactly the staleness the existing
"always rebuild" rule exists to prevent — narrowing the hash's scope would reopen it for two of the
five sources.

### Header lines are excluded from the hash

**ID:** header-excluded-from-hash
**Status:** active
**Chosen:** the `generated:` timestamp and `head:` short sha are written fresh every call but never
enter the hash. They are metadata about *when* the bundle was produced, not part of the plan a
dispatch reads; including them would make the hash differ on every call regardless of content and
defeat the whole mechanism.
**Considered:** none — this follows directly from `hash-full-body`'s own rationale and needed no
separate weighing.

## Open questions

<!-- none -->
