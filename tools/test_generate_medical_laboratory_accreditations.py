import importlib.util
import json
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path

SCRIPT = Path(__file__).with_name("generate_medical_laboratory_accreditations.py").resolve()
REPO = Path(__file__).resolve().parents[1]

spec = importlib.util.spec_from_file_location("generate_medical_laboratory_accreditations", SCRIPT)
generator = importlib.util.module_from_spec(spec)
spec.loader.exec_module(generator)


def fixture_row(position, name, address, acc_no, approval="2026-01-01", expiry="2030-01-01"):
    state_id, raw_state = generator.detect_state(address, name)
    return {
        "source_position": position,
        "source_name": name,
        "source_accreditation_number": acc_no,
        "source_accreditation_status": "listed_accredited_facility",
        "source_approval_date": approval,
        "source_expiry_date": expiry,
        "raw_state": raw_state,
        "raw_address": generator.normalize_address(address),
        "normalized_name": generator.slug(name),
        "normalized_state_id": state_id,
        "certificate_url": None,
    }


class MedicalLaboratoryAccreditationGeneratorTest(unittest.TestCase):
    def test_same_name_branches_receive_distinct_deterministic_ids(self):
        rows = [
            fixture_row(1, "Everight Diagnostic and Laboratory Services Limited", "No 2 Fez Street, Abuja, Nigeria", "ML0014"),
            fixture_row(2, "Everight Diagnostic and Laboratory Services Limited", "20 Asumpta/World Bank Road, Owerri, Imo State, Nigeria", "ML0030"),
        ]
        records, reconciliation = generator.build_outputs(rows, "2026-09-09")
        self.assertEqual(len(records), 2)
        self.assertEqual({r["id"] for r in records}, {
            "everight-diagnostic-and-laboratory-services-limited-fct-no-2-fez-street-ml0014",
            "everight-diagnostic-and-laboratory-services-limited-imo-20-asumpta-world-bank-road-ml0030",
        })
        self.assertEqual([r["decision"] for r in sorted(reconciliation, key=lambda r: r["source_position"])], ["retain", "retain"])

    def test_exact_duplicate_requires_all_meaningful_fields_to_match(self):
        first = fixture_row(1, "Example Laboratory", "Yaba, Lagos State, Nigeria", "ML0999")
        duplicate = dict(first)
        duplicate["source_position"] = 2
        records, reconciliation = generator.build_outputs([first, duplicate], "2026-09-09")
        self.assertEqual(len(records), 1)
        by_position = {row["source_position"]: row for row in reconciliation}
        self.assertEqual(by_position[1]["decision"], "retain")
        self.assertEqual(by_position[2]["decision"], "merge_exact_duplicate")
        self.assertEqual(by_position[2]["merge_target_id"], by_position[1]["final_record_id"])

    def test_repeated_annual_observation_merges_to_latest_certificate(self):
        old = fixture_row(1, "Example Laboratory", "Yaba, Lagos State, Nigeria", "ML0999", "2021-01-01", "2025-01-01")
        new = fixture_row(2, "Example Laboratory", "Yaba, Lagos State, Nigeria", "ML0999", "2026-01-01", "2030-01-01")
        records, reconciliation = generator.build_outputs([old, new], "2026-09-09")
        self.assertEqual(len(records), 1)
        self.assertEqual(records[0]["approval_date"], "2026-01-01")
        self.assertEqual(records[0]["expiry_date"], "2030-01-01")
        by_position = {row["source_position"]: row for row in reconciliation}
        self.assertEqual(by_position[1]["decision"], "merge_same_premises")
        self.assertEqual(by_position[2]["decision"], "retain")

    def test_expired_status_uses_retrieval_date(self):
        expired = fixture_row(1, "Expired Laboratory", "Abuja, Nigeria", "ML0888", "2021-01-01", "2025-01-01")
        current = fixture_row(2, "Current Laboratory", "Lagos State, Nigeria", "ML0889", "2026-01-01", "2030-01-01")
        records, _ = generator.build_outputs([expired, current], "2026-09-09")
        statuses = {record["accreditation_number"]: record["accreditation_status"] for record in records}
        self.assertEqual(statuses["ML0888"], "expired")
        self.assertEqual(statuses["ML0889"], "accredited")

    def test_deterministic_reconciliation_replay_is_byte_identical_twice(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            tmp = Path(tmpdir)
            for relative in [
                "datasets/healthcare/medical_laboratory_accreditations.json",
                "datasets/schemas/healthcare/medical_laboratory_accreditations.schema.json",
                "datasets/metadata/healthcare/medical_laboratory_accreditations.json",
                "datasets/metadata/healthcare/medical_laboratory_accreditations_reconciliation",
            ]:
                source = REPO / relative
                target = tmp / relative
                if source.is_dir():
                    shutil.copytree(source, target)
                else:
                    target.parent.mkdir(parents=True, exist_ok=True)
                    shutil.copy2(source, target)
            before = snapshot_bytes(tmp)
            subprocess.run(["python3", str(SCRIPT), "--repo", str(tmp), "--from-reconciliation", "--retrieved-at", "2026-09-09"], check=True)
            after_once = snapshot_bytes(tmp)
            subprocess.run(["python3", str(SCRIPT), "--repo", str(tmp), "--from-reconciliation", "--retrieved-at", "2026-09-09"], check=True)
            after_twice = snapshot_bytes(tmp)
            self.assertEqual(before, after_once)
            self.assertEqual(after_once, after_twice)

    def test_no_stale_licensing_contract_or_personal_fields_remain(self):
        paths = [
            REPO / "datasets/healthcare/medical_laboratory_accreditations.json",
        ] + sorted((REPO / "datasets/metadata/healthcare/medical_laboratory_accreditations_reconciliation").glob("*.json"))
        forbidden = tuple(
            "".join(parts)
            for parts in (
                ("superintendent",),
                ("phone", "_", "number"),
                ("email", "_", "address"),
                ("personal", "_", "registration"),
                ("practitioner", "_", "record"),
                ("registration", "_", "status"),
                ("licence", "_", "year"),
                ("lga", "_", "id"),
                ("laboratory", "_", "type"),
                ("ownership", "_", "type"),
                ("mlscn", "_", "premises", "_", "number"),
                ("licensed", "_", "medical", "_", "laboratories"),
                ("ng", "-", "licensed", "-", "medical", "-", "laboratories"),
            )
        )
        for path in paths:
            value = json.loads(path.read_text(encoding="utf-8"))
            text = json.dumps(value, ensure_ascii=False).lower()
            for marker in forbidden:
                self.assertNotIn(marker, text, path)


def snapshot_bytes(root):
    paths = [
        "datasets/healthcare/medical_laboratory_accreditations.json",
        "datasets/schemas/healthcare/medical_laboratory_accreditations.schema.json",
        "datasets/metadata/healthcare/medical_laboratory_accreditations.json",
        "datasets/metadata/healthcare/medical_laboratory_accreditations_reconciliation/index.json",
    ]
    recon_dir = root / "datasets/metadata/healthcare/medical_laboratory_accreditations_reconciliation"
    paths.extend(str(path.relative_to(root)) for path in sorted(recon_dir.glob("*.json")) if path.name != "index.json")
    return {path: (root / path).read_bytes() for path in sorted(paths)}


if __name__ == "__main__":
    unittest.main()
