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


def _evaluate_acceptance_region(
    m: int,
    p: float,
    k_min: int | None,
    k_max: int | None,
    lower_bound: float,
    upper_bound: float,
    extra: dict,
) -> dict:
    """Computes power at p and exact Type I errors at lower/upper TOST boundaries."""
    if k_min is None or k_max is None or k_min > k_max:
        out = {
            "M": m,
            "p": p,
            "power_pct": 0.0,
            "k_min": None,
            "k_max": None,
            "type_i_error_lower_pct": 0.0,
            "type_i_error_upper_pct": 0.0,
        }
        out.update(extra)
        return out

    power = sum(binomial_pmf(k, m, p) for k in range(k_min, k_max + 1))
    err_lo = sum(binomial_pmf(k, m, lower_bound) for k in range(k_min, k_max + 1))
    err_hi = sum(binomial_pmf(k, m, upper_bound) for k in range(k_min, k_max + 1))
    out = {
        "M": m,
        "p": p,
        "power_pct": round(power * 100.0, 2),
        "k_min": k_min,
        "k_max": k_max,
        "type_i_error_lower_pct": round(err_lo * 100.0, 4),
        "type_i_error_upper_pct": round(err_hi * 100.0, 4),
    }
    out.update(extra)
    return out


def compute_exact_power(m: int, p: float, z: float, lower_bound: float = 0.80, upper_bound: float = 0.90) -> dict:
    """Acceptance region via two-sided Wilson interval containment."""
    k_min = None
    k_max = None
    for k in range(m + 1):
        lo, hi = wilson_interval(k, m, z)
        if lo >= lower_bound and hi <= upper_bound:
            if k_min is None:
                k_min = k
            k_max = k
    return _evaluate_acceptance_region(m, p, k_min, k_max, lower_bound, upper_bound, {"z": round(z, 6)})


def compute_exact_tost_power(
    m: int,
    p: float,
    alpha: float = 0.05,
    lower_bound: float = 0.80,
    upper_bound: float = 0.90,
) -> dict:
    """Exact Binomial TOST (equivalent to Clopper-Pearson 1-2*alpha CI inside [L, U])."""
    k_min = None
    tail_upper = 0.0
    for k in range(m, -1, -1):
        pmf = binomial_pmf(k, m, lower_bound)
        if tail_upper + pmf <= alpha + 1e-12:
            tail_upper += pmf
            k_min = k
        else:
            break

    k_max = None
    tail_lower = 0.0
    for k in range(m + 1):
        pmf = binomial_pmf(k, m, upper_bound)
        if tail_lower + pmf <= alpha + 1e-12:
            tail_lower += pmf
            k_max = k
        else:
            break

    if k_min is None or k_max is None or k_min > k_max:
        k_min, k_max = None, None

    return _evaluate_acceptance_region(
        m,
        p,
        k_min,
        k_max,
        lower_bound,
        upper_bound,
        {"alpha": alpha},
    )


def main():
    m_values = [200, 300, 400, 500, 600, 800, 1000]
    p_values = [0.83, 0.85, 0.87]
    z_95 = 1.959963984540054   # IC 95% de Wilson (conservador)
    z_90 = 1.6448536269514722  # Wilson 90% aproximado (referencia comparativa)
    alpha = 0.05               # TOST Binomial Exato / Clopper-Pearson 90% (criterio primario)

    results = {"wilson_95": {}, "wilson_90": {}, "tost_exact_90": {}}
    print("============================================================================================================")
    print("  Poder Exato de H1b: Inclusao em [80%, 90%] sob Binomial(M, p)")
    print("============================================================================================================")
    print(
        f"{'M':>5} | {'W95(.85)':>8} | {'W90(.85) [TypeI lo/hi]':>24} | "
        f"{'ExTOST(.83)':>11} | {'ExTOST(.85)':>11} | {'ExTOST(.87)':>11} | {'Regiao [TypeI lo/hi]':>26}"
    )
    print("-" * 112)

    for m in m_values:
        results["wilson_95"][str(m)] = {str(p): compute_exact_power(m, p, z_95) for p in p_values}
        results["wilson_90"][str(m)] = {str(p): compute_exact_power(m, p, z_90) for p in p_values}
        results["tost_exact_90"][str(m)] = {
            str(p): compute_exact_tost_power(m, p, alpha=alpha) for p in p_values
        }

        w95_85 = results["wilson_95"][str(m)]["0.85"]["power_pct"]
        w90_entry = results["wilson_90"][str(m)]["0.85"]
        t83 = results["tost_exact_90"][str(m)]["0.83"]["power_pct"]
        t85_entry = results["tost_exact_90"][str(m)]["0.85"]
        t87 = results["tost_exact_90"][str(m)]["0.87"]["power_pct"]
        reg = f"[{t85_entry['k_min']},{t85_entry['k_max']}]" if t85_entry["k_min"] is not None else "vazio"
        print(
            f"{m:5d} | {w95_85:7.2f}% | "
            f"{w90_entry['power_pct']:6.2f}% [{w90_entry['type_i_error_lower_pct']:4.2f}/{w90_entry['type_i_error_upper_pct']:4.2f}%] | "
            f"{t83:10.2f}% | {t85_entry['power_pct']:10.2f}% | {t87:10.2f}% | "
            f"{reg:>10} [{t85_entry['type_i_error_lower_pct']:4.2f}/{t85_entry['type_i_error_upper_pct']:4.2f}%]"
        )

    # Assert key properties from EST-01 / ADR-001 (PK-02.2)
    assert results["wilson_95"]["200"]["0.85"]["power_pct"] == 0.0, "M=200 under Wilson 95% must have 0% power"
    ex500 = results["tost_exact_90"]["500"]["0.85"]
    assert ex500["k_min"] == 416 and ex500["k_max"] == 438, f"Expected [416, 438], got [{ex500['k_min']}, {ex500['k_max']}]"
    assert ex500["power_pct"] >= 80.0, "M=500 under Exact TOST 90% at p=0.85 must achieve >= 80% power"
    assert ex500["type_i_error_lower_pct"] <= 5.0 and ex500["type_i_error_upper_pct"] <= 5.0, "Exact TOST must satisfy alpha <= 5%"

    script_dir = os.path.dirname(os.path.abspath(__file__))
    data_dir = os.path.join(os.path.dirname(script_dir), "data")
    os.makedirs(data_dir, exist_ok=True)
    out_path = os.path.join(data_dir, "power_h1b_results.json")
    with open(out_path, "w", encoding="utf-8") as f:
        json.dump(results, f, indent=2, ensure_ascii=False)
    print(f"\nResultados exportados com sucesso para {out_path}")


if __name__ == "__main__":
    main()

