-- 0058_session_activity_ended.sql: ended -- the harness session closed
-- (SessionEnd), so the live state board can tell a dead interactive session
-- from one waiting for its next request (KAN-1070).
ALTER TABLE session_activity DROP CONSTRAINT IF EXISTS session_activity_state_check;
ALTER TABLE session_activity ADD CONSTRAINT session_activity_state_check
    CHECK (state IN ('working','question','permission','idle','ended'));
