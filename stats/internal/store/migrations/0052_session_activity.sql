-- 0052_session_activity.sql: what each Claude Code session is doing now.
--
-- hooks/flow-session-activity.py reports working, question (an
-- AskUserQuestion awaits an answer), permission (a permission prompt awaits
-- one) or idle (the turn ended), keyed by the harness session id. The live
-- state board joins it on the change's open stage run's session_id, so a
-- Running now card can say the session is waiting on the operator.
-- updated_at moves only when state changes (store.SetSessionActivity), so
-- it measures how long the session has been in that state.
CREATE TABLE IF NOT EXISTS session_activity (
  session_id TEXT PRIMARY KEY,
  state      TEXT NOT NULL CHECK (state IN ('working','question','permission','idle')),
  detail     TEXT NOT NULL DEFAULT '',
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
