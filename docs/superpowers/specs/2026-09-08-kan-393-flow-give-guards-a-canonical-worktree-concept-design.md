# kan-393 — Flow: give guards a canonical-worktree concept for cross-repo changes — design

Date: 2026-09-08
Status: approved at the merged convergence confirm (the `/flow` design gate)
Source: KAN-393, filed from KAN-29's self-review (angle: problems encountered)

## Context

KAN-29 was run as a three-repo change (agents canonical, two satellite repos) before any
cross-repo mechanism existed. Its self-review recorded three guard failures, and KAN-393 carries
them. Two of the three are already closed by the archived change
`kan-363-let-a-change-s-spectre-tree-live-in-every-repo`, which shipped `spectre link` + `link.md`
(`## Part of` / `## Parts`), the tree-level `spectre/peers` file, and
`scripts/lib/change-plan.sh`:

- `check-unfinished-work.sh` (KAN-363 task 8) takes an optional `[canonical-worktree]` argument
  and counts a satellite worktree against the canonical plan instead of reporting
  `OUTSTANDING: no plan at …` — KAN-393's second bullet.
- `check-task-commit-fields.sh` (KAN-363 task 9, plus KAN-367's optional sixth `<change-name>`
  argument) resolves a satellite's task through the link — the `check-task-commit-fields.sh` half
  of KAN-393's first bullet. `/flow` passes the canonical worktree on both guards' calls
  (`skills/flow/implement.md`, `skills/flow/integrate.md:39`; the caller-side resolution rule —
  "the one member of the resolved set whose own `<project>/<spec-root>/changes/<name>/tasks.md`
  exists" — lives in `skills/flow-contracts/finish-contract-run1.md`).

KAN-363's own proposal lists the guards it fixed; `gather-dispatch-context.sh` is not among
them. That gap survived KAN-363 rather than being created by it, and it is the remaining substance
of KAN-393:

- `gather-dispatch-context.sh` has no link awareness at all. It requires `<change-root>` to sit
  inside `<worktree>` (exit 2 otherwise) and reads `proposal.md`, `design.md` and `tasks.md` from
  that one change directory.
- Under KAN-363's pointer-tree model a satellite change directory carries **only** `link.md`. A
  bundle gathered for a satellite worktree therefore reports the three plan sections as
  `skipped: … (absent)` and the satellite implementer is dispatched with no plan in the bundle.
- The hand workaround — passing the canonical worktree and canonical change dir as the first two
  arguments and writing the bundle into the satellite's `.superpowers/sdd/` — silently mislabels
  the bundle: the `project commands` section then carries the **canonical** repository's
  lint/test/run commands rather than the satellite repository's own, the `incidents` section reads
  the canonical project's incident log, and the header `head:` line names the canonical sha.
  This is the "hand-substituted every time" cost KAN-393 records.
- The Jira's literal proposal — "read the canonical worktree from `.flow/project.md`" — predates
  KAN-363's shipped design. The canonical repository is a **per-change** property (KAN-343's
  canonical was `gymie`; KAN-363's was `agents`), which a static per-project key in one
  repository's `.flow/project.md` cannot name. KAN-363's mechanism — per-change `link.md`,
  tree-level `peers`, and the caller passing the canonical worktree from the resolved set
  (design decision `guards-take-the-canonical-worktree-path`) — supersedes it.

Call sites checked while fixing the caller prose: `skills/flow/implement.md` (one gather per
dispatch bundle, five- or six-argument shape) and `skills/flow/review-panel.md` (one rebuild per
review round, five-argument shape). The shipped symlink farm carries the script through
`skills/flow/scripts/` with a single `lib` directory symlink, so a fifth `lib/` source needs no
new farm entries; `check-guard-symlinks.sh`'s `$SCRIPT_DIR/<name>` sibling rule already matches
this script's four existing `lib/` sources, and the fifth takes the identical shape.
`spectre/specs/` is empty — no capability spec exists to amend, so this change plans no spec edit.

## Approach

Two shapes were considered for closing the gap:

- **A. Script-side resolution (chosen).** `gather-dispatch-context.sh` learns what KAN-363 taught
  the other two guards: resolve the change's plan through `scripts/lib/change-plan.sh`, with an
  optional canonical-worktree argument. One mechanism, already built, tested and mutated by
  KAN-363, reused rather than duplicated.
- **B. Caller-side only (rejected).** Prose in `implement.md` tells the conductor to gather
  satellite bundles from the canonical `<worktree>`/`<change-root>` pair and write them into the
  satellite's `.superpowers/sdd/`. Rejected for two reasons: the bundle then carries the
  canonical repository's `project commands`, `incidents` and `head:` — the exact mislabeling a
  satellite implementer must not get — and the script stays link-blind for every future caller
  that finds it, leaving the next cross-repo change to rediscover the workaround.

## Design

**1. Signature and plan resolution.** The argument list keeps its existing order and meanings; an
optional seventh argument `[canonical-worktree]` is appended. Both guards KAN-363 gave an
argument appended it rather than inserting it between existing positions
(`check-unfinished-work.sh` arg 3, `check-task-commit-fields.sh` arg 5), and appending keeps every
existing five- and six-argument call site — the panel's rebuild and the scoped per-bundle gather —
valid unchanged. After the existing `<worktree>`/`<change-root>` validation (unchanged, including
the exit-2 containment contract), the script resolves one **content directory** through
`change_plan_dir <worktree> <name> <canonical-worktree>`:

- A plain change (local `tasks.md` exists) resolves to its own change directory; the script then
  behaves byte-for-byte as today — including the bundle body, so `SKIP-WHEN-UNCHANGED` hashes
  match and nothing re-fires for single-repo changes.
- A satellite (no local `tasks.md`, a `link.md` carrying `## Part of`) resolves to the canonical
  change directory, reached through the supplied canonical worktree, or through
  `<worktree>/<spec-root>/peers` when no argument was passed — the resolution order
  `change_plan_dir` already implements, unchanged.

`proposal.md`, `design.md` and `tasks.md` are read from the content directory. Every other part of
the bundle keeps its existing source: the change name for the header line stays the invocation's
`<name>` (a link names the same change id on both sides by construction — `spectre link` writes
the satellite's directory under the canonical id).

**2. The trust boundary moves with the resolution.** The three plan leaves' per-source refusal
check (`within_root`, the "resolves outside the change directory" refusal) is checked against the
**resolved content directory**, not the argument `<change-root>`. The boundary's invariant is
unchanged — a change directory's tracked, pull-request-editable content must never be read from
outside that change — because the canonical directory is the same change's tracked content,
reached through a link whose peer name and change id `change-plan.sh` allowlist-checks character
for character. The invocation-level containment rules are untouched: `<change-root>` must still
resolve inside `<worktree>`, exit 2 on any divergence or breach, exactly as documented in the
script's header today.

**3. An unresolvable satellite skips; it never refuses.** When `change_plan_dir` returns 1 — a
genuine satellite whose canonical plan neither the argument nor peers can reach — the three plan
sections are reported as skips with a distinct label
(`skipped: tasks.md (satellite plan unresolved — link.md at <path>)`, and the same shape for
`proposal.md` and `design.md`), and the run still exits 0. The rationale is the deliberate inverse
of `check-unfinished-work.sh`'s exit-2 refusal for the same condition: a guard verdict must never
silently clear, but a bundle has no verdict to give — `implement.md`'s "the bundle never gates a
run" means refusing here would abort a dispatching stage the way a guard refusal does, which is
the wrong failure direction for an advisory bundle. The distinct label keeps a satellite's
unresolved plan from reading like a plain change's legitimately-absent `design.md`, so the
operator who reads the census can tell "the plan is elsewhere and unreachable" from "this change
has no design" — the same two-shapes distinction that motivated `change-plan.sh` itself.

**4. Remote resolution is labeled.** When the content directory differs from the argument
`<change-root>`, the three plan sections' labels carry a `(canonical <peer>:<change-id>)` suffix,
composed with the existing `(scoped to task(s) …)` suffix where both apply — e.g.
`## tasks.md (canonical gymie:kan-410) (scoped to task(s) 3,7)`. A dispatch reading the bundle can
see the plan came from another tree, and the census stays honest about where each section
resolved. Plain changes' labels are untouched.

**5. Worktree-local sections stay worktree-local.** `project commands` (extracted from
`<worktree>/.flow/project.md`'s `## lint`, `## test` and `## run`), `incidents`
(`flow record incidents -C <worktree>`), and the header `head:` sha keep resolving from the given
`<worktree>`, never from the resolved content directory. This is the property approach B could
not deliver: a satellite implementer's bundle carries its own repository's lint/test/run list and
its own project's incident log, plus the canonical plan.

**6. Caller prose.** `implement.md`'s per-bundle gather step and `review-panel.md`'s rebuild step
both pass `<canonical-worktree>` on **every** call — the uniform rule `finish-contract-run1.md`
already states for `check-unfinished-work.sh` ("Pass `[canonical-worktree]` on every call in the
run") — rather than branching on whether this dispatch targets a satellite. Canonical means the
member of the resolved worktree set whose own `<project>/<spec-root>/changes/<name>/tasks.md`
exists; for a single-repo change that is the worktree itself and the argument is inert. The
panel's five-argument call simply gains argument 7; the six-argument scoped call inserts nothing.

**7. Tests.** `scripts/test-gather-dispatch-context.sh` gains the satellite shapes, mirroring the
fixture style `test-check-unfinished-work.sh` and `test-lib-change-plan.sh` already use
(`peers` file + `link.md` fixtures):

- a satellite bundle resolves the three plan sections from the canonical worktree via the
  argument, with the `(canonical …)` labels, and keeps its own worktree's `project commands`;
- the same resolution via `peers` when no argument is passed (the main-checkout shape, where the
  peer path resolves on disk);
- an unresolvable satellite → three distinct skip labels, exit 0, bundle still written;
- a plain change → labels and body byte-identical to today's (hash stability);
- the invocation containment rules still exit 2 (unchanged argument contract);
- `(canonical …)` and `(scoped to task(s) …)` labels compose.
- `check-guard-symlinks.sh` keeps passing: the fifth `lib/` source matches the rule the existing
  four already satisfy, and no new farm entry is needed.

**8. Consequences that ride along.** `implement.md` and `review-panel.md` are budget-ratcheted
owned files under `check-contract-budget.sh` — measured during planning, the prose additions land
well inside both rows (implement.md 29726 bytes against its 32500-byte row, review-panel.md 50032
against 58732, both with this change's edits applied), so no budget row moves. No capability-spec
task exists (`spectre/specs/` is empty). No staging note deletion: this change was seeded from no
research note.
