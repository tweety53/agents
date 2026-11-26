-- The dispatch hook begins a panel row with no -slot, so every panel row since it took over
-- recorded none and dropped out of Reviewer findings. ApplyDispatchRecord now reads the slot off
-- the key (panelSlotFromKey); this applies the same reading to the rows begun before it: each
-- +-joined part of panel-<round>-<parts> is a panel slot, a trailing -<suffix> dropped, or the
-- row is left alone.
WITH parsed AS (
  SELECT d.id, p.ord,
    substring(p.part FROM '^(primary|principles|test-sanity|call-efficiency|failure-modes|mutation|security)(?:-.*)?$') AS slot
  FROM dispatches d
  CROSS JOIN LATERAL unnest(string_to_array(substring(d.dispatch_key FROM '^panel-[0-9]+-(.+)$'), '+'))
    WITH ORDINALITY AS p(part, ord)
  WHERE d.slot IS NULL AND dispatch_class(d.role) = 'reviewer'
), derived AS (
  SELECT id, string_agg(slot, '+' ORDER BY ord) AS slot
  FROM parsed GROUP BY id HAVING bool_and(slot IS NOT NULL)
)
UPDATE dispatches d SET slot = derived.slot FROM derived WHERE d.id = derived.id;
