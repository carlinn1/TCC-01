#!/usr/bin/env python3
"""
Demonstracao exploratoria de parametrizacao estocastica da Dimensao 3
(Arvore em 2 Niveis — EST-02, EST-03, DAD-05, ADR-002, ADR-003, ADR-005, C1).

Nota metodologica (PK-01 / C1 / B2): No TCC 1 este script executa EXCLUSIVAMENTE a calibracao
exploratoria da regra em 2 niveis (Nível 1: Q05/m_HSM/Q95; Fallback: Bootstrap empirico quando Q50 <= 0.25)
sobre as distribuicoes empiricas extraidas de `cli/cli`, removendo o antigo Nivel 2 (casamento Q10/Q90)
que nunca disparava por coincidir Q10 com o piso operacional. Sem vazamento temporal (`y_real`), e o
backtesting descritivo com congelamento causal fica reservado a E09 (TCC 2).
"""

import json
import os

FLOOR_BD = 0.25


def pct(xs: list[float], q: float) -> float:
    s = sorted(xs)
    k = (len(s) - 1) * q
    f = int(k)
    c = min(f + 1, len(s) - 1)
    return s[f] + (s[c] - s[f]) * (k - f)


def half_sample_mode(xs: list[float]) -> float:
    """Estimador Half-Sample Mode (Bickel & Fruhwirth, 2006) recursivo O(n log n)."""
    s = sorted(xs)
    while len(s) > 3:
        n = len(s)
        half = (n + 1) // 2
        best_width = float("inf")
        best_idx = 0
        for i in range(n - half + 1):
            w = s[i + half - 1] - s[i]
            if w < best_width:
                best_width = w
                best_idx = i
        s = s[best_idx : best_idx + half]
    if len(s) == 1:
        return s[0]
    if len(s) == 2:
        return 0.5 * (s[0] + s[1])
    # len == 3
    if (s[1] - s[0]) < (s[2] - s[1]):
        return 0.5 * (s[0] + s[1])
    if (s[2] - s[1]) < (s[1] - s[0]):
        return 0.5 * (s[1] + s[2])
    return s[1]


MIN_SAMPLE_SIZE_N = 15


def classify_and_parameterize(durations: list[float]) -> dict:
    """
    Aplica a arvore de decisao em 2 niveis (C1 / B2 / ADR-003) sobre um vetor de duracoes
    de engenharia (em dias uteis, com piso de 0,25 d.u.), sem uso de dados futuros.
    """
    if not durations:
        raise ValueError("durations nao pode ser vazio")
    q05 = pct(durations, 0.05)
    q10 = pct(durations, 0.10)
    q25 = pct(durations, 0.25)
    q50 = pct(durations, 0.50)
    q75 = pct(durations, 0.75)
    q90 = pct(durations, 0.90)
    q95 = pct(durations, 0.95)
    sample_mean = sum(durations) / len(durations)
    interior = [x for x in durations if x > FLOOR_BD + 1e-9]
    m_hsm = half_sample_mode(interior) if len(interior) >= 3 else half_sample_mode(durations)

    passes_min_n = len(durations) >= MIN_SAMPLE_SIZE_N

    # Verifica condicao de degenerescencia no piso (Q50 == 0.25 d.u.)
    if q50 <= FLOOR_BD + 1e-9:
        return {
            "selected_rule": "fallback_bootstrap_empirico",
            "reason": "Mediana Q50 atinge o piso de truncamento (0,25 d.u.); parametrizacao Beta degenerada.",
            "passes_min_sample_n15": passes_min_n,
            "q05": round(q05, 4),
            "q10": round(q10, 4),
            "q50": round(q50, 4),
            "q90": round(q90, 4),
            "q95": round(q95, 4),
            "sample_mean": round(sample_mean, 4),
            "sample_iqr": round(q75 - q25, 4),
            "m_hsm": round(m_hsm, 4),
            "to": None,
            "tm": None,
            "tp": None,
            "pert_expected_mean": None,
        }

    to = max(FLOOR_BD, q05)
    tp = max(to + 0.05, q95)
    tm = min(max(m_hsm, to), tp)
    pert_mean = (to + 4.0 * tm + tp) / 6.0
    return {
        "selected_rule": "primaria_quantis_regularizados_Q05_HSM_Q95",
        "reason": "Q50 > 0,25 d.u.; suporte Beta definido por Q05/Q95 com moda HSM projetada em [to, tp].",
        "passes_min_sample_n15": passes_min_n,
        "q05": round(q05, 4),
        "q10": round(q10, 4),
        "q50": round(q50, 4),
        "q90": round(q90, 4),
        "q95": round(q95, 4),
        "sample_mean": round(sample_mean, 4),
        "sample_iqr": round(q75 - q25, 4),
        "m_hsm": round(m_hsm, 4),
        "to": round(to, 4),
        "tm": round(tm, 4),
        "tp": round(tp, 4),
        "pert_expected_mean": round(pert_mean, 4),
    }


def fmt_br(val: float, dec: int = 2) -> str:
    return f"{val:.{dec}f}".replace(".", "{,}")


