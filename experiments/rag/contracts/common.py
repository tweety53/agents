"""Shared by the evaluation scripts: questions are stored as (file, opening of the source chunk), so they
survive the repository changing; this maps each back to a chunk id in the current chunks.json."""
import json, os

HERE = os.path.dirname(os.path.abspath(__file__))


def load_questions(chunks):
    """Returns (ids, questions, kinds); a question whose source chunk no longer exists is skipped and named."""
    ids, qs, kinds = [], [], []
    for item in json.load(open(f"{HERE}/questions.json")):
        norm = lambda t: " ".join(t.split())
        hit = [c["id"] for c in chunks if c["file"] == item["file"] and norm(c["text"]).startswith(item["snippet"])]
        if not hit:
            print(f"skipped (source chunk gone): {item['file']} -- {item['question']}")
            continue
        ids.append(hit[0])
        qs.append(item["question"])
        kinds.append(item["kind"])
    return ids, qs, kinds
