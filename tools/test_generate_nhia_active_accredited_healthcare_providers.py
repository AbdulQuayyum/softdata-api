import json
import tempfile
import unittest
from pathlib import Path

import generate_nhia_active_accredited_healthcare_providers as gen


class NHIAActiveAccreditedHealthcareProvidersGeneratorTest(unittest.TestCase):
    def test_safe_chunk_extraction_discards_address(self):
        payload = json.dumps(
            [
                {
                    "value": {
                        "healthcareprovidercode": "la/1632 /p",
                        "healthcareprovidername": "Example Clinic - LA/1632/P",
                        "facilitytype": "PRIMARY AND SECONDARY",
                        "address": "Synthetic discarded location placeholder",
                    }
                }
            ]
        ).encode()
        rows = gen.safe_rows_from_chunk(payload, 0, 0)
        self.assertEqual(rows[0]["source_provider_code"], "la/1632 /p")
        self.assertNotIn("address", rows[0])
        self.assertNotIn("Synthetic Person", json.dumps(rows))
        self.assertNotIn("discarded location", json.dumps(rows))

    def test_code_and_facility_type_normalization(self):
        code, rules = gen.normalize_code(" Gentle Hearts Global Harvest Medical Centre- OY/0013/P ")
        self.assertEqual(code, "OY/0013/P")
        self.assertIn("extract_unambiguous_code", rules)
        code, rules = gen.normalize_code("LA/1632 /P")
        self.assertEqual(code, "LA/1632/P")
        self.assertIn("collapse_separator_whitespace", rules)
        self.assertEqual(gen.normalize_facility_type("PRIMARY AND SECONDARY"), "primary_and_secondary")
        self.assertEqual(gen.normalize_facility_type("Primary"), "primary")

    def test_build_outputs_excludes_conflicting_duplicate_codes(self):
        rows = []
        for i in range(1, gen.EXPECTED_RAW_ROWS + 1):
            rows.append(
                {
                    "source_position": i,
                    "chunk_number": 0,
                    "chunk_row_position": i,
                    "source_provider_code": f"LA/{i:04d}/P",
                    "source_provider_name": f"Provider {i} - LA/{i:04d}/P",
                    "source_facility_type": "Primary",
                }
            )
        rows[1]["source_provider_code"] = rows[0]["source_provider_code"]
        rows[1]["source_provider_name"] = "Different Provider"
        records, reconciliation, analysis = gen.build_outputs(rows)
        self.assertEqual(len(records), gen.EXPECTED_RAW_ROWS - 2)
        self.assertEqual(analysis["decision_counts"]["exclude_unresolved_code_conflict"], 2)
        self.assertEqual(sum(1 for row in reconciliation if row["target_id"] is None), 2)

    def test_public_records_keep_only_approved_fields(self):
        rows = []
        for i in range(1, gen.EXPECTED_RAW_ROWS + 1):
            rows.append(
                {
                    "source_position": i,
                    "chunk_number": i // 3000,
                    "chunk_row_position": i,
                    "source_provider_code": f"AB/{i:04d}/P",
                    "source_provider_name": f"Provider {i} - AB/{i:04d}/P",
                    "source_facility_type": "Primary",
                }
            )
        records, reconciliation, _ = gen.build_outputs(rows)
        for record in records:
            self.assertEqual(set(record), {"id", "name", "country_code", "provider_code", "facility_type", "listing_status"})
        for row in reconciliation:
            self.assertNotIn("address", row)
            self.assertNotIn("phone", row)
            self.assertNotIn("email", row)

    def test_deterministic_reconciliation_replay_is_byte_identical(self):
        rows = []
        for i in range(1, gen.EXPECTED_RAW_ROWS + 1):
            rows.append(
                {
                    "source_position": i,
                    "chunk_number": i // 3000,
                    "chunk_row_position": i,
                    "source_provider_code": f"AB/{i:04d}/P",
                    "source_provider_name": f"Provider {i} - AB/{i:04d}/P",
                    "source_facility_type": "Primary",
                }
            )
        source_retrieval = {"synthetic": True}
        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            gen.generate(repo, rows, source_retrieval)
            first = {path.relative_to(repo).as_posix(): path.read_bytes() for path in sorted(repo.rglob("*.json"))}
            replay_rows, replay_source = gen.rows_from_reconciliation(repo)
            gen.generate(repo, replay_rows, replay_source)
            second = {path.relative_to(repo).as_posix(): path.read_bytes() for path in sorted(repo.rglob("*.json"))}
        self.assertEqual(replay_source, source_retrieval)
        self.assertEqual(first, second)


if __name__ == "__main__":
    unittest.main()
