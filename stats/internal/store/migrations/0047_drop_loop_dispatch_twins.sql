-- 0047_drop_loop_dispatch_twins.sql — drop the hand-recorded twins of loop dispatches.
--
-- kan-1032's self-review: a parent told by hooks/enforce-agent-baseline.py to
-- `flow record dispatch begin` a pipeline-fix loop key under the run's change
-- left an agentless row there, while hooks/enforce-loop-dispatch.py recorded
-- the real, stamped row under the loop's own change. The twin is a dispatch
-- with no agent, a pipeline-fix- key, under a change that is not a loop change,
-- whose same key exists under that change's loop (`pipeline-fix-<change>-<n>`)
-- in the same project. A finding pointing at a twin is unlinked first.

WITH twins AS (
  SELECT d.id
  FROM dispatches d
  JOIN changes c ON c.id = d.change_id
  WHERE d.agent_id IS NULL
    AND d.dispatch_key LIKE 'pipeline-fix-%'
    AND c.name NOT LIKE 'pipeline-fix-%'
    AND EXISTS (
      SELECT 1 FROM dispatches t
      JOIN changes lc ON lc.id = t.change_id
      WHERE t.dispatch_key = d.dispatch_key
        AND lc.project_key = c.project_key
        AND lc.name LIKE 'pipeline-fix-' || c.name || '-%'
    )
), unlinked AS (
  UPDATE findings SET dispatch_id = NULL WHERE dispatch_id IN (SELECT id FROM twins)
)
DELETE FROM dispatches WHERE id IN (SELECT id FROM twins);
