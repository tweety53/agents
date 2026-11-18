-- 0070_dispatch_retimed.sql: the audit of `flow record dispatch retime`
-- (decision retime-jsonb-drop-old, KAN-1132). A retimed row carries
-- {reason, at, startedAt, endedAt}: reason and at are the latest retime's,
-- startedAt and endedAt the instants the first retime replaced, kept across
-- later retimes. NULL on a row never retimed.
ALTER TABLE dispatches ADD COLUMN IF NOT EXISTS retimed JSONB;
