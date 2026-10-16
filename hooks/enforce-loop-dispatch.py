#!/usr/bin/env python3
"""PreToolUse hook: admit a pipeline-fix or self-review fix-loop dispatch only off its row.

A loop dispatch is an Agent/Task call whose prompt carries a `Key: pipeline-fix-<k>-...` or
`Key: self-review-...` line (skills/flow-contracts/pipeline.md **Pipeline defects found
mid-run**, skills/flow-self-review/SKILL.md step 3). The loops were prose only, and a parent
skipped their steps: fixers dispatched on flow-high without a planner, fixes landed with no
review. This hook hands every such call to scripts/check-loop-dispatch.sh in the agents
checkout it is installed from, which refuses a pair off the canonical table, a fix with no
completed plan, a dispatch while an earlier one is still open -- and records the admitted
dispatch in flowd, so check-loop-landing.sh can refuse a landing those records do not cover.

The PreToolUse contract cannot rewrite tool input, so a refusal is a deny whose reason is the
checker's own lines.

Fails open for anything that is not a loop dispatch -- unparseable input, another tool, a
prompt with no loop key and no loop worktree or branch named. Fails CLOSED for a loop dispatch: a checker that cannot answer, or
cannot run, denies it, because an unchecked loop dispatch is the skip this hook exists to stop.
A prompt that names a loop's worktree or branch (`pipeline-fix-<name>-<k>`, `self-review-<name>`)
with no Key line is denied too: dropping the line must not drop the check. A loop dispatch whose
prompt lacks the agent-baseline pointer is denied before the checker records it, since
enforce-agent-baseline.py would deny it after, leaving its row open.

The checker's wait stays well under the 300s timeout setup.sh registers, so a slow checker is
denied here rather than the harness timing the hook out, which fails open.
"""

import json
import os
import re
import subprocess
import sys

DISPATCH_TOOLS = {"Agent", "Task"}

# Any Key line whose value starts as a loop key; the value must then be a whole VALID_KEY.
KEY_RE = re.compile(r"^[ \t]*Key:[ \t]*((?:pipeline-fix|self-review)-.*?)[ \t]*$", re.M)
VALID_KEY = re.compile(r"pipeline-fix-[0-9]+-\S+|self-review-\S+-\S+")

# A loop's branch or worktree name, the way a prompt that dropped its Key line still names it.
LOOP_NAME_RE = re.compile(r"(?<![\w-])(?:pipeline-fix-[\w.-]+-[0-9]+|self-review-[\w.-]+)(?![\w-])")

BASELINE = "~/.claude/rules/agent-baseline.md"

# Under setup.sh's registered 300s.
CHECK_TIMEOUT = 240

REPO = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))

# The checker the hook runs; overridable for the hook's own harness only.
CHECK = os.environ.get("LOOP_DISPATCH_CHECK") or os.path.join(REPO, "scripts", "check-loop-dispatch.sh")


def deny(reason):
    json.dump(
        {
            "hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "deny",
                "permissionDecisionReason": reason,
            }
        },
        sys.stdout,
    )
    return 0


def main():
    try:
        payload = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0
    if not isinstance(payload, dict) or payload.get("tool_name") not in DISPATCH_TOOLS:
        return 0
    tool_input = payload.get("tool_input")
    if not isinstance(tool_input, dict) or not isinstance(tool_input.get("prompt"), str):
        return 0
    prompt = tool_input["prompt"]
    m = KEY_RE.search(prompt)
    if not m:
        named = LOOP_NAME_RE.search(prompt)
        if named:
            return deny(
                f"Loop dispatch blocked: the prompt names {named.group(0)} but carries no "
                "`Key: pipeline-fix-<k>-<suffix>` or `Key: self-review-<name>-<suffix>` line."
            )
        return 0
    key = m.group(1)
    if not VALID_KEY.fullmatch(key):
        return deny(
            f"Loop dispatch blocked: `Key: {key}` is no valid loop key — "
            "`pipeline-fix-<k>-<suffix>` (numeric <k>) or `self-review-<name>-<suffix>`, nothing after it."
        )
    if BASELINE not in prompt:
        return deny(f"Loop dispatch {key} blocked: the prompt carries no {BASELINE} pointer.")
    args = [CHECK, REPO, key, str(tool_input.get("subagent_type") or ""), str(tool_input.get("model") or "")]
    try:
        r = subprocess.run(args, capture_output=True, text=True, timeout=CHECK_TIMEOUT)
    except (OSError, subprocess.SubprocessError) as e:
        return deny(f"Loop dispatch {key} blocked: check-loop-dispatch.sh could not run ({e}).")
    if r.returncode == 0:
        return 0
    return deny(
        f"Loop dispatch {key} blocked (check-loop-dispatch.sh exit {r.returncode}):\n"
        f"{r.stdout}{r.stderr}"
    )


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception:  # only reached before a loop key is known: fail open
        sys.exit(0)
