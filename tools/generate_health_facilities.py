#!/usr/bin/env python3
"""Generate the traceable Nigerian health-facilities snapshot."""

import argparse
import datetime as dt
import difflib
import hashlib
import json
import re
import unicodedata
from collections import Counter, defaultdict
from pathlib import Path

SOURCE_URL = (
    "https://services3.arcgis.com/BU6Aadhn6tbBEdyk/ArcGIS/rest/services/"
    "GRID3_NGA_health_facilities_v2_0/FeatureServer/0/query"
)
SOURCE_LAYER_URL = (
    "https://services3.arcgis.com/BU6Aadhn6tbBEdyk/ArcGIS/rest/services/"
    "GRID3_NGA_health_facilities_v2_0/FeatureServer/0"
)

TYPE_MAP = {
    "Health Post": "health-post",
    "Primary Health Center": "primary-health-centre",
    "Primary Health Clinic": "clinic",
    "General Hospital": "general-hospital",
    "Teaching/Tertiary Hospital": "teaching-hospital",
    "Specialized Hospital": "specialist-hospital",
    "Unknown": "other",
}

STATE_ALIASES = {"fct": "fct"}

# These are source spelling or naming variants that map unambiguously to the
# existing geography snapshot. Other mismatches are quarantined.
LGA_ALIASES = {
    ("fct", "abuja-municipal-area-council"): "fct-abuja-municipal",
    ("ekiti", "ado-ekiti"): "ekiti-ado",
    ("ogun", "sagamu"): "ogun-shagamu",
    ("plateau", "barkin-ladi"): "plateau-barikin-ladi",
    ("bayelsa", "yenagoa"): "bayelsa-yenegoa",
    ("niger", "paikoro"): "niger-pailoro",
    ("benue", "otukpo"): "benue-oturkpo",
    ("cross-river", "calabar-municipal"): "cross-river-calabar-municipality",
    ("kogi", "olamaboro"): "kogi-olamabolo",
    ("abia", "obi-ngwa"): "abia-obi-nwa",
    ("cross-river", "obubra"): "cross-river-odubra",
    ("gombe", "yamaltu-deba"): "gombe-yamaltu-delta",
    ("osun", "ayedaade"): "osun-aiyedade",
    ("cross-river", "bekwarra"): "cross-river-bekwara",
    ("bayelsa", "southern-ijaw"): "bayelsa-southern-jaw",
    ("osun", "ilesa-west"): "osun-ilesha-west",
    ("osun", "ilesa-east"): "osun-ilesha-east",
    ("cross-river", "yakurr"): "cross-river-yarkur",
    ("kaduna", "zangon-kataf"): "kaduna-zango-kataf",
    ("katsina", "matazu"): "katsina-matazuu",
    ("edo", "uhunmwode"): "edo-uhunmwonde",
    ("imo", "ezinihitte-mbaise"): "imo-ezinihitte",
    ("gombe", "nafada"): "gombe-nafada-bajoga",
    ("kano", "danbatta"): "kano-dambatta",
    ("delta", "ukwuani"): "delta-ukwani",
    ("yobe", "busari"): "yobe-bursari",
    ("jigawa", "birniwa"): "jigawa-biriniwa",
    ("kano", "garum-mallam"): "kano-garun-mallam",
    ("gombe", "shongom"): "gombe-shomgom",
    ("ekiti", "aiyekire-gbonyin"): "ekiti-gbonyin",
    ("kebbi", "danko-wasagu"): "kebbi-wasagu-danko",
}


def clean(value):
    if value is None:
        return None
    value = unicodedata.normalize("NFKC", str(value)).replace("\xa0", " ")
    return re.sub(r"\s+", " ", value).strip()


def slug(value):
    value = unicodedata.normalize("NFKD", clean(value) or "")
    value = value.encode("ascii", "ignore").decode().lower()
    return re.sub(r"[^a-z0-9]+", "-", value).strip("-")


