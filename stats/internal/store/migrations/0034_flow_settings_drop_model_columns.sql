-- 0034_flow_settings_drop_model_columns.sql: drop default_model and
-- self_review_model from flow_settings.
--
-- The Decide step now picks every dispatch's model per pair
-- (remove-default-model), and self-review runs only as the deferred
-- /flow-self-review pass on the operator's own session model
-- (remove-self-review-model), so neither column has a reader. reviewers
-- is kept untouched.
ALTER TABLE flow_settings DROP COLUMN default_model, DROP COLUMN self_review_model;
