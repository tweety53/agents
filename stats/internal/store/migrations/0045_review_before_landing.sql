-- 0045_review_before_landing.sql — whether a change keeps the review handoff.
--
-- TRUE keeps /flow at the IN_PROGRESS review handoff for this change, waiting
-- for a fresh bare /flow before integrate. FALSE — the default, and every
-- existing row — lets the run that writes IN_PROGRESS continue straight into
-- integrate in the same invocation. NOT NULL DEFAULT FALSE so existing rows
-- read as auto-landing without a backfill.

ALTER TABLE changes ADD COLUMN review_before_landing BOOLEAN NOT NULL DEFAULT FALSE;
