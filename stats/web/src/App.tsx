// App is the shell every view mounts inside: the period and project
// controls (shared, because every view is driven by both -- design.md's
// "Filtering, searching and sorting" and "Project is a first-class
// filter"), and a minimal hash router with no added dependency -- eight
// static routes plus a redirect is well inside what `window.location.hash`
// can carry, and pulling in a router library for that would be exactly the
// "config switch/registry with one implementation" KISS violation
// (engineering-principles.md) for a change this size.
//
// The live state board is the default route (task 13's own requirement:
// "it is the view that replaces reading a state file, so it is what opens
// first") -- an empty or unrecognised hash both resolve to it.
import { useEffect, useState, type ReactElement } from "react";
import { VIEW_NAMES, type ViewName } from "./api";
import { DashboardBar } from "./components/DashboardBar";
import { defaultPeriod, type Period } from "./components/PeriodPicker";
import type { ViewProps } from "./viewTypes";
import { CacheEfficiency } from "./views/CacheEfficiency";
import { CostPerChange } from "./views/CostPerChange";
import { ModelComparison } from "./views/ModelComparison";
import { PanelEconomics } from "./views/PanelEconomics";
import { ReworkRate } from "./views/ReworkRate";
import { RunDetail } from "./views/RunDetail";
import { StageLeaderboard } from "./views/StageLeaderboard";
import { StateBoard } from "./views/StateBoard";
import { Trend } from "./views/Trend";

const DEFAULT_VIEW: ViewName = "state-board";

const VIEW_LABELS: Record<ViewName, string> = {
  "state-board": "Live state board",
  "cost-per-change": "Cost per change",
  "stage-leaderboard": "Stage leaderboard",
  trend: "Trend over time",
  "cache-efficiency": "Cache efficiency",
  "panel-economics": "Panel economics",
  "model-comparison": "Model comparison",
  "rework-rate": "Rework rate",
};

const VIEW_COMPONENTS: Record<ViewName, (props: ViewProps) => ReactElement> = {
  "state-board": StateBoard,
  "cost-per-change": CostPerChange,
  "stage-leaderboard": StageLeaderboard,
  trend: Trend,
  "cache-efficiency": CacheEfficiency,
  "panel-economics": PanelEconomics,
  "model-comparison": ModelComparison,
  "rework-rate": ReworkRate,
};

function isViewName(v: string): v is ViewName {
  return (VIEW_NAMES as readonly string[]).includes(v);
}

/**
 * The route this app can be on: one of the eight static statistics views,
 * or the parametrised run-detail route a state-board row links into
 * (task 18, "One change opens on its own dashboard"). A discriminated
 * union rather than a ninth ViewName -- run detail's URL carries data
 * (project, change) no static view slug does, so it cannot be a member of
 * VIEW_NAMES without forcing every other view-name check in this file to
 * also handle a "name" that isn't really one.
 */
type Route = { kind: "view"; view: ViewName } | { kind: "run"; project: string; change: string };

/**
 * Matches "run/<project>/<change>" against the hash (already stripped of
 * its leading "#/"). Both segments are percent-decoded here -- they were
 * percent-encoded by StateBoard's own link (encodeURIComponent per
 * segment), which is what lets a project key or change name containing a
 * URL-significant character (a literal "/", "#", "?", ...) survive the
 * round trip without being mistaken for a route separator.
 */
function parseRunRoute(hash: string): { project: string; change: string } | null {
  const match = /^run\/([^/]+)\/([^/]+)$/.exec(hash);
  if (!match) return null;
  return { project: decodeURIComponent(match[1]), change: decodeURIComponent(match[2]) };
}

// The hash is split into path and query before either matcher below ever
// sees it. `currentRoute()` used to match the *whole* hash, so appending
// `?from=…&to=…` (task 2) broke both matchers silently:
// `cost-per-change?from=…` is not a view name, so every static view fell
// back to the default view; and `^run/([^/]+)/([^/]+)$` absorbed
// `?from=…` into the change segment, so run detail opened on a change
// that did not exist. Neither matcher is taught about query strings --
// the point of the split is that they never see one.
function splitHash(hash: string): { path: string; query: string } {
  const qIndex = hash.indexOf("?");
  return qIndex === -1 ? { path: hash, query: "" } : { path: hash.slice(0, qIndex), query: hash.slice(qIndex + 1) };
}

