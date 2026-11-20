-- 0090_drop_retired_bookkeeping.sql: the /flow guard slimming (2026-11-19)
-- deleted about fifty guards and the settings and learning features.
-- Rows naming a deleted guard move to archive_* tables, then leave the
-- live ones; the settings and learning tables are renamed to archive_*
-- since no code reads them any more. role_pairs already holds only
-- dispatched roles (0089), session_token_giveups is still written by the
-- harvest watcher, and the old planning stage keys fold into flow.plan in
-- stages.Retired, so neither is touched here.
CREATE TABLE dead_guard (name TEXT PRIMARY KEY);
INSERT INTO dead_guard (name) VALUES
	('aside-planning-artifacts'),
	('check-archive-scope'),
	('check-ask-silence'),
	('check-cleanup-complete'),
	('check-dispatch-pair'),
	('check-dispatch-paragraphs'),
	('check-done-when-paths'),
	('check-foreign-staged'),
	('check-guard-symlinks'),
	('check-hand-notes-in-step'),
	('check-installed-citations'),
	('check-installed-rules'),
	('check-late-fix-trigger'),
	('check-live-verification'),
	('check-loop-landing'),
	('check-markdown-integrity'),
	('check-mockup-citations'),
	('check-model-keys'),
	('check-mutation-reproducer-pin'),
	('check-narrative-runs'),
	('check-normative-inventory'),
	('check-panel-citation-trigger'),
	('check-panel-diff-size'),
	('check-panel-docs-only'),
	('check-panel-findings-closed'),
	('check-panel-findings-closed.sh'),
	('check-panel-fix-single-dispatch'),
	('check-panel-reproducer-exit-contract'),
	('check-panel-reproducer-exit-contract.sh'),
	('check-panel-reproducers'),
	('check-panel-reproducers.sh'),
	('check-plan-departures'),
	('check-plan-provenance'),
	('check-plan-shape'),
	('check-plan-shape.sh'),
	('check-plan-unchanged'),
	('check-plan-unchanged.sh'),
	('check-planning-commit-location'),
	('check-prompt-tuning-cases'),
	('check-rerun-citations'),
	('check-review-gate'),
	('check-review-gate.sh'),
	('check-self-review-report'),
	('check-spec-reach.sh'),
	('check-stage-mark-calls'),
	('check-task-build-green'),
	('check-task-commit-fields'),
	('check-task-commit-fields.sh'),
	('check-task-commit-planning-paths'),
	('check-task-records'),
	('check-task-reviewer-single-dispatch'),
	('check-task-reviewer-single-dispatch.sh'),
	('check-tree-markers'),
	('check-untasked-commits'),
	('check-verbatim-moves'),
	('check-visual-preflight.sh'),
	('check-visual-trigger.sh'),
	('check-visual-verify-dispatched'),
	('check-vocabulary.sh'),
	('check-workspace-isolation'),
	('check-worktree-location'),
	('check-worktree-location.sh'),
	('check-worktree-processes'),
	('classify-untracked'),
	('close-task.sh'),
	('commit-archive'),
	('commit-split'),
	('decide-groups-panel-visual'),
	('fix-run-floor'),
	('fix-slot-judge'),
	('flow.decide'),
	('fold-fixup'),
	('full-suite'),
	('gather-dispatch-context'),
	('gather-dispatch-context.sh'),
	('generate-relocation-comparison'),
	('guard-autosquash'),
	('guard-autosquash.sh'),
	('land-self-review-report'),
	('live-check'),
	('migrate-worktrees'),
	('mutate-and-verify'),
	('panel-round-begins'),
	('panel-slot-triggers'),
	('plan-agent-pair'),
	('plan-class'),
	('plan-dispatch-bundles'),
	('plan-dispatch-groups'),
	('plan-files-paths'),
	('plan-group-tasks'),
	('plan-script-paths'),
	('plan-suite-steps'),
	('plan-task-id-rule'),
	('prepare-archive-branch'),
	('prove-reproducer'),
	('ready-groups'),
	('refresh-plan-base'),
	('render-slot-prompt'),
	('reshape-branch'),
	('resolve-visual-screenshots.sh'),
	('resume-or-fresh'),
	('run-capped'),
	('run-guard-tests'),
	('run-reproducer.sh'),
	('run-verify-list.sh'),
	('slot-judge'),
	('spectre-link'),
	('spectre-validate'),
	('stage-task-files'),
	('sync-panel-base'),
	('sync-panel-base.sh'),
	('unlanded-archives'),
	('verify-auto-detected-test');

CREATE TABLE archive_guard_runs AS SELECT * FROM guard_runs WHERE guard IN (SELECT name FROM dead_guard);
DELETE FROM guard_runs WHERE guard IN (SELECT name FROM dead_guard);

CREATE TABLE archive_guard_verdicts AS SELECT * FROM guard_verdicts WHERE guard IN (SELECT name FROM dead_guard);
DELETE FROM guard_verdicts WHERE guard IN (SELECT name FROM dead_guard);

CREATE TABLE archive_guard_substitutions AS SELECT * FROM guard_substitutions WHERE guard IN (SELECT name FROM dead_guard);
DELETE FROM guard_substitutions WHERE guard IN (SELECT name FROM dead_guard);

CREATE TABLE archive_incidents AS SELECT * FROM incidents WHERE guard IN (SELECT name FROM dead_guard);
DELETE FROM incidents WHERE guard IN (SELECT name FROM dead_guard);

DROP TABLE dead_guard;

ALTER TABLE flow_settings RENAME TO archive_flow_settings;
ALTER TABLE flow_learning_topics RENAME TO archive_flow_learning_topics;
