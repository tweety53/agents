Working tree clean relative to diff (untracked files are unrelated spectre state, not touched by me). Diff scope confirmed matches the three files reviewed.

No suppression, no lint-config changes, no layering violations — this is test-only shell code. Both hard-invariant sources (CLAUDE.md, AGENTS.md) contain no rule about `scripts/lib/` composition beyond what fix round 3's own comment already verified against (check-guard-symlinks.sh, owned-corpus.sh, check-installed-citations.py, `.flow/project.md`'s test list) — confirmed accurate by direct inspection above.

## Summary

Diff touches `scripts/lib/test-git-shim.sh` (shared shim library) and its two test consumers. F10's resolution (keep in `scripts/lib/`, strengthen the header) is well-justified and verified against all four cited enumeration mechanisms. F14's real-git hoist and exact-match documentation, plus the `.fired` sentinel/`assert_shim_fired`, were reproduced directly — reverting the hoist reproduces the chaining bug and the new test catches it (confirmed then restored). No hard-invariant or principle violations found.

### Issues

None.

### Assessment
**Principles-compliant?** Yes
**Reasoning:** F10's directory-placement tradeoff is deliberate, documented, and verified against every enumeration mechanism named; F14's hardening (real-git hoist, exact-match semantics, fired-sentinel) is proven by direct reproduction, not just asserted by comment.

VERDICT: clean
