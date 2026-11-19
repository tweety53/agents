#!/usr/bin/env python3
"""PreToolUse hook: block subagent dispatches that omit the agent-baseline pointer.

Rules in ~/.claude/CLAUDE.md are injected into a session's system prompt and are NOT
inherited by dispatched agents, nor by agents those agents dispatch. The prompt is the
only channel, so every dispatch must carry a pointer to ~/.claude/rules/agent-baseline.md
plus the sentence that makes the receiving agent pass it on in turn.

The PreToolUse contract has no way to rewrite tool input, so this hook cannot inject the
line itself. It denies the call instead and returns the exact text to add; the model then
re-dispatches with it.

It denies, in a flow run, a dispatch off the **Allowed pairs** table of
skills/flow-contracts/model-policy.md (KAN-1045): flow-guard check-dispatch-pair reads the
prompt's `Role: <role id>` line and refuses no Role line, an unknown role, a pair off the
role's rows, or a marked row (`No-progress escalation`, `Operator-approved high`) whose
marker is not a line of the prompt -- and a guard that cannot answer (exit 2) denies too. A
loop dispatch, whose `Key:` line hooks/enforce-loop-dispatch.py checks against the same
table, is exempt. On ZCode the Role line must still name a role of the table, and the pair is
the harness mapping's alone.

It denies every fork in a flow session -- one whose transcript records a /flow, /flow-fast,
/flow-interactive, /flow-plan or /flow-self-review command, or whose prompt names a flow skill: a
flow dispatch is a fresh agent whose prompt carries the paths it needs (KAN-1000: a parent forked
itself into a self-review loop and dragged ~300k tokens of its own context into it).

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

It begins, in a flow session, a dispatch's ledger row when none is open (KAN-1115, decision
hook-same-file): the session's change and session token are read off the store as its latest open
stage run, else its latest, of any command, skipping a loop's own runs -- a change named
`pipeline-fix-...` or `self-review-...`, whose stage runs check-loop-dispatch's begins open in the
parent's session; `flow record dispatches` lists its rows, and when none is open (empty
`outcome`) under the dispatch's key the hook runs `flow record dispatch begin` itself -- the key
the prompt's column-0 `Key:` line, else the description's first word; the role the prompt's
`Role:` line mapped back as launchRoleID (stats/cmd/flow/record.go) maps it; the call's model and
effort; `-pair-pick` from the `.pick` line flow-guard pair-judge saved as
<worktree>/.superpowers/sdd/pair-judge/<key>.pick. A gated reviewer's plan tree is snapshotted to
<worktree>/.superpowers/sdd/plan-snapshot-<key>.txt first (decision plan-check-rides-row). The
begin is written only after every deny check has passed, so a launch the hook denies leaves no row;
a begin `flow` refuses denies the launch with its stderr, and a failed snapshot denies it with no
begin. The row's end stays the parent's (decision begin-moves-end-stays).

It denies a dispatch with no open row whose row it cannot begin: a `Task <ids>/<n>` description
with no `Key:` line and no open `task-<ids>-` row, or a prompt whose `Role:` line is missing or
names no role a row records. A loop-key dispatch -- a `Key:` line, or a description opening with a
loop key -- is never checked. An unreachable store, no stage begin for the session, a stage run
with no session token, an absent or failing `flow` CLI, or an unparseable answer lets the dispatch
through with no begin.

None of those row checks applies once the session's change is FINISHED in the store: the run is
over, and a later dispatch in the same session is no flow dispatch.

A planning agent (PLANNING_AGENT_TYPES: Explore, Plan) is begun under its `explore-<n>` or
`design-<n>` key only while the session's latest stage run is `flow.brainstorm` and its prompt
names the session change's worktree, <main-checkout>-worktrees/<change> (decisions
planning-agents-hook-begun, planning-agent-begins-in-brainstorm); a loop planner's agents -- naming
the loop's worktree, or the run's own outside `flow.brainstorm` -- keep the planner's explicit begin. Otherwise no row check applies to it;
the baseline-pointer and allowed-pairs checks still do.

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

# A planner's own read-only agents pass the description and loop key line checks, and are begun
# only when their prompt names the session change's worktree (begin_row); the baseline-pointer and
# allowed-pairs checks still apply (kan-1032's self-review: a planner's Explore dispatch was denied
# mid-plan).
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

# Harness mapping (skills/flow-contracts/model-policy.md): on zcode every pair is dispatched as
# general-purpose / glm-5.3-flash -- ZCode's Agent tool offers no flow-<effort> type (KAN-609).
# The zcode registration runs the hook from ~/.zcode/hooks, which is how the hook tells the
# harness.
ZCODE_PAIR = ("general-purpose", "glm-5.3-flash")

REPO = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))
# The checker the hook runs; overridable for the hook's own harness only.
PAIR_CHECK = [os.environ["DISPATCH_PAIR_CHECK"]] if os.environ.get("DISPATCH_PAIR_CHECK") else [
    os.path.join(REPO, "scripts", "flow-guard"), "check-dispatch-pair"
]
PLAN_SNAPSHOT = [os.environ["DISPATCH_PLAN_SNAPSHOT"]] if os.environ.get("DISPATCH_PLAN_SNAPSHOT") else [
    os.path.join(REPO, "scripts", "flow-guard"), "check-plan-unchanged"
]


# A description naming its row: plan tasks by id over the plan's count, any other row by a first
# word of a Key shape the closed list names (**Dispatch sites -- the parent's closed list**,
# skills/flow/implement.md) -- `panel-4-primary`, `visual-verify-fix-11-gymie`, `full-suite-fix-2`, a planner's `plan-2`, the plan reviewer's `plan-review-1`, a /flow-interactive request's operator-asked `operator-58`, the `operator-<n>` shape admitted in a
# /flow-interactive session only (OPERATOR_KEY). A lowercase word off
# the list -- `verify-fix-1`, `verify` -- names no row (kan-1029: both launched with no row).
ROW_KEY = (
    r"task-\d+(?:\+\d+)*-(?:implementer|reviewer)(?:-[a-z0-9-]+)?"
    r"|panel-\d+-[a-z][a-z0-9+-]*"
    r"|panel-fix-\d+(?:-\d+)?(?:-retry)?"
    r"|full-suite-fix-\d+"
    r"|visual-verify(?:-[a-z0-9][a-z0-9-]*)?"
    r"|tooling-analysis-\d+(?:-[a-z0-9][a-z0-9-]*)?"
    r"|fix-run-judge-\d+"
    r"|pipeline-fix-[a-z0-9][a-z0-9-]*"
    r"|plan-review-\d+"
    r"|plan-\d+"
    r"|operator-\d+(?:-[a-z0-9][a-z0-9-]*)?"
)
ROW_DESCRIPTION = re.compile(r"(?:Tasks? \d+(?:\+\d+)*/\d+|" + ROW_KEY + r")(?:\s|$)")
OPERATOR_KEY = re.compile(r"operator-\d+(?:-[a-z0-9][a-z0-9-]*)?(?:\s|$)")
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

OPERATOR_KEY_REASON = (
    "Dispatch blocked: an `operator-<k>` key names an operator-asked dispatch of /flow-interactive "
    "(skills/flow-interactive/SKILL.md), and this session's transcript records no /flow-interactive. "
    "A step a /flow contract gives the parent's own calls -- the Live check among them "
    "(skills/flow/verify-and-handoff.md) -- runs in the parent, never as a dispatch."
)


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


def description_reason(payload: dict, tool_input: dict, prompt: str) -> "str | None":
    if tool_input.get("subagent_type") in PLANNING_AGENT_TYPES:
        return None
    if LOOP_KEY_LINE.search(prompt) or loop_word(tool_input):
        return None
    description = tool_input.get("description")
    if isinstance(description, str) and ROW_DESCRIPTION.match(description):
        if OPERATOR_KEY.match(description) and in_flow_run(payload) and not in_interactive_session(payload):
            return OPERATOR_KEY_REASON
        return None
    return DESCRIPTION_REASON if in_flow_run(payload) else None


TASK_DESCRIPTION = re.compile(r"Tasks? (\d+(?:\+\d+)*)/\d+")
# A key the hook begins a row under: a closed-list key, or a planner's Explore/Plan agent's.
BEGIN_KEY = re.compile(r"(?:" + ROW_KEY + r"|explore-\d+|design-\d+)")
KEY_LINE = re.compile(r"^Key:[ \t]*(\S+)[ \t]*$", re.M)
ROLE_LINE = re.compile(r"^Role:[ \t]*(\S+)[ \t]*$", re.M)
SINGLE_TASK_KEY = re.compile(r"task-(\d+)-")
# A gated reviewer's key, suffixed or not: its plan tree is snapshotted before the begin.
GATED_REVIEWER_KEY = re.compile(r"task-\d+(?:\+\d+)*-reviewer")
# A review-panel reviewer key, `panel-<round>-<slots>`: its row needs the slot panel-round-begins passes.
PANEL_REVIEWER_KEY = re.compile(r"panel-(\d+)-(.+?)(?:-part\d+|-unread)*")
FLOW_EFFORT = re.compile(r"flow-(?:planner-)?([a-z]+)")
# A prompt's `Role:` id -> the row role `flow record dispatch begin -role` takes: the inverse of
# launchRoleID (stats/cmd/flow/record.go).
ROW_ROLE = {
    "implementer": "implementer", "operator-asked": "implementer",
    "gated-reviewer": "reviewer", "gated-re-reviewer": "reviewer", "design-reviewer": "reviewer",
    "panel-reviewer": "reviewer", "panel-sanity-efficiency": "reviewer",
    "panel-mutation-failure": "reviewer", "panel-rerun": "reviewer", "late-fix": "reviewer",
    "panel-merger": "merger",
    "panel-fix": "panel-fix",
    "visual-verifier": "verifier",
    "planner": "planner", "plan-designer": "planner", "tooling-analyst": "planner", "fix-run-judge": "planner",
    "explorer": "explorer",
}


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
    """The row key a dispatch is begun under: the prompt's column-0 `Key:` line, else the
    description's first word, when it is a key the hook begins."""
    line = KEY_LINE.search(prompt)
    if line and BEGIN_KEY.fullmatch(line.group(1)):
        return line.group(1)
    description = tool_input.get("description")
    words = description.split() if isinstance(description, str) else []
    return words[0] if words and BEGIN_KEY.fullmatch(words[0]) else None


