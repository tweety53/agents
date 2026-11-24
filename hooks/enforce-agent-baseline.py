#!/usr/bin/env python3
"""PreToolUse hook: block subagent dispatches that omit the agent-baseline pointer.

Rules in ~/.claude/CLAUDE.md are injected into a session's system prompt and are NOT
inherited by dispatched agents, nor by agents those agents dispatch. The prompt is the
only channel, so every dispatch must carry a pointer to ~/.claude/rules/agent-baseline.md
plus the sentence that makes the receiving agent pass it on in turn.

The PreToolUse contract has no way to rewrite tool input, so this hook cannot inject the
line itself. It denies the call instead and returns the exact text to add; the model then
re-dispatches with it.

It denies every fork in a flow session -- one whose transcript records a /flow, /flow-fast,
/flow-interactive, /flow-plan or /flow-self-review command, or whose prompt names a flow skill: a
flow dispatch is a fresh agent whose prompt carries the paths it needs (KAN-1000: a parent forked
itself into a self-review loop and dragged ~300k tokens of its own context into it).

It denies, in a flow session, a dispatch whose description opens with a loop key (`pipeline-fix-`,
`self-review-`) and whose prompt carries no `Key:` line: hooks/enforce-loop-dispatch.py records the
row under the loop's change off that line, so the fix is the line, never a begin under the run's
change (kan-1032's self-review: that begin left a twin row).

It begins, best-effort, in a flow session, a dispatch's ledger row when none is open (KAN-1115,
decision hook-same-file): the session's change and session token are read off the store as its
latest open stage run, else its latest, of any command, skipping a loop's own runs -- a change
named `pipeline-fix-...` or `self-review-...`, whose stage runs check-loop-dispatch's begins open
in the parent's session; `flow record dispatches` lists its rows, and when none is open (empty
`outcome`) under the dispatch's key the hook runs `flow record dispatch begin` itself -- the key
the prompt's column-0 `Key:` line, else the description's first word; the role the prompt's
`Role:` line mapped back as launchRoleID (stats/cmd/flow/record.go) maps it; the call's model and
effort. The begin is written only after every deny check has passed, so a
launch the hook denies leaves no row. The row's end stays the parent's (decision
begin-moves-end-stays). The ledger never denies a dispatch: a dispatch with no key or no mapped
role, a begin `flow` refuses, an unreachable store, no stage begin for the session, a stage run
with no session token, an absent or failing `flow` CLI, or an unparseable answer lets the dispatch
through with no begin. Nothing is begun once the session's change is FINISHED in the store.

A planning agent (PLANNING_AGENT_TYPES: Explore, Plan) is begun under its `explore-<n>` or
`design-<n>` key only while the session's latest stage run is `flow.plan` and its prompt
names the session change's worktree, <main-checkout>-worktrees/<change> (decisions
planning-agents-hook-begun, planning-agent-begins-in-brainstorm); a loop planner's agents -- naming
the loop's worktree, or the run's own outside `flow.plan` -- keep the planner's explicit begin.

It denies, in a flow session, a `flow-*` dispatch whose prompt has no `Role: <id>` line, whose id is
not a role of `flow policy get`, or whose `<model>/<subagent_type>` pair is in neither the role's base
nor its optin list: the deny names the role's admitted pairs, or the roles. A `flow policy get` that
fails lets the dispatch through (skills/flow-contracts/model-policy.md, **Allowed pairs**).

Each deny, once written, also records one `flow record incident` row, with the session's change
where the store names one; that write fails open, leaving the deny as it stands.

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

# A planner's own read-only agents pass the loop key line check, and are begun only when their
# prompt names the session change's worktree (begin_row); the baseline-pointer check still applies
# (kan-1032's self-review: a planner's Explore dispatch was denied mid-plan).
PLANNING_AGENT_TYPES = {"Explore", "Plan"}

# A flow run's slash command as the harness writes it into the session transcript.
FLOW_COMMAND = re.compile(r"<command-name>/flow(?:-fast|-interactive|-plan|-self-review)?</command-name>")
FLOW_SKILLS = ("flow", "flow-fast", "flow-interactive", "flow-plan", "flow-self-review")

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


# A key the hook begins a row under, by a Key shape the closed list names (**Dispatch sites -- the
# parent's closed list**, skills/flow/implement.md) -- `panel-4-primary`, `visual-verify-fix-11-gymie`,
# `full-suite-fix-2`, a fix's implementer `fix-2-implementer` or reviewer `fix-2-review`, a planner's `plan-2` or `plan-short-3`, its designer's `designer-1`, a /flow-interactive
# request's operator-asked `operator-58`.
ROW_KEY = (
    r"task-\d+(?:\+\d+)*-(?:implementer|reviewer)(?:-[a-z0-9-]+)?"
    r"|panel-\d+-[a-z][a-z0-9+-]*"
    r"|panel-fix-\d+(?:-\d+)?(?:-retry)?"
    r"|fix-\d+-(?:implementer|review)"
    r"|full-suite-fix-\d+"
    r"|runner-\d+"
    r"|visual-verify(?:-[a-z0-9][a-z0-9-]*)?"
    r"|pipeline-fix-[a-z0-9][a-z0-9-]*"
    r"|plan-\d+"
    r"|designer-\d+"
    r"|plan-short-\d+"
    r"|operator-\d+(?:-[a-z0-9][a-z0-9-]*)?"
)
# The loop key line exactly as the harvester reads it (loopKeyPattern, stats/internal/harvest/transcript.go).
LOOP_KEY_LINE = re.compile(r"^[ \t]*Key:[ \t]*(?:pipeline-fix|self-review)-\S+[ \t]*$", re.M)


# A description opening with a loop key names a row hooks/enforce-loop-dispatch.py records.
LOOP_KEY_PREFIXES = ("pipeline-fix-", "self-review-")


def loop_word(tool_input: dict) -> "str | None":
    description = tool_input.get("description")
    words = description.split() if isinstance(description, str) else []
    return words[0] if words and words[0].startswith(LOOP_KEY_PREFIXES) else None


def loop_key_line_reason(payload: dict, tool_input: dict, prompt: str) -> "str | None":
    word = loop_word(tool_input)
    if (
        tool_input.get("subagent_type") in PLANNING_AGENT_TYPES
        or not word
        or LOOP_KEY_LINE.search(prompt)
        or not in_flow_run(payload)
    ):
        return None
    return (
        f"Dispatch blocked: `{word}` is a loop key — its prompt carries `Key: {word}` as a line of "
        "its own, and hooks/enforce-loop-dispatch.py records its row under the loop's change off "
        "that line. Add the line and re-send; never `flow record dispatch begin` it under the "
        "run's change."
    )


# A key the hook begins a row under: a closed-list key, or a planner's Explore/Plan agent's.
BEGIN_KEY = re.compile(r"(?:" + ROW_KEY + r"|explore-\d+|design-\d+)")
# A `Key:` line of any slug names a row (a parent's own `fix-visual-4` too, never left unrecorded); a loop key is no row of this hook's.
AD_HOC_KEY = re.compile(r"[a-z0-9][a-z0-9+-]*")
KEY_LINE = re.compile(r"^Key:[ \t]*(\S+)[ \t]*$", re.M)
ROLE_LINE = re.compile(r"^Role:[ \t]*(\S+)[ \t]*$", re.M)
SINGLE_TASK_KEY = re.compile(r"task-(\d+)-")
FLOW_EFFORT = re.compile(r"flow-(?:planner-|ui-)?([a-z]+)")
# A prompt's `Role:` id -> the row role `flow record dispatch begin -role` takes: the inverse of
# launchRoleID (stats/cmd/flow/record.go).
ROW_ROLE = {
    "implementer": "implementer", "operator-asked": "implementer",
    "gated-reviewer": "reviewer", "gated-re-reviewer": "reviewer",
    "panel-reviewer": "reviewer", "panel-rerun": "reviewer",
    "panel-merger": "merger",
    "panel-fix": "panel-fix", "minor-fix": "panel-fix", "no-progress-fixer": "panel-fix",
    "visual-verifier": "verifier", "runner": "verifier",
    "planner": "planner", "plan-designer": "planner", "designer": "planner",
    "planner-short": "planner",
    "explorer": "explorer",
}


# Every role a begun row maps takes the pair judge's pick.
JUDGED_ROLES = set(ROW_ROLE)


def bound(seconds: float) -> "float | None":
    """A store or subprocess call's wall-clock bound. FLOW_HOOK_UNBOUNDED=1 lifts it: the hook's
    tests inject that so a deny case never fails open because the full -race suite starved its
    stub of CPU past the bound (kan-1033) -- the bound is production's fail-open, not under test."""
    return None if os.environ.get("FLOW_HOOK_UNBOUNDED") == "1" else seconds


