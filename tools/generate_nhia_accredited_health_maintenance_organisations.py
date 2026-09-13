#!/usr/bin/env python3
"""Generate the dated NHIA accredited health maintenance organisation snapshot and reconciliation artifacts."""

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

DATASET_KEY = "ng-nhia-accredited-health-maintenance-organisations"
DATASET_FILE = "nhia_accredited_health_maintenance_organisations"
TITLE = "Nigeria NHIA Accredited Health Maintenance Organisations Snapshot"
COUNTRY_CODE = "NG"
SOURCE_URL = "https://www.nhia.gov.ng/hmo/"
SOURCE_API_URL = "https://www.nhia.gov.ng/wp-json/wp/v2/pages/910"
PRIVATE_HEALTH_PLAN_URL = "https://www.nhia.gov.ng/service/health-insurance/"
SSHIA_URL = "https://www.nhia.gov.ng/sshias/"
SOURCE_TITLE = "Health Maintenance Organisations (HMOs)"
SOURCE_TABLE_TITLE = "NHIA Accredited Health Maintenance Organizations"
SOURCE_PUBLISHER = "National Health Insurance Authority"
SOURCE_DATE_GMT = "2024-03-17T00:32:49Z"
SOURCE_MODIFIED_GMT = "2025-04-16T15:26:30Z"
SOURCE_CONTENT_TYPE = "text/html; charset=UTF-8"
SOURCE_API_CONTENT_TYPE = "application/json; charset=UTF-8"
SOURCE_HTML_RESPONSE_BYTES = 227880
SOURCE_JSON_RESPONSE_BYTES = 50093
PRIVATE_HEALTH_PLAN_RESPONSE_BYTES = 164295
SSHIA_RESPONSE_BYTES = 203703
SOURCE_HTML_SHA256_VALUES = (
    "cdb3f58bd62c0910f8a1987d43f147680ce045b2e5400e170c255bbe11f09b60",
    "c45f3c774b0d3b230c3700535af14afb822703551374faece549edda689995c2",
)
SOURCE_JSON_SHA256_VALUES = (
    "794e9719e10dfb3deb6e54c1947a4c96bd0709777216bfd6890f2308dd7c0fe5",
    "02b1bd99b8ef0583793b1aba10e0df641e7076e5a311b4f78cba07cb45ecc5ab",
)
PRIVATE_HEALTH_PLAN_SHA256 = "0b3a7ce60c46583ed9806028cb2144595c9d97ceebc9bfa01f695443d9a0ea39"
SSHIA_SHA256 = "7498f7478a124511d0ffe3dd6e3f6cec4904e4bec2b1c9926fcf16e4d4b5e195"
EXTRACTED_TABLE_SHA256 = "f3d33b441200aac2035f7c654d8486697b4ff21fb867cfbaab6c58a2cc6051a3"
EXPECTED_RAW_ROWS = 94
RETRIEVED_AT_DEFAULT = "2026-09-11"
ID_MAX_LENGTH = 255
ORGANISATION_TYPE = "health_maintenance_organisation"
ACCREDITATION_STATUS = "accredited"
SOURCE_STATUS = "listed_accredited_health_maintenance_organisation"


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


def canonical_json(value: Any) -> str:
    return json.dumps(value, ensure_ascii=False, indent=2, sort_keys=False) + "\n"


def compact_json(value: Any) -> str:
    return json.dumps(value, ensure_ascii=False, sort_keys=False, separators=(",", ":"))


