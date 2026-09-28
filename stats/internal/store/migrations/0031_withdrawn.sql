-- 0031_withdrawn.sql — the withdrawal marker on a change record.
--
-- The withdrawal route (a change abandoned before planning) terminates a
-- record FINISHED with this flag set, rather than introducing a fourth
-- pipeline state: every FINISHED exclusion — candidate resolution,
-- /flow-status's open set — already applies, and the record stays as the
-- audit trail. NOT NULL DEFAULT FALSE so existing rows read as not
-- withdrawn without a backfill.

ALTER TABLE changes ADD COLUMN withdrawn BOOLEAN NOT NULL DEFAULT FALSE;