def main_checkout(cwd: str) -> "str | None":
    """cwd's main checkout, from its git common dir."""
    done = subprocess.run(
        ["git", "-C", cwd, "rev-parse", "--path-format=absolute", "--git-common-dir"],
        capture_output=True, text=True, timeout=bound(5), check=False,
    )
    if done.returncode != 0:
        return None
    return os.path.realpath(os.path.dirname(os.path.normpath(done.stdout.strip())))


def project_key(cwd: str, main: "str | None" = None) -> "str | None":
    """The store's project key for cwd: stats/internal/fallback/statefile.go ProjectKey. A caller
    that already holds cwd's main checkout passes it as main."""
    main = main or main_checkout(cwd)
    if main is None:
        return None
    return f"{os.path.basename(main)}-{hashlib.sha1(main.encode()).hexdigest()[:8]}"


def session_change(session_id: str) -> "tuple[str | None, str | None, str | None, str | None]":
    """The session's latest open stage run, else its latest stage run, of any command, skipping a
    loop's own runs -- a change named `pipeline-fix-...` or `self-review-...`, whose stage runs
    check-loop-dispatch's begins open in the parent's session: its change, project key, session
    token and stage. An open run wins over a later one already ended, so a pipeline-fix
    run that began and ended inside an open sdd-tdd run does not name the session's change from
    then on (kan-1099 self-review, F12: row 139 was never begun)."""
    addr = os.environ.get("FLOW_ADDR") or "http://127.0.0.1:4173"
    # ponytail: an open run older than the 50 latest is missed; filter ended_at server-side if that bites.
    query = parse.urlencode({"session_id": session_id, "sort": "-started_at", "limit": "50"})
    with request.urlopen(f"{addr}/api/v1/stage-runs?{query}", timeout=bound(2)) as resp:
        runs = json.load(resp).get("stageRuns")
    runs = [r for r in runs if isinstance(r, dict) and not str(r.get("changeName") or "").startswith(LOOP_KEY_PREFIXES)] if isinstance(runs, list) else []
    run = next((r for r in runs if not r.get("endedAt")), runs[0] if runs else {})
    values = (run.get("changeName"), run.get("projectKey"), run.get("sessionToken"), run.get("stage"))
    name, key, token, stage = (v if isinstance(v, str) and v else None for v in values)
    return name, key, token, stage


