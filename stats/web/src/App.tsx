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
import type { Period } from "./components/PeriodPicker";
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

function currentRoute(): Route {
  const hash = window.location.hash.replace(/^#\/?/, "");
  const run = parseRunRoute(hash);
  if (run) {
    return { kind: "run", ...run };
  }
  return { kind: "view", view: isViewName(hash) ? hash : DEFAULT_VIEW };
}

function useHashRoute(): Route {
  const [route, setRoute] = useState<Route>(currentRoute());
  useEffect(() => {
    const onHashChange = () => setRoute(currentRoute());
    window.addEventListener("hashchange", onHashChange);
    return () => window.removeEventListener("hashchange", onHashChange);
  }, []);
  return route;
}

function defaultPeriod(): Period {
  const to = new Date();
  const from = new Date(to.getTime() - 30 * 24 * 60 * 60 * 1000);
  return { from, to };
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
  const route = useHashRoute();
  const [period, setPeriod] = useState<Period>(defaultPeriod);
  const [project, setProject] = useState<string | undefined>(undefined);
  const [model, setModel] = useState<string | undefined>(undefined);

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
          ActiveView && <ActiveView period={period} project={project} model={modelDisabled ? undefined : model} />
        )}
      </main>
    </div>
  );
}
