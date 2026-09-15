#!/usr/bin/env python3
"""Generate a privacy-safe NHIA active-accredited healthcare providers snapshot."""

from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import json
import re
import unicodedata
import urllib.request
from collections import Counter, defaultdict
from dataclasses import dataclass
from pathlib import Path
from typing import Any

DATASET_KEY = "ng-nhia-active-accredited-healthcare-providers"
DATASET_FILE = "nhia_active_accredited_healthcare_providers"
TITLE = "Nigeria NHIA Active Accredited Healthcare Providers Snapshot"
COUNTRY_CODE = "NG"
LISTING_STATUS = "active_accredited"
SOURCE_URL = "https://www.nhia.gov.ng/hcps/"
SOURCE_API_URL = "https://www.nhia.gov.ng/wp-json/wp/v2/pages/1059"
TABLE_ID = 1722
TABLE_TITLE = "ACTIVEACCREDITED NHIA HEALTHCARE PROVIDER.csv"
EXPECTED_CHUNK_ROWS = [3000, 3000, 540, 0]
EXPECTED_RAW_ROWS = 6540
RETRIEVED_AT_DEFAULT = "2026-09-14"
ID_MAX_LENGTH = 255
PROHIBITED_PUBLIC_KEYS = {
    "address",
    "website",
    "website_url",
    "logo",
    "logo_url",
    "phone",
    "phone_number",
    "email",
    "email_address",
    "director",
    "contact",
    "contact_person",
    "state_id",
    "lga_id",
    "ownership",
    "coordinates",
}


@dataclass(frozen=True)
class Retrieval:
    url: str
    retrieved_at: str
    status: int
    content_type: str
    size_bytes: int
    sha256: str
    last_modified: str | None
    etag: str | None
    row_count: int | None = None
    chunk_number: int | None = None


def canonical_json(value: Any) -> str:
    return json.dumps(value, ensure_ascii=False, indent=2, sort_keys=False) + "\n"


def compact_json(value: Any) -> str:
    return json.dumps(value, ensure_ascii=False, sort_keys=False, separators=(",", ":"))


def sha256_bytes(value: bytes) -> str:
    return hashlib.sha256(value).hexdigest()


