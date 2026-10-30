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
The checker runs in the payload's `cwd`, so a self-review on a change's own branch in another
project files its rows in that project's flowd project, not the agents checkout's.

The PreToolUse contract cannot rewrite tool input, so a refusal is a deny whose reason is the
checker's own lines. Each deny, once written, also records one `flow record incident` row, failing
open: a write that fails or hangs leaves the deny as it stands.

Fails open for anything that is not a loop dispatch -- unparseable input, another tool, a
prompt with no loop key and no loop worktree or branch named. Fails CLOSED for a loop dispatch: a checker that cannot answer, or
cannot run, denies it, because an unchecked loop dispatch is the skip this hook exists to stop.
A prompt that names an open loop's worktree or branch (`pipeline-fix-<name>-<k>`, `self-review-<name>`
whose local branch still exists in the payload `cwd`'s repository or this checkout) with no Key line is
denied too -- a closed loop's name in prose is no dispatch, and a planner's read-only Explore/Plan
sub-dispatch is admitted. Dropping the line must not drop the check -- unless that name is a
`/flow` change's own, a `spectre/<name>` branch, local or on origin, in the payload `cwd`'s
repository: a `/flow` change may itself be named `self-review-...`, and a loop's branch is its bare
name, never `spectre/<name>`. A loop dispatch whose prompt lacks the agent-baseline pointer is
denied before the checker records it, since enforce-agent-baseline.py would deny it after, leaving
its row open.

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

# Read-only agents a loop planner may dispatch inside its loop's worktree.
PLANNER_SUBAGENTS = {"Explore", "Plan"}

BASELINE = "~/.claude/rules/agent-baseline.md"

# Under setup.sh's registered 300s.
CHECK_TIMEOUT = 240

REPO = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))

# The checker the hook runs; overridable for the hook's own harness only.
CHECK = os.environ.get("LOOP_DISPATCH_CHECK") or os.path.join(REPO, "scripts", "check-loop-dispatch.sh")

def bound(seconds):
    """A subprocess call's wall-clock bound. FLOW_HOOK_UNBOUNDED=1 lifts it: the hook's tests
    inject that so the full -race suite starving a stub past the bound never fails a case
    (kan-1060) -- the bound is production's fail-open, not under test."""
    return None if os.environ.get("FLOW_HOOK_UNBOUNDED") == "1" else seconds


# The payload's cwd, set by main for deny's incident row.
PAYLOAD_CWD = None


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
    sys.stdout.flush()
    record_incident("enforce-loop-dispatch", reason, PAYLOAD_CWD)
    return 0


def record_incident(guard, reason, cwd):
    """One `flow record incident` row per deny, so a self-review pass can count the denials;
    FLOW_HOOK_INCIDENT_CMD overrides `flow` for the hook's own harness only. Output is captured
    and every failure swallowed: the deny already written stands either way."""
    lines = reason.strip().splitlines()
    args = [os.environ.get("FLOW_HOOK_INCIDENT_CMD") or "flow", "record", "incident",
            "-guard", guard, "-symptom", (lines[0] if lines else reason)[:300],
            "-recovery", "re-sent per the deny reason", "-minutes-lost", "0"]
    if cwd:
        args += ["-C", cwd]
    try:
        subprocess.run(args, capture_output=True, text=True, timeout=bound(10), check=False)
    except (OSError, ValueError, subprocess.SubprocessError):
        pass


def has_ref(repo, *refs):
    """Does any of refs exist in repo's repository?"""
    try:
        r = subprocess.run(
            ["git", "-C", repo, "for-each-ref", "--count=1", "--format=x", *refs],
            capture_output=True, text=True, timeout=bound(10),
        )
    except (OSError, subprocess.SubprocessError):
        return False
    return r.returncode == 0 and r.stdout.strip() == "x"


def flow_change(name, cwd):
    """Is name a /flow change's own -- a spectre/<name> branch, local or on origin, in cwd's repository?"""
    return has_ref(cwd, f"refs/heads/spectre/{name}", f"refs/remotes/origin/spectre/{name}")


def open_loop(name, cwd):
    """Is name an open loop -- its branch still present in cwd's repository or this checkout's?

    A loop's branch is removed when the loop closes, so a prose mention of a finished loop is no dispatch."""
    return any(has_ref(r, f"refs/heads/{name}") for r in (cwd, REPO) if r)


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
    cwd = payload.get("cwd")
    cwd = cwd if isinstance(cwd, str) and os.path.isdir(cwd) else None
    global PAYLOAD_CWD
    PAYLOAD_CWD = cwd
    m = KEY_RE.search(prompt)
    if not m:
        # A loop planner's sanctioned read-only sub-dispatch names the loop's worktree, never its key.
        if tool_input.get("subagent_type") in PLANNER_SUBAGENTS:
            return 0
        for named in LOOP_NAME_RE.finditer(prompt):
            if cwd and flow_change(named.group(0), cwd):
                continue
            if not open_loop(named.group(0), cwd):
                continue
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
        r = subprocess.run(args, input=prompt, capture_output=True, text=True, timeout=bound(CHECK_TIMEOUT), cwd=cwd)
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