function currentRoute(): Route {
  const hash = window.location.hash.replace(/^#\/?/, "");
  const { path } = splitHash(hash);
  const run = parseRunRoute(path);
  if (run) {
    return { kind: "run", ...run };
  }
  return { kind: "view", view: isViewName(path) ? path : DEFAULT_VIEW };
}

/**
 * Review finding F1: a `hashchange` can arrive with no reload -- pasting a
 * shared deep link into a tab where the app is already mounted fires it
 * just the same as clicking an in-app link -- so `useState`'s initialiser
 * never re-runs and the in-memory period survives untouched unless this
 * listener also re-derives it. `onUrlPeriod` is called only when the new
 * hash carries an explicit, validly-parsed period (`periodFromQuery`,
 * task 2's same fallback rule): an in-app nav click's hash carries no
 * query at all, and that case must leave the current period alone rather
 * than resetting it to the default on every click.
 */
function useHashRoute(onUrlPeriod: (period: Period) => void): Route {
  const [route, setRoute] = useState<Route>(currentRoute());
  useEffect(() => {
    const onHashChange = () => {
      setRoute(currentRoute());
      const hash = window.location.hash.replace(/^#\/?/, "");
      const { query } = splitHash(hash);
      const period = periodFromQuery(query);
      if (period) onUrlPeriod(period);
    };
    window.addEventListener("hashchange", onHashChange);
    return () => window.removeEventListener("hashchange", onHashChange);
    // onUrlPeriod is NOT a stable identity -- App creates it fresh on every
    // render. This effect subscribes once and so captures the mount-time
    // closure permanently, which is safe only because that closure's body
    // reads nothing from App's scope: it forwards to setPeriod's functional
    // update form, and setPeriod is the genuinely stable one.
    //
    // That safety is a property of the closure's body, not of the dependency
    // array. If onUrlPeriod is ever edited to read an outer variable directly
    // -- `route` or `period`, say, instead of the functional update -- this
    // handler will read that variable's mount-time value forever, because the
    // effect never resubscribes. Change the body and you must revisit this.
  }, []);
  return route;
}

/**
 * Reads `from`/`to` from the hash's query, applying them only when both
 * are present, both parse to a valid instant, and `from` is strictly
 * before `to`. Anything else -- missing, incomplete, unparsable or
 * reversed -- returns null so the caller falls back to the default period
 * rather than partially applying it. The URL is untrusted input: a
 * crafted one must not be able to produce an empty view with no
 * explanation, which is the failure this change exists to remove.
 */
function periodFromQuery(query: string): Period | null {
  const params = new URLSearchParams(query);
  const fromStr = params.get("from");
  const toStr = params.get("to");
  if (!fromStr || !toStr) return null;
  const from = new Date(fromStr);
  const to = new Date(toStr);
  if (Number.isNaN(from.getTime()) || Number.isNaN(to.getTime())) return null;
  if (from.getTime() >= to.getTime()) return null;
  return { from, to };
}

function currentPeriod(): Period {
  const hash = window.location.hash.replace(/^#\/?/, "");
  const { query } = splitHash(hash);
  return periodFromQuery(query) ?? defaultPeriod();
}

function routePath(route: Route): string {
  return route.kind === "run"
    ? `run/${encodeURIComponent(route.project)}/${encodeURIComponent(route.change)}`
    : route.view;
}

/**
 * Stated on the dashboard bar's model variable when it is disabled, and
 * echoing the server's own reason for rejecting the parameter there
 * (internal/api/stats.go's viewStateBoard branch) so a reader sees the
 * same explanation whichever side told them.
 */
const MODEL_DISABLED_REASON =
  "The live state board's rows are changes, not stage runs, so a model restriction cannot apply here.";

export function App() {
  const [period, setPeriod] = useState<Period>(currentPeriod);
  // A hashchange carrying the same period as the current one (e.g. the
  // write-back effect below round-tripping through history, or a repeat
  // paste of the same link) must not thrash: comparing by value, not by
  // the freshly-parsed Date objects' identity, is what keeps a no-op
  // arrival from triggering a re-render and re-fetch.
  const onUrlPeriod = (next: Period) => {
    setPeriod((prev) => (prev.from.getTime() === next.from.getTime() && prev.to.getTime() === next.to.getTime() ? prev : next));
  };
  const route = useHashRoute(onUrlPeriod);
  const [project, setProject] = useState<string | undefined>(undefined);
  const [model, setModel] = useState<string | undefined>(undefined);

  // Carries the period into the hash's query on every change, via
  // `replaceState` rather than a hash assignment -- adjusting a range is
  // exploration, not navigation, and one history entry per adjustment
  // would make the back button useless. `replaceState` does not dispatch
  // `hashchange`, so this write never feeds back into useHashRoute's own
  // listener as a route change.
  useEffect(() => {
    const params = new URLSearchParams({ from: period.from.toISOString(), to: period.to.toISOString() });
    const newHash = `#/${routePath(route)}?${params.toString()}`;
    if (window.location.hash !== newHash) {
      window.history.replaceState(null, "", newHash);
    }
  }, [route, period]);

  const ActiveView = route.kind === "view" ? VIEW_COMPONENTS[route.view] : null;
  // The state board is the one dashboard today whose rows are changes
  // rather than stage runs (proposal.md's Fix round 1) -- every other
  // static view aggregates stage runs and accepts the restriction.
  const modelDisabled = route.kind === "view" && route.view === "state-board";

  return (
    <div className="app-shell">
      <header>
        <h1>myflow stats</h1>
        <nav aria-label="Statistics views">
          <ul>
            {VIEW_NAMES.map((v) => (
              <li key={v}>
                <a href={`#/${v}`} aria-current={route.kind === "view" && route.view === v ? "page" : undefined}>
                  {VIEW_LABELS[v]}
                </a>
              </li>
            ))}
          </ul>
        </nav>
        <DashboardBar
          period={period}
          onPeriodChange={setPeriod}
          project={project}
          onProjectChange={setProject}
          model={model}
          onModelChange={setModel}
          modelDisabled={modelDisabled}
          modelDisabledReason={MODEL_DISABLED_REASON}
        />
      </header>
      <main>
        {route.kind === "run" ? (
          // The run-detail route honours the model variable rather than
          // leaving it rendering and inert (task 20's own decision): it is
          // never disabled here (modelDisabled is computed only for the
          // static ViewName routes, false by construction for "run"), and
          // RunDetail.tsx's own header comment explains why honouring is
          // cheap on this route specifically.
          <RunDetail project={route.project} change={route.change} model={model} />
        ) : (
          ActiveView && (
            <ActiveView
              period={period}
              onPeriodChange={setPeriod}
              project={project}
              model={modelDisabled ? undefined : model}
            />
          )
        )}
      </main>
    </div>
  );
}
