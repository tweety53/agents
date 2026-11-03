// Reviewers -- "Which reviewer slots earn their cost, find the most severe
// defects, and raise false alarms?" (decision reviewers-three-questions),
// built as the approved prototype draws it.
//
// The view owns its filters: a KAN-1048 chip bar of Period (7d / 30d / 90d,
// a rolling window computed when the chip is chosen) and Project (from the
// response's `projects`, each with its reviewer dispatch count), replacing the
// shared dashboard bar, so App's period, project and model props are ignored
// (reviewers-chip-filters-period-project). A Rank by group re-ranks the table,
// retitles its panel and moves the `aria-sort` marker; it is a view choice, so
// Clear filters leaves it alone (reviewers-rank-by-chips). Chip state is local,
// never written to the URL (reviewers-chip-state-local).
//
// Every ratio, rank, tile and footer total is derived here from the rows:
// a weighted finding is critical 10 · important 4 · minor 1, withdrawn findings
// excluded (reviewers-withdrawn-excluded-shown); Severity ranks on weighted per
// dispatch, ties by critical count (reviewers-severity-rank-weighted); a slot
// ranks on its withdrawn share whatever its sample (reviewers-rank-small-samples-as-is);
// $ per weighted finding is priced cost per priced dispatch over weighted per
// dispatch, `—` with no priced dispatch (reviewers-unpriced-dispatch-rules).
// One tile per question, each headed by the first slot of its own rank
// (reviewers-one-tile-per-question, reviewers-tile-edge-states). Formats follow
// reviewers-display-formats; the Changes column, the model variable and
// DataTable's search and dropdowns are gone (reviewers-dropped-elements).
import { useState } from "react";
import type { ReviewerProject, ReviewerRow } from "../api";
import { Chip, ChipGroup } from "../components/Chip";
import { Panel } from "../components/Panel";
import { DAY_MS, formatUsdCompact } from "../format";
import { useStatsView } from "../hooks/useStatsView";
import { projectLabel } from "../lib/projectLabel";
import type { ViewProps } from "../viewTypes";

const PERIOD_DAYS = { "7d": 7, "30d": 30, "90d": 90 } as const;
type PeriodChip = keyof typeof PERIOD_DAYS;
const RANKS = { value: "Value", severity: "Severity", alarms: "False alarms" } as const;
type Rank = keyof typeof RANKS;
const TITLE: Record<Rank, string> = {
  value: "Every slot, best value first",
  severity: "Every slot, most severe first",
  alarms: "Every slot, most false alarms first",
};

interface Filters {
  period: PeriodChip;
  project: string | null;
  from: Date;
  to: Date;
}

// The window ending now for a Period chip, computed only when one is chosen.
function windowFor(period: PeriodChip): { period: PeriodChip; from: Date; to: Date } {
  const to = Date.now();
  return { period, from: new Date(to - PERIOD_DAYS[period] * DAY_MS), to: new Date(to) };
}

const defaults = (): Filters => ({ ...windowFor("30d"), project: null });

interface Slot {
  row: ReviewerRow;
  weighted: number;
  wpd: number;
  /** $ per weighted finding; null when nothing was found or nothing priced. */
  value: number | null;
  /** Withdrawn share of every finding; null with no finding at all. */
  share: number | null;
}

const weightedOf = (r: { critical: number; important: number; minor: number }) => 10 * r.critical + 4 * r.important + r.minor;
const sum = (xs: number[]) => xs.reduce((a, x) => a + x, 0);
const pct = (v: number) => `${Math.round(v * 100)}%`;
const int = (v: number) => v.toLocaleString("en-US");

function valueOf(costUsd: number, priced: number, weighted: number, dispatches: number): number | null {
  if (!priced || !weighted || !dispatches) return null;
  return costUsd / priced / (weighted / dispatches);
}

