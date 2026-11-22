"""Contract search: hybrid BM25 + bge-small retrieval over chunks.json, reranked by bge-reranker-base over the top 20.
Read-only; not wired into any run.

  python search.py "what happens if the review panel is skipped" [-k 5] [--no-rerank]
Embeddings are cached in emb.npy, keyed by a hash of the chunk texts, so a changed corpus re-embeds.
"""
import hashlib, json, os, re, sys
import numpy as np
from rank_bm25 import BM25Okapi

HERE = os.path.dirname(os.path.abspath(__file__))
tok = lambda t: re.findall(r"[a-z0-9]+", t.lower())


def doc(c):
    return c["heading"] + ". " + c["text"]


class Index:
    def __init__(self, chunks=None):
        self.chunks = chunks or json.load(open(f"{HERE}/chunks.json"))
        self.docs = [doc(c) for c in self.chunks]
        self.bm = BM25Okapi([tok(d) for d in self.docs])
        self._model = None
        key = hashlib.md5("\x00".join(self.docs).encode()).hexdigest()
        cache = f"{HERE}/emb.npy"
        if os.path.exists(cache) and os.path.exists(cache + ".key") and open(cache + ".key").read() == key:
            self.E = np.load(cache)
        else:
            self.E = self.model.encode(self.docs, normalize_embeddings=True, batch_size=32, show_progress_bar=False)
            np.save(cache, self.E)
            open(cache + ".key", "w").write(key)

    def rerank(self, q, order, n=20):
        """Reorder the first n of `order` by the cross-encoder; the rest keep their place."""
        if getattr(self, "_ce", None) is None:
            from sentence_transformers import CrossEncoder
            self._ce = CrossEncoder("BAAI/bge-reranker-base", max_length=512)
        head = list(order[:n])
        sc = self._ce.predict([(q, self.docs[j]) for j in head], show_progress_bar=False, batch_size=16)
        return [head[j] for j in np.argsort(-sc)] + list(order[n:])

    def search(self, q, k=5, rerank=True):
        order = self.rank(q, "hybrid", k=max(k, 20))
        return (self.rerank(q, order) if rerank else order)[:k]

    @property
    def model(self):
        if self._model is None:
            from sentence_transformers import SentenceTransformer
            self._model = SentenceTransformer("BAAI/bge-small-en-v1.5")
        return self._model

    def scores(self, q):
        b = self.bm.get_scores(tok(q))
        v = self.E @ self.model.encode([q], normalize_embeddings=True, show_progress_bar=False)[0]
        return b, v

    def rank(self, q, mode="hybrid", k=5, w=1.0):
        b, v = self.scores(q)
        rk = lambda s: np.argsort(np.argsort(-s)) + 1
        s = {"bm25": b, "vec": v, "hybrid": w / (60 + rk(b)) + 1 / (60 + rk(v))}[mode]
        return list(np.argsort(-s)[:k])


if __name__ == "__main__":
    a = sys.argv[1:]
    k = int(a[a.index("-k") + 1]) if "-k" in a else 5
    rerank = "--no-rerank" not in a
    q = " ".join(x for i, x in enumerate(a) if x not in ("-k", "--no-rerank") and (i == 0 or a[i - 1] != "-k"))
    ix = Index()
    for r, i in enumerate(ix.search(q, k=k, rerank=rerank), 1):
        c = ix.chunks[i]
        snippet = " ".join(c["text"].split())[:150]
        print(f"{r}. {c['file']}:{c['line']}  [{c['heading'][:80]}]\n   {snippet}")
