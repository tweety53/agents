package store

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Embedder turns texts into embedding vectors, one per text.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

const embedBatch = 64

type httpEmbedder struct {
	url    string
	client *http.Client
}

// NewHTTPEmbedder returns an Embedder calling the embedding sidecar's POST /embed at baseURL,
// 64 texts per request, 60 s per request.
func NewHTTPEmbedder(baseURL string) Embedder {
	return &httpEmbedder{url: strings.TrimRight(baseURL, "/"), client: &http.Client{Timeout: 60 * time.Second}}
}

func (e *httpEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	var out [][]float32
	for i := 0; i < len(texts); i += embedBatch {
		chunk := texts[i:min(i+embedBatch, len(texts))]
		body, _ := json.Marshal(map[string]any{"texts": chunk})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.url+"/embed", bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("embed: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := e.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("embed: %w", err)
		}
		b, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("embed: read: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("embed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
		}
		var r struct {
			Vectors [][]float32 `json:"vectors"`
		}
		if err := json.Unmarshal(b, &r); err != nil {
			return nil, fmt.Errorf("embed: decode: %w", err)
		}
		if len(r.Vectors) != len(chunk) {
			return nil, fmt.Errorf("embed: got %d vectors for %d texts", len(r.Vectors), len(chunk))
		}
		out = append(out, r.Vectors...)
	}
	return out, nil
}

// ConfigureHybrid turns on hybrid (BM25 + embedding) similarity: texts are embedded by e and the
// vectors live in the pgvector database at dsn, opened on first use. Either empty/nil leaves BM25 only.
func (s *Store) ConfigureHybrid(e Embedder, pgvectorDSN string) {
	s.embedder, s.vecDSN = e, pgvectorDSN
}

func (s *Store) hybridConfigured() bool { return s.embedder != nil && s.vecDSN != "" }

// vecPool opens the pgvector pool and its schema once; a failure is retried on the next call.
func (s *Store) vecPool(ctx context.Context) (*pgxpool.Pool, error) {
	s.vecMu.Lock()
	defer s.vecMu.Unlock()
	if s.vec != nil {
		return s.vec, nil
	}
	pool, err := pgxpool.New(ctx, s.vecDSN)
	if err != nil {
		return nil, fmt.Errorf("pgvector: open: %w", err)
	}
	if _, err := pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS vector;
		CREATE TABLE IF NOT EXISTS finding_embeddings (project_key text NOT NULL, finding_id bigint NOT NULL, text_hash text NOT NULL, embedding vector(384) NOT NULL, PRIMARY KEY (project_key, finding_id));`); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pgvector: schema: %w", err)
	}
	s.vec = pool
	return pool, nil
}

func vecLiteral(v []float32) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, f := range v {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(f), 'g', -1, 32))
	}
	b.WriteByte(']')
	return b.String()
}

// vectorRank backfills missing or stale embeddings of the project's findings (ids[i] has text
// docs[i]), embeds query and returns the indexes into ids of the 50 nearest, best first.
func (s *Store) vectorRank(ctx context.Context, projectKey string, ids []int64, docs []string, query string) ([]int, error) {
	pool, err := s.vecPool(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := pool.Query(ctx, `SELECT finding_id, text_hash FROM finding_embeddings WHERE project_key = $1`, projectKey)
	if err != nil {
		return nil, fmt.Errorf("pgvector: read hashes: %w", err)
	}
	have := map[int64]string{}
	for rows.Next() {
		var id int64
		var h string
		if err := rows.Scan(&id, &h); err != nil {
			rows.Close()
			return nil, fmt.Errorf("pgvector: scan: %w", err)
		}
		have[id] = h
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("pgvector: read hashes: %w", err)
	}
	var stale []int
	hashes := make([]string, len(docs))
	for i, d := range docs {
		sum := md5.Sum([]byte(d))
		hashes[i] = hex.EncodeToString(sum[:])
		if have[ids[i]] != hashes[i] {
			stale = append(stale, i)
		}
	}
	if len(stale) > 0 {
		texts := make([]string, len(stale))
		for j, i := range stale {
			texts[j] = docs[i]
		}
		vecs, err := s.embedder.Embed(ctx, texts)
		if err != nil {
			return nil, err
		}
		if len(vecs) != len(stale) {
			return nil, fmt.Errorf("embed: got %d vectors for %d texts", len(vecs), len(stale))
		}
		var b pgx.Batch
		for j, i := range stale {
			b.Queue(`INSERT INTO finding_embeddings (project_key, finding_id, text_hash, embedding) VALUES ($1, $2, $3, $4::vector)
				ON CONFLICT (project_key, finding_id) DO UPDATE SET text_hash = EXCLUDED.text_hash, embedding = EXCLUDED.embedding`,
				projectKey, ids[i], hashes[i], vecLiteral(vecs[j]))
		}
		if err := pool.SendBatch(ctx, &b).Close(); err != nil {
			return nil, fmt.Errorf("pgvector: upsert: %w", err)
		}
	}
	qv, err := s.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	if len(qv) != 1 {
		return nil, fmt.Errorf("embed: got %d vectors for the query", len(qv))
	}
	near, err := pool.Query(ctx, `SELECT finding_id FROM finding_embeddings
		WHERE project_key = $1 AND finding_id = ANY($2) ORDER BY embedding <=> $3::vector LIMIT 50`,
		projectKey, ids, vecLiteral(qv[0]))
	if err != nil {
		return nil, fmt.Errorf("pgvector: nearest: %w", err)
	}
	defer near.Close()
	pos := make(map[int64]int, len(ids))
	for i, id := range ids {
		pos[id] = i
	}
	var out []int
	for near.Next() {
		var id int64
		if err := near.Scan(&id); err != nil {
			return nil, fmt.Errorf("pgvector: scan: %w", err)
		}
		out = append(out, pos[id])
	}
	if err := near.Err(); err != nil {
		return nil, fmt.Errorf("pgvector: nearest: %w", err)
	}
	return out, nil
}
