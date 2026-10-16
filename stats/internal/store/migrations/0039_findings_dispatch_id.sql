-- 0039_findings_dispatch_id.sql: findings.dispatch_id is the reviewers view's join key (aggregate.go
-- Reviewers); without an index every period scan seq-scans findings.
CREATE INDEX IF NOT EXISTS findings_dispatch_id ON findings (dispatch_id);
