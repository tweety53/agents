# pipeline-fix-kan-939-step-4-frontend-cardio-plan-view-start-log-16

## Context

Defect: the explain-then-ask turn split (b8e0e063, narrowed by pipeline-fix 15 in 1a709822 and
d1085fb4) lives in the shared contract `skills/flow-contracts/operator-prompts.md` as **Explain
before asking**, though the operator scoped it (2026-10-16) to the flow-eval candidate mining ask
of `skills/flow-self-review/SKILL.md` step 4 alone.

Blast radius — not `big`: three run-loaded files cite or carry the section —
`skills/flow-contracts/operator-prompts.md`, `skills/flow-self-review/SKILL.md`,
`skills/flow/unfinished-work-gate.md` — plus this change's `verbatim-moves.txt`.
<!-- measured: grep -rln "Explain before asking" --exclude-dir=.git --exclude-dir=archive . @ a618d04f — 3 files -->

The pre-bb273948 text of the finding and filing-ask sentences is restored from
`git show bb273948`'s pre-image, never paraphrased. The saved memory
`explain-in-chat-before-ask` already states the narrowed scope; nothing to update there.

## Decisions

### Move the rule into self-review step 4

**ID:** rule-lives-in-step-4
**Status:** active
**Chosen:** the mining-ask split and its reasons move into `skills/flow-self-review/SKILL.md`
step 4, next to 3fd761da's rendering block; operator-prompts.md loses the section — the rule's
only caller is that step, and the shared contract then states nothing a non-mining ask must avoid.
**Considered:** keeping the narrowed section in operator-prompts.md (pipeline-fix 15's shape) — a
one-caller rule in the shared contract, which is what led every ask site to read it as theirs.

### Restore the pre-bb273948 text, not a rewording

**ID:** restore-original-text
**Status:** active
**Chosen:** step 4's opening sentences and the gate's filing-ask paragraph take back their
pre-bb273948 wording verbatim — the operator's "as before" is that text.
**Considered:** keeping pipeline-fix 15's "in chat, ahead of the ask in the same turn" wording with
the citation dropped — a paraphrase of the restored meaning, which the cut-never-paraphrase rule
forbids.

### Drop the sentences step 4 already states

**ID:** dedupe-moved-sentences
**Status:** active
**Chosen:** the section's bold lead, its "That covers … step 2" sentence, "Either way …" and
"Option text records the decision only." are not moved: step 4 already states the angle-7 split
("When the call offers an angle-7 candidate …"), that a candidate's explanation is its template, and
that a prompt's option text cannot carry the explanation and records the decision only. "Every
other ask explains and asks in the same turn" is not moved either — with the section gone, no
contract tells a non-mining ask to split, which is the operator's "as before".
**Considered:** moving the whole section verbatim — restates step 4 twice over in one paragraph.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** low — a prose-only move across three files, guarded by check-verbatim-moves and
check-references, no code or guard seam.
**Considered:** medium — for a change whose seam a cheap review could miss; this one has none.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one fixer, one task — the three files change together or the citation breaks.
**Considered:** one task per file — operator-prompts.md's removal alone leaves two dangling citers.

## Open questions
