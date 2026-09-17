-- 0028_canonical_slots.sql: one spelling per review-panel slot.
--
-- The reviewers view groups dispatches and findings by slot, and the
-- roster the review panel is dispatched from names its slots in
-- lower-kebab form (primary, code-review-low, exp-failure-modes). Before
-- 2026-09-09 the skill prompts wrote whatever the dispatcher typed --
-- "Primary", "Code review (low)", "agents-Primary", "primary a580c6..." --
-- so the same role surfaced as up to eight rows. canonical_slot folds every
-- spelling onto the roster's, and RecordDispatch/UpsertFinding apply it on
-- the way in so the rule lives once, in SQL, for history and future alike.
--
-- Rules, in order: strip a "<project>-" prefix and a trailing 17-hex
-- dispatch hash; lower-case; collapse every run of non [a-z0-9+] to one
-- dash and trim dashes; then map the few spellings the generic rule cannot
-- reach. "+" survives so a bundle stays "primary+principles".
CREATE FUNCTION canonical_slot(raw text) RETURNS text
LANGUAGE sql IMMUTABLE STRICT AS $$
  SELECT CASE s
    WHEN 'codereviewlow' THEN 'code-review-low'
    WHEN 'codereview' THEN 'code-review-low'
    WHEN 'primary-primary' THEN 'primary'
    WHEN 'visual-verification' THEN 'visual-verify'
    WHEN 'visualverify' THEN 'visual-verify'
    WHEN 'panelfix' THEN 'panel-fix'
    ELSE s
  END
  FROM (
    SELECT btrim(
      regexp_replace(
        lower(regexp_replace(regexp_replace(raw, '^(agents|gymie)-', ''), ' [0-9a-f]{17}$', '')),
        '[^a-z0-9+]+', '-', 'g'),
      '-') AS s
  ) t
$$;

UPDATE dispatches SET slot = canonical_slot(slot) WHERE slot IS NOT NULL AND slot <> canonical_slot(slot);
UPDATE findings SET slot = canonical_slot(slot) WHERE slot <> canonical_slot(slot);
