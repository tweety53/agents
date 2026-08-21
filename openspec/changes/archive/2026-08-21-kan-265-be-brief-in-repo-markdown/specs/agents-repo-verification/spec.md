# agents-repo-verification delta — kan-265-be-brief-in-repo-markdown

## ADDED Requirements

### Requirement: A guard inventories the corpus's normative sentences

`scripts/check-normative-inventory.sh` SHALL print every normative sentence in this repository's
owned Markdown, so a change that edits prose in bulk can prove it dropped none of them.

It exists because no guard already in `## lint` would notice a deleted requirement.
`check-references.sh` and `check-installed-citations.sh` catch a citation that stops resolving;
`check-markdown-integrity.py` catches structural damage; the harnesses under `## test` exercise
script behaviour. A sentence carrying a SHALL can be deleted with every one of them green — which is
the failure mode a corpus-wide trim risks, and the reason this guard is a precondition of one.

- A **normative sentence** SHALL be one containing `SHALL`, `SHALL NOT`, `MUST` or `MUST NOT` as a
  whole word. Nothing else is a keyword: `SHOULD` and `MAY` state no obligation whose loss changes
  behaviour, and including them would make the inventory churn on ordinary editing.
- It SHALL scan the same corpus `scripts/check-contract-budget.sh` covers, resolved the same way, so
  the two cannot disagree about which files this repository owns.
- It SHALL normalise each sentence's internal whitespace to single spaces and strip surrounding
  whitespace, so a reflowed paragraph does not read as a changed requirement. It SHALL NOT normalise
  anything else — case, punctuation and wording are the content being protected.
- It SHALL print one sentence per line, **sorted**, so its output is comparable with `diff` between
  two runs and is independent of file order.
- It SHALL be argument-free and self-scoped from its own location, with an opt-in environment
  override for its own test harness alone, like the guards beside it.
- Its exit codes SHALL be `0` the inventory printed, and `2` it cannot answer at all — an unreadable
  file or a scope root that does not exist. It SHALL NOT have a violation exit code: it reports a
  set, and comparing two sets is the caller's act.
- It SHALL be named in `.myflow/project.md` under `## lint`, and its harness
  `scripts/test-check-normative-inventory.sh` under `## test`.

A change that edits prose across the corpus SHALL capture the inventory before its first edit and
after its last, and SHALL require the two to be byte-identical. A difference SHALL be resolved by
restoring the sentence, never by accepting the new inventory.

#### Scenario: A deleted requirement is caught

- **WHEN** a trim removes a paragraph containing a SHALL sentence
- **THEN** the after-inventory differs from the before-inventory by that line

#### Scenario: Reflowing a paragraph is not a change

- **WHEN** a normative sentence is rewrapped across different line boundaries with no wording change
- **THEN** the inventory is unchanged, because internal whitespace is normalised

#### Scenario: Rewording a requirement is a change

- **WHEN** a normative sentence is reworded to mean the same thing
- **THEN** the inventory differs, because only whitespace is normalised

#### Scenario: The output is order-independent

- **WHEN** the guard runs twice with files traversed in a different order
- **THEN** it prints the same bytes both times

#### Scenario: The guard is a declared lint step

- **WHEN** `.myflow/project.md`'s `## lint` section is read
- **THEN** `scripts/check-normative-inventory.sh` is listed
- **AND** `scripts/test-check-normative-inventory.sh` is listed under `## test`