def source_info(path, retrieved_at):
    data = Path(path).read_bytes()
    return {
        "publisher": "Center for International Earth Science Information Network (CIESIN), Columbia University / GRID3",
        "title": "GRID3 NGA - Health Facilities v2.0",
        "url": SOURCE_LAYER_URL,
        "query_endpoint": SOURCE_URL,
        "retrieved_at": retrieved_at,
        "response_bytes": len(data),
        "sha256": hashlib.sha256(data).hexdigest(),
        "format": "ArcGIS FeatureServer JSON",
        "pagination": "OBJECTID ASC, resultOffset/resultRecordCount; retrieved in deterministic 500-record pages",
        "raw_record_count": None,
        "last_updated": "2024-11-11",
        "terms": "The service provides citation and copyright information but no machine-readable licence field; preserve source attribution and consult GRID3/CIESIN terms before redistribution.",
    }


def load_geography(repo):
    states = json.loads((repo / "datasets/geography/states.json").read_text())
    lgas = json.loads((repo / "datasets/geography/lgas.json").read_text())
    state_ids = {item["id"] for item in states}
    lga_by_state = defaultdict(dict)
    for item in lgas:
        lga_by_state[item["state_id"]][slug(item["name"])] = item["id"]
    return state_ids, lga_by_state


def resolve_lga(state_id, raw_lga, lga_by_state):
    if not raw_lga:
        return None, None
    source_slug = slug(raw_lga)
    exact = lga_by_state[state_id].get(source_slug)
    if exact:
        return exact, "exact geography name"
    alias = LGA_ALIASES.get((state_id, source_slug))
    if alias and alias in lga_by_state[state_id].values():
        return alias, "documented source spelling/name variant"
    candidates = []
    for name_slug, lga_id in lga_by_state[state_id].items():
        candidates.append((difflib.SequenceMatcher(None, source_slug, name_slug).ratio(), lga_id))
    candidates.sort(reverse=True)
    if candidates and candidates[0][0] >= 0.93 and (
        len(candidates) == 1 or candidates[0][0] - candidates[1][0] >= 0.08
    ):
        return candidates[0][1], "high-confidence source spelling variant"
    return None, None


def normalize_type(raw_type):
    return TYPE_MAP.get(clean(raw_type), "other")


def normalize_ownership(raw_ownership, raw_type):
    ownership = clean(raw_ownership)
    category = clean(raw_type)
    if ownership == "Unknown":
        return None
    if ownership == "Public":
        if category == "Local Government":
            return "local-government"
        if category == "State Government":
            return "state"
        if category == "Federal Government":
            return "federal"
        if category == "Military & Paramilitary formations":
            return "military"
        return "other-public"
    if ownership == "Private":
        return "private"
    return "other"


def coordinate(value, minimum, maximum):
    if value is None:
        return None
    value = float(value)
    if not (minimum <= value <= maximum):
        return None
    return value


def record_identity(name, state_id, lga_id, ward):
    return "|".join((slug(name), state_id, lga_id or "", slug(ward)))


