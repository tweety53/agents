# Prototype tools for the prompt audit

Prototypes written for the 2026-09-29 audit (plan: [`README.md`](README.md)). They are reference
material, not repository scripts: the guard version of `verbatim_check.py` and the reports belong
in `flow-guard` / `scripts/` under KAN-852, written to `.flow/project.md`'s rules for new guards.
Copy a block out to run it.

## `loadsets.sh` — static bytes each session loads

Sums `wc -c` over each session's load set as the flow files' load directives define it at
`700e184` (planning, implementation, finish — `/clear` between them), plus the files cited at
point of use and the conditional loads. Tokens are approximated as bytes/4. Edit the arrays when
a trim changes what a session loads.

```bash
SB="$(mktemp -d)"; HOME="$SB" ./setup.sh global        # render the always-on block, sandboxed
bash loadsets.sh "$PWD" "$SB/.claude/CLAUDE.md"
```

```bash
#!/usr/bin/env bash
# loadsets.sh <repo> [rendered-global-CLAUDE.md] — per-session instruction bytes this repo
# contributes (tokens ~ bytes/4). Excludes superpowers skills, MEMORY.md and project code/artifacts.
# The second argument defaults to ~/.claude/CLAUDE.md; render a sandboxed one with
#   SB="$(mktemp -d)"; HOME="$SB" ./setup.sh global   # then pass "$SB/.claude/CLAUDE.md"
set -euo pipefail
R="$1"; GLOBAL="${2:-$HOME/.claude/CLAUDE.md}"
F="$R/skills/flow"; C="$R/skills/flow-contracts"
ALWAYS=("$GLOBAL" "$R/CLAUDE.md")
ROUTER=("$R/commands-claude/flow.md" "$F/SKILL.md" "$C/pipeline.md")
sum() { local t=0 f; for f in "$@"; do t=$((t + $(wc -c < "$f"))); done; echo "$t"; }
row() { local name="$1"; shift; local b; b=$(sum "$@"); printf '%-44s %7d bytes  ~%6d tok\n' "$name" "$b" $((b/4)); }
row "always-on (global block + project CLAUDE.md)" "${ALWAYS[@]}"
row "router (command + SKILL.md + pipeline.md)"    "${ROUTER[@]}"
echo "--- planning session (creating run, all inline) ---"
PD=("$F/brainstorm.md" "$F/brainstorm-planner.md" "$C/jira-integration.md" "$C/plan-provenance.md" "$C/build-green.md")
PC=("$C/worktree-resolution.md" "$C/git-boundaries.md" "$C/handoff-blocks.md" "$C/operator-prompts.md")
row "  phase files + Load directives" "${PD[@]}"
row "  + cited at point of use"        "${PC[@]}"
row "  TOTAL definite (always+router+directives)" "${ALWAYS[@]}" "${ROUTER[@]}" "${PD[@]}"
row "  TOTAL incl. cited"               "${ALWAYS[@]}" "${ROUTER[@]}" "${PD[@]}" "${PC[@]}"
echo "--- implementation session (inline default; enters via Resuming at STARTED in brainstorm.md) ---"
ID=("$F/brainstorm.md" "$F/implement.md" "$C/artifacts-registry.md" "$C/worktree-resolution.md" "$F/review-panel.md" "$F/verify-and-handoff.md" "$C/session-records.md" "$C/git-boundaries.md")
IC=("$C/operator-prompts.md" "$C/model-policy.md" "$C/known-bugs.md" "$F/primary-reviewer-prompt.md" "$F/principles-reviewer-prompt.md")
IX=("$F/review-panel-optional-slots.md" "$C/workspace-isolation.md" "$F/visual-verify.md" "$F/bugbot-reviewer-prompt.md" "$F/security-reviewer-prompt.md")
row "  phase files + Load directives" "${ID[@]}"
row "  + cited / reviewer templates"   "${IC[@]}"
row "  + conditional (big roster, isolation, UI)" "${IX[@]}"
row "  TOTAL definite"                  "${ALWAYS[@]}" "${ROUTER[@]}" "${ID[@]}"
row "  TOTAL incl. cited"               "${ALWAYS[@]}" "${ROUTER[@]}" "${ID[@]}" "${IC[@]}"
row "  TOTAL worst case"                "${ALWAYS[@]}" "${ROUTER[@]}" "${ID[@]}" "${IC[@]}" "${IX[@]}"
echo "--- finish session (merge-and-push: run 1 chained into run 2) ---"
FD=("$F/integrate.md" "$C/worktree-resolution.md" "$C/finish-contract-run1.md" "$C/git-boundaries.md" "$C/session-records.md" "$C/jira-integration.md" "$F/archive.md" "$C/artifacts-registry.md" "$C/finish-contract-run2.md")
FC=("$C/operator-prompts.md" "$C/model-policy.md")
FX=("$C/jira-followups.md" "$C/state-file.md" "$C/project-configuration.md")
row "  phase files + Load directives" "${FD[@]}"
row "  + cited"                         "${FC[@]}"
row "  + conditional (follow-ups, state-file, config)" "${FX[@]}"
row "  TOTAL definite"                  "${ALWAYS[@]}" "${ROUTER[@]}" "${FD[@]}"
row "  TOTAL incl. cited"               "${ALWAYS[@]}" "${ROUTER[@]}" "${FD[@]}" "${FC[@]}"
echo "--- subagents (each also inherits whatever always-on context the harness gives it) ---"
row "  reviewer: baseline + primary template"    "$R/rules/agent-baseline.md" "$F/primary-reviewer-prompt.md"
row "  reviewer: baseline + principles + EP"     "$R/rules/agent-baseline.md" "$F/principles-reviewer-prompt.md" "$F/engineering-principles.md"
row "  project-configuration.md (cited everywhere)" "$C/project-configuration.md"
row "  state-file.md (pipeline: load before any state r/w)" "$C/state-file.md"
```

