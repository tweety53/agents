# Design — kan-852-verbatim-move-guard-and-load-reports

## Decisions

### D1 — the verbatim guard compares sentences, not files

`check-verbatim-moves` splits the run-loaded corpus (`skills/`, `commands-claude/`, `rules/`) into
sentences at the merge base and in the working tree. A sentence that left the run-loaded files
passes only if it is still stated in one or landed verbatim in a `-rationale.md`; a new
run-loaded sentence passes only if it is a load directive or a citation. Anything else a change
means to say is listed in its `verbatim-moves.txt`, so a rewording is always a visible,
reviewed act rather than a silent one.

A bold numbered heading (`**4. Execute (SDD + TDD)**`) does not end a sentence: the splitter's
first run cut every citation of one in two, so each such rewording read as two unrelated
fragments.

### D2 — a split citation is judged the way a one-line citation is

`check-references` joins a line with the next when a bold span or its path straddles the
break. A path such a pair names passes when any bold token the joined text associates with it
resolves — the rule the one-line check already applies. Without that, a line citing a live
heading failed because a wrapped bold prompt string beside it was associated with the same path.

### D3 — the 38 stale citations are fixed, not baselined

Reading split citations found 38 that named no heading of their target. Each now names a
heading that exists, keeping the old bold phrase as plain words so no rule text is lost; five
whose bold text is a paragraph lead rather than a citation carry `refs-guard:allow`, the
guard's documented remedy for emphasis. Every reworded sentence is listed in
`verbatim-moves.txt`. Five expected-zero coverage declarations stopped being true (each file now
holds a checked split citation) and were removed.

### D4 — the load reports never fail

`scripts/load-sets.sh` (static bytes per session) and `scripts/loaded-files.py` (what a real
transcript read) are reports: both exit 0. A missing file in a load set is printed and counted
as zero, so a stale array shows in the output instead of aborting it.

## First measurement

Static load sets at `4bd7c4ec`, always-on block rendered sandboxed
(`SB="$(mktemp -d)"; HOME="$SB" ./setup.sh global`), tokens ≈ bytes/4:

```text verified:the output of scripts/load-sets.sh run on this tree with the sandboxed always-on block above
always-on (global block + project CLAUDE.md)         19932 bytes  ~  4983 tok
router (command + SKILL.md + pipeline.md)            46907 bytes  ~ 11726 tok
--- planning session (creating run) ---
  phase files + load directives                      88223 bytes  ~ 22055 tok
  + cited at point of use                            35302 bytes  ~  8825 tok
  TOTAL definite                                    155062 bytes  ~ 38765 tok
  TOTAL incl. cited                                 190364 bytes  ~ 47591 tok
--- implementation session (enters via Resuming at STARTED) ---
  phase files + load directives                     247484 bytes  ~ 61871 tok
  + cited / reviewer templates                       39847 bytes  ~  9961 tok
  + conditional (optional slots, isolation, UI)      93847 bytes  ~ 23461 tok
  TOTAL definite                                    314323 bytes  ~ 78580 tok
  TOTAL incl. cited                                 354170 bytes  ~ 88542 tok
  TOTAL worst case                                  448017 bytes  ~112004 tok
--- finish session (merge-and-push: run 1 chained into run 2) ---
  phase files + load directives                     149381 bytes  ~ 37345 tok
  + cited                                            13196 bytes  ~  3299 tok
  + conditional (follow-ups, state file, config)    110633 bytes  ~ 27658 tok
  TOTAL definite                                    216220 bytes  ~ 54055 tok
  TOTAL incl. cited                                 229416 bytes  ~ 57354 tok
--- subagents (each also inherits whatever always-on context the harness gives it) ---
  reviewer: baseline + primary template              11609 bytes  ~  2902 tok
  reviewer: baseline + principles + EP               23767 bytes  ~  5941 tok
  reviewer: baseline + failure modes                 12766 bytes  ~  3191 tok
  project-configuration.md (cited everywhere)        53834 bytes  ~ 13458 tok
  state-file.md (load before any state r/w)          21258 bytes  ~  5314 tok
```

The transcript measurement (`scripts/loaded-files.py`) is still to be taken from real `/flow`
sessions on the operator's machine. The transcripts in the container this change was written in
are development sessions that edit the flow files, and what they read is not what a run loads.

## Open questions

### Q1 — load-set arrays are maintained by hand

`load-sets.sh` states each session's load set as arrays that a trim must update. Deriving them
from the load directives themselves would remove that step; it is left for a trim that changes a
load set.
