#!/usr/bin/env python3
"""Generate the dated NHIA-listed State Social Health Insurance Agencies snapshot."""

from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import html.parser
import json
import re
from collections import Counter
from pathlib import Path
from typing import Any

DATASET_KEY = "ng-nhia-state-social-health-insurance-agencies"
DATASET_FILE = "nhia_state_social_health_insurance_agencies"
TITLE = "Nigeria NHIA-Listed State Social Health Insurance Agencies Snapshot"
COUNTRY_CODE = "NG"
ORGANISATION_TYPE = "state_social_health_insurance_agency"
SOURCE_URL = "https://www.nhia.gov.ng/sshias/"
SOURCE_API_URL = "https://www.nhia.gov.ng/wp-json/wp/v2/pages/921"
SOURCE_PUBLISHER = "National Health Insurance Authority"
SOURCE_TABLE_TITLE = "STATE SOCIAL HEALTH INSURANCE AGENCIES (SSHIAs)"
SOURCE_CONTENT_TYPE = "text/html; charset=UTF-8"
SOURCE_API_CONTENT_TYPE = "application/json; charset=UTF-8"
SOURCE_MODIFIED_GMT = "2024-04-18T03:09:54"
EXPECTED_RAW_ROWS = 37
EXPECTED_SAFE_TABLE_SHA256 = "95f01387bd327d7a3280ed5f68ecfb5f00b16a017a93697525a73d3cf6303aa2"
RETRIEVED_AT_DEFAULT = "2026-09-13"
ID_MAX_LENGTH = 255

STATE_ALIASES = {
    "ABIA": "abia",
    "ADAMAWA": "adamawa",
    "AKS": "akwa-ibom",
    "AKWA IBOM": "akwa-ibom",
    "ANAMBRA": "anambra",
    "BAUCHI": "bauchi",
    "BAYELSA": "bayelsa",
    "BENUE": "benue",
    "BORNO": "borno",
    "CRS": "cross-river",
    "CROSS RIVER": "cross-river",
    "DELTA": "delta",
    "EBONYI": "ebonyi",
    "EDO": "edo",
    "EKITI": "ekiti",
    "ENUGU": "enugu",
    "FCT": "fct",
    "FEDERAL CAPITAL TERRITORY": "fct",
    "GOMBE": "gombe",
    "IMO": "imo",
    "JIGAWA": "jigawa",
    "KADUNA": "kaduna",
    "KANO": "kano",
    "KATSINA": "katsina",
    "KEBBI": "kebbi",
    "KOGI": "kogi",
    "KWARA": "kwara",
    "LAGOS": "lagos",
    "NASARAWA": "nasarawa",
    "NIGER": "niger",
    "OGUN": "ogun",
    "ONDO": "ondo",
    "OSUN": "osun",
    "OYO": "oyo",
    "PLATEAU": "plateau",
    "RIVER STATE": "rivers",
    "RIVERS": "rivers",
    "RIVERS STATE": "rivers",
    "SOKOTO": "sokoto",
    "TARABA": "taraba",
    "YOBE": "yobe",
    "ZAMFARA": "zamfara",
}

CANONICAL_STATES = {
    "abia",
    "adamawa",
    "akwa-ibom",
    "anambra",
    "bauchi",
    "bayelsa",
    "benue",
    "borno",
    "cross-river",
    "delta",
    "ebonyi",
    "edo",
    "ekiti",
    "enugu",
    "fct",
    "gombe",
    "imo",
    "jigawa",
    "kaduna",
    "kano",
    "katsina",
    "kebbi",
    "kogi",
    "kwara",
    "lagos",
    "nasarawa",
    "niger",
    "ogun",
    "ondo",
    "osun",
    "oyo",
    "plateau",
    "rivers",
    "sokoto",
    "taraba",
    "yobe",
    "zamfara",
}


class TableParser(html.parser.HTMLParser):
    def __init__(self) -> None:
        super().__init__(convert_charrefs=True)
        self.in_table = 0
        self.in_row = 0
        self.in_cell = 0
        self.cell_parts: list[str] = []
        self.row: list[str] = []
        self.rows: list[list[str]] = []

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        if tag == "table":
            self.in_table += 1
        elif self.in_table and tag == "tr":
            self.in_row += 1
            self.row = []
        elif self.in_table and self.in_row and tag in {"td", "th"}:
            self.in_cell += 1
            self.cell_parts = []

    def handle_endtag(self, tag: str) -> None:
        if self.in_table and self.in_row and self.in_cell and tag in {"td", "th"}:
            self.row.append(normalize_text("".join(self.cell_parts)))
            self.in_cell -= 1
        elif self.in_table and self.in_row and tag == "tr":
            if self.row:
                self.rows.append(self.row)
            self.in_row -= 1
        elif self.in_table and tag == "table":
            self.in_table -= 1

    def handle_data(self, data: str) -> None:
        if self.in_table and self.in_row and self.in_cell:
            self.cell_parts.append(data)


