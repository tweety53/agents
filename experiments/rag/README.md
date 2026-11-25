# RAG experiments

Standalone, read-only experiments on retrieval over this repository's own text. Nothing here is wired into
a `/flow*` run, a guard or the stats service. The production piece (`flow self-review similar`, hybrid BM25 +
pgvector) lives in `stats/`; this directory holds what was measured to choose it.

```bash
cd experiments/rag
uv venv --python 3.12 .venv && uv pip install --python .venv/bin/python -r requirements.txt
```

## contracts/ — search over `skills/**/*.md` and `rules/*.mdc`

| Script | Does |
|---|---|
| `chunk.py` | Splits the corpus by heading and bold lead (max 350 words a chunk, `-rationale.md` files skipped) into `chunks.json`. |
| `search.py "<question>" [-k 5] [--no-rerank]` | Hybrid BM25 + `bge-small-en-v1.5` (RRF), top 20 reranked by `bge-reranker-base`. Embeddings cached in `emb.npy`, keyed by a hash of the chunk texts. |
| `evaluate.py` | Scores bm25, bge-small, hybrid and weighted hybrids, then the reranker, on `questions.json`; lists near-duplicate chunk pairs. |
| `rerank_cmp.py` | `bge-reranker-base` against `bge-reranker-v2-m3` (512 and 1024 tokens), top 10 and 20. |
| `questions.json` | 59 model-written questions, each stored as its source file plus the chunk's opening, so it survives the corpus changing. |

Measured on those 59 questions (1,343 chunks, 87 files), chunk recall at 1 / 5 / 10 and file recall at 1 / 5:

| Method | chunk r@1 | r@5 | r@10 | file r@1 | file r@5 | per query |
|---|---|---|---|---|---|---|
| bm25 | 0.47 | 0.75 | 0.77 | 0.63 | 0.87 | |
| bge-small | 0.37 | 0.72 | 0.83 | 0.55 | 0.83 | |
| hybrid RRF | 0.50 | 0.75 | 0.85 | 0.60 | 0.92 | |
| hybrid + `bge-reranker-base`, top 20 | 0.60 | 0.85 | 0.88 | 0.70 | 0.95 | ~0.8 s |
| hybrid + `bge-reranker-v2-m3` 1024, top 20 | 0.65 | 0.83 | 0.90 | 0.73 | 0.95 | ~4.4 s |

`bge-reranker-base` over the top 20 is the chosen default: v2-m3 is inside the noise of 59 questions and about
six times slower. Three sonnet agents given only grep, glob and read answered the same questions at chunk
r@1 0.57 and r@3 0.63, against 0.60 and 0.78 for the chosen method.

## findings/ — search over recorded self-review findings

`theme_retrieval.py` reads the dev stats API and measures whether findings sharing a theme tag retrieve each
other. `rerank_eval.py <data dir> [--v2]` takes a snapshot of the findings (`corpus.pkl`, `questions.json`; not
committed, they are the project's own finding text) and compares hybrid with and without a reranker on question
lookup and on theme retrieval. Offline runs need `HF_HUB_OFFLINE=1`, or a Hub check can hang.

Question lookup, 40 questions over 817 findings (recall at 1 / 5 / 10, MRR):

| Method | r@1 | r@5 | r@10 | MRR | per query |
|---|---|---|---|---|---|
| hybrid | 0.88 | 0.93 | 0.95 | 0.91 | |
| + `bge-reranker-base`, top 20 | 0.85 | 0.95 | 0.97 | 0.89 | 208 ms |
| + `bge-reranker-v2-m3`, top 20 | 0.93 | 0.97 | 0.97 | 0.95 | 676 ms |

Theme retrieval, a finding's title as the query, hit@5 / P@5 / MRR over all 815 themed findings: hybrid
0.65 / 0.26 / 0.47, base top 20 0.65 / 0.25 / 0.44, v2-m3 top 20 0.64 / 0.26 / 0.45. A theme tag is a judgement
about cause, not about wording, so no ranker recovers it, and a reranker adds nothing on either task: hybrid
already ranks the right finding first 88% of the time, and v2-m3's few points are one or two questions of 40.
