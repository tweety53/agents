"""Does a cross-encoder reranker help retrieval over self-review findings? Two tasks, hybrid (BM25 + bge-small, RRF)
with and without reranking the top N:
  qa     a question written for a finding -> does its source finding come back
  theme  a finding's title as the query    -> do findings of the same theme tag come back
Data is a snapshot of the dev store, not committed (it is the project's own finding text):
  <data>/corpus.pkl     {"rows": [stats API finding rows], "T": [indexed text per row]}   (rows carry `theme`)
  <data>/questions.json {"<row index>": "<question>"}
Run: python rerank_eval.py <data dir>
"""
import json, pickle, re, sys, time
from collections import Counter
import numpy as np
from rank_bm25 import BM25Okapi
from sentence_transformers import CrossEncoder, SentenceTransformer

data = sys.argv[1] if len(sys.argv) > 1 else "."
c = pickle.load(open(f"{data}/corpus.pkl", "rb"))
rows, T = c["rows"], c["T"]
Q = json.load(open(f"{data}/questions.json"))
tok = lambda t: re.findall(r"[a-z0-9]+", t.lower())
title = lambda r: re.sub(r"^\[theme: [a-z0-9-]+\]\s*", "", r["note"])
bm = BM25Okapi([tok(t) for t in T])
emb = SentenceTransformer("BAAI/bge-small-en-v1.5")
E = emb.encode(T, normalize_embeddings=True, show_progress_bar=False)


def hybrid_order(q, qtext_vec=None, exclude=None):
    b = bm.get_scores(tok(q))
    v = E @ (qtext_vec if qtext_vec is not None else emb.encode([q], normalize_embeddings=True, show_progress_bar=False)[0])
    if exclude is not None:
        b[exclude] = -1e9
        v[exclude] = -1e9
    rk = lambda s: np.argsort(np.argsort(-s)) + 1
    return list(np.argsort(-(1 / (60 + rk(b)) + 1 / (60 + rk(v)))))


def rerank(ce, q, order, n):
    head = order[:n]
    sc = ce.predict([(q, T[j]) for j in head], show_progress_bar=False, batch_size=16)
    return [head[j] for j in np.argsort(-sc)] + order[n:]


models = {"base": CrossEncoder("BAAI/bge-reranker-base", max_length=512)}
if "--v2" in sys.argv:
    models["v2-m3"] = CrossEncoder("BAAI/bge-reranker-v2-m3", max_length=1024)

# ---- qa
qi = [int(k) for k in Q]
qs = [Q[k] for k in Q]
base_orders = [hybrid_order(q) for q in qs]
print(f"\nQA: {len(qs)} questions over {len(rows)} findings   (recall@1 / @5 / @10 / MRR | ms per query)")
def qa_row(name, orders, ms=0):
    r = [o.index(i) + 1 for o, i in zip(orders, qi)]
    f = lambda n: np.mean([x <= n for x in r])
    print(f"  {name:28s} {f(1):.2f} / {f(5):.2f} / {f(10):.2f} / {np.mean([1/x for x in r]):.2f} | {ms:5.0f}")
qa_row("hybrid", base_orders)
for mname, ce in models.items():
    for N in (10, 20, 50):
        t0 = time.time()
        orders = [rerank(ce, q, o, N) for q, o in zip(qs, base_orders)]
        qa_row(f"hybrid + {mname} top {N}", orders, (time.time() - t0) / len(qs) * 1000)

# ---- theme
rows_t = [i for i, r in enumerate(rows) if r["theme"] and r["theme"] != "uncategorized"]
cnt = Counter(rows[i]["theme"] for i in rows_t)
rows_t = [i for i in rows_t if cnt[rows[i]["theme"]] >= 2]
big = {t for t, k in cnt.items() if k > 60}
small = [i for i in rows_t if rows[i]["theme"] not in big]
print(f"\nTHEME: {len(rows_t)} findings as queries, {len(cnt)} themes ({len(small)} queries in themes of <=60 rows)   (hit@5 / P@5 / MRR | ms per query)")
labels = [r["theme"] for r in rows]
cache = {i: hybrid_order(title(rows[i]), E[i], exclude=i) for i in rows_t}

def th_row(name, order_of, subset, ms=0):
    hit = p5 = mrr = 0
    for i in subset:
        same = [labels[j] == labels[i] for j in order_of(i)]
        hit += any(same[:5]); p5 += np.mean(same[:5]); mrr += 1 / (same.index(True) + 1)
    k = len(subset)
    print(f"  {name:28s} {hit/k:.2f} / {p5/k:.2f} / {mrr/k:.2f} | {ms:5.0f}")

for label, subset in (("all themes", rows_t), ("themes <=60 rows", small)):
    print(f" {label}")
    th_row("hybrid", lambda i: cache[i], subset)
    for mname, ce in models.items():
        for N in (10, 20):
            t0 = time.time()
            rer = {i: rerank(ce, title(rows[i]), cache[i], N) for i in subset}
            th_row(f"hybrid + {mname} top {N}", lambda i: rer[i], subset, (time.time() - t0) / len(subset) * 1000)
