// Flow guards -- "Which guards catch something, which never fire, and what
// does each cost in run time?" (decision flow-health-question), built as the
// approved prototype draws it. The eight period totals it used to show are
// gone, each answered by a sibling view (drop-eight-tiles); the name and
// route stay (keep-name-and-route).
//
// One endpoint, internal/api/health.go's guard-health: one row per guard with
// runs in the window and project, flow-guard's helper commands left out.
// The view owns its filters: a KAN-1048 chip bar of Period (7d / 30d / 90d,
// a rolling window computed when the chip is chosen, 30d by default --
// default-period-30d), Project (from the response's `projects`, each with
// its guard-run count) and Rank by (Fires by default -- default-rank-fires),
// replacing the shared dashboard bar, so App's period, project and model
// props are ignored (chip-filters-period-project-rank). Chip state is local.
//
// Every rank, tile, share and footer total is derived here from the rows
// (candidates-derived-in-spa): ties break by guard name (ties-by-guard-name);
// Each tile is a toggle filtering the table to its set
// (tiles-toggle-table-filter). The panel note is the prototype's text
// (panel-note-as-prototype); formats follow flow-health-display-formats;
// loading, error, not-recorded and an empty window read as Reviewers renders
// them (flow-health-empty-and-loading).
import { useState } from "react";
import type { GuardHealthRow, GuardProject } from "../api";
import { Chip, ChipGroup } from "../components/Chip";
import { Panel } from "../components/Panel";
import { PERIOD_DAYS, type PeriodChip, windowFor } from "../format";
import { useStatsView } from "../hooks/useStatsView";
import { projectLabel } from "../lib/projectLabel";

const RANKS = { fires: "Fires", rate: "Fire rate", silent: "Never fired", time: "Run time" } as const;
type Rank = keyof typeof RANKS;
const TITLE: Record<Rank, string> = {
  fires: "Every guard, most fires first",
  rate: "Every guard, highest fire rate first",
  silent: "Every guard, never-fired first, most runs first",
  time: "Every guard, most run time first",
};

interface Filters {
  period: PeriodChip;
  project: string | null;
  from: Date;
  to: Date;
}

const defaults = (): Filters => ({ ...windowFor("30d"), project: null });

const int = (v: number) => v.toLocaleString("en-US");
const pct = (v: number) => (v === 0 ? "0%" : v < 0.001 ? "<0.1%" : v < 0.1 ? `${(v * 100).toFixed(1)}%` : `${Math.round(v * 100)}%`);
const dur = (ms: number) => {
  if (Math.round(ms) < 1000) return `${Math.round(ms)} ms`;
  const tenths = Math.round(ms / 100);
  if (tenths < 600) return `${(tenths / 10).toFixed(1)} s`;
  const m = Math.round(ms / 60000);
  if (m < 60) return `${m} min`;
  return `${Math.floor(m / 60)} h ${m % 60} min`;
};
const total = (rows: GuardHealthRow[]) => rows.reduce((s, r) => s + r.totalMs, 0);

