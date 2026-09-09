#!/usr/bin/env python3
"""Generate the Nigerian licensed/accredited medical laboratory dataset.

The authoritative public row-level source used in this pass is the MLSCN
Accreditation Service "Accredited Facilities" HTML table. Raw downloads are not
committed; after an initial source parse, deterministic regeneration can replay
from the committed reconciliation partitions.
"""

from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import html.parser
import json
import re
from collections import Counter, defaultdict
from pathlib import Path
from typing import Any
from urllib.parse import urljoin

DATASET_KEY = "ng-medical-laboratory-accreditations"
COUNTRY_CODE = "NG"
SOURCE_URL = "https://mlscn-as.org.ng/accredited.html"
SOURCE_TITLE = "MLSCN Accreditation Service Accredited Facilities"
SOURCE_PUBLISHER = "Medical Laboratory Science Council of Nigeria Accreditation Service"
SOURCE_SHA256 = "0e0ac1856ee89bea9371294b191baec601e25a79be4ae92815f3e5f0b584ba4a"
SOURCE_BYTES = 30802
SOURCE_LAST_MODIFIED = "2026-06-19T11:05:19Z"
SOURCE_ETAG = '"7852-6549944b0a1e1"'
RETRIEVED_AT_DEFAULT = "2026-09-09"
ID_MAX_LENGTH = 255

STATE_ALIASES = {
    "akwa ibom": "akwa-ibom",
    "anambra": "anambra",
    "cross river": "cross-river",
    "fct": "fct",
    "abuja": "fct",
    "gombe": "gombe",
    "imo": "imo",
    "kaduna": "kaduna",
    "lagos": "lagos",
    "osun": "osun",
    "oyo": "oyo",
    "rivers": "rivers",
}

MONTHS = {
    "january": 1,
    "february": 2,
    "march": 3,
    "april": 4,
    "may": 5,
    "june": 6,
    "july": 7,
    "august": 8,
    "september": 9,
    "october": 10,
    "november": 11,
    "december": 12,
}

PERSONAL_FIELD_HINTS = (
    "scientist",
    "practitioner",
    "phone",
    "email",
    "superintendent",
    "mls no",
    "registration number",
)


def canonical_json(value: Any) -> str:
    return json.dumps(value, ensure_ascii=False, indent=2, sort_keys=False) + "\n"


