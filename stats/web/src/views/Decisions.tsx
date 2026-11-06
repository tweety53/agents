// Decisions -- "For each class of change, which implementer choice and review
// grouping deliver it cheapest, fastest and with the fewest findings?"
// (decision decisions-question-choice-and-grouping), built as the approved
// prototype draws it, on the Reviewers view's shape.
//
// The view owns its filters: a KAN-1048 chip bar of Period (7d / 30d / 90d),
// Project (run-count order, decisions-project-chip-order), Class and Review
// grouping, each counting its runs within the selections before it; chip state
// is local, never in the URL, and App's period, project and model props are
// ignored (decisions-chip-filters, decisions-drops-bar-and-old-tables). The
// view requests `decisions` once per period with no project and filters the
// whole unpaged result client-side, so every chip can count its runs.
//
// Everything else is derived here from the rows: the implementer choice folds
// every inline spelling into `inline`, else `<model> · <effort>`
// (decisions-implementer-choice-folding); one table grouped by class whose rows
// are implementer choice × review grouping pairs (decisions-rows-choice-by-grouping);
// median cost and wall-clock, mean weighted findings and fix rounds
// (decisions-medians, decisions-weighted-findings-withdrawn-excluded); rows of
// fewer than 3 runs tagged and never ranked (decisions-few-runs-rule); one tile
// per class (decisions-tiles-per-class); formats per decisions-display-formats.
import { Fragment, useState } from "react";
import type { DecisionRow } from "../api";
import { Chip, ChipGroup } from "../components/Chip";
import { Panel } from "../components/Panel";
import { PERIOD_DAYS, type PeriodChip, formatSpan, formatUsdCompact, sum, windowFor } from "../format";
import { useStatsView } from "../hooks/useStatsView";
import { projectLabel } from "../lib/projectLabel";
import type { ViewProps } from "../viewTypes";

const TITLE = "Every implementer choice and review grouping, by class";
const MIN_RUNS = 3;
const CLASSES = ["micro", "small", "regular", "big", "unclassified"];
const GROUPINGS = ["free", "static", "default"];

interface Filters {
  period: PeriodChip;
  project: string | null;
  cls: string | null;
  grp: string | null;
  from: Date;
  to: Date;
}

const defaults = (): Filters => ({ ...windowFor("30d"), project: null, cls: null, grp: null });

interface Run {
  row: DecisionRow;
  cls: string;
  choice: string;
  weighted: number;
}

interface Choice {
  key: string;
  choice: string;
  grouping: string;
  runs: Run[];
  cost: number;
  secs: number;
  wf: number;
  fr: number;
  fallbacks: number;
}

type Measure = "cost" | "secs" | "wf" | "fr";

interface ClassGroup {
  cls: string;
  choices: Choice[];
  runs: number;
}

const weightedOf = (r: { critical: number; important: number; minor: number }) => 10 * r.critical + 4 * r.important + r.minor;
const mean = (xs: number[]) => sum(xs) / xs.length;
function median(xs: number[]): number {
  const s = [...xs].sort((a, b) => a - b);
  const h = s.length >> 1;
  return s.length % 2 ? s[h] : (s[h - 1] + s[h]) / 2;
}

function choiceOf(r: DecisionRow): string {
  if (r.execution === "inline" || r.implementerModel.includes("inline")) return "inline";
  if (!r.implementerModel) return "—";
  return r.implementerEffort ? `${r.implementerModel} · ${r.implementerEffort}` : r.implementerModel;
}

const derive = (row: DecisionRow): Run => ({ row, cls: row.class || "unclassified", choice: choiceOf(row), weighted: weightedOf(row) });

// The known names first in their own order, any other after them by name.
const ordered = (known: string[], present: Iterable<string>) => {
  const set = new Set(present);
  return [...known.filter((k) => set.has(k)), ...[...set].filter((k) => !known.includes(k)).sort()];
};

function countBy(runs: Run[], key: (r: Run) => string): Map<string, number> {
  const m = new Map<string, number>();
  for (const r of runs) m.set(key(r), (m.get(key(r)) ?? 0) + 1);
  return m;
}

