# pipeline-fix-kan-urgent-5-oct18-20

## Why

A self-review loop run from the agents checkout files every dispatch row under the agents
project, yet a plan task it carries into another repository is committed on a satellite branch
of the same name there (**Pipeline defects found mid-run**, `skills/flow-contracts/pipeline.md`).
`scripts/check-loop-landing.sh <repo> <branch>` looks the rows up under the pushing repository's
project only, so landing the satellite — `git push origin HEAD:main` in a gymie or spectre
worktree, or `git -C <that worktree> push …` from the agents cwd — is refused by
`hooks/protect-main-checkout.py` for a loop whose plan, fix and clean last review are all on
record (self-review-kan-urgent-5-oct18, 2026-10-18). `pipeline-fix.sh land` lands agents
worktrees only, so the satellite had no sanctioned landing path.

## What changes

- `check-loop-landing` reads a loop branch's rows under the agents checkout's own project when
  the pushing repository's project holds none for it; a loop with rows nowhere is still refused.
- A change's own `spectre/<name>` branch keeps reading its own project alone.
- `scripts/check-loop-landing.sh` exports the checkout it sits in as `FLOW_GUARD_REPO_ROOT` and
  states the fallback in its header.
