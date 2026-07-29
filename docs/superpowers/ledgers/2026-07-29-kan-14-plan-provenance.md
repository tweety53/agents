# SDD ledger — plan: openspec/changes/kan-14-plan-provenance/tasks.md
Task 1: complete (uncommitted, review clean) — contract + discoverability; both guards clean
Task 1: implementer also fixed "Four narrower contracts" -> "Five"; reviewer judged it in scope
Task 2: complete pending review — 14/14 harness cases; guard flags 4 numeric claims in this
  change's own tasks.md (lines 56, 133-135) — the KAN-6 story quoting "194 tests". Task 3 tags them.
Task 2: fence walk implemented as an array state machine, not the brief's sketched regex
  (brief block was tagged unverified: — the rule working as designed)
SCOPE ADDED MID-RUN (operator-directed, x3): task 7 state-integrity; task 8 remove dead
  myflow-state-advance + widen vocab guard to openspec/specs; task 9 remove dead
  myflow-review-panel-economics + tombstone prose + 5 suppression markers.
Task 4: complete pending review — myflow-start section D + one Guardrail; both guards clean
Task 4: flagged that its guard-run instruction is conditional ("if it declares one"). RESOLVED by
  parent: myflow-start runs against any project and most declare no provenance guard, so the
  conditional is correct — an unconditional run would fail every other repo.
Task 5: complete pending review — myflow-do fourth standing clause + one Guardrail; guards clean
Task 5: correctly left the originStage sentence untouched for task 9
Task 5: complete (uncommitted, review clean) — clause in the dispatch set, both halves present
Task 4: complete (uncommitted, review clean) — reviewer found house precedent for the conditional
  ("run its ## stop command if declared", pipeline.md:308), confirming it reads as obligatory-when-present
Task 2: fix round 1/5 dispatched — CRITICAL over-firing in CLAIM_RE (no right-hand word boundary:
  "10 filesystems", "5 minuteswalk", "10 errorsome" all flag); IMPORTANT missing zero-files-scanned
  refusal that check-references.sh:459-462 has. Reviewer verified fence walk + scope + root override
  correct against adversarial cases beyond the harness.
Task 7: complete (uncommitted, review clean) — framed as clarification of state-file.md's duty; no automated test possible, stated honestly
Task 3: complete pending review — all 6 declared commands exit 0; 4 numeric claims tagged in this
  change's own tasks.md; kan-8 needed none; no exemption or scope narrowing added
Task 3: OBSERVED COST OF PARALLEL LANES — saw two transient failures (test-setup.sh checksum, 3
  provenance fixtures) caused by concurrent edits in scripts//skills/ by other agents. Both cleared
  on re-run. Evidence for KAN-15: lanes need isolation even in a repo with no build lock, because a
  guard reading a file another lane is mid-edit reports a failure that is not real. Parent must
  re-run everything at task 6 rather than trusting any mid-run green.
Task 2: fix round 1/5 (3 addressed, 0 open) — CLAIM_RE boundary ([[:space:][:punct:]]|$),
  zero-scan refusal matching check-references.sh:459-462, lookahead fixture pinned
Task 2: complete (uncommitted, review clean) — 21/21 harness cases
Task 2: NOTE FOR THE PANEL — the Critical was found only because the reviewer INVENTED cases the
  harness lacked. Guard author and harness author were the same agent, so the harness could not
  cover its own blind spot. This is the argument for the independent review slot.
Task 3: complete (uncommitted, review clean) — all six declared commands exit 0 on a fresh run
Task 3: the cross-repo citation was traced by the reviewer to a real commit (c515c42 in
  intellij-review-queue), confirmed to touch only a markdown doc so the suite is unchanged, and the
  test report there reports exactly 197. Honest and traceable, not a fabricated local measurement.
Task 8: complete pending review — DEFAULT_TARGETS widened to openspec/specs only; guard now exits 1
Task 8: found FOUR hits, not the two expected:
  1-2. the two dead capabilities (expected; removed by delta, self-clear at archive)
  3. agents-repo-verification — 4 hits on the banned token. SELF-REFUTING: the parent's OWN delta
     for that capability also carried the token at line 41, so after archive it would stay flagged.
     Parent reworded the delta to name no dead command. Fixed.
  4. myflow-contract-distribution:86 — a requirement anchors the contracts-skill listing "alongside"
     the sibling being removed. Would NOT self-clear at archive. Parent added a MODIFIED delta
     dropping the anchor; the obligation (skill is discoverable) is unchanged.
Task 8: no exclusion mechanism exists in check-vocabulary.sh (only the forbidden per-line marker and
  a CLI override that replaces the whole target array). Guard correctly left failing at handoff.