def names_worktree(prompt: str, worktree: str) -> bool:
    return re.search(re.escape(worktree) + r"(?:/|\s|$)", prompt) is not None


def begin_row(payload: dict, tool_input: dict, prompt: str) -> "tuple[str | None, list | None]":
    """The begun-row check: (a deny reason, None), (None, the begin to write once every other check
    has passed), or (None, None) -- a row already open, a dispatch the hook neither checks nor
    begins, or an unanswered store. The hook writes the begin; the row's end stays the parent's.
    A review-panel reviewer key (`panel-<round>-<slots>`) is denied with a pointer at
    `flow-guard panel-round-begins`: its row needs a -slot an Agent call cannot pass, so a begin
    from here is one `flow` refuses (kan-1090 self-review, F4)."""
    if LOOP_KEY_LINE.search(prompt) or loop_word(tool_input):
        return None, None
    planning = tool_input.get("subagent_type") in PLANNING_AGENT_TYPES
    description = tool_input.get("description")
    if not isinstance(description, str):
        return None, None
    key = begin_key(tool_input, prompt)
    task = TASK_DESCRIPTION.match(description)
    if key is None and (planning or not task):
        return None, None
    session_id, cwd = payload.get("session_id"), payload.get("cwd")
    if not isinstance(session_id, str) or not session_id or not isinstance(cwd, str):
        return None, None
    if not in_flow_run(payload):
        return None, None
    try:
        change, project, token, stage = session_change(session_id)
        if not change:
            return None, None
        # The rows are read under cwd's project: another repo's cwd reads none, so fail open.
        main = main_checkout(cwd)
        if project and project_key(cwd, main) != project:
            return None, None
        worktree = f"{main}-worktrees/{change}" if main else None
        if planning and not (stage == "flow.brainstorm" and worktree and names_worktree(prompt, worktree)):
            return None, None
        done = subprocess.run(
            ["flow", "record", "dispatches", "-change", change, "-C", cwd],
            capture_output=True, text=True, timeout=bound(5), check=False,
        )
        rows = json.loads(done.stdout) if done.returncode == 0 else None
    except (OSError, ValueError, subprocess.SubprocessError, AttributeError):
        return None, None
    if not isinstance(rows, list):
        return None, None
    open_keys = [r["key"] for r in rows if isinstance(r, dict) and not r.get("outcome") and isinstance(r.get("key"), str)]
    if key in open_keys if key else any(k.startswith(f"task-{task.group(1)}-") for k in open_keys):
        return None, None
    role = ROLE_LINE.search(prompt)
    role = ROW_ROLE.get(role.group(1)) if role else None
    if key is None or role is None:
        word = key or f"task-{task.group(1)}-"
        shown = key or f"task-{task.group(1)}-implementer"
        return (
            f"Dispatch blocked: no open ledger row for '{word}' in {change} — record it first with "
            f"`flow record dispatch begin -change {change} -key {shown} …` (skills/flow/implement.md), "
            "then re-send this call unchanged."
        ), None
    if not token or not worktree:
        return None, None
    if planning:
        effort = tool_input.get("effort")
    else:
        typed = FLOW_EFFORT.fullmatch(str(tool_input.get("subagent_type") or ""))
        effort = typed.group(1) if typed else None
    model = str(tool_input.get("model") or "")
    panel = PANEL_REVIEWER_KEY.fullmatch(key) if role == "reviewer" else None
    if panel:
        # A -part<k> or -unread suffix is written <slots>@<suffix> (kan-1132 self-review, F1).
        suffix = key[panel.end(2):].removeprefix("-")
        slot = panel.group(2) + (f"@{suffix}" if suffix else "")
        if "-" in panel.group(2):
            # An unsplit -<repo> suffix: never hand back a <slot>-<repo> slot (kan-1132 self-review).
            slot = "<slots>@<repo>"
        return (
            f"Dispatch blocked: no open ledger row for '{key}' in {change} — a review-panel reviewer "
            "row needs its slot, which an Agent call cannot pass. Record the round's begins first with "
            f"`flow-guard panel-round-begins {worktree} {change} {panel.group(1)} {token} "
            f"{slot}:{model}:{effort or ''}` (skills/flow/review-panel.md); a key carrying a -<repo> "
            "suffix is begun the same way, its slot written <slots>@<repo>; "
            "then re-send this call unchanged."
        ), None
    args = ["flow", "record", "dispatch", "begin", "-C", cwd, "-change", change, "-role", role,
            "-model", model, "-effort", str(effort or "")]
    single = SINGLE_TASK_KEY.match(key)
    if single:
        args += ["-task", single.group(1)]
    args += ["-key", key, "-session-token", token]
    pick = read_pick(os.path.join(worktree, ".superpowers", "sdd", "pair-judge", key + ".pick"))
    if pick:
        args += ["-pair-pick", pick]
    snapshot = None
    if GATED_REVIEWER_KEY.match(key):
        snapshot = [*PLAN_SNAPSHOT, "snapshot", worktree, change,
                    os.path.join(worktree, ".superpowers", "sdd", f"plan-snapshot-{key}.txt")]
    return None, [snapshot, args]


