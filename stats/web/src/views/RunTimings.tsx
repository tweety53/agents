// RunTimings is one run's timings and costs page (KAN-1132, the
// prototype's view B), opened from run detail's "Timings and costs →"
// button at #/run-costs/<project>/<change>. Every figure is the server's
// (GET /api/v1/stats/run/{project}/{change}); each carries its definition
// and kind as a `src` badge.
import { useEffect, useState, type ReactNode } from "react";
import { fetchStatsView, getRunTimings, type RunMetric, type RunTimings as RunTimingsDTO, type StateBoardRow } from "../api";
import { Chip } from "../components/Chip";
import { RunTimeline, hm } from "../components/RunTimeline";
import { projectLabel } from "../lib/projectLabel";
import { StatePill } from "./StateBoard";

const usd = (v: number) => `$${v.toFixed(2)}`;
const pct = (v: number) => `${Math.round(v * 100)}%`;
const stamp = (iso: string) =>
  new Date(iso).toLocaleString("en-US", { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit", hourCycle: "h23" }).replace(",", "");

const GROUPS = [
  ["category", "Category"],
  ["role", "Role"],
  ["model", "Model / effort"],
  ["repo", "Repo"],
  ["round", "Fix round"],
] as const;
type Group = (typeof GROUPS)[number][0];

// The rows the prototype always draws; any other gap shows only when present.
const DQ_ALWAYS = new Set(["open-stage-runs", "unlinked", "no-tokens", "aborted-ran", "uncategorised", "journal", "retimed"]);
const NO_COST = new Set(["parent commands", "idle / parent"]);
const FRACTION: Record<number, string> = { 2: "½", 3: "⅓", 4: "¼" };

function Src({ kind, title }: { kind: string; title?: string }) {
  return (
    <span className={`src ${kind}`} title={title}>
      {kind}
    </span>
  );
}

function Lbl({ name, m }: { name: string; m: RunMetric<unknown> }) {
  return (
    <span className="lb-lbl">
      {name}
      <Src kind={m.kind} title={m.definition} />
    </span>
  );
}

// Kpi is one labelled figure: a totals-strip tile, or (bare) a counts-panel cell.
function Kpi({ name, m, value, hint, bare }: { name: string; m: RunMetric<unknown>; value: ReactNode; hint?: ReactNode; bare?: boolean }) {
  return (
    <div className={bare ? undefined : "lb-kpi"}>
      <Lbl name={name} m={m} />
      <span className="lb-kpi-val lb-num">{value}</span>
      {hint !== undefined && <span className="lb-hint">{hint}</span>}
    </div>
  );
}

function Body({ run }: { run: RunTimingsDTO }) {
  const [group, setGroup] = useState<Group>("category");
  const t = run.totals;
  const wall = t.wallClockMs.value;
  const share = (ms: number) => (wall ? ms / wall : 0);
  const laneSum = run.lanes.filter((l) => l.kind === "category").reduce((a, l) => a + l.intervals.reduce((b, iv) => b + Date.parse(iv.to) - Date.parse(iv.from), 0), 0);
  const vvCost = run.costPerVerifyRound.filter((r) => r.round.startsWith("visual-verify")).reduce((a, r) => a + r.cost, 0);

  const b = run.breakdowns[group];
  const col = (name: string) => b.columns.find((c) => c.name === name);
  const max = Math.max(0, ...b.rows.filter((r) => !NO_COST.has(r.label)).map((r) => r.cost));

  const c = run.counts;
  const deps = c.departures.value.map((d) => d.departures);
  const tasks = [...run.costPerTask].sort((x, y) => y.cost - x.cost).slice(0, 8);
  const dq = run.dataQuality.filter((d) => DQ_ALWAYS.has(d.id) || d.count > 0 || d.unknown);

  return (
    <>
      <section className="lb-panel" aria-label="Run totals">
        <div className="lb-kpis">
          <Kpi name="Wall clock" m={t.wallClockMs} value={hm(wall)} hint={`${stamp(run.from)} → ${stamp(run.to)}`} />
          <Kpi name="Agent time" m={t.agentTimeMs} value={hm(t.agentTimeMs.value)} hint={`sum ${hm(laneSum)} · ${pct(share(t.agentTimeMs.value))} of wall`} />
          <Kpi name="Parent commands" m={t.parentCommandsMs} value={hm(t.parentCommandsMs.value)} hint={`${t.parentCommandRuns} runs · ${pct(share(t.parentCommandsMs.value))}`} />
          <Kpi name="Idle / parent" m={t.idleMs} value={hm(t.idleMs.value)} hint={`${pct(share(t.idleMs.value))} of wall`} />
          <Kpi name="Cost" m={t.cost} value={usd(t.cost.value)} hint={`${t.dispatches} dispatches · ${t.noTokens} unpriced`} />
          <Kpi name="Visual-verify loops" m={t.visualVerifyShare} value={pct(t.visualVerifyShare.value)} hint={`${hm(t.visualVerifyShare.value * wall)} · ${usd(vvCost)}`} />
        </div>
      </section>

      <section className="lb-panel" aria-label="Timeline">
        <div className="lb-panel-h">
          <span>Timeline</span>
          <span className="lb-legend">
            <span><i className="lb-sw completed" />completed</span>
            <span><i className="lb-sw failed" />failed / aborted</span>
            <span><i className="lb-sw open" />open</span>
            <span><i className="lb-sw idle" />idle / parent</span>
          </span>
        </div>
        <div className="lb-panel-b lb-scroll lb-timeline">
          <RunTimeline run={run} />
        </div>
      </section>

      <div className="lb-2">
        <section className="lb-panel" aria-label="Time and cost by">
          <div className="lb-panel-h">
            <span>Time and cost by</span>
            <span className="lb-fgroup" role="group" aria-label="Group by">
              {GROUPS.map(([g, name]) => (
                <Chip key={g} on={group === g} onClick={() => setGroup(g)}>
                  {name}
                </Chip>
              ))}
            </span>
          </div>
          <div className="lb-scroll">
            <table className="lb-table">
              <thead>
                <tr>
                  <th>{GROUPS.find(([g]) => g === group)![1]}</th>
                  <th className="r">Rows</th>
                  <th className="r" title={col("Time share")?.definition}>Time share<Src kind={col("Time share")?.kind ?? "derived"} /></th>
                  <th className="r" title={col("Busy")?.definition}>Busy<Src kind={col("Busy")?.kind ?? "derived"} /></th>
                  <th className="r" title={col("Cost")?.definition}>Cost<Src kind={col("Cost")?.kind ?? "measured"} /></th>
                  <th className="r">No tokens</th>
                </tr>
              </thead>
              <tbody>
                {b.rows.map((r) => (
                  <tr key={r.label}>
                    <td>{r.label}</td>
                    <td className="r lb-num">{r.rows}</td>
                    <td className="r lb-num">{hm(r.shareMs)}</td>
                    <td className="r lb-num lb-muted">{hm(r.busyMs)}</td>
                    <td className="r">
                      {NO_COST.has(r.label) ? (
                        <span className="lb-muted">—</span>
                      ) : (
                        <span className="lb-costcell">
                          <span className="lb-num">{usd(r.cost)}</span>
                          <span className="lb-costbar"><span style={{ width: `${max ? (100 * r.cost) / max : 0}%` }} /></span>
                        </span>
                      )}
                    </td>
                    <td className={`r lb-num ${r.noTokens ? "lb-err" : "lb-muted"}`}>{r.noTokens}</td>
                  </tr>
                ))}
              </tbody>
              <tfoot>
                <tr>
                  <td>Total</td>
                  <td className="r lb-num">{b.total.rows}</td>
                  <td className="r lb-num">{hm(b.total.shareMs)}</td>
                  <td className="r lb-num lb-muted">—</td>
                  <td className="r lb-num">{usd(b.total.cost)}</td>
                  <td className="r lb-num">{b.total.noTokens}</td>
                </tr>
              </tfoot>
            </table>
          </div>
        </section>

        <section className="lb-panel dq" aria-label="Data quality">
          <div className="lb-panel-h">
            <span>Data quality</span>
            <span className="lb-panel-note">every gap still present in this run</span>
          </div>
          <div className="lb-panel-b">
            <ul>
              {dq.map((d) => (
                <li key={d.id}>
                  <span className={`n lb-num ${d.count || d.unknown ? "lb-err" : "ok"}`}>{d.unknown ? "?" : d.count}</span>
                  <span>
                    {d.detail.length ? (
                      <details>
                        <summary>{d.label}</summary>
                        <ul>
                          {d.detail.map((x) => (
                            <li key={x}>{x}</li>
                          ))}
                        </ul>
                      </details>
                    ) : (
                      d.label
                    )}
                  </span>
                </li>
              ))}
            </ul>
          </div>
        </section>
      </div>

      <section className="lb-panel" aria-label="Counts">
        <div className="lb-panel-h"><span>Counts</span></div>
        <div className="counts">
          <Kpi bare name="Fix rounds" m={c.fixRounds} value={`${c.fixRounds.value.verify} · ${c.fixRounds.value.panel}`} hint="verify / visual · panel" />
          <Kpi bare name="Departures / visual round" m={c.departures} value={deps.length ? `${deps[0]} → ${deps[deps.length - 1]}` : "—"} hint={deps.join(" · ")} />
          <Kpi bare name="Findings" m={c.findings} value={`${c.findings.value.critical} · ${c.findings.value.important} · ${c.findings.value.minor}`} hint="critical · important · minor" />
          <Kpi bare name="Flaky retries" m={c.flakyRetries} value={c.flakyRetries.value.retries} hint={`over ${c.flakyRetries.value.runs} verify runs`} />
          <Kpi bare name="Rows open · aborted · blocked" m={c.rows} value={`${c.rows.value.open} · ${c.rows.value.aborted} · ${c.rows.value.blocked}`} />
          <Kpi bare name="Journal pending" m={c.journalPending} value={c.journalPending.value} />
        </div>
      </section>

      <div className="lb-2">
        <section className="lb-panel" aria-label="Cost per task">
          <div className="lb-panel-h">
            <span>
              Cost per task
              <Src kind="derived" title="Every dispatch whose key names the task (implementer, fixer, reviewer); a group key's cost is split evenly across its tasks and marked estimated." />
            </span>
            <span className="lb-panel-note">{`top ${tasks.length} of ${run.costPerTask.length}`}</span>
          </div>
          <div className="lb-scroll">
            <table className="lb-table">
              <thead>
                <tr><th>Task</th><th className="r">Rows</th><th className="r">Time</th><th className="r">Cost</th></tr>
              </thead>
              <tbody>
                {tasks.map((k) => {
                  const n = k.group ? k.group.split("+").length : 1;
                  return (
                    <tr key={k.task}>
                      <td>{k.group ? `${k.task} · ${FRACTION[n] ?? `1/${n}`} of group ${k.group}` : k.task}</td>
                      <td className="r lb-num">{k.rows}</td>
                      <td className="r lb-num">{hm(k.shareMs)}</td>
                      <td className="r lb-num">
                        {usd(k.cost)}
                        {k.estimated && <> <Src kind="estimated" title="A group dispatch’s cost split evenly across its tasks." /></>}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </section>
        <section className="lb-panel" aria-label="Cost per verify round">
          <div className="lb-panel-h">
            <span>
              Cost per verify round
              <Src kind="derived" title="A round's verify rows plus the in-run fix tasks whose Opener names that round, with their gated reviewers." />
            </span>
            <span className="lb-panel-note">{`${run.costPerVerifyRound.length} rounds`}</span>
          </div>
          <div className="lb-scroll">
            <table className="lb-table">
              <thead>
                <tr><th>Round</th><th className="r">Opener</th><th className="r">Departures</th><th className="r">Time</th><th className="r">Cost</th></tr>
              </thead>
              <tbody>
                {run.costPerVerifyRound.map((r) => {
                  const dep = c.departures.value.find((d) => d.round === r.round || d.round.startsWith(`${r.round}-`));
                  return (
                    <tr key={r.round}>
                      <td>{/fix-(\d+)/.exec(r.round)?.[1] ?? r.round}</td>
                      <td className="r lb-num">{r.round}</td>
                      <td className="r lb-num">{dep ? dep.departures : "—"}</td>
                      <td className="r lb-num">{hm(r.shareMs)}</td>
                      <td className="r lb-num">{usd(r.cost)}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </section>
      </div>
    </>
  );
}

export function RunTimings({ project, change }: { project: string; change: string }) {
  const [state, setState] = useState<{ run?: RunTimingsDTO; error?: string }>({});
  useEffect(() => {
    const ctl = new AbortController();
    setState({});
    getRunTimings(project, change, { signal: ctl.signal }).then(
      (run) => setState({ run }),
      (e: unknown) => {
        if (!ctl.signal.aborted) setState({ error: e instanceof Error ? e.message : String(e) });
      },
    );
    return () => ctl.abort();
  }, [project, change]);
  // The heading's state pill, as run detail fetches it; a failed fetch leaves the heading without it.
  const [board, setBoard] = useState<StateBoardRow>();
  useEffect(() => {
    const ctl = new AbortController();
    setBoard(undefined);
    fetchStatsView<StateBoardRow[]>("state-board", { from: new Date(0), to: new Date(Date.now() + 86_400_000), project, change }, { signal: ctl.signal }).then(
      (r) => setBoard(r.rows.find((x) => x.name === change)),
      () => undefined,
    );
    return () => ctl.abort();
  }, [project, change]);

  return (
    <section aria-label={`Run timings and costs: ${project}/${change}`} className="dashboard lb lb-run">
      <div className="lb-crumb">
        <a href={`#/run/${encodeURIComponent(project)}/${encodeURIComponent(change)}`}>← Run detail</a>
      </div>
      <div className="lb-runhead">
        <div className="lb-title">
          <h2>{`${change} · timings and costs`}</h2>
          {board && <StatePill row={board} />}
        </div>
        <p className="lb-sub">
          Project {projectLabel(project)} · <a href="#/run-costs">Compare with another run →</a>
        </p>
      </div>
      {state.run ? (
        <Body run={state.run} />
      ) : state.error ? (
        <p role="alert" className="view-error">{state.error}</p>
      ) : (
        <p role="status">Loading…</p>
      )}
    </section>
  );
}
