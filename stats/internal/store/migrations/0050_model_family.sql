-- 0050_model_family.sql: one name per model family.
--
-- dispatches.model holds an alias (opus), a full ID (claude-opus-5-5) or a
-- harness spelling (GLM-5.3-Flash (inline session)). The Review dynamics
-- view keys every per-model series twice -- by family, versions merged, and
-- by version (model-version-split) -- and model_family folds every spelling
-- onto the family: opus, sonnet, haiku, fable, glm, else the lower-cased
-- input.
CREATE FUNCTION model_family(raw text) RETURNS text
LANGUAGE sql IMMUTABLE STRICT AS $$
  SELECT CASE
    WHEN s LIKE '%opus%' THEN 'opus'
    WHEN s LIKE '%sonnet%' THEN 'sonnet'
    WHEN s LIKE '%haiku%' THEN 'haiku'
    WHEN s LIKE '%fable%' THEN 'fable'
    WHEN s LIKE '%glm%' THEN 'glm'
    ELSE s
  END
  FROM (SELECT lower(btrim(raw)) AS s) t
$$;
