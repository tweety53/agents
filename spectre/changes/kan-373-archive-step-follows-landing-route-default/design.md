## Context

`skills/flow/integrate.md`'s landing question (`## 2. Ask how the branch should land`) offers three
routes — pull request, merge and push, handle manually — and `skills/flow/archive.md`'s step 10
always opens its own PR for the archive commit, independent of which route the operator chose for
the change itself. The two files never share any state today beyond the change's own state file,
which records `prUrl` (set only by the PR route) but nothing for merge-and-push or manual.

## Decisions

### No new state-file field for "which route was chosen"

**ID:** no-route-field
**Status:** active
**Chosen:** rely on the existing same-invocation continuation `skills/flow/integrate.md`'s "After
merge-and-push specifically" section already defines. `archive.md`'s step 10 checks whether it is
running as that continuation (a value already in scope for that one call path) rather than reading
a new persisted field.
**Considered:** adding a `landingRoute` field to the state file, written at run 1 and read at run
2. Rejected: the PR and manual routes always defer archiving to a separate, later invocation
(sometimes a different session entirely) that has no way to verify the recorded route is still
accurate — the operator could have merged the PR through any mechanism the forge offers, including
one this pipeline never chose. A field that can silently go stale is worse than one that does not
exist; the same-invocation check is accurate by construction because nothing external can have
happened in between.

### The new key's malformed-value handling matches `## jira`'s

**ID:** malformed-value-matches-jira
**Status:** active
**Chosen:** a `## default landing route` body that is absent, or present but not exactly one of the
three literal route names, is reported by name and dropped — falling back to today's behavior
(pull request stays the recommended default) — the same disposition `## jira`'s own malformed-body
case and `## standards`'s own malformed-entry case already use.
**Considered:** silently ignoring a malformed value with no report. Rejected: this project's own
convention (`## jira`, `## standards`) already treats a malformed row as worth surfacing to the
operator even though it degrades gracefully; a second, quieter convention for one more optional key
would be an inconsistency nobody asked for.

### The question is still always asked

**ID:** default-changes-recommendation-not-the-ask
**Status:** active
**Chosen:** `## default landing route` changes which option `skills/flow/integrate.md`'s landing
prompt marks `(default, recommended)`; it never skips the prompt itself.
**Considered:** skipping the prompt entirely when a project default is declared, landing
automatically. Rejected: `skills/flow/SKILL.md`'s own guardrails forbid `/flow` from accepting a
flag or silently choosing a route on a creating run, and `skills/flow/integrate.md` already states
this same question is asked once and answered explicitly every run — a project default narrows
which answer is recommended, never whether the operator is asked to give one.

## Open questions

<!-- none -->