def begin_key(tool_input: dict, prompt: str) -> "str | None":
    """The row key a dispatch is begun under: the prompt's column-0 `Key:` line of any slug, else the
    description's first word, when it is a key the hook begins."""
    line = KEY_LINE.search(prompt)
    if line and AD_HOC_KEY.fullmatch(line.group(1)):
        return line.group(1)
    description = tool_input.get("description")
    words = description.split() if isinstance(description, str) else []
    return words[0] if words and BEGIN_KEY.fullmatch(words[0]) else None


def names_worktree(prompt: str, worktree: str) -> bool:
    return re.search(re.escape(worktree) + r"(?:/|\s|$)", prompt) is not None


def begin_row(payload: dict, tool_input: dict, prompt: str) -> "list | str | None":
    """The begin argv to write once every deny check has passed, a deny reason (str) when a judged
    role's dispatch has no saved pair-judge pick, or
    None: a row already open, a dispatch with no key or no mapped role, or an unanswered store.
    The hook writes the begin; the row's end stays the parent's."""
    if LOOP_KEY_LINE.search(prompt) or loop_word(tool_input):
        return None
    planning = tool_input.get("subagent_type") in PLANNING_AGENT_TYPES
    key = begin_key(tool_input, prompt)
    role_line = ROLE_LINE.search(prompt)
    role_id = role_line.group(1) if role_line else None
    role = ROW_ROLE.get(role_id) if role_id else None
    session_id, cwd = payload.get("session_id"), payload.get("cwd")
    if key is None or role is None or not isinstance(session_id, str) or not session_id or not isinstance(cwd, str):
        return None
    if not in_flow_run(payload):
        return None
    try:
        change, project, token, stage = session_change(session_id)
        if not change:
            return None
        # The rows are read under cwd's project: another repo's cwd reads none, so fail open.
        main = main_checkout(cwd)
        if project and project_key(cwd, main) != project:
            return None
        worktree = f"{main}-worktrees/{change}" if main else None
        if not token or not worktree:
            return None
        if planning and not (stage == "flow.plan" and names_worktree(prompt, worktree)):
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
    if any(isinstance(r, dict) and not r.get("outcome") and r.get("key") == key for r in rows):
        return None
    if planning:
        effort = tool_input.get("effort")
    else:
        typed = FLOW_EFFORT.fullmatch(str(tool_input.get("subagent_type") or ""))
        effort = typed.group(1) if typed else None
    args = ["flow", "record", "dispatch", "begin", "-C", cwd, "-change", change, "-role", role,
            "-model", str(tool_input.get("model") or ""), "-effort", str(effort or "")]
    single = SINGLE_TASK_KEY.match(key)
    if single:
        args += ["-task", single.group(1)]
    args += ["-key", key, "-session-token", token, "-policy-role", role_id]
    pick = read_pick(os.path.join(worktree, ".superpowers", "sdd", "pair-judge", key + ".pick"))
    if pick:
        args += ["-pair-pick", pick]
    elif role_id in JUDGED_ROLES and os.environ.get("FLOW_HOOK_ROLE_CHECK") != "off":
        return (f"Dispatch blocked: `{role_id}` dispatches take the pair judge's pick, and "
                f"`.superpowers/sdd/pair-judge/{key}.pick` is not there.\n\n"
                f"Write the dispatch's prompt to `{worktree}/.superpowers/sdd/pair-judge/{key}.md`, run\n"
                f"flow-guard pair-judge {role_id} {worktree}/.superpowers/sdd/pair-judge/{key}.md\n"
                "(it never blocks: a judge that cannot answer prints the role's first pair), then send the "
                "dispatch on the pair it prints (`skills/flow/brainstorm-planner.md`, **Model and effort**).")
    return args


