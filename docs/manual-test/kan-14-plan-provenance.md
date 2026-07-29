# Manual test guide — kan-14-plan-provenance

**Change:** kan-14-plan-provenance · [KAN-14](https://tweety53.atlassian.net/browse/KAN-14) — Plan provenance
**Worktree:** `/Users/tweety53/Projects/agents-worktrees/openspec-kan-14-plan-provenance`
**Branch:** `openspec/kan-14-plan-provenance`

Automated state at handoff: **595 harness assertions, 0 failures**; `check-vocabulary.sh` 0,
`check-references.sh` 0, `check-plan-provenance.sh` **exit 1 with exactly 25 hits — expected, and
explained in check 2**; `openspec validate --strict` clean for kan-14 and kan-8.

**This repository has no runnable application.** It is the source of the myflow skills, commands and
rules that `setup.sh` installs elsewhere. So there is nothing to launch — the checks below are
scripts you run, plus one section that can only be verified by running the pipeline itself.

```bash
cd /Users/tweety53/Projects/agents-worktrees/openspec-kan-14-plan-provenance
```

---

## 1. The declared lint and test sets

- [x] `scripts/test-check-plan-provenance.sh` → all pass. This is the guard's own harness.
- [x] `scripts/test-setup.sh` and `scripts/test-check-references.sh` → both pass, unchanged by this
      change.
- [x] `scripts/check-vocabulary.sh` and `scripts/check-references.sh` → both exit 0.

## 2. The one guard that is *expected* to fail

- [x] `scripts/check-plan-provenance.sh` → **exit 1**, exactly **25** hits, every one in
      `openspec/changes/kan-8-myflow-updates/tasks.md`.

```bash
./scripts/check-plan-provenance.sh 2>&1 | grep -c 'kan-8-myflow-updates/tasks.md'   # 25
./scripts/check-plan-provenance.sh 2>&1 | grep 'tasks.md:' | grep -v kan-8          # empty
```

This is the new guard firing on **another open change's** plan, whose fenced snippets nobody can now
retroactively attribute. They clear on their own when kan-8 archives, because the guard excludes
`openspec/changes/archive/`. **The fix is never to narrow the guard's scope or add a suppression
marker** — confirm you are comfortable shipping a lint set with one expected-red guard, or say so
and it becomes follow-up work.

## 3. The guard's behaviour — the cases that cost the most to get right

Each of these was a shipped defect at some point in this change's review. Run them against a
sandbox root so nothing touches the repo:

```bash
S=$(mktemp -d); mkdir -p "$S/openspec/changes/demo"
# write a fixture to "$S/openspec/changes/demo/tasks.md", then:
CHECK_PLAN_PROVENANCE_ROOT="$S" ./scripts/check-plan-provenance.sh; echo "exit=$?"
```

- [x] **A tagged, correctly closed fence passes.** A block whose info string carries
      `verified:<how>` and a numeric claim with a `measured:` comment within two lines → exit 0.
- [x] **An untagged fence is reported**, naming its line.
- [x] **A number with no provenance comment is reported**, naming its line.
- [x] **Ambiguous indentation refuses rather than guessing.** A fence run indented 4+ columns with
      no container marker on its own line → **exit 4**, with a message naming the cause and the
      remedy. This is the deliberate fail-loud choice; confirm the refusal message would tell you
      what to change.
- [x] **A `tasks.md` that is a symlink is refused** → exit 3, and the file is never opened.
- [x] **A change directory that escapes the repo is refused** → exit 3.
- [x] **`openspec/changes` itself being a symlink is refused** → exit 3. (Before the last fix wave
      this returned exit 0, "nothing in flight", over a repo containing an unattributed plan.)

## 4. The exit-code contract

The guard's whole value is that a caller can tell these apart. Confirm each is distinguishable:

| Code | Means | How to trigger |
|------|-------|----------------|
| 0 | clean, or nothing in flight | a sandbox root with no changes |
| 1 | real violations found | the repo itself (check 2) |
| 2 | environment — cannot run | `CHECK_PLAN_PROVENANCE_ROOT=` (empty) |
| 3 | containment / security | the symlink cases in check 3 |
| 4 | content-classification — refused to guess | the 4+ column case in check 3 |

- [x] **Violations elsewhere are still reported when something else fails.** Put a real violation in
      one change directory and an unreadable file in another → the violation is still printed.
- [x] **A containment refusal outranks everything** → exit 3, but the other directory's violations
      are still listed above it.
- [x] **A broken (not absent) `python3` exits 2, not 1.** Put a shim on `PATH` that exits non-zero;
      the guard must not report "violations found" when it never ran.

## 5. What only running the pipeline can verify

The skill and contract edits cannot be checked by any script. These are the real reason this guide
exists.

- [x] **`/myflow-start` tags what it writes.** Run it on a throwaway idea. Every fenced block in the
      generated `tasks.md` carries `verified:` or `unverified:`, every number carries `measured:` or
      `predicted:`, and the guard passes on the plan it just wrote.
- [x] **An `unverified:` block reaches the implementer as a hypothesis.** During `/myflow-do`,
      confirm the implementer dispatch carries the standing clause telling it to establish the real
      API first and report what it found — rather than transcribing the snippet.
- [x] **Implementers are dispatched on Opus** (the policy added in this change), and the panel slots
      on Sonnet. Watch an actual `/myflow-do` run.
- [x] **The ledger records the model per dispatch**, with `unknown (agent-defined)` for the two
      slots whose model the dispatcher genuinely cannot observe.
- [x] **`commands/myflow-do.md` no longer contradicts the policy.** It previously said "Opus is
      reserved for `/myflow-start`'s brainstorming stage", which this change makes false.

## 6. Known and deliberate — please confirm these are acceptable

- [x] **The guard scans `openspec/changes/*/tasks.md` and nothing else.** Not `design.md`, not
      `proposal.md`, not `specs/`. That narrow scope is deliberate — but it is also why several
      defects in *this change's own* `design.md` survived many review passes.
      **Resolved: widening approved as follow-up — [KAN-19](https://tweety53.atlassian.net/browse/KAN-19).**
      Ships as-is here.
- [ ] **`measured:`'s `@ <ref>` half is advisory, not enforced.** The contract defines the shape;
      the guard checks only that a `measured:`/`predicted:` comment is present. Tightening it would
      fire on existing plans.
- [x] **The SDD ledger's model record is session-scoped and does not survive archiving.** It lives
      under `.superpowers/`, which is gitignored and removed with the worktree. The spec states this
      plainly rather than claiming an audit trail it cannot provide.
      **Resolved: durability approved as follow-up — [KAN-19](https://tweety53.atlassian.net/browse/KAN-19).**
      The scope paragraph in `specs/myflow-model-policy/spec.md` stays accurate until that lands.
- [ ] **A `verified:` tag can still be wrong.** The guard enforces that provenance is *stated*, not
      that it is true. This change never claimed otherwise, and this is the residual risk.

---

## What automation could not establish

- **That `/myflow-start` actually tags well.** The guard checks a tag is present, not that
  `verified:javap …` names a check anyone ran. Check 5's first box is the only real test.
- **That the model policy is followed.** Nothing reads the ledger; the requirement is prose. Check
  5's third and fourth boxes are the only verification that exists.
- **That the fail-loud refusals are tolerable in practice.** They fire on zero of the 76 fence lines
  in this repository's two real plans, so their cost is unmeasured on real content. Check 3's
  exit-4 box is where you judge whether the message earns the refusal.
