-- 0122_rename_call_efficiency_slot.sql: the call-efficiency review slot is
-- renamed performance, so recorded history shows under the new id: a
-- dispatch slot naming it alone or as one part of a +-joined bundle, a
-- finding's slot, and every slot list a decision recorded (panel[].slots,
-- extra_dispatch.slots, ...) -- a JSON string exactly "call-efficiency",
-- wherever it sits. A slot only containing the text (exp-call-efficiency)
-- and every dispatch key are kept. Idempotent: each update matches only the
-- old name.
UPDATE dispatches
   SET slot = array_to_string(array_replace(string_to_array(slot, '+'), 'call-efficiency', 'performance'), '+')
 WHERE 'call-efficiency' = ANY (string_to_array(slot, '+'));

UPDATE findings SET slot = 'performance' WHERE slot = 'call-efficiency';

UPDATE decisions
   SET decision = replace(decision::text, '"call-efficiency"', '"performance"')::jsonb
 WHERE decision::text LIKE '%"call-efficiency"%';
