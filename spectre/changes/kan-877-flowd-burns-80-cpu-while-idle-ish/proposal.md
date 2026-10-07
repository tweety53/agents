# kan-877-flowd-burns-80-cpu-while-idle-ish

## Why

`flowd` was measured at ~80 % CPU while the pipeline was nearly idle (KAN-877). The harvest watcher's
retried-give-up scan re-reads and re-parses every transcript on disk on every 5 s cycle while any token
seeded from a persisted give-up stays pending, and matches each command against every such token. On this
machine one such cycle costs about 90 s of CPU, cycles run back to back, and the scan stays live for the
whole bounded window — so every restart that finds persisted give-ups pins a core for well over an hour,
looking for tokens whose transcripts have already been pruned. Separately, even an idle cycle makes one
store query and one file open per transcript on disk.

## What changes

- The retried-give-up scan reads the transcripts once per process instead of on every cycle.
- Matching a batch of commands against the pending session tokens costs one pass over the commands,
  whatever the number of pending tokens.
- A transcript whose size has not changed since the watcher last saw it committed is skipped without a
  store query or a read.
- No change to what binds, what gives up, or what is attributed.
