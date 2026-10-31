// Stage leaderboard -- "Which stages take the most of the spend, and which
// are growing?" (decision leaderboard-spend-and-growth), built as the
// approved prototype draws it.
//
// The view owns its filters: a KAN-1048 chip bar of Period (7d / 30d / 90d,
// a rolling window ending when the chip is chosen), Project and Command,
// replacing the shared dashboard bar, so App's period and project props are
// ignored (leaderboard-chip-filters). Project chips come from the response's
// `projects` -- only projects with spend, by readable name and spend
// (leaderboard-project-chips-with-spend). The Project chip reaches the
// request; the Command chip filters the returned rows client-side.
//
// Two requests -- the chosen window and the equal window before it -- are
// joined on command/stage for each stage's growth
// (leaderboard-previous-window-call). Their bounds are computed when a Period
// chip is chosen, never per render: a per-render `new Date()` would refetch
// forever.
//
// Spend over Biggest riser beside Top 3 (leaderboard-tile-layout); a plain
// ranked `lb-table` of the top ten plus an "Other N stages" row, unsorted
// (leaderboard-plain-ranked-table), with the prototype's columns
// (leaderboard-columns); every stage named by stageLabel with its id beneath
// (leaderboard-readable-names); Chip is shared with the state board
// (leaderboard-shared-chip).
import { useState, type ReactNode } from "react";
import type { StageLeaderboardRow } from "../api";
import { Chip } from "../components/Chip";
import { Panel } from "../components/Panel";
import { DAY_MS } from "../format";
import { useStatsView } from "../hooks/useStatsView";
import { projectLabel } from "../lib/projectLabel";
import { stageLabel } from "../lib/stageLabel";
import type { ViewProps } from "../viewTypes";

const PERIOD_DAYS = { "7d": 7, "30d": 30, "90d": 90 } as const;
type PeriodChip = keyof typeof PERIOD_DAYS;

interface Filters {
  period: PeriodChip;
  project: string | null;
  command: string | null;
  from: Date;
  to: Date;
}

// The window ending now for a Period chip, computed only when one is chosen.
function windowFor(period: PeriodChip): { period: PeriodChip; from: Date; to: Date } {
  const to = Date.now();
  return { period, from: new Date(to - PERIOD_DAYS[period] * DAY_MS), to: new Date(to) };
}

const defaults = (): Filters => ({ ...windowFor("30d"), project: null, command: null });

interface Ranked {
  row: StageLeaderboardRow;
  prev: number | null;
}

const usd = (v: number) => (v >= 100 ? "$" + Math.round(v).toLocaleString("en-US") : "$" + v.toFixed(2));
const pct = (part: number, whole: number) => `${Math.round((part / whole) * 100)}%`;
const sum = (xs: number[]) => xs.reduce((a, x) => a + x, 0);
const key = (r: StageLeaderboardRow) => `${r.command}\u0000${r.stage}`;

// The change against the previous window; null ("new") without a previous spend.
function growth(cur: number, prev: number | null): { cls: string; text: string; percent: string } | null {
  if (prev == null || prev < 0.005) return null;
  const d = cur - prev;
  const p = Math.abs((d / prev) * 100);
  const sign = d > 0 ? "+" : "−";
  const percent = sign + (p >= 1000 ? `${Math.round(p / 100) / 10}k` : `${Math.round(p)}`) + "%";
  return { cls: d > 0 ? "up" : "down", text: `${sign}${usd(Math.abs(d))} (${percent})`, percent };
}

function Delta({ cur, prev }: { cur: number; prev: number | null }) {
  const g = growth(cur, prev);
  return g ? <span className={g.cls}>{g.text}</span> : <span className="lb-muted">new</span>;
}

function StageCell({ stage, command }: { stage: string; command?: string }) {
  const label = stageLabel(stage);
  return (
    <>
      <div className="sl-name">
        {label}
        {command && (
          <>
            {" "}
            <span className="sl-cmd">{command}</span>
          </>
        )}
      </div>
      {label !== stage && <div className="sl-id">{stage}</div>}
    </>
  );
}

function Group({ name, children }: { name: string; children: ReactNode }) {
  return (
    <span className="lb-fgroup" role="group" aria-label={name}>
      <span className="lb-fl">{name}</span>
      {children}
    </span>
  );
}

