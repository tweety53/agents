-- 0066_dispatch_closes_loop.sql — whether a scoped visual-verify verifier row closed the fix loop (KAN-1122).
--
-- dispatches.closes_loop is set by `flow record dispatch end -closes-loop`,
-- which flow-guard visual-verify-pass close passes when every frame on the
-- change's list was re-captured during the row and every frame whose capture
-- changed since its round base was judged. `flow record dispatch begin` and
-- flow-guard check-visual-verify-dispatched accept such a row as the loop's
-- close, as they accept a completed -fix-<k>-full row. Every existing row
-- reads false.
ALTER TABLE dispatches ADD COLUMN closes_loop boolean NOT NULL DEFAULT false;
