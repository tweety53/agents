# Roles and models

Every agent the pipeline starts has a **role**: planner, implementer, panel reviewer, visual
verifier, and so on. A role is what decides which model the agent runs on and how hard it thinks.

## Pairs

A **pair** is a model plus an effort level. Each role has a list of pairs it may run on, kept in the
stats service and edited in the dashboard's **Agent roles** view. A pair is either:

- **base** — always allowed; or
- **opt-in** — allowed only once you say so for a given change.

No pair is written into a prompt or a skill, so changing which model a role uses is one edit in one
place.

## The pair judge

When an agent is about to start, a small headless call, the **pair judge**, reads the task's prompt,
and the file parts the prompt names, and picks one of the role's allowed pairs: a mechanical task gets a cheaper pair, a hard one a
stronger pair. A judge that cannot answer never blocks the run; the role's first pair is used. Every
role with base pairs is judged except the pair judge itself.

Telling the session in plain words to use a different model for a role overrides the pick for that
run only. Every dispatch records the pair it actually ran on, which is what the dashboard's
**Role costs** view adds up.

## The roles

Roles are grouped by the stage they run in: planning, implementation, review, verification,
self-review, pipeline fix, plus a `general` role for anything else and the pair judge itself. The
full list, with one line on what each does, is the **Allowed pairs** section of the canonical file.

Canonical: [`skills/flow-contracts/model-policy.md`](../skills/flow-contracts/model-policy.md)
(roles, tiers, overrides, and **Model and effort**, the judge call).
