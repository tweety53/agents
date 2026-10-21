-- 0041_decision_parent_effort_default.sql: a decision's parent effort is the -effort every inline
-- dispatch row copies, so it takes that flag's closed set (low, medium, high, default). Decide
-- wrote `unknown` where the session stated none, which the flag refuses; `default` is the flag's
-- word for that case (kan-920 self-review). `flow decision render` refuses it from now on.
UPDATE decisions
SET decision = jsonb_set(decision, '{parent,effort}', '"default"')
WHERE decision->'parent'->>'effort' = 'unknown';
