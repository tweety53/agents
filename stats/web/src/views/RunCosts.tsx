// RunCosts is the across-runs Run costs view at #/run-costs (KAN-1132, the
// prototype's view C): the Period and Project chips, the category table
// and the runs table from GET /api/v1/stats/run-costs, Compare joining two
// ticked runs' one-run results (getRunTimings) client-side, and the
// Definitions those results carry.
import { useEffect, useRef, useState } from "react";
import { fetchProjects, fetchStatsView, getRunTimings, type RunCostsResult, type RunTimings } from "../api";
import { Chip, ChipGroup } from "../components/Chip";
import { hm } from "../components/RunTimeline";
import { PERIOD_DAYS, windowFor, type PeriodChip } from "../format";
import { projectLabel } from "../lib/projectLabel";

const usd = (v: number) => `$${v.toFixed(2)}`;
const PERIOD_LABELS: Record<PeriodChip, string> = { "7d": "7 days", "30d": "30 days", "90d": "90 days" };
// The breakdown rows that are not categories (runtimings.go's labelParent, labelIdle).
const NOT_CATEGORIES = ["parent commands", "idle / parent"];

type Run = RunCostsResult["runs"][number];
const runId = (r: Pick<Run, "project" | "change">) => `${r.project}/${r.change}`;

// compareLines is the prototype's compare(): each metric of A and B, then
// one cost row per category either run carries, A's order first.
function compareLines(a: RunTimings, b: RunTimings): [string, number, number, (v: number) => string][] {
  const tasks = (t: RunTimings) => t.costPerTask.length;
  const fix = (t: RunTimings) => t.counts.fixRounds.value.verify + t.counts.fixRounds.value.panel;
  const cats = (t: RunTimings) => t.breakdowns.category.rows.filter((r) => !NOT_CATEGORIES.includes(r.label));
  const cost = (t: RunTimings, label: string) => cats(t).find((r) => r.label === label)?.cost ?? 0;
  const labels = [...new Set([...cats(a), ...cats(b)].map((r) => r.label))];
  return [
    ["Wall clock", a.totals.wallClockMs.value, b.totals.wallClockMs.value, hm],
    ["Agent time", a.totals.agentTimeMs.value, b.totals.agentTimeMs.value, hm],
    ["Idle / parent", a.totals.idleMs.value, b.totals.idleMs.value, hm],
    ["Cost", a.totals.cost.value, b.totals.cost.value, usd],
    ["Tasks", tasks(a), tasks(b), String],
    ["Fix rounds", fix(a), fix(b), String],
    ...labels.map((l): [string, number, number, (v: number) => string] => [`${l} · cost`, cost(a, l), cost(b, l), usd]),
  ];
}

