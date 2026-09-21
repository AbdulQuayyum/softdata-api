#!/usr/bin/env python3
"""Generate the FRSC zonal commands snapshot."""

from __future__ import annotations

import argparse
import hashlib
import json
import re
from collections import Counter
from pathlib import Path
from typing import Any

DATASET_KEY = "ng-frsc-zonal-commands"
TITLE = "Nigeria FRSC Zonal Commands Snapshot"
COUNTRY_CODE = "NG"
RETRIEVED_AT = "2026-09-21"
COMMAND_TYPE = "zonal_command"

SOURCE_KEY = "frsc-zonal-commands-api"
SOURCES: dict[str, dict[str, Any]] = {
    SOURCE_KEY: {
        "canonical_url": "https://www.frsc.gov.ng/commands/zonal-commands/",
        "retrieval_url": "https://staging.frsc.gov.ng/frsc-admin/api/zonal-commands?populate=image&sort=order:asc",
        "publisher": "Federal Road Safety Corps",
        "title": "Zonal Commands",
        "retrieval_date": RETRIEVED_AT,
        "retrievals": [
            {
                "http_status": 200,
                "content_type": "application/json; charset=utf-8",
                "byte_size": 26169,
                "sha256": "2474a790bf576fb8b63366b5a60bb4e63ca179728a05013b347f9fbd7adc7a1c",
                "last_modified": None,
                "etag": None,
            },
            {
                "http_status": 200,
                "content_type": "application/json; charset=utf-8",
                "byte_size": 26169,
                "sha256": "2474a790bf576fb8b63366b5a60bb4e63ca179728a05013b347f9fbd7adc7a1c",
                "last_modified": None,
                "etag": None,
            },
        ],
        "row_count": 12,
        "command_types": ["zonal_command"],
        "pagination": "Single Strapi collection response consumed by the official FRSC Zonal Commands page; no export endpoint found.",
        "reuse_terms": "No explicit reuse or attribution terms found on the inspected public page.",
        "extraction_method": "Manual safe extraction of order, code and state fields from the public API used by the official page; names, phones, emails, addresses and images were not committed.",
        "geographic_coverage": "Twelve FRSC zonal command headquarters states/FCT published by the official FRSC Zonal Commands directory.",
        "stability": "Two retrievals of the public JSON response were byte-identical.",
    }
}

COMMAND_ROWS = [
    (1, "RS1HQ", "Kaduna", "kaduna"),
    (2, "RS2HQ", "Lagos", "lagos"),
    (3, "RS3HQ", "Adamawa", "adamawa"),
    (4, "RS4HQ", "Plateau", "plateau"),
    (5, "RS5HQ", "Edo", "edo"),
    (6, "RS6HQ", "Rivers", "rivers"),
    (7, "RS7HQ", "FCT", "fct"),
    (8, "RS8HQ", "Kwara", "kwara"),
    (9, "RS9HQ", "Enugu", "enugu"),
    (10, "RS10HQ", "Sokoto", "sokoto"),
    (11, "RS11HQ", "Osun", "osun"),
    (12, "RS12HQ", "Bauchi", "bauchi"),
]

STATE_IDS = {state_id for _, _, _, state_id in COMMAND_ROWS}
DECISIONS = {
    "retain",
    "merge_exact_duplicate",
    "exclude_unresolved_identity",
    "exclude_unresolved_geography",
    "exclude_unclassified_command_type",
    "exclude_non_command_entry",
    "exclude_historical_entry",
    "exclude_privacy_unsafe",
}


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
    for _, code, state_label, state_id in COMMAND_ROWS:
        records.append(
            {
                "id": f"frsc-{slug(code)}-{state_id}-zonal-command",
                "name": f"FRSC {code} {state_label} Zonal Command",
                "command_type": COMMAND_TYPE,
                "command_code": code,
                "state_id": state_id,
                "country_code": COUNTRY_CODE,
            }
        )
    return sorted(records, key=lambda row: row["id"])


def build_observations(records_by_code: dict[str, dict[str, str]]) -> list[dict[str, Any]]:
    observations = []
    for position, (order, code, state_label, state_id) in enumerate(COMMAND_ROWS, start=1):
        record = records_by_code[code]
        observations.append(
            {
                "source_key": SOURCE_KEY,
                "source_position": position,
                "source_order": order,
                "source_name": code,
                "source_command_type": "Zonal Commands",
                "source_command_code": code,
                "source_state": state_label,
                "source_evidence_id": f"{SOURCE_KEY}:{order}",
                "normalized_name": record["name"],
                "normalized_command_type": COMMAND_TYPE,
                "state_id": state_id,
                "normalized": record,
                "decision": "retain",
                "target_id": record["id"],
                "reason": "Official FRSC Zonal Commands public API row publishes this command code and state in the Zonal Commands directory.",
            }
        )
    return sorted(observations, key=lambda row: row["normalized"]["id"])


