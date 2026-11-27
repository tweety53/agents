-- 0110_drop_runner_role.sql: flow-guard run-verify-list re-runs a failing
-- test command once itself and prints the list's verdict, so no agent runs
-- the test list and the runner role loses its allowed pairs; a failure the
-- re-run cannot classify is triaged by a `general` agent. Historical dispatch
-- rows keep their role, and dispatch_class still folds `runner` to verifier;
-- only the role rows go.
DELETE FROM role_pairs WHERE role = 'runner';
