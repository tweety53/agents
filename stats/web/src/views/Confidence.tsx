// Confidence -- every confidence estimate below 100% an agent filed (KAN-1130): how confident it
// was that the plan is executed as written, and what would move it to 100%. Self-review's layout
// as is: a chip per tag where Self-review has its Show chips, the list paged 25 rows at a time,
// and the selected estimate's gaps in full beside it. Read only.
import { useEffect, useState } from "react";
import { Chip, ChipGroup } from "../components/Chip";
import { ViewFrame } from "../components/ViewFrame";
import { fetchConfidencePage, type ConfidencePage, type ConfidenceRow } from "../api";

const PAGE = 25;

const title = (r: ConfidenceRow) => `${r.score}% · ${r.change}`;

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
  const [tag, setTag] = useState("");
  const [offset, setOffset] = useState(0);
  const [selId, setSelId] = useState<number | null>(null);
  const [page, setPage] = useState<ConfidencePage | null>(null);
  const [allTotal, setAllTotal] = useState(0);
  const [error, setError] = useState("");

  useEffect(() => {
    let live = true;
    setError("");
    fetchConfidencePage({ tag, limit: PAGE, offset }).then(
      (p) => {
        if (!live) return;
        setPage(p);
        if (!tag) setAllTotal(p.total);
      },
      (e: unknown) => live && setError(e instanceof Error ? e.message : String(e)),
    );
    return () => {
      live = false;
    };
  }, [tag, offset]);

  const rows = page?.rows ?? [];
  const sel = rows.find((r) => r.id === selId) ?? rows[0];
  const last = page ? Math.min(page.offset + rows.length, page.total) : 0;
  const pick = (t: string) => {
    setTag(t);
    setOffset(0);
    setSelId(null);
    setPage(null);
  };

  return (
    <ViewFrame title="Confidence" description="Every estimate below 100% an agent gave of executing the plan as written, and what it said would raise it.">
      <div className="lb-fbar">
        <ChipGroup name="Tag">
          <Chip on={tag === ""} onClick={() => pick("")}>
            All {allTotal}
          </Chip>
          {(page?.tags ?? []).map((t) => (
            <Chip key={t.tag} on={tag === t.tag} onClick={() => pick(t.tag)}>
              {t.tag} {t.count}
            </Chip>
          ))}
        </ChipGroup>
      </div>
      <section className="lb-panel" aria-label={tag || "All"}>
        <div className="lb-panel-b">
          {error && <p className="sr-err">{error}</p>}
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
                <h4 className="sr-h">What would move it to 100%</h4>
                <ul className="sr-list">
                  {sel.gaps.map((g, i) => (
                    <li key={i}>{g}</li>
                  ))}
                </ul>
              </article>
            </div>
          )}
        </div>
      </section>
    </ViewFrame>
  );
}
