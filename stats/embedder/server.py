# /// script
# requires-python = ">=3.10"
# dependencies = ["sentence-transformers==6.3.0", "numpy"]
# ///
"""Embedding sidecar for flowd's hybrid self-review similarity.

Serves BAAI/bge-small-en-v1.5 (384-d, normalized) on 127.0.0.1:
  POST /embed  {"texts": ["..."]} -> {"vectors": [[...], ...]}
  GET  /healthz                   -> 200 "ok"

Run: uv run stats/embedder/server.py [--port 8099]
"""
import argparse
import json
from http.server import BaseHTTPRequestHandler, HTTPServer

from sentence_transformers import SentenceTransformer

model = SentenceTransformer("BAAI/bge-small-en-v1.5")


class Handler(BaseHTTPRequestHandler):
    def _send(self, code, body, ctype="application/json"):
        data = body if isinstance(body, bytes) else body.encode()
        self.send_response(code)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def do_GET(self):
        if self.path == "/healthz":
            self._send(200, "ok", "text/plain")
        else:
            self._send(404, "not found", "text/plain")

    def do_POST(self):
        if self.path != "/embed":
            self._send(404, "not found", "text/plain")
            return
        try:
            n = int(self.headers.get("Content-Length", 0))
            texts = json.loads(self.rfile.read(n))["texts"]
            vectors = model.encode(texts, normalize_embeddings=True, batch_size=32)
        except Exception as e:  # bad body or model failure: report, keep serving
            self._send(400, json.dumps({"error": str(e)}))
            return
        self._send(200, json.dumps({"vectors": vectors.tolist()}))


if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("--port", type=int, default=8099)
    HTTPServer(("127.0.0.1", ap.parse_args().port), Handler).serve_forever()
