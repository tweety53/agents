# kan-297-myflow-panel-findings-are-never-marked-fixed

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

Three tasks: the guard that catches an unclosed finding, the prose that closes findings so the guard
has nothing to catch, and the wiring that puts the guard in the panel's path. `design.md` is
canonical for the decisions; `proposal.md` for why the change exists.

**Baseline, measured before any edit:**

- 43 harnesses matched by `scripts/test-*.sh`.
  <!-- measured: ls scripts/test-*.sh | wc -l @ 69d0c0c -->
- 26 guard symlinks under `skills/flow/scripts/`.
  <!-- measured: ls skills/flow/scripts/ | wc -l @ 69d0c0c -->
- No `scripts/check-panel-findings-closed.sh` exists.
  <!-- measured: ls scripts/check-panel-findings-closed.sh @ 69d0c0c -->
- `skills/flow/review-panel.md` names `flow record status` exactly once — the sample at line 224,
  in the recording section, with no call site in the fix round's own procedure.
  <!-- measured: grep -c 'record status' skills/flow/review-panel.md @ 69d0c0c -->
- `stats/cmd/flow/record.go`'s `validateFindingStatus` already rejects a status outside
  `open`/`fixed`/`withdrawn <reason>`; KAN-297's second defect needs no work.
  <!-- verified: sed -n '45,56p' stats/cmd/flow/record.go @ 69d0c0c -->

- [x] 1. `scripts/check-panel-findings-closed.sh` — the guard, with its harness

The gate `design.md`'s `gate-is-a-guard` decision chose. Built to `scripts/check-panel-reproducers.sh`'s
shape, because it reads the same store array for the same change at the same point in the run.

```text unverified:confirm the argument order matches check-panel-reproducers.sh's once the guard exists
check-panel-findings-closed.sh <worktree> <change-name>
```

Carry, from `check-panel-reproducers.sh`, in its own words rather than by sourcing: `export LC_ALL=C`,
the change-name containment allowlist as an enumerated `case` (never a collating range), and the
`-d "$WORKTREE"` check. Per `design.md`'s `duplicate-the-predicate`, the open-finding jq predicate is
this guard's own copy, not a `scripts/lib/` helper.

**Capture stdout and stderr separately** on the store read — never `2>&1`. `flow` prints
`flow: using FLOW_ADDR=…` to stderr whenever the address is overridden, and folding that into the
payload puts a non-JSON line at the head of what `jq` parses, breaking a run that succeeded.
`check-unfinished-work.sh`'s own comment on this is canonical; do not restate it, cite it.

Exit codes, and nothing else may exit:

- **0** — no finding's status is open.
- **1** — one or more are; each still-open ref is named on stderr.
- **2** — cannot answer at all: no worktree, no change name, a name outside the allowlist, a
  worktree that is not a directory, the store unreachable, or `jq` failing. Per `design.md`'s
  `no-journal-excuse`, a close that only reached the local journal is **not** an exemption — it
  reads as open and exits 1.

Write `scripts/test-check-panel-findings-closed.sh` **first, RED before GREEN**, covering every row
of `design.md`'s **Testing** table and all three of its mutation bullets. Follow
`scripts/test-check-panel-reproducers.sh` for how it stubs `flow` on `PATH` to return a fixture
array without a live daemon.

**Files:** `scripts/check-panel-findings-closed.sh`, `scripts/test-check-panel-findings-closed.sh`
**Tests:** `scripts/test-check-panel-findings-closed.sh`
**Regression:** reverting this commit removes the only check that a review panel closed what it
verified, returning the failure to where KAN-297 found it — discovered at `/flow`'s integrate gate,
after the work is done, as an OUTSTANDING verdict naming every finding at once.
**Baseline:** before=43 after=44 harnesses matched by `scripts/test-*.sh`
<!-- predicted: ls scripts/test-*.sh | wc -l after task 1 -->
**Commit:** `feat(scripts): refuse a panel close while a finding is still open`
**Build:** green

- [x] 2. `skills/flow/review-panel.md` — close each finding where the round verifies it

