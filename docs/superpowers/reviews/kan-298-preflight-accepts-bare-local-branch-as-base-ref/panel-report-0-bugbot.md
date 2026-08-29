## Mutation table

| # | File | Mutation | Caught? |
|---|------|----------|---------|
| M1 | check-finish-preflight.sh | Delete entire heredoc usage block | Caught — `usage message no longer states the base-ref rule` |
| M2 | check-finish-preflight.sh | Drop `>&2` (heredoc → stdout instead of stderr) | **Survived** |
| M3 | check-finish-preflight.sh | Remove `exit 2` after the block | Survived (coincidental — next check `[ ! -d "$WORKTREE" ]` also exits 2 for empty arg), pre-existing redundancy, not diff-introduced |
| M4 | check-finish-preflight.sh | Move `EOF` off column 0 (indented) | Caught — heredoc unterminated, whole script breaks, 28 cases fail |
| M5 | check-finish-preflight.sh | Unquote heredoc delimiter (`<<EOF`) | Survived, but no observable diff (no `$`/backticks in body) — not a real defect |
| M6 | check-finish-preflight.sh | Reword grepped phrase (`prefers`→`favors`) | Caught — exact-string `grep -qF` assertion fails |
| CBM-M1 | check-base-moved.sh | Delete entire heredoc | Caught |
| CBM-M2 | check-base-moved.sh | Drop `>&2` | **Survived** |
| CBM-M3 | check-base-moved.sh | Reword grepped phrase | Caught |
| CBM-M4 | check-base-moved.sh | Move `EOF` off column 0 | Caught (41 cases fail) |

All files verified byte-identical to originals after every mutation; `git status --porcelain` shows only the pre-existing untracked `spectre/changes/...` dir. No commits, no index/HEAD changes.

## Findings

**1. Important — usage message routed to stdout goes undetected (surviving mutant M2/CBM-M2)**
`scripts/check-finish-preflight.sh:54`, `scripts/check-base-moved.sh:50`
The heredoc is piped through `cat >&2 <<'EOF' ... EOF`. If the `>&2` redirect is dropped (or ever regresses), the usage text lands on **stdout** instead of stderr — violating the script's own stated contract ("Prints ONE verdict line to stdout"; a caller parsing stdout for the verdict would now see the usage banner mixed in). No test in either harness catches this: `run_guard` captures stdout+stderr merged (`2>&1`) into `$OUT`, and the missing-argument test case (`test-check-finish-preflight.sh:317-321` / equivalent in `test-check-base-moved.sh`) checks only the exit code, never which stream the text arrived on. The new KAN-298 assertion (`grep -qF ... "$SCRIPT_DIR/check-finish-preflight.sh"`) greps the **source file**, not runtime output, so it can't catch a stream-routing regression either.
- Concrete failure scenario: a future edit accidentally changes `cat >&2 <<'EOF'` to `cat <<'EOF'` (e.g. during a refactor); nothing in either test suite fails, and any caller that does `OUT=$(guard ... 2>/dev/null)` to read only the verdict on success would now leak the usage text into what it treats as stdout on the failure path too.
- Fix: add a runtime assertion in both `test-check-finish-preflight.sh` and `test-check-base-moved.sh` for the missing-argument case that captures stdout and stderr separately (e.g. `OUT=$("$GUARD" 2>/tmp/err); [ -z "$OUT" ] && [ -s /tmp/err ]`) and asserts the usage text is on stderr, not stdout.
- Reproducer:
```
cd /Users/tweety53/Projects/agents-worktrees/spectre-kan-298-preflight-accepts-bare-local-branch-as-base-ref
sed -i.bak "s/cat >&2 <<'EOF'/cat <<'EOF'/" scripts/check-finish-preflight.sh
./scripts/test-check-finish-preflight.sh; echo "rc=$?"   # rc=0, all cases still pass
mv scripts/check-finish-preflight.sh.bak scripts/check-finish-preflight.sh
```

**2. Minor — `exit 2` after the usage block is redundant, not defect-introducing**
`scripts/check-finish-preflight.sh:60-61`, `scripts/check-base-moved.sh:56-57`
Removing the explicit `exit 2` still yields exit 2 because empty `$WORKTREE` fails the very next `[ ! -d "$WORKTREE" ]` check. Pre-existing behavior, not introduced by this diff — no fix required, noted only because it's a surviving mutant against the "missing arguments → exit 2" test.
- Reproducer: `none — pre-existing redundancy independently confirmed via manual mutation of check-finish-preflight.sh:60-61, restored, not a regression from this diff`

VERDICT: FINDINGS
