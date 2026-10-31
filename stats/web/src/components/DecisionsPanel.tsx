// The run's latest recorded decision, one line per role: the model and
// effort the run gave it, the decision's reason on hover.

// Older rows hold `pairs` as prose rather than a list; this panel ignores those.
type Override = { to?: string; pairs?: unknown; instruction?: string };

type Pick = { model?: string; effort?: string; reason?: string; slots?: string[]; bundles?: number[] };

/** A recorded decision, as `flow record decision` stores it; only the fields this panel reads. */
export interface Decision {
  class?: string;
  execution?: string;
  implementer?: Pick;
  groups?: Pick[];
  fixer?: Pick;
  panel?: { dispatches?: Pick[]; rerun_dispatch?: Pick };
  extra_dispatch?: Pick;
  extra_dispatch_mutation?: Pick;
  /** An array of Override; legacy rows hold an object, which this panel ignores. */
  overrides?: unknown;
}

type Line = { role: string; model: string; reason?: string };

const pick = (role: string, p: Pick | undefined): Line[] =>
  p?.model ? [{ role, model: p.effort ? `${p.model} · ${p.effort}` : p.model, reason: p.reason }] : [];

const review = (p: Pick | undefined): Line[] => pick(`Review: ${p?.slots?.join(", ") ?? ""}`, p);

function lines(d: Decision): Line[] {
  // An override names its pair as "<model>/<agent definition>", the definition ending in its effort.
  const overrides = (Array.isArray(d.overrides) ? (d.overrides as Override[]) : []).flatMap((o) => {
    const [model, agent] = (o.to ?? "").split("/");
    const effort = agent?.split("-").at(-1);
    return (Array.isArray(o.pairs) ? o.pairs : []).map((role) => ({ role: role[0].toUpperCase() + role.slice(1), model: `${model} · ${effort}`, reason: o.instruction }));
  });
  const implementers = d.groups?.length
    ? d.groups.flatMap((g) => pick(`${(g.bundles?.length ?? 0) > 1 ? "Tasks" : "Task"} ${(g.bundles ?? []).join(", ")}`, g))
    : pick("Implementer", d.implementer);
  return [
    ...overrides,
    ...implementers,
    ...pick("Fixer", d.fixer),
    ...(d.panel?.dispatches ?? []).flatMap(review),
    ...pick("Review re-run", d.panel?.rerun_dispatch),
    ...review(d.extra_dispatch),
    ...review(d.extra_dispatch_mutation),
  ];
}

export function DecisionsPanel({ decision }: { decision: Decision }) {
  return (
    <section className="lb-panel" aria-label="Decisions">
      <div className="lb-panel-h">
        <span>Decisions</span>
        {decision.class && <span className="lb-panel-note">{[decision.class, decision.execution].filter(Boolean).join(" · ")}</span>}
      </div>
      <div className="lb-panel-b">
        {lines(decision).map((l, i) => (
          <div key={i} className="lb-decision" data-testid="decision-row" title={l.reason}>
            <span className="lb-n">{l.role}</span>
            <span className="r">{l.model}</span>
          </div>
        ))}
      </div>
    </section>
  );
}
