-- 0044_flow_settings_drop_extra_dispatch_roles.sql: strip failure-modes and
-- mutation from the stored reviewer list.
--
-- kan-1028 moved both roles onto the always-on mutation+failure-modes extra
-- dispatch, which every pass 1 runs, and ValidReviewers (settings.go) no
-- longer accepts either; a row still listing one would make /flow-settings'
-- "keep current" send back a list PutSettings now refuses. Order of the
-- surviving ids is kept, as 0033 keeps it.
UPDATE flow_settings
SET reviewers = COALESCE((
  SELECT jsonb_agg(r ORDER BY ord)
  FROM jsonb_array_elements_text(reviewers) WITH ORDINALITY AS t(r, ord)
  WHERE r NOT IN ('failure-modes', 'mutation')
), '[]'::jsonb);
