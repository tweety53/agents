"""Chunk skills/ and rules/ markdown into retrievable sections.

A chunk is one heading's body (headings: #..####, plus a line that is only a bold lead like `**Foo.**`
at column 0 or indented by up to 3 spaces opens a section too); a section over MAX_WORDS is cut at
paragraph breaks. Every chunk keeps file, heading path and 1-based start line.
Run: python chunk.py [<repo root> [<out file>]]  -> chunks.json next to this file
"""
import glob, json, os, re, sys

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = sys.argv[1] if len(sys.argv) > 1 else os.path.abspath(os.path.join(HERE, "..", "..", ".."))
OUT = sys.argv[2] if len(sys.argv) > 2 else os.path.join(HERE, "chunks.json")
MAX_WORDS = 350
HEAD = re.compile(r"^(#{1,4})\s+(.*\S)\s*$")
BOLD = re.compile(r"^\s{0,3}(?:\d+\.\s+)?\*\*([^*]{3,90})\*\*")


def split(path):
    rel = path[len(ROOT) + 1:]
    lines = open(path, encoding="utf-8").read().split("\n")
    sections, stack, cur, start = [], [], [], 1

    def flush():
        body = "\n".join(cur).strip()
        if body:
            sections.append((" > ".join(stack), start, body))

    for i, line in enumerate(lines, 1):
        h = HEAD.match(line)
        b = BOLD.match(line) if not h else None
        if h or b:
            flush()
            cur.clear()
            start = i
            if h:
                depth = len(h.group(1))
                stack[:] = stack[: depth - 1] + [h.group(2)]
                # a heading's own text is part of its chunk
            else:
                stack[:] = stack[:1] + [b.group(1).strip(" .:")]
        cur.append(line)
    flush()
    out = []
    for heading, start, body in sections:
        paras, buf, words, off = body.split("\n\n"), [], 0, 0
        first = start
        for p in paras:
            w = len(p.split())
            if buf and words + w > MAX_WORDS:
                out.append({"file": rel, "heading": heading, "line": first, "text": "\n\n".join(buf)})
                first, buf, words = start + off, [], 0
            buf.append(p)
            words += w
            off += p.count("\n") + 2
        if buf:
            out.append({"file": rel, "heading": heading, "line": first, "text": "\n\n".join(buf)})
    return out


chunks = []
for pat in ("skills/**/*.md", "rules/*.mdc"):
    for f in sorted(glob.glob(f"{ROOT}/{pat}", recursive=True)):
        if "/node_modules/" in f or f.endswith("-rationale.md"):
            continue
        chunks += split(f)
for i, c in enumerate(chunks):
    c["id"] = i
    c["words"] = len(c["text"].split())
json.dump(chunks, open(OUT, "w"), ensure_ascii=False)
ws = sorted(c["words"] for c in chunks)
print(f"{len(chunks)} chunks from {len({c['file'] for c in chunks})} files; words/chunk median {ws[len(ws)//2]}, p90 {ws[int(len(ws)*.9)]}, max {ws[-1]}")
