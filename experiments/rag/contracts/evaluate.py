"""Score retrieval on questions.json: does the source chunk (and its file) come back? Also finds near-duplicate chunk pairs."""
import json, os, sys
import numpy as np
from common import load_questions
from search import Index

ix = Index()
ids, qs, kinds = load_questions(ix.chunks)
file_of = [c["file"] for c in ix.chunks]
print(f"{len(ids)} questions over {len(ix.chunks)} chunks in {len(set(file_of))} files")
print(f"{'method':22s} chunk r@1 r@5 r@10 MRR | file r@1 r@5")
for name, mode, w in (("bm25", "bm25", 1), ("bge-small", "vec", 1), ("hybrid RRF", "hybrid", 1), ("hybrid w_bm25=2", "hybrid", 2), ("hybrid w_bm25=0.5", "hybrid", .5)):
    rc, rf = [], []
    for i, q in zip(ids, qs):
        order = ix.rank(q, mode, k=len(ix.chunks), w=w)
        rc.append(order.index(i) + 1)
        seen = []
        for j in order:
            if file_of[j] not in seen:
                seen.append(file_of[j])
        rf.append(seen.index(file_of[i]) + 1)
    f = lambda r, n: np.mean([x <= n for x in r])
    print(f"{name:22s}        {f(rc,1):.2f} {f(rc,5):.2f} {f(rc,10):.2f} {np.mean([1/x for x in rc]):.2f} |       {f(rf,1):.2f} {f(rf,5):.2f}")
# misses for the hybrid
# cross-encoder reranker over the hybrid's top-N
from sentence_transformers import CrossEncoder
rer = CrossEncoder("BAAI/bge-reranker-base")
print("\nhybrid + bge-reranker-base over the top N (chunk r@1 / r@5 / r@10 / MRR, file r@1 / r@5):")
import time
for N in (10, 20, 50):
    rc, rf, t0 = [], [], time.time()
    for i, q in zip(ids, qs):
        order = ix.rank(q, "hybrid", k=len(ix.chunks))
        head = order[:N]
        sc = rer.predict([(q, ix.docs[j][:2000]) for j in head], show_progress_bar=False)
        order = [head[j] for j in np.argsort(-sc)] + order[N:]
        rc.append(order.index(i) + 1)
        seen = []
        for j in order:
            if file_of[j] not in seen:
                seen.append(file_of[j])
        rf.append(seen.index(file_of[i]) + 1)
    f = lambda r, n: np.mean([x <= n for x in r])
    print(f"  top {N:2d}: {f(rc,1):.2f} / {f(rc,5):.2f} / {f(rc,10):.2f} / {np.mean([1/x for x in rc]):.2f} | {f(rf,1):.2f} / {f(rf,5):.2f}   ({(time.time()-t0)/len(ids)*1000:.0f} ms per query)")
print("\nhybrid, by question kind (chunk r@1 / r@5 / MRR):")
for kind in sorted(set(kinds)):
    r = [ix.rank(q, "hybrid", k=len(ix.chunks)).index(i) + 1 for i, q, kd in zip(ids, qs, kinds) if kd == kind]
    print(f"  {kind:5s} n={len(r)}  {np.mean([x <= 1 for x in r]):.2f} / {np.mean([x <= 5 for x in r]):.2f} / {np.mean([1/x for x in r]):.2f}")
print("\nhybrid misses outside top 5:")
for i, q in zip(ids, qs):
    order = ix.rank(q, "hybrid", k=len(ix.chunks))
    r = order.index(i) + 1
    if r > 5:
        c = ix.chunks[i]
        top = ix.chunks[order[0]]
        print(f"  rank {r}: Q={q}\n     want {c['file']}:{c['line']} [{c['heading'][:60]}]\n     got  {top['file']}:{top['line']} [{top['heading'][:60]}]")

# use case 3: near-duplicate chunks across different files (candidate overlaps or contradictions)
E = ix.E
S = E @ E.T
np.fill_diagonal(S, -1)
pairs = []
n = len(E)
for i in range(n):
    for j in range(i + 1, n):
        if S[i, j] >= 0.88 and file_of[i] != file_of[j] and ix.chunks[i]["words"] >= 25 and ix.chunks[j]["words"] >= 25:
            pairs.append((S[i, j], i, j))
pairs.sort(reverse=True)
print(f"\nnear-duplicate chunk pairs across different files (cosine >= 0.88): {len(pairs)}")
json.dump([{"cos": float(s), "a": int(i), "b": int(j)} for s, i, j in pairs[:60]], open(os.path.join(os.path.dirname(os.path.abspath(__file__)), "dup_pairs.json"), "w"))
for s, i, j in pairs[:12]:
    a, b = ix.chunks[i], ix.chunks[j]
    print(f"  {s:.3f}  {a['file']}:{a['line']} [{a['heading'][:45]}]  <->  {b['file']}:{b['line']} [{b['heading'][:45]}]")
