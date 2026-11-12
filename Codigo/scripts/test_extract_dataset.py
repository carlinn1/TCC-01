import json
import sys
import tempfile
import unittest
from datetime import datetime, timezone
from pathlib import Path

SCRIPTS_DIR = Path(__file__).resolve().parent
if str(SCRIPTS_DIR) not in sys.path:
    sys.path.insert(0, str(SCRIPTS_DIR))

import extract_dataset as ed  # noqa: E402


class TestExtractDataset(unittest.TestCase):
    def test_compute_engineering_start_takes_min_of_pr_commit_and_assigned(self):
        created = datetime(2024, 1, 10, 12, 0, tzinfo=timezone.utc)
        merged = datetime(2024, 1, 12, 12, 0, tzinfo=timezone.utc)
        # First commit is earlier in -03:00 timezone (11:00 UTC on Jan 9)
        fc_iso = "2024-01-09T08:00:00-03:00"
        # Issue assigned is even earlier (12:00 UTC on Jan 8)
        assigned = datetime(2024, 1, 8, 12, 0, tzinfo=timezone.utc)

        res = ed.compute_engineering_start(created, merged, fc_iso, assigned)
        self.assertTrue(res["commit_precedes_pr"])
        self.assertTrue(res["issue_assigned_precedes_pr"])
        self.assertGreaterEqual(res["raw_eng_bd"], res["ct_bd"] - 1e-9)
        self.assertAlmostEqual(res["ct_bd"], 2.0, places=6)
        self.assertAlmostEqual(res["raw_eng_bd"], 4.0, places=6)

    def test_compute_engineering_start_never_less_than_created_when_commit_is_later(self):
        created = datetime(2024, 1, 10, 12, 0, tzinfo=timezone.utc)
        merged = datetime(2024, 1, 12, 12, 0, tzinfo=timezone.utc)
        later_fc_iso = "2024-01-11T10:00:00Z"

        res = ed.compute_engineering_start(created, merged, later_fc_iso, None)
        self.assertFalse(res["commit_precedes_pr"])
        self.assertFalse(res["issue_assigned_precedes_pr"])
        self.assertAlmostEqual(res["raw_eng_bd"], res["ct_bd"], places=9)

    def test_extract_closing_issue_numbers_parses_hash_and_url(self):
        body = "Fixes #1234 and closes https://github.com/cli/cli/issues/5678\nAlso Refs #999"
        nums = ed.extract_closing_issue_numbers(body, "cli/cli")
        self.assertEqual(nums, [1234, 5678])

    def test_verify_local_dataset_rejects_raw_lt_created(self):
        with tempfile.TemporaryDirectory() as tmp:
            bad_path = Path(tmp) / "bad.json"
            valid_item = {
                "item_id": "PR_01",
                "created_to_merged_bd": 2.0,
                "first_commit_to_merged_raw_bd": 2.0,
                "first_commit_to_merged_floored_bd": 2.0,
                "at_floor": False,
                "commit_precedes_pr": False,
                "has_issue_assigned_at": False,
                "issue_assigned_precedes_pr": False,
                "first_commit_source": "rest_api",
                "squash_or_rebase_checked": True,
                "has_closing_keyword": False,
                "labels": [],
                "fs_edges": [],
            }
            bad_item = dict(valid_item, first_commit_to_merged_raw_bd=1.0, first_commit_to_merged_floored_bd=1.0, commit_precedes_pr=True)

            def make_case(target, item):
                dur = item["created_to_merged_bd"]
                eng = item["first_commit_to_merged_floored_bd"]
                return {
                    "target": target,
                    "prs_merged_non_bot": 1,
                    "prs_with_closing_keyword": 0,
                    "prs_with_commit_before_pr": 1 if item["commit_precedes_pr"] else 0,
                    "prs_with_issue_assigned_before_pr": 0,
                    "cycle_time_business_days": {
                        "n": 1, "p5": dur, "p10": dur, "p50": dur, "p90": dur, "p95": dur, "mean": dur
                    },
                    "engineering_duration_business_days": {
                        "n": 1, "p5": eng, "p10": eng, "p50": eng, "p90": eng, "p95": eng, "mean": eng
                    },
                    "floor_count": 0,
                    "floor_fraction": 0.0,
                    "floor_sensitivity": {},
                    "items": [item],
                }

            bad_data = {
                "repo": "cli/cli",
                "cases": [
                    make_case("c1", bad_item),
                    make_case("c2", valid_item),
                    make_case("c3", valid_item),
                ],
            }
            bad_path.write_text(json.dumps(bad_data), encoding="utf-8")
            with self.assertRaisesRegex(AssertionError, "created_to_merged_bd"):
                ed.verify_local_dataset(str(bad_path))


if __name__ == "__main__":
    unittest.main()
