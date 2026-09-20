#!/usr/bin/env python3
"""Generate the official Nigerian emergency service contacts snapshot."""

from __future__ import annotations

import argparse
import hashlib
import json
import re
from collections import Counter
from pathlib import Path
from typing import Any

DATASET_KEY = "ng-emergency-service-contacts"
TITLE = "Nigeria Official Emergency Service Contacts Snapshot"
COUNTRY_CODE = "NG"
RETRIEVED_AT = "2026-09-20"
CONTACT_TYPES = {"short_code", "telephone"}
COVERAGE_TYPES = {"national", "state"}
SERVICE_TYPES = {"general_emergency", "disaster_management", "road_emergency", "police", "fire", "ambulance", "other"}
AVAILABILITY = {"24_hours"}
CALL_COST = {"toll_free"}

SOURCES: dict[str, dict[str, Any]] = {
    "ncc-special-duties-service-charter": {
        "canonical_url": "https://www.ncc.gov.ng/about/servicom-charter/departmental-charters/special-duties-department-service-charter",
        "publisher": "Nigerian Communications Commission",
        "title": "Nigerian Communications Commission Special Duties Department SERVICOM Charter",
        "retrieval_date": RETRIEVED_AT,
        "retrievals": [
            {
                "http_status": 200,
                "content_type": "text/html; charset=UTF-8",
                "byte_size": 112391,
                "sha256": "0ae11916eb79e90e94178f0c554107a9bddd0accec52835c96516b8df01c9764",
                "last_modified": None,
                "etag": None,
            },
            {
                "http_status": 200,
                "content_type": "text/html; charset=UTF-8",
                "byte_size": 112391,
                "sha256": "11fd016df9b291275eaa2faf4ccb066726c895e59e712d6296eea4da6e8c31e3",
                "last_modified": None,
                "etag": None,
            },
        ],
        "extraction_method": "Manual safe extraction from the Emergency Communication Centre and Accessibility sections; full dynamic HTML was not committed.",
        "geographic_coverage": "Across the country",
        "stability": "Byte size stable across two retrievals; full-page hash varied because the Drupal page is dynamic.",
    },
    "frsc-services": {
        "canonical_url": "https://www.frsc.gov.ng/services/",
        "publisher": "Federal Road Safety Corps",
        "title": "FRSC Services",
        "retrieval_date": RETRIEVED_AT,
        "retrievals": [
            {
                "http_status": 200,
                "content_type": "text/html",
                "byte_size": 54747,
                "sha256": "c12790988ad02a2563a0d337eb12d6f6dd8c0b9adc1480cbd5ab11ebf811d13d",
                "last_modified": "Fri, 17 Jul 2026 17:20:37 GMT",
                "etag": None,
            },
            {
                "http_status": 200,
                "content_type": "text/html",
                "byte_size": 54747,
                "sha256": "c12790988ad02a2563a0d337eb12d6f6dd8c0b9adc1480cbd5ab11ebf811d13d",
                "last_modified": "Fri, 17 Jul 2026 17:20:37 GMT",
                "etag": None,
            },
        ],
        "extraction_method": "Manual safe extraction from the Emergency Ambulance Service Scheme section; full HTML was not committed.",
        "geographic_coverage": "Across Nigeria",
        "stability": "Repeated retrieval was byte-identical.",
    },
    "nema-flood-operations-press-release": {
        "canonical_url": "https://nema.gov.ng/flood-incidents-nema-activates-operational-offices-nationwide-to-support-states-conduct-assessments/",
        "publisher": "National Emergency Management Agency",
        "title": "FLOOD INCIDENTS: NEMA ACTIVATES OPERATIONAL OFFICES NATIONWIDE TO SUPPORT STATES, CONDUCT ASSESSMENTS",
        "publication_date": "2024-07-07",
        "modification_date": "2024-07-07",
        "retrieval_date": RETRIEVED_AT,
        "retrievals": [
            {
                "http_status": 200,
                "content_type": "text/html; charset=UTF-8",
                "byte_size": 111374,
                "sha256": "652058f496404f6ddd9a448020dabc767de68381558edb39e3eda55adea5807f",
                "last_modified": None,
                "etag": None,
            },
            {
                "http_status": 200,
                "content_type": "text/html; charset=UTF-8",
                "byte_size": 111775,
                "sha256": "02e9a995142516a26d3dc7c4ccbc342a7044a5b5d6ce9be6df659d40dcd9f461",
                "last_modified": None,
                "etag": None,
            },
        ],
        "extraction_method": "Manual safe extraction from the press-release paragraph mentioning NEMA headquarters emergency contact line; full WordPress HTML was not committed.",
        "geographic_coverage": "Nationwide",
        "stability": "Safe extracted paragraph was stable; full-page hash and byte size varied because the WordPress page includes dynamic content.",
    },
    "federal-fire-service-contact-us": {
        "canonical_url": "https://fedfire.gov.ng/contact-us/",
        "publisher": "Federal Fire Service",
        "title": "Contact Us - Federal Fire Service Nigeria",
        "retrieval_date": RETRIEVED_AT,
        "retrievals": [
            {
                "http_status": 200,
                "content_type": "text/html; charset=UTF-8",
                "byte_size": 76719,
                "sha256": "afd4efca340503ac28c13785e6788482240e5320494ce348294c34aca335a9c1",
                "last_modified": None,
                "etag": None,
            },
            {
                "http_status": 200,
                "content_type": "text/html; charset=UTF-8",
                "byte_size": 76744,
                "sha256": "9c8143124493482086416b9fccc65ee7e004f6d7198b60701c9e32bdd109215f",
                "last_modified": None,
                "etag": None,
            },
        ],
        "extraction_method": "Manual safe extraction from Contact & Emergency and opening-hours text; full WordPress HTML was not committed.",
        "geographic_coverage": "Federal Fire Service national contact page",
        "stability": "Safe extracted contact text was stable; full-page hash and byte size varied because the WordPress page includes dynamic content.",
    },
}

