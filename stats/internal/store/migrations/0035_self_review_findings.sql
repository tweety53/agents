-- 0035_self_review_findings.sql: one row per /flow-self-review finding and
-- what became of it -- fixed and landed (ref = the landed sha), filed
-- (ref = the issue key) or declined (no ref) -- KAN-875.
--
-- Not the panel's findings table: that table's status vocabulary, its
-- Minor-only deferral rule, canonical_slot folding and the panel guards
-- that read it are all panel-specific (design.md, decision
-- self-review-findings-table). disposition is closed here as at the store;
-- the ref-shape rules per disposition live in the store's validation.
--
-- change is text, not a changes(id) FK: the change is FINISHED and possibly
-- archived when its self-review runs. blast_radius is NULL for a finding
-- with no file count, a product-code finding among them.

CREATE TABLE self_review_findings (
  id           BIGSERIAL PRIMARY KEY,
  project_key  TEXT NOT NULL REFERENCES projects(project_key),
  change       TEXT NOT NULL,
  angle        TEXT NOT NULL,
  note         TEXT NOT NULL,
  disposition  TEXT NOT NULL CHECK (disposition IN ('fixed','filed','declined')),
  ref          TEXT,
  blast_radius INT,
  recorded_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX self_review_findings_project_change ON self_review_findings (project_key, change);