def main():
    script_dir = os.path.dirname(os.path.abspath(__file__))
    codigo_dir = os.path.dirname(script_dir)
    root_dir = os.path.dirname(codigo_dir)
    data_dir = os.path.join(codigo_dir, "data")
    tab_dir = os.path.join(root_dir, "Docs", "monografia", "tabelas")

    ds_path = os.path.join(data_dir, "dataset_backtesting.json")
    with open(ds_path, "r", encoding="utf-8") as f:
        data = json.load(f)

    cases = data["cases"]
    window_params = []
    for c in cases:
        durs = [it["first_commit_to_merged_floored_bd"] for it in c["items"]]
        p_info = classify_and_parameterize(durs)
        p_info["target"] = c["target"]
        p_info["n"] = len(durs)
        p_info["floor_count"] = c["floor_count"]
        p_info["floor_fraction_pct"] = round(c["floor_fraction"] * 100.0, 1)
        p_info["window_business_days"] = c["window_business_days"]
        window_params.append(p_info)

    # Invariante estrutural: janelas com mediana == 0.25 acionam fallback empirico deterministicamente
    for wp in window_params:
        if wp["q50"] <= FLOOR_BD + 1e-9:
            assert wp["selected_rule"] == "fallback_bootstrap_empirico", (
                f"{wp['target']} com Q50={wp['q50']} deveria acionar fallback_bootstrap_empirico"
            )
            assert wp["to"] is None
        else:
            assert wp["selected_rule"] == "primaria_quantis_regularizados_Q05_HSM_Q95", (
                f"{wp['target']} com Q50={wp['q50']} deveria acionar regra primaria Q05/HSM/Q95"
            )
            assert wp["to"] is not None and wp["to"] >= FLOOR_BD - 1e-9

    pilot_out = {
        "classification": "exploratory_stochastic_parameterization_only",
        "temporal_leakage": False,
        "min_sample_size_n": MIN_SAMPLE_SIZE_N,
        "note": (
            "Demonstracao exploratoria de parametrizacao da arvore em 2 niveis no TCC 1 (C1: Nivel 2 Q10/Q90 removido); "
            "backtesting descritivo com congelamento causal fica reservado a E09 do TCC 2."
        ),
        "window_parameterizations": window_params,
    }

    os.makedirs(data_dir, exist_ok=True)
    out_json = os.path.join(data_dir, "pilot_backtesting_results.json")
    with open(out_json, "w", encoding="utf-8") as f:
        json.dump(pilot_out, f, indent=2, ensure_ascii=False)

    # Exporta fragmento LaTeX da tabela exploratoria de parametrizacao em Docs/monografia/tabelas/
    os.makedirs(tab_dir, exist_ok=True)
    tex_lines = [
        "% Gerado automaticamente por Codigo/scripts/pilot_backtesting.py (Estudo Exploratorio de Parametrizacao)\n",
        "\\providecommand{\\tabRowsPilotoBacktesting}{%\n",
    ]
    for idx, wp in enumerate(window_params):
        rule_lbl = (
            "Primária ($Q_{05}/\\hat{m}_{\\mathrm{HSM}}/Q_{95}$)"
            if "primaria" in wp["selected_rule"]
            else "Fallback (Bootstrap Empírico)"
        )
        params_str = (
            f"$({fmt_br(wp['to'])};\\,{fmt_br(wp['tm'])};\\,{fmt_br(wp['tp'])})$"
            if wp["to"] is not None
            else "Reamostragem ($n=" + str(wp["n"]) + "$)"
        )
        n_flag = "$\\ge 15$" if wp["passes_min_sample_n15"] else "$< 15$ (piloto)"
        ed_str = (
            f"${fmt_br(wp['pert_expected_mean'])}$ (PERT)"
            if wp["pert_expected_mean"] is not None
            else f"${fmt_br(wp['sample_mean'])}$ (emp.)"
        )
        row_end = "\\\\%\n" if idx == len(window_params) - 1 else "\\\\\n"
        tex_lines.append(
            f"{wp['target']} & {wp['floor_count']}/{wp['n']} ($"
            + fmt_br(wp["floor_fraction_pct"], 1)
            + f"\\%$) & {rule_lbl} & {params_str} & ${fmt_br(wp['q50'])}$ & {ed_str} & {n_flag} {row_end}"
        )
    tex_lines.append("}\n")
    with open(os.path.join(tab_dir, "tab_piloto_backtesting.tex"), "w", encoding="utf-8") as f:
        f.writelines(tex_lines)

    print("=========================================================================")
    print("  Estudo Exploratorio de Parametrizacao (Arvore em 2 Niveis, Sem Replay)")
    print("=========================================================================")
    for wp in window_params:
        ed_val = wp["pert_expected_mean"] if wp["pert_expected_mean"] is not None else wp["sample_mean"]
        print(
            f"Janela {wp['target']:<26} | n={wp['n']:2d} (n>=15: {wp['passes_min_sample_n15']}) | "
            f"Piso: {wp['floor_count']:2d}/{wp['n']:2d} ({wp['floor_fraction_pct']:4.1f}%) | "
            f"Regra: {wp['selected_rule']} | E[d]={ed_val:.2f}"
        )
    print(f"\nResultados exportados com sucesso para {out_json}")


if __name__ == "__main__":
    main()
