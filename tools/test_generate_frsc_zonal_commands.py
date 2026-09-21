import json
import tempfile
import unittest
from pathlib import Path

import generate_frsc_zonal_commands as gen


class FRSCZonalCommandsGeneratorTest(unittest.TestCase):
    def test_artifacts_are_deterministic_and_closed(self):
        artifacts = gen.build_outputs()
        records = json.loads(artifacts["datasets/emergency/frsc_zonal_commands.json"])
        metadata = json.loads(artifacts["datasets/metadata/emergency/frsc_zonal_commands.json"])
        index = json.loads(artifacts["datasets/metadata/emergency/frsc_zonal_commands_reconciliation/index.json"])

        self.assertEqual(12, len(records))
        self.assertEqual("ng-frsc-zonal-commands", metadata["dataset_key"])
        self.assertEqual("Zonal Commands", metadata["official_terminology"])
        self.assertEqual({"retain": 12}, metadata["decision_counts"])
        self.assertEqual({"zonal_command": 12}, metadata["command_type_counts"])
        self.assertEqual(12, index["source_rows"])
        self.assertEqual(12, index["final_record_count"])
        self.assertEqual({"retain": 12}, index["decision_counts"])
        self.assertEqual(12, len(index["partitions"]))
        self.assertEqual(12, len({row["id"] for row in records}))
        self.assertEqual(12, len({row["command_code"] for row in records}))
        self.assertEqual(records, sorted(records, key=lambda row: row["id"]))

        for record in records:
            self.assertEqual("NG", record["country_code"])
            self.assertEqual("zonal_command", record["command_type"])
            self.assertIn(record["state_id"], metadata["state_counts"])
            self.assertRegex(record["command_code"], r"^RS(?:[1-9]|1[0-2])HQ$")
            self.assertEqual({"id", "name", "command_type", "command_code", "state_id", "country_code"}, set(record))

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
        for marker in ("@", "password", "api_key", "secret", "token", "victim", "patient", "offender", "whatsapp"):
            self.assertNotIn(marker, combined)
        privacy = json.loads(gen.build_outputs()["datasets/metadata/emergency/frsc_zonal_commands.json"])["privacy"].lower()
        for omitted in ("commander", "phone", "email", "address"):
            self.assertIn(omitted, privacy)

    def test_write_outputs_to_tempdir(self):
        with tempfile.TemporaryDirectory() as tmp:
            gen.write_outputs(Path(tmp), gen.build_outputs())
            generated = Path(tmp) / "datasets/emergency/frsc_zonal_commands.json"
            self.assertTrue(generated.exists())
            self.assertEqual(12, len(json.loads(generated.read_text())))


if __name__ == "__main__":
    unittest.main()