export function RunCosts() {
  const [win, setWin] = useState(() => windowFor("30d"));
  const [project, setProject] = useState<string | undefined>(undefined);
  const [projects, setProjects] = useState<string[]>([]);
  const [data, setData] = useState<RunCostsResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  // The two compared runs, older pick first; ticking a third drops the older.
  const [picks, setPicks] = useState<string[]>([]);
  const [timings, setTimings] = useState<Record<string, RunTimings>>({});
  // gen counts period/project loads; a timings request outlived by one is stale.
  const gen = useRef(0);

  useEffect(() => {
    fetchProjects().then(setProjects, () => setProjects([]));
  }, []);

  useEffect(() => {
    let live = true;
    gen.current++;
    setError(null);
    fetchStatsView("run-costs", { from: win.from, to: win.to, project }).then(
      (d) => {
        if (!live) return;
        setData(d);
        setPicks(d.runs.slice(0, 2).map(runId));
      },
      (e: unknown) => live && setError(e instanceof Error ? e.message : String(e)),
    );
    return () => {
      live = false;
    };
  }, [win, project]);

  // Each picked run is fetched once: requested holds the ids asked for, and a
  // result or error that lands after the period or project changed is dropped
  // (its id released so the next pick asks again).
  const requested = useRef(new Set<string>());
  useEffect(() => {
    for (const id of picks) {
      const r = data?.runs.find((x) => runId(x) === id);
      if (!r || requested.current.has(id)) continue;
      requested.current.add(id);
      const g = gen.current;
      const stale = () => gen.current !== g && requested.current.delete(id);
      getRunTimings(r.project, r.change).then(
        (t) => stale() || setTimings((prev) => ({ ...prev, [id]: t })),
        (e: unknown) => {
          if (stale()) return;
          requested.current.delete(id);
          setError(e instanceof Error ? e.message : String(e));
        },
      );
    }
  }, [picks, data]);

  const tick = (id: string) => setPicks((p) => (p.includes(id) ? p.filter((x) => x !== id) : [...p, id].slice(-2)));
  const pick = (slot: 0 | 1, id: string) => setPicks((p) => (slot === 0 ? [id, p[1]] : [p[0], id]));
  const [a, b] = [timings[picks[0]], timings[picks[1]]];
  const definitions = (a ?? b)?.definitions ?? [];
  const runs = data?.runs ?? [];

  return (
    <section className="dashboard lb" aria-label="Run costs">
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

      <section className="lb-panel" aria-label="Time and cost by category, across runs">
        <div className="lb-panel-h">
          <span>Time and cost by category</span>
          <span className="lb-panel-note">{runs.length} finished runs · union per run, summed across runs</span>
        </div>
        <div className="lb-scroll">
          <table className="lb-table">
            <thead>
              <tr>
                <th>Category</th>
                <th className="r">Runs</th>
                <th className="r">Rows</th>
                <th className="r">Time (union)</th>
                <th className="r">Cost</th>
                <th className="r">Share of cost</th>
              </tr>
            </thead>
            <tbody>
              {(data?.categories ?? []).map((c) => (
                <tr key={c.category}>
                  <td>{c.category}</td>
                  <td className="r lb-num">{c.runs}</td>
                  <td className="r lb-num">{c.rows}</td>
                  <td className="r lb-num">{hm(c.seconds * 1000)}</td>
                  <td className="r lb-num">{usd(c.cost)}</td>
                  <td className="r lb-num">{`${(100 * c.shareOfCost).toFixed(1)}%`}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className="lb-panel" aria-label="Runs">
        <div className="lb-panel-h">
          <span>Runs</span>
          <span className="lb-panel-note">tick two to compare</span>
        </div>
        <div className="lb-scroll">
          <table className="lb-table">
            <thead>
              <tr>
                <th></th>
                <th>Run</th>
                <th>Project</th>
                <th className="r">Wall clock</th>
                <th className="r">Agent time</th>
                <th className="r">Cost</th>
                <th className="r">Tasks</th>
                <th className="r">Fix rounds</th>
              </tr>
            </thead>
            <tbody>
              {runs.map((r) => (
                <tr key={runId(r)}>
                  <td>
                    <input type="checkbox" checked={picks.includes(runId(r))} onChange={() => tick(runId(r))} aria-label={`Compare ${r.change}`} />
                  </td>
                  <td>
                    <a href={`#/run-costs/${r.project}/${r.change}`}>{r.change}</a>
                  </td>
                  <td>{projectLabel(r.project)}</td>
                  <td className="r lb-num">{hm(r.wallClockMs.value)}</td>
                  <td className="r lb-num">{hm(r.agentTimeMs.value)}</td>
                  <td className="r lb-num">{usd(r.cost.value)}</td>
                  <td className="r lb-num">{r.tasks.value}</td>
                  <td className="r lb-num">{r.fixRounds.value.verify + r.fixRounds.value.panel}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className="lb-panel" aria-label="Compare runs">
        <div className="lb-panel-h">
          <span>Compare runs</span>
          <span className="lb-panel-note">B − A</span>
        </div>
        <div className="lb-panel-b cmp-pick">
          {(["A", "B"] as const).map((name, slot) => (
            <label key={name}>
              {name}{" "}
              <select value={picks[slot] ?? ""} onChange={(e) => pick(slot as 0 | 1, e.target.value)}>
                {runs.map((r) => (
                  <option key={runId(r)} value={runId(r)}>
                    {r.change}
                  </option>
                ))}
              </select>
            </label>
          ))}
        </div>
        <div className="lb-scroll">
          <table className="lb-table">
            <thead>
              <tr>
                <th>Metric</th>
                <th className="r">A</th>
                <th className="r">B</th>
                <th className="r">Δ</th>
                <th className="r">Δ %</th>
              </tr>
            </thead>
            <tbody>
              {a &&
                b &&
                compareLines(a, b).map(([n, x, y, f]) => {
                  const d = y - x;
                  const cls = d > 0 ? "delta-up" : d < 0 ? "delta-down" : "lb-muted";
                  const sign = d > 0 ? "+" : d < 0 ? "−" : "";
                  return (
                    <tr key={n}>
                      <td>{n}</td>
                      <td className="r lb-num">{f(x)}</td>
                      <td className="r lb-num">{f(y)}</td>
                      <td className={`r lb-num ${cls}`}>{`${sign}${f(Math.abs(d))}`}</td>
                      <td className={`r lb-num ${cls}`}>{x ? `${sign}${Math.abs((100 * d) / x).toFixed(0)}%` : "—"}</td>
                    </tr>
                  );
                })}
            </tbody>
          </table>
        </div>
      </section>

      <section className="lb-panel" aria-label="Definitions">
        <div className="lb-panel-h">
          <span>Definitions</span>
        </div>
        <div className="lb-panel-b">
          <dl className="defs">
            {definitions.map((d) => [
              <dt key={`t-${d.name}`}>
                {d.name} <span className={`src ${d.kind}`}>{d.kind}</span>
              </dt>,
              <dd key={`d-${d.name}`}>{d.definition}</dd>,
            ])}
          </dl>
        </div>
      </section>
    </section>
  );
}