def build_schema(record_count: int) -> dict[str, Any]:
    return {
        "$schema": "https://json-schema.org/draft/2020-12/schema",
        "$id": "https://softdata-api.local/schemas/emergency/frsc_zonal_commands.schema.json",
        "title": TITLE,
        "type": "array",
        "minItems": record_count,
        "maxItems": record_count,
        "uniqueItems": True,
        "items": {"$ref": "#/$defs/frscZonalCommand"},
        "$defs": {
            "frscZonalCommand": {
                "type": "object",
                "additionalProperties": False,
                "required": ["id", "name", "command_type", "command_code", "state_id", "country_code"],
                "properties": {
                    "id": {"type": "string", "pattern": "^[a-z0-9]+(?:-[a-z0-9]+)*$", "maxLength": 255},
                    "name": {"type": "string", "minLength": 1},
                    "command_type": {"type": "string", "enum": [COMMAND_TYPE]},
                    "command_code": {"type": "string", "pattern": "^RS(?:[1-9]|1[0-2])HQ$"},
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
        "relative_path": "emergency/frsc_zonal_commands.json",
        "schema_path": "schemas/emergency/frsc_zonal_commands.schema.json",
        "reconciliation_path": "metadata/emergency/frsc_zonal_commands_reconciliation/index.json",
        "status": "dataset_foundation_only",
        "official_terminology": "Zonal Commands",
        "field_contract": {
            "required": ["id", "name", "command_type", "command_code", "state_id", "country_code"],
            "optional": [],
            "omitted": [
                "lga_id",
                "address",
                "telephone",
                "email",
                "website",
                "logo_url",
                "coordinates",
                "commander",
                "staff",
                "operational_status",
                "parent_command_id",
            ],
        },
        "command_type_counts": dict(Counter(row["command_type"] for row in records)),
        "state_counts": dict(Counter(row["state_id"] for row in records)),
        "decision_counts": dict(Counter(row["decision"] for row in observations)),
        "sources": SOURCES,
        "limitations": [
            "This is a dated institutional-command snapshot, not a live operational-status guarantee.",
            "The dataset is limited to FRSC Zonal Commands published by the official FRSC Zonal Commands directory; sector commands, unit commands, outposts, driver-licence centres and other offices are excluded.",
            "State IDs identify the published zonal command headquarters state/FCT, not a full service-coverage area.",
            "The public dataset omits addresses, phone numbers, emails, images, commanders, staff names, coordinates, websites and operational status.",
            "No FRSC office, command contact or emergency line was called, messaged or operationally tested.",
            "Repository, service, HTTP API, OpenAPI, Postman and startup verification integration are deferred to later phases.",
        ],
        "privacy": "The source contains contact and personnel-adjacent fields, including commander, phone, email, address and image fields. These values are discarded and are not present in generated artifacts.",
        "excluded_command_types": ["sector_command", "unit_command", "outpost", "headquarters", "driver_licence_centre", "road_safety_office"],
        "hierarchy": "No parent-child hierarchy is published by the retained safe fields; parent_command_id is omitted.",
    }


def partition_observations(observations: list[dict[str, Any]]) -> dict[str, dict[str, Any]]:
    partitions: dict[str, dict[str, Any]] = {}
    for observation in observations:
        state_id = observation["state_id"]
        partitions[state_id] = {
            "dataset_key": DATASET_KEY,
            "partition_key": state_id,
            "records": [observation],
        }
    return dict(sorted(partitions.items()))


def build_outputs() -> dict[str, bytes]:
    records = build_records()
    records_by_code = {row["command_code"]: row for row in records}
    observations = build_observations(records_by_code)
    validate(records, observations)

    outputs: dict[str, bytes] = {
        "datasets/emergency/frsc_zonal_commands.json": canonical_json(records).encode(),
        "datasets/schemas/emergency/frsc_zonal_commands.schema.json": canonical_json(build_schema(len(records))).encode(),
        "datasets/metadata/emergency/frsc_zonal_commands.json": canonical_json(build_metadata(records, observations)).encode(),
    }

    partition_index = []
    for state_id, partition in partition_observations(observations).items():
        path = f"datasets/metadata/emergency/frsc_zonal_commands_reconciliation/{state_id}.json"
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
    outputs["datasets/metadata/emergency/frsc_zonal_commands_reconciliation/index.json"] = canonical_json(index).encode()
    return outputs


def validate(records: list[dict[str, str]], observations: list[dict[str, Any]]) -> None:
    ids = [row["id"] for row in records]
    if len(records) != 12:
        raise ValueError(f"expected 12 records, found {len(records)}")
    if ids != sorted(ids) or len(set(ids)) != len(ids):
        raise ValueError("record IDs must be unique and deterministically sorted")
    command_codes = [row["command_code"] for row in records]
    if len(set(command_codes)) != len(command_codes):
        raise ValueError("duplicate command code requires explicit review")
    for row in records:
        if set(row) != {"id", "name", "command_type", "command_code", "state_id", "country_code"}:
            raise ValueError(f"unsupported fields in {row['id']}")
        if len(row["id"]) > 255 or not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", row["id"]):
            raise ValueError(f"invalid ID {row['id']}")
        if not row["name"] or row["command_type"] != COMMAND_TYPE or row["state_id"] not in STATE_IDS or row["country_code"] != COUNTRY_CODE:
            raise ValueError(f"invalid record {row['id']}")
        if not re.fullmatch(r"RS(?:[1-9]|1[0-2])HQ", row["command_code"]):
            raise ValueError(f"invalid command code {row['command_code']}")
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
    excluded = sum(v for k, v in decisions.items() if k.startswith("exclude_"))
    if len(observations) - decisions.get("merge_exact_duplicate", 0) - excluded != len(records):
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