function groups(runs: Run[]): ClassGroup[] {
  const byCls = new Map<string, Map<string, Run[]>>();
  for (const r of runs) {
    const m = byCls.get(r.cls) ?? byCls.set(r.cls, new Map()).get(r.cls)!;
    const key = `${r.choice}|${r.row.grouping}`;
    (m.get(key) ?? m.set(key, []).get(key)!).push(r);
  }
  return ordered(CLASSES, byCls.keys()).map((cls) => {
    const choices = [...byCls.get(cls)!.entries()]
      .map(([key, rs]) => ({
        key,
        choice: rs[0].choice,
        grouping: rs[0].row.grouping,
        runs: rs,
        cost: median(rs.map((r) => r.row.costUsd)),
        secs: median(rs.map((r) => r.row.wallClockSeconds)),
        wf: mean(rs.map((r) => r.weighted)),
        fr: mean(rs.map((r) => r.row.fixRounds)),
        fallbacks: sum(rs.map((r) => r.row.fallbacks)),
      }))
      .sort((a, b) => b.runs.length - a.runs.length || a.choice.localeCompare(b.choice) || a.grouping.localeCompare(b.grouping));
    return { cls, choices, runs: sum(choices.map((c) => c.runs.length)) };
  });
}

const ranked = (choices: Choice[]) => choices.filter((c) => c.runs.length >= MIN_RUNS);

// Best / worst per measure among the class's rows of 3 or more runs, when at least two exist; lower is better.
function tone(choices: Choice[], c: Choice, key: Measure): string {
  const ok = ranked(choices);
  if (c.runs.length < MIN_RUNS || ok.length < 2) return "";
  const vals = ok.map((x) => x[key]);
  return c[key] === Math.min(...vals) ? "rv-best" : c[key] === Math.max(...vals) ? "rv-worst" : "";
}

// The lowest value's row, ties to the first in table order.
const pick = (choices: Choice[], key: Measure) => choices.reduce<Choice | undefined>((best, c) => (!best || c[key] < best[key] ? c : best), undefined);

const Name = ({ c }: { c: Choice }) => (
  <span className="sl-name">
    {c.choice} · {c.grouping}
  </span>
);

function Tiles({ gs }: { gs: ClassGroup[] }) {
  return (
    <div className="dc-tiles">
      {gs.map((g) => {
        const ok = ranked(g.choices);
        const [cheap, fast, clean] = [pick(ok, "cost"), pick(ok, "secs"), pick(ok, "wf")];
        return (
          <div key={g.cls} className="lb-tile static">
            <span className="lb-lbl">{g.cls} · cheapest</span>
            <span className="lb-val">{cheap ? formatUsdCompact(cheap.cost) : "—"}</span>
            {cheap && fast && clean ? (
              <>
                <span className="sl-note">
                  <Name c={cheap} /> median per run
                </span>
                <span className="sl-note">
                  fastest: <Name c={fast} /> {formatSpan(fast.secs * 1000)}
                </span>
                <span className="sl-note">
                  fewest findings: <Name c={clean} /> {clean.wf.toFixed(1)} / run
                </span>
              </>
            ) : (
              <span className="sl-note">
                no choice has {MIN_RUNS} runs yet · {g.runs} run{g.runs === 1 ? "" : "s"}
              </span>
            )}
          </div>
        );
      })}
    </div>
  );
}

