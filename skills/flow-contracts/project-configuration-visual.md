# Project configuration — visual verification

The `## visual verification` section of **Project configuration** (`skills/flow-contracts/project-configuration.md`).

## visual verification

**How a `## visual verification` section is written.** One settings table, then one commands table,
and nothing else in the section is read — prose beside them is for the reader, exactly as
`## workspace isolation` above. **Declare the section at most once**: a file carrying the
`## visual verification` heading twice is an ambiguous declaration, not a merge, and neither table
is read.

The settings table's header folds to `setting|value` before it is compared — case and internal
whitespace do not matter, so `| setting | VALUE |` matches as readily as `| Setting | Value |`, but
the two columns and their order do. **Its `Setting` vocabulary is closed**: a name outside the six
rows below is reported and its row dropped, never silently ignored.

| Setting | Required | Meaning |
|---------|----------|---------|
| `ui paths` | yes | Comma-separated globs, relative to each app root in `## apps`. A run whose diff matches none of them skips the `flow.visual-verify` stage. |
| `screenshots` | yes | The root beneath which `verify` and `capture` write PNGs, **searched recursively rather than joined with a filename** — relative to the regression checkout when one is declared, otherwise to the project root. A per-change `capture` spec lands its PNGs in its own nested snapshot directory, a different one for every spec, so no single leaf path is correct for both the baseline suite and an arbitrary capture; naming the root instead is. |
| `regression checkout` | no | Absolute path to the repository the per-change spec and its PNGs are committed to. **Absent means they are committed to the change's own branch instead.** When this same repository is also declared in `## apps`, its root follows that table's own rule — **Roots in `## apps` are main checkouts** above — resolving to the change's apply worktree there rather than the main checkout; a `regression checkout` with no matching `## apps` row has no worktree to resolve and is always the main checkout. Also the root `<agents repo>/scripts/check-spec-reach.sh` enumerates `*.spec.ts` under and diffs against what the checkout's `package.json` scripts list — its header is canonical for that gate and its exit codes. |
| `regression repo` | only with `regression checkout` | The remote URL that checkout's real `origin` must equal — **an identity assertion, never an authorisation.** Nothing pushes automatically; see below. |
| `mockups` | no | Directory holding one `<frame id>.png` per drawn frame, relative to the project root. Declared, it enables the compose step of `flow.visual-verify` for any capture spec carrying a `<spec>.mockups` sidecar; absent, the stage composes nothing and reports `mockups: not declared`. |
| `mockup frame` | no | `scale=<int> status=<px> border=<px>`: the mockup PNGs' scale factor, the status-line height and the border width, the last two in logical px. Declared, the compose step crops each frame to its content area — the sides and the top from these values, the bottom found per frame as the first row below the status line every one of whose pixels between the borders equals the page colour at (0, 0) — and reports a capture whose size differs from the cropped frame as a finding rather than composing it. Absent, frames compose at native size, uncropped, with no size check. |

The commands table's header folds the same way, to `command|runs` — the same heading
`## workspace isolation` above already establishes for a table of project-declared commands, reused
here rather than invented a second time. **Its `Command` vocabulary is closed** too: a name outside
the six rows below is reported and its row dropped.

| Command | Required | Meaning |
|---------|----------|---------|
| `setup` | no | Run once before `verify` when the toolchain is missing. |
| `verify` | yes | Runs the checked-in baseline suite. A `<specs>` in it is where `flow.visual-verify` puts the spec paths `specs` printed, space-separated — or nothing, when there are none — so the command must run its whole suite when `<specs>` is empty. |
| `capture` | yes | Runs the per-change spec and **creates** this change's baseline — it is expected to write PNGs that do not yet exist, and that is success, not failure. `verify` above is the regression gate, guarding an already-committed baseline; `capture` is not, and the project's own command must be the variant that succeeds on a first-run write rather than the one built to fail a comparison against nothing. |
| `fingerprint` | no | Exits 0 when the app the stage's probe answered from is serving the worktree's own build, non-zero otherwise. What a bundle's identity is differs per stack, so the project owns the comparison; `flow.visual-verify` owns what a non-zero exit does (one restart from `## run`, then block). Absent means the stage reports `fingerprint: not declared` and proves nothing about the served bundle — a supported state for a project with no served bundle, not a misconfiguration. |
| `start` | no | Starts the stack for this stage's own probe, from the worktree; absent means `## run`. This is where a project moves a port the operator's `## run` keeps fixed. |
| `specs` | no | Prints the specs covering the change's touched ui paths, one per line, run from the regression checkout with `<frontend-root>` and `<merge-base>` substituted. Its output reaches `verify` through `verify`'s `<specs>`. Absent means `verify` and `capture` run as declared. |

**The `mockups` sidecar.** A capture spec finds its frames through a `<spec>.mockups` file beside
it — one `<screenshot name> <frame id>` line per pair, `#` comments and blank lines ignored. A
`<screenshot name>` is the name the spec passed to `toHaveScreenshot()`; Playwright's own
`-<platform>` suffix is tolerated on match.

With `mockup frame` declared, the composite `flow.visual-verify` writes is three panels — the
captured frame, the cropped mockup, and a difference panel white wherever any channel differs —
and the stage prints one `<composite path> diff=<ratio>` line per pair, the ratio being the share
of differing pixels. **That ratio is a number the verifier reads beside the panels, never a
threshold**: nothing blocks on it, because an antialiasing difference and a missing control
produce comparable ratios. Without the row the composite is the captured frame on the left and the
mockup on the right, both at native size, and its ratio field carries this literal instead:

```text
diff=n/a
```
 `<agents
repo>/scripts/compose-mockup-frames.sh`'s own header is canonical for its exit codes, the way this
section already cites `check-visual-verification.sh` for its shape.

**The full app suite.** One spec, `full-app-suite.spec.ts`, capturing every screen the app has —
the whole-app regression suite, beside the per-change specs and reached by `verify` like them.
`flow.visual-verify` updates it on every run and creates it where the checkout has none, then
writes every PNG the suite produced into one archive, `full-app-suite.zip`, directly under the
`screenshots` root, rebuilt from scratch each time. Both names are fixed rather than declared: no
row of the settings table names them, and the spec, its PNGs and the zip are committed wherever
the per-change spec is.

**No push to a `regression checkout` is ever automatic** — the change branch's own pushes are **Branch backup** (`skills/flow-contracts/git-boundaries.md`), inside the one repository the run owns. `flow.visual-verify` commits the per-change spec and its PNGs to the
`regression checkout` when one is declared and stops there; the handoff prints the push command for
the operator to run by hand. `regression repo` records which repository the checkout is expected to
be, and a mismatch against its real `origin` is reported — but that is an identity assertion, not an
authorisation, because `<project>/.flow/project.md` is tracked and editable in any pull request, and
a file inside a repository cannot authorise a push to another repository.

**Mechanically enforced** by `<agents repo>/scripts/check-visual-verification.sh`, canonical for the
section's exact shape and every violation it reports.
