# kan-1027-flow-learning-setting

## Why

The operator wants each implementation run to teach something: a short piece of software
engineering theory tied to the task at hand — its language, its design approaches, agentic coding,
LLMs and AI infrastructure — with a coding puzzle where one fits and links to study further, never
the same topic more than three times (KAN-1027).

## What changes

- A harness-wide `learning` setting, on by default, beside the reviewer slots in `flow_settings`:
  `flow settings set -learning on|off`, shown and changed by `/flow-settings`.
- A `flow_learning_topics` store table recording every note shown — topic slug, one-line note,
  change, time — with `flow learning add` (refusing a topic already shown three times) and
  `flow learning list` (each topic with its count).
- With the setting on, `/flow`'s implementation run and `/flow-fast` compose a learning note at
  implementation start and print it as a `## Learning` section of their final message; fix runs
  do not.
