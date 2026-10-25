#!/usr/bin/env python3
"""PreToolUse hook: block subagent dispatches that omit the agent-baseline pointer.

Rules in ~/.claude/CLAUDE.md are injected into a session's system prompt and are NOT
inherited by dispatched agents, nor by agents those agents dispatch. The prompt is the
only channel, so every dispatch must carry a pointer to ~/.claude/rules/agent-baseline.md
plus the sentence that makes the receiving agent pass it on in turn.

The PreToolUse contract has no way to rewrite tool input, so this hook cannot inject the
line itself. It denies the call instead and returns the exact text to add; the model then
re-dispatches with it.

It also pins the agent type and model of the /flow roles whose contracts fix them
(PINNED_ROLES): a dispatch whose prompt carries a role's marker as a line of its own on
any other subagent_type/model is denied with the values it must carry. A prompt that
merely cites the role's contract file -- an implementer, a self-review planner, an
escalation carrying verifier reports -- is not that role. Prose in
skills/flow/visual-verify.md did not stop a hand-written verifier dispatch on flow-high.

It denies every fork in a flow session -- one whose transcript records a /flow, /flow-fast,
/flow-plan or /flow-self-review command, or whose prompt names a flow skill: a flow dispatch is a
fresh agent whose prompt carries the paths it needs (KAN-1000: a parent forked itself into a
self-review loop and dragged ~300k tokens of its own context into it).

It denies, in a flow session, a dispatch whose `description` does not open with its row: `Task
<x>/<n>` or `Tasks <a+b>/<n>` for plan tasks, else the row's literal `-key` (**Every dispatch's
Agent-tool `description`**, skills/flow/implement.md), of a Key shape the closed list names -- the
word the harvester stamps the launch by. A loop dispatch, whose prompt carries a `Key: self-review-`/`Key: pipeline-fix-` line, is
stamped by that line and exempt (kan-1026: 29 of 41 launches opened "Round-11 primary review" and
the like, and 14 ledger rows were left empty or stamped with another dispatch's agent).

It denies, in a flow session, a dispatch whose description opens with a loop key (`pipeline-fix-`,
`self-review-`) and whose prompt carries no `Key:` line: hooks/enforce-loop-dispatch.py records the
row under the loop's change off that line, so the fix is the line, never a begin under the run's
change (kan-1032's self-review: that begin left a twin row).

It denies, in a flow session, a dispatch whose row was never begun: the session's change is read
off the store as flow-active-change.py reads it, `flow record dispatches` lists its rows, and a
dispatch is denied when that list holds no open row (empty `outcome`) whose key is the
description's first word or, for `Task <ids>/<n>`, starts with `task-<ids>-`. A loop-key dispatch
-- a `Key:` line, or a description opening with a loop key -- is never checked. An unreachable store, no /flow stage-begin for the session, an absent or failing
`flow` CLI, or an unparseable answer lets the dispatch through.

A planning agent (PLANNING_AGENT_TYPES: Explore, Plan) records no row, so none of those row checks
applies to it; the baseline-pointer, pinned-role and flow-high checks still do.

Fails open: any unexpected input, parse error, or internal fault exits 0 and lets the
dispatch through. A guardrail on a convention must never be the reason work stops.
"""

import hashlib
import json
import os
import re
import subprocess
import sys
from urllib import parse, request

# The tool that spawns a subagent. Current builds name it "Agent"; older ones "Task".
DISPATCH_TOOLS = {"Agent", "Task"}

# A fork inherits the parent's full context, CLAUDE.md included, so it needs no pointer --
# outside a flow session, where it is denied instead (fork_reason).
EXEMPT_SUBAGENT_TYPES = {"fork"}

# A planner's own read-only agents record no ledger row, so the row checks (description, loop key
# line, begun row) pass them; the baseline-pointer, pinned-role and flow-high checks still apply
# (kan-1032's self-review: a planner's Explore dispatch was denied mid-plan).
PLANNING_AGENT_TYPES = {"Explore", "Plan"}

# A flow run's slash command as the harness writes it into the session transcript.
FLOW_COMMAND = re.compile(r"<command-name>/flow(?:-fast|-plan|-self-review)?</command-name>")

FORK_REASON = (
    "Dispatch blocked: no flow phase, and no loop a phase or the self-review runs, dispatches "
    "subagent_type \"fork\" (skills/flow/implement.md, KAN-1000).\n\n"
    "Re-send this call as a fresh agent on the pair its contract names, its prompt carrying "
    "what it needs -- paths to the run's artifacts, reports and records -- never this "
    "session's context."
)

