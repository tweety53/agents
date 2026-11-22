-- KAN-1130: the estimate an implementer holds once its tests are green, set by `flow confidence
-- after` on the group's row; NULL until then. The first estimate stays as it was filed.
ALTER TABLE confidence_estimates ADD COLUMN score_after INT CHECK (score_after BETWEEN 1 AND 100);
