-- 0095_self_review_reason.sql: the operator's optional reason for approving or
-- declining a suggested self-review finding, shown on the Self-review
-- dashboard beside the decision. NULL until a decision carries one.
ALTER TABLE self_review_findings ADD COLUMN IF NOT EXISTS review_reason TEXT;
