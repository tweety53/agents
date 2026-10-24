## Context

- Today `pipeline.md`'s **Every invocation is re-entrant** forbids chaining from implementation into
  integrate without a fresh bare `/flow`. This change makes that chain the default, while the rule
  keeps banning integrate inferred from mid-turn talk.
- The flag has to survive the `/clear` between the plan gate and the implementing run, which is why it
  lives on the state record and not in session memory.
- Precedent: `withdrawn` (a0cc01e7) added a boolean column, a DTO field and contract text the same way.
  The `flow` CLI passes record fields through byte-transparently, and integrate's `writeState`
  rewrites the record as a raw map, so the new field is carried without a CLI change.

## Decisions

### auto-land-default

**ID:** auto-land-default
**Status:** active
**Chosen:** after **Write `IN_PROGRESS`**, a run whose record's `reviewBeforeLanding` is not `true`
prints no `IN_PROGRESS` handoff and continues at integrate's **Deciding which run this is**
(`skills/flow/integrate.md`) in the same invocation. Integrate then prints its own handoff
(`## Branch integrated` or `## Finished`), or a stop.
**Considered:** a project-level `.flow/project.md` key — ruled out because the operator asked for a
per-run override, and a project key would add configuration nobody asked for. Skipping the plan gate
too — ruled out because the request names only the implementation handoff.

### review-flag-in-state-record

**ID:** review-flag-in-state-record
**Status:** active
**Chosen:** a boolean `reviewBeforeLanding` field on the change's state record (a store column
`review_before_landing BOOLEAN NOT NULL DEFAULT FALSE`, a DTO field with `omitempty`). It is set
through `flow state get | jq | flow state set`, so absent or false means auto-land. This was the
operator's choice at brainstorming.
**Considered:** a `## Decisions` entry in `design.md` read as prose — no Go change, but the operator
chose the state record. `decision.json`'s `overrides` — rejected because it holds model pairs only,
and mixing landing into it would blur that record.

### fix-runs-follow-flag

**ID:** fix-runs-follow-flag
**Status:** active
**Chosen:** a fix run (including one from a plain message) ends the same way as a creating run. It
continues into integrate unless `reviewBeforeLanding` is `true`. This was the operator's choice at
brainstorming.
**Considered:** only the implementing run auto-lands — the operator declined this.

### override-any-time

**ID:** override-any-time
**Status:** active
**Chosen:** a plain-language instruction asking to review before landing, or to land without review,
given at any point of any `/flow` run of the change — the creating run's planning included, since
`STARTED` is recorded at kickoff — is written to the record at once. It is one line in the run's
narrative and one `Decisions:` bullet. The latest instruction wins. Such an instruction is never
itself an integrate trigger: integrate still begins only at the run's own chain point, or on a fresh
bare `/flow`.
**Considered:** a `/flow` flag — rejected because `/flow` accepts no flags but `--base`, and the
operator wants to say it mid-run.

### reentrancy-rule-narrowed

**ID:** reentrancy-rule-narrowed
**Status:** active
**Chosen:** **Every invocation is re-entrant** keeps banning integrate inferred from operator words
said mid-turn, and gains one sanctioned chain: the post-`IN_PROGRESS` continuation of
`auto-land-default`. Nothing else chains into integrate.
**Considered:** deleting the rule — rejected because its ban on inferred consent still guards
against starting integrate from a stray "merge and push" said mid-run.

## Open questions

## Live check

- Exercise: build `flowd` from this worktree and run it against a scratch database on a scratch port,
  per the workspace isolation contract (`skills/flow-contracts/workspace-isolation.md`). This is the
  stand-in for the shared dev daemon, which is never restarted.
- Steps: `flow state set` a scratch change at `STARTED` with `"reviewBeforeLanding": true`, then run
  `flow state get` on it. Write it again at `IN_PROGRESS` without the field, and `flow state get`
  once more.
- Before/after: the first get prints `"reviewBeforeLanding": true`, and the second omits the field.
- Not working looks like: the first `state set` falls back to the journal (an unknown-field refusal),
  or the first get omits the field.
