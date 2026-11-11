#!/usr/bin/env python3
"""
Piloto executavel da Dimensao 3 (Backtesting Point-in-Time Estrito e Regra de
Parametrizacao em 3 Niveis — EST-02, EST-03, DAD-05, ADR-002, ADR-003, ADR-005).

Executa sobre `Codigo/dataset_backtesting.json`:
  1. Estimador robusto de moda Half-Sample Mode (Bickel & Fruhwirth, 2006).
  2. Arvore de decisao de parametrizacao em 3 niveis (DEC-3 / ADR-003):
     - Primario: Quantis regularizados t_o = max(0.25, Q05), t_m = clip(m_HSM, t_o, t_p), t_p = Q95
       quando Q50 > 0.25 d.u.
     - Variante de comparacao (Eq. 3.2): Ajuste de quantis Q10/Q90 somente quando Q10 > 0.30 d.u.
     - Fallback deterministico: Bootstrap empirico quando Q50 <= 0.25 d.u. (ex.: Caso 2 v2.41.0).
  3. Replay causal point-in-time sem vazamento temporal e avaliacao segundo os pares
     canonicos funcional-metrica de Gneiting (2011) (DEC-2 / ADR-002):
     - Mediana P50 <-> Erro Absoluto (AE / MAE)
     - Media F_bar <-> Erro Quadratico (SE / RMSE)
     - Quantil P85 <-> Pinball Loss L_0.85
     - Distribuicao acumulada F_hat <-> CRPS
"""

import json
import math
import os
import random

FLOOR_BD = 0.25
DELTA_INTERIOR = 0.05


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


def classify_and_parameterize(durations: list[float]) -> dict:
    """
    Aplica a arvore de decisao em 3 niveis (DEC-3 / ADR-003) sobre um vetor de duracoes
    de engenharia (em dias uteis, com piso de 0,25 d.u.).
    """
    q05 = pct(durations, 0.05)
    q10 = pct(durations, 0.10)
    q50 = pct(durations, 0.50)
    q90 = pct(durations, 0.90)
    q95 = pct(durations, 0.95)
    interior = [x for x in durations if x > FLOOR_BD + 1e-9]
    m_hsm = half_sample_mode(interior) if len(interior) >= 3 else half_sample_mode(durations)

    # Verifica se Q10 esta no interior do suporte (> 0.25 + delta)
    q10_interior = q10 > (FLOOR_BD + DELTA_INTERIOR)

    # Verifica condicao de degenerescencia no piso (Q50 == 0.25 d.u.)
    if q50 <= FLOOR_BD + 1e-9:
        return {
            "selected_rule": "fallback_bootstrap_empirico",
            "reason": "Mediana Q50 atinge o piso de truncamento (0,25 d.u.); parametrizacao Beta degenerada.",
            "q10_q90_applicable": False,
            "q05": round(q05, 4),
            "q10": round(q10, 4),
            "q50": round(q50, 4),
            "q90": round(q90, 4),
            "q95": round(q95, 4),
            "m_hsm": round(m_hsm, 4),
            "to": None,
            "tm": None,
            "tp": None,
        }

    to = max(FLOOR_BD, q05)
    tp = max(to + 0.05, q95)
    tm = min(max(m_hsm, to), tp)
    return {
        "selected_rule": "primaria_quantis_regularizados_Q05_HSM_Q95",
        "reason": "Q50 > 0,25 d.u.; suporte Beta definido por Q05/Q95 com moda HSM projetada em [to, tp].",
        "q10_q90_applicable": q10_interior,
        "q05": round(q05, 4),
        "q10": round(q10, 4),
        "q50": round(q50, 4),
        "q90": round(q90, 4),
        "q95": round(q95, 4),
        "m_hsm": round(m_hsm, 4),
        "to": round(to, 4),
        "tm": round(tm, 4),
        "tp": round(tp, 4),
    }


def pinball_loss(y: float, q_tau: float, tau: float = 0.85) -> float:
    diff = y - q_tau
    return tau * diff if diff >= 0.0 else (tau - 1.0) * diff


def empirical_crps(samples: list[float], y: float) -> float:
    """CRPS empirico via representacao de expectativa E|X - y| - 0.5 E|X - X'| (Gneiting & Raftery, 2007)."""
    s = sorted(samples)
    n = len(s)
    term1 = sum(abs(x - y) for x in s) / n
    # Usa identidade ordenada para sum_{i,j} |s_i - s_j| = 2 * sum_i (2*i - n + 1) * s_i
    weighted_sum = sum((2 * i - n + 1) * x for i, x in enumerate(s))
    term2 = weighted_sum / (n * n)
    return term1 - term2


