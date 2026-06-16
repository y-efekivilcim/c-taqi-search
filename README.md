# Taqi Search

A search engine for English vocabulary with the ranking math shown.

Type a word. The Go backend scores it with BM25 and returns the definition, pronunciation, usage examples, and the full scoring breakdown: TF, DF, IDF, and final score displayed as individual values. If you misspell it, a BK-Tree suggests the closest match. The correction is labelled — you always know when Taqi has intervened.

**[→ kivilcimlab.org/taqi-search](https://kivilcimlab.org/taqi-search)**

---

## BM25

BM25 (Best Match 25) is the ranking function used by most production search engines. Unlike raw term frequency, it saturates — a word appearing 100 times in a document doesn't score 100× better than one appearing once. It also accounts for document length.

Taqi exposes every component of the BM25 calculation:

```
λ("conquer")    TF: 1    DF: 1    IDF: 12.416    Score: 62.080
                                                  Σ BM25 = 62.080
```

Most search tools hide this. Taqi treats it as the product.

## Fuzzy correction: BK-Tree

A BK-Tree exploits the triangle inequality of edit distance to prune the search space. For a query word `q` and target `t`: if `d(q, t) > threshold`, then any word `w` where `d(t, w) > d(q, t) + threshold` can be skipped entirely — no comparison needed.

Levenshtein distance is the primary metric. Longest common prefix breaks ties at equal distance.

## Context-aware cancellation

The HTTP server handles cancellation on in-flight requests. If you type faster than results arrive, stale responses are dropped — the result panel always reflects the latest query, never a previous one.

## Stack

- Go — BM25 scorer, BK-Tree, HTTP server with cancellation
- React + TypeScript — frontend
- Public dictionary API — real-time definition enrichment
