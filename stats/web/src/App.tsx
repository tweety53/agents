// App is the shell every view mounts inside: a minimal hash router with no
// added dependency (every view carries its own Period and Project chips) -- the
// static routes plus a redirect are well inside what `window.location.hash`
// can carry, and pulling in a router library for that would be exactly the
// "config switch/registry with one implementation" KISS violation
// (engineering-principles.md) for a change this size.
//
// The live state board is the default route (task 13's own requirement:
// "it is the view that replaces reading a state file, so it is what opens
// first") -- an empty or unrecognised hash both resolve to it.
import { useEffect, useState, type ReactElement } from "react";
import { VIEW_NAMES, type ViewName } from "./api";
import { AgentRoles } from "./views/AgentRoles";
import { FlowHealth } from "./views/FlowHealth";
import { RunDetail } from "./views/RunDetail";
import { RunTimings } from "./views/RunTimings";
import { ReviewDynamics } from "./views/ReviewDynamics";
import { Reviewers } from "./views/Reviewers";
import { SelfReview } from "./views/SelfReview";
import { Confidence } from "./views/Confidence";
import { StageLeaderboard } from "./views/StageLeaderboard";
import { StateBoard } from "./views/StateBoard";
import { Trend } from "./views/Trend";
import { RoleCosts } from "./views/RoleCosts";
import { WeeklyCosts } from "./views/WeeklyCosts";

const DEFAULT_VIEW: ViewName = "state-board";

const VIEW_LABELS: Record<ViewName, string> = {
  "state-board": "Live state board",
  "stage-leaderboard": "Stage leaderboard",
  trend: "Trend over time",
  "weekly-costs": "Weekly costs",
  reviewers: "Reviewers",
  "flow-health": "Flow guards",
  "self-review": "Self-review",
  confidence: "Confidence",
  "review-dynamics": "Reviewer findings",
  "role-costs": "Role costs",
  "agent-roles": "Agent roles",
};

const VIEW_COMPONENTS: Record<ViewName, () => ReactElement> = {
  "state-board": StateBoard,
  "stage-leaderboard": StageLeaderboard,
  trend: Trend,
  "weekly-costs": WeeklyCosts,
  reviewers: Reviewers,
  "flow-health": FlowHealth,
  "self-review": SelfReview,
  confidence: Confidence,
  "review-dynamics": ReviewDynamics,
  "role-costs": RoleCosts,
  "agent-roles": AgentRoles,
};

function isViewName(v: string): v is ViewName {
  return (VIEW_NAMES as readonly string[]).includes(v);
}

/**
 * The route this app can be on: one of the static statistics views, or
 * the parametrised run-detail route a state-board row links into
 * (task 18, "One change opens on its own dashboard"). A discriminated
 * union rather than another ViewName -- run detail's URL carries data
 * (project, change) no static view slug does, so it cannot be a member of
 * VIEW_NAMES without forcing every other view-name check in this file to
 * also handle a "name" that isn't really one.
 */
type Route =
  | { kind: "view"; view: ViewName }
  | { kind: "run"; project: string; change: string }
  | { kind: "run-costs"; project: string; change: string };

/**
 * Matches "run/<project>/<change>" against the hash (already stripped of
 * its leading "#/"). Both segments are percent-decoded here -- they were
 * percent-encoded by StateBoard's own link (encodeURIComponent per
 * segment), which is what lets a project key or change name containing a
 * URL-significant character (a literal "/", "#", "?", ...) survive the
 * round trip without being mistaken for a route separator.
 */
function parseRunRoute(hash: string, re = /^run\/([^/]+)\/([^/]+)$/): { project: string; change: string } | null {
  const match = re.exec(hash);
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
  const runCosts = parseRunRoute(path, /^run-costs\/([^/]+)\/([^/]+)$/);
  if (runCosts) {
    return { kind: "run-costs", ...runCosts };
  }
  return { kind: "view", view: isViewName(path) ? path : DEFAULT_VIEW };
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

export function App() {
  const route = useHashRoute();

  const ActiveView = route.kind === "view" ? VIEW_COMPONENTS[route.view] : null;

  return (
    <div className="app-shell">
      <header>
        <h1>flow stats</h1>
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
      </header>
      <main>
        {route.kind === "run" ? (
          <RunDetail project={route.project} change={route.change} />
        ) : route.kind === "run-costs" ? (
          <RunTimings project={route.project} change={route.change} />
        ) : (
          ActiveView && (
            <ActiveView />
          )
        )}
      </main>
    </div>
  );
}
