---
name: protect-main-checkout-hook-quirks
description: "protect-main-checkout hook: a cat/tee heredoc body to a named file is data unless the rest of the command may run it — a runner word, a substitution, or a command word naming the target (kan-1001 self-review, 2026-10-21)."
metadata:
  node_type: memory
  type: project
  originSessionId: 0ee7404c-f6ea-4286-9207-021bc71e8d5d
  modified: 2026-10-21T00:00:00.000Z
---

`~/.claude/hooks/protect-main-checkout.py` token-scans Bash commands while cwd is a main checkout.

- A `cat > T <<'EOF'` / `tee [-a] T <<'EOF'` body is data — even inside a longer command
  (`D=…; cat > $D/r.md <<'EOF' … EOF; check-plan-shape.sh $D/r.md`) — so prose such as `a > b; rm x`
  in it is never read as a redirect. Only T itself is judged.
- The body stays scanned when the rest of the command may run it: a runner word anywhere (`sh`,
  `bash`, `source`, `.`, `eval`, `xargs`, `python3`, …), a `$(…)`/backtick/process substitution, or
  a command word that may name T — the same spelling, the same file name under another path, or a
  name hidden in an expansion (`$X`). Naming T as a plain argument, or running another path-named
  command (`~/.claude/skills/flow/scripts/check-plan-shape.sh`, `./gradlew`), keeps it data
  (agents, kan-1001 self-review).
- An unquoted-delimiter body's `$(…)` and backtick spans stay scanned; the opener must be a plain
  `cat`/`tee` with a target (`python3 - <<EOF`, `x=$(cat <<EOF` are scanned whole).
- `$VAR` targets are not expanded and pass (fail-open by design) — except a `git push` whose
  repository, remote or refspec is an unexpanded `$VAR`/`$(…)`/backtick and whose destination is the
  default branch or unknown: that is denied, so write pushes with literal names.

**How to apply:** a heredoc report or plan write may share its Bash call with the guard that reads
the file; run the written file through a runner (`bash x.sh`) only in its own call if its body
mentions the main checkout.
