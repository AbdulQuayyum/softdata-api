#!/usr/bin/env python3
"""Generate the traceable Nigerian health-facilities snapshot."""

import argparse
import difflib
import hashlib
import json
import re
import unicodedata
from collections import Counter, defaultdict
from pathlib import Path

ID_MAX_LENGTH = 255
REVIEW_PATH = "datasets/metadata/healthcare/health_facilities_identity_reviews.json"
RAW_FIELDS = {
    "facility_name": "raw_name", "globalid": "raw_globalid",
    "nhfr_facility_code": "raw_nhfr_facility_code", "OBJECTID": "raw_object_id",
    "facility_level_option": "raw_facility_type", "facility_level": "raw_facility_level",
    "ownership": "raw_ownership", "ownership_type": "raw_ownership_type",
    "iso": "raw_country", "state": "raw_state", "lga": "raw_lga",
    "ward": "raw_ward_or_town", "latitude": "raw_latitude", "longitude": "raw_longitude",
}

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
        "pagination": "Source positions preserve the supplied combined response order; they are not an OBJECTID sort or observation chronology.",
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


def normalized_attributes(raw):
    """Compare every supplied attribute, not just the facility name.

    OBJECTID is the ArcGIS storage key. The stable source ID, all provenance,
    classification, ownership, location and coordinate attributes remain part
    of equality. Missing and unknown values do not equal known values.
    """
    return {key: clean(value).casefold() if isinstance(value, str) else value
            for key, value in raw.items() if key != "OBJECTID"}


def exact_duplicate(left, right):
    return normalized_attributes(left) == normalized_attributes(right)


def public_id(raw, state_id, lga_id, disambiguator=None):
    parts = [raw.get("facility_name"), state_id, lga_id, disambiguator]
    if not disambiguator:
        parts.append(raw.get("globalid"))
    value = slug("-".join(filter(None, parts)))
    if not value or len(value) > ID_MAX_LENGTH:
        raise ValueError("generated facility ID exceeds the public ID contract")
    return value


def load_reconciliation(repo, reviews):
    """Replay the preserved normalization input without fetching a live dataset.

    The original response byte hash is provenance, not a claim that this
    projection recreates the entire ArcGIS transport response. Full attributes
    for the ten reviewed observations are separately pinned in the review file.
    """
    metadata = json.loads((repo / "datasets/metadata/healthcare/health_facilities.json").read_text())
    index = json.loads((repo / "datasets/metadata/healthcare/health_facilities_reconciliation/index.json").read_text())
    rows = []
    for partition in index["partitions"]:
        data = (repo / "datasets" / partition["path"]).read_bytes()
        if len(data) != partition["size_bytes"] or hashlib.sha256(data).hexdigest() != partition["sha256"]:
            raise ValueError("reconciliation partition hash/size mismatch")
        value = json.loads(data)
        if len(value["records"]) != partition["source_rows"]:
            raise ValueError("reconciliation partition row count mismatch")
        rows.extend(value["records"])
    rows.sort(key=lambda row: row["source_position"])
    if [row["source_position"] for row in rows] != list(range(1, index["source_rows"] + 1)):
        raise ValueError("source positions must be unique and contiguous")
    full = {obs["source_position"]: obs["raw_attributes"]
            for pair in reviews["pairs"] for obs in pair["observations"]}
    features = []
    for row in rows:
        raw = {key: row[field] for key, field in RAW_FIELDS.items()}
        if row["source_position"] in full:
            expected = full[row["source_position"]]
            if any(clean(raw[key]) != clean(expected[key]) for key in RAW_FIELDS):
                raise ValueError("reviewed source attributes disagree with reconciliation")
            raw = expected.copy()
        features.append({"attributes": raw})
    return features, metadata["source"]


def reviewed_decisions(features, state_ids, lga_by_state, reviews):
    """Fail closed on an unreviewed identity collision or a stale review."""
    decisions = {}
    for pair in reviews["pairs"]:
        if pair["final_decision"] not in {"exclude_unresolved_identity", "retain_separate_facility"}:
            raise ValueError("this review requires an explicit supported resolution")
        if not pair.get("evidence_sources") or not pair.get("reason"):
            raise ValueError("identity review must include evidence and a reason")
        for obs in pair["observations"]:
            position = obs["source_position"]
            if position in decisions or position < 1 or position > len(features):
                raise ValueError("invalid or overlapping review positions")
            raw = features[position - 1]["attributes"]
            # Compare the complete pinned observation when full attributes are
            # supplied, and every preserved field in projection-only input.
            expected = obs["raw_attributes"]
            if any(key not in expected or clean(value) != clean(expected[key]) for key, value in raw.items()):
                raise ValueError("source changed since identity review")
            decisions[position] = pair
    groups = defaultdict(list)
    source_groups = defaultdict(list)
    for position, feature in enumerate(features, 1):
        raw = feature["attributes"]
        if raw.get("globalid"):
            source_groups[clean(raw["globalid"])].append(position)
        state = slug(raw.get("state"))
        lga, _ = resolve_lga(state, raw.get("lga"), lga_by_state)
        if state not in state_ids or not clean(raw.get("facility_name")) or (raw.get("lga") and not lga):
            continue
        key = record_identity(raw["facility_name"], state, lga, raw.get("ward"))
        groups[key].append(position)
    for positions in list(groups.values()) + list(source_groups.values()):
        if len(positions) < 2:
            continue
        first = features[positions[0] - 1]["attributes"]
        if all(exact_duplicate(first, features[pos - 1]["attributes"]) for pos in positions[1:]):
            continue
        pairs = [decisions.get(pos) for pos in positions]
        if any(pair is None for pair in pairs) or len({pair["pair_id"] for pair in pairs}) != 1:
            raise ValueError(f"conflicting facility observations require identity review: {positions}")
    return decisions


