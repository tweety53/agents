// Confidence -- every confidence estimate below 100% an agent filed (KAN-1130): how confident it
// was that the plan is executed as written, and what would move it to 100%. Self-review's layout
// as is: its To review / Approved / Declined chips, the list paged 25 rows at a time, and the
// selected estimate's gaps in full beside it, approved or declined here with an optional reason.
import { useEffect, useState } from "react";
import { Chip, ChipGroup } from "../components/Chip";
import { ViewFrame } from "../components/ViewFrame";
import { fetchConfidencePage, reviewConfidenceEstimate, type ConfidencePage, type ConfidenceRow, type ConfidenceStatus } from "../api";

const SHOWS: { key: ConfidenceStatus; label: string }[] = [
  { key: "to-review", label: "To review" },
  { key: "approved", label: "Approved" },
  { key: "declined", label: "Declined" },
];
const PAGE = 25;

const title = (r: ConfidenceRow) => `${r.score}% · ${r.change}`;

/** The estimate, whole, for pasting to an agent: nothing here is readable by a store verb, so the text carries it. */
const agentText = (r: ConfidenceRow) =>
  `Confidence estimate ${r.project}/${r.change} (${r.stage}${r.task ? ` task ${r.task}` : ""}): ${r.score}%.\nWhat would move it to 100%:\n${r.gaps.map((g) => `- ${g}`).join("\n")}\nTags: ${r.tags.join(", ")}\n\nAnalyze what can be done and where, and say what generic fix would prevent this kind of gap in any future task — a fix specific to this task is useless.`;

function Meta({ row }: { row: ConfidenceRow }) {
  return (
    <span className="sr-meta">
      {row.tags.map((t) => (
        <code key={t} className="sr-theme">
          {t}
        </code>
      ))}
      <span className="sr-pill">{row.stage}</span>
      {row.disposition && <span className="sr-pill">{row.disposition}</span>}
      <span className="lb-muted">
        {row.task ? `task ${row.task} · ` : ""}
        {row.recordedAt.slice(0, 10)}
      </span>
    </span>
  );
}

export function Confidence() {
  const [show, setShow] = useState<ConfidenceStatus>("to-review");
  const [offset, setOffset] = useState(0);
  const [selId, setSelId] = useState<number | null>(null);
  const [reason, setReason] = useState("");
  const [reload, setReload] = useState(0);
  const [page, setPage] = useState<ConfidencePage | null>(null);
  const [error, setError] = useState("");
  const [failed, setFailed] = useState("");

  // The list on screen, one page; a decision made here refetches it.
  useEffect(() => {
    let live = true;
    setError("");
    fetchConfidencePage({ status: show, limit: PAGE, offset }).then(
      (p) => {
        if (!live) return;
        if (p.rows.length === 0 && offset > 0) setOffset(Math.max(0, offset - PAGE));
        else setPage(p);
      },
      (e: unknown) => live && setError(e instanceof Error ? e.message : String(e)),
    );
    return () => {
      live = false;
    };
  }, [show, offset, reload]);

  const rows = page?.rows ?? [];
  const sel = rows.find((r) => r.id === selId) ?? rows[0];
  const last = page ? Math.min(page.offset + rows.length, page.total) : 0;
  const counts: Record<ConfidenceStatus, number> = { "to-review": page?.toReview ?? 0, approved: page?.approved ?? 0, declined: page?.declined ?? 0 };

  const decide = (row: ConfidenceRow, to: ConfidenceStatus) =>
    reviewConfidenceEstimate(row.project, row.id, to, to === "to-review" ? "" : reason.trim()).then(
      () => {
        setFailed("");
        setReason("");
        setSelId(null);
        setReload((n) => n + 1);
      },
      (err: unknown) => setFailed(err instanceof Error ? err.message : String(err)),
    );

  return (
    <ViewFrame title="Confidence" description="Every estimate below 100% an agent gave of executing the plan as written, and what it said would raise it, for you to approve or decline.">
      <div className="lb-fbar">
        <ChipGroup name="Show">
          {SHOWS.map(({ key, label }) => (
            <Chip
              key={key}
              on={show === key}
              onClick={() => {
                setShow(key);
                setOffset(0);
                setSelId(null);
                setPage(null);
              }}
            >
              {label} {counts[key]}
            </Chip>
          ))}
        </ChipGroup>
      </div>
      <section className="lb-panel" aria-label={SHOWS.find((s) => s.key === show)!.label}>
        <div className="lb-panel-b">
          {error && <p className="sr-err">{error}</p>}
          {failed && <p className="sr-err">Could not save: {failed}</p>}
          {!page && !error && <p className="lb-muted">Loading…</p>}
          {page && rows.length === 0 && <p className="lb-muted">Nothing here.</p>}
          {page && sel && (
            <div className="sr-split">
              <div className="sr-listcol">
                {rows.map((r) => (
                  <button key={r.id} className={`sr-item${r.id === sel.id ? " on" : ""}`} data-testid="estimate" onClick={() => setSelId(r.id)}>
                    <strong>{title(r)}</strong>
                    <Meta row={r} />
                  </button>
                ))}
                {page.total > page.limit && (
                  <div className="sr-pager">
                    <button type="button" disabled={page.offset === 0} onClick={() => setOffset(Math.max(0, page.offset - page.limit))}>
                      Previous
                    </button>
                    <span className="lb-muted">
                      {page.offset + 1}–{last} of {page.total}
                    </span>
                    <button type="button" disabled={last >= page.total} onClick={() => setOffset(page.offset + page.limit)}>
                      Next
                    </button>
                  </div>
                )}
              </div>
              <article className="sr-pane" aria-label="Estimate">
                <h3>{title(sel)}</h3>
                <Meta row={sel} />
                <div className="sr-meta sr-id">
                  <button type="button" onClick={() => void navigator.clipboard.writeText(agentText(sel))}>
                    Copy for agent
                  </button>
                </div>
                <h4 className="sr-h">What would move it to 100%</h4>
                <ul className="sr-list">
                  {sel.gaps.map((g, i) => (
                    <li key={i}>{g}</li>
                  ))}
                </ul>
                {show === "to-review" ? (
                  <footer className="sr-actions">
                    <button type="button" className="sr-approve" onClick={() => void decide(sel, "approved")}>
                      Approve
                    </button>
                    <button type="button" className="sr-decline" onClick={() => void decide(sel, "declined")}>
                      Decline
                    </button>
                    <input aria-label="Reason" placeholder="Reason (optional, stored with the decision)" value={reason} onChange={(e) => setReason(e.target.value)} />
                  </footer>
                ) : (
                  <footer className="sr-actions">
                    <span>{sel.reason ? `Reason: ${sel.reason}` : "No reason given."}</span>
                    <button type="button" onClick={() => void decide(sel, "to-review")}>Move back to review</button>
                  </footer>
                )}
              </article>
            </div>
          )}
        </div>
      </section>
    </ViewFrame>
  );
}
