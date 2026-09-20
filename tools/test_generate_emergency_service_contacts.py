import json
import tempfile
import unittest
from pathlib import Path

import generate_emergency_service_contacts as gen


class EmergencyServiceContactsGeneratorTest(unittest.TestCase):
    def test_build_artifacts_contract_and_privacy(self):
        artifacts = gen.build_artifacts(Path("."))
        records = json.loads(artifacts["datasets/emergency/emergency_service_contacts.json"])
        metadata = json.loads(artifacts["datasets/metadata/emergency/emergency_service_contacts.json"])
        reconciliation = json.loads(artifacts["datasets/metadata/emergency/emergency_service_contacts_reconciliation/index.json"])
        partition = json.loads(artifacts["datasets/metadata/emergency/emergency_service_contacts_reconciliation/national.json"])

        self.assertEqual(len(records), 5)
        self.assertEqual(metadata["record_count"], len(records))
        self.assertEqual(metadata["source_rows"], 5)
        self.assertEqual(metadata["decision_counts"], {"retain": 5})
        self.assertEqual(reconciliation["final_record_count"], len(records))
        self.assertEqual(reconciliation["source_rows"], len(partition["records"]))
        self.assertEqual(reconciliation["source_rows"] - reconciliation["decision_counts"].get("merge_exact_duplicate", 0), len(records))

        ids = [record["id"] for record in records]
        self.assertEqual(ids, sorted(ids))
        self.assertEqual(len(ids), len(set(ids)))
        for record in records:
            self.assertEqual(record["country_code"], "NG")
            self.assertIn(record["service_type"], gen.SERVICE_TYPES)
            self.assertIn(record["contact_type"], gen.CONTACT_TYPES)
            self.assertIn(record["coverage_type"], gen.COVERAGE_TYPES)
            self.assertNotIn("state_id", record)
            for forbidden in ("email", "address", "website", "person", "staff", "cookie", "token"):
                self.assertNotIn(forbidden, record)

        values = {record["contact_value"] for record in records}
        self.assertIn("112", values)
        self.assertIn("122", values)
        self.assertIn("080022556362", values)
        self.assertIn("+2348032003557", values)

        public_json = json.dumps(records).lower()
        for forbidden in ("@", "password", "api_key", "secret", "cookie", "victim", "patient", "caller"):
            self.assertNotIn(forbidden, public_json)

    def test_reconciliation_targets_and_partition_hashes(self):
        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            gen.write_outputs(repo)
            index_path = repo / "datasets/metadata/emergency/emergency_service_contacts_reconciliation/index.json"
            index = json.loads(index_path.read_text())
            records = {
                row["id"]
                for row in json.loads((repo / "datasets/emergency/emergency_service_contacts.json").read_text())
            }
            seen_source_positions = set()
            for partition in index["partitions"]:
                path = repo / "datasets" / partition["path"]
                data = path.read_bytes()
                self.assertEqual(len(data), partition["size_bytes"])
                self.assertEqual(gen.sha256_bytes(data), partition["sha256"])
                value = json.loads(data)
                for row in value["records"]:
                    key = (row["source_key"], row["source_position"])
                    self.assertNotIn(key, seen_source_positions)
                    seen_source_positions.add(key)
                    if row["decision"] == "retain":
                        self.assertIn(row["target_id"], records)
                    else:
                        self.assertIsNone(row["target_id"])
            self.assertEqual(len(seen_source_positions), index["source_rows"])

    def test_deterministic_generation_is_byte_identical(self):
        with tempfile.TemporaryDirectory() as tmp:
            first = Path(tmp) / "first"
            second = Path(tmp) / "second"
            gen.write_outputs(first)
            gen.write_outputs(second)
            first_files = {path.relative_to(first).as_posix(): path.read_bytes() for path in sorted(first.rglob("*.json"))}
            second_files = {path.relative_to(second).as_posix(): path.read_bytes() for path in sorted(second.rglob("*.json"))}
            self.assertEqual(first_files, second_files)

    def test_committed_artifacts_match_generator(self):
        repo = Path(__file__).resolve().parents[1]
        for rel, expected in gen.build_artifacts(repo).items():
            self.assertEqual((repo / rel).read_bytes(), expected, rel)


if __name__ == "__main__":
    unittest.main()
