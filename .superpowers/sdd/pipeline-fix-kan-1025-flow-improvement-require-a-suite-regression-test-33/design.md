# Design: review-panel end resolves the change's worktree

Without `-C`, `gatePanelFindingsClosed` (`stats/cmd/flow/stage.go`) passes `"."` to `check-panel-findings-closed`; the /flow parent's cwd is the main checkout, so the guard reads the main checkout's `.superpowers/sdd/` and refuses (kan-1025). Fix: when `-C` is absent, resolve the dir with the existing `changeCheckouts` (`stats/cmd/flow/record.go`) — the first non-main checkout carrying the change, else `"."`. The prescribed `-C`-less line in the docs then becomes correct and stays as written.
