# Brainstorm — a change that carries UI

Items of **The checklist** (`skills/flow/brainstorm-planner.md`) a planner loads only when the
change carries UI or can alter what a user sees.

- **A mockup is read from where the project declares mockups**, the `## visual verification`
  section (**Project configuration — visual verification**,
  `skills/flow-contracts/project-configuration-visual.md`), never copied into an app repository; a
  design specified by a mockup that is not there yet is not approvable until it is.
- **Shared constants are listed before any frame is approved.** For a change that carries UI, read
  the constants its frames share — bar heights, the title's y, rail colours, row heights, greys,
  label wording — from the app's own token file, and list them in `design.md` under
  `## Shared constants`, one per item: ``- `<name>`: <value> — <token and file it was read from>``.
  A frame that disagrees with the list is corrected in the frame, or the list is changed by an
  explicit decision recorded in `design.md` — never left for `flow.verify` to find. Where two
  frames the plan cites disagree with each other (copy, label case, order), `design.md` names the
  governing frame and why, or records the departure, before the plan is written (the lens
  **Design critic** of `skills/flow/plan-review-rubric.md`). Each frame of the change states every listed
  constant — the value it draws, or `n/a` where it does not draw it — in
  a sidecar `<frame id>.constants` beside it, one `<name>: <value>` per line, or in
  `data-const-<name>="<value>"` attributes in an HTML frame. A sidecar's value is read off that
  frame — measured on its PNG with `flow-guard measure-visual-properties`, or read from the design
  source at the frame's tag — never copied from the list, which is what it is checked against.
  `flow-guard check-frame-constants <design.md> <frames dir or frame file>...`, given the change's
  frames, exits 0 before the design is approved; it exits 1 naming each frame that leaves a listed
  constant unstated and each constant two frames draw differently. A rendered frame's sidecar sits
  in its tag's directory, which is never committed: the check is this gate's, re-stated if the
  frames are rendered afresh. Whoever adds or changes a listed constant runs
  `flow-guard check-frame-constants` in the same step and fixes every frame it names — each
  `FRAME-CONSTANT-ADD:` line it prints is the sidecar line to add, its value read off the frame. Every colour a cited frame uses is also mapped to a theme token, or a
  new one the design adds, before the design is approved — never a hex at the use site.
- **A computed rule is run against every cited frame.** Where a screen's rows, titles or copy come
  from a computed rule, the planner runs it for each cited frame and compares; every mismatch is
  decided before the design is approved.
- **Every visible element a task builds is drawn before the design is approved, and the
  `designer` draws what the mockups leave undrawn.** For a change that carries UI, list in
  `design.md` under `## Undrawn`, from the tasks as written, each element, state, icon, copy string,
  alignment and chrome variant (title placement, close control) that no cited frame draws, and each
  icon the project's icon set lacks — never inferred from a design template, a sibling screen or
  the nearest component. `flow-guard design-repo <abs-worktree>` — never
  `flow-guard project-get`, never the main checkout — says who draws them. Exit 0 prints
  `<clone> <remote>`: the planner leaves the items listed and returns (**The designer**,
  `skills/flow/designer-brief.md`, draws them). Exit 1, no design repo: the planner draws every item into the prototype HTML
  (`plan-visual-prototype`, below), cites it and empties the list. Exit 2 is a defect fixed in the
  run. **The answer to an undrawn item is always draw**: it is never put to the operator as a
  question, and neither ruling it out nor reusing something as is is ever offered. The design is not approvable while
  `## Undrawn` holds an item. An implementer meeting an element the plan missed reports it in its
  REPORT FILE and builds the nearest drawn pattern.
- **A visible change is prototyped before its design is approved** (`plan-visual-prototype`). When the change can alter what a user sees — a page, a view, a chart, a dialog — and
  no mockup frame already specifies it, build a clickable prototype of the proposed result with
  sample data as a single HTML file under `<project>-worktrees/<name>/.superpowers/sdd/`, and
  iterate it on the operator's feedback each round until they approve; that approval is the convergence confirm's
  **approve the design and move on** (**Convergence**, `skills/flow/brainstorm-planner.md`). The prototype is
  built on the app's own shell and styles — the real header and nav, the project's theme,
  stylesheet tokens and existing component classes, the content width and chart size of the
  sibling views — never a standalone restyle. The file draws the proposed page
  alone — no caption, note or screenshot beside it — since `flow.verify` renders the whole
  file as the capture's frame; a note to the operator goes in the message. **The operator sees the prototype
  whenever one exists**: a planner only writes the file; the session that dispatched it publishes
  it with the Artifact tool after every return that wrote or changed it and prints its URL —
  `/flow`'s parent in the relay. Where the harness has no Artifact tool, that session prints the file's absolute path
  and says it could not be published. Cite the approved file, with its Artifact URL
  or, where it could not be published, that it was not, from `design.md`'s `## Context` as the
  spec the implementation and `flow.verify` compare
  against. A change with nothing visible builds none. Where the confirm is not asked — the
  `## decisions: recommended` mode, which takes the recommendation (**Auto-resolution**, `skills/flow-contracts/operator-prompts-auto-resolution.md`)
  — the published prototype is the spec without an approval round. **The approved prototype is
  the source of truth**: any departure from it found later in the run — by a reviewer,
  verify or `flow.verify` — is fixed to match it, auto-resolved and never asked, as **A departure
from the mockup** (`skills/flow-contracts/operator-prompts-auto-resolution.md`) resolves it; a UI
  change the operator wants after using it is a later change, not this one.
