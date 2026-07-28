# Manual test template

Copy this structure into `docs/manual-test/<change-name>.md`. Replace every `<…>` placeholder with **real absolute paths** for the apply worktrees (or main checkouts if no worktree), and with the **real commands, URLs, and credentials for this project**.

Those project specifics come from `<main checkout>/.myflow/project.md` at generation time — `## apps`, `## run`, and `## credentials` — or from auto-detection when the project has no such file. See **Project configuration** in `skills/myflow-contracts/project-configuration.md`. This template deliberately names no build tool, task, port, URL, or account: anything of that kind appearing below as a placeholder must be filled from the project, never from another project.

**Hard rule:** every shell command runs from the involved app's current apply worktree, as an absolute path. Never a relative sibling `cd ../<other-app>`.

```markdown
# Manual test — <change-name>

**Purpose:** <one sentence from proposal>
**Branch / worktree:** `<branch>` · <app name> `<absolute root>` · <app name> `<absolute root>`
**(One segment per involved app; omit apps that are out of scope.)**
**Next after sign-off:** `/myflow-review <change-name>`
<!-- Skip mode only: add a line here — **Manual test status:** SKIPPED — YYYY-MM-DD (Gate C intentionally bypassed) -->
<!-- and leave every checklist / sign-off box below unchecked. -->

## How to run (involved apps)

### Prerequisites

- <Runtime / toolchain / service prerequisites for this project>
- Apply worktrees listed in the header (do **not** run from the main-branch checkouts for this change)

### Start local stack

<!-- The start sequence from the project's `## run`, rooted at the app that owns it. -->
<!-- When a command takes another app's root as a parameter, pass that app's ABSOLUTE worktree root. -->

```bash
cd <absolute root of the app that owns the run commands>

<start command(s) from the project's `## run`>
<seed / fixture command, only when the checklist needs seeded data>
```

Useful:

```bash
cd <same absolute root>
<status / logs / stop commands from the project's `## run`, if it defines them>
```

### Apps in scope for this change

<!-- One row per involved app, from the project's `## apps`. Delete rows that are out of scope. -->

| App | URL | Notes |
|-----|-----|-------|
| <app name> | <local URL from `## apps`> | <how it is started; anything that surprises> |

**Sign-in credentials:** <from the project's `## credentials`, and only when the checklist needs a signed-in session. State which app each account is for and what seeds it. Omit this line entirely otherwise — never invent an account.>

### Optional: run only touched surfaces

<!-- Include only if helpful; otherwise delete this subsection. Always absolute cds. -->

**<App name> only:**

```bash
cd <absolute root for that app>
<that app's single-service command(s) from the project's `## run`>
```

## Functionality checklist

<!-- One concrete manual observation per item. Group by capability/screen. -->

### <Capability or screen 1>

- [ ] <User-visible behavior from spec>
- [ ] <Edge / empty / error case if specified>

### <Capability or screen 2>

- [ ] …

## Sign-off

- [ ] All checklist items above verified on a running local stack
- [ ] No blocking bugs found (or filed / fixed via `/myflow-do-fix`)
- [ ] Ready for `/myflow-review <change-name>`
```
