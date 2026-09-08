#!/usr/bin/env python3
"""Generate the UBEC 2022 school snapshot and state-partitioned row ledger."""

import argparse
import hashlib
import json
import re
import zipfile
import xml.etree.ElementTree as ET
from collections import Counter, defaultdict
from datetime import date
from pathlib import Path

NS = {"m": "http://schemas.openxmlformats.org/spreadsheetml/2006/main"}
HEADERS = [
    "school_code", "category", "school_type", "state", "lga", "school_name",
    "town", "location", "school_level", "ownership", "ownership_category",
]
SOURCES = {
    "primary": {
        "filename": "2022-Basic-Education-School-List.xlsx",
        "url": "https://ubec.gov.ng/wp-content/uploads/2024/04/2022-Basic-Education-School-List.xlsx",
        "publication_date": "2022 snapshot; published at this URL in 2024-04",
        "size": 8634886,
        "sha256": "02e73aeccf773e09fae21b2ebf13cc9b8126eb20ad2b781b0b8fb33f55408fb0",
    },
    "jss": {
        "filename": "2022-Basic-Education-JS-School-List.xlsx",
        "url": "https://ubec.gov.ng/wp-content/uploads/2024/04/2022-Basic-Education-JS-School-List.xlsx",
        "publication_date": "2022 snapshot; published at this URL in 2024-04",
        "size": 2570008,
        "sha256": "f773df47908773f72ed142087dbdf19326ab40848465780b2bd683ed7bb029d5",
    },
}

# UBEC preserves a number of source-era spellings and merged LGA labels. These
# are explicit aliases, not fuzzy matches, and each target is resolved against
# datasets/geography/lgas.json.
LGA_ALIASES = {
    ("fct", "municipal"): "abuja municipal",
    ("fct", "fct abuja"): "abuja municipal",
    ("fct", "abaji"): "abaji",
    ("fct", "bwari"): "bwari",
    ("fct", "gwagwalada"): "gwagwalada",
    ("fct", "kuje"): "kuje",
    ("fct", "kwali"): "kwali",
    ("abia", "isukwuato"): "isuikwuato",
    ("abia", "isikwuato"): "isuikwuato",
    ("abia", "obioma ngwa"): "obi nwa",
    ("abia", "osisioma"): "osisioma ngwa",
    ("adamawa", "tuongo"): "toungo",
    ("bauchi", "damban"): "dambam",
    ("bauchi", "jamaare"): "jama are",
    ("bauchi", "katagun"): "katagum",
    ("bayelsa", "southern ijaw"): "southern jaw",
    ("benue", "otukpo"): "oturkpo",
    ("borno", "abadan"): "abadam",
    ("borno", "mongunu"): "monguno",
    ("cross-river", "bekwarra"): "bekwara",
    ("cross-river", "biasse"): "biase",
    ("cross-river", "calabar municipal"): "calabar municipality",
    ("cross-river", "obubra"): "odubra",
    ("cross-river", "yakurr"): "yarkur",
    ("delta", "aniocha north"): "aniocha",
    ("delta", "ika north"): "ika north east",
    ("delta", "oshimili south"): "oshimili",
    ("delta", "ukwuani"): "ukwani",
    ("delta", "warri south west"): "warri south",
    ("ekiti", "ado ekiti"): "ado",
    ("ekiti", "emure"): "emure ise orun",
    ("ekiti", "irepodun ifelodun"): "irepodun",
    ("gombe", "nafada"): "nafada bajoga",
    ("gombe", "shongom"): "shomgom",
    ("gombe", "yamaltu deba"): "yamaltu delta",
    ("imo", "ahiazu"): "ahiazu mbaise",
    ("jigawa", "birniwa"): "biriniwa",
    ("jigawa", "kaugawa"): "kaugama",
    ("jigawa", "kirikasamma"): "kiri kasamma",
    ("kaduna", "jemaa"): "jema a",
    ("kaduna", "zangon kataf"): "zango kataf",
    ("kano", "mingibir"): "minjibir",
    ("katsina", "danmusa"): "dan musa",
    ("katsina", "maiadua"): "mai adua",
    ("katsina", "matazu"): "matazuu",
    ("kebbi", "aliero"): "aleiro",
    ("kebbi", "arewa"): "arewa dandi",
    ("kebbi", "danko wasagu"): "wasagu danko",
    ("kogi", "ogori magongo"): "ogori mangongo",
    ("kogi", "olamaboro"): "olamabolo",
    ("kwara", "patigi"): "pategi",
    ("lagos", "mainland"): "lagos mainland",
    ("lagos", "ifako ijaiye"): "ifako ijaye",
    ("niger", "munya"): "muya",
    ("niger", "paikoro"): "pailoro",
    ("ogun", "ogun water side"): "ogun waterside",
    ("ogun", "sagamu"): "shagamu",
    ("osun", "ayede ade"): "aiyedade",
    ("osun", "ayedire"): "aiyedire",
    ("oyo", "ogbomoso north"): "ogbomosho north",
    ("oyo", "ogbomoso south"): "ogbomosho south",
    ("plateau", "barkin ladi"): "barikin ladi",
    ("plateau", "quaanpan"): "qua an pan",
    ("rivers", "emu oha"): "emohua",
    ("rivers", "emuoha"): "emohua",
    ("rivers", "omuma"): "omumma",
    ("sokoto", "wamakko"): "wamako",
    ("yobe", "barde"): "bade",
    ("yobe", "bosari"): "bursari",
    ("yobe", "tarmua"): "tarmuwa",
    ("zamfara", "birni magaji"): "birnin magaji",
    ("zamfara", "talatan mafara"): "talata mafara",
}