/** What each guard does, and whether the operator approved it as useful. Edited by hand: a guard missing here shows neither. */
const GUARD_INFO: Record<string, { does: string; approved: boolean }> = {
  "check-guard-calls": { does: "Lint: every flow-guard call in docs, skills and hooks names a real guard and has its own test", approved: true },
  "render-frames": { does: "Renders the design repo's tagged frames a plan cites into the project's mockups folder", approved: true },
  "check-base-behind": { does: "Compares each worktree's recorded base with the branch origin tracks, before a stale base fails things", approved: true },
  "end-dispatch-row": { does: "Closes a stopping subagent's dispatch row with its outcome and commit", approved: true },
  "check-visual-trigger": { does: "Tells the verify stage whether a diff touches ui paths, so the UI check runs", approved: true },
  "check-unfinished-work": { does: "Integrate gate: reports whether the change still has unfinished work", approved: true },
  "sweep-dispatch-rows": { does: "Closes dispatch rows whose stop was missed, from the agent's hand-back report", approved: true },
  "affected-specs": { does: "Prints the capture specs a diff can move, so a fix round recaptures those and no others", approved: true },
  "break-and-prove": { does: "Flips a file with a sed or patch, runs the test command and restores the file, proving the test fails without the fix", approved: true },
  "capture-workers": { does: "Prints how many Playwright workers a capture run takes, bounded by the machine's cores", approved: true },
  "check-dev-stack-fresh": { does: "Before a handoff prints a dev stack's URL, checks it serves the worktree's own build", approved: true },
  "check-finish-preflight": { does: "Decides whether /flow should integrate, from the worktree, its base and the recorded merge base", approved: true },
  "check-main-checkout-drift": { does: "Names how a main checkout drifts from its expected post-merge state, before a run's preflight reaches it", approved: true },
  "check-identical-baselines": { does: "Names the baseline PNGs of a directory that are byte-identical, so the verifier explains each group", approved: true },
  "recapture": { does: "Records a receipt for the specs a recapture run covered, on exactly the source files the round changed", approved: true },
  "run-reproducer": { does: "Runs a reviewer's command in a throwaway copy of the worktree, so it touches neither the worktree nor the machine", approved: true },
  "smoke-keep": { does: "Prints the heavy tests the files a branch touched keep in a default test run", approved: true },
  "test-cost": { does: "Reads go test -json on stdin and prints what each top-level test costs the run", approved: true },
  "rebuild-full-app-suite-zip": { does: "Rebuilds the full-app-suite zip from exactly the PNGs the full-app suite produced", approved: true },
  "resolve-visual-screenshots": { does: "Resolves which screenshots a visual run produced, as absolute PNG paths", approved: true },
  "visual-verify-pass": { does: "The parent's side of a visual verification: one call before each verifier dispatch, one after", approved: true },
  "measure-visual-properties": { does: "Measures each capture against its mockup frame and prints the differing-pixel ratio", approved: true },
  "unchanged": { does: "Tells a test line whether the branch left every path it reads untouched, so the suite can be skipped", approved: true },
  "compose-mockup-frames": { does: "Pairs each captured screenshot with the mockup frame it was drawn from, side by side", approved: true },
  "write-panel-diff": { does: "Writes one review-panel diff and its touched-files list", approved: true },
  "kickoff-worktree": { does: "Creates a change's worktree on its own branch and records its base", approved: true },
  "prepare-workspace": { does: "Prepares a worktree for a change: its own database, cache index and ports, then its lint list", approved: true },
  "workspace": { does: "Creates and removes a change's own database and bucket", approved: true },
  "throwaway-worktree": { does: "Creates or removes the throwaway copy of a worktree that a review slot works in", approved: true },
  "remove-change-worktrees": { does: "Run 2's worktree cleanup: every check first, then the worktrees and the local and remote branch", approved: true },
  "resolve-base-branch": { does: "The single place the base branch of a worktree is resolved; prints the bare branch name", approved: true },
  "sync-onto-base": { does: "Syncs a finishing worktree onto its base's tip, with a resume after a conflict", approved: true },
  "refresh-main-checkout": { does: "Fast-forwards a main checkout onto origin once a change has landed, only when nothing in it can be lost", approved: true },
  "live-store-copy": { does: "Copies the dev database into a scratch one so a live check runs on real data without touching the dev stack", approved: true },
  "installer-sandbox-diff": { does: "Runs setup.sh global from two trees under fresh sandboxes and diffs what each installed", approved: true },
  "project-get": { does: "Reads one section of a project's .flow/project.md", approved: true },
  "integrate-change": { does: "The mechanical half of a bare /flow at IN_PROGRESS: integrate, then cleanup, through the guards", approved: true },
  "close-task": { does: "Ticks and pushes every task a returning implementer committed", approved: true },
  "plan-check": { does: "Reviews an approved plan and decides its implementer groups, review panel and visual verdict in one headless call", approved: true },
  "pair-judge": { does: "Picks the model and effort of one judged dispatch from its role's allowed pairs", approved: true },
  "pipeline-fix": { does: "The in-run fix loop for a pipeline defect: begin, land and end", approved: true },
  "load-sets": { does: "Reports the static instruction bytes each /flow session loads; never a verdict", approved: true },
  "recover-guard-incident": { does: "Aborts a stuck git revert after printing the reflog diagnosis", approved: true },
  "run-verify-list": { does: "Runs a project's lint, test or visual list for a worktree and stamps the green commit", approved: true },
  "check-loop-dispatch": { does: "Begins and ends the dispatch rows of a pipeline-fix or self-review loop under the change it belongs to", approved: true },
  "check-fix-revert": { does: "After a fix commit, refuses one that undoes an earlier fix of the same change", approved: false },
  "check-frame-constants": { does: "Refuses frames of one change that draw a shared constant two ways, such as bar heights or greys", approved: false },
  "check-ui-task-measured": { does: "Gate for a UI task: answers whether it was measured clean against its frames after its last source commit", approved: false },
  "check-verify-env": { does: "Before a verify round, proves every backend the suite talks to accepts a connection", approved: false },
  "end-stale-dispatch-rows": { does: "Ends the change's open dispatch rows whose agent finished and went quiet", approved: false },
  "await-rows": { does: "The parent's one bounded wait for its children's dispatch rows to end", approved: false },
  "run-detached": { does: "Starts a command that can outrun the shell's foreground limit, detached, recording its log and exit code", approved: false },
  "await-detached": { does: "Waits one short call for a detached command, printing RUNNING, EXITED with its exit code, or LOST", approved: false },
  "sync-base-behind": { does: "Runs the whole base sync that check-base-behind prints the commands for, so no raw git push is typed", approved: false },
  "stop-worktree-processes": { does: "Stops every process running inside a worktree, before the worktree is removed", approved: false },
};

