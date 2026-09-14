import json
import tempfile
import unittest
from pathlib import Path

import generate_nhia_state_social_health_insurance_agencies as gen


class NHIAStateSocialHealthInsuranceAgenciesGeneratorTest(unittest.TestCase):
    def test_state_aliases_cover_source_values(self):
        for value in ["ABIA", "ADAMAWA", "AKS", "ANAMBRA", "CRS", "FCT", "GOMBE", "KATSINA", "NASARAWA", "RIVERS", "River State"]:
            self.assertIn(gen.normalize_state(value), gen.CANONICAL_STATES)
        self.assertEqual(gen.normalize_state("AKS"), "akwa-ibom")
        self.assertEqual(gen.normalize_state("CRS"), "cross-river")
        self.assertEqual(gen.normalize_state("FCT"), "fct")
        self.assertEqual(gen.normalize_state("River State"), "rivers")
        with self.assertRaises(ValueError):
            gen.normalize_state("UNKNOWN")

    def test_build_outputs_keeps_only_safe_public_fields(self):
        rows = [{"source_position": i + 1, "raw_organisation_name": f"{state.title()} State Health Insurance Agency", "raw_state": state.replace("-", " ").upper()} for i, state in enumerate(sorted(gen.CANONICAL_STATES))]
        records, reconciliation = gen.build_outputs(rows, "2026-09-13")
        self.assertEqual(len(records), 37)
        self.assertEqual({record["state_id"] for record in records}, gen.CANONICAL_STATES)
        self.assertEqual({record["organisation_type"] for record in records}, {gen.ORGANISATION_TYPE})
        self.assertEqual({record["country_code"] for record in records}, {"NG"})
        for record in records:
            self.assertEqual(set(record), {"id", "name", "state_id", "country_code", "organisation_type"})
        for row in reconciliation:
            self.assertEqual(set(row), {"source_position", "raw_organisation_name", "raw_state", "normalized_name", "state_id", "final_record_id", "decision", "reason", "safe_evidence"})

    def test_duplicate_state_is_rejected_by_final_coverage(self):
        rows = [{"source_position": i + 1, "raw_organisation_name": f"{state.title()} State Health Insurance Agency", "raw_state": state.replace("-", " ").upper()} for i, state in enumerate(sorted(gen.CANONICAL_STATES))]
        rows[-1]["raw_state"] = rows[0]["raw_state"]
        with self.assertRaisesRegex(ValueError, "unexpected retained SSHIA count"):
            gen.build_outputs(rows, "2026-09-13")

    def test_extracts_safe_columns_and_discards_contacts(self):
        html = """
        <table><tr><th>S/N</th><th>ORGANIZATION</th><th>STATE</th><th>DIRECTOR</th><th>PHONE NO.</th><th>E-MAIL ADDRESS</th><th>WEBSITES</th><th>ADDRESS</th></tr>
        """
        for i, state in enumerate(sorted(gen.CANONICAL_STATES), start=1):
            html += f"<tr><td>{i}</td><td>{state.title()} State Health Insurance Agency</td><td>{state.replace('-', ' ').upper()}</td><td>Synthetic Person</td><td>00000000000</td><td>synthetic@example.invalid</td><td>Example</td><td>Synthetic Address</td></tr>"
        html += "</table>"
        with tempfile.TemporaryDirectory() as tmp:
            source = Path(tmp) / "source.html"
            source.write_text(html, encoding="utf-8")
            old_hash = gen.EXPECTED_SAFE_TABLE_SHA256
            rows = []
            try:
                parsed = gen.TableParser()
                parsed.feed(source.read_text())
                rows = [{"source_position": int(row[0]), "raw_organisation_name": row[1], "raw_state": row[2]} for row in parsed.rows[1:]]
                gen.EXPECTED_SAFE_TABLE_SHA256 = gen.sha256_bytes(gen.compact_json(rows).encode("utf-8"))
                loaded = gen.load_source_rows(source)
            finally:
                gen.EXPECTED_SAFE_TABLE_SHA256 = old_hash
        self.assertEqual(loaded, rows)
        self.assertNotIn("Synthetic Person", json.dumps(loaded))
        self.assertNotIn("synthetic@example.invalid", json.dumps(loaded))

    def test_deterministic_reconciliation_replay_is_byte_identical_twice(self):
        rows = [{"source_position": i + 1, "raw_organisation_name": f"{state.title()} State Health Insurance Agency", "raw_state": state.replace("-", " ").upper()} for i, state in enumerate(sorted(gen.CANONICAL_STATES))]
        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            gen.generate(repo, rows, "2026-09-13")
            first = {path.relative_to(repo).as_posix(): path.read_bytes() for path in sorted(repo.rglob("*.json"))}
            replay = gen.rows_from_reconciliation(repo)
            gen.generate(repo, replay, "2026-09-13")
            second = {path.relative_to(repo).as_posix(): path.read_bytes() for path in sorted(repo.rglob("*.json"))}
        self.assertEqual(first, second)


if __name__ == "__main__":
    unittest.main()
