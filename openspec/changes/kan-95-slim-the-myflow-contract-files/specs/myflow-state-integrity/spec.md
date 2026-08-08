## REMOVED Requirements

### Requirement: A self-heal rewrite carries forward every field it did not infer

**Reason**: State self-heal is removed from the pipeline. `/myflow-status` was the only command that
performed it, and it no longer validates a state file against artifacts — it reads the state file
plus local git and reports what it finds. A requirement governing a rewrite that nothing performs
describes behaviour no command has.

**Migration**: The carry-forward duty this requirement restated at the self-heal path is unaffected
and remains in force for every state-file write, under **State file**
(`skills/myflow-contracts/state-file.md`): a write renders the whole object, so every command reads
the existing file first and re-emits every field it does not own. Nothing is lost by the removal —
what goes away is the second statement of that duty for a write path that no longer exists.

### Requirement: A field that cannot be recovered is announced, never silently nulled

**Reason**: The announcement this requirement governs is made by a self-heal rewrite, and there is no
longer any such rewrite. No command infers a state file's contents, so no command has a field it
could fail to recover.

**Migration**: A state file that is missing or unparseable is now **reported and skipped**, never
rebuilt from inference — the rule already stated under **Change name resolution**
(`skills/myflow-contracts/pipeline.md`), which requires an unreadable state-directory file to be
named in the resolution's own output rather than folded into a "no change" result.

**The cost of this removal is stated rather than left implicit.** A state file that goes stale, that
is contradicted by the artifacts on disk, or that records a `prUrl` for a pull request that no longer
exists is now never corrected by anything, and never flagged with a `⚠`. This was raised during
planning and the decision was to delete the mechanism rather than move it to another command.
