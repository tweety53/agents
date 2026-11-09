-- 0060_dispatch_request.sql: the /flow-interactive request a dispatch ran
-- for (kan-1081 self-review, F5). An operator-asked row carries no task, so
-- without it the ledger read `- Task: no task` for every such row. Named
-- 0060 so it sorts after both 0059_* files.
ALTER TABLE dispatches ADD COLUMN IF NOT EXISTS request integer CHECK (request > 0);
