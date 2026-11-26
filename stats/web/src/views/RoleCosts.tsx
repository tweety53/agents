// RoleCosts is the Role costs view at #/role-costs, drawn as the approved
// prototype (.superpowers/sdd/<change>/role-costs-prototype.html): the
// Period and Project chips; each allowed-pairs role from GET
// /api/v1/stats/role-costs under its pipeline stage -- spend, agent time,
// per-dispatch figures, failure rate and a model · effort mix, every pair on
// click -- and the weekly spend split by stage.
import { Fragment, useEffect, useState } from "react";
import { fetchProjects, fetchStatsView, type RoleCostsResult } from "../api";
import { Chip, ChipGroup } from "../components/Chip";
import { formatUsd as formatUsd2, hm, PERIOD_DAYS, windowFor, type PeriodChip } from "../format";
import { projectLabel } from "../lib/projectLabel";
import { SHORT_DOES } from "./AgentRoles";

type Role = RoleCostsResult["roles"][number];

const PERIOD_LABELS: Record<PeriodChip, string> = { "7d": "7 days", "30d": "30 days", "90d": "90 days" };
const STAGE_COLOUR: Record<string, string> = {
  Planning: "#d97706",
  Implementation: "#dc2626",
  Review: "#2563eb",
  Verification: "#16a34a",
  "Self-review": "#9333ea",
  "Pipeline fix": "#0891b2",
  "Pair choice": "#64748b",
  Other: "#94a3b8",
};
const MODEL_COLOUR: Record<string, string> = { opus: "#ea580c", sonnet: "#2563eb", fable: "#9333ea" };
const EFFORT_OPACITY: Record<string, number> = { low: 0.4, default: 0.4, medium: 0.6, high: 0.8, xhigh: 1, max: 1 };

/** A priced amount under a cent reads <$0.01 -- a headless judge call costs a fraction of one. */
const formatUsd = (v: number) => (v > 0 && v < 0.005 ? "<$0.01" : formatUsd2(v));
const span = (s: number) => (s < 60 ? `${Math.round(s)}s` : s < 3600 ? `${Math.round(s / 60)}m` : hm(s * 1000));
const pct = (n: number, d: number) => (d > 0 ? `${Math.round((100 * n) / d)}%` : "—");
const sum = (xs: number[]) => xs.reduce((a, b) => a + b, 0);
const split = (pair: string) => pair.split(" · ") as [string, string];
const label = (role: string) => SHORT_DOES[role] ?? (role === "other" ? "Unplaced dispatches" : role);

/** Roles in stage order (the server's), one entry per stage. */
function byStage(roles: Role[]): { stage: string; roles: Role[] }[] {
  const out: { stage: string; roles: Role[] }[] = [];
  for (const r of roles) {
    const last = out[out.length - 1];
    if (last?.stage === r.group) last.roles.push(r);
    else out.push({ stage: r.group, roles: [r] });
  }
  return out;
}

function RoleName({ role }: { role: string }) {
  return (
    <>
      <b>{label(role)}</b>
      <code>{role}</code>
    </>
  );
}

function Mix({ r }: { r: Role }) {
  const priced = r.pairs.filter((p) => p.costUSD > 0);
  if (priced.length === 0) return <span className="rc-muted">no priced runs</span>;
  const [model, effort] = split(priced[0].pair);
  return (
    <div className="rc-mix">
      <div className="rc-mixbar" aria-hidden="true">
        {priced.map((p) => {
          const [m, e] = split(p.pair);
          return <span key={p.pair} style={{ width: `${(100 * p.costUSD) / r.costUSD}%`, background: MODEL_COLOUR[m] ?? "#888", opacity: EFFORT_OPACITY[e] ?? 0.6 }} />;
        })}
      </div>
      <span>
        <span className={`rc-model ${model}`}>{model}</span> {effort} · {pct(priced[0].costUSD, r.costUSD)}
      </span>
      {r.pairs.length > 1 && <span className="rc-muted">+{r.pairs.length - 1}</span>}
    </div>
  );
}

function Pairs({ r }: { r: Role }) {
  const priced = r.pairs.filter((p) => p.costUSD > 0);
  const unpriced = r.pairs.filter((p) => p.costUSD === 0);
  const n = sum(unpriced.map((p) => p.dispatches));
  const max = Math.max(0, ...priced.map((p) => p.costUSD));
  return (
    <div data-testid="role-pairs">
      {priced.map((p) => {
        const [m, e] = split(p.pair);
        return (
          <div key={p.pair} className="rc-pair" data-testid="role-pair">
            <span className={`rc-model ${m}`}>{m}</span>
            <span className="rc-muted">{e}</span>
            <span className="r">{formatUsd(p.costUSD)}</span>
            <span className="r rc-muted">×{p.dispatches}</span>
            <div className="rc-track">
              <div style={{ width: `${(100 * p.costUSD) / max}%` }} />
            </div>
          </div>
        );
      })}
      {n > 0 && (
        <div className="rc-muted rc-unpriced">
          + {n} unpriced run{n > 1 ? "s" : ""} ({unpriced.map((p) => `${p.pair} ×${p.dispatches}`).join(", ")})
        </div>
      )}
    </div>
  );
}

