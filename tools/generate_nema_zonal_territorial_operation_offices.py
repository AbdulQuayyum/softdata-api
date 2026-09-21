#!/usr/bin/env python3
"""Generate the NEMA zonal, territorial and operation offices snapshot."""

from __future__ import annotations

import argparse
import hashlib
import json
import re
from collections import Counter
from pathlib import Path
from typing import Any

DATASET_KEY = "ng-nema-zonal-territorial-operation-offices"
TITLE = "Nigeria NEMA Zonal, Territorial and Operation Offices Snapshot"
COUNTRY_CODE = "NG"
RETRIEVED_AT = "2026-09-21"
OFFICE_TYPE = "zonal_territorial_operation_office"

SOURCES: dict[str, dict[str, Any]] = {
    "nema-flood-operations-press-release": {
        "canonical_url": "https://nema.gov.ng/flood-incidents-nema-activates-operational-offices-nationwide-to-support-states-conduct-assessments/",
        "retrieval_url": "https://nema.gov.ng/wp-json/wp/v2/posts/15629",
        "publisher": "National Emergency Management Agency",
        "title": "FLOOD INCIDENTS: NEMA ACTIVATES OPERATIONAL OFFICES NATIONWIDE TO SUPPORT STATES, CONDUCT ASSESSMENTS",
        "publication_date": "2024-07-07",
        "modification_date": "2024-07-07",
        "retrieval_date": RETRIEVED_AT,
        "retrievals": [
            {
                "http_status": 200,
                "content_type": "application/json; charset=UTF-8",
                "byte_size": 5860,
                "sha256": "a3e1df80fd1f32ff8d0cbb658229e25b018dd08f7c3f6a5f1bc509697dd4598c",
                "last_modified": None,
                "etag": None,
            },
            {
                "http_status": 200,
                "content_type": "application/json; charset=UTF-8",
                "byte_size": 5860,
                "sha256": "a3e1df80fd1f32ff8d0cbb658229e25b018dd08f7c3f6a5f1bc509697dd4598c",
                "last_modified": None,
                "etag": None,
            },
        ],
        "row_count": 17,
        "pagination": "Single WordPress post JSON object; no pagination or export endpoint.",
        "reuse_terms": "No explicit reuse terms found on the source page.",
        "extraction_method": "Manual safe extraction of the paragraph listing NEMA Zonal, Territorial and Operation office locations; raw WordPress HTML/JSON was not committed.",
        "geographic_coverage": "Nationwide list of office locations published in a NEMA press release.",
        "stability": "Repeated retrieval of the public WordPress JSON representation was byte-identical; the HTML page varied because it includes dynamic WordPress content.",
    },
    "nema-about-nema": {
        "canonical_url": "https://nema.gov.ng/about-nema/",
        "retrieval_url": "https://nema.gov.ng/wp-json/wp/v2/pages/2131",
        "publisher": "National Emergency Management Agency",
        "title": "About NEMA",
        "retrieval_date": RETRIEVED_AT,
        "retrievals": [
            {
                "http_status": 200,
                "content_type": "application/json; charset=UTF-8",
                "byte_size": 12690,
                "sha256": "f46e4106cb53897d2b49cda5b3dc838faeff29f0fbc96c57ec9460e5861b10c9",
                "last_modified": None,
                "etag": None,
            }
        ],
        "row_count": 6,
        "pagination": "Single WordPress page JSON object; no pagination or export endpoint.",
        "reuse_terms": "No explicit reuse terms found on the source page.",
        "extraction_method": "Supporting terminology extraction only: the page states NEMA has six zonal offices and lists their geopolitical-zone locations.",
        "geographic_coverage": "Six geopolitical-zone zonal offices.",
        "stability": "Retrieved public WordPress JSON representation once for supporting terminology and source metadata.",
    },
    "nema-zonal-territorial-offices": {
        "canonical_url": "https://nema.gov.ng/zonal-territorial-offices/",
        "retrieval_url": "https://nema.gov.ng/wp-json/wp/v2/pages/17130",
        "publisher": "National Emergency Management Agency",
        "title": "Zonal & Territorial Offices",
        "retrieval_date": RETRIEVED_AT,
        "retrievals": [
            {
                "http_status": 200,
                "content_type": "application/json; charset=UTF-8",
                "byte_size": 38950,
                "sha256": "ab1615a1a976600cbb5e28890b89f69dc14a248beabd3cc1ba8a37c40c77bd0a",
                "last_modified": None,
                "etag": None,
            }
        ],
        "row_count": 20,
        "pagination": "Single WordPress page JSON object; no pagination or export endpoint.",
        "reuse_terms": "No explicit reuse terms found on the source page.",
        "extraction_method": "Inspected only as supporting context for NEMA office terminology and locations; staff names and personal phone numbers were not retained.",
        "geographic_coverage": "NEMA staff-contact table covering zonal, territorial and operation-office labels.",
        "stability": "Retrieved public WordPress JSON representation once; page combines office labels with staff contacts and is not used as public row-level evidence.",
        "privacy_note": "The page contains staff names and phone numbers. They are excluded from all generated artifacts.",
    },
}

