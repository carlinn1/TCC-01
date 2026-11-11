#!/usr/bin/env python3
"""
Calculo exato do poder estatistico de H1b (Dimensao 2) sob distribuicao Binomial(M, p).
Compara o criterio de inclusao do IC 95% de Wilson em [80%, 90%] contra o teste de
equivalencia TOST (duas caudas unilaterais a alpha = 0,05, equivalente a IC 90% de Wilson
inteiramente contido em [80%, 90%]) para p in {0.83, 0.85, 0.87} (EST-01 / ADR-001).
"""

import json
import math
import os


def wilson_interval(k: int, n: int, z: float) -> tuple[float, float]:
    p_hat = k / n
    z2 = z * z
    denom = 1.0 + z2 / n
    center = (p_hat + z2 / (2.0 * n)) / denom
    half = (z / denom) * math.sqrt((p_hat * (1.0 - p_hat) / n) + (z2 / (4.0 * n * n)))
    return center - half, center + half


def binomial_pmf(k: int, n: int, p: float) -> float:
    if p <= 0.0:
        return 1.0 if k == 0 else 0.0
    if p >= 1.0:
        return 1.0 if k == n else 0.0
    log_pmf = (
        math.lgamma(n + 1)
        - math.lgamma(k + 1)
        - math.lgamma(n - k + 1)
        + k * math.log(p)
        + (n - k) * math.log(1.0 - p)
    )
    return math.exp(log_pmf)


def compute_exact_power(m: int, p: float, z: float, lower_bound: float = 0.80, upper_bound: float = 0.90) -> dict:
    power = 0.0
    k_min = None
    k_max = None
    for k in range(m + 1):
        lo, hi = wilson_interval(k, m, z)
        if lo >= lower_bound and hi <= upper_bound:
            prob = binomial_pmf(k, m, p)
            power += prob
            if k_min is None:
                k_min = k
            k_max = k
    return {
        "M": m,
        "p": p,
        "z": round(z, 6),
        "power_pct": round(power * 100.0, 2),
        "k_min": k_min,
        "k_max": k_max,
    }


def main():
    m_values = [200, 300, 400, 500, 600, 800, 1000]
    p_values = [0.83, 0.85, 0.87]
    z_95 = 1.959963984540054   # IC 95% de Wilson (two-sided alpha=0.05)
    z_90 = 1.6448536269514722  # TOST alpha=0.05 (IC 90% de Wilson)

    results = {"wilson_95": {}, "tost_90": {}}
    print("=========================================================================")
    print("  Poder Exato de H1b: Inclusao em [80%, 90%] sob Binomial(M, p)")
    print("=========================================================================")
    print(f"{'M':>6} | {'W95 (p=0.85)':>14} | {'TOST90 (p=0.83)':>16} | {'TOST90 (p=0.85)':>16} | {'TOST90 (p=0.87)':>16}")
    print("-" * 78)

    for m in m_values:
        w95_85 = compute_exact_power(m, 0.85, z_95)
        results["wilson_95"][str(m)] = {
            str(p): compute_exact_power(m, p, z_95) for p in p_values
        }
        results["tost_90"][str(m)] = {
            str(p): compute_exact_power(m, p, z_90) for p in p_values
        }
        t83 = results["tost_90"][str(m)]["0.83"]["power_pct"]
        t85 = results["tost_90"][str(m)]["0.85"]["power_pct"]
        t87 = results["tost_90"][str(m)]["0.87"]["power_pct"]
        print(f"{m:6d} | {w95_85['power_pct']:13.1f}% | {t83:15.1f}% | {t85:15.1f}% | {t87:15.1f}%")

    # Assert key properties from EST-01 / ADR-001
    assert results["wilson_95"]["200"]["0.85"]["power_pct"] == 0.0, "M=200 under Wilson 95% must have 0% power"
    assert results["tost_90"]["500"]["0.85"]["power_pct"] >= 80.0, "M=500 under TOST 90% at p=0.85 must achieve >= 80% power"

    out_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "power_h1b_results.json")
    with open(out_path, "w", encoding="utf-8") as f:
        json.dump(results, f, indent=2, ensure_ascii=False)
    print(f"\nResultados exportados com sucesso para {out_path}")


if __name__ == "__main__":
    main()

