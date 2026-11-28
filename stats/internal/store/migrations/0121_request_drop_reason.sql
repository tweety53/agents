-- 0121_request_drop_reason.sql: a dropped /flow-interactive request's reason,
-- the operator-readable why `flow record request end -outcome dropped -reason`
-- requires. Empty on every other outcome and on rows written before it.
ALTER TABLE requests ADD COLUMN IF NOT EXISTS reason TEXT NOT NULL DEFAULT '';
