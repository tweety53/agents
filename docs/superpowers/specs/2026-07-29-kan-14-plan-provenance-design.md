# KAN-14 — plan provenance: make unverified plan claims visible and refusable

**Jira:** KAN-14 "Plan provenance — make unverified plan claims visible and refusable"
**Date:** 2026-07-29
**Repo:** `/Users/tweety53/Projects/agents`
**Status:** approved

## Why

A plan written by `/myflow-start` asserts things about the world. Nothing checks any of them.

The KAN-6 run (the Review Queue plugin) is the evidence. Six plan claims were wrong, and each one
cost an implementer a detour:

| The plan said | The world says | Cost |
|---|---|---|
| `override fun getToolWindowIds()` | it is a Kotlin **property**, not a function | Task 2 rework |
| `DiffNotificationProvider`'s SAM is non-null | parameter **and** return are `@Nullable` | Task 3 rework |
| the marker is readable from `DiffRequest` | it never propagates there — only to `DiffContext` | Task 4 rework |
| "baseline is 194 tests" | 197 | confusion at every task |
| "expect 4 fewer tests" | 11 | — |
| rename fixture `"original"` → `"staged change"` | undetectable as a rename at **any** `-M` threshold | Task 5 rework |

Roughly forty minutes of a ~two-hour run. But the time is the smaller half. The third row is the
one that matters: had the implementer transcribed the plan's snippet instead of investigating,
`request.getUserData(REVIEW_DIFF)` would have been permanently `null`, the context-menu guard
permanently false, and **the entire feature a no-op that every unit test still passed.** It was
caught because the plan happened to tell that implementer to check — a one-off instruction on one
task, not a rule.

The pattern across all six: **in-repo code was fine; claims about things outside the repo were
not.** Snippets calling a platform SDK, and numbers that were estimated rather than run.

## What this is not

This does not make plans correct, and nothing proposed here would have made the KAN-6 plan correct.
It makes unchecked claims **visible and refusable**: labelled at plan time, and treated as
hypotheses at implementation time. A wrong `verified:` tag still gets through. That limit is
deliberate — see *Decisions*.

## What changes

Four pieces. One new contract, one guard script plus its harness, two skill edits.

### 1. The contract — `skills/myflow-contracts/plan-provenance.md`

A new file beside `state-file.md`, `jira-integration.md` and the rest, because two skills consume it
and the existing contracts directory is exactly where a shared definition lives. It defines four
tags.

**On fenced code blocks**, in the info string:

- `verified:<how>` — the symbol and its signature were checked against the real artifact. The
  `<how>` is the evidence, concretely: `verified:javap intellij.platform.diff.jar`,
  `verified:compiled in-tree`, `verified:PlatformActions.xml:452`.
- `unverified:<what-to-check>` — a hypothesis. Names the check that would settle it:
  `unverified:confirm the member is a function not a property`.

**On numeric and factual claims**, as an HTML comment on the following line:

- `measured:<command> @ <ref>` — the number came from running something.
  `<!-- measured: ./gradlew test @ c515c42 -->`
- `predicted:<what-confirms-it>` — a forecast, naming the command that settles it.

**The load-bearing rule: a number carrying neither tag may not appear in a plan at all.** "194
tests" was not mislabelled — it was invented. Labelling only helps if stating an unsourced number is
itself the violation.

Every fenced block needs a tag; in-repo code satisfies it cheaply with `verified:compiled in-tree`.
The cost the rule imposes falls almost entirely on claims about things outside the repository, which
is where every KAN-6 error was.

### 2. The guard — `scripts/check-plan-provenance.sh`

Mechanical enforcement, in this repository's established guard style: a header comment stating the
rule *and the failure that motivated it*, `file:line` output, and a `CHECK_PLAN_PROVENANCE_ROOT`
environment override so a harness can point it at a fixture tree. Added to the `## lint` key in
`.myflow/project.md`.

**Scoped narrowly, on purpose: `openspec/changes/*/tasks.md` only, excluding `archive/`.**
`check-references.sh`'s own header records what over-firing cost this repo — 28 false failures on
its own tree, after which suppression markers were the only way to silence them, and those markers
then switched off the real checks sharing those lines. A guard that fires outside the one file type
its rule is about earns exactly that outcome. Archived plans predate the rule and are history; they
are not rewritten.

The guard checks that provenance is **stated**, never whether it is true. A script cannot verify
that `javap` was actually run. Conflating the two would make the guard both unimplementable and
dishonest about what it proves.

### 3. The harness — `scripts/test-check-plan-provenance.sh`

