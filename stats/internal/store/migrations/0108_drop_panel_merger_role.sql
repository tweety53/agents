-- 0108_drop_panel_merger_role.sql: a panel round's findings are merged by
-- flow-guard merge-panel-findings, never a dispatched agent, so the
-- panel-merger role loses its allowed pairs. Historical dispatch rows keep
-- their role; only the role rows go.
DELETE FROM role_pairs WHERE role = 'panel-merger';
