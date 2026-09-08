#!/usr/bin/env python3
"""Generate traceable education snapshots from downloaded first-party sources."""

import argparse
import datetime as dt
import hashlib
import html
import json
import re
import subprocess
import unicodedata
from difflib import SequenceMatcher
from pathlib import Path

STATES = {
    "abia": (1, 1), "adamawa": (2, 6), "akwa-ibom": (7, 11), "anambra": (12, 19),
    "bayelsa": (20, 20), "bauchi": (21, 27), "benue": (28, 30), "cross-river": (31, 32),
    "delta": (33, 38), "edo": (39, 43), "ekiti": (44, 46), "enugu": (47, 48),
    "gombe": (49, 50), "imo": (51, 54), "jigawa": (55, 58), "kaduna": (59, 62),
    "kano": (63, 66), "katsina": (67, 71), "kebbi": (72, 74), "kwara": (75, 78),
    "lagos": (79, 84), "nasarawa": (85, 88), "niger": (89, 93), "ogun": (94, 100),
    "ondo": (101, 105), "osun": (106, 107), "plateau": (108, 108), "rivers": (109, 113),
    "sokoto": (114, 118), "taraba": (119, 119), "yobe": (120, 120), "zamfara": (121, 121),
    "fct": (122, 123),
}

def clean(value):
    value = html.unescape(value).replace("\xa0", " ")
    value = unicodedata.normalize("NFKC", value)
    return re.sub(r"\s+", " ", value).strip(" .")

def slug(value):
    value = unicodedata.normalize("NFKD", value).encode("ascii", "ignore").decode().lower()
    value = re.sub(r"[^a-z0-9]+", "-", value).strip("-")
    return value[:120].strip("-")

def source_info(path, url):
    data = Path(path).read_bytes()
    return {"url": url, "accessed_at": dt.date.today().isoformat(), "response_bytes": len(data),
            "sha256": hashlib.sha256(data).hexdigest()}

def parse_nmcn(path):
    if str(path).lower().endswith(".pdf"):
        text = subprocess.check_output(["pdftotext", "-layout", str(path), "-"]).decode("utf-8", "replace")
    else:
        text = Path(path).read_text(errors="replace")
    rows, current, section = [], None, None
    for page, page_text in enumerate(text.split("\f"), 1):
      for line in page_text.splitlines():
        match = re.match(r"^\s*(\d+\.\d+)\s+(.*)$", line)
        if match:
            if current:
                rows.append(current)
            if match.group(1).endswith(".0"):
                section = clean(match.group(2))
                current = None
                continue
            current = {"source_page": page, "printed_row_number": match.group(1), "section_heading": section, "lines": [clean(match.group(2))]}
        elif current and line.strip() and not line.strip().startswith("|"):
            current["lines"].append(clean(line))
    if current:
        rows.append(current)
    result = []
    for row in rows:
        lines = [x for x in row["lines"] if x and x != "Page"]
        programmes = [x.lstrip("- ") for x in lines if x.startswith(("", "-"))]
        institution = clean(" ".join(x for x in lines if not x.startswith(("", "-"))))
        result.append({"source_page": row["source_page"], "printed_row_number": row["printed_row_number"],
                       "section_heading": row["section_heading"], "raw_institution_name": institution,
                       "raw_location": None, "raw_programme": programmes or None, "raw_lines": lines})
    return result

def nmcn_match(row, candidates):
    raw = re.sub(r"[^a-z0-9 ]", " ", row["raw_institution_name"].lower())
    raw_tokens = {x for x in raw.split() if len(x) > 2 and x not in {"college", "school", "department", "nursing", "sciences"}}
    best, score = None, 0
    for candidate in candidates:
        name = candidate["name"].lower()
        tokens = {x for x in re.sub(r"[^a-z0-9 ]", " ", name).split() if len(x) > 2 and x not in {"college", "school", "department", "nursing", "sciences"}}
        overlap = len(raw_tokens & tokens) / max(1, len(tokens))
        similarity = SequenceMatcher(None, raw, re.sub(r"[^a-z0-9 ]", " ", name)).ratio()
        value = max(overlap, similarity)
        if value > score:
            best, score = candidate, value
    return (best, score) if score >= 0.58 else (None, score)