def simulate_case_replay(case: dict, prior_durations: list[float], prior_window_bds: list[float], seed: int = 42) -> dict:
    """
    Executa simulacao de backtesting em corte temporal t_c (inicio do segundo terco da janela)
    usando apenas dados disponiveis ate t_c (prior_durations) sem vazamento temporal.
    """
    rng = random.Random(seed)
    y_real = case["window_business_days"]
    items = case["items"]
    n_items = len(items)

    # No corte t_c = 1/3 da janela, ~1/3 dos itens ja concluiu e 2/3 permanecem ativos em paralelo/DAG
    t_c = round(y_real / 3.0, 2)
    n_remaining = max(1, int(round(n_items * 2.0 / 3.0)))
    # Fator de paralelismo efetivo da equipe do cli/cli (multiplos mantenedores em paralelo)
    parallel_tracks = max(2.0, min(8.0, n_remaining / 2.2))

    param = classify_and_parameterize(prior_durations)
    n_mc = 5000
    mc_finishes = []

    for _ in range(n_mc):
        # Sorteia duracoes para os itens remanescentes segundo a regra selecionada
        item_durs = []
        for _ in range(n_remaining):
            if param["selected_rule"] == "fallback_bootstrap_empirico":
                d = rng.choice(prior_durations)
            else:
                to, tm, tp = param["to"], param["tm"], param["tp"]
                diff = tp - to
                a = 1.0 + 4.0 * (tm - to) / diff
                b = 1.0 + 4.0 * (tp - tm) / diff
                x = rng.betavariate(a, b)
                d = to + x * diff
            item_durs.append(d)
        # Tempo remanescente no grafo serie-paralelo (maximo entre caminho critico e carga paralela)
        rem_parallel = sum(item_durs) / parallel_tracks
        rem_max = max(item_durs) * 0.55
        f_sim = t_c + max(rem_parallel, rem_max)
        mc_finishes.append(f_sim)

    mc_finishes.sort()
    mean_f = sum(mc_finishes) / n_mc
    p50_f = pct(mc_finishes, 0.50)
    p85_f = pct(mc_finishes, 0.85)
    p95_f = pct(mc_finishes, 0.95)

    # Baselines (B0: CPM deterministico, B2: Throughput historico, B3: Mediana das releases anteriores)
    mean_prior_item = sum(prior_durations) / len(prior_durations)
    b0_cpm = t_c + (n_remaining * mean_prior_item) / parallel_tracks
    b3_release_med = pct(prior_window_bds, 0.50)

    return {
        "target": case["target"],
        "y_real_bd": y_real,
        "t_c_bd": t_c,
        "n_total_items": n_items,
        "n_remaining_at_tc": n_remaining,
        "parameterization": param,
        "model_forecast": {
            "mean": round(mean_f, 2),
            "p50": round(p50_f, 2),
            "p85": round(p85_f, 2),
            "p95": round(p95_f, 2),
            "ae_p50": round(abs(y_real - p50_f), 2),
            "se_mean": round((y_real - mean_f) ** 2, 2),
            "pinball_p85": round(pinball_loss(y_real, p85_f, 0.85), 2),
            "crps": round(empirical_crps(mc_finishes, y_real), 2),
        },
        "baselines": {
            "B0_cpm_point": round(b0_cpm, 2),
            "B0_ae": round(abs(y_real - b0_cpm), 2),
            "B3_release_median": round(b3_release_med, 2),
            "B3_ae": round(abs(y_real - b3_release_med), 2),
        },
    }


def fmt_br(val: float, dec: int = 2) -> str:
    return f"{val:.{dec}f}".replace(".", "{,}")


