#!/usr/bin/env python3
"""Claude Code hook shim: runs `flow-guard hook <this file's name>` on the event payload, through
the scripts/flow-guard of the checkout this file (after symlinks) lives in. The hook itself is Go;
its contract is the leading comment of stats/internal/guard/hook*.go that registers this name.

Exits 0 whatever happens -- a hook fault, or a flow-guard that cannot build, never blocks work.
"""

import os
import subprocess
import sys

HERE = os.path.realpath(__file__)

try:
    subprocess.run(
        [os.path.join(os.path.dirname(os.path.dirname(HERE)), "scripts", "flow-guard"),
         "hook", os.path.splitext(os.path.basename(HERE))[0]],
        check=False,
    )
except Exception:  # fail open, never block work on a hook fault
    pass
sys.exit(0)