PARENT ERROR: task-8-brief.md and task-9-brief.md were never generated — briefs were regenerated
  BEFORE groups 8 and 9 were appended. Task 8's agent proceeded from embedded dispatch context and
  completed correctly, but that was luck. All nine briefs now regenerated before task 9 dispatched.
Task 8: fix round 1/5 — rationale comment added at DEFAULT_TARGETS naming the widen-to-openspec trap
Task 8: its first draft quoted a retired token literally as the example, which made the guard flag
  its OWN source file. Reworded generically, no marker added. The guard dogfooding itself.
Task 9: complete pending fix — spec PASS; rewrite-not-delete verified; task 5's clause untouched
Task 9: fix round 1/5 — report claimed 4 markers removed, enumerated 5, true count 6
  (check-references.sh had 2 trailers, not 1). Code correct, RECORD wrong. Corroborated twice:
  reviewer counted from the diff; parent measured live-tree markers 51 -> 45.
Task 8: complete (uncommitted, review clean) — rationale comment verified to name the widen-trap
  explicitly, array unchanged, comment itself does not trip the guard, hit set unchanged at 4 files
Task 9: fix round found TWO MORE unbacked claims beyond the one raised — test-check-references.sh
  assertion count 30 -> 31, and the reproduced hit-list block was missing the
  myflow-contract-distribution:86 line. Asking it to recheck was worth more than the correction asked for.
Task 9: fix round 1/5 (1 addressed + 2 self-found, 0 open) — count now 6 (3+1+2), verified by
  independent git-diff count; all remaining report claims audited and backed
Task 9: complete (uncommitted, review clean)
Task 6: complete (uncommitted, verification PASS) — installer placed plan-provenance.md in all three
  harness trees; 5 declared commands exit 0, check-vocabulary.sh exits 1 by design; guard proven
  non-vacuous in BOTH directions (block tag stripped -> failed at that line; measured: comment
  stripped -> failed at that line; both restored byte-identical); zero suppression additions.
Task 6: its first numeric vacuity attempt used "4 scripts", which did not trip the guard — correctly
  diagnosed as the unit list being deliberately narrow, not vacuity, then retried with "194 tests".
PARENT ERROR (found by panel slot 0): task 6 was dispatched, passed, and reported — but its
  checkboxes were never ticked and it had no ledger line until now. The ledger is the record; a
  passing task absent from it is an unchecked claim, which is this change's own subject.

=== MODEL LEDGER (backfilled, pass-14 fix wave) ===
Backfilled because task 10's own rule requires it and this change's ledger held zero model entries
— including for the pass-13 Opus wave the proposal cites as its evidence. An empty ledger under a
requirement to keep one is the exact shape of claim this change exists to refuse.

Every value below is either observed by the dispatcher, read out of an existing written record, or
marked unknowable. Nothing is inferred from plausibility.

IMPLEMENTERS
  Tasks 1-9: unrecorded. These dispatches predate task 10's rule; no model was named at dispatch
    and none was written down, so the values are not recoverable now. Recorded as UNRECORDED, not
    reconstructed — a plausible guess here is exactly the fabrication this backfill removes.
  Fix wave pass 13 (one agent, whole wave): opus (operator-directed)
  Fix wave pass 14 (one agent, whole wave — this one): opus (operator-directed)

REVIEW PANEL, passes 1-3 (the run recorded in final-review-panel.md)
  Slot 0 Primary                    : sonnet  (spawned directly, model named at dispatch)
  Slot 2 Principles (Merged)        : sonnet  (spawned directly, model named at dispatch)
  Slot 4 Adversarial                : sonnet  (spawned directly, model named at dispatch)
  Slot 5B Principles, simplicity    : sonnet  (spawned directly, model named at dispatch)
  Slot 5C Principles, robustness    : sonnet  (spawned directly, model named at dispatch)
  Slot 1 Bug hunter                 : sonnet, BY SUBSTITUTION — `subagent_type: bugbot` was
    unavailable in that session, so the slot ran as `general-purpose` with the model named
    explicitly. That is why a value is knowable here at all; it is read from the substitution note
    in final-review-panel.md, not assumed.
  Slot 3 Security                   : sonnet, BY SUBSTITUTION — same reason, same source.

  As DESIGNED, slots 1 and 3 dispatch by `subagent_type` and resolve their model from their own
  agent definitions, which the dispatcher never reads. In that configuration the correct value is
  `unknown (agent-defined)`, and the panel record's roster table asserting `sonnet` for them was a
  value the dispatcher had no way to know. The table has been corrected; the substitution note is
  what makes the pass-1..3 values legitimate.

REVIEW PANEL, pass 14 (the run that produced this fix wave)
  Seven slots: opus (operator-directed override of the standing Sonnet default). Recorded here
  because the model policy requires an override to be written down with the dispatch it governs,
  and this one was not.