OBSERVATIONS: list[dict[str, Any]] = [
    {
        "source_key": "ncc-special-duties-service-charter",
        "source_position": 1,
        "source_section": "Accessibility",
        "agency_label": "Nigerian Communications Commission",
        "service_label": "112 Toll free Emergency Number",
        "published_contact": "112",
        "published_coverage": "across the country",
        "evidence": "112 Toll free Emergency Number is available to receive Emergency Calls from the public 24hrs daily 7 days weekly (24/7) across the country.",
        "normalized": {
            "id": "nigerian-communications-commission-general-emergency-112-national",
            "service_name": "112 Emergency Number",
            "agency_name": "Nigerian Communications Commission",
            "service_type": "general_emergency",
            "contact_type": "short_code",
            "contact_value": "112",
            "coverage_type": "national",
            "country_code": COUNTRY_CODE,
            "availability": "24_hours",
            "call_cost": "toll_free",
            "notes": "NCC describes 112 as an emergency number for emergency calls across the country.",
        },
        "decision": "retain",
        "reason": "Official NCC charter directly identifies 112 as a toll-free emergency number across the country and states 24/7 availability.",
    },
    {
        "source_key": "frsc-services",
        "source_position": 1,
        "source_section": "Emergency Ambulance Service Scheme (EASS) Zebra",
        "agency_label": "Federal Road Safety Corps",
        "service_label": "Emergency Ambulance Service Scheme (EASS) Zebra",
        "published_contact": "122",
        "published_coverage": "across Nigeria",
        "evidence": "Members of the public can call the emergency toll-free number 122 to seek for help or report crashes promptly and emergencies.",
        "normalized": {
            "id": "federal-road-safety-corps-road-emergency-122-national",
            "service_name": "Emergency Ambulance Service Scheme (EASS) Zebra",
            "agency_name": "Federal Road Safety Corps",
            "service_type": "road_emergency",
            "contact_type": "short_code",
            "contact_value": "122",
            "coverage_type": "national",
            "country_code": COUNTRY_CODE,
            "call_cost": "toll_free",
            "notes": "FRSC describes 122 as an emergency toll-free number for help, crash reports and emergencies.",
        },
        "decision": "retain",
        "reason": "Official FRSC services page directly publishes 122 as an emergency toll-free number for road-crash/emergency help across Nigeria.",
    },
    {
        "source_key": "nema-flood-operations-press-release",
        "source_position": 1,
        "source_section": "Press release body",
        "agency_label": "National Emergency Management Agency",
        "service_label": "NEMA headquarters toll-free emergency contact line",
        "published_contact": "0800CALLNEMA (080022556362)",
        "published_coverage": "nationwide",
        "evidence": "At the headquarters NEMA also operates a toll-free emergency contact line: 0800CALLNEMA (080022556362) ... feedback from members of the public can be received nationwide.",
        "normalized": {
            "id": "national-emergency-management-agency-disaster-management-080022556362-national",
            "service_name": "NEMA Emergency Contact Line",
            "agency_name": "National Emergency Management Agency",
            "service_type": "disaster_management",
            "contact_type": "telephone",
            "contact_value": "080022556362",
            "coverage_type": "national",
            "country_code": COUNTRY_CODE,
            "call_cost": "toll_free",
            "notes": "Source publishes the vanity form 0800CALLNEMA with the numeric form 080022556362; the numeric form is used as the public contact value.",
        },
        "decision": "retain",
        "reason": "Official NEMA press release directly publishes the toll-free emergency contact line and states nationwide public feedback coverage.",
    },
    {
        "source_key": "federal-fire-service-contact-us",
        "source_position": 1,
        "source_section": "Contact & Emergency",
        "agency_label": "Federal Fire Service",
        "service_label": "Federal Fire Service distress calls",
        "published_contact": "+2348032003557",
        "published_coverage": "Federal Fire Service contact page",
        "evidence": "Federal Fire Service as first responder is available 24hrs 7days in a week to respond to distress calls from citizens. Phone: +2348032003557 | 112.",
        "normalized": {
            "id": "federal-fire-service-fire-2348032003557-national",
            "service_name": "Federal Fire Service Emergency Response",
            "agency_name": "Federal Fire Service",
            "service_type": "fire",
            "contact_type": "telephone",
            "contact_value": "+2348032003557",
            "coverage_type": "national",
            "country_code": COUNTRY_CODE,
            "availability": "24_hours",
            "notes": "Source also publishes 08032003557 elsewhere on the same page; +2348032003557 is retained as the Contact & Emergency phone form.",
        },
        "decision": "retain",
        "reason": "Official FFS contact page directly publishes this emergency/distress phone and states 24-hour availability.",
    },
    {
        "source_key": "federal-fire-service-contact-us",
        "source_position": 2,
        "source_section": "Contact & Emergency",
        "agency_label": "Federal Fire Service",
        "service_label": "Federal Fire Service distress calls",
        "published_contact": "112",
        "published_coverage": "Federal Fire Service contact page",
        "evidence": "Federal Fire Service as first responder is available 24hrs 7days in a week to respond to distress calls from citizens. Phone: +2348032003557 | 112.",
        "normalized": {
            "id": "federal-fire-service-fire-112-national",
            "service_name": "Federal Fire Service Emergency Response",
            "agency_name": "Federal Fire Service",
            "service_type": "fire",
            "contact_type": "short_code",
            "contact_value": "112",
            "coverage_type": "national",
            "country_code": COUNTRY_CODE,
            "availability": "24_hours",
            "notes": "The same short code is also documented by NCC as the national emergency number; this record preserves the distinct Federal Fire Service source observation.",
        },
        "decision": "retain",
        "reason": "Official FFS contact page directly publishes 112 as part of its distress-call contact text and states 24-hour availability.",
    },
]


