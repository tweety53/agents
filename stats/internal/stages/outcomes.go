package stages

// OutcomeCompleted is the outcome value a stage end mark carries when the
// stage's work finished successfully -- the value `flow stage end`
// writes and every reader of stage_runs.outcome that keys on success
// must match. One constant, because a writer and a reader spelling the
// value independently could each drift and never meet: the self-review
// bundle's records-loss note reads a completed flow.review-panel run,
// and a silent spelling drift there would read as no panel ever having
// run (KAN-621's F5).
const OutcomeCompleted = "completed"