function derive(row: ReviewerRow): Slot {
  const weighted = weightedOf(row);
  const findings = row.critical + row.important + row.minor + row.withdrawn;
  return {
    row,
    weighted,
    wpd: row.dispatches ? weighted / row.dispatches : 0,
    value: valueOf(row.costUsd, row.pricedDispatches, weighted, row.dispatches),
    share: findings ? row.withdrawn / findings : null,
  };
}

// Nulls last, then ascending.
const nullsLast = (a: number | null, b: number | null) => (a === null ? (b === null ? 0 : 1) : b === null ? -1 : a - b);
const byName = (a: Slot, b: Slot) => a.row.slot.localeCompare(b.row.slot);
const ORDER: Record<Rank, (a: Slot, b: Slot) => number> = {
  value: (a, b) => nullsLast(a.value, b.value) || byName(a, b),
  severity: (a, b) => b.wpd - a.wpd || b.row.critical - a.row.critical || byName(a, b),
  alarms: (a, b) => nullsLast(a.share === null ? null : -a.share, b.share === null ? null : -b.share) || byName(a, b),
};
const ranked = (slots: Slot[], rank: Rank) => [...slots].sort(ORDER[rank]);

const Name = ({ s }: { s: Slot }) => <span className="sl-name">{s.row.slot}</span>;

function Tiles({ slots }: { slots: Slot[] }) {
  const byValue = ranked(slots, "value").filter((s) => s.value !== null);
  const bySeverity = ranked(slots, "severity");
  const byAlarms = ranked(slots, "alarms").filter((s) => s.share !== null);
  const nothing = slots.filter((s) => s.weighted === 0);
  const maxCritical = Math.max(...slots.map((s) => s.row.critical));
  const critical = slots.filter((s) => s.row.critical === maxCritical).sort(byName);
  const [best, least] = [byValue[0], byValue.at(-1)];
  const [most, fewest] = [byAlarms[0], byAlarms.at(-1)];
  return (
    <div className="rv-tiles">
      <div className="lb-tile static">
        <span className="lb-lbl">Cost · best value</span>
        <span className={best ? "lb-val rv-best" : "lb-val"}>{best ? formatUsdCompact(best.value!) : "—"}</span>
        {best && (
          <span className="sl-note">
            <Name s={best} /> per weighted finding
          </span>
        )}
        {least && least !== best && (
          <span className="sl-note">
            least: <Name s={least} /> <span className="rv-worst">{formatUsdCompact(least.value!)}</span>
          </span>
        )}
        {nothing.length > 0 && (
          <span className="sl-note">
            found nothing: {nothing.map((s) => `${s.row.slot} (${formatUsdCompact(s.row.costUsd)})`).join(", ")}
          </span>
        )}
      </div>
      <div className="lb-tile static">
        <span className="lb-lbl">Severity · most found</span>
        <span className="lb-val">{bySeverity[0].wpd.toFixed(2)}</span>
        <span className="sl-note">
          <Name s={bySeverity[0]} /> weighted per dispatch
        </span>
        {maxCritical > 0 && (
          <span className="sl-note">
            most critical:{" "}
            {critical.map((s, i) => (
              <span key={s.row.slot}>
                {i > 0 && ", "}
                <Name s={s} />
              </span>
            ))}{" "}
            ({maxCritical}
            {critical.length > 1 && " each"})
          </span>
        )}
      </div>
      <div className="lb-tile static">
        <span className="lb-lbl">False alarms · most withdrawn</span>
        <span className={most ? "lb-val rv-worst" : "lb-val"}>{most ? pct(most.share!) : "—"}</span>
        {most && (
          <span className="sl-note">
            <Name s={most} /> {most.row.withdrawn} of {most.row.critical + most.row.important + most.row.minor + most.row.withdrawn} findings
          </span>
        )}
        {fewest && fewest !== most && (
          <span className="sl-note">
            fewest: <Name s={fewest} /> <span className="rv-best">{pct(fewest.share!)}</span>
          </span>
        )}
      </div>
    </div>
  );
}

