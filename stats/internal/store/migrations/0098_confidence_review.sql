-- 0098_confidence_review.sql: a confidence estimate is reviewed on the Confidence dashboard as a
-- self-review suggestion is on its own (0092, 0095): review_status is to-review on insert, then
-- approved or declined, and review_reason is the operator's optional reason for that decision,
-- NULL until one carries it. Every row has a status -- there is no unreviewable kind of estimate.
ALTER TABLE confidence_estimates ADD COLUMN IF NOT EXISTS review_status TEXT NOT NULL DEFAULT 'to-review'
    CHECK (review_status IN ('to-review', 'approved', 'declined'));
ALTER TABLE confidence_estimates ADD COLUMN IF NOT EXISTS review_reason TEXT;