def read_pick(path: str) -> "str | None":
    """The `<source>: <reason>` of the `<model> <effort> <source> <reason…>` line flow-guard
    pair-judge saved, or None when the file is absent or its first line does not parse."""
    try:
        with open(path, encoding="utf-8") as f:
            fields = f.readline().split(maxsplit=3)
    except (OSError, ValueError):
        return None
    return f"{fields[2]}: {fields[3].strip()}" if len(fields) == 4 else None


def write_begin(args: list) -> None:
    """Write the row's begin. Best-effort: no failure denies the dispatch."""
    try:
        subprocess.run(args, capture_output=True, text=True, timeout=bound(10), check=False)
    except (OSError, subprocess.SubprocessError):
        pass


def change_finished(payload: dict) -> bool:
    """Is the session's change already FINISHED in the store? Then the session runs no flow
    phase any more, and the row checks have no row to hold a dispatch to. Unanswered: False."""
    session_id, cwd = payload.get("session_id"), payload.get("cwd")
    if not isinstance(session_id, str) or not session_id:
        return False
    try:
        change, key, _, _ = session_change(session_id)
        key = key or (project_key(cwd) if isinstance(cwd, str) else None)
        if not change or not key:
            return False
        addr = os.environ.get("FLOW_ADDR") or "http://127.0.0.1:4173"
        url = f"{addr}/api/v1/changes/{parse.quote(key)}/{parse.quote(change)}"
        with request.urlopen(url, timeout=bound(2)) as resp:
            return json.load(resp).get("state") == "FINISHED"
    except (OSError, ValueError, subprocess.SubprocessError, AttributeError):
        return False


_FLOW_RUN: "dict[int, bool]" = {}


def in_flow_run(payload: dict) -> bool:
    """A flow session whose change is not yet FINISHED -- the scope of the row checks. Answered
    once per payload: every check asks, and each answer reads the transcript and the store."""
    if id(payload) not in _FLOW_RUN:
        _FLOW_RUN[id(payload)] = in_flow_session(payload) and not change_finished(payload)
    return _FLOW_RUN[id(payload)]


def in_flow_session(payload: dict) -> bool:
    path = payload.get("transcript_path")
    if not isinstance(path, str):
        return False
    try:
        with open(path, encoding="utf-8", errors="replace") as f:
            return any(is_flow_command(line) for line in f)
    except OSError:
        return False