def write_json(path: Path, value: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(canonical_json(value), encoding="utf-8")


def sha256_bytes(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def normalize_text(value: str) -> str:
    return " ".join(value.replace("\xa0", " ").split()).strip()


def slug(value: str) -> str:
    lowered = value.lower().replace("&", " and ")
    lowered = re.sub(r"[^a-z0-9]+", "-", lowered)
    return re.sub(r"-+", "-", lowered).strip("-")


def load_source_rows(source_path: Path) -> list[dict[str, Any]]:
    raw = source_path.read_text(encoding="utf-8", errors="replace")
    if source_path.suffix.lower() == ".json":
        payload = json.loads(raw)
        raw = payload["content"]["rendered"]
    parser = TableParser()
    parser.feed(raw)
    if not parser.rows:
        raise ValueError("NHIA HMO table was not found")
    header, data_rows = parser.rows[0], parser.rows[1:]
    expected_header = ["S/NO.", "HMO", "HMO ID", "WEBSITES", "Address", "Email", "Call Center Numbers"]
    if header != expected_header:
        raise ValueError(f"unexpected NHIA HMO table header: {header!r}")
    table_hash = sha256_bytes(compact_json(parser.rows).encode("utf-8"))
    if table_hash != EXTRACTED_TABLE_SHA256:
        raise ValueError(f"unexpected extracted table hash: {table_hash}")
    if len(data_rows) != EXPECTED_RAW_ROWS:
        raise ValueError(f"unexpected NHIA HMO row count: {len(data_rows)}")
    rows: list[dict[str, Any]] = []
    for index, row in enumerate(data_rows, start=1):
        if len(row) != 7:
            raise ValueError(f"unexpected column count at source row {index}: {row!r}")
        position_text, name, hmo_id, _website, address, _email, _phone = row
        if int(position_text) != index:
            raise ValueError(f"unexpected source position {position_text!r} at row {index}")
        rows.append(
            {
                "source_position": index,
                "source_name": name,
                "source_hmo_id": hmo_id,
                "source_category": SOURCE_TABLE_TITLE,
                "source_accreditation_status": SOURCE_STATUS,
                "raw_organisation_level_location": address,
                "normalized_name": slug(name),
                "normalized_hmo_id": hmo_id,
            }
        )
    return rows


def rows_from_reconciliation(repo: Path) -> list[dict[str, Any]]:
    recon_dir = repo / "datasets" / "metadata" / "healthcare" / f"{DATASET_FILE}_reconciliation"
    rows: list[dict[str, Any]] = []
    for path in sorted(recon_dir.glob("*.json")):
        if path.name == "index.json":
            continue
        payload = json.loads(path.read_text(encoding="utf-8"))
        rows.extend(payload["records"])
    rows.sort(key=lambda row: row["source_position"])
    return [
        {
            "source_position": row["source_position"],
            "source_name": row["source_name"],
            "source_hmo_id": row["source_hmo_id"],
            "source_category": row["source_category"],
            "source_accreditation_status": row["source_accreditation_status"],
            "raw_organisation_level_location": row.get("raw_organisation_level_location", ""),
            "normalized_name": row["normalized_name"],
            "normalized_hmo_id": row["normalized_hmo_id"],
        }
        for row in rows
    ]


def build_outputs(rows: list[dict[str, Any]], retrieved_at: str) -> tuple[list[dict[str, str]], list[dict[str, Any]]]:
    seen_identity: dict[tuple[str, str], str] = {}
    records: list[dict[str, str]] = []
    reconciliation: list[dict[str, Any]] = []
    for row in rows:
        name = normalize_text(row["source_name"])
        hmo_id = normalize_text(row["source_hmo_id"])
        if not name or not hmo_id or not hmo_id.isdigit():
            raise ValueError(f"row {row['source_position']} is missing required identity")
        record_id = slug(f"{name} {hmo_id}")
        if len(record_id) > ID_MAX_LENGTH:
            raise ValueError(f"public ID exceeds {ID_MAX_LENGTH}: {record_id}")
        identity = (name.casefold(), hmo_id)
        if identity in seen_identity:
            decision = "merge_exact_duplicate"
            final_id = None
            merge_target_id = seen_identity[identity]
        else:
            decision = "retain"
            final_id = record_id
            merge_target_id = None
            seen_identity[identity] = record_id
            records.append(
                {
                    "id": record_id,
                    "name": name,
                    "country_code": COUNTRY_CODE,
                    "organisation_type": ORGANISATION_TYPE,
                    "accreditation_status": ACCREDITATION_STATUS,
                    "hmo_id": hmo_id,
                }
            )
        reconciliation.append(
            {
                "source_position": row["source_position"],
                "source_page_section": SOURCE_TABLE_TITLE,
                "source_name": name,
                "source_hmo_id": hmo_id,
                "source_category": row["source_category"],
                "source_accreditation_status": row["source_accreditation_status"],
                "raw_organisation_level_location": normalize_text(row.get("raw_organisation_level_location", "")),
                "normalized_name": slug(name),
                "normalized_hmo_id": hmo_id,
                "decision": decision,
                "final_record_id": final_id,
                "merge_target_id": merge_target_id,
                "reason": "Retained as an organisation-level HMO row listed on the NHIA accredited HMO page." if decision == "retain" else "Merged only because source name and HMO ID matched exactly.",
                "evidence": {
                    "source_url": SOURCE_URL,
                    "source_api_url": SOURCE_API_URL,
                    "source_title": SOURCE_TITLE,
                    "source_table_title": SOURCE_TABLE_TITLE,
                    "publisher": SOURCE_PUBLISHER,
                    "retrieved_at": retrieved_at,
                    "source_modified_gmt": SOURCE_MODIFIED_GMT,
                    "extracted_table_sha256": EXTRACTED_TABLE_SHA256,
                },
            }
        )
    records.sort(key=lambda record: (record["name"].casefold(), int(record["hmo_id"]), record["id"]))
    ids = [record["id"] for record in records]
    if len(ids) != len(set(ids)):
        raise ValueError("duplicate public IDs")
    hmo_ids = [record["hmo_id"] for record in records]
    if len(hmo_ids) != len(set(hmo_ids)):
        raise ValueError("duplicate HMO IDs")
    return records, reconciliation


def schema() -> dict[str, Any]:
    return {
        "$schema": "https://json-schema.org/draft/2020-12/schema",
        "$id": f"https://softdata-api.local/schemas/healthcare/{DATASET_FILE}.schema.json",
        "title": TITLE,
        "description": "A dated organisation-level snapshot of NHIA-listed accredited Health Maintenance Organisations. It is not a complete live health-insurance licensing or registration register.",
        "type": "array",
        "minItems": EXPECTED_RAW_ROWS,
        "maxItems": EXPECTED_RAW_ROWS,
        "uniqueItems": True,
        "items": {"$ref": "#/$defs/nhiaAccreditedHealthMaintenanceOrganisation"},
        "$defs": {
            "nhiaAccreditedHealthMaintenanceOrganisation": {
                "type": "object",
                "additionalProperties": False,
                "required": ["id", "name", "country_code", "organisation_type", "accreditation_status", "hmo_id"],
                "properties": {
                    "id": {"type": "string", "pattern": "^[a-z0-9]+(?:-[a-z0-9]+)*$", "maxLength": ID_MAX_LENGTH},
                    "name": {"type": "string", "minLength": 1},
                    "country_code": {"const": COUNTRY_CODE},
                    "organisation_type": {"type": "string", "enum": [ORGANISATION_TYPE]},
                    "accreditation_status": {"type": "string", "enum": [ACCREDITATION_STATUS]},
                    "hmo_id": {"type": "string", "pattern": "^[0-9]+$"},
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
        "description": "A dated organisation-level snapshot of Health Maintenance Organisations listed by the National Health Insurance Authority as accredited. It is not a complete live health-insurance licensing or registration register.",
        "country_code": COUNTRY_CODE,
        "record_count": len(records),
        "source_rows": len(reconciliation),
        "decision_counts": dict(sorted(decision_counts.items())),
        "organisation_type_counts": dict(sorted(Counter(record["organisation_type"] for record in records).items())),
        "accreditation_status_counts": dict(sorted(Counter(record["accreditation_status"] for record in records).items())),
        "relative_path": f"healthcare/{DATASET_FILE}.json",
        "schema_path": f"schemas/healthcare/{DATASET_FILE}.schema.json",
        "source": {
            "publisher": SOURCE_PUBLISHER,
            "title": SOURCE_TITLE,
            "table_title": SOURCE_TABLE_TITLE,
            "url": SOURCE_URL,
            "api_url": SOURCE_API_URL,
            "retrieved_at": retrieved_at,
            "publication_date": SOURCE_DATE_GMT,
            "last_modified": SOURCE_MODIFIED_GMT,
            "content_type": SOURCE_CONTENT_TYPE,
            "api_content_type": SOURCE_API_CONTENT_TYPE,
            "retrievals": [
                {"retrieved_at": "2026-09-11T10:39:25+01:00", "url": SOURCE_URL, "response_bytes": SOURCE_HTML_RESPONSE_BYTES, "sha256": SOURCE_HTML_SHA256_VALUES[0]},
                {"retrieved_at": "2026-09-11T10:39:29+01:00", "url": SOURCE_URL, "response_bytes": SOURCE_HTML_RESPONSE_BYTES, "sha256": SOURCE_HTML_SHA256_VALUES[1]},
                {"retrieved_at": "2026-09-11T10:39:43+01:00", "url": SOURCE_API_URL, "response_bytes": SOURCE_JSON_RESPONSE_BYTES, "sha256": SOURCE_JSON_SHA256_VALUES[0]},
                {"retrieved_at": "2026-09-11T10:39:47+01:00", "url": SOURCE_API_URL, "response_bytes": SOURCE_JSON_RESPONSE_BYTES, "sha256": SOURCE_JSON_SHA256_VALUES[1]},
            ],
            "extracted_table_sha256": EXTRACTED_TABLE_SHA256,
            "last_modified_header": None,
            "etag": None,
            "format": "HTML table embedded in NHIA WordPress page",
            "worksheet_or_table": SOURCE_TABLE_TITLE,
            "extraction_method": "HTML table cell extraction from the NHIA HMO page or its WordPress page JSON; dynamic page wrappers are ignored and the extracted table hash is verified.",
            "raw_record_count": len(reconciliation),
            "available_organisation_fields": ["HMO name", "HMO ID", "website", "address", "email", "call center numbers"],
            "published_public_fields": ["id", "name", "country_code", "organisation_type", "accreditation_status", "hmo_id"],
            "omitted_fields": ["website", "address", "email", "call center numbers"],
            "personal_fields_present": "The HMO table includes email and telephone columns; those contact fields are omitted from public records and reconciliation artifacts.",
            "category_semantics": "Rows are listed under the NHIA page heading for accredited Health Maintenance Organizations.",
            "status_semantics": "The source page heading identifies all listed rows as accredited; no row-level expiry or revocation status is published in the table.",
            "terms": "No machine-readable copyright or reuse terms were found on the public source page; preserve NHIA attribution and review source terms before downstream redistribution.",
        },
        "supporting_sources": [
            {
                "publisher": SOURCE_PUBLISHER,
                "title": "Private Health Plan",
                "url": PRIVATE_HEALTH_PLAN_URL,
                "retrieved_at": retrieved_at,
                "format": "HTML page",
                "content_type": SOURCE_CONTENT_TYPE,
                "response_bytes": PRIVATE_HEALTH_PLAN_RESPONSE_BYTES,
                "sha256": PRIVATE_HEALTH_PLAN_SHA256,
                "last_modified_header": None,
                "etag": None,
                "use": "Confirms NHIA describes private health plans as provided by Health Maintenance Organisations and links users to accredited HMOs.",
            },
            {
                "publisher": SOURCE_PUBLISHER,
                "title": "State Social Health Insurance Agencies (SSHIAs)",
                "url": SSHIA_URL,
                "retrieved_at": retrieved_at,
                "format": "HTML table",
                "content_type": SOURCE_CONTENT_TYPE,
                "response_bytes": SSHIA_RESPONSE_BYTES,
                "sha256": SSHIA_SHA256,
                "last_modified_header": None,
                "etag": None,
                "row_count_including_header": 40,
                "raw_record_count": 39,
                "use": "Identified as a separate category from HMOs; not combined into this dataset because it lists state schemes and includes named directors and personal contact fields.",
            },
        ],
        "rejected_dataset_names": {
            "ng-health-insurance-organisations": "Rejected because the verified row-level source covers accredited HMOs only and does not cover all NHIA-recognised health-insurance organisation categories.",
            "ng-nhia-accredited-organisations": "Rejected because it is broader than the HMO-only source table.",
            "ng-accredited-health-maintenance-organisations": "Rejected because it omits the NHIA authority prefix that distinguishes the source and regulatory context.",
        },
        "reconciliation_path": f"metadata/healthcare/{DATASET_FILE}_reconciliation/index.json",
        "source_limitations": [
            "The source is an NHIA HMO page snapshot, not a complete live health-insurance licensing or registration register.",
            "The source table lists accredited Health Maintenance Organizations but does not publish row-level expiry dates, revocation status or historical accreditation periods.",
            "The source table includes contact fields; websites, addresses, emails and telephone numbers are omitted from the public contract in this foundation phase.",
            "No websites or logos are included in the public dataset.",
            "State Social Health Insurance Agencies are a separate NHIA-listed category and are not included in this HMO roster.",
            "The dataset does not establish that every Nigerian health-insurance organisation is licensed, registered or currently operational.",
            "No enrollee, patient, practitioner, staff, claims, policy, payment, banking or identity-document information is published.",
        ],
        "licensing": "NHIA source material retains its own rights. SoftData claims only its independent normalization, schema, identifiers, reconciliation and metadata.",
        "generation": {
            "input": "NHIA HMO HTML table projection or committed reconciliation replay",
            "extracted_table_sha256": EXTRACTED_TABLE_SHA256,
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
    partition_path = recon_dir / f"{ORGANISATION_TYPE}s.json"
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
    for record in records:
        if set(record) != {"id", "name", "country_code", "organisation_type", "accreditation_status", "hmo_id"}:
            raise ValueError(f"unexpected public fields: {record}")
        if not id_pattern.fullmatch(record["id"]) or len(record["id"]) > ID_MAX_LENGTH:
            raise ValueError(f"invalid public ID: {record['id']}")
        for field, value in record.items():
            if not value or normalize_text(value) != value:
                raise ValueError(f"empty or untrimmed field {field}: {record!r}")
        if record["country_code"] != COUNTRY_CODE or record["organisation_type"] != ORGANISATION_TYPE or record["accreditation_status"] != ACCREDITATION_STATUS:
            raise ValueError(f"invalid classification: {record!r}")
        if not record["hmo_id"].isdigit():
            raise ValueError(f"invalid HMO ID: {record['hmo_id']}")


def generate(repo: Path, rows: list[dict[str, Any]], retrieved_at: str) -> None:
    records, reconciliation = build_outputs(rows, retrieved_at)
    validate_records(records)
    write_json(repo / "datasets" / "healthcare" / f"{DATASET_FILE}.json", records)
    write_json(repo / "datasets" / "schemas" / "healthcare" / f"{DATASET_FILE}.schema.json", schema())
    write_json(repo / "datasets" / "metadata" / "healthcare" / f"{DATASET_FILE}.json", metadata(records, reconciliation, retrieved_at))
    write_reconciliation(repo, reconciliation)


def main() -> None:
    parser = argparse.ArgumentParser(description="Generate the NHIA accredited HMO snapshot and reconciliation artifacts.")
    parser.add_argument("--repo", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument("--source", type=Path, help="Captured NHIA HMO HTML page or WordPress page JSON.")
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
