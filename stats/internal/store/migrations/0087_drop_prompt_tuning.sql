-- 0087_drop_prompt_tuning.sql: /flow-prompt-tuning and the prompt-change
-- markers are gone. Its two replay roles' pairs and the prompt_changes table
-- (0049, 0073) go with them.
DELETE FROM role_pairs WHERE role IN ('prompt-tuning-replay', 'prompt-tuning-grader');
DROP TABLE prompt_changes;

-- 0067's theme column and index: the cost angle and its cross-change counting are gone.
DROP INDEX IF EXISTS self_review_findings_theme;
ALTER TABLE self_review_findings DROP COLUMN IF EXISTS theme;