def normalize_text(value: str) -> str:
    value = value.replace("\ufeff", "").replace("\xa0", " ")
    value = re.sub(r"\s+", " ", value)
    value = re.sub(r"\s+\.", ".", value)
    value = re.sub(r"\.\(", " (", value)
    return value.strip()


def slug(value: str) -> str:
    lowered = value.lower().replace("&", " and ")
    lowered = re.sub(r"[^a-z0-9]+", "-", lowered)
    return re.sub(r"-+", "-", lowered).strip("-")


def canonical_json(value: Any) -> str:
    return json.dumps(value, ensure_ascii=False, indent=2, sort_keys=False) + "\n"


def compact_json(value: Any) -> str:
    return json.dumps(value, ensure_ascii=False, sort_keys=False, separators=(",", ":"))


def sha256_bytes(value: bytes) -> str:
    return hashlib.sha256(value).hexdigest()


def write_json(path: Path, value: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(canonical_json(value), encoding="utf-8")


def load_source_rows(source_path: Path) -> list[dict[str, Any]]:
    raw = source_path.read_text(encoding="utf-8", errors="replace")
    if source_path.suffix.lower() == ".json":
        payload = json.loads(raw)
        raw = payload["content"]["rendered"]
    parser = TableParser()
    parser.feed(raw)
    expected_header = ["S/N", "ORGANIZATION", "STATE", "DIRECTOR", "PHONE NO.", "E-MAIL ADDRESS", "WEBSITES", "ADDRESS"]
    for table_start in range(len(parser.rows)):
        header = parser.rows[table_start]
        if header == expected_header:
            data_rows = parser.rows[table_start + 1 :]
            break
    else:
        raise ValueError("NHIA SSHIA table was not found")
    rows: list[dict[str, Any]] = []
    for row in data_rows:
        if len(row) != len(expected_header):
            raise ValueError("unexpected SSHIA table column count")
        position, organisation, state = row[0], row[1], row[2]
        if not position and not organisation and not state:
            continue
        if not position.isdigit():
            raise ValueError("unexpected SSHIA source position")
        rows.append(
            {
                "source_position": int(position),
                "raw_organisation_name": organisation,
                "raw_state": state,
            }
        )
    safe_hash = sha256_bytes(compact_json(rows).encode("utf-8"))
    if safe_hash != EXPECTED_SAFE_TABLE_SHA256:
        raise ValueError(f"unexpected safe table hash: {safe_hash}")
    if len(rows) != EXPECTED_RAW_ROWS:
        raise ValueError(f"unexpected SSHIA row count: {len(rows)}")
    if [row["source_position"] for row in rows] != list(range(1, EXPECTED_RAW_ROWS + 1)):
        raise ValueError("unexpected SSHIA source position sequence")
    return rows


def rows_from_reconciliation(repo: Path) -> list[dict[str, Any]]:
    recon_path = repo / "datasets" / "metadata" / "healthcare" / f"{DATASET_FILE}_reconciliation" / "state_social_health_insurance_agencies.json"
    payload = json.loads(recon_path.read_text(encoding="utf-8"))
    rows = [
        {
            "source_position": row["source_position"],
            "raw_organisation_name": row["raw_organisation_name"],
            "raw_state": row["raw_state"],
        }
        for row in payload["records"]
    ]
    rows.sort(key=lambda row: row["source_position"])
    return rows


def normalize_state(value: str) -> str:
    key = normalize_text(value).upper()
    try:
        return STATE_ALIASES[key]
    except KeyError as exc:
        raise ValueError(f"unknown SSHIA state alias at safe state column: {key}") from exc


def build_outputs(rows: list[dict[str, Any]], retrieved_at: str) -> tuple[list[dict[str, str]], list[dict[str, Any]]]:
    records: list[dict[str, str]] = []
    reconciliation: list[dict[str, Any]] = []
    seen_states: dict[str, str] = {}
    seen_ids: set[str] = set()
    for row in rows:
        name = normalize_text(row["raw_organisation_name"])
        state_id = normalize_state(row["raw_state"])
        record_id = slug(name)
        decision = "retain"
        reason = "Retained as an organisation row listed in the NHIA SSHIA table."
        final_id = record_id
        if not name or not record_id:
            decision = "exclude_invalid_organisation_identity"
            reason = "Excluded because the safe organisation name was empty or could not produce a stable public ID."
            final_id = None
        elif state_id in seen_states:
            decision = "exclude_duplicate_state"
            reason = "Excluded because another retained source row already represents this state/FCT."
            final_id = None
        elif len(record_id) > ID_MAX_LENGTH:
            decision = "exclude_invalid_organisation_identity"
            reason = "Excluded because the generated public ID exceeds the supported length."
            final_id = None
        elif record_id in seen_ids:
            decision = "exclude_invalid_organisation_identity"
            reason = "Excluded because the generated public ID duplicates another retained row."
            final_id = None
        else:
            seen_states[state_id] = record_id
            seen_ids.add(record_id)
            records.append(
                {
                    "id": record_id,
                    "name": name,
                    "state_id": state_id,
                    "country_code": COUNTRY_CODE,
                    "organisation_type": ORGANISATION_TYPE,
                }
            )
        reconciliation.append(
            {
                "source_position": row["source_position"],
                "raw_organisation_name": name,
                "raw_state": normalize_text(row["raw_state"]),
                "normalized_name": name,
                "state_id": state_id,
                "final_record_id": final_id,
                "decision": decision,
                "reason": reason,
                "safe_evidence": {
                    "source_url": SOURCE_URL,
                    "source_api_url": SOURCE_API_URL,
                    "source_table_title": SOURCE_TABLE_TITLE,
                    "publisher": SOURCE_PUBLISHER,
                    "retrieved_at": retrieved_at,
                    "normalized_safe_table_sha256": EXPECTED_SAFE_TABLE_SHA256,
                },
            }
        )
    records.sort(key=lambda record: record["state_id"])
    if len(records) != EXPECTED_RAW_ROWS:
        raise ValueError(f"unexpected retained SSHIA count: {len(records)}")
    if {record["state_id"] for record in records} != CANONICAL_STATES:
        raise ValueError("SSHIA state/FCT coverage mismatch")
    return records, reconciliation


def schema() -> dict[str, Any]:
    return {
        "$schema": "https://json-schema.org/draft/2020-12/schema",
        "$id": f"https://softdata-api.local/schemas/healthcare/{DATASET_FILE}.schema.json",
        "title": TITLE,
        "description": "A dated organisation/state snapshot of NHIA-listed State Social Health Insurance Agencies. Inclusion means listed by NHIA, not necessarily accredited, licensed, registered or operational.",
        "type": "array",
        "minItems": EXPECTED_RAW_ROWS,
        "maxItems": EXPECTED_RAW_ROWS,
        "uniqueItems": True,
        "items": {"$ref": "#/$defs/nhiaStateSocialHealthInsuranceAgency"},
        "$defs": {
            "nhiaStateSocialHealthInsuranceAgency": {
                "type": "object",
                "additionalProperties": False,
                "required": ["id", "name", "state_id", "country_code", "organisation_type"],
                "properties": {
                    "id": {"type": "string", "pattern": "^[a-z0-9]+(?:-[a-z0-9]+)*$", "maxLength": ID_MAX_LENGTH},
                    "name": {"type": "string", "minLength": 1},
                    "state_id": {"type": "string", "enum": sorted(CANONICAL_STATES)},
                    "country_code": {"const": COUNTRY_CODE},
                    "organisation_type": {"type": "string", "enum": [ORGANISATION_TYPE]},
                },
            }
        },
    }


def metadata(records: list[dict[str, str]], reconciliation: list[dict[str, Any]], retrieved_at: str) -> dict[str, Any]:
    decision_counts = Counter(row["decision"] for row in reconciliation)
    return {
        "dataset_key": DATASET_KEY,
        "status": "active",
        "snapshot": True,
        "publicly_published": True,
        "title": TITLE,
        "description": "A dated organisation/state snapshot of State Social Health Insurance Agencies listed by the National Health Insurance Authority. Inclusion means listed by NHIA, not necessarily currently accredited, licensed, registered or operational.",
        "group": "healthcare",
        "country_code": COUNTRY_CODE,
        "version": "2026-09-13",
        "record_count": len(records),
        "source_rows": len(reconciliation),
        "decision_counts": dict(sorted(decision_counts.items())),
        "organisation_type_counts": dict(sorted(Counter(record["organisation_type"] for record in records).items())),
        "state_count": len({record["state_id"] for record in records}),
        "state_fct_coverage": sorted(record["state_id"] for record in records),
        "relative_path": f"healthcare/{DATASET_FILE}.json",
        "schema_path": f"schemas/healthcare/{DATASET_FILE}.schema.json",
        "source": {
            "publisher": SOURCE_PUBLISHER,
            "title": SOURCE_TABLE_TITLE,
            "url": SOURCE_URL,
            "api_url": SOURCE_API_URL,
            "retrieved_at": retrieved_at,
            "modified_gmt": SOURCE_MODIFIED_GMT,
            "content_type": SOURCE_CONTENT_TYPE,
            "api_content_type": SOURCE_API_CONTENT_TYPE,
            "retrievals": [
                {"url": SOURCE_URL, "retrieved_at": "2026-09-13T16:00:20Z", "http_status": 200, "content_type": SOURCE_CONTENT_TYPE, "response_bytes": 203707, "sha256": "e70fa85090020ba392156d3e4d3985733dd86095dc63b12a1441da9f3a201a70", "last_modified": None, "etag": None, "extraction_method": "HTML table parsed and immediately projected to source row, organisation and state columns.", "extracted_table_row_count": 37},
                {"url": SOURCE_URL, "retrieved_at": "2026-09-13T16:00:19Z", "http_status": 200, "content_type": SOURCE_CONTENT_TYPE, "response_bytes": 203709, "sha256": "37cb5cdd2e05b54c22749bd903b6e959104592134a1b9991d29e69274953c867", "last_modified": None, "etag": None, "extraction_method": "HTML table parsed and immediately projected to source row, organisation and state columns.", "extracted_table_row_count": 37},
                {"url": SOURCE_API_URL, "retrieved_at": "2026-09-13T16:00:36Z", "http_status": 200, "content_type": SOURCE_API_CONTENT_TYPE, "response_bytes": 24895, "sha256": "3a44fd946cc038e1c7145b0e06c5330fa6990b54d749b9a93c4e70683cd5589b", "last_modified": None, "etag": None, "wordpress_modified_gmt": SOURCE_MODIFIED_GMT, "extraction_method": "WordPress rendered content parsed and immediately projected to source row, organisation and state columns.", "extracted_table_row_count": 37},
            ],
            "raw_html_hash_note": "The two full HTML captures had different byte sizes and SHA-256 hashes; reproducibility is asserted only for the normalized safe table and generated artifacts.",
            "normalized_safe_table_sha256": EXPECTED_SAFE_TABLE_SHA256,
            "raw_record_count": len(reconciliation),
            "published_source_columns_used": ["S/N", "ORGANIZATION", "STATE"],
            "discarded_source_columns": ["DIRECTOR", "PHONE NO.", "E-MAIL ADDRESS", "WEBSITES", "ADDRESS"],
            "privacy_policy": "Director, phone, email, website and address columns were deliberately discarded immediately after table identification. Public and reconciliation artifacts contain organisation/state information only.",
            "category_semantics": "Rows appear under NHIA's State Social Health Insurance Agencies table.",
            "status_semantics": "The source lists SSHIAs but does not publish row-level accreditation, licensing, registration or operational status evidence.",
            "source_quality_notes": [
                "State values mix full names and abbreviations including AKS, CRS and FCT.",
                "Some organisation names are acronym-only in the source and are retained without expansion.",
                "The source contains inconsistent capitalization and punctuation; only mechanical whitespace and punctuation-spacing normalization is applied.",
                "Rivers is represented in the safe state column as RIVERS while the discarded address column says River State.",
                "Cross River and Ebonyi website-link mismatches were observed, so websites are deferred and not published.",
                "Several website cells are empty.",
            ],
        },
        "reconciliation_path": f"metadata/healthcare/{DATASET_FILE}_reconciliation/index.json",
        "update_frequency": "Snapshot; refresh only after a new source verification pass.",
        "source_limitations": [
            "This is a dated NHIA-listed SSHIA snapshot, not a complete live licensing, accreditation, registration or operational-status register.",
            "Inclusion means listed by NHIA under State Social Health Insurance Agencies, not necessarily currently accredited, licensed, registered or operational.",
            "The source publishes directors and contact fields; those fields were intentionally discarded and are absent from public and reconciliation artifacts.",
            "The dataset contains organisation and state information only.",
            "Websites and addresses are intentionally deferred pending separate verification.",
            "Health Maintenance Organisations are a separate NHIA category and are not included in this dataset.",
        ],
        "licensing": "NHIA source material retains its own rights. SoftData claims only its independent normalization, schema, identifiers, reconciliation and metadata.",
        "generation": {
            "generator": f"tools/generate_{DATASET_FILE}.py",
            "input": "NHIA SSHIA HTML table projection or committed reconciliation replay",
            "normalized_safe_table_sha256": EXPECTED_SAFE_TABLE_SHA256,
            "state_aliases": dict(sorted(STATE_ALIASES.items())),
            "partition_count": 1,
        },
        "verified_at": retrieved_at,
    }


def write_reconciliation(repo: Path, reconciliation: list[dict[str, Any]]) -> None:
    recon_dir = repo / "datasets" / "metadata" / "healthcare" / f"{DATASET_FILE}_reconciliation"
    recon_dir.mkdir(parents=True, exist_ok=True)
    for old in recon_dir.glob("*.json"):
        old.unlink()
    partition = {
        "dataset_key": DATASET_KEY,
        "organisation_type": ORGANISATION_TYPE,
        "source_rows": len(reconciliation),
        "records": sorted(reconciliation, key=lambda row: row["source_position"]),
    }
    partition_path = recon_dir / "state_social_health_insurance_agencies.json"
    write_json(partition_path, partition)
    data = partition_path.read_bytes()
    index = {
        "dataset_key": DATASET_KEY,
        "source_rows": len(reconciliation),
        "final_record_count": sum(1 for row in reconciliation if row["decision"] == "retain"),
        "decision_counts": dict(sorted(Counter(row["decision"] for row in reconciliation).items())),
        "partitions": [
            {
                "path": f"metadata/healthcare/{DATASET_FILE}_reconciliation/{partition_path.name}",
                "organisation_type": ORGANISATION_TYPE,
                "source_rows": len(reconciliation),
                "size_bytes": len(data),
                "sha256": sha256_bytes(data),
            }
        ],
    }
    write_json(recon_dir / "index.json", index)


def validate_records(records: list[dict[str, str]]) -> None:
    id_pattern = re.compile(r"^[a-z0-9]+(?:-[a-z0-9]+)*$")
    if len(records) != EXPECTED_RAW_ROWS:
        raise ValueError("unexpected public record count")
    ids = [record["id"] for record in records]
    states = [record["state_id"] for record in records]
    if len(ids) != len(set(ids)) or len(states) != len(set(states)):
        raise ValueError("duplicate SSHIA public IDs or states")
    if set(states) != CANONICAL_STATES:
        raise ValueError("SSHIA state coverage mismatch")
    for record in records:
        if set(record) != {"id", "name", "state_id", "country_code", "organisation_type"}:
            raise ValueError("unexpected SSHIA public fields")
        if not id_pattern.fullmatch(record["id"]) or len(record["id"]) > ID_MAX_LENGTH:
            raise ValueError("invalid SSHIA public ID")
        for value in record.values():
            if not value or normalize_text(value) != value:
                raise ValueError("empty or untrimmed SSHIA public field")
        if record["country_code"] != COUNTRY_CODE or record["organisation_type"] != ORGANISATION_TYPE:
            raise ValueError("invalid SSHIA public classification")


def generate(repo: Path, rows: list[dict[str, Any]], retrieved_at: str) -> None:
    records, reconciliation = build_outputs(rows, retrieved_at)
    validate_records(records)
    write_json(repo / "datasets" / "healthcare" / f"{DATASET_FILE}.json", records)
    write_json(repo / "datasets" / "schemas" / "healthcare" / f"{DATASET_FILE}.schema.json", schema())
    write_json(repo / "datasets" / "metadata" / "healthcare" / f"{DATASET_FILE}.json", metadata(records, reconciliation, retrieved_at))
    write_reconciliation(repo, reconciliation)


def main() -> None:
    parser = argparse.ArgumentParser(description="Generate the NHIA-listed SSHIA snapshot and reconciliation artifacts.")
    parser.add_argument("--repo", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument("--source", type=Path, help="Captured NHIA SSHIA HTML page or WordPress page JSON.")
    parser.add_argument("--from-reconciliation", action="store_true", help="Replay from committed reconciliation artifacts instead of a captured source page.")
    parser.add_argument("--retrieved-at", default=RETRIEVED_AT_DEFAULT)
    args = parser.parse_args()
    dt.date.fromisoformat(args.retrieved_at)
    if args.from_reconciliation == bool(args.source):
        parser.error("provide exactly one of --source or --from-reconciliation")
    rows = rows_from_reconciliation(args.repo) if args.from_reconciliation else load_source_rows(args.source)
    generate(args.repo, rows, args.retrieved_at)


if __name__ == "__main__":
    main()
