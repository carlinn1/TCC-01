import sys
import unittest
from pathlib import Path

SCRIPTS_DIR = Path(__file__).resolve().parent
if str(SCRIPTS_DIR) not in sys.path:
    sys.path.insert(0, str(SCRIPTS_DIR))

import power_h1b  # noqa: E402


class TestPowerH1B(unittest.TestCase):
    def test_exact_tost_m500_bounds_power_and_type_i(self):
        res = power_h1b.compute_exact_tost_power(
            m=500, p=0.85, alpha=0.05, lower_bound=0.80, upper_bound=0.90
        )
        self.assertEqual(res["k_min"], 416)
        self.assertEqual(res["k_max"], 438)
        self.assertAlmostEqual(res["power_pct"], 83.93, delta=0.02)
        self.assertLessEqual(res["type_i_error_lower_pct"], 5.0)
        self.assertLessEqual(res["type_i_error_upper_pct"], 5.0)
        self.assertAlmostEqual(res["type_i_error_lower_pct"], 3.94, delta=0.05)
        self.assertAlmostEqual(res["type_i_error_upper_pct"], 4.65, delta=0.05)

    def test_wilson_90_accepts_415_and_inflates_type_i_error(self):
        z_90 = 1.6448536269514722
        res = power_h1b.compute_exact_power(500, 0.85, z_90)
        self.assertEqual(res["k_min"], 415)
        self.assertEqual(res["k_max"], 438)
        self.assertGreater(res["type_i_error_lower_pct"], 5.0)

    def test_exact_tost_controls_type_i_error_for_all_sample_sizes(self):
        for m in [200, 300, 400, 500, 600, 800, 1000]:
            res = power_h1b.compute_exact_tost_power(m=m, p=0.85, alpha=0.05)
            self.assertLessEqual(res["type_i_error_lower_pct"], 5.0, f"M={m} violated lower alpha")
            self.assertLessEqual(res["type_i_error_upper_pct"], 5.0, f"M={m} violated upper alpha")


if __name__ == "__main__":
    unittest.main()