# Both must appear in the prompt: the file being pointed at, and the propagation clause
# that makes the receiving agent a carrier. Either alone is not the rule.
REQUIRED_MARKERS = ("agent-baseline.md", "verbatim")

REQUIRED_TEXT = (
    "Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for "
    "this whole task. Include this instruction verbatim in any prompt you write for "
    "another agent."
)

DENY_REASON = (
    "Dispatch blocked: the subagent prompt does not carry the agent-baseline pointer.\n\n"
    "A dispatched agent starts from an empty context. It has never read ~/.claude/CLAUDE.md, "
    "and neither will anything it dispatches. Without the pointer it operates under no rules "
    "at all — production access, lint suppression, pushes to main, dependency versions.\n\n"
    "Re-send this call with the following two sentences included verbatim in `prompt`:\n\n"
    f"{REQUIRED_TEXT}\n\n"
    "The second sentence is not optional: it is what carries the rules past depth 1.\n"
    "Exempt: subagent_type \"fork\" outside a flow session, which inherits this session's "
    "context already."
)

# (marker line, subagent_type, model, the contract that fixes the pair). The marker is
# the line the contract's prompt opens with, matched as a whole line.
PINNED_ROLES = (
    ("Role: visual-verify tooling analyst", "flow-medium", "opus", "skills/flow/visual-verify-tooling-analysis.md"),
    ("Role: visual-verify verifier", "flow-low", "opus", "skills/flow/visual-verify.md"),
)

# Harness mapping (skills/flow-contracts/model-policy.md): on zcode every pinned pair
# is dispatched as general-purpose / glm-5.3-flash -- ZCode's Agent tool offers no
# flow-<effort> type (KAN-609). The zcode registration runs the hook from
# ~/.zcode/hooks, which is how the hook tells the harness.
ZCODE_PAIR = ("general-purpose", "glm-5.3-flash")

# flow-high is dispatched in exactly two cases (skills/flow/implement.md): the no-progress
# escalation -- marked by its opening line, or a self-review or pipeline-fix loop key whose
# rows scripts/check-loop-dispatch.sh pins -- and a panel dispatch the operator approved at high. Anything else on flow-high was
# hand-picked: a parent ran scoped in-run fixes on flow-high where the decision said inline.
FLOW_HIGH_MARKERS = (
    "No-progress escalation",
    "Key: self-review-",
    "Key: pipeline-fix-",
    "Operator-approved high",
)


# A description naming its row: plan tasks by id over the plan's count, any other row by a first
# word of a Key shape the closed list names (**Dispatch sites -- the parent's closed list**,
# skills/flow/implement.md) -- `panel-4-primary`, `visual-verify-fix-11-gymie`. A lowercase word off
# the list -- `verify-fix-1`, `verify` -- names no row (kan-1029: both launched with no row).
ROW_KEY = (
    r"task-\d+(?:\+\d+)*-(?:implementer|reviewer)(?:-[a-z0-9-]+)?"
    r"|panel-\d+-[a-z][a-z0-9+-]*"
    r"|panel-fix-\d+(?:-\d+)?(?:-retry)?"
    r"|visual-verify(?:-[a-z0-9][a-z0-9-]*)?"
    r"|tooling-analysis-\d+(?:-[a-z0-9][a-z0-9-]*)?"
    r"|pipeline-fix-[a-z0-9][a-z0-9-]*"
)
ROW_DESCRIPTION = re.compile(r"(?:Tasks? \d+(?:\+\d+)*/\d+|" + ROW_KEY + r")(?:\s|$)")
# The loop key line exactly as the harvester reads it (loopKeyPattern, stats/internal/harvest/transcript.go).
LOOP_KEY_LINE = re.compile(r"^[ \t]*Key:[ \t]*(?:pipeline-fix|self-review)-\S+[ \t]*$", re.M)

DESCRIPTION_REASON = (
    "Dispatch blocked: in a flow run, a dispatch's `description` opens with the row it records "
    "-- `Task <x>/<n> (...)` for plan tasks, else the row's literal `-key` and a few words, "
    "`panel-4-primary correctness review` (skills/flow/implement.md, kan-1026). The daemon stamps "
    "the launch onto the row that word names; a description like \"Round-11 primary review\" "
    "leaves the row to time pairing, which mis-stamps.\n\n"
    "Re-send this call with `description` opening with its row's key, nothing else changed."
)