LOCATION_ROWS = [
    ("Lagos", "lagos", "NEMA Lagos Office"),
    ("Ibadan", "oyo", "NEMA Ibadan Office"),
    ("Ekiti", "ekiti", "NEMA Ekiti Office"),
    ("Abuja", "fct", "NEMA Abuja Office"),
    ("Minna", "niger", "NEMA Minna Office"),
    ("Jos", "plateau", "NEMA Jos Office"),
    ("Enugu", "enugu", "NEMA Enugu Office"),
    ("Owerri", "imo", "NEMA Owerri Office"),
    ("Port Harcourt", "rivers", "NEMA Port Harcourt Office"),
    ("Edo", "edo", "NEMA Edo Office"),
    ("Uyo", "akwa-ibom", "NEMA Uyo Office"),
    ("Kano", "kano", "NEMA Kano Office"),
    ("Sokoto", "sokoto", "NEMA Sokoto Office"),
    ("Kaduna", "kaduna", "NEMA Kaduna Office"),
    ("Maiduguri", "borno", "NEMA Maiduguri Office"),
    ("Yola", "adamawa", "NEMA Yola Office"),
    ("Gombe", "gombe", "NEMA Gombe Office"),
]

STATE_ALIASES = {location: state_id for location, state_id, _ in LOCATION_ROWS}
DECISIONS = {"retain", "merge_exact_duplicate", "exclude_unresolved_identity", "exclude_invalid_geography", "exclude_non_office_entry", "exclude_personal_contact_only"}
STATE_IDS = {state_id for _, state_id, _ in LOCATION_ROWS}


def slug(value: str) -> str:
    normalized = re.sub(r"[^a-z0-9]+", "-", value.lower()).strip("-")
    normalized = re.sub(r"-+", "-", normalized)
    if not normalized:
        raise ValueError("empty slug")
    return normalized


def canonical_json(value: Any) -> str:
    return json.dumps(value, ensure_ascii=False, indent=2, sort_keys=True) + "\n"


def build_records() -> list[dict[str, str]]:
    records = []
    for location, state_id, name in LOCATION_ROWS:
        records.append(
            {
                "id": f"nema-{slug(location)}-zonal-territorial-operation-office",
                "name": name,
                "office_type": OFFICE_TYPE,
                "state_id": state_id,
                "country_code": COUNTRY_CODE,
            }
        )
    return sorted(records, key=lambda row: row["id"])


def build_observations(records_by_location: dict[str, dict[str, str]]) -> list[dict[str, Any]]:
    observations = []
    for position, (location, state_id, name) in enumerate(LOCATION_ROWS, start=1):
        record = records_by_location[location]
        observations.append(
            {
                "source_key": "nema-flood-operations-press-release",
                "source_position": position,
                "source_section": "Press release body",
                "published_office_category": "Zonal, Territorial and Operation offices",
                "published_location": location,
                "published_state_or_location": location,
                "evidence": "The NEMA Zonal, Territorial and Operation offices are located in Lagos, Ibadan, Ekiti, Abuja, Minna, Jos, Enugu, Owerri, Port Harcourt, Edo, Uyo, Kano, Sokoto, Kaduna, Maiduguri, Yola and Gombe.",
                "normalized": record,
                "decision": "retain",
                "target_id": record["id"],
                "reason": "Official NEMA press release directly lists this location as one of the Agency's Zonal, Territorial and Operation office locations.",
            }
        )
    return sorted(observations, key=lambda row: row["normalized"]["id"])


