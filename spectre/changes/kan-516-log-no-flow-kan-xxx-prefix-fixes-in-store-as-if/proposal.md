# kan-516-log-no-flow-kan-xxx-prefix-fixes-in-store-as-if

## Why

Problems found during testing are typed as plain messages, without the `/flow kan-XXX` prefix,
because that is more convenient. A store "run" is only ever the set of stage marks sharing one
session token, and `flow.document-fix` never fires for a plain message, so today those fixes leave
no run row, no fix iteration, no dispatch rows, no findings, no task-count observation and no
token/cost figures behind — the session does the work and the store records a session that went
idle.

## What changes

A `UserPromptSubmit` hook (`hooks/flow-active-change.py`) names the session's active `/flow` change
on every plain prompt, derived from the transcript's last `/flow` stage-begin mark. A new
`skills/flow/SKILL.md` subsection and a matching `pipeline.md` transition-table row make a plain
message that reports a problem or asks for a change, in a session whose last `/flow` run left the
change at `IN_PROGRESS`, run as the fix run `/flow <name> <message>` would — planning pass,
implementation, panel, verify, handoff — under a fresh session token, so the store records it as one
more ordinary fix run.
