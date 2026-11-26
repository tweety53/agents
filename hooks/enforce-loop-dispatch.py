#!/usr/bin/env python3
"""PreToolUse hook: record a pipeline-fix or self-review fix-loop dispatch's row. It never denies.

A loop dispatch is an Agent/Task call whose prompt carries a `Key: pipeline-fix-<k>-...` or
`Key: self-review-...` line (skills/flow-contracts/pipeline.md **Pipeline defects found
mid-run**, skills/flow-self-review/SKILL.md step 3). This hook hands every such call to
flow-guard check-loop-dispatch in the agents checkout it is installed from, which opens the
dispatch's row in flowd -- the row the per-role cost stats read. The checker runs in the
payload's `cwd`, so a self-review on a change's own branch in another project files its rows in
that project's flowd project, not the agents checkout's.

Recording is best-effort and never blocks: a checker that cannot answer, cannot run or times out
is logged on stderr and the Agent call proceeds. A loop dispatch whose prompt lacks the
agent-baseline pointer is not recorded, since enforce-agent-baseline.py denies it after, which
would leave its row open. Anything that is not a loop dispatch -- unparseable input, another
tool, a prompt with no valid loop key -- passes untouched.

The checker's wait stays well under the 300s timeout setup.sh registers.
"""

import json
import os
import re
import subprocess
import sys

from harness_io import emit_pre_tool_allow, get_tool_call

DISPATCH_TOOLS = {"Agent", "Task", "invoke_subagent"}

# Any Key line whose value starts as a loop key; the value must then be a whole VALID_KEY.
KEY_RE = re.compile(r"^[ \t]*Key:[ \t]*((?:pipeline-fix|self-review)-.*?)[ \t]*$", re.M)
VALID_KEY = re.compile(r"pipeline-fix-[0-9]+-\S+|self-review-\S+-\S+")

BASELINES = (
    "~/.claude/rules/agent-baseline.md",
    "~/.gemini/config/plugins/flow/rules-full/agent-baseline.md",
)

# Under setup.sh's registered 300s.
CHECK_TIMEOUT = 240

REPO = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))

# The checker the hook runs; overridable for the hook's own harness only.
CHECK = [os.environ["LOOP_DISPATCH_CHECK"]] if os.environ.get("LOOP_DISPATCH_CHECK") else [
    os.path.join(REPO, "scripts", "flow-guard"), "check-loop-dispatch"
]


# A prompt's `Role:` line: the allowed-pairs role the row is counted under in Role costs.
ROLE_LINE = re.compile(r"^Role:[ \t]*(\S+)[ \t]*$", re.M)


def bound(seconds):
    """A subprocess call's wall-clock bound. FLOW_HOOK_UNBOUNDED=1 lifts it: the hook's tests
    inject that so the full -race suite starving a stub past the bound never fails a case
    (kan-1060)."""
    return None if os.environ.get("FLOW_HOOK_UNBOUNDED") == "1" else seconds


def main():
    try:
        payload = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0
    if not isinstance(payload, dict):
        return 0
    tool_name, tool_input, _ = get_tool_call(payload)
    if tool_name not in DISPATCH_TOOLS:
        emit_pre_tool_allow(payload)
        return 0
    prompt = tool_input.get("prompt") if isinstance(tool_input.get("prompt"), str) else tool_input.get("Prompt")
    if not isinstance(prompt, str):
        emit_pre_tool_allow(payload)
        return 0
    cwd = payload.get("cwd")
    cwd = cwd if isinstance(cwd, str) and os.path.isdir(cwd) else None
    m = KEY_RE.search(prompt)
    if not m or not VALID_KEY.fullmatch(m.group(1)) or not any(b in prompt for b in BASELINES):
        emit_pre_tool_allow(payload)
        return 0
    key = m.group(1)
    role = ROLE_LINE.search(prompt)
    subagent_type = tool_input.get("subagent_type") or tool_input.get("TypeName") or ""
    model = tool_input.get("model") or tool_input.get("Model") or ""
    args = [*CHECK, REPO, key, str(subagent_type), str(model),
            role.group(1) if role else ""]
    try:
        r = subprocess.run(args, capture_output=True, text=True, timeout=bound(CHECK_TIMEOUT), cwd=cwd)
    except (OSError, subprocess.SubprocessError) as e:
        print(f"enforce-loop-dispatch: {key} not recorded: flow-guard check-loop-dispatch could not run ({e})", file=sys.stderr)
        emit_pre_tool_allow(payload)
        return 0
    if r.returncode != 0:
        print(f"enforce-loop-dispatch: {key} not recorded (flow-guard check-loop-dispatch exit {r.returncode}):\n{r.stdout}{r.stderr}", file=sys.stderr)
    emit_pre_tool_allow(payload)
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception:  # recording is best-effort: never block the dispatch
        sys.exit(0)