def generate(source_path, repo, retrieved_at):
    source_path = Path(source_path)
    payload = json.loads(source_path.read_text())
    features = payload["features"]
    state_ids, lga_by_state = load_geography(repo)
    source = source_info(source_path, retrieved_at)
    source["raw_record_count"] = len(features)

    records = []
    reconciliation = defaultdict(list)
    seen_identity = {}
    counters = Counter()
    for position, feature in enumerate(features, 1):
        raw = feature["attributes"]
        raw_name = clean(raw.get("facility_name"))
        state_id = slug(raw.get("state"))
        lga_id, lga_note = resolve_lga(state_id, raw.get("lga"), lga_by_state)
        source_id = clean(raw.get("globalid"))
        decision = "retain"
        final_id = None
        merge_target = None
        notes = []
        if state_id not in state_ids or not raw_name:
            decision = "exclude_invalid_geography"
            notes.append("Missing name or state does not match the existing Nigerian states snapshot.")
        elif raw.get("lga") and not lga_id:
            decision = "exclude_invalid_geography"
            notes.append("Source LGA did not match the stated state and was not safely mappable to the existing LGA snapshot.")
        else:
            if lga_note:
                notes.append(lga_note)
            final_id = slug("-".join(filter(None, [raw_name, state_id, lga_id, raw.get("globalid")])))
            identity = record_identity(raw_name, state_id, lga_id, raw.get("ward"))
            if identity in seen_identity:
                decision = "merge_exact_duplicate"
                merge_target = seen_identity[identity]
                final_id = None
                notes.append("Exact normalized facility identity already retained from an earlier source row.")
            else:
                seen_identity[identity] = final_id
                output = {
                    "id": final_id,
                    "name": raw_name,
                    "facility_type": normalize_type(raw.get("facility_level_option")),
                    "state_id": state_id,
                    "country_code": "NG",
                    "source_facility_id": source_id,
                }
                facility_level = clean(raw.get("facility_level"))
                if facility_level in {"Primary", "Secondary", "Tertiary"}:
                    output["facility_level"] = facility_level.lower()
                ownership = normalize_ownership(raw.get("ownership"), raw.get("ownership_type"))
                if ownership:
                    output["ownership_type"] = ownership
                if lga_id:
                    output["lga_id"] = lga_id
                lat = coordinate(raw.get("latitude"), 4.281710, 13.865239)
                lon = coordinate(raw.get("longitude"), 2.707790, 14.636383)
                if lat is not None and lon is not None:
                    output["latitude"] = lat
                    output["longitude"] = lon
                records.append(output)

        counters[decision] += 1
        reconciliation[state_id if state_id in state_ids else "unresolved"].append({
            "source_position": position,
            "source_facility_id": source_id,
            "raw_nhfr_facility_code": clean(raw.get("nhfr_facility_code")),
            "raw_globalid": source_id,
            "raw_object_id": raw.get("OBJECTID"),
            "raw_name": raw_name,
            "raw_facility_type": clean(raw.get("facility_level_option")),
            "raw_facility_level": clean(raw.get("facility_level")),
            "raw_ownership": clean(raw.get("ownership")),
            "raw_ownership_type": clean(raw.get("ownership_type")),
            "raw_country": clean(raw.get("iso")),
            "raw_state": clean(raw.get("state")),
            "raw_lga": clean(raw.get("lga")),
            "raw_ward_or_town": clean(raw.get("ward")),
            "raw_status": None,
            "raw_latitude": raw.get("latitude"),
            "raw_longitude": raw.get("longitude"),
            "normalized_name": raw_name if final_id else None,
            "normalized_state_id": state_id if state_id in state_ids else None,
            "normalized_lga_id": lga_id,
            "decision": decision,
            "final_record_id": final_id,
            "merge_target_id": merge_target,
            "evidence": [SOURCE_LAYER_URL],
            "notes": " ".join(notes) or "Retained from the dated GRID3 HFR-derived row-level snapshot.",
        })

    records.sort(key=lambda item: (item["state_id"], item.get("lga_id", ""), item["name"].casefold(), item["id"]))
    healthcare = repo / "datasets/healthcare"
    metadata_dir = repo / "datasets/metadata/healthcare"
    reconciliation_dir = metadata_dir / "health_facilities_reconciliation"
    healthcare.mkdir(parents=True, exist_ok=True)
    reconciliation_dir.mkdir(parents=True, exist_ok=True)
    dataset_path = healthcare / "health_facilities.json"
    dataset_path.write_text(json.dumps(records, ensure_ascii=False, indent=2) + "\n")

    partition_index = []
    for state_id in sorted(reconciliation):
        partition_path = reconciliation_dir / f"{state_id}.json"
        partition = {
            "dataset_key": "ng-health-facilities",
            "state_id": state_id,
            "source_rows": len(reconciliation[state_id]),
            "decision_counts": dict(sorted(Counter(item["decision"] for item in reconciliation[state_id]).items())),
            "records": reconciliation[state_id],
        }
        partition_path.write_text(json.dumps(partition, ensure_ascii=False, indent=2) + "\n")
        data = partition_path.read_bytes()
        partition_index.append({
            "state_id": state_id,
            "path": f"metadata/healthcare/health_facilities_reconciliation/{state_id}.json",
            "source_rows": len(reconciliation[state_id]),
            "size_bytes": len(data),
            "sha256": hashlib.sha256(data).hexdigest(),
        })

    index_path = reconciliation_dir / "index.json"
    index_path.write_text(json.dumps({
        "dataset_key": "ng-health-facilities",
        "source_rows": len(features),
        "final_record_count": len(records),
        "decision_counts": dict(sorted(counters.items())),
        "partitions": partition_index,
    }, ensure_ascii=False, indent=2) + "\n")

    ownership_counts = Counter(item.get("ownership_type", "<omitted>") for item in records)
    type_counts = Counter(item["facility_type"] for item in records)
    level_counts = Counter(item.get("facility_level", "<omitted>") for item in records)
    state_counts = Counter(item["state_id"] for item in records)
    lga_counts = Counter(item.get("lga_id", "<omitted>") for item in records)
    metadata = {
        "dataset_key": "ng-health-facilities",
        "status": "active",
        "snapshot": True,
        "publicly_published": True,
        "title": "Nigeria Health Facilities (GRID3 HFR-derived 2024 Snapshot)",
        "description": "Source-verified row-level health facility snapshot derived from the public GRID3 NGA Health Facilities v2.0 layer, which documents NHFR 2024 inputs. It is not a continuously current live registry and should not be treated as a completeness claim for all active facilities.",
        "country_code": "NG",
        "record_count": len(records),
        "source_rows": len(features),
        "decision_counts": dict(sorted(counters.items())),
        "facility_type_counts": dict(sorted(type_counts.items())),
        "facility_level_counts": dict(sorted(level_counts.items())),
        "ownership_counts": dict(sorted(ownership_counts.items())),
        "state_coverage_count": len(state_counts),
        "lga_coverage_count": len([key for key in lga_counts if key != "<omitted>"]),
        "coordinate_coverage_count": sum(1 for item in records if "latitude" in item and "longitude" in item),
        "relative_path": "healthcare/health_facilities.json",
        "schema_path": "schemas/healthcare/health_facilities.schema.json",
        "source": source,
        "reconciliation_path": "metadata/healthcare/health_facilities_reconciliation/index.json",
        "source_limitations": [
            "The NHFR external API is API-key protected; the public portal documents it as the canonical registry interface.",
            "This published row-level snapshot uses the public GRID3 layer whose metadata says it incorporates NHFR 2024 inputs.",
            "Rows with unresolved state/LGA geography and exact normalized duplicates are excluded and retained in reconciliation metadata.",
            "NHFR facility codes are not globally unique in this layer; GRID3 globalid is retained as the source row identifier.",
        ],
        "licensing": "Source ownership remains with GRID3/CIESIN and its credited contributors. SoftData claims only its independent normalization, schema, identifiers, reconciliation, and metadata; public availability does not transfer source ownership.",
        "verified_at": retrieved_at,
    }
    (metadata_dir / "health_facilities.json").write_text(json.dumps(metadata, ensure_ascii=False, indent=2) + "\n")
    return len(features), len(records), counters


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--source", required=True, help="Combined ArcGIS JSON query response")
    parser.add_argument("--repo", default=".")
    parser.add_argument("--retrieved-at", required=True, help="Fixed retrieval date for deterministic output")
    args = parser.parse_args()
    print(generate(args.source, Path(args.repo), args.retrieved_at))
