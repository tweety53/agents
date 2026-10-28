# Live verification — kan-1046-review-dynamics-dashboard

Run 2026-10-28 against `flow_kan1046_live`, a pg_dump copy of the dev `flow` database, served by a scratch `flowd` built from the change at HEAD 57d13adc on 127.0.0.1:5299. The scratch database was dropped afterwards. The dev flowd and its database were not touched.

Merge base: c4202f609b117c26c8f5f0b69c002cb958069733

| Check | Before (merge base, the copy as the base build left it) | After (change build) | Failure look |
|-------|------|-------|------|
| `select count(*) from prompt_changes` | `ERROR: relation "prompt_changes" does not exist` | 0 after migration | table missing after start |
| `flow prompt-tuning import -since 2026-08-15` | n/a (no command on the base build) | `imported: 148` | 0 rows |
| `flowd backfill-models` | n/a | `stamped: 2208` | 0 stamped |
| review-dynamics, 56 days: `bucket` | n/a (view absent) | `week` | not week |
| 56 days: markers | n/a | 138, `lastPromptChangeAt` 2026-10-27T20:43:36Z | empty |
| 56 days: series | n/a | every metric present. Slot series include `primary` and `principles` | every series empty |
| review-dynamics, 14 days: `bucket` | n/a | `day` | not day |
| Hand check, week 2026-10-19, minor findings per change | n/a | API 5.1351 over 37 changes. psql 5.1351 over 37 (cohort is each change's first reviewer dispatch with a panel slot) | the two disagree |

Result: matched.
