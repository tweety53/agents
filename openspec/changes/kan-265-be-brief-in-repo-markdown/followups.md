# Follow-ups found during kan-265, not fixed here

Both are **pre-existing** defects in `skills/myflow-contracts/finish-contract.md`, confirmed present
at `764eb77` — before this change's first trim commit — by task 4.1's implementer and independently
by its reviewer. Neither is fixed here: this change is cuts-only, and fixing either requires
rewording, which decision `cuts-not-paraphrase` prohibits.

File these at `/myflow-finish` run 1.

## 1. § Worktree cleanup cites the wrong check number

The bullet reads "This includes check 4: a `## stop` command that…". The `## stop` command is
**check 5**. Check 4 is the ignored-files disclosure, which the same section states explicitly is
*not* a gate — so the bullet contradicts the paragraph four lines above it.

Consequence: a reader following the bullet would treat the ignored-files disclosure as a blocking
gate, which is the opposite of what the contract requires, and `/myflow-fast` overrides that
disclosure by design.

## 2. The file's last line carries a stale count

"the check is **skipped, not failed**, and cleanup proceeds on the strength of the other two" —
there are **four** other checks, not two. A count left behind from when the list was shorter.

## 3. Two live requirements contradict each other about when commits happen

Found by task 6.1, which could not cut either: both are normative, and section 3 forbids cutting a
normative sentence. Fixing this is a decision about which requirement is correct, not an edit.

`openspec/specs/myflow-command-surface/spec.md` § Git actions are bounded by state:

- "`/myflow-do` … SHALL commit and push **only** when the state file already records a `prUrl`"
- "No command other than `/myflow-finish`, and `/myflow-do` when a PR is already open, SHALL create
  a commit."

Contradicted by `openspec/specs/myflow-task-commits/spec.md` § Each task commits after finishing,
before review, and by `skills/myflow-contracts/pipeline.md` § Git boundaries — both of which have
`/myflow-do` committing every task.

The pipeline as it actually runs follows the second pair: this change's own tasks each committed
without any `prUrl` recorded. So `myflow-command-surface` is the stale side, but retiring a
requirement is a spec decision.

The same rot in prose form was cut from `skills/README.md` by task 4.3, which is evidence it
propagated from here.

## 4. `agents-repo-verification` names four guards where the project declares sixteen

`openspec/specs/agents-repo-verification/spec.md` § The repository's guards run during
implementation enumerates four guards as the project's lint and test set and closes "Four guards
remain." `.myflow/project.md` declares **16** `## lint` entries and **33** `## test` entries.

The enumeration is a `SHALL`, so task 6.1 left it. "Four guards remain." is not normative, but no
file carries "four guards" any more — it is a stale fact with no holder, not restatement, so the
naming rule gave task 6.1 nothing to point at. Its own scenario, never cut, still says "all four
scripts run".

## 5. Canonical always-on text still says `/myflow-do` stages a diff

Found by task 7.1, which cut the *copies* of this claim from `README.md`, `CLAUDE.md` and
`AGENTS.md` but could not touch the canonical statements.

"Review the staged diff" survives in `rules/myflow-manual-review.mdc` (`alwaysApply: true`, so in
every session's prompt), in the three-line pipeline digest, and in every gate row that describes
`IN_PROGRESS`. But `/myflow-do` commits each task, and `pipeline.md` § Git boundaries plus
`skills/myflow-do/SKILL.md` § 7 both say `do.stage-diff` "only confirms nothing slipped in, it does
not stage anything itself".

So the always-on rule tells every session to review something the pipeline no longer produces. This
is not a rotted copy — it is the canonical language lagging the pipeline, which is why no trim task
could fix it: correcting it is a reword.

## 6. `README.md`'s global-install table omits `~/.codex/skills/`

Found by task 7.1. It cut the README's claim that "nothing is installed under `~/.codex/`", which is
false — `setup.sh:775` runs `install_skills "$home_dir/.codex/skills"`, and `AGENTS.md` confirms a
Codex session gets skills. But the README's own install table, in § Global install, still lists the
global targets without `~/.codex/skills/`.

Left because the table is the README's own reference, not a copy of a holder the naming rule could
point at — the same shape as `skills/README.md`'s contracts list, which task 4.3 could cut only
because `myflow-manual-review.mdc` held the canonical version.

## 7. `stats/README.md` is outside the owned corpus, and probably should not be

Found while task 8.1 reconciled a count gap. The corpus is defined by enumerated roots — `skills/`,
`rules/`, `openspec/specs/`, `commands/`, `commands-claude/`, `.myflow/` and the repository root
(non-recursive). `stats/README.md` therefore sits outside it: it is not a record, it is
documentation, and nothing in this change's design argues for excluding it.

It was **not** added here. Widening the corpus mid-change would move `normative-baseline.txt`, and
that baseline is what all seven completed trim tasks were verified against — invalidating them to
pick up one file is the wrong trade.

`docs/self-review/*.md` (21 files) also sit outside, and that one is deliberate: they are session
records, the same class as `docs/superpowers/`, and trimming a record rewrites history.

Adding `stats/README.md` means widening `scripts/lib/owned-corpus.sh`, re-capturing the baseline,
and giving the file a budget row.
