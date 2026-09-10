-- 0021_flow_settings_drop_planning_model.sql: drops flow_settings.planning_model
-- (0017_flow_settings_planning_model.sql).
--
-- Nothing reads PLANNING_MODEL any more: /flow's brainstorming and /flow-research both run
-- entirely in the current session, on the session's own model, with no dispatched planner or
-- researcher subagent to pick a model for (kan-488). A setting no run reads is dead
-- configurability, not a field worth keeping for a future caller that does not exist.
ALTER TABLE flow_settings DROP COLUMN planning_model;