def generate_nursing(pdf, dataset_path, out_path):
    candidates = json.loads(Path(dataset_path).read_text())
    rows = parse_nmcn(pdf)
    used, records = set(), []
    reconciliation = []
    for row in rows:
        match, score = nmcn_match(row, candidates)
        if match and match["id"] not in used:
            used.add(match["id"])
            reconciliation.append({**row, "normalized_name": match["name"], "state_id": match["state_id"],
                "ownership_type": match["ownership_type"], "decision": "retain", "final_record_id": match["id"],
                "merge_target_id": None, "evidence": ["https://nmcn.gov.ng/docs/New_List_of_Approved_Schools_Dec_2025.pdf"],
                "notes": "Matched to the existing stable candidate by source text; programme lines remain in raw_programme.", "match_score": round(score, 4)})
            records.append(match)
        else:
            reconciliation.append({**row, "normalized_name": None, "state_id": None, "ownership_type": None,
                "decision": "exclude_unresolved", "final_record_id": None, "merge_target_id": None,
                "evidence": ["https://nmcn.gov.ng/docs/New_List_of_Approved_Schools_Dec_2025.pdf"],
                "notes": "The extracted row could not be mapped to a distinct public institution without guessing.", "match_score": round(score, 4)})
    records.sort(key=lambda x: (x["name"].lower(), x["id"]))
    Path(out_path).write_text(json.dumps({"dataset_key": "ng-colleges-of-nursing-and-midwifery", "status": "active",
        "source_rows": len(rows), "programme_merges": 0, "wrapped_row_merges": 0, "exact_duplicate_merges": 0,
        "wrong_category_exclusions": 0, "unresolved_exclusions": len(rows) - len(records),
        "published_records": len(records), "records": reconciliation}, ensure_ascii=False, indent=2) + "\n")
    Path(dataset_path).write_text(json.dumps(records, ensure_ascii=False, indent=2) + "\n")
    source = source_info(pdf, "https://nmcn.gov.ng/docs/New_List_of_Approved_Schools_Dec_2025.pdf")
    counts, states = {}, set()
    for record in records:
        counts[record["ownership_type"]] = counts.get(record["ownership_type"], 0) + 1
        states.add(record["state_id"])
    Path(Path(dataset_path).parent.parent / "metadata/education/colleges_of_nursing_and_midwifery.json").write_text(json.dumps({
        "dataset_key": "ng-colleges-of-nursing-and-midwifery", "status": "active", "snapshot": True,
        "publicly_published": True, "title": "Nigeria Colleges of Nursing and Midwifery (NMCN December 2025 Snapshot)",
        "description": "Traceable institution and campus snapshot derived from the NMCN December 2025 approved-schools register; not a complete live national register.",
        "country_code": "NG", "record_count": len(records), "ownership_counts": counts, "state_coverage_count": len(states),
        "relative_path": "education/colleges_of_nursing_and_midwifery.json", "schema_path": "schemas/education/colleges_of_nursing_and_midwifery.schema.json",
        "source": source, "official_webpage_aggregate": 290, "source_rows": len(rows), "unresolved_exclusions": len(rows) - len(records),
        "coverage_limitation": "The source publishes programme rows and repeated numbering; rows not mapped to a distinct institution are explicitly excluded.",
        "reconciliation_path": "metadata/education/colleges_of_nursing_and_midwifery_reconciliation.json", "verified_at": dt.date.today().isoformat()
    }, ensure_ascii=False, indent=2) + "\n")
    return len(rows), len(records), reconciliation

def table_cells(fragment):
    cells = []
    for cell in re.findall(r"<td[^>]*>(.*?)</td>", fragment, re.I | re.S):
        cells.append(clean(re.sub(r"<[^>]+>", " ", cell)))
    return cells

def technical_state(number):
    for state, (first, last) in STATES.items():
        if first <= number <= last:
            return state
    return None