def generate(source_path, repo, retrieved_at, reconciliation_repo=None, review_path=None):
    review_path = Path(review_path) if review_path else Path(__file__).resolve().parents[1] / REVIEW_PATH
    reviews = json.loads(review_path.read_text())
    if reconciliation_repo is not None:
        features, source = load_reconciliation(Path(reconciliation_repo), reviews)
    else:
        source_path = Path(source_path)
        features = json.loads(source_path.read_text())["features"]
        source = source_info(source_path, retrieved_at)
    source["raw_record_count"] = len(features)
    source["pagination"] = "Archived source_position order preserved from the combined response; not an OBJECTID sort or observation chronology."
    state_ids, lga_by_state = load_geography(repo)
    decisions = reviewed_decisions(features, state_ids, lga_by_state, reviews)

    records = []
    reconciliation = defaultdict(list)
    seen_exact = {}
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
        review = decisions.get(position)
        if review and review["final_decision"] == "exclude_unresolved_identity":
            decision = "exclude_unresolved_identity"
            notes.append(f"Identity review {review['pair_id']}: {review['reason']}")
        elif state_id not in state_ids or not raw_name:
            decision = "exclude_invalid_geography"
            notes.append("Missing name or state does not match the existing Nigerian states snapshot.")
        elif raw.get("lga") and not lga_id:
            decision = "exclude_invalid_geography"
            notes.append("Source LGA did not match the stated state and was not safely mappable to the existing LGA snapshot.")
        else:
            if lga_note:
                notes.append(lga_note)
            disambiguator = None
            if review:
                decision = "retain_separate_facility"
                disambiguator = review.get("disambiguators", {}).get(source_id)
            final_id = public_id(raw, state_id, lga_id, disambiguator)
            signature = json.dumps(normalized_attributes(raw), sort_keys=True, ensure_ascii=False)
            if signature in seen_exact:
                decision = "merge_exact_duplicate"
                merge_target = seen_exact[signature]
                final_id = None
                notes.append("All normalized source attributes, including stable source identity and coordinates, equal an earlier observation.")
            else:
                seen_exact[signature] = final_id
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
            "raw_latitude": raw.get("latitude"),
            "raw_longitude": raw.get("longitude"),
            "normalized_name": raw_name if final_id else None,
            "normalized_state_id": state_id if state_id in state_ids else None,
            "normalized_lga_id": lga_id,
            "decision": decision,
            "final_record_id": final_id,
            "merge_target_id": merge_target,
            "evidence": [SOURCE_LAYER_URL] + ([REVIEW_PATH + "#" + review["pair_id"]] if review else []),
            "notes": " ".join(notes) or "Retained from the dated GRID3 HFR-derived row-level snapshot.",
        })

    ids = [item["id"] for item in records]
    if len(set(ids)) != len(ids):
        raise ValueError("public IDs must be unique")
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
            "Unresolved geography and reviewed identity conflicts are excluded with row-level evidence; name equality does not establish an exact duplicate.",
            "NHFR facility codes are not globally unique in this layer; GRID3 globalid is retained as the source row identifier.",
        ],
        "licensing": "Source ownership remains with GRID3/CIESIN and its credited contributors. SoftData claims only its independent normalization, schema, identifiers, reconciliation, and metadata; public availability does not transfer source ownership.",
        "identity_review_path": "metadata/healthcare/health_facilities_identity_reviews.json",
        "generation": {
            "input": "preserved reconciliation attributes with pinned full GRID3 observations for reviewed pairs",
            "input_projection_sha256": hashlib.sha256(json.dumps(
                [{key: feature["attributes"].get(key) for key in RAW_FIELDS} for feature in features],
                ensure_ascii=False, sort_keys=True, separators=(",", ":")
            ).encode()).hexdigest(),
            "identity_reviews_sha256": hashlib.sha256(review_path.read_bytes()).hexdigest(),
            "original_response_hash_note": "source.sha256 identifies the original archived response; reconciliation replay does not reconstruct its transport bytes.",
        },
        "verified_at": reviews["review_date"],
    }
    (metadata_dir / "health_facilities.json").write_text(json.dumps(metadata, ensure_ascii=False, indent=2) + "\n")
    schema_path = repo / "datasets/schemas/healthcare/health_facilities.schema.json"
    schema_source = Path(__file__).resolve().parents[1] / "datasets/schemas/healthcare/health_facilities.schema.json"
    schema = json.loads(schema_source.read_text())
    schema["minItems"] = schema["maxItems"] = len(records)
    schema["$defs"]["healthFacility"]["properties"]["id"]["maxLength"] = ID_MAX_LENGTH
    schema_path.parent.mkdir(parents=True, exist_ok=True)
    schema_path.write_text(json.dumps(schema, ensure_ascii=False, indent=2) + "\n")
    return len(features), len(records), counters


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    inputs = parser.add_mutually_exclusive_group(required=True)
    inputs.add_argument("--source", help="Combined ArcGIS JSON query response in the reviewed source-position order")
    inputs.add_argument("--from-reconciliation", help="Repository containing the committed source evidence to replay offline")
    parser.add_argument("--reviews", help="Pinned per-pair identity decisions; defaults to the repository review manifest")
    parser.add_argument("--repo", default=".")
    parser.add_argument("--retrieved-at", required=True, help="Fixed retrieval date for deterministic output")
    args = parser.parse_args()
    print(generate(args.source, Path(args.repo), args.retrieved_at, args.from_reconciliation, args.reviews))
