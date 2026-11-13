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

    def test_classify_two_level_decision_tree(self):
        # Fallback when median hits floor (0.25)
        floor_durs = [0.25] * 6 + [0.5, 1.0, 2.0]
        res_fb = pb.classify_and_parameterize(floor_durs)
        self.assertEqual(res_fb["selected_rule"], "fallback_bootstrap_empirico")
        self.assertIsNone(res_fb["to"])

        # Primary rule with mass at the floor
        mixed_durs = [0.25, 0.25, 0.25, 1.0, 2.0, 3.0, 3.2, 3.4, 5.0, 8.0, 10.0, 12.0, 15.0, 16.0, 20.0]
        res_l1 = pb.classify_and_parameterize(mixed_durs)
        self.assertEqual(res_l1["selected_rule"], "primaria_quantis_regularizados_Q05_HSM_Q95")
        self.assertNotIn("q10_q90_applicable", res_l1)
        self.assertAlmostEqual(res_l1["to"], 0.25, places=4)
        self.assertTrue(res_l1["passes_min_sample_n15"])

        # Primary rule without mass at the floor
        interior_durs = [0.5 + i * 0.2 for i in range(16)]
        res_int = pb.classify_and_parameterize(interior_durs)
        self.assertEqual(res_int["selected_rule"], "primaria_quantis_regularizados_Q05_HSM_Q95")
        self.assertNotIn("q10_q90_applicable", res_int)
        self.assertIsNotNone(res_int["pert_expected_mean"])

    def test_constant_and_low_dispersion_use_bootstrap(self):
        for durations in ([1.0]*15, [1.0, 1.05]*15, [1.0, 1.0555555555555556]):
            with self.subTest(durations=durations):
                result = pb.classify_and_parameterize(durations)
                self.assertEqual(result["selected_rule"], "fallback_bootstrap_empirico")
                self.assertIsNone(result["tp"])
        result = pb.classify_and_parameterize([1.0, 1.06]*15)
        self.assertEqual(result["selected_rule"], "primaria_quantis_regularizados_Q05_HSM_Q95")
        self.assertAlmostEqual(result["tp"], result["q95"])

    def test_hsm_uses_interior_when_at_least_three_observations(self):
        for durations in ([0.25]*4 + [1, 2, 3, 4, 5, 6], [0.25, 1, 2]):
            interior = [x for x in durations if x > pb.FLOOR_BD + 1e-9]
            expected = pb.half_sample_mode(interior if len(interior) >= 3 else durations)
            self.assertAlmostEqual(pb.classify_and_parameterize(durations)["m_hsm"], expected, places=4)


if __name__ == "__main__":
    unittest.main()