def generate_technical(page, output, metadata, reconciliation):
    source = source_info(page, "https://web.nbte.gov.ng/technical%20colleges1")
    records, decisions = [], []
    for fragment in re.findall(r"<tr[^>]*>(.*?)</tr>", Path(page).read_text(errors="replace"), re.I | re.S):
        cells = table_cells(fragment)
        if len(cells) < 2 or not re.fullmatch(r"\d+\.?", cells[0]):
            continue
        number = int(cells[0].rstrip(".")); name = cells[1]; owner_text = " ".join(cells[2:])
        combined = f"{name} {owner_text}".lower()
        decision = "retain"
        if any(x in combined for x in ["vocational technical training", "craft development", "secondary school"]):
            decision = "exclude_wrong_category"
        elif "academy" in name.lower() or "training centres" in name.lower() or "training centers" in name.lower():
            decision = "exclude_wrong_category"
        state = technical_state(number)
        if decision == "retain" and not state:
            decision = "exclude_unresolved"
        ownership = "federal" if "federal" in combined else "private" if "private" in combined else "state"
        final_id = slug(name) if decision == "retain" else None
        row = {"source_page": None, "printed_row_number": number, "section_heading": "Technical Colleges",
               "raw_institution_name": name, "raw_location": owner_text or None, "raw_programme": None,
               "normalized_name": name if decision == "retain" else None, "state_id": state,
               "ownership_type": ownership if decision == "retain" else None, "decision": decision,
               "final_record_id": final_id, "merge_target_id": None, "evidence": [source["url"]],
               "notes": "NBTE directory row; the homepage aggregate is tracked separately."}
        decisions.append(row)
        if decision == "retain":
            records.append({"id": final_id, "name": name, "ownership_type": ownership, "state_id": state, "country_code": "NG"})
    records.sort(key=lambda x: (x["name"].lower(), x["id"]))
    Path(output).write_text(json.dumps(records, ensure_ascii=False, indent=2) + "\n")
    Path(reconciliation).write_text(json.dumps({"dataset_key": "ng-technical-colleges", "status": "active",
        "source": source, "official_aggregate_total": 153, "row_level_directory_entries": 123,
        "records": decisions}, ensure_ascii=False, indent=2) + "\n")
    counts = {}
    for r in records: counts[r["ownership_type"]] = counts.get(r["ownership_type"], 0) + 1
    Path(metadata).write_text(json.dumps({"dataset_key": "ng-technical-colleges", "status": "active",
        "snapshot": True, "publicly_published": True, "title": "Nigeria Technical Colleges (NBTE Directory Snapshot)",
        "description": "Dated row-level snapshot of the official NBTE Technical Colleges directory; not a complete live national register.",
        "country_code": "NG", "record_count": len(records), "ownership_counts": counts,
        "official_aggregate_total": 153, "row_level_directory_entries": 123,
        "coverage_limitation": "NBTE separately reports 153 institutions while this directory exposes 123 numbered entries; unidentified institutions are not inferred.",
        "relative_path": "education/technical_colleges.json", "schema_path": "schemas/education/technical_colleges.schema.json",
        "sources": [source], "reconciliation_path": "metadata/education/technical_colleges_reconciliation.json",
        "verified_at": dt.date.today().isoformat()}, ensure_ascii=False, indent=2) + "\n")
    return len(decisions), len(records)

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--nmcn-pdf", required=True)
    parser.add_argument("--nbte-html", required=True)
    parser.add_argument("--repo", default=".")
    args = parser.parse_args(); root = Path(args.repo)
    print("nursing", generate_nursing(args.nmcn_pdf, root / "datasets/education/colleges_of_nursing_and_midwifery.json", root / "datasets/metadata/education/colleges_of_nursing_and_midwifery_reconciliation.json")[:2])
    print("technical", generate_technical(args.nbte_html, root / "datasets/education/technical_colleges.json", root / "datasets/metadata/education/technical_colleges.json", root / "datasets/metadata/education/technical_colleges_reconciliation.json"))
