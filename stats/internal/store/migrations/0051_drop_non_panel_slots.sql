-- 0051_drop_non_panel_slots.sql: a reviewer row's slot names panel slots only.
--
-- Reviewer rows were recorded with a gated per-task reviewer's key
-- (task-27+28-reviewer), a self-review pass (self-review-pass) or a typo
-- (pp, verify) as their slot, and the Review dynamics heatmap split each
-- "+"-joined part into a slot row of its own. internal/api's
-- ApplyDispatchRecord now refuses such a slot; this drops every part
-- outside the panel's vocabulary -- the current slots, the retired
-- simple-reviewer, bugbot, security and code-review-low, and any
-- experimental exp-<name> -- and NULLs a slot left with no part.
UPDATE dispatches d SET slot = NULLIF(k.kept, '')
FROM (
  SELECT id, coalesce(string_agg(part, '+' ORDER BY ord) FILTER (
    WHERE part ~ '^(primary|principles|test-sanity|call-efficiency|failure-modes|mutation|simple-reviewer|bugbot|security|code-review-low|exp-[a-z0-9-]+)$'
  ), '') AS kept
  FROM dispatches, unnest(string_to_array(slot, '+')) WITH ORDINALITY AS u(part, ord)
  WHERE role = 'reviewer' AND slot IS NOT NULL
  GROUP BY id
) k
WHERE d.id = k.id AND d.slot IS DISTINCT FROM NULLIF(k.kept, '');
