-- 0107_drop_dispatch_closes_loop.sql — nothing writes or reads dispatches.closes_loop since
-- `flow-guard visual-verify-pass close` stopped passing `-closes-loop` (0066 added it); the flag is gone.
ALTER TABLE dispatches DROP COLUMN IF EXISTS closes_loop;
