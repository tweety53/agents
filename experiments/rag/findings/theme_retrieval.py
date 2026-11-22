"""RAG experiment: do findings that share a theme retrieve each other? Standalone; reads the dev stats API."""
import json, re, sys, urllib.request, numpy as np
from collections import Counter
from sklearn.feature_extraction.text import TfidfVectorizer
from sentence_transformers import SentenceTransformer

API = "http://127.0.0.1:4173/api/v1/stats/self-review-findings"
def fetch():
    rows = []
    for show, status in (("fixed", None), ("to-review", "approved"), ("to-review", "declined"), ("to-review", "to-review")):
        off = 0
        while True:
            q = f"from=2000-01-01T00:00:00Z&to=2027-12-31T00:00:00Z&show={show}&limit=100&offset={off}" + (f"&status={status}" if status else "")
            p = json.load(urllib.request.urlopen(f"{API}?{q}"))
            rows += p["rows"]; off += 100
            if off >= p["total"]: break
    return rows

rows = [r for r in fetch() if r["theme"] and r["theme"] != "uncategorized"]
cnt = Counter(r["theme"] for r in rows)
rows = [r for r in rows if cnt[r["theme"]] >= 2]          # a theme needs a second member to be retrievable
def title(r): return re.sub(r"^\[theme: [a-z0-9-]+\]\s*", "", r["note"])
def sec(r, name):
    m = re.search(rf"## {name}\n(.*?)(?=\n## |\Z)", r["detail"] or "", re.S)
    return " ".join(m.group(1).split()) if m else ""
variants = {
    "note": [title(r) for r in rows],
    "note+detail": [title(r) + ". " + " ".join((r["detail"] or "").split())[:600] for r in rows],
}
labels = np.array([r["theme"] for r in rows]); n = len(rows)
print(f"{n} findings, {len(set(labels))} themes (themes with >=2 members; uncategorized excluded)")
big = {t for t, c in Counter(labels).items() if c > 60}

def evaluate(sim, name, mask=None):
    np.fill_diagonal(sim, -1)
    idx = np.where(mask)[0] if mask is not None else np.arange(n)
    hit5 = p5 = mrr = 0
    for i in idx:
        order = np.argsort(-sim[i]); same = labels[order] == labels[i]
        hit5 += same[:5].any(); p5 += same[:5].mean()
        first = np.argmax(same); mrr += 1 / (first + 1)
    k = len(idx); print(f"  {name:34s} hit@5 {hit5/k:.2f}  P@5 {p5/k:.2f}  MRR {mrr/k:.2f}  (n={k})")

# random baseline: expected P@5 if retrieval ignored theme
rnd = np.mean([(Counter(labels)[t] - 1) / (n - 1) for t in labels]); print(f"random P@5 ≈ {rnd:.2f}")
small = np.array([l not in big for l in labels])
model = SentenceTransformer("BAAI/bge-small-en-v1.5")
for vname, texts in variants.items():
    print(vname)
    tf = TfidfVectorizer(stop_words="english", sublinear_tf=True).fit_transform(texts); S = (tf @ tf.T).toarray()
    evaluate(S.copy(), "tf-idf, all themes"); evaluate(S.copy(), "tf-idf, themes <=60 rows", small)
    E = model.encode(texts, normalize_embeddings=True, show_progress_bar=False); S = E @ E.T
    evaluate(S.copy(), "bge-small, all themes"); evaluate(S.copy(), "bge-small, themes <=60 rows", small)