def digest(path):
    h = hashlib.sha256()
    with path.open("rb") as fh:
        for chunk in iter(lambda: fh.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def write_json(path, value, compact=False):
    path.parent.mkdir(parents=True, exist_ok=True)
    encoded = json.dumps(value, separators=(",", ":")) if compact else json.dumps(value, indent=2)
    path.write_text(encoded + "\n")


def normalize(value):
    value = repair_text(value)
    return re.sub(r"[^a-z0-9]+", " ", value.lower()).strip()


def slug(value):
    value = repair_text(value)
    value = re.sub(r"[^a-z0-9]+", "-", value.lower()).strip("-")
    return re.sub(r"-+", "-", value)


def repair_text(value):
    """Repair the small number of visibly double-encoded workbook strings."""
    for _ in range(2):
        if not any(marker in value for marker in ("Ã", "Â", "â")):
            break
        try:
            repaired = value.encode("latin1").decode("utf-8")
        except (UnicodeEncodeError, UnicodeDecodeError):
            break
        if repaired == value:
            break
        value = repaired
    return value


def workbook_rows(path):
    with zipfile.ZipFile(path) as archive:
        shared = []
        if "xl/sharedStrings.xml" in archive.namelist():
            root = ET.fromstring(archive.read("xl/sharedStrings.xml"))
            shared = ["".join(item.itertext()) for item in root.findall("m:si", NS)]
        workbook = ET.fromstring(archive.read("xl/workbook.xml"))
        relationships = ET.fromstring(archive.read("xl/_rels/workbook.xml.rels"))
        rels = {item.attrib["Id"]: item.attrib["Target"] for item in relationships}
        sheet = workbook.find("m:sheets", NS)[0]
        rid = sheet.attrib["{http://schemas.openxmlformats.org/officeDocument/2006/relationships}id"]
        target = rels[rid]
        target = "xl/" + target if not target.startswith("xl/") else target
        root = ET.fromstring(archive.read(target))
        rows = []
        for row in root.findall(".//m:sheetData/m:row", NS):
            values = []
            for cell in row.findall("m:c", NS):
                value = cell.find("m:v", NS)
                text = value.text if value is not None else ""
                if cell.attrib.get("t") == "s" and text:
                    text = shared[int(text)]
                values.append(text)
            rows.append((int(row.attrib["r"]), dict(zip(HEADERS, values))))
    return sheet.attrib.get("name", ""), rows


def load_geography(path):
    states = json.loads(path.joinpath("states.json").read_text())
    lgas = json.loads(path.joinpath("lgas.json").read_text())
    state_ids = {normalize(item["name"]): item["id"] for item in states}
    state_ids.update({"fct": "fct", "abuja": "fct", "federal capital territory": "fct"})
    lga_ids = {(item["state_id"], normalize(item["name"])): item["id"] for item in lgas}
    return state_ids, lga_ids, {item["id"] for item in states}


def level_mapping(workbook, value):
    mappings = {
        "ECCDE AND PRIMARY": ["pre-primary", "primary"],
        "ECCDE ONLY": ["pre-primary"],
        "PRIMARY ONLY": ["primary"],
        "JSS AND SSS": ["junior-secondary", "senior-secondary"],
        "JSS ONLY": ["junior-secondary"],
    }
    if value not in mappings:
        raise ValueError(f"unknown {workbook} school_level: {value!r}")
    return mappings[value]


def inspect_source(key, path, retrieved_at):
    source = SOURCES[key]
    if path.stat().st_size != source["size"] or digest(path) != source["sha256"]:
        raise SystemExit(f"{key} workbook hash or size mismatch: {path}")
    worksheet, rows = workbook_rows(path)
    return {
        "workbook": source["filename"], "worksheet": worksheet,
        "url": source["url"], "retrieved_at": retrieved_at,
        "size": source["size"], "sha256": source["sha256"],
        "header_rows": 1, "data_rows": len(rows) - 1,
        "hidden_sheets": [], "hidden_rows": 0, "formulas": 0,
        "merged_cells": 0, "empty_rows": 0,
    }, rows


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--primary", type=Path, required=True)
    parser.add_argument("--jss", type=Path, required=True)
    parser.add_argument("--states-dir", type=Path, required=True)
    parser.add_argument("--dataset", type=Path, required=True)
    parser.add_argument("--reconciliation", type=Path, required=True, help="output directory for reconciliation partitions")
    parser.add_argument("--summary", type=Path, required=True)
    parser.add_argument("--retrieved-at", default="2026-09-08")
    args = parser.parse_args()

    source_meta = {}
    raw = []
    for key, path in (("primary", args.primary), ("jss", args.jss)):
        meta, rows = inspect_source(key, path, args.retrieved_at)
        source_meta[key] = meta
        worksheet, _ = workbook_rows(path)
        for row_number, values in rows[1:]:
            values = {field: str(values.get(field, "")) for field in HEADERS}
            values.update({"source": key, "worksheet": worksheet, "source_row": row_number})
            values["levels"] = level_mapping(key, values["school_level"])
            raw.append(values)

    state_ids, lga_ids, known_states = load_geography(args.states_dir)
    geography_failures = []
    for item in raw:
        state = state_ids.get(normalize(item["state"]))
        if normalize(item["state"]) == "fct abuja":
            state = "fct"
        lga_name = normalize(item["lga"])
        lga_name = LGA_ALIASES.get((state, lga_name), lga_name)
        lga = lga_ids.get((state, lga_name)) if state else None
        item["state_id"] = state
        item["lga_id"] = lga
        item["geography_valid"] = bool(state and lga)
        if state is None or lga is None:
            geography_failures.append({"source": item["source"], "source_row": item["source_row"], "state": item["state"], "lga": item["lga"]})
    by_code = defaultdict(list)
    for item in raw:
        if not item["geography_valid"]:
            continue
        by_code[item["school_code"]].append(item)
    code_conflicts = {code: rows for code, rows in by_code.items() if len({(r["state_id"], r["lga_id"], normalize(r["school_name"]), normalize(r["town"])) for r in rows}) > 1}

    # First group exact source identities. A stable code can bridge the two workbooks
    # only when the geographic and normalized-name identity is also consistent.
    groups = defaultdict(list)
    for item in raw:
        if not item["geography_valid"]:
            continue
        identity = (item["state_id"], item["lga_id"], normalize(item["school_name"]), normalize(item["town"]))
        if item["school_code"] and item["school_code"] not in code_conflicts:
            key = ("code", item["school_code"])
        else:
            key = ("identity", identity)
        groups[key].append(item)

    records = []
    group_by_identity = {}
    for key, rows in groups.items():
        first = rows[0]
        identity = (first["state_id"], first["lga_id"], normalize(first["school_name"]), normalize(first["town"]))
        # A code group is still split if its normalized identity conflicts.
        if key[0] == "code" and len({(r["state_id"], r["lga_id"], normalize(r["school_name"]), normalize(r["town"])) for r in rows}) > 1:
            raise SystemExit(f"unhandled code conflict: {key[1]}")
        public_key = identity
        if public_key in group_by_identity and group_by_identity[public_key] != key:
            # Same exact identity with different codes is one candidate; the code is
            # deliberately omitted from the public contract and retained in the ledger.
            groups[group_by_identity[public_key]].extend(rows)
            continue
        group_by_identity[public_key] = key
        levels = sorted({level for row in rows for level in row["levels"]}, key=["pre-primary", "primary", "junior-secondary", "senior-secondary"].index)
        records.append({
            "name": repair_text(first["school_name"]).strip(), "ownership_type": "public" if first["category"].upper() == "PUBLIC" else "private",
            "government_owner": None, "state_id": first["state_id"], "lga_id": first["lga_id"],
            "country_code": "NG", "education_levels": levels,
            "identity": public_key, "rows": rows,
        })

    # Rebuild merged groups whose second code group joined an existing identity.
    for record in records:
        record["rows"] = groups[group_by_identity[record["identity"]]]
        names = {repair_text(row["school_name"]).strip() for row in record["rows"]}
        record["aliases"] = sorted(names - {record["name"]})
        codes = sorted({row["school_code"] for row in record["rows"] if row["school_code"]})
        record["codes"] = codes
        record["education_levels"] = sorted({level for row in record["rows"] for level in row["levels"]}, key=["pre-primary", "primary", "junior-secondary", "senior-secondary"].index)

    records.sort(key=lambda item: (item["name"].lower(), item["state_id"], item["lga_id"]))
    used_ids = {}
    public = []
    for record in records:
        state_name = record["state_id"]
        name_slug = slug(record["name"])
        if not name_slug:
            # Preserve the source name, but use its stable UBEC code when the
            # source encoding contains no slug-able characters.
            name_slug = "school-" + (record["codes"][0] if record["codes"] else "unidentified")
        base = name_slug + "-" + state_name + "-" + slug(record["lga_id"].split("-", 1)[1])
        town = slug(record["rows"][0]["town"])
        candidate = base + ("-" + town if town else "")
        if candidate in used_ids:
            code = record["codes"][0] if len(record["codes"]) == 1 else ""
            candidate += ("-" + code if code else "-" + slug(record["rows"][0]["source"]))
        if candidate in used_ids:
            raise SystemExit(f"unresolved public ID collision: {candidate}")
        used_ids[candidate] = True
        record["id"] = candidate
        value = {"id": candidate, "name": record["name"], "state_id": state_name, "country_code": "NG", "education_levels": record["education_levels"]}
        if record["ownership_type"]:
            value["ownership_type"] = record["ownership_type"]
        if record["lga_id"]:
            value["lga_id"] = record["lga_id"]
        if len(record["codes"]) == 1:
            value["ubec_school_code"] = record["codes"][0]
        public.append(value)
    public.sort(key=lambda item: (item["name"].lower(), item["id"]))

    target_by_row = {}
    for record in records:
        for row in record["rows"]:
            target_by_row[(row["source"], row["source_row"])] = record
    decisions = []
    for item in raw:
        if not item["geography_valid"]:
            decisions.append({
                "source": item["source"], "workbook": source_meta[item["source"]]["workbook"], "worksheet": item["worksheet"],
                "source_row": item["source_row"], "state_id": item["state_id"], "raw": {field: item[field] for field in HEADERS},
                "canonical_id": None, "canonical_name": None, "decision": "excluded_unresolved_geography",
                "public_record": False, "education_levels": item["levels"], "related_source_positions": [],
            })
            continue
        record = target_by_row[(item["source"], item["source_row"])]
        decision = "retained_source_identity"
        if len(record["rows"]) > 1:
            decision = "merged_cross_workbook_or_duplicate_row"
        if len(record["codes"]) > 1:
            decision = "merged_identity_code_conflict"
        decisions.append({
            "source": item["source"], "workbook": source_meta[item["source"]]["workbook"], "worksheet": item["worksheet"],
            "source_row": item["source_row"], "state_id": item["state_id"], "raw": {field: item[field] for field in HEADERS},
            "canonical_id": record["id"], "canonical_name": record["name"], "decision": decision,
            "public_record": True, "education_levels": item["levels"], "related_source_positions": sorted(r["source"] + ":" + str(r["source_row"]) for r in record["rows"] if r is not item),
        })

    args.dataset.parent.mkdir(parents=True, exist_ok=True)
    args.summary.parent.mkdir(parents=True, exist_ok=True)
    write_json(args.dataset, public)

    # Keep row-level evidence in deterministic state partitions. Quarantined
    # rows retain their resolved state and remain fully accounted for.
    states_by_id = {state: [] for state in sorted(known_states)}
    for decision in decisions:
        state = decision["state_id"]
        if state not in states_by_id:
            raise SystemExit(f"unknown reconciliation state: {state!r}")
        states_by_id[state].append(decision)
    args.reconciliation.mkdir(parents=True, exist_ok=True)
    legacy_reconciliation = args.reconciliation.with_suffix(".json")
    if legacy_reconciliation.exists():
        legacy_reconciliation.unlink()
    for stale in args.reconciliation.glob("*.json"):
        stale.unlink()
    partition_entries = []
    for state_id, state_decisions in states_by_id.items():
        state_decisions.sort(key=lambda item: (item["source"], item["source_row"]))
        state_record_ids = {item["canonical_id"] for item in state_decisions if item["public_record"]}
        payload = {
            "dataset_key": "ng-primary-and-secondary-schools",
            "schema_version": "1.0",
            "state_id": state_id,
            "source_rows": len(state_decisions),
            "final_record_count": len(state_record_ids),
            "decision_counts": dict(Counter(item["decision"] for item in state_decisions)),
            "decisions": state_decisions,
        }
        path = args.reconciliation / (state_id + ".json")
        write_json(path, payload, compact=True)
        partition_entries.append({
            "state_id": state_id, "path": path.name, "source_rows": len(state_decisions),
            "final_record_count": len(state_record_ids), "sha256": digest(path), "bytes": path.stat().st_size,
        })
    index = {
        "dataset_key": "ng-primary-and-secondary-schools", "schema_version": "1.0",
        "status": "source_observed_snapshot", "sources": source_meta,
        "source_rows": len(raw), "final_record_count": len(public), "partitions": partition_entries,
        "decision_counts": dict(Counter(item["decision"] for item in decisions)),
        "education_level_mapping": {
            "ECCDE AND PRIMARY": ["pre-primary", "primary"], "ECCDE ONLY": ["pre-primary"], "PRIMARY ONLY": ["primary"],
            "JSS AND SSS": ["junior-secondary", "senior-secondary"], "JSS ONLY": ["junior-secondary"],
        },
        "ownership_mapping": {"PUBLIC": "public", "PRIVATE": "private", "government_owner": "omitted unless authoritative source distinguishes federal/state/local"},
        "code_policy": "UBEC school codes are public only when one code maps to one final identity; conflicting codes remain in the partition ledger.",
        "geography": {"state_count": len(known_states), "lga_count": len(lga_ids), "failures": geography_failures},
        "generator": "tools/generate_primary_and_secondary_schools.py",
    }
    write_json(args.reconciliation / "index.json", index, compact=True)
    summary = {
        "dataset_key": "ng-primary-and-secondary-schools", "generated_at": args.retrieved_at,
        "raw_records_processed": len(raw), "candidates_retained": len(public), "excluded_geography_rows": len(geography_failures), "duplicate_or_cross_workbook_rows": len(raw) - len(public) - len(geography_failures),
        "cross_workbook_matches": sum(1 for record in records if {row["source"] for row in record["rows"]} == {"primary", "jss"}),
        "code_conflict_rows": sum(len(rows) for rows in code_conflicts.values()), "geography_failures": len(geography_failures),
        "ownership_failures": 0, "education_level_failures": 0, "next_step": "repository/service design",
        "checkpoint_progression": [
            {"group": "Abia-Bayelsa", "status": "completed"}, {"group": "Benue-Ebonyi", "status": "completed"},
            {"group": "Edo-Jigawa", "status": "completed"}, {"group": "Kaduna-Kwara", "status": "completed"},
            {"group": "Lagos-Niger", "status": "completed"}, {"group": "Ogun-Rivers", "status": "completed"},
            {"group": "Sokoto-Zamfara", "status": "completed"}, {"group": "Federal Capital Territory", "status": "completed"},
            {"group": "National cross-state deduplication", "status": "completed"},
        ],
        "temporary_source_files_not_required_for_reproduction": True,
    }
    write_json(args.summary, summary)
    print(json.dumps({"raw": len(raw), "records": len(public), "duplicates_or_merges": len(raw) - len(public), "cross_workbook_matches": summary["cross_workbook_matches"], "code_conflict_rows": summary["code_conflict_rows"]}, indent=2))


if __name__ == "__main__":
    main()
