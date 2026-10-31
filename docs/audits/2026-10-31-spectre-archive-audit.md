# Spectre archive audit — 2026-10-31

What `spectre/changes/archive/` holds in every repository flow has run in, who reads it, what it
costs, and what replaces it. Every number below sits beside the command that produced it; all were
measured on 2026-10-31 against each repository's main checkout (`~/Projects/<repo>`), read-only.

## Scope

Six repositories carry a `spectre/changes/archive/` directory: agents, gymie, gymie-frontend,
gymie-playwright, flow-eval and spectre. gymie-admin-frontend and intellij-review-queue carry none.

```bash
for r in agents gymie gymie-frontend gymie-playwright flow-eval spectre gymie-admin-frontend intellij-review-queue; do
  [ -d ~/Projects/$r/spectre/changes/archive ] && echo "$r archive" || echo "$r none"; done
```

## Size

| Repository | Archive (working tree) | Archived changes | PNG bytes (share) | `spectre/` blobs in history (on disk) | `.git` storage |
|---|---|---|---|---|---|
| agents | 104M | 364 | 86.5M (86%) | 99.1M | 278M |
| gymie | 638M | 127 | 603.6M (97%) | 530.8M | 726M |
| gymie-frontend | 356K | 39 | 0 | 0.1M | 35M |
| gymie-playwright | 112K | 18 | 0 | <0.1M | 562M (all loose — see below) |
| flow-eval | 472K | 5 | 0 | 0.2M | 55M |
| spectre | 220K | 4 | 0 | 0.1M | 7.2M |

The share is PNG bytes over the archive's file bytes (agents 100.5M, gymie 620.0M). Commands, with
`d=~/Projects/<repo>`:

