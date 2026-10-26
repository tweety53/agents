# self-review-kan-1035-flow-self-review-analyse-visual-verify-prompts

## Why

The self-review pass of `kan-1035-flow-self-review-analyse-visual-verify-prompts`
(`.superpowers/sdd/self-review-pass.md` in this worktree) found three non-`big` defects in the
change's own text and the tooling it leans on:

- A. Angle 8 judges visual-verify findings by Critical/Important/Minor severity, but the verifier
  grades none (a departure blocks or not), and `check-prompt-tuning-cases.sh` refuses a case whose
  `source.slot` or `source.severity` is empty — so every visual-verify case would carry a made-up
  severity and slot.
- B. A `reviewer` case targeting the verifier prompt can never reproduce under
  `/flow-prompt-tuning`'s replay, which runs a reviewer over `bad.diff`/`good.diff`; the verifier
  judges a capture against a mockup, and no case file carries either.
- C. `/flow-fast` section 5 tells the run to "Run every command in `## lint`", and the list
  `project-get.sh <root> lint` resolves carries its triple-backtick fence lines; the run hand-strips
  them, and the reviewed run lost one iteration running the fence lines.

## What changes

- Angle 8 states that a visual-verify departure is judged as an Important and that its case
  records `source.slot` `visual-verify` and `source.severity` `Important`.
- A verifier-target case carries `capture.png` and `mockup.png`, and `/flow-prompt-tuning`
  replays it with a verifier dispatch judging that capture against that mockup.
- `/flow-fast` section 5 runs the lint list through the existing `scripts/run-verify-list.sh`,
  which already runs only the fenced commands, instead of by hand.
