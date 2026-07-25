# Manual test template

Copy this structure into `docs/manual-test/<change-name>.md`. Replace every `<…>` placeholder with **real absolute paths** for the apply worktrees (or main checkouts if no worktree). Omit entire app subsections that are out of scope.

**Hard rule:** every shell command runs from the involved app's current apply worktree. Never use relative `../gymie-frontend` / `../gymie-admin-frontend`.

```markdown
# Manual test — <change-name>

**Purpose:** <one sentence from proposal>
**Branch / worktree:** `<branch>` · backend `<BACKEND_ROOT>` · frontend `<FRONTEND_ROOT>` · admin `<ADMIN_ROOT>`
**(Omit frontend/admin path segments when out of scope.)**
**Next after sign-off:** `/myflow-review <change-name>`
<!-- Skip mode only: add a line here — **Manual test status:** SKIPPED — YYYY-MM-DD (Gate C intentionally bypassed) -->
<!-- and leave every checklist / sign-off box below unchecked. -->

## How to run (involved apps)

### Prerequisites

- Docker running
- JDK 21 on `PATH`
- Apply worktrees listed in the header (do **not** run from main `develop` checkouts for this change)

### Start local stack

From the **backend apply worktree**, point `devStart` at the **frontend apply worktree** when KMP is in scope (default `../gymie-frontend` is wrong from `.worktrees/`):

```bash
cd <BACKEND_ROOT>

docker compose up -d

./gradlew dbSeed

./gradlew devStart \
  -PfrontendRoot=<FRONTEND_ROOT>
```

<!-- If KMP is out of scope, drop -PfrontendRoot and the frontend header path. -->
<!-- If admin is in scope and started by the stack, note it; otherwise use the admin-only block below. -->

Useful:

```bash
cd <BACKEND_ROOT>
./gradlew devStatus
./gradlew devLogs
./gradlew devStop
```

### Apps in scope for this change

<!-- Keep only rows that apply -->

| App | URL | Notes |
|-----|-----|-------|
| Gateway + API | http://localhost:8080 | Clients always hit gateway, not :8081 |
| Backend app | http://localhost:8081/actuator/health | Health only |
| KMP web (`gymie-frontend` worktree) | http://localhost:3000 | Started by `devStart` when `-PfrontendRoot` is set |
| Admin panel (`gymie-admin-frontend` worktree) | http://localhost:3001 | `superadmin@gymie.dev` / `GymieDev1` after `dbSeed` |

**Seeded credentials (after `dbSeed`):** document which apply (`user@…` / `superadmin@…`).

### Optional: run only touched surfaces

<!-- Include only if helpful; otherwise delete this subsection. Always absolute cds. -->

**Backend only:**

```bash
cd <BACKEND_ROOT>
docker compose up -d
./gradlew :app:bootRun --args='--spring.profiles.active=local'
# other terminal:
cd <BACKEND_ROOT>
./gradlew :gateway:bootRun --args='--spring.profiles.active=local'
```

**Admin frontend only** (gateway already up):

```bash
cd <ADMIN_ROOT>
npm install
npm run dev                  # http://localhost:3001
```

**KMP web only** (gateway already up):

```bash
cd <FRONTEND_ROOT>
./gradlew :webApp:jsBrowserDevelopmentRun
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
