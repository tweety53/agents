// The run's latest recorded decision, one line per role: the model and
// effort the run gave it, the decision's reason on hover. A plan-check
// decision names no model: each line shows the pair the run's own dispatch for that point was given
// (matched by its key), or "not run yet", with the decision's reason beneath it.

import type { ConfidenceRow, RecordDispatch } from "../api";

// Older rows hold `pairs` as prose rather than a list; this panel ignores those.
type Override = { to?: string; pairs?: unknown; instruction?: string };

type Pick = { model?: string; effort?: string; reason?: string; slots?: string[]; bundles?: number[]; tasks?: string[] };

/** A recorded decision, as `flow record decision` stores it; only the fields this panel reads. */
export interface Decision {
  class?: string;
  execution?: string;
  implementer?: Pick;
  groups?: Pick[];
  fixer?: Pick;
  /** A plan-check decision lists its review bundles; an older one nests them under dispatches. */
  panel?: Pick[] | { dispatches?: Pick[]; rerun_dispatch?: Pick };
  visual?: { verify?: string; reason?: string };
  extra_dispatch?: Pick;
  extra_dispatch_mutation?: Pick;
  /** An array of Override; legacy rows hold an object, which this panel ignores. */
  overrides?: unknown;
}

type Line = { role: string; model: string; reason?: string; prose?: boolean; pct?: string };

/** The pair of the latest dispatch whose key `want` accepts, "not run yet" when none ran. */
function decided(dispatches: RecordDispatch[], want: (key: string) => boolean): string {
  const d = dispatches.filter((x) => x.key && want(x.key)).sort((a, b) => Date.parse(b.startedAt) - Date.parse(a.startedAt))[0];
  return d ? (d.effort && d.effort !== "default" ? `${d.model} · ${d.effort}` : d.model) : "not run yet";
}

/** An implementer dispatch covers a group's tasks when its key `task-<ids>-implementer` names every one: the run may bundle tasks the plan listed apart. */
const ranWith = (key: string, ids: string[]) => {
  const m = /^task-(\d+(?:\+\d+)*)-implementer$/.exec(key);
  return !!m && ids.every((i) => m[1].split("+").includes(i));
};

const pick = (role: string, p: Pick | undefined, run?: string): Line[] =>
  p?.model
    ? [{ role, model: p.effort ? `${p.model} · ${p.effort}` : p.model, reason: p.reason }]
    : p?.reason
      ? [{ role, model: run ?? "", reason: p.reason, prose: true }]
      : [];

const escapeRe = (x: string) => x.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");

/** The lowest estimate an implementer filed for any task of a group, "70% → 95%" once its tests were green; none when it filed none. */
const confidenceOf = (rows: ConfidenceRow[], ids: string[]): string | undefined => {
  const hit = rows.filter((r) => r.stage === "implement" && (r.task ?? "").split("+").some((t) => ids.includes(t)));
  if (hit.length === 0) return undefined;
  const low = hit.reduce((a, b) => (b.score < a.score ? b : a));
  return low.scoreAfter ? `${low.score}% → ${low.scoreAfter}%` : `${low.score}%`;
};

function lines(d: Decision, dispatches: RecordDispatch[], confidence: ConfidenceRow[]): Line[] {
  const review = (p: Pick | undefined): Line[] => {
    // A multi-repository change dispatches each slot per repo: `panel-<round>-<slots>-<repo>`.
    const re = new RegExp(`^panel-\\d+-${escapeRe((p?.slots ?? []).join("+"))}(-[a-z0-9-]+)?$`);
    return pick(`Review: ${p?.slots?.join(", ") ?? ""}`, p, decided(dispatches, (k) => re.test(k)));
  };
  // An override names its pair as "<model>/<agent definition>", the definition ending in its effort.
  const overrides = (Array.isArray(d.overrides) ? (d.overrides as Override[]) : []).flatMap((o) => {
    const [model, agent] = (o.to ?? "").split("/");
    const effort = agent?.split("-").at(-1);
    return (Array.isArray(o.pairs) ? o.pairs : []).map((role) => ({ role: role[0].toUpperCase() + role.slice(1), model: `${model} · ${effort}`, reason: o.instruction }));
  });
  const implementers = d.groups?.length
    ? d.groups.flatMap((g) =>
        ((ids) => pick(`${ids.length > 1 ? "Tasks" : "Task"} ${ids.join(", ")}`, g, decided(dispatches, (k) => ranWith(k, ids))).map((l) => ({ ...l, pct: confidenceOf(confidence, ids) })))(
          g.tasks ?? (g.bundles ?? []).map(String),
        ),
      )
    : pick("Implementer", d.implementer);
  return [
    ...overrides,
    ...implementers,
    ...pick("Fixer", d.fixer),
    ...(Array.isArray(d.panel) ? d.panel : (d.panel?.dispatches ?? [])).flatMap(review),
    ...pick("Review re-run", Array.isArray(d.panel) ? undefined : d.panel?.rerun_dispatch),
    ...review(d.extra_dispatch),
    ...review(d.extra_dispatch_mutation),
    ...(d.visual?.verify ? pick(`Visual check: ${d.visual.verify}`, { reason: d.visual.reason ?? "" }, decided(dispatches, (k) => /^visual-verify(?:-|$)/.test(k))) : []),
  ];
}

/** `confidence` is the change's estimates (KAN-1130): a group that filed one shows its percentage beside its pair. */
export function DecisionsPanel({ decision, dispatches = [], confidence = [] }: { decision: Decision; dispatches?: RecordDispatch[]; confidence?: ConfidenceRow[] }) {
  return (
    <section className="lb-panel" aria-label="Decisions">
      <div className="lb-panel-h">
        <span>Decisions</span>
        {decision.class && <span className="lb-panel-note">{[decision.class, decision.execution].filter(Boolean).join(" · ")}</span>}
      </div>
      <div className="lb-panel-b lb-decisions">
        {lines(decision, dispatches, confidence).map((l, i) => (
          <div key={i} className={l.prose ? "lb-decision lb-decision-prose" : "lb-decision"} data-testid="decision-row" title={l.prose ? undefined : l.reason}>
            <span className="lb-decision-head">
              <span className="lb-n">{l.role}</span>
              <span className={l.model === "not run yet" ? "lb-muted r" : "r"}>
                {l.model}
                {l.pct !== undefined && (
                  <>
                    {" · "}
                    <span className="lb-num" data-testid="decision-confidence">{l.pct}</span>
                  </>
                )}
              </span>
            </span>
            {l.prose && <span className="lb-decision-why">{l.reason}</span>}
          </div>
        ))}
      </div>
    </section>
  );
}
