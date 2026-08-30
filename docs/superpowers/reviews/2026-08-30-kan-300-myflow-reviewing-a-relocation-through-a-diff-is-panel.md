# Review panel — kan-300-myflow-reviewing-a-relocation-through-a-diff-is

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Bugbot | Critical | scripts/generate-relocation-comparison.py:373 | UnicodeDecodeError on a non-UTF-8 scoped file or git blob crashes with a traceback (exit 1), violating the documented 0/2 exit-code contract and the never-blocks constraint |
| F2 | Principles | Important | scripts/generate-relocation-comparison.py:320 | _git_show collapses every non-zero git show exit into silent empty text, not only the documented missing-path case, so a real git failure is misreported as an added passage with no diagnostic |
| F3 | Bugbot | Important | scripts/generate-relocation-comparison.py:210 | repoint_signature's BOLD_SPAN_RE collapses every bold span in the passage, not only the one beside the repointed citation, so an unrelated bold-text edit is misclassified repointed instead of removed+added |
| F4 | Bugbot | Important | scripts/test-generate-relocation-comparison.sh:1 | test harness has zero coverage for the exit-2 paths, a missing (not just no/malformed) Relocation header, and an already-unchanged in-scope passage — five surviving mutants found |
| F5 | Bugbot | Minor | scripts/generate-relocation-comparison.py:307 | passage/heading text is written into a Markdown table cell with no pipe escaping, which can corrupt the rendered table's column count |
| F6 | Primary | Minor | scripts/generate-relocation-comparison.py:335 | two byte-identical passages at the same (file, heading), where only one occurrence is later removed, collapse to one removed row rather than reflecting the duplicate |
| F7 | Principles | Critical | scripts/generate-relocation-comparison.py:226 | CITATION_RE's separator only matches an optional comma+whitespace between the backtick path and the bold span, but the corpus's real citation style uses no separator or a possessive apostrophe-s, so a genuine repoint of both path and bold token is misclassified removed+added instead of repointed |
| F8 | Principles | Minor | scripts/generate-relocation-comparison.py:356 | PATH_ABSENT_AT_REF_RE matches only git's default English fatal message; a localized git under a non-C LANG/LC_ALL would fail to match, misreporting a genuinely absent path as a git failure |
| F9 | Primary | Important | scripts/test-generate-relocation-comparison.sh:1 | F1 (UnicodeDecodeError) and F2 (GitShowError non-missing-path case) have no dedicated automated regression test — verified only by ad hoc manual reproduction during review |
| F10 | Bugbot | Important | scripts/test-generate-relocation-comparison.sh:1 | F3 (bold-span adjacency narrowing) and F5 (pipe escaping) have no dedicated automated regression test; case f's merge-base-resolves mutation is not actually caught by its own assertion — it passes coincidentally via a different exit-2 path |
| F11 | Bugbot | Minor | scripts/generate-relocation-comparison.py:226 | _collapse_citation preserves the captured comma/whitespace separator verbatim in the placeholder rather than normalizing it away, so a repoint that also drops the comma is misclassified removed+added |
| F12 | Primary | Important | skills/flow/review-panel.md:146 | the relocation-comparison.md citation uses <worktree>/... instead of the required <abs-worktree>/... prefix, failing check-installed-citations.sh |
| F13 | Principles | Critical | scripts/generate-relocation-comparison.py:210 | CITATION_RE can splice across two unrelated backtick pairs when a second literal code span like `**Squash-with:**` follows shortly after, silently swallowing a genuine content edit inside the first span into the citation placeholder and misclassifying it repointed |
| F14 | Bugbot | Critical | scripts/generate-relocation-comparison.py:365 | PATH_ABSENT_AT_REF_RE only matches git's 'does not exist in' message; a file that exists on disk but not at the ref (the ordinary shape of a file this plan creates fresh) instead emits 'exists on disk, but not in', which the regex misses, so the script wrongly exits 2 instead of treating it as empty/added text |
| F15 | Bugbot | Important | scripts/test-generate-relocation-comparison.sh:1 | no test exercises a bullet-list or table-row passage, the heading (§) suffix in output locations, scoped_files de-duplication across two tasks sharing a file, or a mixed-separator-style citation repoint (F11's own harder case) — several mutations of already-correct code survive the full suite |
| F16 | Primary | Critical | scripts/generate-relocation-comparison.py:253 | TRAILING_POSITION_WORD_RE strips a trailing above/below together with its trailing punctuation on the before-text, but the after-text (which never had the word) keeps its own punctuation untouched, so a genuine repoint with a dropped stale above/below followed by punctuation is misclassified removed+added |
| F17 | Primary | Minor | docs/superpowers/specs/2026-08-30-kan-300-myflow-reviewing-a-relocation-through-a-diff-is-design.md:82 | design doc states a spec addition lands in myflow-contract-economy for the Relocation header field, but no such addition exists — the header requirement was consolidated entirely into myflow-review-panel-economics during implementation |
| F18 | Bugbot | Important | scripts/generate-relocation-comparison.py:414 | _path_exists_at_ref's git cat-file -e succeeds for a tree object, not only a blob, so a scoped path that was a directory at the merge-base ref has its git show tree-listing output silently read as file text and fed into extract_passages, producing spurious rows instead of exit 2 |
| F19 | Bugbot | Minor | scripts/test-generate-relocation-comparison.sh:1 | no test exercises a ### heading, a * bullet, the CITATION_SEPARATOR_RE 0-3 char bound, the trailing-position-word \b boundary, or a Relocation: yes plan with zero scoped Files: entries — all manually verified correct today but unguarded against regression |
| F20 | Bugbot | Important | scripts/generate-relocation-comparison.py:253 | TRAILING_POSITION_WORD_RE and TRAILING_PUNCT_RE only strip a single optional punctuation mark from the [.,;:] set; a run of punctuation or ! / ? still diverges the two signatures, reopening F16's hole one layer deeper |
| F21 | Bugbot | Important | scripts/generate-relocation-comparison.py:414 | git cat-file -t reports blob for a symlink regardless of what it points to, so a scoped symlink passes the F18 blob-type guard but its merge-base content is the raw target-path string while its worktree content is read through open() which follows the link, producing spurious rows even for an unchanged symlinked file |
| F22 | Primary | Minor | scripts/generate-relocation-comparison.py:261 | TRAILING_POSITION_WORD_RE requires the punctuation run to sit immediately after above/below with no intervening whitespace; a passage with whitespace between the word and its punctuation falls through to removed+added instead of repointed — not silent data loss, just a less precise label |

findings-total: 22
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 withdrawn confirmed as correct multiset-diff behavior on inspection, not a defect — operator agreed to withdraw rather than spend a fix round on it
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 fixed
finding-status: F15 fixed
finding-status: F16 fixed
finding-status: F17 fixed
finding-status: F18 fixed
finding-status: F19 fixed
finding-status: F20 fixed
finding-status: F21 fixed
finding-status: F22 fixed

reproducers-total: 22
finding-reproducer: F1 none — needs a non-UTF-8 worktree file/blob fixture; not scriptable as a single argv
finding-reproducer: F2 none — needs a git show failure other than a missing path (corrupt object, malformed ref)
finding-reproducer: F3 none — constructed fixture case, not a single runnable command
finding-reproducer: F4 none — coverage gap, not a single reproducer
finding-reproducer: F5 none — cosmetic, needs a scoped passage containing a literal pipe
finding-reproducer: F6 none — exercised via python REPL, not a spec scenario; multiset dedup on exact-location matches is a reasonable interpretation and no spec scenario covers duplicate passages
finding-reproducer: F7 none — python REPL repro against repoint_signature, not a single shell command
finding-reproducer: F8 none — requires a git build with non-English locale catalogs, not present in this environment
finding-reproducer: F9 none — coverage gap
finding-reproducer: F10 none — coverage gap
finding-reproducer: F11 none — python REPL repro, not a single shell command
finding-reproducer: F12 bash scripts/check-installed-citations.sh
finding-reproducer: F13 none — python repro against repoint_signature with two real corpus-shaped backtick spans
finding-reproducer: F14 none — requires a new uncommitted worktree file scoped and diffed against a merge-base predating it
finding-reproducer: F15 none — coverage gaps, multiple mutations survive
finding-reproducer: F16 none — python repro against repoint_signature with trailing punctuation on one side only
finding-reproducer: F17 none — documentation-only
finding-reproducer: F18 none — requires a fixture where a scoped path is a directory (tree object) at the merge-base ref
finding-reproducer: F19 none — coverage gaps, code manually verified correct
finding-reproducer: F20 none — full-generator fixture with a fluggering above followed by ellipsis or ?!
finding-reproducer: F21 none — full-generator fixture scoping a symlink whose target changes, or an unchanged symlinked file compared via open() (follows) vs git show (does not resolve)
finding-reproducer: F22 none — cosmetic, both classes still surface to the reviewer as removed+added rather than one repointed row
