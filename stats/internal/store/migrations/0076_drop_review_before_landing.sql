-- 0076_drop_review_before_landing.sql — slim-flow-loop removed the
-- review-before-landing hold: a run that writes IN_PROGRESS continues into
-- integrate, so the per-change flag has no reader. 0045 added the column.
ALTER TABLE changes DROP COLUMN IF EXISTS review_before_landing;