def write_json(path: Path, value: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(canonical_json(value), encoding="utf-8")


def fetch_bytes(url: str) -> tuple[bytes, Retrieval]:
    req = urllib.request.Request(url, headers={"User-Agent": "softdata-dataset-generator/1.0"})
    retrieved_at = dt.datetime.now(dt.timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")
    with urllib.request.urlopen(req, timeout=45) as response:  # nosec B310 - official public source.
        chunks: list[bytes] = []
        digest = hashlib.sha256()
        size = 0
        while True:
            chunk = response.read(1024 * 128)
            if not chunk:
                break
            digest.update(chunk)
            size += len(chunk)
            chunks.append(chunk)
        data = b"".join(chunks)
        retrieval = Retrieval(
            url=url,
            retrieved_at=retrieved_at,
            status=response.status,
            content_type=response.headers.get("Content-Type", ""),
            size_bytes=size,
            sha256=digest.hexdigest(),
            last_modified=response.headers.get("Last-Modified"),
            etag=response.headers.get("ETag"),
        )
    return data, retrieval


def normalize_text(value: str) -> str:
    value = unicodedata.normalize("NFKC", value or "")
    value = value.replace("\ufeff", "").replace("\xa0", " ")
    value = re.sub(r"\s+", " ", value)
    value = re.sub(r"\s+([,.;:)])", r"\1", value)
    value = re.sub(r"([(/])\s+", r"\1", value)
    return value.strip()


def normalize_code(value: str) -> tuple[str, list[str]]:
    original = value or ""
    text = unicodedata.normalize("NFKC", original).strip().upper()
    rules: list[str] = []
    if text != original:
        rules.append("trim_or_case_normalization")
    collapsed = re.sub(r"\s*/\s*", "/", text)
    collapsed = re.sub(r"\s+", " ", collapsed).strip()
    if collapsed != text:
        rules.append("collapse_separator_whitespace")
    pattern = re.compile(r"(?:[A-Z]{2,3})/\d{4}/P$")
    if pattern.fullmatch(collapsed):
        return collapsed, rules or ["none"]
    matches = re.findall(r"(?:[A-Z]{2,3})/\d{4}/P", collapsed)
    if len(matches) == 1:
        rules.append("extract_unambiguous_code")
        return matches[0], rules
    return collapsed, rules or ["none"]


def normalize_name(value: str, provider_code: str) -> str:
    name = normalize_text(value)
    if provider_code:
        escaped = re.escape(provider_code)
        name = re.sub(rf"\s*[-–—]?\s*{escaped}\s*$", "", name, flags=re.IGNORECASE).strip()
    return normalize_text(name)


def normalize_facility_type(value: str) -> str:
    text = normalize_text(value).lower()
    text = text.replace("&", " and ")
    text = re.sub(r"\s+", " ", text).strip()
    if text == "primary":
        return "primary"
    if text == "primary and secondary":
        return "primary_and_secondary"
    raise ValueError("unexpected facility type vocabulary")


def slug_from_provider_code(provider_code: str) -> str:
    slug = provider_code.lower().replace("/", "-")
    slug = re.sub(r"[^a-z0-9-]+", "-", slug)
    return re.sub(r"-+", "-", slug).strip("-")


def parse_page_config(html: bytes) -> dict[str, str]:
    text = html.decode("utf-8", errors="replace")
    table_markers = [f"table_id&quot;:{TABLE_ID}", f'"table_id":{TABLE_ID}', f'table_id="{TABLE_ID}"', f"table_id={TABLE_ID}"]
    if TABLE_TITLE not in text or not any(marker in text for marker in table_markers):
        raise ValueError("expected NHIA HCP table was not found")
    nonce_match = re.search(r'ninja_table_public_nonce["\\]*[:=]["\\]*([a-zA-Z0-9]+)', text)
    if not nonce_match:
        raise ValueError("Ninja Tables public nonce was not found")
    return {"nonce": nonce_match.group(1)}


def chunk_url(nonce: str, chunk_number: int) -> str:
    return (
        "https://www.nhia.gov.ng/wp-admin/admin-ajax.php"
        f"?action=wp_ajax_ninja_tables_public_action&table_id={TABLE_ID}"
        "&target_action=get-all-data&default_sorting=old_first&skip_rows=0&limit_rows=0"
        f"&ninja_table_public_nonce={nonce}&chunk_number={chunk_number}"
    )


def safe_rows_from_chunk(payload: bytes, chunk_number: int, offset: int) -> list[dict[str, Any]]:
    decoded = json.loads(payload)
    if not isinstance(decoded, list):
        raise ValueError("unexpected HCP chunk payload")
    safe_rows: list[dict[str, Any]] = []
    for row_index, row in enumerate(decoded, start=1):
        value = row.get("value") if isinstance(row, dict) else None
        if not isinstance(value, dict):
            raise ValueError("unexpected HCP row shape")
        # The address value is deliberately not read, copied or propagated.
        safe_rows.append(
            {
                "source_position": offset + row_index,
                "chunk_number": chunk_number,
                "chunk_row_position": row_index,
                "source_provider_code": normalize_text(str(value.get("healthcareprovidercode", ""))),
                "source_provider_name": normalize_text(str(value.get("healthcareprovidername", ""))),
                "source_facility_type": normalize_text(str(value.get("facilitytype", ""))),
            }
        )
    return safe_rows


def fetch_live_safe_rows() -> tuple[list[dict[str, Any]], dict[str, Any]]:
    page_bytes, page_retrieval = fetch_bytes(SOURCE_URL)
    wp_bytes, wp_retrieval = fetch_bytes(SOURCE_API_URL)
    wp_payload = json.loads(wp_bytes)
    modified_gmt = wp_payload.get("modified_gmt")
    nonce = parse_page_config(page_bytes)["nonce"]
    rows: list[dict[str, Any]] = []
    chunks: list[dict[str, Any]] = []
    offset = 0
    for chunk_number, expected_count in enumerate(EXPECTED_CHUNK_ROWS):
        payload, retrieval = fetch_bytes(chunk_url(nonce, chunk_number))
        safe = safe_rows_from_chunk(payload, chunk_number, offset)
        if len(safe) != expected_count:
            raise ValueError("unexpected HCP chunk row count")
        rows.extend(safe)
        offset += len(safe)
        chunks.append({**retrieval.__dict__, "row_count": len(safe), "chunk_number": chunk_number})
        payload = b""
    if len(rows) != EXPECTED_RAW_ROWS:
        raise ValueError("unexpected HCP source row count")
    return rows, {
        "page": page_retrieval.__dict__,
        "wordpress_json": {**wp_retrieval.__dict__, "modified_gmt": modified_gmt},
        "chunks": chunks,
    }


def rows_from_reconciliation(repo: Path) -> tuple[list[dict[str, Any]], dict[str, Any]]:
    metadata = json.loads((repo / "datasets/metadata/healthcare" / f"{DATASET_FILE}.json").read_text(encoding="utf-8"))
    index = json.loads((repo / "datasets/metadata/healthcare" / f"{DATASET_FILE}_reconciliation/index.json").read_text(encoding="utf-8"))
    rows: list[dict[str, Any]] = []
    for partition in index["partitions"]:
        data = json.loads((repo / "datasets" / partition["path"]).read_text(encoding="utf-8"))
        for record in data["records"]:
            rows.append(
                {
                    "source_position": record["source_position"],
                    "chunk_number": record["chunk_number"],
                    "chunk_row_position": record["chunk_row_position"],
                    "source_provider_code": record["source_provider_code"],
                    "source_provider_name": record["source_provider_name"],
                    "source_facility_type": record["source_facility_type"],
                }
            )
    rows.sort(key=lambda row: row["source_position"])
    return rows, metadata["source_retrieval"]


def build_outputs(rows: list[dict[str, Any]]) -> tuple[list[dict[str, Any]], list[dict[str, Any]], dict[str, Any]]:
    normalized: list[dict[str, Any]] = []
    code_rules: Counter[str] = Counter()
    code_rule_positions: dict[str, list[int]] = defaultdict(list)
    type_counts: Counter[str] = Counter()
    malformed_codes: list[int] = []
    for row in rows:
        code, rules = normalize_code(row["source_provider_code"])
        name = normalize_name(row["source_provider_name"], code)
        facility_type = normalize_facility_type(row["source_facility_type"])
        for rule in rules:
            code_rules[rule] += 1
            if rule != "none":
                code_rule_positions[rule].append(row["source_position"])
        if not re.fullmatch(r"(?:[A-Z]{2,3})/\d{4}/P", code):
            malformed_codes.append(row["source_position"])
        type_counts[facility_type] += 1
        normalized.append({**row, "provider_code": code, "normalized_name": name, "facility_type": facility_type})

    by_code: dict[str, list[dict[str, Any]]] = defaultdict(list)
    for row in normalized:
        by_code[row["provider_code"]].append(row)

    records: list[dict[str, Any]] = []
    reconciliation: list[dict[str, Any]] = []
    decisions: Counter[str] = Counter()
    for code in sorted(by_code):
        group = sorted(by_code[code], key=lambda row: row["source_position"])
        names = {row["normalized_name"] for row in group}
        types = {row["facility_type"] for row in group}
        if not code or not all(row["normalized_name"] and row["facility_type"] for row in group):
            decision = "exclude_missing_identity"
            target_id = None
        elif len(group) > 1 and (len(names) > 1 or len(types) > 1):
            decision = "exclude_unresolved_code_conflict" if len(names) > 1 else "exclude_unresolved_safe_field_conflict"
            target_id = None
        else:
            decision = "retain"
            target_id = slug_from_provider_code(code)
            record = {
                "id": target_id,
                "name": group[0]["normalized_name"],
                "country_code": COUNTRY_CODE,
                "provider_code": code,
                "facility_type": group[0]["facility_type"],
                "listing_status": LISTING_STATUS,
            }
            if len(record["id"]) > ID_MAX_LENGTH:
                raise ValueError("generated provider ID exceeds maximum length")
            records.append(record)

        for index, row in enumerate(group):
            row_decision = decision
            row_target = target_id
            reason = "retained as a unique safe provider-code observation"
            if decision == "retain" and index > 0:
                row_decision = "merge_exact_safe_duplicate"
                reason = "merged with exact safe duplicate"
            elif decision == "exclude_missing_identity":
                reason = "missing required safe identity evidence"
            elif decision == "exclude_unresolved_code_conflict":
                reason = "same normalized provider code has conflicting safe provider names"
            elif decision == "exclude_unresolved_safe_field_conflict":
                reason = "same normalized provider code has conflicting safe facility types"
            decisions[row_decision] += 1
            reconciliation.append(
                {
                    "source_position": row["source_position"],
                    "chunk_number": row["chunk_number"],
                    "chunk_row_position": row["chunk_row_position"],
                    "source_provider_code": row["source_provider_code"],
                    "source_provider_name": row["source_provider_name"],
                    "source_facility_type": row["source_facility_type"],
                    "provider_code": row["provider_code"],
                    "normalized_name": row["normalized_name"],
                    "facility_type": row["facility_type"],
                    "normalization_rules": [rule for rule in normalize_code(row["source_provider_code"])[1] if rule != "none"],
                    "decision": row_decision,
                    "target_id": row_target,
                    "reason": reason,
                }
            )

    records.sort(key=lambda record: record["id"])
    reconciliation.sort(key=lambda row: row["source_position"])
    if len({record["id"] for record in records}) != len(records):
        raise ValueError("duplicate public provider IDs")
    if len({record["provider_code"] for record in records}) != len(records):
        raise ValueError("duplicate retained provider codes")
    if len(rows) != EXPECTED_RAW_ROWS:
        raise ValueError("unexpected HCP row count")
    analysis = {
        "decision_counts": dict(sorted(decisions.items())),
        "facility_type_counts": dict(sorted(type_counts.items())),
        "code_normalization_rule_counts": dict(sorted(code_rules.items())),
        "code_normalization_rule_source_positions": {key: value for key, value in sorted(code_rule_positions.items())},
        "malformed_provider_code_source_positions": malformed_codes,
        "duplicate_provider_code_groups": {
            code: [row["source_position"] for row in group]
            for code, group in sorted(by_code.items())
            if len(group) > 1
        },
    }
    return records, reconciliation, analysis


def schema() -> dict[str, Any]:
    return {
        "$schema": "https://json-schema.org/draft/2020-12/schema",
        "$id": f"https://softdata.example/datasets/schemas/healthcare/{DATASET_FILE}.schema.json",
        "title": TITLE,
        "type": "array",
        "minItems": 6536,
        "maxItems": 6536,
        "uniqueItems": True,
        "items": {"$ref": "#/$defs/nhiaActiveAccreditedHealthcareProvider"},
        "$defs": {
            "nhiaActiveAccreditedHealthcareProvider": {
                "type": "object",
                "additionalProperties": False,
                "required": ["id", "name", "country_code", "provider_code", "facility_type", "listing_status"],
                "properties": {
                    "id": {"type": "string", "pattern": "^[a-z0-9]+(?:-[a-z0-9]+)*$", "maxLength": ID_MAX_LENGTH},
                    "name": {"type": "string", "minLength": 1},
                    "country_code": {"type": "string", "const": COUNTRY_CODE},
                    "provider_code": {"type": "string", "pattern": "^(?:[A-Z]{2,3})/[0-9]{4}/P$"},
                    "facility_type": {"type": "string", "enum": ["primary", "primary_and_secondary"]},
                    "listing_status": {"type": "string", "enum": [LISTING_STATUS]},
                },
            }
        },
    }


def generate(repo: Path, rows: list[dict[str, Any]], source_retrieval: dict[str, Any]) -> None:
    records, reconciliation, analysis = build_outputs(rows)
    safe_table_sha = sha256_bytes(compact_json(rows).encode("utf-8"))
    source_retrieval = json.loads(json.dumps(source_retrieval, ensure_ascii=False))
    for index, chunk in enumerate(source_retrieval.get("chunks", [])):
        chunk["chunk_number"] = index
    partition_rel = f"metadata/healthcare/{DATASET_FILE}_reconciliation/active_accredited_healthcare_providers.json"
    partition_value = {
        "dataset_key": DATASET_KEY,
        "records": reconciliation,
    }
    partition_bytes = canonical_json(partition_value).encode("utf-8")
    partition_sha = sha256_bytes(partition_bytes)
    decision_counts = Counter(row["decision"] for row in reconciliation)
    metadata = {
        "dataset_key": DATASET_KEY,
        "title": TITLE,
        "description": "Dated privacy-safe snapshot of healthcare providers listed on the NHIA Health Care Providers page and embedded active-accredited provider table.",
        "group": "healthcare",
        "country_code": COUNTRY_CODE,
        "version": "2026.09.14",
        "record_count": len(records),
        "schema_path": f"datasets/schemas/healthcare/{DATASET_FILE}.schema.json",
        "source_url": SOURCE_URL,
        "source_retrieval": source_retrieval,
        "source_rows": EXPECTED_RAW_ROWS,
        "normalized_safe_table_sha256": safe_table_sha,
        "decision_counts": dict(sorted(decision_counts.items())),
        "facility_type_counts": analysis["facility_type_counts"],
        "listing_status_counts": {LISTING_STATUS: len(records)},
        "source_quality": {
            "table_title": TABLE_TITLE,
            "duplicate_provider_code_groups": analysis["duplicate_provider_code_groups"],
            "malformed_provider_code_source_positions": analysis["malformed_provider_code_source_positions"],
            "code_normalization_rule_counts": analysis["code_normalization_rule_counts"],
            "code_normalization_rule_source_positions": analysis["code_normalization_rule_source_positions"],
            "state_or_lga_not_published": True,
        },
        "privacy_policy": {
            "safe_source_fields_retained": ["healthcare provider code", "healthcare provider name", "facility type", "source position"],
            "discarded_source_columns": ["ADDRESS"],
            "public_dataset_excludes": [
                "addresses",
                "websites",
                "logos",
                "phone numbers",
                "emails",
                "directors",
                "contacts",
                "practitioner data",
                "patient or enrollee data",
                "claims or policy data",
                "banking or payment data",
                "credentials or tokens",
            ],
        },
        "limitations": [
            "This is a dated NHIA active-accredited provider table snapshot, not a complete live licensing, registration or operational-status register.",
            "Provider inclusion reflects the source table title and page context; it should not be used as independent proof of current facility operation.",
            "States, LGAs, addresses, coordinates, ownership, websites, logos and contact details are not published in this dataset.",
            "Duplicate provider-code conflicts are excluded rather than merged without definitive safe identity evidence.",
        ],
        "generator": f"tools/generate_{DATASET_FILE}.py",
        "reconciliation_index_path": f"datasets/metadata/healthcare/{DATASET_FILE}_reconciliation/index.json",
    }
    index = {
        "dataset_key": DATASET_KEY,
        "source_rows": EXPECTED_RAW_ROWS,
        "final_record_count": len(records),
        "decision_counts": dict(sorted(decision_counts.items())),
        "partitions": [
            {
                "path": partition_rel,
                "source_rows": len(reconciliation),
                "size_bytes": len(partition_bytes),
                "sha256": partition_sha,
            }
        ],
    }
    for record in records:
        if set(record) != {"id", "name", "country_code", "provider_code", "facility_type", "listing_status"}:
            raise ValueError("public record contains unsupported keys")
        if set(record) & PROHIBITED_PUBLIC_KEYS:
            raise ValueError("public record contains prohibited keys")
    write_json(repo / "datasets/healthcare" / f"{DATASET_FILE}.json", records)
    write_json(repo / "datasets/schemas/healthcare" / f"{DATASET_FILE}.schema.json", schema())
    write_json(repo / "datasets/metadata/healthcare" / f"{DATASET_FILE}.json", metadata)
    write_json(repo / "datasets" / partition_rel, partition_value)
    write_json(repo / "datasets/metadata/healthcare" / f"{DATASET_FILE}_reconciliation/index.json", index)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo", type=Path, default=Path.cwd())
    parser.add_argument("--live", action="store_true")
    parser.add_argument("--from-reconciliation", action="store_true")
    args = parser.parse_args()
    if args.live == args.from_reconciliation:
        raise SystemExit("choose exactly one of --live or --from-reconciliation")
    if args.live:
        rows, source_retrieval = fetch_live_safe_rows()
    else:
        rows, source_retrieval = rows_from_reconciliation(args.repo)
    generate(args.repo, rows, source_retrieval)


if __name__ == "__main__":
    main()