# A description opening with a loop key names a row hooks/enforce-loop-dispatch.py records.
LOOP_KEY_PREFIXES = ("pipeline-fix-", "self-review-")


def loop_word(tool_input: dict) -> "str | None":
    description = tool_input.get("description")
    words = description.split() if isinstance(description, str) else []
    return words[0] if words and words[0].startswith(LOOP_KEY_PREFIXES) else None


def loop_key_line_reason(payload: dict, tool_input: dict, prompt: str) -> "str | None":
    word = loop_word(tool_input)
    if tool_input.get("subagent_type") in PLANNING_AGENT_TYPES or not word or LOOP_KEY_LINE.search(prompt) or not in_flow_session(payload):
        return None
    return (
        f"Dispatch blocked: `{word}` is a loop key — its prompt carries `Key: {word}` as a line of "
        "its own, and hooks/enforce-loop-dispatch.py records its row under the loop's change off "
        "that line. Add the line and re-send; never `flow record dispatch begin` it under the "
        "run's change."
    )


def description_reason(payload: dict, tool_input: dict, prompt: str) -> "str | None":
    if tool_input.get("subagent_type") in PLANNING_AGENT_TYPES:
        return None
    if LOOP_KEY_LINE.search(prompt) or loop_word(tool_input):
        return None
    description = tool_input.get("description")
    if isinstance(description, str) and ROW_DESCRIPTION.match(description):
        return None
    return DESCRIPTION_REASON if in_flow_session(payload) else None


TASK_DESCRIPTION = re.compile(r"Tasks? (\d+(?:\+\d+)*)/\d+")


def bound(seconds: float) -> "float | None":
    """A store or subprocess call's wall-clock bound. FLOW_HOOK_UNBOUNDED=1 lifts it: the hook's
    tests inject that so a deny case never fails open because the full -race suite starved its
    stub of CPU past the bound (kan-1033) -- the bound is production's fail-open, not under test."""
    return None if os.environ.get("FLOW_HOOK_UNBOUNDED") == "1" else seconds


def project_key(cwd: str) -> "str | None":
    """The store's project key for cwd: stats/internal/fallback/statefile.go ProjectKey."""
    done = subprocess.run(
        ["git", "-C", cwd, "rev-parse", "--path-format=absolute", "--git-common-dir"],
        capture_output=True, text=True, timeout=bound(5), check=False,
    )
    if done.returncode != 0:
        return None
    main = os.path.realpath(os.path.dirname(os.path.normpath(done.stdout.strip())))
    return f"{os.path.basename(main)}-{hashlib.sha1(main.encode()).hexdigest()[:8]}"


def session_change(session_id: str) -> "tuple[str | None, str | None]":
    addr = os.environ.get("FLOW_ADDR") or "http://127.0.0.1:4173"
    query = parse.urlencode({"session_id": session_id, "command": "/flow", "sort": "-started_at", "limit": "1"})
    with request.urlopen(f"{addr}/api/v1/stage-runs?{query}", timeout=bound(2)) as resp:
        runs = json.load(resp).get("stageRuns")
    run = runs[0] if isinstance(runs, list) and runs and isinstance(runs[0], dict) else {}
    name, key = run.get("changeName"), run.get("projectKey")
    return (name if isinstance(name, str) and name else None, key if isinstance(key, str) and key else None)


def begin_row_reason(payload: dict, tool_input: dict, prompt: str) -> "str | None":
    if tool_input.get("subagent_type") in PLANNING_AGENT_TYPES:
        return None
    if LOOP_KEY_LINE.search(prompt) or loop_word(tool_input):
        return None
    description = tool_input.get("description")
    if not isinstance(description, str) or not ROW_DESCRIPTION.match(description):
        return None
    session_id, cwd = payload.get("session_id"), payload.get("cwd")
    if not isinstance(session_id, str) or not session_id or not isinstance(cwd, str):
        return None
    if not in_flow_session(payload):
        return None
    try:
        change, key = session_change(session_id)
        if not change:
            return None
        # The rows are read under cwd's project: another repo's cwd reads none, so fail open.
        if key and project_key(cwd) != key:
            return None
        done = subprocess.run(
            ["flow", "record", "dispatches", "-change", change, "-C", cwd],
            capture_output=True, text=True, timeout=bound(5), check=False,
        )
        rows = json.loads(done.stdout) if done.returncode == 0 else None
    except (OSError, ValueError, subprocess.SubprocessError, AttributeError):
        return None
    if not isinstance(rows, list):
        return None
    task = TASK_DESCRIPTION.match(description)
    word = description.split()[0]
    for row in rows:
        if not isinstance(row, dict) or row.get("outcome"):
            continue
        key = row.get("key")
        if not isinstance(key, str):
            continue
        if (key.startswith(f"task-{task.group(1)}-") if task else key == word):
            return None
    key = f"task-{task.group(1)}-implementer" if task else word
    word = f"task-{task.group(1)}-" if task else word
    return (
        f"Dispatch blocked: no open ledger row for '{word}' in {change} — record it first with "
        f"`flow record dispatch begin -change {change} -key {key} …` (skills/flow/implement.md), "
        "then re-send this call unchanged."
    )