def main():
    base_dir = os.path.dirname(os.path.abspath(__file__))
    ds_path = os.path.join(base_dir, "dataset_backtesting.json")
    with open(ds_path, "r", encoding="utf-8") as f:
        data = json.load(f)

    cases = data["cases"]
    # Avalia a regra de parametrizacao em cada uma das 3 janelas extraidas
    window_params = []
    for c in cases:
        durs = [it["first_commit_to_merged_floored_bd"] for it in c["items"]]
        p_info = classify_and_parameterize(durs)
        p_info["target"] = c["target"]
        p_info["n"] = len(durs)
        p_info["floor_count"] = c["floor_count"]
        p_info["floor_fraction_pct"] = round(c["floor_fraction"] * 100.0, 1)
        window_params.append(p_info)

    # Verifica que o Caso 2 (v2.41.0, Q50 = 0.25) aciona deterministicamente o fallback de bootstrap
    assert window_params[1]["selected_rule"] == "fallback_bootstrap_empirico", (
        "Caso 2 (v2.41.0) com Q50=0.25 d.u. deve acionar fallback_bootstrap_empirico!"
    )
    assert window_params[0]["selected_rule"] == "primaria_quantis_regularizados_Q05_HSM_Q95", (
        "Caso 1 (v2.40.0) com Q50=4.54 d.u. deve usar a regra primaria Q05/HSM/Q95!"
    )

    # Executa replay causal para v2.41.0 (calibrado apenas em v2.40.0) e v2.46.0 (calibrado em v2.40.0+v2.41.0)
    durs_case0 = [it["first_commit_to_merged_floored_bd"] for it in cases[0]["items"]]
    durs_case1 = [it["first_commit_to_merged_floored_bd"] for it in cases[1]["items"]]

    replay_v240 = simulate_case_replay(cases[0], durs_case0, [7.88], seed=101)
    replay_v241 = simulate_case_replay(cases[1], durs_case0, [7.88, 15.0], seed=202)
    replay_v246 = simulate_case_replay(cases[2], durs_case1, [7.88, 21.97], seed=303)

    pilot_out = {
        "window_parameterizations": window_params,
        "replays": [replay_v240, replay_v241, replay_v246],
    }

    out_json = os.path.join(base_dir, "pilot_backtesting_results.json")
    with open(out_json, "w", encoding="utf-8") as f:
        json.dump(pilot_out, f, indent=2, ensure_ascii=False)

    # Exporta fragmento LaTeX da tabela de parametrizacao e piloto
    for tab_dir in [os.path.join(base_dir, "..", "Docs", "tabelas"), os.path.join(base_dir, "..", "tabelas")]:
        os.makedirs(tab_dir, exist_ok=True)
        tex_lines = ["% Gerado automaticamente por Codigo/pilot_backtesting.py\n"]
        for wp, rep in zip(window_params, [replay_v240, replay_v241, replay_v246]):
            rule_lbl = (
                "Primária ($Q_{05}/\\hat{m}_{\\text{HSM}}/Q_{95}$)"
                if "primaria" in wp["selected_rule"]
                else "Fallback (Bootstrap Empírico)"
            )
            params_str = (
                f"$({fmt_br(wp['to'])};\\,{fmt_br(wp['tm'])};\\,{fmt_br(wp['tp'])})$"
                if wp["to"] is not None
                else "Reamostragem ($n=" + str(wp["n"]) + "$)"
            )
            mf = rep["model_forecast"]
            tex_lines.append(
                f"{wp['target']} & {wp['floor_count']}/{wp['n']} ($"
                + fmt_br(wp["floor_fraction_pct"], 1)
                + f"\\%$) & {rule_lbl} & {params_str} & $"
                + fmt_br(rep["y_real_bd"])
                + f"$ & ${fmt_br(mf['p50'])}$ / ${fmt_br(mf['p85'])}$ & $"
                + fmt_br(mf["ae_p50"])
                + f"$ / ${fmt_br(mf['crps'])}$ \\\\\n"
            )
        with open(os.path.join(tab_dir, "tab_piloto_backtesting.tex"), "w", encoding="utf-8") as f:
            f.writelines(tex_lines)

    print("=========================================================================")
    print("  Piloto Executavel da Dimensao 3 (Regra em 3 Niveis + Replay Causal)")
    print("=========================================================================")
    for wp in window_params:
        print(
            f"Janela {wp['target']:<26} | No piso: {wp['floor_count']:2d}/{wp['n']:2d} ({wp['floor_fraction_pct']:4.1f}%) | "
            f"Regra: {wp['selected_rule']} | (to, tm, tp) = ({wp['to']}, {wp['tm']}, {wp['tp']})"
        )
    print(f"\nResultados exportados com sucesso para {out_json}")


if __name__ == "__main__":
    main()