function Table({ gs }: { gs: ClassGroup[] }) {
  const [open, setOpen] = useState<Set<string>>(new Set());
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set());
  const flip = (set: Set<string>, k: string) => {
    const next = new Set(set);
    if (!next.delete(k)) next.add(k);
    return next;
  };
  const maxCost = Math.max(...gs.flatMap((g) => g.choices.map((c) => c.cost))) || 1;
  const all = gs.flatMap((g) => g.choices.flatMap((c) => c.runs));
  return (
    <table className="lb-table">
      <thead>
        <tr>
          <th>Implementer choice</th>
          <th>Review grouping</th>
          <th className="r">Runs</th>
          <th className="r">Median cost</th>
          <th className="r">Median wall-clock</th>
          <th className="r">Weighted findings / run</th>
          <th className="r">Fix rounds / run</th>
          <th className="r">Fallbacks</th>
        </tr>
      </thead>
      <tbody>
        {gs.map((g) => {
          const isCollapsed = collapsed.has(g.cls);
          const toggle = () => setCollapsed((s) => flip(s, g.cls));
          return (
            <Fragment key={g.cls}>
              <tr
                className={`lb-group${isCollapsed ? " collapsed" : ""}`}
                tabIndex={0}
                aria-expanded={!isCollapsed}
                onClick={toggle}
                onKeyDown={(e) => e.key === "Enter" && toggle()}
              >
                <td colSpan={8}>
                  <span className="lb-caret">▾</span>
                  {g.cls}
                  <span className="lb-cnt">{g.runs} runs</span>
                </td>
              </tr>
              {!isCollapsed &&
                g.choices.map((c) => {
                  const id = `${g.cls}|${c.key}`;
                  const isOpen = open.has(id);
                  const t = (key: Measure) => `r ${tone(g.choices, c, key)}`.trim();
                  return (
                    <Fragment key={id}>
                      <tr
                        className="lb-row expandable"
                        tabIndex={0}
                        aria-expanded={isOpen}
                        onClick={() => setOpen((s) => flip(s, id))}
                        onKeyDown={(e) => e.key === "Enter" && setOpen((s) => flip(s, id))}
                      >
                        <td>
                          <span className="lb-caret" style={{ transform: isOpen ? "none" : "rotate(-90deg)" }}>
                            ▾
                          </span>
                          <span className="dc-choice">{c.choice}</span>
                          {c.runs.length < MIN_RUNS && <span className="dc-few">few runs</span>}
                        </td>
                        <td>
                          <span className="dc-grp">{c.grouping}</span>
                        </td>
                        <td className="r">{c.runs.length}</td>
                        <td className={t("cost")}>
                          <span className="lb-costcell">
                            <span className="lb-costbar">
                              <span style={{ width: `${((c.cost / maxCost) * 100).toFixed(1)}%` }} />
                            </span>
                            <span className="sl-amt">{formatUsdCompact(c.cost)}</span>
                          </span>
                        </td>
                        <td className={t("secs")}>{formatSpan(c.secs * 1000)}</td>
                        <td className={t("wf")}>{c.wf.toFixed(1)}</td>
                        <td className={t("fr")}>{c.fr.toFixed(2)}</td>
                        <td className="r">{c.fallbacks ? <span className="lb-err">{c.fallbacks}</span> : <span className="lb-muted">0</span>}</td>
                      </tr>
                      {isOpen &&
                        c.runs.map(({ row: r, weighted }) => (
                          <tr key={`${r.project}/${r.change}/${r.recordedAt}`} className="dc-run">
                            <td>
                              <a href={`#/run/${encodeURIComponent(r.project)}/${encodeURIComponent(r.change)}`} onClick={(e) => e.stopPropagation()}>
                                {r.change}
                              </a>{" "}
                              <span className="lb-proj">
                                {projectLabel(r.project)} · {r.recordedAt.slice(0, 10)}
                              </span>
                            </td>
                            <td></td>
                            <td className="r"></td>
                            <td className="r">{formatUsdCompact(r.costUsd)}</td>
                            <td className="r">{formatSpan(r.wallClockSeconds * 1000)}</td>
                            <td className="r">
                              {weighted}{" "}
                              <span className="lb-muted">
                                {r.critical} · {r.important} · {r.minor}
                              </span>
                            </td>
                            <td className="r">{r.fixRounds}</td>
                            <td className="r">{r.fallbacks || ""}</td>
                          </tr>
                        ))}
                    </Fragment>
                  );
                })}
            </Fragment>
          );
        })}
      </tbody>
      <tfoot>
        <tr>
          <td>All runs</td>
          <td></td>
          <td className="r">{all.length}</td>
          <td className="r">{formatUsdCompact(median(all.map((r) => r.row.costUsd)))}</td>
          <td className="r">{formatSpan(median(all.map((r) => r.row.wallClockSeconds)) * 1000)}</td>
          <td className="r">{mean(all.map((r) => r.weighted)).toFixed(1)}</td>
          <td className="r">{mean(all.map((r) => r.row.fixRounds)).toFixed(2)}</td>
          <td className="r">{sum(all.map((r) => r.row.fallbacks))}</td>
        </tr>
      </tfoot>
    </table>
  );
}