def in_flow_session(payload: dict) -> bool:
    path = payload.get("transcript_path")
    if not isinstance(path, str):
        return False
    try:
        with open(path, encoding="utf-8", errors="replace") as f:
            return any(is_flow_command(line) for line in f)
    except OSError:
        return False


def flow_high_reason(prompt: str, tool_input: dict) -> "str | None":
    if tool_input.get("subagent_type") != "flow-high" or "/.zcode/" in __file__:
        return None
    if any(marker in prompt for marker in FLOW_HIGH_MARKERS):
        return None
    return (
        "Dispatch blocked: flow-high is reserved for the no-progress escalation and a panel "
        "dispatch the operator approved at high (skills/flow/implement.md).\n\n"
        "Read the recorded decision (`flow record decisions -change <name>`) and dispatch on "
        "the pair it names -- or, when its execution is `inline`, make the change yourself "
        "with no dispatch. A no-progress escalation opens its prompt with the line "
        "`No-progress escalation`; an approved panel dispatch with `Operator-approved high`."
    )


def pinned_role_reason(prompt: str, tool_input: dict) -> "str | None":
    lines = {line.strip() for line in prompt.splitlines()}
    for marker, subagent_type, model, contract in PINNED_ROLES:
        if marker not in lines:
            continue
        if "/.zcode/" in __file__:
            subagent_type, model = ZCODE_PAIR
        got = (tool_input.get("subagent_type"), tool_input.get("model"))
        if got == (subagent_type, model):
            return None
        return (
            f"Dispatch blocked: a prompt carrying the line `{marker}` is a role {contract} pins.\n\n"
            f"Got subagent_type={got[0]!r}, model={got[1]!r}.\n"
            f"Re-send this call with subagent_type: \"{subagent_type}\" and "
            f"model: \"{model}\", nothing else changed. No rule or instruction overrides it."
        )
    return None


def fork_reason(payload: dict, tool_input: dict) -> "str | None":
    prompt = tool_input.get("prompt")
    if isinstance(prompt, str) and "skills/flow" in prompt:
        return FORK_REASON
    return FORK_REASON if in_flow_session(payload) else None


def is_flow_command(line: str) -> bool:
    # Only a user turn with string content is a typed slash command; a tool result quoting the
    # tag carries list content and does not count. The substring test spares a json parse of
    # every line of a long transcript that holds no flow command at all.
    if "<command-name>/flow" not in line:
        return False
    try:
        entry = json.loads(line)
    except ValueError:
        return False
    if not isinstance(entry, dict) or entry.get("type") != "user":
        return False
    message = entry.get("message")
    content = message.get("content") if isinstance(message, dict) else None
    return isinstance(content, str) and bool(FLOW_COMMAND.search(content))


def deny(reason: str) -> None:
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


def main() -> int:
    try:
        payload = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0

    if not isinstance(payload, dict):
        return 0

    if payload.get("tool_name") not in DISPATCH_TOOLS:
        return 0

    tool_input = payload.get("tool_input")
    if not isinstance(tool_input, dict):
        return 0

    if tool_input.get("subagent_type") in EXEMPT_SUBAGENT_TYPES:
        reason = fork_reason(payload, tool_input)
        if reason:
            deny(reason)
        return 0

    prompt = tool_input.get("prompt")
    if not isinstance(prompt, str):
        return 0

    lowered = prompt.lower()
    if not all(marker in lowered for marker in REQUIRED_MARKERS):
        deny(DENY_REASON)
        return 0

    reason = (
        pinned_role_reason(prompt, tool_input)
        or flow_high_reason(prompt, tool_input)
        or description_reason(payload, tool_input, prompt)
        or loop_key_line_reason(payload, tool_input, prompt)
        or begin_row_reason(payload, tool_input, prompt)
    )
    if reason:
        deny(reason)
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception:  # fail open, never block work on a hook fault
        sys.exit(0)