- Archive (working tree): `du -sh $d/spectre/changes/archive`
- Archived changes: `ls $d/spectre/changes/archive | wc -l`
- PNG bytes: `find $d/spectre/changes/archive -name '*.png' -exec stat -f %z {} + | awk '{s+=$1}END{printf "%.1fM",s/1048576}'`
- Archive file bytes (the share's denominator): the same with `-type f` in place of `-name '*.png'`
- `spectre/` blobs in history: `git -C $d rev-list --objects --all -- spectre/ | cut -d' ' -f1 | git -C $d cat-file --batch-check='%(objecttype) %(objectsize:disk)' | awk '$1=="blob"{s+=$2}END{printf "%.1fM", s/1048576}'`
- `.git` storage: `du -sh $d/.git`

### gymie-playwright's loose storage

`git -C ~/Projects/gymie-playwright count-objects -v` reports 10912 loose objects, 574180 KiB, and
`packs: 0`: the repository has never been packed, so every object version is stored whole and
zlib-compressed only. Of the 504.7M of blob bytes on disk
(`git -C $d rev-list --objects --all | git -C $d cat-file --batch-check='%(objecttype) %(objectsize:disk) %(rest)' | awk '$1=="blob"{s[$3]+=$2;t+=$2}END{for(k in s)print s[k],k; print t,"TOTAL"}' | sort -n | tail`),
205.6M is `full-app-suite.zip`, a ~3.4M tracked archive rewritten by 101 commits
(`git -C $d log --all --oneline -- full-app-suite.zip | wc -l`); the rest is regression-suite
snapshot PNGs. None of it is spectre: its `spectre/` blobs are under 0.1M. A `git gc` would delta the
zip versions; that is the repository's own housekeeping, not this audit's.

## Commits

| Repository | Base | Non-merge commits | `chore(spectre)` | `docs(self-review)` | Touching only `spectre/` or `docs/self-review/` |
|---|---|---|---|---|---|
| agents | origin/main | 3704 | 907 (24.5%) | 426 (11.5%) | 1249 (33.7%) |
| gymie | origin/develop | 1451 | 739 (50.9%) | 133 (9.2%) | 867 (59.8%) |
| gymie-frontend | origin/develop | 290 | 75 (25.9%) | 1 | 63 (21.7%) |
| gymie-playwright | origin/main | 165 | 33 (20.0%) | 1 | 33 (20.0%) |
| flow-eval | origin/main | 58 | 12 (20.7%) | 5 (8.6%) | 17 (29.3%) |
| spectre | origin/main | 112 | 10 (8.9%) | 2 | 6 (5.4%) |

Every share is over the non-merge count (a merge lists no files). Commands, with `b` the base:

- Base: `git -C $d symbolic-ref --short refs/remotes/origin/HEAD`
- Non-merge commits: `git -C $d rev-list --no-merges --count $b`
- `chore(spectre)`: `git -C $d log --no-merges --format=%s $b | grep -c '^chore(spectre)'`
- `docs(self-review)`: `git -C $d log --no-merges --format=%s $b | grep -c '^docs(self-review)'`
- Touching only `spectre/` or `docs/self-review/`: `git -C $d log --no-merges --format='@%H' --name-only $b | awk '/^@/{if(n&&ok)c++;n=0;ok=1;next} NF{n++; if($0!~/^(spectre\/|docs\/self-review\/)/)ok=0} END{if(n&&ok)c++; print c}'`

## Consumers

Cited at the merge base `9bf31b17e2a4c88620ad5da6043fb1d515d87c85` (`git merge-base HEAD origin/main`
on the change branch), before this change's tooling tasks; found with
`git grep -n -e archive 9bf31b17 -- <path>`.

- **Readers of past changes' archives.**
  - `flow lesson resolve` scans `spectre/changes/archive/*/narrative.md`
    (`stats/internal/lessons/lessons.go:74`, used at `:125`).
  - spectre-planning's rejection search reads archived decisions
    (`skills/spectre-planning/SKILL.md:34-37`).
- **Readers of archives of changes still in flight.** The unlanded-archive scan
  (`stats/internal/guard/unlandedarchives.go:14-23`) and the plan-shape, build-green, plan-class and
  provenance guards that call it.
- **Readers of the just-archived change only.**
  - The self-review bundle's `git show` reads (`stats/internal/selfreview/git.go:8-36`).
  - Every "live path, else `archive/<name>`" fallback: `guard.ChangeDir`
    (`stats/internal/guard/changeplan.go:36`), the pipeline-fix key scan
    (`stats/internal/guard/pipelinefix.go`), plan-unchanged (`stats/internal/guard/planunchanged.go`),
    the record checkout check (`stats/cmd/flow/record.go:3785-3843`), the visual-verify record's
    "archive/ twin" (`stats/internal/guard/visualverifydispatched.go:336`), and cleanup's
    moved-never-deleted check.
- **No reader after landing.** `visual-verification/` PNGs, `.frame.png`, `matrix/*.json`,
  `evidence/` (readers are in-run only: `stats/internal/guard/visualverifypass_close.go:274`,
  `stats/cmd/flow/record.go:1279`); self-review `patches/task-N.diff`; `verbatim-moves.txt`.
- **flowd** holds dispatches, stage runs, findings, panel passes and mutations, self-review findings,
  learning notes, task counts and summaries — no plan text, no PNGs, no report prose.
  `ledger.md`/`panel.md` are renders of its rows (`stats/internal/records/render.go`).
- **`docs/self-review/`**: the `land` guard refuses until the report is committed at HEAD
  (`stats/internal/guard/integratechange_land.go:17,64`); nothing reads past reports.

## Verdicts

The rule (operator's direction): every change artifact is ephemeral — it exists while the run needs
it and is gone from the tree once the change lands. An artifact survives only if a named consumer
reads it after landing, and then it is served from flowd or history, not from a committed archive.

| Artifact | Verdict | Post-land consumer and its replacement |
|---|---|---|
| `proposal.md`, `design.md`, `tasks.md` | committed on the change branch during the run, deleted by the landing commit | rejection search → decisions recorded in flowd at the plan gate (follow-up); the text stays in history |
| `narrative.md` | deleted at land | `flow lesson resolve` → `docs/briefs/` plus flowd learning notes (follow-up) |
| `visual-verification.md`, `live-verification.md` | deleted at land | the self-review bundle, which reads them from history |
| `visual-verification/` PNGs, `.frame.png`, `matrix/*.json`, `evidence/` | never committed (follow-up) | none |
| `link.md` (canonical and satellite) | committed during the run, deleted at land in every repository | none — merge order is read before landing, or from history |
| `ledger.md`, `panel.md` | not committed | `flow record render` from flowd |
| self-review `patches/task-N.diff`, `verbatim-moves.txt` | leave with the change directory | none |
| `docs/self-review/<name>.md` report | not committed (follow-up) | the `land` guard checks flowd's self-review pass row |
| `docs/self-review/<name>-context.md` bundle | unchanged | already transient |
| `spectre/changes/archive/` (existing content) | one tip-only commit per repository deletes it | none; history keeps every file |

### Spectre surfaces

Call sites across `skills/`, `scripts/`, `stats/` and `rules/` at the merge base:
`git grep -n -E "spectre <verb>([^a-z-]|$)" 9bf31b17 -- skills scripts stats rules | wc -l`.

| Surface | Sites | Verdict |
|---|---|---|
| `spectre new` | 12 | keep |
| `spectre validate` | 12 | keep |
| `spectre list` (incl. `--specs`) | 9 | keep |
| `spectre link` | 18 | keep |
| `spectre archive` | 19 — all the archive step this change removes | cut (follow-up) |
| `spectre refs` | 0 | cut, with cross-tree requirement citations (follow-up) |
| `spectre init` | 0 — one-off bootstrap | keep |
| `peers` file | read by `link` | keep |
| `config.md` | present in gymie-frontend, gymie-playwright, flow-eval, spectre; 0 keys set outside the defaults fence (`git -C $d show HEAD:spectre/config.md \| awk '/^```/{f=!f;next} !f && /^- /' \| wc -l`) | no key to cut |
| `proposal.md` / `design.md` / `tasks.md` | read in-run by implementers, reviewers, visual-verify | keep the three files; ephemeral, so no post-land cost |
| `specs/` placeholders (agents, gymie-frontend, flow-eval: 1 file each — `git -C $d ls-files spectre/specs \| wc -l`) | no reader | delete (follow-up) |
| `specs/` in gymie: 33 files, 772K, 109 commits, last 2026-10-24 (`git -C $d ls-files spectre/specs \| wc -l`, `du -sh $d/spectre/specs`, `git -C $d rev-list --count HEAD -- spectre/specs`, `git -C $d log -1 --format=%ad --date=short -- spectre/specs`) | live; read by planners | keep |

## What this change did

- **The landing drop.** Integrate's run 1 no longer moves a change into
  `spectre/changes/archive/<name>/`. One commit, `chore(spectre): drop <name>`, deletes
  `spectre/changes/<name>/`, each `<name>-fix-<N>/` sub-change, and each satellite's `link.md`
  directory from the tip. The self-review sub-change and the pipeline-fix loop drop theirs the same
  way. A fix run after the drop restores the change directory from the drop commit's parent.
- **The collapsed readers.** Every live-else-archive fallback reads the live path. Readers that run
  after the drop — the self-review bundle, the merge-order read, the board's plan of a finished
  change, the record checkout check — read the change's files from history through
  `stats/internal/landed`. The unlanded-archive scans are cut; cleanup reports a change directory
  still at the tip; the pipeline-fix key scan counts landed keys from the log.
- **The tip cleanup.** One `chore(spectre): remove the change archive` commit per repository deletes
  `spectre/changes/archive/` in agents, gymie, gymie-frontend, gymie-playwright, flow-eval and spectre.
  History keeps every file.

## What it left

- This change's own run 1 integrates with the main checkout's installed tooling, which still
  archives. After it lands, `spectre/changes/archive/kan-1072-audit-whether-archived-spectre-change-artifacts/`
  remains in agents and an `archive/<name>/link.md` remains in each satellite.
- Until the follow-up lands, `flow lesson resolve`'s narrative scan and the rejection search find
  nothing at the tip.

## History cleanup — recommended against

Rewriting history, for example
`git filter-repo --path-glob 'spectre/changes/archive/*/visual-verification/*' --invert-paths`,
would reclaim up to the `spectre/` blob bytes above: ~531M of gymie's 726M and ~99M of agents' 278M.
The procedure: rewrite with `git filter-repo`, force-push every branch, re-clone every checkout and
recreate every worktree. It is irreversible once the old refs are gone, and it breaks every cited sha
— Jira comments, `docs/`, flowd's `commit_sha` and `round_base` rows, narratives. The tip deletion
fixes working-tree size, not clone size. Nothing rewrites history.

## The umbrella follow-up

One task carries what this change does not do.

1. **Visual evidence out of git.** Gitignore `visual-verification/` PNGs, `.frame.png`,
   `matrix/*.json` and `evidence/` in the worktree and remove them at cleanup;
   `visual-verification.md` keeps its text. Why: PNGs are 86% of agents' archive bytes and 97% of
   gymie's, and no reader opens them after landing.
2. **Self-review report out of git.** Stop writing and committing `docs/self-review/<name>.md`; the
   `land` guard checks flowd's self-review pass row; `check-self-review-report.sh` drops its
   committed-report check. Why: no reader of past reports exists, findings already live in flowd, and
   the `docs(self-review)` commits are 11.5% of agents' non-merge history.
3. **Lessons and rejection search off the archive.** `flow lesson resolve` drops its narrative scan
   and reads `docs/briefs/` plus flowd learning notes; each `## Decisions` entry is recorded in flowd
   at the plan gate and the spectre-planning rejection search queries flowd. Why: both read the
   archive, which is gone from the tip, and briefs are the lessons home by contract; flowd keeps the
   kan-970 safeguard against designing a twice-rejected remedy again.
4. **Spectre CLI cuts.** Remove `spectre archive`, `spectre refs` and cross-tree requirement
   citations from spectre, with its README and its vendored `spectre-planning` copy; delete the
   placeholder `specs/` trees in agents, gymie-frontend and flow-eval. Why: zero call sites, or call
   sites only in the archive step this change removed.
5. **The residual archives.** Delete
   `spectre/changes/archive/kan-1072-audit-whether-archived-spectre-change-artifacts/` in agents and
   each satellite's `archive/kan-1072-audit-whether-archived-spectre-change-artifacts/` — left by this
   change's own landing, which ran on the tooling installed before it.
