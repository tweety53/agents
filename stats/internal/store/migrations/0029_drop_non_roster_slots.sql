-- 0029_drop_non_roster_slots.sql: a slot names a review-panel role, nothing else.
--
-- Some sessions tagged panel-fix, verifier and one-off reviewer dispatches
-- with a -slot of their own invention (fix, fix-round-1, verify-gymie,
-- go-audit-correctness), which no recorded roster names. The reviewers
-- view now joins on roster membership, so they no longer surface, and
-- this clears the column on the rows themselves so the ledger says what
-- was true: those dispatches ran under no panel slot. The rows stay -- they
-- carry the change's cost and timing -- only the slot is dropped. A bundle
-- keeps its slot when any one of its roles is roster-named.
UPDATE dispatches d SET slot = NULL
WHERE d.slot IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM decisions, jsonb_path_query(decision, '$.panel.roster[*].slot') r
    WHERE r #>> '{}' = ANY (string_to_array(d.slot, '+'))
  );
