-- KAN-1130: the gaps still open once an implementer's tests are green, filed by `flow confidence after`
-- below 100%: a residual score traces back to named gaps. Empty until then, and at 100%.
ALTER TABLE confidence_estimates ADD COLUMN gaps_after TEXT[] NOT NULL DEFAULT '{}';
