-- 0026_task_counts.sql: one append-only row per observation of a change
-- plan's total column-0 task count -- recorded when the plan is written
-- (flow.writing-plans) and again whenever a fix round re-plans it
-- (flow.document-fix).
--
-- planned and appended are derived at read time, never stored: planned is
-- the first observation's total, appended the latest minus the first. The
-- series itself is the trend KAN-29's self-review asked for -- gate-time
-- re-planning visible as growth over a change's life rather than a single
-- per-change surprise (KAN-415).

CREATE TABLE change_task_counts (
  id          BIGSERIAL PRIMARY KEY,
  change_id   BIGINT NOT NULL REFERENCES changes(id),
  total_tasks INT NOT NULL CHECK (total_tasks > 0),
  observed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX change_task_counts_change ON change_task_counts (change_id, observed_at);