def read_pick(path: str) -> "str | None":
    """The `<source>: <reason>` of the `<model> <effort> <source> <reason…>` line flow-guard
    pair-judge saved, or None when the file is absent or its first line does not parse."""
    try:
        with open(path, encoding="utf-8") as f:
            fields = f.readline().split(maxsplit=3)
    except (OSError, ValueError):
        return None
    return f"{fields[2]}: {fields[3].strip()}" if len(fields) == 4 else None


def write_begin(begin: list) -> "str | None":
    """Snapshot a gated reviewer's plan tree, then write the row's begin. A failed snapshot or a
    begin `flow` refuses is a deny reason; a command that cannot run fails open with no begin."""
    snapshot, args = begin
    try:
        if snapshot:
            os.makedirs(os.path.dirname(snapshot[-1]), exist_ok=True)
            done = subprocess.run(snapshot, capture_output=True, text=True, timeout=bound(60), check=False)
            if done.returncode != 0:
                return (
                    f"Dispatch blocked: flow-guard check-plan-unchanged snapshot exit {done.returncode}, "
                    f"so this gated reviewer's row was not begun:\n{done.stdout}{done.stderr}"
                )
        done = subprocess.run(args, capture_output=True, text=True, timeout=bound(10), check=False)
    except (OSError, subprocess.SubprocessError):
        return None
    if done.returncode != 0:
        return f"Dispatch blocked: flow refused this dispatch's begin:\n{done.stderr}"
    return None


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