function Table({ slots, rank }: { slots: Slot[]; rank: Rank }) {
  const values = slots.flatMap((s) => (s.value === null ? [] : [s.value]));
  const [best, worst] = [Math.min(...values), Math.max(...values)];
  const maxUsd = Math.max(...slots.map((s) => s.row.costUsd)) || 1;
  const sort = (k: Rank) => (k === rank ? (k === "value" ? "ascending" : "descending") : undefined);
  const rows = slots.map((s) => s.row);
  const t = {
    dispatches: sum(rows.map((r) => r.dispatches)),
    priced: sum(rows.map((r) => r.pricedDispatches)),
    cost: sum(rows.map((r) => r.costUsd)),
    critical: sum(rows.map((r) => r.critical)),
    important: sum(rows.map((r) => r.important)),
    minor: sum(rows.map((r) => r.minor)),
    withdrawn: sum(rows.map((r) => r.withdrawn)),
  };
  const tWeighted = weightedOf(t);
  const tValue = valueOf(t.cost, t.priced, tWeighted, t.dispatches);
  const tFindings = t.critical + t.important + t.minor + t.withdrawn;
  return (
    <table className="lb-table">
      <thead>
        <tr className="rv-grp">
          <th colSpan={2}></th>
          <th className="rv-g" colSpan={3}>
            Cost
          </th>
          <th className="rv-g" colSpan={4}>
            Severity
          </th>
          <th className="rv-g">False alarms</th>
        </tr>
        <tr>
          <th>#</th>
          <th>Slot</th>
          <th className="r rv-g">Dispatches</th>
          <th className="r">Spend</th>
          <th className="r" aria-sort={sort("value")}>
            $ per weighted finding
          </th>
          <th className="r rv-g">Critical</th>
          <th className="r">Important</th>
          <th className="r">Minor</th>
          <th className="r" aria-sort={sort("severity")}>
            Weighted per dispatch
          </th>
          <th className="r rv-g" aria-sort={sort("alarms")}>
            Withdrawn
          </th>
        </tr>
      </thead>
      <tbody>
        {ranked(slots, rank).map((s, i) => {
          const r = s.row;
          const vcls = s.value === null ? "lb-muted" : s.value === best ? "rv-best" : s.value === worst ? "rv-worst" : "";
          return (
            <tr key={r.slot}>
              <td className="lb-muted">{i + 1}</td>
              <td>
                <div className="rv-slot">
                  <span>{r.slot}</span>
                  {r.experimental && (
                    <>
                      {" "}
                      <span role="img" aria-label="experimental reviewer" title={r.description}>
                        🧪
                      </span>
                    </>
                  )}
                </div>
                {r.description && <div className="rv-desc">{r.description}</div>}
              </td>
              <td className="r rv-g">{int(r.dispatches)}</td>
              <td className="r">
                {r.pricedDispatches ? (
                  <span className="lb-costcell">
                    <span className="lb-costbar">
                      <span style={{ width: `${((r.costUsd / maxUsd) * 100).toFixed(1)}%` }} />
                    </span>
                    <span className="sl-amt">{formatUsdCompact(r.costUsd)}</span>
                  </span>
                ) : (
                  "—"
                )}
              </td>
              <td className={`r ${vcls}`.trim()}>{s.value !== null ? formatUsdCompact(s.value) : s.weighted === 0 ? "found nothing" : "—"}</td>
              <td className="r rv-g">{int(r.critical)}</td>
              <td className="r">{int(r.important)}</td>
              <td className="r">{int(r.minor)}</td>
              <td className="r">{s.wpd.toFixed(2)}</td>
              <td className="r rv-g">
                {int(r.withdrawn)} <span className="lb-muted">{s.share === null ? "—" : pct(s.share)}</span>
              </td>
            </tr>
          );
        })}
      </tbody>
      <tfoot>
        <tr>
          <td colSpan={2}>All slots</td>
          <td className="r rv-g">{int(t.dispatches)}</td>
          <td className="r">
            {t.priced ? <span className="sl-amt">{formatUsdCompact(t.cost)}</span> : "—"}
            {t.priced < t.dispatches && (
              <>
                {" "}
                <span className="lb-muted">
                  {int(t.priced)} of {int(t.dispatches)} priced
                </span>
              </>
            )}
          </td>
          <td className="r">{tValue === null ? "—" : formatUsdCompact(tValue)}</td>
          <td className="r rv-g">{int(t.critical)}</td>
          <td className="r">{int(t.important)}</td>
          <td className="r">{int(t.minor)}</td>
          <td className="r">{t.dispatches ? (tWeighted / t.dispatches).toFixed(2) : "—"}</td>
          <td className="r rv-g">
            {int(t.withdrawn)} <span className="lb-muted">{tFindings ? pct(t.withdrawn / tFindings) : "—"}</span>
          </td>
        </tr>
      </tfoot>
    </table>
  );
}