Follows `test-check-references.sh` exactly: sandboxed fixture trees under `TMPDIR`, assertions on
both exit status and output, never touching the real repository tree. Added to the `## test` key.

Both directions of error are covered — a tagged block must pass, an untagged block must fail, and an
untagged number must fail. A guard whose failing case is untested is a guard that can silently stop
checking.

### 4. The two skill edits

**`skills/myflow-start/SKILL.md`, step D.** While writing `tasks.md`, tag every fenced block and
every numeric claim, then run the guard before publishing the artifact. Code that cannot be verified
is tagged `unverified:` — never omitted, because a plan without the snippet is worse than a plan
with a labelled guess.

**`skills/myflow-do/SKILL.md`, the implementer dispatch.** A standing clause added beside the
existing three:

> **PLAN PROVENANCE:** a fenced block tagged `unverified:` is a hypothesis, not code to transcribe.
> Establish the real API before writing against it, and report what you found. A block tagged
> `verified:<how>` was checked as stated; if it does not compile, report that — do not contort the
> code to match it.

The KAN-6 implementers did this by instinct and it is why the change is correct. This makes it
instruction rather than luck.

Plus a row in `rules/myflow-manual-review.mdc`'s contracts table, so the new contract is loadable
the same way the other five are.

## Decisions

### How much verification the rule demands

**ID:** provenance-tag-not-hard-gate
**Status:** active
**Chosen:** A stated-provenance tag, enforced by a guard script — verification stays a judgment
call, and the script enforces only that the claim is labelled, which is what a script can actually
check.
**Considered:** A hard verification gate that compiles every external-API snippet and runs every
measurement before publishing — would have caught all six KAN-6 errors, but requires a build in the
planning stage, which `/myflow-start` deliberately has no worktree for, and makes every proposal
round slower whether or not the plan touches an external API. Banning implementation bodies for
external APIs from plans entirely — removes the error class outright, but overrides
`writing-plans`' explicit code-block mandate and makes plans markedly less concrete, trading a
known cost for a vaguer one.

### What an `unverified:` tag obliges downstream

**ID:** unverified-is-a-hypothesis
**Status:** active
**Chosen:** The implementer must establish the fact before writing, and report what it found — the
tag changes behaviour at implementation time, which is the only place it can save anything.
**Considered:** Also writing the confirmed fact back into `tasks.md` so the plan self-heals and
later tasks inherit it — would have stopped Tasks 3, 4 and 5 each rediscovering the same class of
SDK mismatch, but puts implementers in the business of editing the plan mid-run, which is a larger
change to who owns the plan than this proposal wants to make. Leaving the tag purely informational
— cheapest, and the option least likely to change what actually happened in KAN-6.

### Scope of this change

**ID:** provenance-without-parallel-lanes
**Status:** active
**Chosen:** Plan accuracy only. The serialization of independent task groups is filed as KAN-15 and
left in To Do.
**Considered:** Both in one change — they share no code, and the parallel-lane work needs per-lane
build isolation, which is the riskier half and would hold the cheap win hostage. Serialization only
— saves ~20 minutes of wall clock on a KAN-6-shaped run but does nothing about the errors, which
cost more and nearly shipped a defect.

## Testing

This repository's verification is its guard scripts, run through `.myflow/project.md`:

- `scripts/test-setup.sh` and `scripts/test-check-references.sh` must stay green.
- `scripts/test-check-plan-provenance.sh` is new and must cover both directions: tagged blocks pass,
  untagged blocks fail, untagged numbers fail, and `archive/` is not scanned.
- `scripts/check-vocabulary.sh` and `scripts/check-references.sh` must stay green over the new
  contract file — the new file introduces cross-references, which is precisely what
  `check-references.sh` guards.

There is no auto-fix command in this repository, so every guard hit is fixed by editing the
offending line, never by weakening the guard or adding a suppression marker.

## Risks

**The guard over-fires and gets suppressed.** This is the specific failure mode this repo has
already lived through once, documented in `check-references.sh`'s header. Mitigated by the narrow
file scope, and by the harness asserting the passing direction as hard as the failing one. If the
guard cannot be made quiet on a correct plan, the guard is wrong — not the plan.

**A tag becomes ritual.** `verified:` can be typed without verifying, and nothing detects that. The
honest position is that this change raises the cost of an unchecked claim from zero to "write a
false evidence string", and makes the claim visible to the human at the proposal gate and to the
reviewer at the implementation gate. It does not make it impossible.

**`.myflow/project.md` is this repo's own config**, so a mistake in the `## lint` key changes how
every future `/myflow-do` verifies this repository. The key is edited alongside the harness that
proves the script runs.