export function StageLeaderboard(_props: ViewProps) {
  const [f, setF] = useState<Filters>(defaults);
  const project = f.project ?? undefined;
  const length = f.to.getTime() - f.from.getTime();
  const cur = useStatsView<StageLeaderboardRow[]>("stage-leaderboard", { from: f.from, to: f.to, project });
  const prev = useStatsView<StageLeaderboardRow[]>("stage-leaderboard", { from: new Date(f.from.getTime() - length), to: f.from, project });

  const offDefaults = f.period !== "30d" || f.project !== null || f.command !== null;
  const ready = cur.status === "ready" && prev.status === "ready" && cur.data.recorded && !cur.data.unmeasured;
  const all = cur.status === "ready" ? cur.data.rows : [];
  const prevRows = prev.status === "ready" ? prev.data.rows : [];
  const projects = cur.status === "ready" ? (cur.data.projects ?? []) : [];
  const byCommand = (r: StageLeaderboardRow) => f.command === null || r.command === f.command;

  const cmdSpend = new Map<string, number>();
  for (const r of all) cmdSpend.set(r.command, (cmdSpend.get(r.command) ?? 0) + r.totalCostUsd);
  const prevByKey = new Map(prevRows.map((r) => [key(r), r.totalCostUsd]));
  const rs: Ranked[] = all.filter(byCommand).map((row) => ({ row, prev: prevByKey.get(key(row)) ?? null }));

  const total = sum(rs.map((r) => r.row.totalCostUsd));
  const prevTotal = sum(prevRows.filter(byCommand).map((r) => r.totalCostUsd));
  const top3 = rs.slice(0, 3);
  const riser = rs
    .flatMap((r) => (r.prev == null ? [] : [{ ...r, prev: r.prev, d: r.row.totalCostUsd - r.prev }]))
    .sort((a, b) => b.d - a.d)[0];
  const grew = riser !== undefined && riser.d > 0 ? riser : undefined;
  const cmdNote = (r: StageLeaderboardRow) => (f.command === null ? r.command : undefined);
  const top = rs.slice(0, 10);
  const rest = rs.slice(10);
  const restTotal = sum(rest.map((r) => r.row.totalCostUsd));
  const max = top.length ? top[0].row.totalCostUsd : 1;
  const bar = (t: number) => (
    <span className="lb-costcell">
      <span className="lb-costbar">
        <span style={{ width: `${((t / max) * 100).toFixed(1)}%` }} />
      </span>
      <span className="sl-amt">{usd(t)}</span>
    </span>
  );

  return (
    <section aria-label="Stage leaderboard" className="dashboard lb sl">
      <div>
        <h2>Stage leaderboard</h2>
        <p className="lb-sub">Which stages take the most of the spend, and which are growing?</p>
      </div>
      <div className="lb-fbar" role="toolbar" aria-label="Filters">
        <Group name="Period">
          {(Object.keys(PERIOD_DAYS) as PeriodChip[]).map((p) => (
            <Chip key={p} on={f.period === p} onClick={() => setF({ ...f, ...windowFor(p) })}>
              {p}
            </Chip>
          ))}
        </Group>
        <Group name="Project">
          <Chip on={f.project === null} onClick={() => setF({ ...f, project: null })}>
            All
          </Chip>
          {projects.map((p) => (
            <Chip key={p.projectKey} on={f.project === p.projectKey} onClick={() => setF({ ...f, project: p.projectKey })} title={p.projectKey}>
              {projectLabel(p.projectKey)}
              <span className="lb-c">{usd(p.totalCostUsd)}</span>
            </Chip>
          ))}
        </Group>
        <Group name="Command">
          <Chip on={f.command === null} onClick={() => setF({ ...f, command: null })}>
            All
          </Chip>
          {[...cmdSpend.entries()]
            .sort((a, b) => b[1] - a[1])
            .map(([c, t]) => (
              <Chip key={c} on={f.command === c} onClick={() => setF({ ...f, command: c })}>
                {c}
                <span className="lb-c">{usd(t)}</span>
              </Chip>
            ))}
        </Group>
        {offDefaults && (
          <button type="button" className="lb-fclear" onClick={() => setF(defaults())}>
            Clear filters
          </button>
        )}
      </div>
      {ready && (
        <div className="sl-tiles">
          <div className="lb-tile sl-t-spend">
            <span className="lb-lbl">Spend</span>
            <span className="lb-val">{usd(total)}</span>
            <span className="sl-note">
              <Delta cur={total} prev={prevTotal} /> vs the previous {f.period}
            </span>
          </div>
          <div className="lb-tile sl-t-top">
            <span className="lb-lbl">Top 3 stages</span>
            <span className="lb-val">{total ? pct(sum(top3.map((r) => r.row.totalCostUsd)), total) : "0%"}</span>
            <span className="sl-note">of the spend</span>
            <ol className="sl-list">
              {top3.map((r, i) => (
                <li key={key(r.row)}>
                  <span className="sl-rank">{i + 1}</span>
                  <div>
                    <StageCell stage={r.row.stage} command={cmdNote(r.row)} />
                  </div>
                  <span className="sl-num">{usd(r.row.totalCostUsd)}</span>
                  <span className="sl-num lb-muted">{pct(r.row.totalCostUsd, total)}</span>
                </li>
              ))}
            </ol>
          </div>
          <div className="lb-tile sl-t-riser">
            <span className="lb-lbl">Biggest riser</span>
            <span className={grew ? "lb-val up" : "lb-val"}>{grew ? "+" + usd(grew.d) : "—"}</span>
            {grew ? (
              <div className="sl-riser">
                <div>
                  <StageCell stage={grew.row.stage} command={cmdNote(grew.row)} />
                </div>
                <span className="sl-note">
                  {`${usd(grew.prev)} → ${usd(grew.row.totalCostUsd)} vs the previous ${f.period} (${growth(grew.row.totalCostUsd, grew.prev)?.percent ?? "new"})`}
                </span>
              </div>
            ) : (
              <span className="sl-note">no stage grew vs the previous {f.period}</span>
            )}
          </div>
        </div>
      )}
      <div className="lb-panel">
        <div className="lb-panel-h">
          <span>Top 10 stages by spend</span>
          <span className="lb-panel-note">{ready && rs.length ? `${rs.length} stages ran` : ""}</span>
        </div>
        <div className="lb-scroll">
          {!ready ? (
            // Loading, error, not-recorded and unmeasured read as Panel renders them.
            <Panel title="Top 10 stages by spend" state={cur.status === "ready" && prev.status !== "ready" ? prev : cur}>
              {() => null}
            </Panel>
          ) : !rs.length ? (
            <div className="sl-empty">No costed stage runs in this period.</div>
          ) : (
            <table className="lb-table">
              <thead>
                <tr>
                  <th>#</th>
                  <th>Stage</th>
                  <th>Command</th>
                  <th className="r">Spend</th>
                  <th className="r">Share</th>
                  <th className="r">vs previous {f.period}</th>
                  <th className="r">Runs</th>
                  <th className="r">Median run</th>
                  <th className="r">P90 run</th>
                  <th className="r">Main share</th>
                </tr>
              </thead>
              <tbody>
                {top.map(({ row: r, prev: p }, i) => (
                  <tr key={key(r)}>
                    <td className="lb-muted">{i + 1}</td>
                    <td>
                      <StageCell stage={r.stage} />
                    </td>
                    <td className="sl-cmd">{r.command}</td>
                    <td className="r">{bar(r.totalCostUsd)}</td>
                    <td className="r">{pct(r.totalCostUsd, total)}</td>
                    <td className="r">
                      <Delta cur={r.totalCostUsd} prev={p} />
                    </td>
                    <td className="r">{r.runCount.toLocaleString("en-US")}</td>
                    <td className="r">{usd(r.medianCostUsd)}</td>
                    <td className="r">{usd(r.p90CostUsd)}</td>
                    <td className="r">{r.mainSharePct == null ? "—" : `${Math.round(r.mainSharePct)}%`}</td>
                  </tr>
                ))}
                {rest.length > 0 && (
                  <tr className="other">
                    <td></td>
                    <td>Other {rest.length} stages</td>
                    <td></td>
                    <td className="r">{bar(restTotal)}</td>
                    <td className="r">{pct(restTotal, total)}</td>
                    <td className="r">
                      <Delta cur={restTotal} prev={sum(rest.map((r) => r.prev ?? 0))} />
                    </td>
                    <td className="r">{sum(rest.map((r) => r.row.runCount)).toLocaleString("en-US")}</td>
                    <td className="r">—</td>
                    <td className="r">—</td>
                    <td className="r">—</td>
                  </tr>
                )}
              </tbody>
            </table>
          )}
        </div>
      </div>
    </section>
  );
}
