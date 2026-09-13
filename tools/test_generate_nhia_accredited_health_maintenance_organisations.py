import importlib.util
import json
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path

SCRIPT = Path(__file__).with_name("generate_nhia_accredited_health_maintenance_organisations.py").resolve()
REPO = Path(__file__).resolve().parents[1]

spec = importlib.util.spec_from_file_location("generate_nhia_accredited_health_maintenance_organisations", SCRIPT)
generator = importlib.util.module_from_spec(spec)
spec.loader.exec_module(generator)


def fixture_row(position, name, hmo_id, address=""):
    return {
        "source_position": position,
        "source_name": name,
        "source_hmo_id": hmo_id,
        "source_category": generator.SOURCE_TABLE_TITLE,
        "source_accreditation_status": generator.SOURCE_STATUS,
        "raw_organisation_level_location": address,
        "normalized_name": generator.slug(name),
        "normalized_hmo_id": hmo_id,
    }


class NHIAAccreditedHMOGeneratorTest(unittest.TestCase):
    def test_build_outputs_keeps_hmo_ids_as_strings(self):
        records, reconciliation = generator.build_outputs([fixture_row(1, "Example HMO Limited", "007")], "2026-09-11")
        self.assertEqual(records, [{
            "id": "example-hmo-limited-007",
            "name": "Example HMO Limited",
            "country_code": "NG",
            "organisation_type": "health_maintenance_organisation",
            "accreditation_status": "accredited",
            "hmo_id": "007",
        }])
        self.assertEqual(reconciliation[0]["decision"], "retain")
        self.assertEqual(reconciliation[0]["final_record_id"], "example-hmo-limited-007")

    def test_exact_duplicates_only_merge_on_name_and_hmo_id(self):
        rows = [
            fixture_row(1, "Example HMO Limited", "7", "Lagos"),
            fixture_row(2, "Example HMO Limited", "7", "Lagos"),
            fixture_row(3, "Example HMO Limited", "8", "Lagos"),
        ]
        records, reconciliation = generator.build_outputs(rows, "2026-09-11")
        self.assertEqual(len(records), 2)
        decisions = {row["source_position"]: row["decision"] for row in reconciliation}
        self.assertEqual(decisions, {1: "retain", 2: "merge_exact_duplicate", 3: "retain"})

    def test_invalid_identity_is_rejected(self):
        with self.assertRaises(ValueError):
            generator.build_outputs([fixture_row(1, "Example HMO Limited", "ABC")], "2026-09-11")

    def test_deterministic_reconciliation_replay_is_byte_identical_twice(self):
        with tempfile.TemporaryDirectory() as tmpdir:
            tmp = Path(tmpdir)
            for relative in [
                "datasets/healthcare/nhia_accredited_health_maintenance_organisations.json",
                "datasets/schemas/healthcare/nhia_accredited_health_maintenance_organisations.schema.json",
                "datasets/metadata/healthcare/nhia_accredited_health_maintenance_organisations.json",
                "datasets/metadata/healthcare/nhia_accredited_health_maintenance_organisations_reconciliation",
            ]:
                source = REPO / relative
                target = tmp / relative
                if source.is_dir():
                    shutil.copytree(source, target)
                else:
                    target.parent.mkdir(parents=True, exist_ok=True)
                    shutil.copy2(source, target)
            before = snapshot_bytes(tmp)
            subprocess.run(["python3", str(SCRIPT), "--repo", str(tmp), "--from-reconciliation", "--retrieved-at", "2026-09-11"], check=True)
            after_once = snapshot_bytes(tmp)
            subprocess.run(["python3", str(SCRIPT), "--repo", str(tmp), "--from-reconciliation", "--retrieved-at", "2026-09-11"], check=True)
            after_twice = snapshot_bytes(tmp)
            self.assertEqual(before, after_once)
            self.assertEqual(after_once, after_twice)

    def test_no_contact_or_personal_fields_remain(self):
        paths = [
            REPO / "datasets/healthcare/nhia_accredited_health_maintenance_organisations.json",
        ] + sorted((REPO / "datasets/metadata/healthcare/nhia_accredited_health_maintenance_organisations_reconciliation").glob("*.json"))
        forbidden = ("website", "email", "phone", "director", "practitioner", "patient", "enrollee", "policy_number")
        for path in paths:
            value = json.loads(path.read_text(encoding="utf-8"))
            text = json.dumps(value, ensure_ascii=False).lower()
            if path.name == "index.json":
                continue
            for marker in forbidden:
                self.assertNotIn(marker, text, path)


def snapshot_bytes(root):
    paths = [
        "datasets/healthcare/nhia_accredited_health_maintenance_organisations.json",
        "datasets/schemas/healthcare/nhia_accredited_health_maintenance_organisations.schema.json",
        "datasets/metadata/healthcare/nhia_accredited_health_maintenance_organisations.json",
        "datasets/metadata/healthcare/nhia_accredited_health_maintenance_organisations_reconciliation/index.json",
        "datasets/metadata/healthcare/nhia_accredited_health_maintenance_organisations_reconciliation/health_maintenance_organisations.json",
    ]
    return {path: (root / path).read_bytes() for path in paths}


if __name__ == "__main__":
    unittest.main()
