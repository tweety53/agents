"""Compare cross-encoders over the hybrid's top-N on the contract questions."""
import json, pickle, re, sys, time
import numpy as np
from sentence_transformers import CrossEncoder
from common import load_questions
from search import Index

ix = Index()
ids, qs, _ = load_questions(ix.chunks)
file_of = [c["file"] for c in ix.chunks]
orders = [ix.rank(q, "hybrid", k=len(ix.chunks)) for q in qs]
f = lambda r, n: np.mean([x <= n for x in r])
print(f"{len(qs)} contract questions; chunk r@1 / r@5 / r@10 / MRR | file r@1 / r@5 | ms per query")
r = [o.index(i) + 1 for o, i in zip(orders, ids)]
print(f"{'hybrid (no rerank)':36s} {f(r,1):.2f} / {f(r,5):.2f} / {f(r,10):.2f} / {np.mean([1/x for x in r]):.2f}")
for name, maxlen in (("BAAI/bge-reranker-base", 512), ("BAAI/bge-reranker-v2-m3", 512), ("BAAI/bge-reranker-v2-m3", 1024)):
    t0 = time.time(); ce = CrossEncoder(name, max_length=maxlen); load = time.time() - t0
    for N in (10, 20):
        rc, rf, t0 = [], [], time.time()
        for q, i, order in zip(qs, ids, orders):
            head = order[:N]
            sc = ce.predict([(q, ix.docs[j]) for j in head], show_progress_bar=False, batch_size=16)
            new = [head[j] for j in np.argsort(-sc)] + order[N:]
            rc.append(new.index(i) + 1)
            seen = []
            for j in new:
                if file_of[j] not in seen:
                    seen.append(file_of[j])
            rf.append(seen.index(file_of[i]) + 1)
        ms = (time.time() - t0) / len(qs) * 1000
        print(f"{name.split('/')[1]+' L'+str(maxlen)+' top'+str(N):36s} {f(rc,1):.2f} / {f(rc,5):.2f} / {f(rc,10):.2f} / {np.mean([1/x for x in rc]):.2f} | {f(rf,1):.2f} / {f(rf,5):.2f} | {ms:.0f} ms  (load {load:.0f}s)", flush=True)
