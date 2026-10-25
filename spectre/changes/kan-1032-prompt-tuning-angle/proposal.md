# kan-1032-prompt-tuning-angle

## Why

- Many `/flow` runs spend several review-panel fix rounds on defects of a recurring class. Either an
  implementer instruction would have prevented the bad code, or a reviewer instruction would have
  caught it (or not raised a false positive) in round 1. Nothing today turns those rounds into
  instruction changes: the self-review pass looks at the pipeline, its cost, its speed and eval
  candidates, but never at whether the implementer and reviewer prompts themselves have a generic
  gap.
- Patching a prompt from one vivid failure produces incident-specific accretion (prompt-audit's "patch
  accretion" and "overgeneralizing from one case"). A proposed prompt update therefore has to be
  generic, saved with the data that reproduces the problem, and proven by an old-vs-new replay before
  it is applied.

## What changes

- An eighth self-review angle, `flow-prompt-tuning`, runs in the same single pass as the other
  seven, in `/flow`'s integrate, `/flow-fast`'s verify and standalone `/flow-self-review`. It picks
  the review-panel findings of the run whose root cause is a generic gap in the implementer or
  reviewer instructions. Each finding qualifies on its own severity: a Critical one, an Important one, or a
  Minor one judged genuinely useful.
  A `reviewer` case is a miss (caught only in a later round or by a later slot) or a false positive
  (a finding later disproved or withdrawn). Zero cases is a normal result.
- Each picked case is saved to `<agents repo>/docs/prompt-tuning/<case-id>/`: `case.json`, the
  finding, the bad and the good code, the task text, why it was picked, the proposed prompt update
  before and after, and an `index.html` rendered from those files by a new
  `flow prompt-tuning render` command. The angle never fixes anything, never offers a case for
  filing and never asks the operator. Its report lines are `noted` and name the case folder.
- A new guard, `scripts/check-prompt-tuning-cases.sh` (Go, in flow-guard), keeps every case
  complete and generic: the proposed text names no change, Jira key or repro file, and the
  "before" text exists word for word in the target prompt.
- Each case carries a `scope` on two axes, language and side (`fe`/`be`), each pinned or open:
  `generic` targets the shared prompts under `skills/`; `language:<lang>`, `side:<fe|be>` and
  `language:<lang>/side:<fe|be>` target `rules/lang-<lang>.mdc`, `rules/side-<fe|be>.mdc` and
  `rules/lang-<lang>-<fe|be>.mdc`, which projects opt into through `.flow/project.md`'s
  `## standards`. A rule true on both sides leaves the side open.
- A new standalone command, `/flow-prompt-tuning`, walks the pending cases and groups those that
  propose the same generic update. For each group it A/B-replays the update on the cases' repro
  data (3 runs per arm on opus), applies the updates that pass, records every case in
  `docs/prompt-tuning/LEDGER.md`, deletes the case folders and lands the result through the agents
  repository's default landing route.