def in_interactive_session(payload: dict) -> bool:
    path = payload.get("transcript_path")
    if not isinstance(path, str):
        return False
    try:
        with open(path, encoding="utf-8", errors="replace") as f:
            return any(
                is_flow_command(line)
                and ("<command-name>/flow-interactive</command-name>" in line or '"flow-interactive"' in line)
                for line in f
            )
    except OSError:
        return False


def dispatch_pair_reason(payload: dict, prompt: str, tool_input: dict) -> "str | None":
    if LOOP_KEY_LINE.search(prompt) or not in_flow_run(payload):
        return None
    typ, model = tool_input.get("subagent_type"), tool_input.get("model")
    zcode = "/.zcode/" in __file__
    if zcode:
        # The pair is the harness mapping's; the Go guard still judges the Role line.
        args = [*PAIR_CHECK, REPO, "--role-only"]
    else:
        args = [*PAIR_CHECK, REPO, str(typ or ""), str(model or "")]
        if typ in PLANNING_AGENT_TYPES:
            args.append(str(tool_input.get("effort") or ""))
    try:
        done = subprocess.run(args, input=prompt, capture_output=True, text=True, timeout=bound(240), check=False)
    except (OSError, subprocess.SubprocessError) as e:
        return f"Dispatch blocked: flow-guard check-dispatch-pair could not run ({e})."
    if done.returncode != 0:
        return (
            f"Dispatch blocked (flow-guard check-dispatch-pair exit {done.returncode}):\n{done.stdout}{done.stderr}\n"
            "Re-send this call on a pair its role's rows of Allowed pairs (skills/flow-contracts/model-policy.md) "
            "admit, its prompt carrying the `Role: <role id>` line."
        )
    if zcode and (typ, model) != ZCODE_PAIR:
        return (
            f"Dispatch blocked: on ZCode every flow dispatch runs on subagent_type: \"{ZCODE_PAIR[0]}\" "
            f"and model: \"{ZCODE_PAIR[1]}\" (Harness mapping, skills/flow-contracts/model-policy.md); "
            f"got subagent_type={typ!r}, model={model!r}."
        )
    return None


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

    begin = None
    reason = description_reason(payload, tool_input, prompt) or loop_key_line_reason(payload, tool_input, prompt)
    if not reason:
        reason, begin = begin_row(payload, tool_input, prompt)
    reason = reason or dispatch_pair_reason(payload, prompt, tool_input) or (begin and write_begin(begin))
    if reason:
        deny(reason)
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception:  # fail open, never block work on a hook fault
        sys.exit(0)
