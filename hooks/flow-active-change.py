#!/usr/bin/env python3
"""UserPromptSubmit hook: name the session's active /flow or /flow-interactive change on every plain prompt.

A problem report typed as a plain message, with no /flow prefix, needs to know which
change it belongs to before the SKILL.md rule can turn it into a fix run. The change name
is not carried in session state a hook can read directly, but every /flow run marks its
own stages with `flow stage begin -command '/flow' ... -session-token mf-<token> <name>`,
and the stats daemon harvests those marks out of the session's transcript into its store,
binding each to the session that made it. This hook asks the store for this session's latest
stage-begin mark of any command, skipping a loop's own runs -- a change named `pipeline-fix-...`
or `self-review-...`, whose stage runs a loop dispatch's begins open in the parent's session -- one GET against /api/v1/stage-runs, whose rows carry
the owning change's name, the command and the stage -- and branches on its command: a
/flow mark injects the fix-run line naming the change; a /flow-interactive mark, unless
it is flow.write-finished, injects the line pointing a plain message at the session's
request protocol (so it survives a compaction); any other command injects nothing.

Fails open, like enforce-agent-baseline.py: any unexpected input, an unreachable store,
or an answer that names no change exits 0 with no output. A hook that supplies context
must never be the reason a prompt is refused. A mark younger than the daemon's harvest
cycle is not in the store yet, and the hook answers nothing in that window rather than
re-parsing the transcript itself -- a fallback would keep the per-prompt transcript read
this hook exists to avoid.
"""

import json
import os
import re
import sys
from typing import Optional
from urllib import parse, request

DEFAULT_STORE_ADDR = "http://127.0.0.1:4173"
STORE_TIMEOUT_SECONDS = 2

# A zcode rollout transcript is named model-io-sess_<id>.jsonl; a claude-code
# transcript is named after the bare session id. Either way the id survives
# as the basename minus the model-io-sess_ prefix and the .jsonl suffix.
ROLLOUT_NAME_RE = re.compile(r"^model-io-sess_(.+)\.jsonl$")

# A pipeline-fix or self-review loop's own change: its stage runs carry the parent's session id.
LOOP_CHANGE_PREFIXES = ("pipeline-fix-", "self-review-")


def resolve_session_id(payload: dict) -> Optional[str]:
    session_id = payload.get("session_id")
    if isinstance(session_id, str) and session_id:
        return session_id

    transcript_path = payload.get("transcript_path")
    if isinstance(transcript_path, str) and transcript_path:
        basename = os.path.basename(transcript_path)
        m = ROLLOUT_NAME_RE.match(basename)
        if m:
            return m.group(1)
        stem, ext = os.path.splitext(basename)
        if stem and ext == ".jsonl":
            return stem

    env_id = os.environ.get("CLAUDE_SESSION_ID")
    if env_id:
        return env_id
    return None


def first_run(runs) -> Optional[dict]:
    """The first dict row of runs that is not a loop's own (`pipeline-fix-...`, `self-review-...`);
    a row with no changeName is kept."""
    for run in runs:
        if isinstance(run, dict) and not str(run.get("changeName") or "").startswith(LOOP_CHANGE_PREFIXES):
            return run
    return None


def last_stage_run(session_id: str) -> Optional[tuple]:
    """The (change, command, stage) of the session's latest stage run, any command, skipping a
    loop's own runs."""
    addr = os.environ.get("FLOW_ADDR") or DEFAULT_STORE_ADDR
    query = parse.urlencode(
        {"session_id": session_id, "sort": "-started_at", "limit": "50"}
    )
    with request.urlopen(
        f"{addr}/api/v1/stage-runs?{query}", timeout=STORE_TIMEOUT_SECONDS
    ) as resp:
        body = json.load(resp)

    runs = body.get("stageRuns")
    if not isinstance(runs, list):
        return None
    run = first_run(runs)
    if run is None:
        return None
    name = run.get("changeName")
    if not isinstance(name, str) or not name:
        return None
    return name, run.get("command"), run.get("stage")


def main() -> int:
    payload = json.load(sys.stdin)

    if not isinstance(payload, dict):
        return 0

    prompt = payload.get("prompt")
    if isinstance(prompt, str) and prompt.startswith("/"):
        return 0

    session_id = resolve_session_id(payload)
    if not session_id:
        return 0

    last = last_stage_run(session_id)
    if not last:
        return 0
    name, command, stage = last

    if command == "/flow":
        line = (
            f"flow: this session's last /flow run marked change {name}. A plain message "
            "reporting a problem or asking for a change to it is a fix run of "
            f"{name} — **A plain message at IN_PROGRESS** (skills/flow/SKILL.md)."
        )
    elif command == "/flow-interactive" and stage != "flow.write-finished":
        line = (
            f"flow: this session runs /flow-interactive change `{name}`. A plain message "
            "is its next request — or its stop (**One request**, **Explicit stop**, "
            "`skills/flow-interactive/SKILL.md`)."
        )
    else:
        return 0
    json.dump(
        {
            "hookSpecificOutput": {
                "hookEventName": "UserPromptSubmit",
                "additionalContext": line,
            }
        },
        sys.stdout,
    )
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception:  # fail open, never block a prompt on a hook fault
        sys.exit(0)
