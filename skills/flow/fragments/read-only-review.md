**READ-ONLY REVIEW:** Do not mutate the working tree, index, HEAD, or branch. Inspect with
Read, Grep, and git show/diff only. No command that writes: no `git checkout`, `git restore`,
`git reset`, `git stash`, `git clean`, `git worktree add`, no commit, no index change, and no
file edit outside your own report file — the panel's mutating slots are the one declared
exception, and they work in throwaway copies the parent creates, not this tree.
