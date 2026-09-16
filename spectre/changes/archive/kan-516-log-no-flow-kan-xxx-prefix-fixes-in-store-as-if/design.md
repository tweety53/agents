## Context

The fix path already exists — `skills/flow/implement.md` section 3 runs whenever the router reads
`IN_PROGRESS` with an argument. The only gap is a router entry: the router is entered only by a
`/flow` invocation, and a plain message never reaches it. This is one change: a hook that supplies
the change name deterministically on every plain prompt, plus the router rule that classifies a
message and hands it to the existing fix path with a fresh session token.

## Decisions

### Full fix run, not a thinner inline path

**ID:** unprefixed-fix-full-ceremony
**Status:** active
**Chosen:** An unprefixed problem report runs exactly as `/flow <name> <message>` would — the
document-fix planning pass, implementation, review panel, verification and handoff — so the store
sees an ordinary fix run.
**Considered:** An inline fix wrapped in marks only (tokens and a fix iteration would land, but the
planning pass, panel and verify records would not — a second, thinner kind of fix run the store
could not tell apart from the real one) — rejected. Asking each time whether to run it as a fix —
rejected: a click on every message defeats the convenience the ticket is about.

### Trigger is change requests only

**ID:** unprefixed-fix-trigger-change-requests
**Status:** active
**Chosen:** A message that reports a problem or asks for a change to the change's code or artifacts
is a fix run; a question, a discussion, or an unrelated task stays an ordinary turn.
**Considered:** Every non-slash message — rejected: would turn "why did you do that?" into a
planning pass.

### Ambiguous messages ask once

**ID:** unprefixed-fix-ambiguity-ask-once
**Status:** active
**Chosen:** A message that reads either way is one `AskUserQuestion` — "Run this as a fix of
`<name>`?", Yes (recommended) / No — ordinary turn; silence takes Yes and the fix run's handoff
carries a `⚠` line.
**Considered:** Treating doubt as an ordinary turn — rejected: the silent data loss the ticket
reports. Treating doubt as a fix run unconditionally — rejected: a planning pass over a plain
question.

### Cold sessions and bare Jira mentions stay out of scope

**ID:** unprefixed-fix-cold-sessions-out-of-scope
**Status:** active
**Chosen:** Only a session that already ran `/flow <name>` is in scope. A session with no prior run
in the change's worktree, and a Jira key named in prose without the slash, both stay ordinary turns.
**Considered:** Extending the rule to either case — rejected: each is a discovery mechanism of its
own, and neither is the case the ticket reports.

### Discovery via a prompt hook plus skill text

**ID:** unprefixed-fix-discovery-hook-plus-skill
**Status:** active
**Chosen:** The rule itself lives in the `/flow` skill; a `UserPromptSubmit` hook
(`hooks/flow-active-change.py`) supplies the change name deterministically on every plain prompt by
reading the transcript's last `/flow` stage-begin mark, so the rule survives context compaction and
long sessions.
**Considered:** Skill text alone — rejected: works on every harness but is lost once the skill's
text falls out of context. A hook alone — rejected: the hook can name the change but cannot classify
a message, and the rule must be readable by a session with no hook registered.

### The fix run's state gate is the existing one

**ID:** unprefixed-fix-state-gate-existing
**Status:** active
**Chosen:** The fix run starts only when `flow state get <name>` answers `IN_PROGRESS`. Any other
state — `STARTED`, `FINISHED`, no state, or a store that cannot be reached — makes the message an
ordinary turn, never a wrong-state handoff, since the operator invoked nothing.
**Considered:** Treating a mismatch as a wrong-state handoff like an actual `/flow` invocation —
rejected: there is no command to report a wrong state for when nothing was invoked.

## Open questions

None — this round was fully seeded by `docs/research/kan-516.md`, whose own convergence checks
(`/flow-plan`'s per-topic closing check) already resolved every question this design raises.