def build_schema(record_count: int) -> dict[str, Any]:
    return {
        "$schema": "https://json-schema.org/draft/2020-12/schema",
        "$id": "https://softdata-api.local/schemas/emergency/nema_zonal_territorial_operation_offices.schema.json",
        "title": TITLE,
        "type": "array",
        "minItems": record_count,
        "maxItems": record_count,
        "uniqueItems": True,
        "items": {"$ref": "#/$defs/nemaOffice"},
        "$defs": {
            "nemaOffice": {
                "type": "object",
                "additionalProperties": False,
                "required": ["id", "name", "office_type", "state_id", "country_code"],
                "properties": {
                    "id": {"type": "string", "pattern": "^[a-z0-9]+(?:-[a-z0-9]+)*$", "maxLength": 255},
                    "name": {"type": "string", "minLength": 1},
                    "office_type": {"type": "string", "enum": [OFFICE_TYPE]},
                    "state_id": {"type": "string", "enum": sorted(STATE_IDS)},
                    "country_code": {"type": "string", "const": COUNTRY_CODE},
                },
            }
        },
    }


def build_metadata(records: list[dict[str, str]], observations: list[dict[str, Any]]) -> dict[str, Any]:
    return {
        "dataset_key": DATASET_KEY,
        "title": TITLE,
        "group": "emergency",
        "country_code": COUNTRY_CODE,
        "record_count": len(records),
        "source_rows": len(observations),
        "source_row_count": len(observations),
        "snapshot_date": RETRIEVED_AT,
        "retrieval_date": RETRIEVED_AT,
        "relative_path": "emergency/nema_zonal_territorial_operation_offices.json",
        "schema_path": "schemas/emergency/nema_zonal_territorial_operation_offices.schema.json",
        "reconciliation_path": "metadata/emergency/nema_zonal_territorial_operation_offices_reconciliation/index.json",
        "status": "dataset_foundation_only",
        "official_terminology": "Zonal, Territorial and Operation offices",
        "field_contract": {
            "required": ["id", "name", "office_type", "state_id", "country_code"],
            "optional": [],
            "omitted": ["lga_id", "address", "telephone", "email", "coverage_area", "operational_status", "website", "logo_url"],
        },
        "office_type_counts": dict(Counter(row["office_type"] for row in records)),
        "state_counts": dict(Counter(row["state_id"] for row in records)),
        "decision_counts": dict(Counter(row["decision"] for row in observations)),
        "state_aliases": STATE_ALIASES,
        "sources": SOURCES,
        "limitations": [
            "This is a dated institutional-office snapshot, not a live operational-status guarantee.",
            "The primary source lists NEMA Zonal, Territorial and Operation office locations collectively; it does not classify each retained row into a narrower per-office subtype.",
            "The public dataset retains only office identity, collective office type, state_id and country_code; addresses, telephone numbers, emails, LGA IDs, coverage areas and operational status are omitted because they are not consistently published as privacy-safe institutional row fields.",
            "No NEMA office contact or emergency number was called, messaged or operationally tested.",
            "Repository, service, HTTP API, OpenAPI, Postman and startup verification integration are deferred to later phases.",
        ],
        "privacy": "Staff names and personal phone numbers found on the inspected Zonal & Territorial Offices page were not retained in generated artifacts. No personal contacts, incident data, caller data, credentials, cookies, raw HTML pages or temporary extraction files are committed.",
    }


def partition_observations(observations: list[dict[str, Any]]) -> dict[str, dict[str, Any]]:
    partitions: dict[str, dict[str, Any]] = {}
    for observation in observations:
        state_id = observation["normalized"]["state_id"]
        partitions[state_id] = {
            "dataset_key": DATASET_KEY,
            "partition_key": state_id,
            "records": [observation],
        }
    return dict(sorted(partitions.items()))


