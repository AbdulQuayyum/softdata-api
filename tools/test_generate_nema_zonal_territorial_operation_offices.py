import json
import tempfile
import unittest
from pathlib import Path

import generate_nema_zonal_territorial_operation_offices as gen


class NEMAZonalTerritorialOperationOfficesGeneratorTest(unittest.TestCase):
    def test_artifacts_are_deterministic_and_closed(self):
        artifacts = gen.build_outputs()
        records = json.loads(artifacts["datasets/emergency/nema_zonal_territorial_operation_offices.json"])
        metadata = json.loads(artifacts["datasets/metadata/emergency/nema_zonal_territorial_operation_offices.json"])
        index = json.loads(artifacts["datasets/metadata/emergency/nema_zonal_territorial_operation_offices_reconciliation/index.json"])

        self.assertEqual(17, len(records))
        self.assertEqual("ng-nema-zonal-territorial-operation-offices", metadata["dataset_key"])
        self.assertEqual("Zonal, Territorial and Operation offices", metadata["official_terminology"])
        self.assertEqual({"retain": 17}, metadata["decision_counts"])
        self.assertEqual(17, index["source_rows"])
        self.assertEqual(17, index["final_record_count"])
        self.assertEqual({"retain": 17}, index["decision_counts"])
        self.assertEqual(17, len(index["partitions"]))
        self.assertEqual(17, len({row["id"] for row in records}))
        self.assertEqual(records, sorted(records, key=lambda row: row["id"]))

        for record in records:
            self.assertEqual("NG", record["country_code"])
            self.assertEqual("zonal_territorial_operation_office", record["office_type"])
            self.assertIn(record["state_id"], metadata["state_counts"])
            self.assertEqual({"id", "name", "office_type", "state_id", "country_code"}, set(record))

    def test_check_mode_matches_committed_artifacts(self):
        artifacts = gen.build_outputs()
        repo = Path(__file__).resolve().parents[1]
        mismatches = []
        for relative, expected in artifacts.items():
            if (repo / relative).read_bytes() != expected:
                mismatches.append(relative)
        self.assertEqual([], mismatches)

    def test_privacy_markers_are_absent(self):
        combined = "\n".join(data.decode().lower() for data in gen.build_outputs().values())
        for marker in ("@", "password", "api_key", "secret", "token", "victim", "patient", "whatsapp"):
            self.assertNotIn(marker, combined)
        for source_name in ("zubaida", "director general"):
            self.assertNotIn(source_name, combined)

    def test_write_outputs_to_tempdir(self):
        with tempfile.TemporaryDirectory() as tmp:
            gen.write_outputs(Path(tmp), gen.build_outputs())
            generated = Path(tmp) / "datasets/emergency/nema_zonal_territorial_operation_offices.json"
            self.assertTrue(generated.exists())
            self.assertEqual(17, len(json.loads(generated.read_text())))


if __name__ == "__main__":
    unittest.main()