## `loaded_files.py` — what each real session actually loaded

Scans Claude Code transcripts (one JSONL file per session; `/clear` starts a new one). For every
`Read`, `Skill` load, or `cat`/`sed`/`head` of a `skills/`, `rules/` or `commands-claude/`
Markdown file it reports the bytes that entered context, the API turn it entered at, re-reads,
and the turn-weighted re-read cost (`bytes/4 × turns remaining`), since every later turn re-reads
the whole context as cache. Report only; always exits 0.

```bash
python3 loaded_files.py                       # every ~/.claude/projects/*/*.jsonl
python3 loaded_files.py ~/.claude/projects/-Users-me-Projects-agents/<session>.jsonl
```

```python
#!/usr/bin/env python3
"""loaded_files.py [transcript.jsonl | dir ...] — which flow instruction files each Claude Code
session actually pulled into context, and how many bytes each cost. One JSONL transcript is one
session (/clear starts a new one). Counts Read results, Skill loads, and Bash cat/sed/head of
*.md(c) under skills/, rules/, commands-claude/. Report only; always exits 0.
Default input: every ~/.claude/projects/*/*.jsonl."""
import collections, json, pathlib, re, sys
PAT = re.compile(r"((?:skills|rules|commands-claude)/[\w./-]+\.mdc?)")
def text_of(content):
    if isinstance(content, str): return content
    if isinstance(content, list):
        return "".join(c.get("text", "") for c in content if isinstance(c, dict))
    return ""
def scan(path):
    pending, loads, turn = {}, [], 0
    for line in open(path, encoding="utf-8", errors="replace"):
        try: ev = json.loads(line)
        except ValueError: continue
        msg = ev.get("message") or {}
        if ev.get("type") == "assistant" and msg.get("usage"): turn += 1
        for c in msg.get("content") or [] if isinstance(msg.get("content"), list) else []:
            if not isinstance(c, dict): continue
            if c.get("type") == "tool_use":
                name, inp = c.get("name"), c.get("input") or {}
                if name == "Read":
                    m = PAT.search(inp.get("file_path", ""))
                    if m: pending[c["id"]] = m.group(1) + ("" if not inp.get("offset") else f" @{inp.get('offset')}")
                elif name == "Skill":
                    pending[c["id"]] = f"<Skill {inp.get('skill')}>"
                elif name == "Bash" and re.search(r"\b(cat|sed|head|tail|awk)\b", inp.get("command", "")):
                    hits = PAT.findall(inp.get("command", ""))
                    if hits: pending[c["id"]] = "bash:" + ",".join(sorted(set(hits)))
            elif c.get("type") == "tool_result" and c.get("tool_use_id") in pending:
                loads.append((pending.pop(c["tool_use_id"]), len(text_of(c.get("content")).encode()), turn))
    return loads, turn
args = sys.argv[1:] or [str(p) for p in pathlib.Path.home().glob(".claude/projects/*/*.jsonl")]
files = []
for a in args:
    p = pathlib.Path(a)
    files += sorted(p.glob("*.jsonl")) if p.is_dir() else [p]
for f in files:
    loads, turns = scan(f)
    if not loads: continue
    total = sum(b for _, b, _ in loads)
    # cache-read weight: every later API turn re-reads what is already in context
    weight = sum(b // 4 * (turns - t) for _, b, t in loads)
    seen = collections.Counter(n.split(" @")[0] for n, _, _ in loads)
    print(f"== {f.name}: {len(loads)} loads, {total} bytes (~{total//4} tok), {turns} API turns, ~{weight/1e6:.1f}M token-turns re-read")
    print(f"   {'bytes':>7}  {'turn':>4}  file")
    for n, b, t in sorted(loads, key=lambda x: -(x[1] * (turns - x[2])))[:25]:
        print(f"   {b:7d}  {t:4d}  {n}")
    rereads = {n: k for n, k in seen.items() if k > 1}
    if rereads: print(f"   re-read: {rereads}")
```

