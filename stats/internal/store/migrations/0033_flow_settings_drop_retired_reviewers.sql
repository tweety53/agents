-- 0033_flow_settings_drop_retired_reviewers.sql: strip retired slot ids
-- from the stored reviewer list.
--
-- ValidReviewers (settings.go) validates writes only, so a row written
-- before a slot's retirement kept the id: a micro run would resolve a slot
-- with no prompt file behind it, and /flow-settings' "keep current" would
-- send the list back to a PutSettings that now refuses it. bugbot and
-- security were retired with their prompt files; code-review-low and
-- simple-reviewer earlier. Order of the surviving ids is kept.
UPDATE flow_settings
SET reviewers = COALESCE((
  SELECT jsonb_agg(r ORDER BY ord)
  FROM jsonb_array_elements_text(reviewers) WITH ORDINALITY AS t(r, ord)
  WHERE r NOT IN ('bugbot', 'security', 'code-review-low', 'simple-reviewer')
), '[]'::jsonb);