def write_json(path: Path, value: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(canonical_json(value), encoding="utf-8")


def sha256_bytes(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def slug(value: str) -> str:
    value = value.lower().replace("&", " and ")
    value = re.sub(r"[^a-z0-9]+", "-", value)
    return re.sub(r"-+", "-", value).strip("-")


def clean_text(value: str) -> str:
    value = value.replace("\xa0", " ")
    value = re.sub(r"\s+", " ", value)
    return value.strip()


def parse_date(value: str) -> str:
    cleaned = clean_text(value).rstrip(".")
    cleaned = re.sub(r"(\d+)(st|nd|rd|th)", r"\1", cleaned, flags=re.I)
    match = re.fullmatch(r"(\d{1,2})\s+([A-Za-z]+),?\s+(\d{4})", cleaned)
    if not match:
        raise ValueError(f"unsupported MLSCN date: {value!r}")
    day = int(match.group(1))
    month = MONTHS[match.group(2).lower()]
    year = int(match.group(3))
    return dt.date(year, month, day).isoformat()


def normalize_address(value: str) -> str:
    value = clean_text(value)
    value = value.replace(" ,", ",")
    value = value.replace("/ ", "/")
    value = value.replace(" /", "/")
    return value


def detect_state(address: str, name: str = "") -> tuple[str, str]:
    haystack = f"{address} {name}".lower()
    haystack = haystack.replace("-", " ")
    haystack = re.sub(r"[^a-z0-9]+", " ", haystack)
    for alias, state_id in sorted(STATE_ALIASES.items(), key=lambda item: -len(item[0])):
        if re.search(rf"\b{re.escape(alias)}\b", haystack):
            return state_id, alias.title() if state_id != "fct" else "FCT"
    raise ValueError(f"could not resolve Nigerian state from {address!r}")


def location_component(address: str, state_id: str) -> str:
    parts = [clean_text(part) for part in address.split(",") if clean_text(part)]
    if not parts:
        return state_id
    first = parts[0]
    if state_id == "fct" and re.search(r"\b(fct|abuja)\b", first, re.I):
        return "fct"
    if re.search(r"\bstate\b", first, re.I) or first.lower().replace("-", " ") in ("fct", "abuja fct", "abuja"):
        return state_id
    return slug(first) or state_id


class MLSCNAccreditedTableParser(html.parser.HTMLParser):
    def __init__(self) -> None:
        super().__init__()
        self.in_row = False
        self.in_cell = False
        self.current_cell: list[str] = []
        self.current_href: str | None = None
        self.current_row: list[dict[str, str | None]] = []
        self.rows: list[list[dict[str, str | None]]] = []

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        attr = dict(attrs)
        if tag == "tr":
            self.in_row = True
            self.current_row = []
        if self.in_row and tag in {"td", "th"}:
            self.in_cell = True
            self.current_cell = []
            self.current_href = None
        if self.in_cell and tag == "a":
            self.current_href = attr.get("href")

    def handle_data(self, data: str) -> None:
        if self.in_cell:
            self.current_cell.append(data)

    def handle_endtag(self, tag: str) -> None:
        if self.in_cell and tag in {"td", "th"}:
            self.current_row.append({"text": clean_text("".join(self.current_cell)), "href": self.current_href})
            self.in_cell = False
        if self.in_row and tag == "tr":
            if self.current_row:
                self.rows.append(self.current_row)
            self.in_row = False


def parse_source_html(path: Path, retrieved_at: str) -> list[dict[str, Any]]:
    html = path.read_text(encoding="utf-8", errors="replace")
    digest = sha256_bytes(html.encode("utf-8", errors="replace"))
    if digest != SOURCE_SHA256:
        raise ValueError(f"source hash mismatch: got {digest} want {SOURCE_SHA256}")
    parser = MLSCNAccreditedTableParser()
    parser.feed(html)
    source_rows = []
    for cells in parser.rows:
        if len(cells) < 6 or not cells[2]["text"].startswith("ML"):
            continue
        source_name = cells[0]["text"] or ""
        raw_address = normalize_address(cells[1]["text"] or "")
        accreditation_number = cells[2]["text"] or ""
        certificate_href = cells[3].get("href") or "#"
        certificate_url = urljoin(SOURCE_URL, certificate_href) if certificate_href != "#" else None
        if certificate_url == "https://mlscn-as.org.ng/":
            certificate_url = None
        approval_date = parse_date(cells[4]["text"] or "")
        expiry_date = parse_date(cells[5]["text"] or "")
        state_id, raw_state = detect_state(raw_address, source_name)
        source_position = len(source_rows) + 1
        source_rows.append(
            {
                "source_position": source_position,
                "source_name": source_name,
                "source_accreditation_number": accreditation_number,
                "source_accreditation_status": "listed_accredited_facility",
                "source_approval_date": approval_date,
                "source_expiry_date": expiry_date,
                "raw_state": raw_state,
                "raw_address": raw_address,
                "normalized_name": slug(source_name),
                "normalized_state_id": state_id,
                "certificate_url": certificate_url,
            }
        )
    if len(source_rows) != 30:
        raise ValueError(f"unexpected MLSCN source row count: {len(source_rows)}")
    return source_rows


def load_source_rows_from_reconciliation(repo: Path) -> list[dict[str, Any]]:
    index_path = repo / "datasets/metadata/healthcare/medical_laboratory_accreditations_reconciliation/index.json"
    if not index_path.exists():
        raise FileNotFoundError("reconciliation index is missing; pass --source-html for the first generation")
    index = json.loads(index_path.read_text(encoding="utf-8"))
    rows: list[dict[str, Any]] = []
    for partition in index["partitions"]:
        data = json.loads((repo / "datasets" / partition["path"]).read_text(encoding="utf-8"))
        for row in data["records"]:
            rows.append(
                {
                    "source_position": row["source_position"],
                    "source_name": row["source_name"],
                    "source_accreditation_number": row.get("source_accreditation_number"),
                    "source_accreditation_status": row.get("source_accreditation_status"),
                    "source_approval_date": row.get("source_approval_date"),
                    "source_expiry_date": row.get("source_expiry_date"),
                    "raw_state": row.get("raw_state"),
                    "raw_address": row.get("raw_address"),
                    "normalized_name": row["normalized_name"],
                    "normalized_state_id": row["normalized_state_id"],
                    "certificate_url": row.get("evidence", {}).get("certificate_url"),
                }
            )
    return sorted(rows, key=lambda row: row["source_position"])


def status_for(expiry_date: str, retrieved_at: str) -> str:
    if dt.date.fromisoformat(expiry_date) < dt.date.fromisoformat(retrieved_at):
        return "expired"
    return "accredited"


def build_public_id(row: dict[str, Any]) -> str:
    location = location_component(row["raw_address"], row["normalized_state_id"])
    value = f"{row['normalized_name']}-{row['normalized_state_id']}-{location}-{row['source_accreditation_number'].lower()}"
    value = re.sub(r"-+", "-", value).strip("-")
    if len(value) > ID_MAX_LENGTH:
        compact = f"{row['normalized_name']}-{row['normalized_state_id']}-{row['source_accreditation_number'].lower()}"
        value = re.sub(r"-+", "-", compact).strip("-")
    if len(value) > ID_MAX_LENGTH:
        raise ValueError(f"generated ID exceeds {ID_MAX_LENGTH}: {value}")
    return value


def exact_identity_key(row: dict[str, Any]) -> tuple[Any, ...]:
    return (
        row["source_name"],
        row.get("source_accreditation_number"),
        row.get("source_accreditation_status"),
        row.get("source_approval_date"),
        row.get("source_expiry_date"),
        row.get("raw_state"),
        row.get("raw_address"),
        row.get("normalized_name"),
        row.get("normalized_state_id"),
    )


def same_premises_key(row: dict[str, Any]) -> tuple[Any, ...]:
    return (
        row.get("source_accreditation_number"),
        row.get("normalized_name"),
        row.get("normalized_state_id"),
        normalize_address(row.get("raw_address") or "").lower(),
    )


def build_record(row: dict[str, Any], retrieved_at: str, final_id: str) -> dict[str, Any]:
    return {
        "id": final_id,
        "name": row["source_name"],
        "state_id": row["normalized_state_id"],
        "address": row["raw_address"],
        "country_code": COUNTRY_CODE,
        "accreditation_number": row["source_accreditation_number"],
        "accreditation_status": status_for(row["source_expiry_date"], retrieved_at),
        "approval_date": row["source_approval_date"],
        "expiry_date": row["source_expiry_date"],
    }


def build_reconciliation_row(row: dict[str, Any], decision: str, final_id: str | None, merge_target_id: str | None, retrieved_at: str, notes: str) -> dict[str, Any]:
    evidence = {
        "source_url": SOURCE_URL,
        "source_title": SOURCE_TITLE,
        "retrieved_at": retrieved_at,
        "source_sha256": SOURCE_SHA256,
    }
    if row.get("certificate_url"):
        evidence["certificate_url"] = row["certificate_url"]
    return {
        "source_position": row["source_position"],
        "source_name": row["source_name"],
        "source_accreditation_number": row.get("source_accreditation_number"),
        "source_accreditation_status": row.get("source_accreditation_status") or "listed_accredited_facility",
        "source_approval_date": row["source_approval_date"],
        "source_expiry_date": row["source_expiry_date"],
        "raw_state": row.get("raw_state"),
        "raw_address": row["raw_address"],
        "normalized_name": row["normalized_name"],
        "normalized_state_id": row["normalized_state_id"],
        "decision": decision,
        "final_record_id": final_id,
        "merge_target_id": merge_target_id,
        "evidence": evidence,
        "notes": notes,
    }


def build_outputs(source_rows: list[dict[str, Any]], retrieved_at: str) -> tuple[list[dict[str, Any]], list[dict[str, Any]]]:
    public_records: list[dict[str, Any]] = []
    reconciliation_rows: list[dict[str, Any]] = []
    seen_ids: set[str] = set()
    seen_accreditation_numbers: set[str] = set()
    exact_targets: dict[tuple[Any, ...], str] = {}
    premises_groups: dict[tuple[Any, ...], list[dict[str, Any]]] = defaultdict(list)
    for row in sorted(source_rows, key=lambda value: value["source_position"]):
        premises_groups[same_premises_key(row)].append(row)

    for group_rows in premises_groups.values():
        group_rows.sort(key=lambda row: (row["source_expiry_date"], row["source_approval_date"], -row["source_position"]), reverse=True)
        canonical = group_rows[0]
        final_id = build_public_id(canonical)
        if final_id in seen_ids:
            raise ValueError(f"duplicate generated public ID: {final_id}")
        seen_ids.add(final_id)
        acc_no = canonical["source_accreditation_number"]
        if acc_no in seen_accreditation_numbers:
            raise ValueError(f"duplicate conflicting accreditation number: {acc_no}")
        seen_accreditation_numbers.add(acc_no)
        public_records.append(build_record(canonical, retrieved_at, final_id))
        for row in sorted(group_rows, key=lambda value: value["source_position"]):
            exact_key = exact_identity_key(row)
            if exact_key in exact_targets:
                reconciliation_rows.append(
                    build_reconciliation_row(
                        row,
                        "merge_exact_duplicate",
                        None,
                        exact_targets[exact_key],
                        retrieved_at,
                        "Exact duplicate source observation merged because all meaningful institutional source fields match an earlier row.",
                    )
                )
                continue
            exact_targets[exact_key] = final_id
            if row is canonical:
                reconciliation_rows.append(
                    build_reconciliation_row(
                        row,
                        "retain",
                        final_id,
                        None,
                        retrieved_at,
                        "Retained from the MLSCN Accreditation Service public Accredited Facilities table as an institutional accreditation observation. Current validity is determined only from the listed certificate expiry date for this dated snapshot.",
                    )
                )
            else:
                reconciliation_rows.append(
                    build_reconciliation_row(
                        row,
                        "merge_same_premises",
                        None,
                        final_id,
                        retrieved_at,
                        "Merged as the same premises because the MLSCN accreditation number, normalized name, state and address match; the observation with the latest certificate expiry date is retained as canonical.",
                    )
                )
    public_records.sort(key=lambda row: (row["state_id"], row["name"].lower(), row["id"]))
    reconciliation_rows.sort(key=lambda row: (row["normalized_state_id"], row["source_position"]))
    return public_records, reconciliation_rows


def schema(record_count: int) -> dict[str, Any]:
    return {
        "$schema": "https://json-schema.org/draft/2020-12/schema",
        "$id": "https://softdata-api.local/schemas/healthcare/medical_laboratory_accreditations.schema.json",
        "title": "Nigeria Medical Laboratory Accreditation Register Snapshot",
        "description": "A dated snapshot of medical laboratory facility accreditation records published by the MLSCN Accreditation Service. It is not a complete register of all licensed medical laboratory premises in Nigeria.",
        "type": "array",
        "minItems": record_count,
        "maxItems": record_count,
        "uniqueItems": True,
        "items": {"$ref": "#/$defs/medicalLaboratoryAccreditation"},
        "$defs": {
            "medicalLaboratoryAccreditation": {
                "type": "object",
                "additionalProperties": False,
                "required": ["id", "name", "state_id", "country_code", "accreditation_status"],
                "properties": {
                    "id": {"type": "string", "pattern": "^[a-z0-9]+(?:-[a-z0-9]+)*$", "maxLength": ID_MAX_LENGTH},
                    "name": {"type": "string", "minLength": 1},
                    "state_id": {"type": "string", "pattern": "^[a-z0-9]+(?:-[a-z0-9]+)*$"},
                    "country_code": {"const": COUNTRY_CODE},
                    "accreditation_status": {"type": "string", "enum": ["accredited", "expired"]},
                    "accreditation_number": {"type": "string", "pattern": "^ML[0-9]{4}$"},
                    "approval_date": {"type": "string", "format": "date"},
                    "expiry_date": {"type": "string", "format": "date"},
                    "address": {"type": "string", "minLength": 1},
                },
            }
        },
    }

def metadata(records: list[dict[str, Any]], reconciliation_rows: list[dict[str, Any]], retrieved_at: str, partitions: list[dict[str, Any]]) -> dict[str, Any]:
    status_counts = Counter(record["accreditation_status"] for record in records)
    state_counts = Counter(record["state_id"] for record in records)
    return {
        "dataset_key": DATASET_KEY,
        "status": "active",
        "snapshot": True,
        "publicly_published": True,
        "title": "Nigeria Medical Laboratory Accreditation Register Snapshot",
        "description": "A dated snapshot of medical laboratory facility accreditation records published by the MLSCN Accreditation Service. It is not a complete register of all licensed medical laboratory premises in Nigeria.",
        "country_code": COUNTRY_CODE,
        "record_count": len(records),
        "source_rows": len(reconciliation_rows),
        "decision_counts": dict(sorted(Counter(row["decision"] for row in reconciliation_rows).items())),
        "accreditation_status_counts": dict(sorted(status_counts.items())),
        "state_coverage_count": len(state_counts),
        "relative_path": "healthcare/medical_laboratory_accreditations.json",
        "schema_path": "schemas/healthcare/medical_laboratory_accreditations.schema.json",
        "source": {
            "publisher": SOURCE_PUBLISHER,
            "title": SOURCE_TITLE,
            "url": SOURCE_URL,
            "retrieved_at": retrieved_at,
            "retrievals": [
                {"retrieved_at": "2026-09-09T08:19:52+01:00", "response_bytes": SOURCE_BYTES, "sha256": SOURCE_SHA256},
                {"retrieved_at": "2026-09-09T08:19:52+01:00", "response_bytes": SOURCE_BYTES, "sha256": SOURCE_SHA256},
            ],
            "response_bytes": SOURCE_BYTES,
            "sha256": SOURCE_SHA256,
            "format": "HTML table",
            "worksheet_or_table": "Accredited Facilities table",
            "pagination": "Single public HTML page; all 30 source rows were present in table order in both retrieves.",
            "raw_record_count": len(reconciliation_rows),
            "available_institutional_fields": [
                "facility name",
                "public location/address",
                "accreditation number",
                "certificate link when published",
                "certificate effective date",
                "certificate expiry date",
            ],
            "personal_fields_present": False,
            "status_semantics": "The table title identifies rows as accredited facilities; certificate expiry dates are used to derive dated accredited versus expired status for the 2026-09-09 snapshot.",
            "last_modified": SOURCE_LAST_MODIFIED,
            "etag": SOURCE_ETAG,
            "terms": "No machine-readable copyright or reuse terms were found on the public source page; preserve MLSCN attribution and review source terms before downstream redistribution.",
        },
        "supporting_sources": [
            {
                "publisher": "Medical Laboratory Science Council of Nigeria",
                "title": "MLSCN Hub laboratory API route",
                "url": "https://hub.mlscn.gov.ng/api/v1/laboratories",
                "retrieved_at": retrieved_at,
                "format": "JSON API endpoint",
                "access": "HTTP HEAD returned 401 Unauthorized, so it was documented but not used as row-level source.",
            }
        ],
        "reconciliation_path": "metadata/healthcare/medical_laboratory_accreditations_reconciliation/index.json",
        "source_limitations": [
            "The public source is the MLSCN Accreditation Service accredited-facilities table; it is narrower than a complete national medical laboratory premises register.",
            "The MLSCN Hub laboratory API endpoint returned 401 Unauthorized and was not used as a row-level source.",
            "The source does not publish LGA, ownership, laboratory-type taxonomy, annual renewal year or premises-registration numbers with reliable coverage, so those fields are omitted from public records.",
            "Expired certificates remain included as dated accreditation observations and are marked expired from the listed certificate expiry date; inclusion is not a claim of current accreditation or premises authorization.",
            "No individual laboratory scientists, personal registration numbers, personal phone numbers or personal email addresses are present in the committed artifacts.",
        ],
        "licensing": "MLSCN source material retains its own rights. SoftData claims only its independent normalization, schema, identifiers, reconciliation and metadata.",
        "generation": {
            "input": "MLSCN-AS HTML table projection or committed reconciliation replay",
            "source_sha256": SOURCE_SHA256,
            "partition_count": len(partitions),
        },
        "verified_at": retrieved_at,
    }


def validate(records: list[dict[str, Any]], reconciliation_rows: list[dict[str, Any]]) -> None:
    if len(records) != 30 or len(reconciliation_rows) != 30:
        raise ValueError("expected 30 public records and 30 reconciliation rows")
    ids = [record["id"] for record in records]
    if len(ids) != len(set(ids)):
        raise ValueError("duplicate public IDs")
    for record in records:
        if record["country_code"] != COUNTRY_CODE:
            raise ValueError("non-NG record")
        if len(record["id"]) > ID_MAX_LENGTH or not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", record["id"]):
            raise ValueError(f"invalid public ID: {record['id']}")
        for key, value in record.items():
            if isinstance(value, str) and value == "":
                raise ValueError(f"empty optional string in {record['id']}:{key}")
        lowered = canonical_json(record).lower()
        if any(hint in lowered for hint in PERSONAL_FIELD_HINTS):
            raise ValueError(f"personal-field hint leaked into public record {record['id']}")
    by_id = {record["id"] for record in records}
    positions = [row["source_position"] for row in reconciliation_rows]
    if sorted(positions) != list(range(1, 31)):
        raise ValueError("source positions are not complete")
    for row in reconciliation_rows:
        if row["decision"] == "retain" and row.get("final_record_id") not in by_id:
            raise ValueError("retained reconciliation row lacks public record")
        if row.get("merge_target_id") is not None:
            raise ValueError("unexpected merge target in no-merge MLSCN dataset")


def write_outputs(repo: Path, source_rows: list[dict[str, Any]], retrieved_at: str) -> None:
    records, reconciliation_rows = build_outputs(source_rows, retrieved_at)
    validate(records, reconciliation_rows)
    write_json(repo / "datasets/healthcare/medical_laboratory_accreditations.json", records)
    write_json(repo / "datasets/schemas/healthcare/medical_laboratory_accreditations.schema.json", schema(len(records)))

    recon_dir = repo / "datasets/metadata/healthcare/medical_laboratory_accreditations_reconciliation"
    by_state: dict[str, list[dict[str, Any]]] = defaultdict(list)
    for row in reconciliation_rows:
        by_state[row["normalized_state_id"]].append(row)
    partitions = []
    for state_id in sorted(by_state):
        value = {
            "dataset_key": DATASET_KEY,
            "state_id": state_id,
            "source_rows": len(by_state[state_id]),
            "records": by_state[state_id],
        }
        rel = f"metadata/healthcare/medical_laboratory_accreditations_reconciliation/{state_id}.json"
        path = repo / "datasets" / rel
        write_json(path, value)
        data = path.read_bytes()
        partitions.append({"state_id": state_id, "path": rel, "source_rows": len(by_state[state_id]), "size_bytes": len(data), "sha256": sha256_bytes(data)})
    index = {
        "dataset_key": DATASET_KEY,
        "source_rows": len(reconciliation_rows),
        "final_record_count": len(records),
        "decision_counts": dict(sorted(Counter(row["decision"] for row in reconciliation_rows).items())),
        "partitions": partitions,
    }
    write_json(recon_dir / "index.json", index)
    write_json(repo / "datasets/metadata/healthcare/medical_laboratory_accreditations.json", metadata(records, reconciliation_rows, retrieved_at, partitions))


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo", default=".")
    parser.add_argument("--source-html", help="Path to externally stored MLSCN-AS accredited.html snapshot")
    parser.add_argument("--from-reconciliation", action="store_true", help="Replay committed reconciliation rows")
    parser.add_argument("--retrieved-at", default=RETRIEVED_AT_DEFAULT)
    args = parser.parse_args()
    repo = Path(args.repo).resolve()
    if args.source_html:
        rows = parse_source_html(Path(args.source_html), args.retrieved_at)
    else:
        rows = load_source_rows_from_reconciliation(repo)
    write_outputs(repo, rows, args.retrieved_at)


if __name__ == "__main__":
    main()
