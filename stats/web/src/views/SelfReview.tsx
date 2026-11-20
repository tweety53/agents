// SelfReview -- what the self-review pass did, in two read tabs: Fixed (every
// finding it landed a fix for, with its full detail, read only) and To review
// (the improvements it suggests, each approved or declined here; a status
// filter keeps decided ones out of the way). Nothing else: tasks are the
// operator's to create and plan.
import { useState } from "react";
import { DataTable, type Column } from "../components/DataTable";
import { Panel } from "../components/Panel";
import { ViewFrame } from "../components/ViewFrame";
import { reviewSelfReviewFinding, type SelfReviewRow } from "../api";
import { useStatsView } from "../hooks/useStatsView";
import type { ViewProps } from "../viewTypes";

type Tab = "fixed" | "to-review";
type Status = "to-review" | "approved" | "declined";
const STATUS_LABEL: Record<Status, string> = { "to-review": "To review", approved: "Approved", declined: "Declined" };

function Detail({ row }: { row: SelfReviewRow }) {
  return (
    <>
      <div>{row.note}</div>
      {row.detail && (
        <details className="sr-detail">
          <summary>Details</summary>
          <pre>{row.detail}</pre>
        </details>
      )}
    </>
  );
}

const fixedColumns: Column<SelfReviewRow>[] = [
  { key: "recordedAt", header: "Recorded", sortable: true, accessor: (r) => r.recordedAt, render: (r) => r.recordedAt.slice(0, 10) },
  { key: "change", header: "Change", sortable: true, filterable: true, accessor: (r) => r.change },
  { key: "angle", header: "Angle", sortable: true, filterable: true, accessor: (r) => r.angle },
  { key: "severity", header: "Severity", sortable: true, filterable: true, accessor: (r) => r.severity },
  { key: "note", header: "Finding", accessor: (r) => r.note, render: (r) => <Detail row={r} /> },
  { key: "ref", header: "Fix", accessor: (r) => r.ref, render: (r) => (r.commitUrl ? <a href={r.commitUrl}>{r.ref}</a> : r.ref) },
];

export function SelfReview({ period, onPeriodChange, project }: ViewProps) {
  const state = useStatsView<SelfReviewRow[]>("self-review", { from: period.from, to: period.to, project });
  const [tab, setTab] = useState<Tab>("fixed");
  const [status, setStatus] = useState<Status>("to-review");
  // A decision made here shows at once; the next load reads it from the store.
  const [decided, setDecided] = useState<Record<number, Status>>({});
  const [failed, setFailed] = useState("");

  const decide = (row: SelfReviewRow, to: Status) =>
    reviewSelfReviewFinding(row.project, row.id, to).then(
      () => {
        setFailed("");
        setDecided((d) => ({ ...d, [row.id]: to }));
      },
      (err: unknown) => setFailed(err instanceof Error ? err.message : String(err)),
    );

  return (
    <ViewFrame title="Self-review" description="What each self-review pass fixed, and the improvements it suggests for you to approve or decline.">
      <div role="tablist" className="sr-tabs">
        {([["fixed", "Fixed"], ["to-review", "To review"]] as const).map(([t, label]) => (
          <button key={t} role="tab" aria-selected={tab === t} className={tab === t ? "sr-tab on" : "sr-tab"} onClick={() => setTab(t)}>
            {label}
          </button>
        ))}
      </div>
      {tab === "fixed" ? (
        <Panel title="Fixed" state={state}>
          {(data) => (
            <DataTable
              columns={fixedColumns}
              rows={data.rows.filter((r) => r.disposition === "fixed")}
              rowKey={(r) => String(r.id)}
              emptyMessage="No fixed findings in this period."
              period={period}
              onPeriodChange={onPeriodChange}
            />
          )}
        </Panel>
      ) : (
        <Panel title="To review" state={state}>
          {(data) => {
            const suggestions = data.rows
              .filter((r) => r.disposition === "suggested")
              .map((r) => ({ row: r, status: decided[r.id] ?? ((r.reviewStatus || "to-review") as Status) }));
            const shown = suggestions.filter((s) => s.status === status);
            return (
              <>
                <div role="group" aria-label="Status" className="sr-tabs">
                  {(Object.keys(STATUS_LABEL) as Status[]).map((s) => (
                    <button key={s} aria-pressed={status === s} className={status === s ? "sr-tab on" : "sr-tab"} onClick={() => setStatus(s)}>
                      {STATUS_LABEL[s]} {suggestions.filter((x) => x.status === s).length}
                    </button>
                  ))}
                </div>
                {failed && <p className="sr-err">Could not save: {failed}</p>}
                {shown.length === 0 && <p className="lb-muted">Nothing {STATUS_LABEL[status].toLowerCase()}.</p>}
                {shown.map(({ row }) => (
                  <article key={row.id} className="sr-card" data-testid="suggestion">
                    <header>
                      <strong>{row.note}</strong>
                      <span className="lb-muted">
                        {row.change} · {row.angle}
                        {row.severity && ` · ${row.severity}`} · {row.recordedAt.slice(0, 10)}
                      </span>
                    </header>
                    <pre>{row.detail}</pre>
                    {status === "to-review" && (
                      <footer>
                        <button onClick={() => void decide(row, "approved")}>Approve</button>
                        <button onClick={() => void decide(row, "declined")}>Decline</button>
                      </footer>
                    )}
                  </article>
                ))}
              </>
            );
          }}
        </Panel>
      )}
    </ViewFrame>
  );
}
