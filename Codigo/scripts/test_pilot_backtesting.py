import sys
import unittest
from pathlib import Path

SCRIPTS_DIR = Path(__file__).resolve().parent
if str(SCRIPTS_DIR) not in sys.path:
    sys.path.insert(0, str(SCRIPTS_DIR))

import pilot_backtesting as pb  # noqa: E402


class TestPilotBacktesting(unittest.TestCase):
    def test_no_future_leakage_replay_function_present(self):
        self.assertFalse(
            hasattr(pb, "simulate_case_replay"),
            "simulate_case_replay with y_real temporal leakage must be removed per PK-01",
        )

    def test_classify_three_level_decision_tree(self):
        # Level 3 fallback when median hits floor (0.25)
        floor_durs = [0.25] * 6 + [0.5, 1.0, 2.0]
        res_fb = pb.classify_and_parameterize(floor_durs)
        self.assertEqual(res_fb["selected_rule"], "fallback_bootstrap_empirico")
        self.assertIsNone(res_fb["to"])

        # Level 1 primary Q05/HSM/Q95 when Q50 > 0.25 but Q10 <= 0.30
        mixed_durs = [0.25, 0.25, 0.25, 1.0, 2.0, 3.0, 3.2, 3.4, 5.0, 8.0, 10.0, 12.0, 15.0, 16.0, 20.0]
        res_l1 = pb.classify_and_parameterize(mixed_durs)
        self.assertEqual(res_l1["selected_rule"], "primaria_quantis_regularizados_Q05_HSM_Q95")
        self.assertFalse(res_l1["q10_q90_applicable"])
        self.assertAlmostEqual(res_l1["to"], 0.25, places=4)
        self.assertTrue(res_l1["passes_min_sample_n15"])

        # Interior sample where Q10 > 0.30 enables Q10/Q90 check
        interior_durs = [0.5 + i * 0.2 for i in range(16)]
        res_int = pb.classify_and_parameterize(interior_durs)
        self.assertEqual(res_int["selected_rule"], "primaria_quantis_regularizados_Q05_HSM_Q95")
        self.assertTrue(res_int["q10_q90_applicable"])
        self.assertIsNotNone(res_int["pert_expected_mean"])


if __name__ == "__main__":
    unittest.main()

