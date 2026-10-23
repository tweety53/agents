# kan-1027-flow-learning-setting — design

## Context

`flow_settings` (`stats/internal/store/settings.go`, migration `0015_flow_settings.sql`) holds one
harness-wide row with the reviewer slots; `flow settings get|set` (`stats/cmd/flow/settings.go`)
reads and writes it whole through `GET|PUT /api/v1/settings` (`stats/internal/api/settings.go`,
`stats/internal/client/client.go`). No SPA page reads settings. `/flow`'s implementation run ends
on the `IN_PROGRESS` handoff (`skills/flow-contracts/handoff-blocks.md`, printed per **Write
`IN_PROGRESS`**, `skills/flow/verify-and-handoff.md`); `/flow-fast` ends on its own final report
(`skills/flow-fast/SKILL.md`). The operator runs in focus mode, where only a turn's final message is
seen.

## Decisions

### The setting lives in flow_settings

**ID:** learning-in-settings
**Status:** active
**Chosen:** a `learning BOOLEAN NOT NULL DEFAULT TRUE` column on `flow_settings`, `Settings.Learning` in the store, `learning` in the settings DTO — the harness-wide record `/flow-settings` already manages.
**Considered:** a `.flow/project.md` key — per project, where the operator wants one switch for every project; a per-change field — the operator toggles a habit, not a change.

### Default on

**ID:** learning-default-on
**Status:** active
**Chosen:** on when never set — the column default, and `GetSettings`' no-row value — the operator's answer.
**Considered:** off — the recommended option, declined by the operator.

### Settings set merges instead of replacing

**ID:** settings-set-merges
**Status:** active
**Chosen:** `flow settings set` takes `-reviewers` and `-learning on|off`, at least one; it reads the current record, overrides the given fields and writes the whole record back, so setting one never resets the other. The PUT body's `learning` is optional on the wire and reads as on when absent, matching the default.
**Considered:** server-side partial update (PATCH) — a second write path for one row with one client; requiring both flags — every toggle would retype the roster.

### Topic history lives in flowd

**ID:** topics-in-flowd
**Status:** active
**Chosen:** a `flow_learning_topics` table (`id`, `topic`, `note`, `change`, `project_key`, `created_at`), store methods `RecordLearningTopic` and `ListLearningTopics`, `POST|GET /api/v1/learning`, CLI `flow learning add <topic> -note '<one line>' -change <name>` and `flow learning list`. The operator's answer.
**Considered:** a markdown file appended by the agent — counting by eye, per machine, no enforcement of the cap.

### The store enforces the three-times cap

**ID:** cap-in-store
**Status:** active
**Chosen:** `RecordLearningTopic` refuses (`ErrLearningTopicExhausted`, HTTP 409, CLI exit 1) a topic already recorded three times; `flow learning list` prints one JSON array of `{topic, count, lastNote}` so the composing session skips exhausted topics before choosing.
**Considered:** the agent counting from the list alone — the cap would hold only as long as every session obeys prose.

### Topic slugs are normalised

**ID:** topic-slug
**Status:** active
**Chosen:** a topic is a lowercase kebab slug (`[a-z0-9-]+`, at most 64 characters), refused otherwise, so "Go generics" and "go-generics" never count separately.
**Considered:** free text — duplicates would slip under the cap.

### The note is printed in the final message

**ID:** note-in-handoff
**Status:** active
**Chosen:** composed at implementation start, printed as a `## Learning` section of the run's final message — `/flow`'s `IN_PROGRESS` handoff and `/flow-fast`'s final report — the operator's answer, since focus mode hides mid-run text.
**Considered:** printed only at implementation start — invisible in focus mode; both — repetition.

### Which runs show a note

**ID:** learning-scope
**Status:** active
**Chosen:** `/flow`'s implementation run (the first run past the plan gate) and `/flow-fast`; never a fix run, an integrate run or a creating run that stops at the plan gate. The operator's answer.
**Considered:** fix runs too — declined.

### The parent composes the note inline

**ID:** compose-inline
**Status:** active
**Chosen:** the parent session writes the note itself from the plan it is about to implement, following one canonical contract file `skills/flow-contracts/learning.md`: run `flow learning list`, pick a topic related to the plan (its language, a design approach it uses, or agentic coding / LLM / AI-infrastructure practice the run itself exercises) not yet shown three times, else a random programming-language or system-design pattern; write a short theory note, a coding puzzle where the topic admits one, and two or three study links each checked with a credential-free `curl -sfI` (a dead link dropped); record it with `flow learning add`.
**Considered:** a dispatched sonnet writer — a dispatch, a ledger row and a prompt for a few paragraphs the parent already has context for.

### No SPA view

**ID:** no-spa-view
**Status:** active
**Chosen:** the stats web app shows nothing of learning.
**Considered:** a topics page — nobody asked for it.

## Open questions

## Live check

Against the running dev flowd after the change's migration applies: `flow settings get` prints `"learning":true` on a store whose row predates the column; `flow settings set -learning off` then `get` prints `false` with the reviewer list unchanged; `flow learning add live-check-topic -note x -change kan-1027-flow-learning-setting` three times succeeds and the fourth exits 1 naming the cap; `flow learning list` shows `live-check-topic` with count 3. Not working looks like: `learning` missing from `get`, the reviewers reset by `set -learning`, or a fourth add accepted. The check runs against the change's own workspace store (`scripts/workspace.sh`), whose database cleanup drops, never against or by restarting the dev flowd.