/** The panels the page draws, in order: every script of flow-guard by what it is for. A guard missing here lands in Other. */
const GROUPS: { title: string; guards: string[] }[] = [
  { title: "Checks", guards: ["check-guard-calls", "check-base-behind", "check-unfinished-work", "check-finish-preflight", "check-visual-trigger", "check-dev-stack-fresh", "check-workspace-isolation", "check-main-checkout-drift", "check-identical-baselines", "check-fix-revert", "check-frame-constants", "check-ui-task-measured", "check-verify-env"] },
  { title: "Dispatch bookkeeping", guards: ["end-dispatch-row", "sweep-dispatch-rows", "end-stale-dispatch-rows", "check-loop-dispatch", "await-rows"] },
  { title: "Verify and test tooling", guards: ["run-verify-list", "recapture", "affected-specs", "capture-workers", "break-and-prove", "run-reproducer", "smoke-keep", "test-cost", "test-two-streams", "rebuild-full-app-suite-zip", "resolve-visual-screenshots", "visual-verify-pass", "measure-visual-properties", "unchanged", "run-detached", "await-detached"] },
  { title: "Mockups and frames", guards: ["render-frames", "compose-mockup-frames", "write-panel-diff"] },
  { title: "Workspace and git", guards: ["kickoff-worktree", "prepare-workspace", "workspace", "throwaway-worktree", "remove-change-worktrees", "resolve-base-branch", "sync-onto-base", "refresh-main-checkout", "live-store-copy", "installer-sandbox-diff", "project-get", "sync-base-behind", "stop-worktree-processes"] },
  { title: "Pipeline orchestration", guards: ["integrate-change", "close-task", "plan-check", "pair-judge", "pipeline-fix", "load-sets", "recover-guard-incident"] },
];
const GROUP_NOTE: Record<string, string> = {
  Checks: "gates that can refuse",
  "Dispatch bookkeeping": "close and open the rows a dispatch leaves in the store",
  "Verify and test tooling": "run, scope and prove the project's tests and captures",
  "Mockups and frames": "render and compare the design frames a plan cites",
  "Workspace and git": "worktrees, bases and the main checkout",
  "Pipeline orchestration": "plan, land and tidy a change",
  Other: "",
};
const groupOf = (guard: string) => GROUPS.find((g) => g.guards.includes(guard))?.title ?? "Other";

/** A guard that ran inside another guard's run belongs to its caller's panel, nested under it. */
export function panelOf(r: GuardHealthRow, rows: GuardHealthRow[]): string {
  let cur = r;
  for (const seen = new Set([r.guard]); cur.parent; ) {
    const p = rows.find((x) => x.guard === cur.parent);
    if (!p || seen.has(p.guard)) break;
    seen.add(p.guard);
    cur = p;
  }
  return groupOf(cur.guard);
}

