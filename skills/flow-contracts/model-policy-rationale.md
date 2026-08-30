# Model policy — rationale

This file is the reasoning behind `skills/flow-contracts/model-policy.md`.
**A `/flow*` run never loads it — appendices are for whoever edits a contract.**

## Model policy

The field shape, and the rule that an absent key reads as *not recorded*, belong to
**State file** (`skills/flow-contracts/state-file.md`).

**This section is canonical for the model roles, their defaults and how an override applies.** One
location, named here rather than left to be worked out: every `/flow*` command is required to
load this file before acting, and none of them loads `<agents repo>/spectre/specs/`, so the file runtime actually
reads is the file the rule has to live in. **State file** (`skills/flow-contracts/state-file.md`)
cites this section for the `models` field rather than defining the roles a second time, and
`<project>/CLAUDE.md` and `<project>/AGENTS.md` name this section for the same reason.

**There is no requirements layer above this one; change this section.** The rules behind it were
first written as the capability `myflow-model-policy`, whose **Requirement: Implementer subagents run on the strongest available model** (`<agents repo>/openspec/specs/myflow-model-policy/spec.md`) anchored the defaults. That capability was frozen with the rest of the `<agents repo>/openspec/` tree at the spectre cutover and not migrated, so it records where the defaults came from and governs nothing. The two-layer split it was half of has ended everywhere, including at **Planning effort** (`skills/flow-contracts/state-file.md`), which lost the same layer for the same reason: this section is now the requirement as well as the operational form the commands read. The split's second argument died with it — a live spec used to be behind by construction while a change was open, because its delta reached the specs tree only at finish run 2, and under spectre a change edits the specs tree directly on its branch, so no spec lags a change any more.

The citation is still a **checked** one, not a courtesy: the guard associates a bold token with the path
beside it and matches it against the target's headings, a `### Requirement: …` heading
is a heading like any other, and the frozen file resolves — so naming the requirement in full is what makes
`<agents repo>/scripts/check-references.sh` fire if the heading is ever renamed. A bare backticked path with no bold token beside
it is **not** checked and rots silently — which is what this bullet's predecessor did.

The two rules point in opposite directions on purpose. A reviewer's job is to be many independent
readings of a finished diff, so its cost must not scale with the operator's session model. An
implementer's job is to get the diff right the first time, where capability compounds.

**Implementer subagents dispatched by `/myflow-do` run on Opus** (or the harness's strongest
available model). This **explicitly overrides** superpowers:subagent-driven-development's model
guidance — that skill says to pick "the least powerful model that can handle each role" and to use
the cheapest tier where the plan already contains the code to write. That guidance does not govern
this pipeline. The saving it offers is false here: an implementation defect is not avoided by the
review panel, it is *found* by the panel, at the cost of a fix wave and a re-run of every slot.
Buying a cheaper implementer with a more expensive review is the wrong trade.

**Two further instructions in that same upstream skill are also overridden, and are named here
rather than left to be discovered.** subagent-driven-development says to dispatch the *final
review* on the most capable model: flow does not — it fixes every panel slot at the panel's
model, Sonnet by default, for the reason above, and escalates the panel's **breadth** instead (the conditional Security, Adversarial
and extra-principle slots), which buys more independent readings rather than one stronger one.
It also says to *escalate the model in fix rounds 4-5*: flow cannot, because its implementers
already sit at the ceiling from round 1. Fix rounds escalate the same way — more slots, not a
bigger model — and round 5 hands back to the operator rather than pretending an escalation is
available.

**The panel-fix default is the strongest available model, and deliberately not Sonnet.** The role
names the agent that *applies* a fix, which is an implementer — so the implementer rule above
already governs it. Fix rounds escalate the panel's breadth rather than its model precisely because
implementers sit at the ceiling from round 1, and a fix-wave default of Sonnet would contradict both
of those rules at once.

**Every subagent dispatch records the model it used** in the SDD ledger, alongside the task it ran.
A model policy that nothing records is a policy nothing can verify — and the absence of that record
is precisely how this rule came to be missing in the first place.

**This record outlives the change, because it is a row in the store rather than a file in a
worktree.** Each dispatch is written as it closes, and the store's rows are the terminal record — no
worktree removal reaches them. The ledger rendered under `<project>/docs/superpowers/ledgers/` at
run 1 is a readable copy of those rows, for a reader with no daemon running; it is not the record.
**Rows also make the question answerable across changes, by query** — which model ran a given role,
over every change the store holds — where a preserved file answered it only for the one change whose
file you opened. The ledger is authored under `<abs-worktree>/.superpowers/`, which is
gitignored, in a worktree `/myflow-finish` run 2 removes — but run 1 preserves it into the
repository first, under `<project>/docs/superpowers/ledgers/`, so it serves the operator and the panel
*during* the change and stays answerable afterwards. An after-the-fact audit of which model
implemented which task therefore reads the preserved ledger rather than a transcript nobody kept.
The preservation duty itself is stated once, under
**Run 1 — the branch is not merged** (`skills/flow-contracts/finish-contract-run1.md`).

Slots dispatched by `subagent_type` (Bugbot, Security Review)
carry their own agent definitions and take no override from either mechanism.

Durability is a **stronger** reason to leave an unobserved entry unobserved, not a weaker one. A
persisting record makes an invented model slug permanent, so `unknown (agent-defined)` stays exactly
as written and no step fills it in on the way into the repository.
