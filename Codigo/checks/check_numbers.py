#!/usr/bin/env python3
"""
Verificador automatizado de consistencia numerica (BASE-02, FOR-01, REL-02).
Cruza os valores exportados em:
  - Codigo/data/poc_results.json
  - Codigo/data/dataset_backtesting.json
  - Codigo/data/power_h1b_results.json
  - Codigo/data/pilot_backtesting_results.json
contra:
  - Docs/monografia/resultados_poc.tex
  - Docs/monografia/tabelas/*.tex
  - Docs/monografia/Main.tex
  - README.md
Retorna codigo de saida != 0 se encontrar qualquer divergencia numerica.
"""
import json
import os
import re
import sys


def fmt_br(val: float, decimals: int = 2) -> str:
    return f"{val:.{decimals}f}".replace(".", "{,}")


def fmt_br_plain(val: float, decimals: int = 2) -> str:
    return f"{val:.{decimals}f}".replace(".", ",")


def fmt_en(val: float, decimals: int = 2) -> str:
    return f"{val:.{decimals}f}"


def main() -> None:
    checks_dir = os.path.dirname(os.path.abspath(__file__))
    code_dir = os.path.dirname(checks_dir)
    root_dir = os.path.dirname(code_dir)
    data_dir = os.path.join(code_dir, "data")
    docs_dir = os.path.join(root_dir, "Docs", "monografia")

    with open(os.path.join(data_dir, "poc_results.json"), "r", encoding="utf-8") as f:
        poc = json.load(f)
    with open(os.path.join(data_dir, "dataset_backtesting.json"), "r", encoding="utf-8") as f:
        ds = json.load(f)
    with open(os.path.join(data_dir, "power_h1b_results.json"), "r", encoding="utf-8") as f:
        pwr = json.load(f)
    with open(os.path.join(data_dir, "pilot_backtesting_results.json"), "r", encoding="utf-8") as f:
        pilot = json.load(f)

    with open(os.path.join(docs_dir, "resultados_poc.tex"), "r", encoding="utf-8") as f:
        macros_tex = f.read()
    with open(os.path.join(docs_dir, "Main.tex"), "r", encoding="utf-8") as f:
        main_tex = f.read()
    with open(os.path.join(root_dir, "README.md"), "r", encoding="utf-8") as f:
        readme_md = f.read()

    errors = []

    # 1. Verifica sincronia de macros em Docs/monografia/resultados_poc.tex com Codigo/data/poc_results.json
    scen_a = poc["scenario_a"]
    scen_b = poc["scenario_b"]
    scen_c = poc["scenario_c"]
    scen_cp = poc["scenario_c_prime"]
    crn = poc["crn_diff_c_cprime"]
    tcc2 = poc["tcc2_audit"]
    b_res = poc["benchmarks"]["results"]
    speedup2w = b_res[1]["median_ms"] / b_res[2]["median_ms"] if b_res[2]["median_ms"] > 0 else 0.0
    speedup4w = b_res[1]["median_ms"] / b_res[3]["median_ms"] if b_res[3]["median_ms"] > 0 else 0.0
    eff4w = (speedup4w / 4.0) * 100.0
    speedup8w = b_res[1]["median_ms"] / b_res[4]["median_ms"] if b_res[4]["median_ms"] > 0 else 0.0
    margin_rnf = 500.0 / b_res[4]["median_ms"] if b_res[4]["median_ms"] > 0 else 0.0

    expected_macros = {
        "pocACPM": fmt_br(scen_a["cpm_deterministic"], 2),
        "pocAMean": fmt_br(scen_a["mean"], 2),
        "pocAPcinquenta": fmt_br(scen_a["p50"], 2),
        "pocAPoitentaeCinco": fmt_br(scen_a["p85"], 2),
        "pocAPnoventaeCinco": fmt_br(scen_a["p95"], 2),
        "pocAProbMeta": fmt_br(scen_a["prob_target"], 1),
        "pocBMean": fmt_br(scen_b["mean"], 2),
        "pocBPcinquenta": fmt_br(scen_b["p50"], 2),
        "pocBPoitentaeCinco": fmt_br(scen_b["p85"], 2),
        "pocBPnoventaeCinco": fmt_br(scen_b["p95"], 2),
        "pocBICTVTres": fmt_br(scen_b["criticality_index"]["T3"], 1),
        "pocCMean": fmt_br(scen_c["mean"], 2),
        "pocCPoitentaeCinco": fmt_br(scen_c["p85"], 2),
        "pocCICTVTres": fmt_br(scen_c["criticality_index"]["T3"], 1),
        "pocCPrimeMean": fmt_br(scen_cp["mean"], 2),
        "pocCPrimePoitentaeCinco": fmt_br(scen_cp["p85"], 2),
        "pocCPrimeICTVTres": fmt_br(scen_cp["criticality_index"]["T3"], 1),
        "crnDeltaMean": fmt_br(crn["delta_mean"], 2),
        "crnDeltaMeanSE": fmt_br(crn["delta_mean_se"], 4),
        "crnIndepSE": fmt_br(crn["independent_se"], 4),
        "crnVarRedRatio": fmt_br(crn["variance_reduction_ratio"], 1),
        "tccDoisBaseCPM": fmt_br(tcc2["base"]["cpm_deterministic"], 2),
        "tccDoisRiscosPcinquenta": fmt_br(tcc2["with_risks"]["p50"], 2),
        "tccDoisRiscosPcinquentaEN": fmt_en(tcc2["with_risks"]["p50"], 2),
        "tccDoisRiscosPoitentaeCinco": fmt_br(tcc2["with_risks"]["p85"], 2),
        "tccDoisRiscosPoitentaeCincoEN": fmt_en(tcc2["with_risks"]["p85"], 2),
        "tccDoisRiscosProbVinte": fmt_br(tcc2["with_risks"]["prob_target"], 1),
        "tccDoisRiscosProbVinteEN": fmt_en(tcc2["with_risks"]["prob_target"], 1),
        "tccDoisStressPoitentaeCinco": fmt_br(tcc2["stress_25"]["p85"], 2),
        "tccDoisStressProbVinte": fmt_br(tcc2["stress_25"]["prob_target"], 1),
        "tccDoisQuarentaCPM": fmt_br(tcc2["stress_40"]["cpm_deterministic"], 2),
        "tccDoisQuarentaMean": fmt_br(tcc2["stress_40"]["mean"], 2),
        "tccDoisQuarentaPcinquenta": fmt_br(tcc2["stress_40"]["p50"], 2),
        "tccDoisQuarentaPoitentaeCinco": fmt_br(tcc2["stress_40"]["p85"], 2),
        "tccDoisQuarentaPoitentaeCincoEN": fmt_en(tcc2["stress_40"]["p85"], 2),
        "tccDoisQuarentaPnoventaeCinco": fmt_br(tcc2["stress_40"]["p95"], 2),
        "tccDoisQuarentaProbVinte": fmt_br(tcc2["stress_40"]["prob_target"], 1),
        "tccDoisQuarentaProbVinteEN": fmt_en(tcc2["stress_40"]["prob_target"], 1),
        "tccDoisTFEUm": fmt_br(tcc2["task_cpm"]["E01"]["tf"], 2),
        "tccDoisTFEQuatro": fmt_br(tcc2["task_cpm"]["E04"]["tf"], 2),
        "tccDoisTFESeis": fmt_br(tcc2["task_cpm"]["E06"]["tf"], 2),
        "tccDoisTFESete": fmt_br(tcc2["task_cpm"]["E07"]["tf"], 2),
        "tccDoisTFEOnze": fmt_br(tcc2["task_cpm"]["E11"]["tf"], 2),
        "tccDoisRiscosICESeis": fmt_br(tcc2["with_risks"]["criticality_index"]["E06"], 1),
        "tccDoisRiscosICESete": fmt_br(tcc2["with_risks"]["criticality_index"]["E07"], 1),
        "tccDoisRiscosICEOito": fmt_br(tcc2["with_risks"]["criticality_index"]["E08"], 1),
        "tccDoisRiscosICENove": fmt_br(tcc2["with_risks"]["criticality_index"]["E09"], 1),
        "tccDoisRiscosICEOnze": fmt_br(tcc2["with_risks"]["criticality_index"]["E11"], 1),
        "benchBaseMedMs": fmt_br(b_res[0]["median_ms"], 2),
        "benchBaseMedMsEN": fmt_en(b_res[0]["median_ms"], 2),
        "benchSynthWOneMedMs": fmt_br(b_res[1]["median_ms"], 1),
        "benchSynthWOneMedMsEN": fmt_en(b_res[1]["median_ms"], 1),
        "benchSynthWTwoMedMs": fmt_br(b_res[2]["median_ms"], 1),
        "benchSynthWFourMedMs": fmt_br(b_res[3]["median_ms"], 1),
        "benchSynthWFourMedMsEN": fmt_en(b_res[3]["median_ms"], 1),
        "benchSynthWEightMedMs": fmt_br(b_res[4]["median_ms"], 1),
        "benchSynthWEightMedMsEN": fmt_en(b_res[4]["median_ms"], 1),
        "benchSpeedupDoisW": fmt_br(speedup2w, 2),
        "benchSpeedupQuatroW": fmt_br(speedup4w, 2),
        "benchEffQuatroW": fmt_br(eff4w, 1),
        "benchSpeedupOitoW": fmt_br(speedup8w, 2),
        "benchMarginRNF": fmt_br(margin_rnf, 1),
    }

    for name, val in expected_macros.items():
        pattern = rf"\\providecommand\{{\\{name}\}}\{{{re.escape(val)}\}}"
        if not re.search(pattern, macros_tex):
            errors.append(f"Macro \\{name} em Docs/monografia/resultados_poc.tex nao casa com valor esperado '{val}'")

    # 1b. Verifica uso das macros EN/TCC2/Benchmark em Docs/monografia/Main.tex (Resumo, Abstract, Cap. 5-6)
    required_macros_in_main = [
        "\\benchSynthWOneMedMs{}",
        "\\benchSynthWTwoMedMs{}",
        "\\benchSynthWFourMedMs{}",
        "\\benchSynthWEightMedMs{}",
        "\\tccDoisRiscosPcinquentaEN{}",
        "\\tccDoisRiscosPoitentaeCincoEN{}",
        "\\tccDoisRiscosProbVinteEN{}",
        "\\tccDoisQuarentaPoitentaeCincoEN{}",
        "\\tccDoisQuarentaProbVinteEN{}",
        "\\tccDoisTFEUm{}",
        "\\benchSpeedupDoisW{}",
        "\\benchSpeedupQuatroW{}",
        "\\benchMarginRNF{}",
    ]
    for macro_ref in required_macros_in_main:
        if macro_ref not in main_tex:
            errors.append(f"Macro obrigatoria '{macro_ref}' ausente em Docs/monografia/Main.tex")

    # 2. Verifica se os 7 fragmentos LaTeX em Docs/monografia/tabelas/ existem, nao estao vazios e sao incluidos
    expected_tables = [
        "tab_ks.tex",
        "tab_poc_cenarios.tex",
        "tab_poc_ic.tex",
        "tab_reprodutibilidade.tex",
        "tab_benchmarks.tex",
        "tab_tcc2_mc.tex",
        "tab_piloto_backtesting.tex",
    ]
    for tname in expected_tables:
        p_docs = os.path.join(docs_dir, "tabelas", tname)
        if not os.path.exists(p_docs) or os.path.getsize(p_docs) < 20:
            errors.append(f"Tabela gerada ausente ou vazia: {p_docs}")
            continue
        if f"\\input{{tabelas/{tname}}}" not in main_tex:
            errors.append(f"Docs/monografia/Main.tex nao inclui \\input{{tabelas/{tname}}}")
        with open(p_docs, "r", encoding="utf-8") as tf:
            t_content = tf.read()
        for stale_num in ["86,2%", "86{,}2%", "19{,}57", "22{,}40", "18{,}31"]:
            if stale_num in t_content:
                errors.append(f"Valor legado proibido '{stale_num}' encontrado em Docs/monografia/tabelas/{tname}")

    for stale_num in ["86,2%", "86{,}2%"]:
        if stale_num in main_tex:
            errors.append(f"Valor legado proibido '{stale_num}' encontrado em Docs/monografia/Main.tex")

    with open(os.path.join(docs_dir, "tabelas", "tab_tcc2_mc.tex"), "r", encoding="utf-8") as tf:
        tab_tcc2_txt = tf.read()
    if fmt_br(tcc2["stress_40"]["p85"], 2) not in tab_tcc2_txt or fmt_br(tcc2["stress_40"]["prob_target"], 1) not in tab_tcc2_txt:
        errors.append("tab_tcc2_mc.tex dessincronizada com stress_40 do poc_results.json")

    # 3. Verifica poder estatistico de H1b (M=500, TOST binomial exato 90% [416, 438]) em Main.tex
    p500 = pwr["tost_exact_90"]["500"]["0.85"]
    w90_500 = pwr["wilson_90"]["500"]["0.85"]
    if p500["k_min"] != 416 or p500["k_max"] != 438:
        errors.append(f"Regiao exata TOST M=500 inesperada: [{p500['k_min']}, {p500['k_max']}] (esperado [416, 438])")
    if p500["type_i_error_lower_pct"] > 5.0 or p500["type_i_error_upper_pct"] > 5.0:
        errors.append(
            f"Erro Tipo I exato TOST acima de 5%: {p500['type_i_error_lower_pct']}% / {p500['type_i_error_upper_pct']}%"
        )
    if w90_500["type_i_error_lower_pct"] <= 5.0:
        errors.append("Wilson 90% nao acusou inflacao de Erro Tipo I > 5% como esperado na demonstracao PK-02.2")
    expected_p85_pwr_1d = fmt_br_plain(p500["power_pct"], 1)   # 83,9
    expected_p85_pwr_2d = fmt_br(p500["power_pct"], 2)         # 83{,}93
    if expected_p85_pwr_1d not in main_tex and expected_p85_pwr_2d not in main_tex:
        errors.append(f"Poder TOST exato {expected_p85_pwr_1d}% / {expected_p85_pwr_2d}% para M=500 nao encontrado em Docs/monografia/Main.tex")
    if "416" not in main_tex or "438" not in main_tex:
        errors.append("Intervalo de aceitacao exato [416, 438] para M=500 nao encontrado em Docs/monografia/Main.tex")

    # 4. Verifica dados de piso e invariante raw_eng >= ct_bd do dataset_backtesting.json em Main.tex
    for case in ds["cases"]:
        fc = case["floor_count"]
        tot = len(case["items"])
        pct_tex = fmt_br(case["floor_fraction"] * 100.0, 1)
        expected_str = f"{fc} / {tot} (${pct_tex}\\%$)"
        if expected_str not in main_tex:
            errors.append(f"Contagem no piso '{expected_str}' de {case['target']} nao encontrada em Docs/monografia/Main.tex")
        for it in case["items"]:
            if it["first_commit_to_merged_raw_bd"] + 1e-9 < it["created_to_merged_bd"]:
                errors.append(f"Anomalia raw_eng < ct_bd em {case['target']}: {it['anon_id']}")

    # 5. Verifica classificacao exploratoria sem vazamento e fallback no Caso 2 (v2.41.0)
    if pilot.get("classification") != "exploratory_stochastic_parameterization_only":
        errors.append(f"pilot_backtesting_results.json classification inesperada: {pilot.get('classification')}")
    if pilot.get("temporal_leakage") is not False:
        errors.append(f"pilot_backtesting_results.json temporal_leakage deveria ser false, obtido: {pilot.get('temporal_leakage')}")
    wp_v241 = pilot["window_parameterizations"][1]
    if wp_v241["selected_rule"] != "fallback_bootstrap_empirico":
        errors.append("Caso 2 (v2.41.0) nao acionou fallback_bootstrap_empirico em pilot_backtesting_results.json")

    # 6. Verifica consistencia com README.md
    crn_ratio_str = f"{fmt_br_plain(crn['variance_reduction_ratio'], 1)}x"
    pwr_2d_str = f"{fmt_br_plain(p500['power_pct'], 2)}%"
    pwr_1d_str = f"{fmt_br_plain(p500['power_pct'], 1)}%"
    s40_p85_str = fmt_br_plain(tcc2["stress_40"]["p85"], 2)
    s40_prob_plain = fmt_br_plain(tcc2["stress_40"]["prob_target"], 1)
    s40_prob_tex = fmt_br(tcc2["stress_40"]["prob_target"], 1)
    w4_med_str = fmt_br_plain(b_res[3]["median_ms"], 1)
    for token in ["94,1%", crn_ratio_str, "12,00", w4_med_str]:
        if token not in readme_md:
            errors.append(f"Valor chave '{token}' ausente em README.md")
    if (
        f"{s40_prob_plain}%" not in readme_md
        and (s40_prob_plain + r"\%") not in readme_md
        and (s40_prob_tex + r"\%") not in readme_md
    ):
        errors.append(f"Valor chave '{s40_prob_plain}%' ausente em README.md")
    if pwr_2d_str not in readme_md and pwr_1d_str not in readme_md:
        errors.append(f"Valor chave '{pwr_2d_str}' / '{pwr_1d_str}' ausente em README.md")
    if s40_p85_str not in readme_md and fmt_br(tcc2["stress_40"]["p85"], 2) not in readme_md:
        errors.append(f"Valor chave '{s40_p85_str}' ausente em README.md")

    if errors:
        print("FALHA na verificacao de consistencia numerica:")
        for e in errors:
            print(f"  [ERRO] {e}")
        sys.exit(1)

    print("=========================================================================")
    print("  [OK] check_numbers.py: 100% dos numeros sincronizados entre JSON,")
    print("       macros LaTeX, 7 tabelas geradas, Docs/monografia/Main.tex e README.md!")
    print("=========================================================================")


if __name__ == "__main__":
    main()