const byName = (a: GuardHealthRow, b: GuardHealthRow) => a.guard.localeCompare(b.guard);
const ORDER: Record<Rank, (a: GuardHealthRow, b: GuardHealthRow) => number> = {
  fires: (a, b) => b.fired - a.fired || b.runs - a.runs,
  rate: (a, b) => b.fired / b.runs - a.fired / a.runs || b.fired - a.fired,
  silent: (a, b) => Number(a.fired > 0) - Number(b.fired > 0) || b.runs - a.runs,
  time: (a, b) => b.totalMs - a.totalMs,
};
const ranked = (rows: GuardHealthRow[], rank: Rank) => [...rows].sort((a, b) => ORDER[rank](a, b) || byName(a, b));

/** Each guard that ran inside another guard of the same panel sits right under its caller, ranked order kept otherwise. */
export function nested(rows: GuardHealthRow[]): GuardHealthRow[] {
  const here = new Set(rows.map((r) => r.guard));
  const kids = (g: string) => rows.filter((r) => r.parent === g);
  const top = rows.filter((r) => !r.parent || !here.has(r.parent));
  const walk = (r: GuardHealthRow, seen: Set<string>): GuardHealthRow[] =>
    seen.has(r.guard) ? [] : [r, ...kids(r.guard).flatMap((k) => walk(k, new Set([...seen, r.guard])))];
  return top.flatMap((r) => walk(r, new Set()));
}

const Slot = ({ name }: { name: string }) => <span className="rv-slot">{name}</span>;

function Table({ rows, all, rank }: { rows: GuardHealthRow[]; all: number; rank: Rank }) {
  // A panel of helpers only has no fire columns: a helper refuses nothing.
  const fires = rows.some((r) => !r.helper);
  const sort = (k: Rank) => (k === rank ? "descending" : undefined);
  const share = (v: number) => (
    <>
      {pct(v)}
      <span className="wc-sharebar">
        <span style={{ width: `${(v * 100).toFixed(1)}%` }} />
      </span>
    </>
  );
  const t = rows.reduce(
    (s, r) => ({ runs: s.runs + r.runs, fired: s.fired + r.fired, ca: s.ca + r.cannotAnswer, total: s.total + r.totalMs }),
    { runs: 0, fired: 0, ca: 0, total: 0 },
  );
  return (
    <table className="lb-table">
      <thead>
        <tr>
          <th>#</th>
          <th>Guard</th>
          <th className="r">Runs</th>
          {fires && (
            <>
              <th className="r" aria-sort={sort("fires")}>
                Fired
              </th>
              <th className="r" aria-sort={sort("rate")}>
                Fire rate
              </th>
              <th className="r">Cannot answer</th>
            </>
          )}
          <th className="r">Median run</th>
          <th className="r" aria-sort={sort("time")}>
            Run time
          </th>
          <th className="r">Share of run time</th>
          {fires && <th aria-sort={sort("silent")}>Last fired</th>}
        </tr>
      </thead>
      <tbody>
        {nested(ranked(rows, rank)).map((r, i) => (
          <tr key={r.guard} className={r.fired || r.helper ? undefined : "lb-muted"}>
            <td className="lb-num lb-muted">{i + 1}</td>
            <td style={r.parent && rows.some((p) => p.guard === r.parent) ? { paddingLeft: 24 } : undefined}>
              {r.parent && rows.some((p) => p.guard === r.parent) && <span className="lb-muted" title={`runs inside ${r.parent}`}>↳ </span>}
              <Slot name={r.guard} />
              {GUARD_INFO[r.guard]?.approved && <span className="lb-pill fh-approved">approved</span>}
              {GUARD_INFO[r.guard] && (
                <span className="fh-info" role="img" tabIndex={0} title={GUARD_INFO[r.guard].does} aria-label={GUARD_INFO[r.guard].does}>
                  i
                </span>
              )}
            </td>
            <td className="r lb-num">{int(r.runs)}</td>
            {fires && (
              <>
                <td className="r lb-num">{r.helper ? "" : r.fired ? int(r.fired) : "—"}</td>
                <td className="r lb-num">{r.helper ? "" : pct(r.fired / r.runs)}</td>
                <td className="r lb-num">{r.helper ? "" : r.cannotAnswer ? pct(r.cannotAnswer / r.runs) : "—"}</td>
              </>
            )}
            <td className="r lb-num">{dur(r.medianMs)}</td>
            <td className="r lb-num">{dur(r.totalMs)}</td>
            <td className="r lb-num">{share(all ? r.totalMs / all : 0)}</td>
            {fires && <td className="lb-num">{r.helper ? "" : r.lastFiredAt ? r.lastFiredAt.slice(0, 10) : <span className="lb-muted">never</span>}</td>}
          </tr>
        ))}
      </tbody>
      <tfoot>
        <tr>
          <td></td>
          <td>
            {rows.length} {rows.length === 1 ? "script" : "scripts"}
          </td>
          <td className="r lb-num">{int(t.runs)}</td>
          {fires && (
            <>
              <td className="r lb-num">{int(t.fired)}</td>
              <td className="r lb-num">{pct(t.runs ? t.fired / t.runs : 0)}</td>
              <td className="r lb-num">{pct(t.runs ? t.ca / t.runs : 0)}</td>
            </>
          )}
          <td></td>
          <td className="r lb-num">{dur(t.total)}</td>
          <td className="r lb-num">{pct(all ? t.total / all : 0)}</td>
          {fires && <td></td>}
        </tr>
      </tfoot>
    </table>
  );
}

