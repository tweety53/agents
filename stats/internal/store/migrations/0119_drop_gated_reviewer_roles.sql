-- 0119_drop_gated_reviewer_roles.sql: the gated per-task review retires — the
-- whole-branch review panel is the only review of an implementation — so the
-- gated-reviewer and gated-re-reviewer roles lose their allowed pairs.
-- Historical dispatch rows keep their role, and dispatch_class still folds
-- both to reviewer; only the role rows go.
DELETE FROM role_pairs WHERE role IN ('gated-reviewer', 'gated-re-reviewer');