def build_outputs() -> dict[str, bytes]:
    records = build_records()
    records_by_location = {name.replace("NEMA ", "").replace(" Office", ""): row for row in records for name in [row["name"]]}
    observations = build_observations(records_by_location)
    validate(records, observations)

    outputs: dict[str, bytes] = {
        "datasets/emergency/nema_zonal_territorial_operation_offices.json": canonical_json(records).encode(),
        "datasets/schemas/emergency/nema_zonal_territorial_operation_offices.schema.json": canonical_json(build_schema(len(records))).encode(),
        "datasets/metadata/emergency/nema_zonal_territorial_operation_offices.json": canonical_json(build_metadata(records, observations)).encode(),
    }

    partitions = partition_observations(observations)
    partition_index = []
    for state_id, partition in partitions.items():
        path = f"datasets/metadata/emergency/nema_zonal_territorial_operation_offices_reconciliation/{state_id}.json"
        data = canonical_json(partition).encode()
        outputs[path] = data
        partition_index.append(
            {
                "path": path.removeprefix("datasets/"),
                "partition_key": state_id,
                "source_rows": len(partition["records"]),
                "size_bytes": len(data),
                "sha256": hashlib.sha256(data).hexdigest(),
            }
        )

    index = {
        "dataset_key": DATASET_KEY,
        "source_rows": len(observations),
        "final_record_count": len(records),
        "decision_counts": dict(Counter(row["decision"] for row in observations)),
        "partitions": partition_index,
    }
    outputs["datasets/metadata/emergency/nema_zonal_territorial_operation_offices_reconciliation/index.json"] = canonical_json(index).encode()
    return outputs


def validate(records: list[dict[str, str]], observations: list[dict[str, Any]]) -> None:
    ids = [row["id"] for row in records]
    if len(records) != 17:
        raise ValueError(f"expected 17 records, found {len(records)}")
    if ids != sorted(ids) or len(set(ids)) != len(ids):
        raise ValueError("record IDs must be unique and deterministically sorted")
    for row in records:
        if set(row) != {"id", "name", "office_type", "state_id", "country_code"}:
            raise ValueError(f"unsupported fields in {row['id']}")
        if len(row["id"]) > 255 or not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", row["id"]):
            raise ValueError(f"invalid ID {row['id']}")
        if not row["name"] or row["office_type"] != OFFICE_TYPE or row["state_id"] not in STATE_IDS or row["country_code"] != COUNTRY_CODE:
            raise ValueError(f"invalid record {row['id']}")
    seen_positions = set()
    for observation in observations:
        if observation["decision"] not in DECISIONS:
            raise ValueError(f"unknown decision {observation['decision']}")
        key = (observation["source_key"], observation["source_position"])
        if key in seen_positions:
            raise ValueError(f"duplicate source position {key}")
        seen_positions.add(key)
        if observation["decision"] == "retain" and observation["target_id"] not in ids:
            raise ValueError(f"unresolved retained target {observation['target_id']}")
    decisions = Counter(row["decision"] for row in observations)
    if len(observations) - decisions.get("merge_exact_duplicate", 0) - sum(v for k, v in decisions.items() if k.startswith("exclude_")) != len(records):
        raise ValueError("reconciliation arithmetic does not close")


def write_outputs(repo_root: Path, outputs: dict[str, bytes]) -> None:
    for relative, data in outputs.items():
        path = repo_root / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(data)


def check_outputs(repo_root: Path, outputs: dict[str, bytes]) -> None:
    mismatches = []
    for relative, expected in outputs.items():
        path = repo_root / relative
        if not path.exists() or path.read_bytes() != expected:
            mismatches.append(relative)
    if mismatches:
        raise SystemExit("generated artifacts differ: " + ", ".join(mismatches))


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo-root", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    outputs = build_outputs()
    if args.check:
        check_outputs(args.repo_root, outputs)
    else:
        write_outputs(args.repo_root, outputs)


if __name__ == "__main__":
    main()