The defect itself, per `design.md`'s `close-at-verify-point`.

The fix round's verification paragraph already requires the reproducer to re-run to exit 0 **and**
the fix's diff to touch a path the finding named. Extend that paragraph so that a finding meeting
both conditions is recorded closed there and then:

```bash verified:flow record status's flag set, against `flow record`'s own usage output
flow record status -change <name> -ref F<n> -status fixed
```

State the three properties that placement carries, since each is a thing a reader would otherwise
get wrong: the **parent** records it and never the fix subagent, because the parent is what ran the
reproducer and walked the diff; it is recorded **per finding as the verdict is reached**, not
batched at the round's end, so an aborted round still leaves verified findings closed; and a finding
failing either condition is **left untouched** on `open` for the existing handback.

Replace the sentence at line 277 — "Fix subagents record `fixed` when they fixed it, and leave
`open` when they did not." It assigns the close to the subagent and would contradict the paragraph
above. The neighbouring sentence about a `withdrawn` marker's reason is unaffected and stays.

**Do not** add a bulk close, a new CLI verb, or a `-round` argument; `design.md`'s `no-bulk-close`
records why that shape is wrong.

**Files:** `skills/flow/review-panel.md`
**Tests:** **none** — this task changes skill prose, whose verification is
`scripts/check-references.sh` and `scripts/check-vocabulary.sh`; the behaviour it describes is
checked by task 1's guard and its harness.
**Regression:** reverting this commit restores the state KAN-297 reports — a fix round that repairs
and verifies a finding, then leaves it recorded `open` forever.
**Baseline:** before=44 after=44 harnesses matched by `scripts/test-*.sh`
<!-- predicted: no harness is added or removed by a prose edit -->
**Commit:** `fix(flow): close a panel finding where the fix round verifies it`
**Build:** green

- [x] 3. Wire the guard into the panel's path, and correct the stale guard description

The symlink and the invocation **must land in the same commit**: `scripts/check-guard-symlinks.sh`'s
rule 2 requires a symlink in a skill's own `scripts/` directory for every guard that skill's text
invokes by name in a fence, so either half alone fails that guard.

- `skills/flow/scripts/check-panel-findings-closed.sh` — a **relative** symlink to
  `../../../scripts/check-panel-findings-closed.sh`, matching every sibling there. Rule 1 rejects an
  absolute target.
- `skills/flow/review-panel.md` — invoke it immediately before the stage's own
  `flow stage end -command '/flow' -stage flow.review-panel`, in the same fenced form
  `check-panel-reproducers.sh` is invoked in at line 330.
- `skills/flow/SKILL.md` — add the basename to the guard-presence list at line 168, in the
  alphabetical position the list already keeps.
- `skills/flow/review-panel.md` lines 268-276 — the passage stating `check-unfinished-work.sh`
  "reads **only the marker block**" and enumerating the five shapes it parses. That guard reads the
  store through `flow record findings` and parses no Markdown at all; its own header records why the
  parser was removed. Replace the passage with what is now true. Leave the marker block's own
  rendered shape documented — the renderer still emits it; only the claim about who parses it is
  wrong.

`.flow/project.md` needs no edit: `scripts/run-guard-tests.sh` discovers `scripts/test-*.sh` by glob.

**Files:** `skills/flow/scripts/check-panel-findings-closed.sh`, `skills/flow/review-panel.md`,
`skills/flow/SKILL.md`
**Tests:** **none** — this task adds a symlink and skill prose; its verification is
`scripts/check-guard-symlinks.sh` rules 1 and 2, plus `scripts/check-references.sh`.
**Regression:** reverting this commit leaves task 1's guard shipped but never invoked, so a panel
can still end with an open finding in the store and nothing notices until integrate.
**Baseline:** before=26 after=27 guard symlinks under `skills/flow/scripts/`
<!-- predicted: ls skills/flow/scripts/ | wc -l after task 3 -->
**Commit:** `feat(flow): gate the panel close on every finding being closed`
**Build:** green
