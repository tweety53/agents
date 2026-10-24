-- 0046_dispatch_round_base.sql — the round base a scoped verifier re-run was begun with.
--
-- A visual-verify re-run scoped to one fix round's delta records that round's
-- base sha (`flow record dispatch begin -round-base <sha>`), so
-- check-visual-verify-dispatched.sh can refuse a worktree whose newest
-- completed verifier ran scoped: the loop's closing unscoped run was never
-- recorded (skills/flow/verify-fix-loop.md). Nullable: every existing row and
-- every unscoped dispatch carries none.

ALTER TABLE dispatches ADD COLUMN round_base text;
