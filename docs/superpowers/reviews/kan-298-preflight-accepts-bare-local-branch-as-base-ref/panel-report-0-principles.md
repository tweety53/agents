Both test suites reproduced and pass. Reviewing against principles: this is a purely documentation/usage-message change plus a source-grep regression test in each of two guard scripts — no behavior, control flow, or state changes.

### Summary
Diff adds a usage-message clause (base-ref substitution rule) to two sibling guard scripts and a `grep -qF` regression test asserting that clause's presence in each script's source. No project-specific standards resolved (`CLAUDE.md`'s standards section is the unfilled template; `AGENTS.md` present but not cited as authoritative beyond the template). Complies with engineering principles and with all four verbatim design constraints.

### Issues
None.

#### Critical (Must Fix)
none

#### Important (Should Fix)
none

#### Minor (Nice to Have)
none

### Assessment
**Principles-compliant?** Yes
**Reasoning:** Change is a same-shape, twice-applied fix (DRY satisfied by applying identically to both siblings, not by extracting a shared string — reasonable given `usage-message-is-the-only-site` forbids a second copy site anyway); tests assert the guard's own source per `source-grep-over-behaviour-test`; both guard header comment blocks and `finish-contract.md` are untouched, verified by diff and `git diff --stat`; both test suites run green.
