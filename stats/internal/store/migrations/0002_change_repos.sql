-- 0002_change_repos.sql: the repository set for a change that spans more
-- than one repository.
--
-- A change is one unit of work regardless of how many repositories it
-- touches -- one `changes` row, never one per repository. This table
-- normalises the affected repositories beside that single row.
--
-- The FK cascade and the composite primary key were verified directly
-- against this migration's target, postgres:18-alpine, before being
-- committed here: a manual INSERT/DELETE against a throwaway table proved
-- ON DELETE CASCADE removes every change_repos row when its owning change
-- row is deleted, and that (change_id, repo_root) enforces one row per
-- repository per change. Both match design.md's DDL as written.

CREATE TABLE change_repos (
  change_id   BIGINT NOT NULL REFERENCES changes(id) ON DELETE CASCADE,
  repo_root   TEXT   NOT NULL,
  merge_base  TEXT,
  PRIMARY KEY (change_id, repo_root)
);

CREATE INDEX change_repos_repo_root ON change_repos (repo_root);
