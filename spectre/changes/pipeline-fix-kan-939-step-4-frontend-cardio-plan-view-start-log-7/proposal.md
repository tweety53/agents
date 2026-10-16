# pipeline-fix-kan-939-step-4-frontend-cardio-plan-view-start-log-7

## Why

`hooks/protect-main-checkout.py` refuses two kinds of landing push — one that lands an open
`spectre/changes/<change>/`, and a fix loop's landing that flowd's records do not cover
(`scripts/check-loop-landing.sh`). Both read the push's repository and source refspec from the
literal Bash command text. A push written through a shell expansion the hook cannot resolve —
`git -C $W push origin $B:main`, `"${B}:main"`, `$(git branch --show-current):main`, backticks —
leaves the hook with no repository to inspect or an unresolvable source, so both checks return
nothing and the push goes through. On 2026-10-16 that is how a `pipeline-fix-*` branch with no
flowd records landed on the agents repository's `main` (0b055e56).

## What changes

- The hook denies a `git push` whose repository (`-C <path>` or a `cd`), remote or refspec goes
  through an unexpanded shell expansion (`$`, `${…}`, `$(…)`, a backtick) when any of its
  destinations is the default branch, is itself an expansion, or cannot be determined. The deny
  reason asks for the push to be written with literal names.
- A push whose every destination is a literal branch other than the default one still passes, so
  branch-backup pushes typed with a variable path (`git -C $W push origin spectre/<name>`) keep
  working; an assignment the command itself makes (`B=x; git push origin $B:main`) is expanded
  first, as before, and judged by the existing checks.
- Scripts are unaffected: the hook sees only the typed Bash command, never what a script it runs
  does, so `land-self-review-report.sh`, `pipeline-fix.sh land` and the landing scripts keep their
  own pushes.
- The hook header documents the new check and its ceiling; `scripts/test-protect-main-checkout.sh`
  gains cases that fail on the untouched hook.
