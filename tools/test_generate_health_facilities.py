"""Offline regression fixtures and full regeneration checks for healthcare."""
import copy
import json
import hashlib
from collections import Counter
import math
import shutil
import tempfile
import unittest
from pathlib import Path

import generate_health_facilities as gen

ROOT = Path(__file__).resolve().parents[1]


class HealthFacilityGeneratorTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.reviews = json.loads((ROOT / gen.REVIEW_PATH).read_text())
        cls.states, cls.lgas = gen.load_geography(ROOT)

    def fixture(self, pair):
        review = copy.deepcopy(pair)
        review["source_positions"] = [1, 2]
        for i, obs in enumerate(review["observations"], 1):
            obs["source_position"] = i
        return [{"attributes": obs["raw_attributes"].copy()} for obs in review["observations"]], review

    def test_all_five_pairs_require_review_and_are_quarantined(self):
        self.assertEqual(len(self.reviews["pairs"]), 5)
        for pair in self.reviews["pairs"]:
            with self.subTest(pair=pair["pair_id"]):
                features, review = self.fixture(pair)
                self.assertFalse(gen.exact_duplicate(*(f["attributes"] for f in features)))
                with self.assertRaisesRegex(ValueError, "require identity review"):
                    gen.reviewed_decisions(features, self.states, self.lgas, {"pairs": []})
                decisions = gen.reviewed_decisions(features, self.states, self.lgas, {"pairs": [review]})
                self.assertEqual(set(decisions), {1, 2})
                self.assertTrue(all(r["final_decision"] == "exclude_unresolved_identity" for r in decisions.values()))
                self.assertEqual(review["final_record_ids"], [])
                a, b = [f["attributes"] for f in features]
                lat1, lat2 = map(math.radians, [a["latitude"], b["latitude"]])
                delta_lat = lat2 - lat1
                delta_lon = math.radians(b["longitude"] - a["longitude"])
                metres = 2 * 6371008.8 * math.asin(math.sqrt(math.sin(delta_lat / 2)**2 + math.cos(lat1) * math.cos(lat2) * math.sin(delta_lon / 2)**2))
                self.assertAlmostEqual(metres, review["distance_metres"], places=3)

    def test_name_equality_never_hides_conflicting_attributes(self):
        base = self.reviews["pairs"][0]["observations"][0]["raw_attributes"]
        for field, value in [("latitude", base["latitude"] + .01), ("ownership", "Private"),
                             ("facility_level_option", "General Hospital"), ("facility_level", "Secondary"),
                             ("ward", "Another Ward"), ("lga", "Esan South-East"), ("state", "Lagos")]:
            with self.subTest(field=field):
                other = dict(base, **{field: value})
                self.assertFalse(gen.exact_duplicate(base, other))
                # Same stable source ID with conflicting geography is also reviewed.
                with self.assertRaisesRegex(ValueError, "require identity review"):
                    gen.reviewed_decisions([{"attributes": base}, {"attributes": other}], self.states, self.lgas, {"pairs": []})

    def test_every_meaningful_source_field_is_compared(self):
        base = self.reviews["pairs"][0]["observations"][0]["raw_attributes"]
        self.assertTrue(gen.exact_duplicate(base, dict(base, OBJECTID=999999)))
        for key in base:
            if key == "OBJECTID":
                continue
            other = dict(base, **{key: None if base[key] is not None else "changed"})
            self.assertFalse(gen.exact_duplicate(base, other), key)

    def test_stale_or_missing_review_evidence_rejected(self):
        features, review = self.fixture(self.reviews["pairs"][0])
        features[1]["attributes"]["longitude"] += .1
        with self.assertRaisesRegex(ValueError, "source changed"):
            gen.reviewed_decisions(features, self.states, self.lgas, {"pairs": [review]})
        features, review = self.fixture(self.reviews["pairs"][0])
        review["evidence_sources"] = []
        with self.assertRaisesRegex(ValueError, "evidence"):
            gen.reviewed_decisions(features, self.states, self.lgas, {"pairs": [review]})

    def output_root(self, parent, name):
        root = parent / name
        (root / "datasets/geography").mkdir(parents=True)
        for name in ("states", "lgas"):
            shutil.copyfile(ROOT / f"datasets/geography/{name}.json", root / f"datasets/geography/{name}.json")
        return root

    def test_separate_campuses_have_meaningful_deterministic_ids(self):
        features, review = self.fixture(self.reviews["pairs"][0])
        review["final_decision"] = "retain_separate_facility"
        review["reason"] = "Synthetic regression fixture: two independently verified campuses."
        review["disambiguators"] = {features[0]["attributes"]["globalid"]: "north-campus", features[1]["attributes"]["globalid"]: "south-campus"}
        with tempfile.TemporaryDirectory() as directory:
            parent = Path(directory)
            source = parent / "source.json"
            source.write_text(json.dumps({"features": features}))
            reviews = parent / "reviews.json"
            reviews.write_text(json.dumps({"review_date": "2026-09-09", "pairs": [review]}))
            first = self.output_root(parent, "first")
            second = self.output_root(parent, "second")
            for root in (first, second):
                gen.generate(source, root, "2026-09-08", review_path=reviews)
            data = json.loads((first / "datasets/healthcare/health_facilities.json").read_text())
            self.assertEqual(len(data), 2)
            self.assertEqual(len({r["id"] for r in data}), 2)
            self.assertTrue(any(r["id"].endswith("north-campus") for r in data))
            self.assertTrue(any(r["id"].endswith("south-campus") for r in data))
            self.assertEqual((first / "datasets/healthcare/health_facilities.json").read_bytes(), (second / "datasets/healthcare/health_facilities.json").read_bytes())

    def test_true_exact_source_duplicate_merges(self):
        raw = self.reviews["pairs"][0]["observations"][0]["raw_attributes"].copy()
        with tempfile.TemporaryDirectory() as directory:
            parent = Path(directory)
            source = parent / "source.json"
            source.write_text(json.dumps({"features": [{"attributes": raw}, {"attributes": dict(raw, OBJECTID=999999)}]}))
            reviews = parent / "reviews.json"
            reviews.write_text(json.dumps({"review_date": "2026-09-09", "pairs": []}))
            result = gen.generate(source, self.output_root(parent, "output"), "2026-09-08", review_path=reviews)
            self.assertEqual(result[0:2], (2, 1))
            self.assertEqual(result[2]["merge_exact_duplicate"], 1)

    def test_committed_source_decisions_are_complete_and_bijective(self):
        dataset = json.loads((ROOT / "datasets/healthcare/health_facilities.json").read_text())
        ids = {record["id"] for record in dataset}
        self.assertEqual(len(ids), len(dataset))
        self.assertEqual(len({r["source_facility_id"] for r in dataset}), len(dataset))
        metadata = json.loads((ROOT / "datasets/metadata/healthcare/health_facilities.json").read_text())
        index = json.loads((ROOT / "datasets/metadata/healthcare/health_facilities_reconciliation/index.json").read_text())
        rows = []
        for part in index["partitions"]:
            data = (ROOT / "datasets" / part["path"]).read_bytes()
            self.assertEqual(len(data), part["size_bytes"])
            self.assertEqual(hashlib.sha256(data).hexdigest(), part["sha256"])
            partition = json.loads(data)
            self.assertEqual(dict(Counter(row["decision"] for row in partition["records"])), partition["decision_counts"])
            rows.extend(partition["records"])
        self.assertEqual(len(rows), 51022)
        self.assertEqual({row["source_position"] for row in rows}, set(range(1, 51023)))
        self.assertEqual(len({row["raw_object_id"] for row in rows}), 51022)
        counts = dict(Counter(row["decision"] for row in rows))
        self.assertEqual(counts, metadata["decision_counts"])
        self.assertEqual(counts, index["decision_counts"])
        retained = [row for row in rows if row["decision"] == "retain"]
        self.assertEqual(len(retained), len(dataset))
        self.assertEqual({row["final_record_id"] for row in retained}, ids)
        self.assertTrue(all(row["evidence"] for row in retained))
        reviewed_positions = {pos for pair in self.reviews["pairs"] for pos in pair["source_positions"]}
        self.assertEqual({row["source_position"] for row in rows if row["decision"] == "exclude_unresolved_identity"}, reviewed_positions)
        self.assertTrue(all(row["final_record_id"] is None for row in rows if row["decision"] != "retain"))
        self.assertTrue(all(row["merge_target_id"] is None for row in rows))
        self.assertEqual(counts.get("merge_exact_duplicate", 0), 0)
        self.assertEqual(counts.get("merge_same_facility", 0), 0)
        self.assertEqual(metadata["record_count"], index["final_record_count"])
        self.assertEqual(metadata["record_count"], len(dataset))
        self.assertEqual(metadata["coordinate_coverage_count"], len(dataset))
        self.assertEqual(metadata["state_coverage_count"], len({r["state_id"] for r in dataset}))
        for field, key in [("facility_type", "facility_type_counts"), ("facility_level", "facility_level_counts"), ("ownership_type", "ownership_counts")]:
            self.assertEqual(dict(Counter(r.get(field, "<omitted>") for r in dataset)), metadata[key])

    def test_full_offline_regeneration_twice_is_byte_identical(self):
        with tempfile.TemporaryDirectory() as directory:
            parent = Path(directory)
            for name in ("first", "second"):
                output = self.output_root(parent, name)
                total, retained, decisions = gen.generate(None, output, "2026-09-08", reconciliation_repo=ROOT)
                self.assertEqual((total, retained), (51022, 50649))
                self.assertEqual(decisions, {"retain": 50649, "exclude_invalid_geography": 363, "exclude_unresolved_identity": 10})
                for file in (output / "datasets").rglob("*.json"):
                    self.assertEqual(file.read_bytes(), (ROOT / file.relative_to(output)).read_bytes(), str(file.relative_to(output)))


if __name__ == "__main__":
    unittest.main()
