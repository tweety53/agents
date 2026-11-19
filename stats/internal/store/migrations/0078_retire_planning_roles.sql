-- slim-flow-loop folded plan review and the groups/panel/visual decision into flow-guard plan-check
-- on the decider role, so no dispatch runs on plan-triage or design-reviewer and their allowed pairs go.
DELETE FROM role_pairs WHERE role IN ('plan-triage', 'design-reviewer');
