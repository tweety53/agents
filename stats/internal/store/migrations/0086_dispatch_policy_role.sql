-- 0086_dispatch_policy_role.sql: the allowed-pairs role a dispatch ran as,
-- read off its prompt's `Role:` line by the hook that begins its row. Role
-- costs counts a dispatch under it when set -- a no-progress escalation
-- under no-progress-fixer, never under the loop's own fixer role the key
-- names. NULL on rows begun without one: those keep the key's derivation.
ALTER TABLE dispatches ADD COLUMN policy_role text;
