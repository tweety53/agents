## Context

KAN-393's three recorded failures came from KAN-29, a three-repo change run before any cross-repo
mechanism existed. KAN-363 already closed two of them — `check-unfinished-work.sh` (task 8) and
`check-task-commit-fields.sh` (task 9) take an optional `[canonical-worktree]` argument and follow
`link.md` through `scripts/lib/change-plan.sh`, with `/flow` passing the argument from the
resolved worktree set (`skills/flow/implement.md`, `skills/flow/integrate.md`; the caller-side
resolution rule lives in `skills/flow-contracts/finish-contract-run1.md`). This change is one
change because exactly one named gap survives: `gather-dispatch-context.sh`, left out of KAN-363's
guard list, still resolves `proposal.md`/`design.md`/`tasks.md` from a single change directory
inside a single worktree. Under KAN-363's pointer-tree model a satellite change directory carries
only `link.md`, so a satellite worktree's bundle carries no plan, and the hand workaround (gathering
from the canonical pair) mislabels `project commands`, `incidents` and the `head:` sha as the
canonical repository's.

Full design narrative (context, approach, section-by-section design, test surface):
`docs/superpowers/specs/2026-09-08-kan-393-flow-give-guards-a-canonical-worktree-concept-design.md`
in this worktree. This file records the decisions; the design doc carries the design. The Jira's
literal proposal — read the canonical worktree from `.flow/project.md` — is superseded by
KAN-363's shipped mechanism and is recorded below as a rejected alternative, not an open question.

## Decisions

### Close the gather gap; the `.flow/project.md` proposal is superseded

**ID:** close-the-gather-gap-not-project-md
**Status:** active
**Chosen:** scope the change to giving `gather-dispatch-context.sh` the KAN-363 treatment — follow
`link.md` via `scripts/lib/change-plan.sh`, optional `[canonical-worktree]` argument — and record
the Jira's `.flow/project.md` proposal as superseded by KAN-363's shipped mechanism.
**Considered:** The Jira's literal proposal, a per-project `.flow/project.md` key naming the
canonical worktree, rejected because the canonical repository is a per-change property (KAN-343's
canonical was `gymie`, KAN-363's was `agents`), which a static per-project key cannot name —
KAN-363's per-change `link.md` plus tree-level `spectre/peers` already carries the same fact where
it belongs. Also considered mechanizing canonical-worktree resolution itself (a `flow` CLI verb or
script so callers stop deriving "the one member of the resolved set whose `tasks.md` exists" from
prose), rejected as wider than KAN-393's named gap — it touches the `stats` CLI and the
worktree-resolution contract for a rule that already works. Also considered closing KAN-393 as
stale, rejected because the gather gap is real and unfixed.

### The script resolves the plan; the caller does not route around it

**ID:** script-side-resolution
**Status:** active
**Chosen:** `gather-dispatch-context.sh` sources `scripts/lib/change-plan.sh` and resolves the
change's plan through `change_plan_dir`, gaining an optional seventh argument
`[canonical-worktree]`.
**Considered:** Caller-side only — prose in `implement.md` telling the conductor to gather
satellite bundles from the canonical `<worktree>`/`<change-root>` pair and write them into the
satellite's `.superpowers/sdd/` — rejected because the bundle then carries the canonical
repository's `project commands`, `incidents` and `head:` sha, the exact mislabeling a satellite
implementer must not get, and the script stays link-blind for every future caller. Appending the
argument as position 7 rather than inserting it: both guards KAN-363 gave an argument appended it,
and every existing five- and six-argument call site stays valid unchanged.

### The trust boundary moves with the resolution

**ID:** boundary-follows-resolution
**Status:** active
**Chosen:** the three plan leaves' `within_root` refusal is checked against the resolved content
directory, not the argument `<change-root>`; the invocation-level containment rules
(`<change-root>` inside `<worktree>`, exit 2) are unchanged.
**Considered:** Checking the leaves against the argument `<change-root>` still, rejected because
a satellite's canonical plan always resolves outside it and every satellite bundle would report
its plan sections as refused. The invariant is preserved: the canonical directory is the same
change's tracked content, reached through a link whose peer name and change id `change-plan.sh`
allowlist-checks.

### An unresolvable satellite skips; it never refuses

**ID:** skip-not-refuse
**Status:** active
**Chosen:** when `change_plan_dir` returns 1, the three plan sections are reported as skips with a
distinct label (`satellite plan unresolved — link.md at <path>`) and the run still exits 0.
**Considered:** Refusing with exit 2, the shape `check-unfinished-work.sh` chose for the same
condition, rejected for the bundle: the bundle is advisory — `implement.md`'s "the bundle never
gates a run" — so refusing would abort a dispatching stage the way a guard refusal does, the wrong
failure direction for a bundle with no verdict to give. A guard verdict must never silently clear;
a bundle has no verdict. The distinct label keeps "the plan is elsewhere and unreachable" from
reading like a plain change's legitimately-absent `design.md`.

### Remote resolution is labeled, and worktree-local sections stay worktree-local

**ID:** labeled-canonical-plan
**Status:** active
**Chosen:** when the content directory differs from the argument `<change-root>`, the three plan
sections' labels carry a `(canonical <peer>:<change-id>)` suffix, composed with the existing
`(scoped to task(s) …)` suffix; `project commands`, `incidents` and the `head:` sha keep resolving
from the given `<worktree>`.
**Considered:** Gathering everything from the canonical worktree for satellite bundles, rejected —
the satellite implementer would read the canonical repository's lint/test/run list, the exact
defect this change exists to remove. Unlabeled remote sections, rejected because a dispatch reading
the bundle could not tell where the plan resolved from.

### Callers pass the canonical worktree on every call

**ID:** uniform-canonical-arg
**Status:** active
**Chosen:** `implement.md`'s per-bundle gather step and `review-panel.md`'s rebuild step pass
`<canonical-worktree>` on every call — the uniform rule `finish-contract-run1.md` already states
for `check-unfinished-work.sh` — where canonical means the member of the resolved worktree set
whose own `<project>/<spec-root>/changes/<name>/tasks.md` exists.
**Considered:** Passing the argument only when the dispatch targets a non-canonical worktree,
rejected because the branch is one more thing every gather call site must reason about, for an
argument that is inert on a single-repo change (local `tasks.md` wins before the link is
consulted, so plain-change bundles stay byte-identical either way).

## Open questions

None. Every question raised during brainstorming was answered before this stage closed.