def role_pair_reason(payload: dict, tool_input: dict, prompt: str) -> "str | None":
    """The deny reason for a `flow-*` dispatch in a flow session whose Role line or pair the policy
    does not admit, else None. Any failure to read the policy admits the dispatch."""
    agent = tool_input.get("subagent_type")
    if os.environ.get("FLOW_HOOK_ROLE_CHECK") == "off":  # the hooks' own harness, for cases not about the check
        return None
    if not isinstance(agent, str) or not agent.startswith("flow-") or not in_flow_session(payload):
        return None
    try:
        done = subprocess.run(["flow", "policy", "get"], capture_output=True, text=True,
                              timeout=bound(10), check=False, cwd=payload.get("cwd") or None)
        policy = json.loads(done.stdout) if done.returncode == 0 else None
    except (OSError, ValueError, subprocess.SubprocessError):
        return None
    if not isinstance(policy, dict) or not policy:
        return None
    roles = ", ".join(sorted(policy))
    line = ROLE_LINE.search(prompt)
    if line is None:
        return ("Dispatch blocked: a flow-* dispatch's prompt carries a `Role: <id>` line "
                "(skills/flow-contracts/model-policy.md, **Allowed pairs**).\n\n"
                f"Roles: {roles}.\nAn unusual task no role fits takes `general`.")
    role = line.group(1)
    entry = policy.get(role)
    if not isinstance(entry, dict):
        return (f"Dispatch blocked: `Role: {role}` is not a role of the policy; never invent one.\n\n"
                f"Roles: {roles}.\nAn unusual task no role fits takes `general`.")
    admitted = [*(entry.get("base") or []), *(entry.get("optin") or [])]
    pair = f"{tool_input.get('model') or '<no model>'}/{agent}"
    if pair in admitted:
        return None
    return (f"Dispatch blocked: `{role}` does not admit `{pair}`.\n\n"
            f"Admitted pairs: {', '.join(admitted) or 'none'}.\n"
            "Re-send with `model` and `subagent_type` set to one of them.")


def fork_reason(payload: dict, tool_input: dict) -> "str | None":
    prompt = tool_input.get("prompt")
    if isinstance(prompt, str) and "skills/flow" in prompt:
        return FORK_REASON
    return FORK_REASON if in_flow_session(payload) else None


def is_flow_command(line: str) -> bool:
    # A typed slash command is a user turn with string content; a tool result quoting the tag
    # carries list content and does not count. A run the model started is an assistant turn
    # calling the Skill tool on a flow skill (kan-1131: a /flow-fast started that way began no
    # dispatch row). The substring test spares a json parse of every line of a long transcript
    # that holds neither.
    if "<command-name>/flow" not in line and '"Skill"' not in line:
        return False
    try:
        entry = json.loads(line)
    except ValueError:
        return False
    if not isinstance(entry, dict):
        return False
    message = entry.get("message")
    content = message.get("content") if isinstance(message, dict) else None
    if entry.get("type") == "user":
        return isinstance(content, str) and bool(FLOW_COMMAND.search(content))
    if entry.get("type") != "assistant" or not isinstance(content, list):
        return False
    return any(
        isinstance(b, dict)
        and b.get("type") == "tool_use"
        and b.get("name") == "Skill"
        and isinstance(b.get("input"), dict)
        and b["input"].get("skill") in FLOW_SKILLS
        for b in content
    )


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
    sys.stdout.flush()
    cwd = PAYLOAD.get("cwd")
    change = None
    session_id = PAYLOAD.get("session_id")
    if isinstance(session_id, str) and session_id:
        try:
            change, _, _, _ = session_change(session_id)
        except (OSError, ValueError, AttributeError):  # recorded without its change
            change = None
    record_incident("enforce-agent-baseline", reason, cwd if isinstance(cwd, str) else None, change)


def record_incident(guard, reason, cwd, change=None):
    """One `flow record incident` row per deny, so a self-review pass can count the denials;
    FLOW_HOOK_INCIDENT_CMD overrides `flow` for the hook's own harness only. Output is captured
    and every failure swallowed: the deny already written stands either way."""
    lines = reason.strip().splitlines()
    args = [os.environ.get("FLOW_HOOK_INCIDENT_CMD") or "flow", "record", "incident",
            "-guard", guard, "-symptom", (lines[0] if lines else reason)[:300],
            "-recovery", "re-sent per the deny reason", "-minutes-lost", "0"]
    if cwd:
        args += ["-C", cwd]
    if change:
        args += ["-change", change]
    try:
        subprocess.run(args, capture_output=True, text=True, timeout=bound(10), check=False)
    except (OSError, ValueError, subprocess.SubprocessError):
        pass


# The hook's payload, set by main for deny's incident row.
PAYLOAD: dict = {}


def main() -> int:
    global PAYLOAD
    try:
        payload = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0

    if not isinstance(payload, dict):
        return 0
    PAYLOAD = payload

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

    reason = loop_key_line_reason(payload, tool_input, prompt)
    if reason:
        deny(reason)
        return 0
    reason = role_pair_reason(payload, tool_input, prompt)
    if reason:
        deny(reason)
        return 0
    begin = begin_row(payload, tool_input, prompt)
    if isinstance(begin, str):
        deny(begin)
    elif begin:
        write_begin(begin)
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception:  # fail open, never block work on a hook fault
        sys.exit(0)