## `verbatim_check.py` — prove a trim moved text and never reworded it

Compares two checkouts sentence by sentence over `skills/`, `commands-claude/` and `rules/`. A
sentence that leaves the run-loaded files must still be stated in one (a de-duplication), or have
landed verbatim in a `-rationale.md` — reported **REVIEW** when it carries an imperative marker,
since a rule must never move out of a run's reach. A new run-loaded sentence must be a load
directive or a citation. Exit 1 on any **FAIL**. It cannot judge whether a lazily loaded file's
"Load `X` only when Y" condition is right; review each directive by hand.

```bash
git worktree add /tmp/base <base-sha>
python3 verbatim_check.py /tmp/base "$PWD"
```

Self-test at `700e184`: moving "This is why no `*-done` command exists — there would be nothing
for one to write." from `pipeline.md` into `pipeline-rationale.md` passes; rewording "**This
table is authoritative.**" and deleting "**A mark never blocks, delays, or alters the stage it
marks.**" each fail.

```python
#!/usr/bin/env python3
"""verbatim_check.py <base-root> <head-root> — prove a prompt trim moved or de-duplicated text
and never reworded it. Scope: skills/**/*.md, commands-claude/*.md, rules/*.md* in each root.
A sentence that left the run-loaded files must (a) still be stated in a run-loaded file, or
(b) have landed verbatim in a *-rationale.md (REVIEW when it carries an imperative marker — a rule
must not be moved out of a run's reach). A sentence that appeared in a run-loaded file must be a
load directive or a citation. Exit 1 on any FAIL; REVIEW lines need a human look.
Limits: it cannot see whether a lazily loaded file's load condition is right — review each
"Load X only when Y" directive by hand."""
import collections, pathlib, re, sys
IMPERATIVE = re.compile(r"\b(never|always|must|only|do not|don't|exactly|every|each|required|refuse[sd]?|stop|before|after|unless|except)\b", re.I)
ALLOWED_NEW = re.compile(r"(\*\*Load `|[Ll]oad `[^`]+` only when|\(`skills/[^`]+\.md`\))")
def files(root):
    r = pathlib.Path(root)
    for pat in ("skills/**/*.md", "commands-claude/*.md", "rules/*.md", "rules/*.mdc"):
        yield from (p for p in r.glob(pat) if p.is_file() and not p.is_symlink())
def blocks(text):
    cur, fence = [], False
    for l in text.split("\n"):
        s = l.strip()
        if s.startswith(("```", "~~~")):
            if cur: yield " ".join(cur); cur = []
            fence = not fence; continue
        if fence: yield s; continue
        s = re.sub(r"^(>\s?)+", "", s)
        if not s or re.match(r"^(#+ |\||---|[-*+] |\d+[.)] )", s):
            if cur: yield " ".join(cur); cur = []
            if s: yield re.sub(r"^([-*+]|\d+[.)])\s+", "", s)
            continue
        cur.append(s)
    if cur: yield " ".join(cur)
def sentences(text):
    for b in blocks(text):
        for s in re.split(r"(?<=[.!?])[)\]\"'`*_]*\s+", b):
            s = re.sub(r"\s+", " ", s).strip()
            if len(s) >= 12: yield s
def corpus(root):
    run, rat = collections.Counter(), collections.Counter()
    for p in files(root):
        (rat if p.name.endswith("-rationale.md") else run).update(sentences(p.read_text(encoding="utf-8")))
    return run, rat
base_run, base_rat = corpus(sys.argv[1]); head_run, head_rat = corpus(sys.argv[2])
bad = 0
for s, n in (base_run - head_run).items():
    if head_run[s] > 0:
        print(f"ok   dedup   still stated in a run-loaded file :: {s[:100]}")
    elif head_rat[s] - base_rat[s] >= n:
        if IMPERATIVE.search(s):
            print(f"REVIEW moved to a -rationale.md but carries an imperative marker :: {s[:100]}")
        else:
            print(f"ok   moved   to a -rationale.md :: {s[:100]}")
    else:
        bad += 1; print(f"FAIL deleted or reworded :: {s[:100]}")
for s, n in (head_run - base_run).items():
    if base_run[s] == 0 and not ALLOWED_NEW.search(s) and base_rat[s] == 0:
        bad += 1; print(f"FAIL new run-loaded text (paraphrase?) :: {s[:100]}")
print(f"verbatim_check: {bad} violation(s)", file=sys.stderr)
sys.exit(1 if bad else 0)
```