def canonical_json(value: Any) -> str:
    return json.dumps(value, ensure_ascii=False, indent=2, sort_keys=False) + "\n"


def write_json(path: Path, value: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(canonical_json(value), encoding="utf-8")


def sha256_bytes(value: bytes) -> str:
    return hashlib.sha256(value).hexdigest()


def validate_record(record: dict[str, Any]) -> None:
    required = ["id", "service_name", "agency_name", "service_type", "contact_type", "contact_value", "coverage_type", "country_code"]
    for field in required:
        if not isinstance(record.get(field), str) or not record[field].strip():
            raise ValueError(f"missing required field {field}")
    if record["country_code"] != COUNTRY_CODE:
        raise ValueError("invalid country_code")
    if record["service_type"] not in SERVICE_TYPES:
        raise ValueError("invalid service_type")
    if record["contact_type"] not in CONTACT_TYPES:
        raise ValueError("invalid contact_type")
    if record["coverage_type"] not in COVERAGE_TYPES:
        raise ValueError("invalid coverage_type")
    if "availability" in record and record["availability"] not in AVAILABILITY:
        raise ValueError("invalid availability")
    if "call_cost" in record and record["call_cost"] not in CALL_COST:
        raise ValueError("invalid call_cost")
    if "state_id" in record:
        raise ValueError("state_id is unsupported for this national-only snapshot")
    contact = record["contact_value"]
    if record["contact_type"] == "short_code" and not re.fullmatch(r"[0-9]{3}", contact):
        raise ValueError("invalid short code")
    if record["contact_type"] == "telephone" and not re.fullmatch(r"(?:0[0-9]{10}|0800[0-9]{8}|\+234[0-9]{10})", contact):
        raise ValueError("invalid telephone")


def build_artifacts(repo: Path) -> dict[str, bytes]:
    records: list[dict[str, Any]] = []
    decisions: list[dict[str, Any]] = []
    seen_ids: set[str] = set()
    for observation in OBSERVATIONS:
        decision = {
            "source_key": observation["source_key"],
            "source_position": observation["source_position"],
            "source_section": observation["source_section"],
            "agency_label": observation["agency_label"],
            "service_label": observation["service_label"],
            "published_contact": observation["published_contact"],
            "published_coverage": observation["published_coverage"],
            "evidence": observation["evidence"],
            "normalized": observation["normalized"],
            "decision": observation["decision"],
            "reason": observation["reason"],
            "target_id": observation["normalized"]["id"] if observation["decision"] == "retain" else None,
        }
        decisions.append(decision)
        if observation["decision"] != "retain":
            continue
        record = {k: v for k, v in observation["normalized"].items() if v is not None}
        validate_record(record)
        if record["id"] in seen_ids:
            raise ValueError("duplicate ID")
        seen_ids.add(record["id"])
        records.append(record)
    records.sort(key=lambda item: item["id"])
    decisions.sort(key=lambda item: (item["source_key"], item["source_position"]))

    decision_counts = dict(Counter(item["decision"] for item in decisions))
    service_type_counts = dict(sorted(Counter(item["service_type"] for item in records).items()))
    coverage_type_counts = dict(sorted(Counter(item["coverage_type"] for item in records).items()))
    contact_type_counts = dict(sorted(Counter(item["contact_type"] for item in records).items()))

    if len(decisions) - decision_counts.get("merge_exact_duplicate", 0) - sum(v for k, v in decision_counts.items() if k.startswith("exclude_")) != len(records):
        raise ValueError("reconciliation arithmetic does not close")

    schema = {
        "$schema": "https://json-schema.org/draft/2020-12/schema",
        "$id": "https://softdata-api.local/schemas/emergency/emergency_service_contacts.schema.json",
        "title": TITLE,
        "description": "Dated snapshot of official Nigerian institutional emergency service contacts supported by primary public sources.",
        "type": "array",
        "minItems": len(records),
        "maxItems": len(records),
        "uniqueItems": True,
        "items": {"$ref": "#/$defs/emergencyServiceContact"},
        "$defs": {
            "emergencyServiceContact": {
                "type": "object",
                "additionalProperties": False,
                "required": ["id", "service_name", "agency_name", "service_type", "contact_type", "contact_value", "coverage_type", "country_code"],
                "properties": {
                    "id": {"type": "string", "pattern": "^[a-z0-9]+(?:-[a-z0-9]+)*$", "maxLength": 255},
                    "service_name": {"type": "string", "minLength": 1, "maxLength": 160},
                    "agency_name": {"type": "string", "minLength": 1, "maxLength": 160},
                    "service_type": {"type": "string", "enum": sorted(SERVICE_TYPES)},
                    "contact_type": {"type": "string", "enum": sorted(CONTACT_TYPES)},
                    "contact_value": {"type": "string", "pattern": "^(?:[0-9]{3}|0[0-9]{10}|0800[0-9]{8}|\\\\+234[0-9]{10})$"},
                    "coverage_type": {"type": "string", "enum": sorted(COVERAGE_TYPES)},
                    "country_code": {"type": "string", "const": COUNTRY_CODE},
                    "state_id": {"type": "string"},
                    "availability": {"type": "string", "enum": sorted(AVAILABILITY)},
                    "call_cost": {"type": "string", "enum": sorted(CALL_COST)},
                    "notes": {"type": "string", "minLength": 1, "maxLength": 400},
                },
            }
        },
    }
    metadata = {
        "dataset_key": DATASET_KEY,
        "title": TITLE,
        "group": "emergency",
        "country_code": COUNTRY_CODE,
        "status": "active",
        "description": "A dated, privacy-safe snapshot of official Nigerian institutional emergency service contacts directly supported by primary public sources.",
        "relative_path": "emergency/emergency_service_contacts.json",
        "schema_path": "schemas/emergency/emergency_service_contacts.schema.json",
        "reconciliation_path": "metadata/emergency/emergency_service_contacts_reconciliation/index.json",
        "record_count": len(records),
        "source_rows": len(decisions),
        "decision_counts": decision_counts,
        "service_type_counts": service_type_counts,
        "coverage_type_counts": coverage_type_counts,
        "contact_type_counts": contact_type_counts,
        "snapshot": {"retrieved_at": RETRIEVED_AT, "complete_nationwide_register": False},
        "sources": SOURCES,
        "public_fields": ["id", "service_name", "agency_name", "service_type", "contact_type", "contact_value", "coverage_type", "country_code", "availability", "call_cost", "notes"],
        "limitations": [
            "This is a dated source-verified snapshot, not a promise that every retained contact is continuously operational.",
            "The dataset is not a complete nationwide register of all federal, state, local, ambulance, police, fire or disaster-response emergency contacts.",
            "State-level contacts are not included because this foundation phase retained only contacts with explicit national or nationwide evidence.",
            "Nigeria Police Force festive press-release emergency lines were not retained because their official context appeared event-specific rather than a stable general emergency-contact register.",
        ],
        "privacy": "Only institutional public emergency contacts are retained. Personal names, personal phone numbers, personal emails, caller data, incident data, cookies and raw HTML pages are not committed.",
    }
    reconciliation = {
        "dataset_key": DATASET_KEY,
        "source_rows": len(decisions),
        "final_record_count": len(records),
        "decision_counts": decision_counts,
        "partitions": [],
    }
    partition = {
        "dataset_key": DATASET_KEY,
        "source_rows": len(decisions),
        "records": decisions,
    }

    outputs = {
        "datasets/emergency/emergency_service_contacts.json": canonical_json(records).encode(),
        "datasets/schemas/emergency/emergency_service_contacts.schema.json": canonical_json(schema).encode(),
        "datasets/metadata/emergency/emergency_service_contacts.json": canonical_json(metadata).encode(),
        "datasets/metadata/emergency/emergency_service_contacts_reconciliation/national.json": canonical_json(partition).encode(),
    }
    part_bytes = outputs["datasets/metadata/emergency/emergency_service_contacts_reconciliation/national.json"]
    reconciliation["partitions"].append(
        {
            "coverage_type": "national",
            "path": "metadata/emergency/emergency_service_contacts_reconciliation/national.json",
            "source_rows": len(decisions),
            "size_bytes": len(part_bytes),
            "sha256": sha256_bytes(part_bytes),
        }
    )
    outputs["datasets/metadata/emergency/emergency_service_contacts_reconciliation/index.json"] = canonical_json(reconciliation).encode()
    return outputs


def write_outputs(repo: Path) -> None:
    for rel, data in build_artifacts(repo).items():
        path = repo / rel
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(data)


def verify(repo: Path) -> None:
    for rel, data in build_artifacts(repo).items():
        existing = (repo / rel).read_bytes()
        if existing != data:
            raise SystemExit(f"{rel} differs from deterministic generation")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo-root", default=Path(__file__).resolve().parents[1], type=Path)
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    repo = args.repo_root.resolve()
    if args.check:
        verify(repo)
    else:
        write_outputs(repo)


if __name__ == "__main__":
    main()
