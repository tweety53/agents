## Context

- `/flow-self-review` (`skills/flow-self-review/SKILL.md`) runs one inline six-angle pass and
  offers findings for Jira filing; it fixes nothing.
- kan-829 (**Pipeline defects found mid-run**, `skills/flow-contracts/pipeline.md`) defines the
  blast-radius class and a one-shot fix → review → re-review loop landed by the default route; it
  covers defects a live run hits, never self-review findings.
- The flow store already serves `flow self-review bundle` (`stats/cmd/flow/selfreview.go`,
  `GET /api/v1/self-review/{project}/{change}/bundle`); nothing records a self-review finding.
- One change: the skill text, the store row and the report guard move together — a skill that
  writes `fixed: <sha>` fails today's guard, and a record call with no endpoint is refused.

## Decisions

### Fix every non-big finding, all six angles

**ID:** fix-all-non-big
**Status:** active
**Chosen:** every finding not `big` by kan-829's rule is fixed and landed without asking; a
product-code finding is always `big` (it reaches outside `<agents repo>`) — operator chose option
(a) in the KAN-875 discussion.
**Considered:** (b) a per-finding Fix now / File / Drop prompt — keeps the operator choosing, but
the operator chose unattended fixing; angle 1 only (kan-829 as is) — leaves five angles to tickets.

### Cheap by structure, opus throughout

**ID:** inline-fix-one-branch-one-reviewer
**Status:** active
**Chosen:** the self-review session fixes inline (it already holds the context), all findings on
one branch `self-review-<name>`, one commit each; one fresh `opus` `flow-high` reviewer per round
over the whole branch diff; landed once — two dispatches on a clean pass, regardless of finding
count.
**Considered:** kan-829's per-finding loop (fixer + reviewer per finding) — 2N dispatches, each
re-loading context the session already holds; sonnet for the reviewer or fixer — the operator
ruled model downgrades out ("Opus is good").

### A reviewer-rejected fix becomes big

**ID:** rejected-fix-to-prompt
**Status:** active
**Chosen:** a commit the reviewer judges not to fix its finding, or to make things worse, is
dropped from the branch and its finding is offered for filing like a `big` one.
**Considered:** fixing until the reviewer agrees — no bound on a finding whose fix is a judgement
call, against "cheap".

### Own table, not the panel's findings table

**ID:** self-review-findings-table
**Status:** active
**Chosen:** migration `0035_self_review_findings.sql`, one row per finding: change, angle label,
finding text, disposition `fixed`/`filed`/`declined` (closed at the store), ref (sha for `fixed`,
issue key for `filed`, empty for `declined`), blast radius (nullable). Written by
`flow self-review finding`, read by `flow self-review findings`, served under
`/api/v1/self-review/{project}/{change}/findings`.
**Considered:** `flow record finding` with slot `self-review` — its status vocabulary, Minor-only
deferral rule, `canonical_slot` folding and the panel guards that read it are all panel-specific.

### A store miss warns, never blocks; the report stays the durable record

**ID:** store-write-never-blocks
**Status:** active
**Chosen:** `flow self-review finding` on a store failure prints one warning line and exits 0, no
journal; the committed report carries every finding's disposition, so nothing is lost.
**Considered:** journalling like `flow record` — needs a reconcile replay kind for one
operator-attended command whose data the report already holds.

### `fixed: <sha>` disposition

**ID:** fixed-disposition
**Status:** active
**Chosen:** the report's finding line ends `— fixed: <sha>`, a 7–40 character lowercase hex sha,
read off the landed branch after the rebase so it names the commit on `<default-branch>`.
**Considered:** `declined` for a fixed finding — loses the fact; a free-text note — unparseable.

## Open questions