function Cell({ v, max, colour }: { v: number; max: number; colour: string }) {
  if (!v) return <span className="rc-muted">—</span>;
  return (
    <div className="rc-cell">
      <span>{formatUsd(v)}</span>
      <div className="rc-track" aria-hidden="true">
        <div style={{ width: `${(100 * v) / max}%`, background: colour }} />
      </div>
    </div>
  );
}

export function RoleCosts() {
  const [win, setWin] = useState(() => windowFor("7d"));
  const [project, setProject] = useState<string | undefined>(undefined);
  const [projects, setProjects] = useState<string[]>([]);
  const [data, setData] = useState<RoleCostsResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [open, setOpen] = useState<Set<string>>(new Set());

  useEffect(() => {
    fetchProjects().then(setProjects, () => setProjects([]));
  }, []);

  useEffect(() => {
    let live = true;
    setError(null);
    fetchStatsView("role-costs", { from: win.from, to: win.to, project }).then(
      (d) => live && setData(d),
      (e: unknown) => live && setError(e instanceof Error ? e.message : String(e)),
    );
    return () => {
      live = false;
    };
  }, [win, project]);

  const toggle = (role: string) =>
    setOpen((s) => {
      const n = new Set(s);
      if (!n.delete(role)) n.add(role);
      return n;
    });

  const roles = data?.roles ?? [];
  const stages = byStage(roles);
  const total = sum(roles.map((r) => r.costUSD));
  const weeks = data?.weeks ?? [];
  const spent = (rs: Role[], w: (typeof weeks)[number]) => sum(rs.map((r) => w.roles[r.role] ?? 0));
  const weekTotals = weeks.map((w) => spent(roles, w));
  const weekMax = Math.max(0, ...weekTotals);
  const stageMax = Math.max(0, ...stages.flatMap((s) => weeks.map((w) => spent(s.roles, w))));
  const roleMax = Math.max(0, ...weeks.flatMap((w) => Object.values(w.roles)));

  return (
    <section className="dashboard lb role-costs" aria-label="Role costs">
      <div className="lb-fbar">
        <ChipGroup name="Period">
          {(Object.keys(PERIOD_DAYS) as PeriodChip[]).map((p) => (
            <Chip key={p} on={win.period === p} onClick={() => setWin(windowFor(p))}>
              {PERIOD_LABELS[p]}
            </Chip>
          ))}
        </ChipGroup>
        <ChipGroup name="Project">
          <Chip on={project === undefined} onClick={() => setProject(undefined)}>
            All
          </Chip>
          {projects.map((p) => (
            <Chip key={p} on={project === p} onClick={() => setProject(p)} title={p}>
              {projectLabel(p)}
            </Chip>
          ))}
        </ChipGroup>
      </div>
      {error && <p className="lb-muted">{error}</p>}

      <section className="lb-panel">
        <div className="lb-panel-h">
          <span>Spend and agent time per role</span>
          <span className="lb-panel-note">agent time sums every dispatch, parallel agents adding up</span>
        </div>
        <div className="lb-scroll">
          <table className="lb-table rc-table" aria-label="Spend and agent time per role">
            <thead>
              <tr>
                <th>Role</th>
                <th className="r">Dispatches</th>
                <th className="r">Cost</th>
                <th className="r">Share</th>
                <th className="r">Agent time</th>
                <th className="r">Median run</th>
                <th className="r">Cost / dispatch</th>
                <th className="r">Failed</th>
                <th title="click a row for every pair">Model · effort mix</th>
              </tr>
            </thead>
            <tbody>
              {stages.map(({ stage, roles: rs }) => {
                const c = sum(rs.map((r) => r.costUSD));
                const colour = { "--sc": STAGE_COLOUR[stage] ?? STAGE_COLOUR.Other } as React.CSSProperties;
                return (
                  <Fragment key={stage}>
                    <tr className="rc-stage" style={colour}>
                      <td>
                        <i />
                        {stage}
                      </td>
                      <td className="r lb-num">{sum(rs.map((r) => r.dispatches))}</td>
                      <td className="r lb-num">{formatUsd(c)}</td>
                      <td className="r lb-num">{pct(c, total)}</td>
                      <td className="r lb-num">{hm(sum(rs.map((r) => r.agentSeconds)) * 1000)}</td>
                      <td />
                      <td />
                      <td />
                      <td />
                    </tr>
                    {rs.map((r) => (
                      <Fragment key={r.role}>
                        <tr className="rc-role" style={colour} onClick={() => toggle(r.role)} aria-expanded={open.has(r.role)}>
                          <td>
                            <RoleName role={r.role} />
                          </td>
                          <td className="r lb-num">{r.dispatches}</td>
                          <td className="r lb-num">{formatUsd(r.costUSD)}</td>
                          <td className="r lb-num">{pct(r.costUSD, total)}</td>
                          <td className="r lb-num">{hm(r.agentSeconds * 1000)}</td>
                          <td className="r lb-num">{span(r.medianSeconds)}</td>
                          <td className="r lb-num">{r.priced > 0 ? formatUsd(r.costUSD / r.priced) : "—"}</td>
                          <td className="r lb-num">{pct(r.failed, r.dispatches)}</td>
                          <td>
                            <Mix r={r} />
                          </td>
                        </tr>
                        {open.has(r.role) && (
                          <tr className="rc-detail" style={colour}>
                            <td colSpan={8} />
                            <td>
                              <Pairs r={r} />
                            </td>
                          </tr>
                        )}
                      </Fragment>
                    ))}
                  </Fragment>
                );
              })}
            </tbody>
          </table>
        </div>
        {data && roles.length === 0 && <p className="lb-muted">No dispatches in this period.</p>}
      </section>

      <section className="lb-panel">
        <div className="lb-panel-h">
          <span>Spend per week by stage</span>
          <span className="lb-panel-note">each bar split by pipeline stage · hover a segment for its amount</span>
        </div>
        <div className="rc-legend">
          {stages.map(({ stage }) => (
            <span key={stage}>
              <i style={{ background: STAGE_COLOUR[stage] ?? STAGE_COLOUR.Other }} />
              {stage}
            </span>
          ))}
        </div>
        {weeks.map((w, i) => (
          <div key={w.week} className="rc-week" role="group" aria-label={`Week of ${w.week}, ${formatUsd(weekTotals[i])}`}>
            <span className="lb-num">Week of {w.week.slice(5)}</span>
            <div className="rc-weekbar" style={{ width: `${(100 * weekTotals[i]) / weekMax}%` }}>
              {stages.map(({ stage, roles: rs }) => {
                const v = spent(rs, w);
                if (!v) return null;
                const share = (100 * v) / weekTotals[i];
                return (
                  <span
                    key={stage}
                    data-testid="stage-seg"
                    title={`${stage} ${formatUsd(v)}`}
                    style={{ width: `${share}%`, background: STAGE_COLOUR[stage] ?? STAGE_COLOUR.Other }}
                  >
                    {(share * weekTotals[i]) / weekMax >= 12 ? `${stage} ${formatUsd(v)}` : ""}
                  </span>
                );
              })}
            </div>
            <span className="r lb-num">{formatUsd(weekTotals[i])}</span>
          </div>
        ))}
        <div className="lb-scroll">
          <table className="lb-table rc-weeks" aria-label="Spend per week by stage">
            <thead>
              <tr>
                <th>Stage · role</th>
                {weeks.map((w) => (
                  <th key={w.week} className="r">
                    Week of {w.week.slice(5)}
                  </th>
                ))}
                <th className="r">Total</th>
              </tr>
            </thead>
            <tbody>
              {stages.map(({ stage, roles: rs }) => {
                const colour = STAGE_COLOUR[stage] ?? STAGE_COLOUR.Other;
                return (
                  <Fragment key={stage}>
                    <tr className="rc-wstage">
                      <td>
                        <i style={{ background: colour }} />
                        {stage}
                      </td>
                      {weeks.map((w) => (
                        <td key={w.week} className="r">
                          <Cell v={spent(rs, w)} max={stageMax} colour={colour} />
                        </td>
                      ))}
                      <td className="r">{formatUsd(sum(rs.map((r) => r.costUSD)))}</td>
                    </tr>
                    {rs.map((r) => (
                      <tr key={r.role} className="rc-wrole">
                        <td>
                          <RoleName role={r.role} />
                        </td>
                        {weeks.map((w) => (
                          <td key={w.week} className="r">
                            <Cell v={w.roles[r.role] ?? 0} max={roleMax} colour={`${colour}99`} />
                          </td>
                        ))}
                        <td className="r">{formatUsd(r.costUSD)}</td>
                      </tr>
                    ))}
                  </Fragment>
                );
              })}
              <tr className="rc-wstage">
                <td>Total</td>
                {weekTotals.map((v, i) => (
                  <td key={weeks[i].week} className="r">
                    {formatUsd(v)}
                  </td>
                ))}
                <td className="r">{formatUsd(total)}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </section>
  );
}