export function Decisions(_props: ViewProps) {
  const [f, setF] = useState<Filters>(defaults);
  const state = useStatsView<DecisionRow[]>("decisions", { from: f.from, to: f.to });

  const ready = state.status === "ready" && state.data.recorded && !state.data.unmeasured;
  const period = state.status === "ready" ? state.data.rows.map(derive) : [];
  const inProject = period.filter((r) => !f.project || r.row.project === f.project);
  const inClass = inProject.filter((r) => !f.cls || r.cls === f.cls);
  const shown = inClass.filter((r) => !f.grp || r.row.grouping === f.grp);
  const projects = [...countBy(period, (r) => r.row.project).entries()].sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]));
  const clsCounts = countBy(inProject, (r) => r.cls);
  const grpCounts = countBy(inClass, (r) => r.row.grouping);
  const gs = groups(shown);
  const offDefaults = f.period !== "30d" || f.project !== null || f.cls !== null || f.grp !== null;

  return (
    <section aria-label="Decisions" className="dashboard lb sl">
      <div>
        <h2>Decisions</h2>
        <p className="lb-sub">For each class of change, which implementer choice and review grouping deliver it cheapest, fastest and with the fewest findings?</p>
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
          {projects.map(([key, n]) => (
            <Chip key={key} on={f.project === key} onClick={() => setF({ ...f, project: key })} title={key}>
              {projectLabel(key)}
              <span className="lb-c">{n}</span>
            </Chip>
          ))}
        </ChipGroup>
        <ChipGroup name="Class">
          <Chip on={f.cls === null} onClick={() => setF({ ...f, cls: null })}>
            All
          </Chip>
          {ordered(CLASSES, clsCounts.keys()).map((c) => (
            <Chip key={c} on={f.cls === c} onClick={() => setF({ ...f, cls: c })}>
              {c}
              <span className="lb-c">{clsCounts.get(c)}</span>
            </Chip>
          ))}
        </ChipGroup>
        <ChipGroup name="Review grouping">
          <Chip on={f.grp === null} onClick={() => setF({ ...f, grp: null })}>
            All
          </Chip>
          {ordered(GROUPINGS, grpCounts.keys()).map((g) => (
            <Chip key={g} on={f.grp === g} onClick={() => setF({ ...f, grp: g })}>
              {g}
              <span className="lb-c">{grpCounts.get(g)}</span>
            </Chip>
          ))}
        </ChipGroup>
        {offDefaults && (
          <button type="button" className="lb-fclear" onClick={() => setF(defaults())}>
            Clear filters
          </button>
        )}
      </div>
      {ready && gs.length > 0 && <Tiles gs={gs} />}
      <div className="lb-panel">
        <div className="lb-panel-h">
          <span>{TITLE}</span>
          <span className="lb-panel-note">median per run · weighted finding = critical 10 · important 4 · minor 1, withdrawn excluded · fewer than 3 runs never marked best or worst</span>
        </div>
        <div className="lb-scroll">
          {!ready ? (
            // Loading, error, not-recorded and unmeasured read as Panel renders them.
            <Panel title={TITLE} state={state}>
              {() => null}
            </Panel>
          ) : !gs.length ? (
            <div className="sl-empty">No decisions in this period.</div>
          ) : (
            <Table gs={gs} />
          )}
        </div>
      </div>
    </section>
  );
}