export function FlowHealth() {
  const [f, setF] = useState<Filters>(defaults);
  const [rank, setRank] = useState<Rank>("fires");
  const state = useStatsView<GuardHealthRow[], GuardProject>("guard-health", {
    from: f.from,
    to: f.to,
    project: f.project ?? undefined,
    projects: true,
  });

  const offDefaults = f.period !== "30d" || f.project !== null;
  const ready = state.status === "ready" && state.data.recorded && !state.data.unmeasured;
  const rows = state.status === "ready" ? state.data.rows : [];
  const projects = state.status === "ready" ? (state.data.projects ?? []) : [];
  const panels = [...GROUPS.map((g) => g.title), "Other"]
    .map((title) => ({ title, rows: rows.filter((r) => panelOf(r, rows) === title) }))
    .filter((p) => p.rows.length > 0);

  return (
    <section aria-label="Flow guards" className="dashboard lb sl">
      <div>
        <h2>Flow guards</h2>
        <p className="lb-sub">Which guards catch something, which never fire, and what does each cost in run time?</p>
      </div>
      <div className="lb-fbar" role="toolbar" aria-label="Filters">
        <ChipGroup name="Period">
          {(Object.keys(PERIOD_DAYS) as PeriodChip[]).map((p) => (
            <Chip key={p} on={f.period === p} onClick={() => setF({ ...f, ...windowFor(p) })}>
              {p}
            </Chip>
          ))}
        </ChipGroup>
        <ChipGroup name="Project">
          <Chip on={f.project === null} onClick={() => setF({ ...f, project: null })}>
            All
          </Chip>
          {projects.map((p) => (
            <Chip key={p.projectKey} on={f.project === p.projectKey} onClick={() => setF({ ...f, project: p.projectKey })} title={p.projectKey}>
              {projectLabel(p.projectKey)}
              <span className="lb-c">{int(p.runs)}</span>
            </Chip>
          ))}
        </ChipGroup>
        <ChipGroup name="Rank by">
          {(Object.keys(RANKS) as Rank[]).map((k) => (
            <Chip key={k} on={rank === k} onClick={() => setRank(k)}>
              {RANKS[k]}
            </Chip>
          ))}
        </ChipGroup>
        {offDefaults && (
          <button
            type="button"
            className="lb-fclear"
            onClick={() => {
              setF(defaults());
            }}
          >
            Clear filters
          </button>
        )}
      </div>
      {ready && rows.length > 0 && (
        <p className="lb-sub">
          {TITLE[rank]}. fired = the guard refused; helper commands refuse nothing, so show no fire columns.
        </p>
      )}
      {!ready ? (
        // Loading, error, not-recorded and unmeasured read as Panel renders them.
        <Panel title="Flow guards" state={state}>
          {() => null}
        </Panel>
      ) : !rows.length ? (
        <div className="lb-panel">
          <div className="sl-empty">No guard runs in this period.</div>
        </div>
      ) : (
        panels.map((p) => {
          return (
            <div className="lb-panel" key={p.title} aria-label={p.title}>
              <div className="lb-panel-h">
                <span>{p.title}</span>
                <span className="lb-panel-note">{GROUP_NOTE[p.title]}</span>
              </div>
              <div className="lb-scroll">
                <Table rows={p.rows} all={total(p.rows)} rank={rank} />
              </div>
            </div>
          );
        })
      )}
    </section>
  );
}
