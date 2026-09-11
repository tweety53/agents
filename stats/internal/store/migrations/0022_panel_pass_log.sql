-- 0022_panel_pass_log.sql: the review panel's pass log as store rows --
-- the pass-by-pass metadata lines and the fix round's fix-mutation: proof
-- lines that, before this file, lived only in the hand-written
-- .superpowers/sdd/final-review-panel.md, a worktree-lifetime artifact
-- that run 2's cleanup destroyed (KAN-331). Findings and dispatches were
-- already rows; this carries the last two hand-written contents of that
-- file into the same store, so the rendered panel record is the pass log
-- and the file's destruction with the worktree stops mattering.
--
-- Two tables rather than one polymorphic log: a pass line is one note and
-- a mutation line is exactly three fields, and one table carrying both
-- would make every reader dispatch on a kind column to reach the fields
-- it names. Both are append-only, like guard_verdicts and incidents: a
-- pass-log line is recorded once by the parent as the fact arises, and a
-- replayed write can at worst duplicate a line, the same cosmetic risk
-- those tables already carry -- no dedup key exists for "the same fact
-- phrased identically", and inventing one would refuse a legitimate
-- repeat.
--
-- round mirrors findings.round: 0 for the initial panel's passes, 1..n
-- for a fix round's. The rendered pass log groups by it, and the
-- fix-mutations-total count is per round, the scope the review-panel
-- contract's fenced block always gave it.

CREATE TABLE panel_passes (
  id        BIGSERIAL PRIMARY KEY,
  change_id BIGINT NOT NULL REFERENCES changes(id),
  round     INT NOT NULL DEFAULT 0,
  note      TEXT NOT NULL
);
CREATE INDEX panel_passes_change_id ON panel_passes (change_id);

CREATE TABLE panel_mutations (
  id        BIGSERIAL PRIMARY KEY,
  change_id BIGINT NOT NULL REFERENCES changes(id),
  round     INT NOT NULL DEFAULT 0,
  path      TEXT NOT NULL,
  mutated   TEXT NOT NULL,
  test      TEXT NOT NULL
);
CREATE INDEX panel_mutations_change_id ON panel_mutations (change_id);
