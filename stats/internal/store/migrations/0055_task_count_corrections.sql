-- 0055_task_count_corrections.sql: each task-count observation also carries
-- the plan's count of `Correction (YYYY-MM-DD):` paragraphs at that moment --
-- the plan-corrections metric allowed-pairs experiments compare (KAN-1067).
-- Rows recorded before this column, and clients that never send it, read 0.
ALTER TABLE change_task_counts ADD COLUMN IF NOT EXISTS corrections INT NOT NULL DEFAULT 0;
