-- KAN-1130: an estimate of 100% is kept too, with no gap and no tag, so a group that was sure shows
-- as sure and one that never estimated shows as nothing. The review list and its counts read only
-- the rows below 100; the run page reads them all.
ALTER TABLE confidence_estimates DROP CONSTRAINT confidence_estimates_score_check;
ALTER TABLE confidence_estimates ADD CONSTRAINT confidence_estimates_score_check CHECK (score BETWEEN 1 AND 100);