export function Reviewers(_props: ViewProps) {
  const [f, setF] = useState<Filters>(defaults);
  const [rank, setRank] = useState<Rank>("value");
  const state = useStatsView<ReviewerRow[], ReviewerProject>("reviewers", {
    from: f.from,
    to: f.to,
    project: f.project ?? undefined,
    projects: true,
  });

  const offDefaults = f.period !== "30d" || f.project !== null;
  const ready = state.status === "ready" && state.data.recorded && !state.data.unmeasured;
  const slots = state.status === "ready" ? state.data.rows.map(derive) : [];
  const projects = state.status === "ready" ? (state.data.projects ?? []) : [];

  return (
    <section aria-label="Reviewers" className="dashboard lb sl">
      <div>
        <h2>Reviewers</h2>
        <p className="lb-sub">Which reviewer slots earn their cost, find the most severe defects, and raise false alarms?</p>
      </div>
      <div className="lb-fbar" role="toolbar" aria-label="Filters">
        <ChipGroup name="Period">
          {(Object.keys(PERIOD_DAYS) as PeriodChip[]).map((p) => (
            <Chip key={p} on={f.period === p} onClick={() => setF({ ...f, ...windowFor(p) })}>
              {p}
            </Chip>
          ))}
        </ChipGroup>
        <ChipGroup name="Project">
          <Chip on={f.project === null} onClick={() => setF({ ...f, project: null })}>
            All
          </Chip>
          {projects.map((p) => (
            <Chip key={p.projectKey} on={f.project === p.projectKey} onClick={() => setF({ ...f, project: p.projectKey })} title={p.projectKey}>
              {projectLabel(p.projectKey)}
              <span className="lb-c">{int(p.dispatches)}</span>
            </Chip>
          ))}
        </ChipGroup>
        <ChipGroup name="Rank by">
          {(Object.keys(RANKS) as Rank[]).map((k) => (
            <Chip key={k} on={rank === k} onClick={() => setRank(k)}>
              {RANKS[k]}
            </Chip>
          ))}
        </ChipGroup>
        {offDefaults && (
          <button type="button" className="lb-fclear" onClick={() => setF(defaults())}>
            Clear filters
          </button>
        )}
      </div>
      {ready && slots.length > 0 && <Tiles slots={slots} />}
      <div className="lb-panel">
        <div className="lb-panel-h">
          <span>{TITLE[rank]}</span>
          <span className="lb-panel-note">weighted finding = critical 10 · important 4 · minor 1, withdrawn excluded</span>
        </div>
        <div className="lb-scroll">
          {!ready ? (
            // Loading, error, not-recorded and unmeasured read as Panel renders them.
            <Panel title={TITLE[rank]} state={state}>
              {() => null}
            </Panel>
          ) : !slots.length ? (
            <div className="sl-empty">No reviewer dispatches in this period.</div>
          ) : (
            <Table slots={slots} rank={rank} />
          )}
        </div>
      </div>
    </section>
  );
}
