-- KAN-1130 request 200: an implement estimate is one score, taken once its tests are green. A row that carries
-- the after estimate takes it as its score and gaps (a 100% keeps no tag, as a filed 100% does); every
-- other row keeps what it has. Then the after columns go.
UPDATE confidence_estimates
SET score = score_after, gaps = gaps_after, tags = CASE WHEN score_after = 100 THEN '{}' ELSE tags END
WHERE score_after IS NOT NULL;
ALTER TABLE confidence_estimates DROP COLUMN score_after, DROP COLUMN gaps_after;
