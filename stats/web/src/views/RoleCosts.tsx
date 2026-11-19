// RoleCosts is the Role costs view at #/role-costs: the Period and Project
// chips, one row per allowed-pairs role from GET /api/v1/stats/role-costs —
// spend, agent time, per-dispatch figures, failure rate and the model ·
// effort pairs it ran on — and the weekly spend per role.
import { useEffect, useState } from "react";
import { fetchProjects, fetchStatsView, type RoleCostsResult } from "../api";
import { Chip, ChipGroup } from "../components/Chip";
import { formatUsd, hm, PERIOD_DAYS, windowFor, type PeriodChip } from "../format";
import { projectLabel } from "../lib/projectLabel";

const PERIOD_LABELS: Record<PeriodChip, string> = { "7d": "7 days", "30d": "30 days", "90d": "90 days" };

const span = (s: number) => (s < 60 ? `${Math.round(s)}s` : s < 3600 ? `${Math.round(s / 60)}m` : hm(s * 1000));
const pct = (n: number, d: number) => (d > 0 ? `${Math.round((100 * n) / d)}%` : "—");
const colour = (i: number) => `hsl(${(i * 47) % 360} 55% 55%)`;

export function RoleCosts() {
  const [win, setWin] = useState(() => windowFor("30d"));
  const [project, setProject] = useState<string | undefined>(undefined);
  const [projects, setProjects] = useState<string[]>([]);
  const [data, setData] = useState<RoleCostsResult | null>(null);
  const [error, setError] = useState<string | null>(null);

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

  const roles = data?.roles ?? [];
  const total = roles.reduce((t, r) => t + r.costUSD, 0);
  const weeks = data?.weeks ?? [];
  const weekMax = Math.max(0, ...weeks.map((w) => Object.values(w.roles).reduce((a, b) => a + b, 0)));

  return (
    <section className="dashboard lb" aria-label="Role costs">
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
          <table className="lb-table" aria-label="Spend and agent time per role">
            <thead>
              <tr>
                <th>Role</th>
                <th>Stage</th>
                <th className="r">Dispatches</th>
                <th className="r">Cost</th>
                <th className="r">Share</th>
                <th className="r">Agent time</th>
                <th className="r">Median run</th>
                <th className="r">Cost / dispatch</th>
                <th className="r">Failed</th>
                <th>Model · effort</th>
              </tr>
            </thead>
            <tbody>
              {roles.map((r) => (
                <tr key={r.role}>
                  <td className="lb-mono">{r.role}</td>
                  <td>{r.group}</td>
                  <td className="r lb-num">{r.dispatches}</td>
                  <td className="r lb-num">{formatUsd(r.costUSD)}</td>
                  <td className="r lb-num">{pct(r.costUSD, total)}</td>
                  <td className="r lb-num">{hm(r.agentSeconds * 1000)}</td>
                  <td className="r lb-num">{span(r.medianSeconds)}</td>
                  <td className="r lb-num">{r.priced > 0 ? formatUsd(r.costUSD / r.priced) : "—"}</td>
                  <td className="r lb-num">{pct(r.failed, r.dispatches)}</td>
                  <td className="role-pairs">
                    {r.pairs.map((p) => (
                      <div key={p.pair}>
                        {p.pair} {formatUsd(p.costUSD)} ×{p.dispatches}
                      </div>
                    ))}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {data && roles.length === 0 && <p className="lb-muted">No dispatches in this period.</p>}
      </section>

      <section className="lb-panel">
        <div className="lb-panel-h">
          <span>Spend per week by role</span>
        </div>
        <div className="role-legend">
          {roles.map((r, i) => (
            <span key={r.role}>
              <i style={{ background: colour(i) }} /> {r.role}
            </span>
          ))}
        </div>
        {weeks.map((w) => {
          const sum = Object.values(w.roles).reduce((a, b) => a + b, 0);
          return (
            <div key={w.week} className="role-week" role="group" aria-label={`Week of ${w.week}, ${formatUsd(sum)}`}>
              <span className="lb-num">{w.week}</span>
              <div className="role-week-bar">
                {roles.map((r, i) =>
                  w.roles[r.role] ? (
                    <span
                      key={r.role}
                      data-testid="role-seg"
                      title={`${r.role} ${formatUsd(w.roles[r.role])}`}
                      style={{ width: `${(100 * w.roles[r.role]) / weekMax}%`, background: colour(i) }}
                    />
                  ) : null,
                )}
              </div>
              <span className="r lb-num">{formatUsd(sum)}</span>
            </div>
          );
        })}
      </section>
    </section>
  );
}
