#!/usr/bin/env python3
"""
Gera todas as 8 figuras do TCC 1 em PNG RGB opaco (200 DPI; > 315 DPI efetivos em A4)
diretamente em Docs/monografia/figuras/ (fonte unica de imagens):
  - fig01_dag_exemplo.png
  - fig02_pert_beta.png
  - fig03_niveis_gemeo_digital.png
  - fig04_hash_chain.png
  - fig05_arquitetura.png
  - fig06_sequencia.png
  - fig07_histograma_cdf_poc.png (lido diretamente de Codigo/data/poc_results.json — FOR-03)
  - fig08_gantt.png
"""
import json
import math
import os
import numpy as np
import matplotlib.pyplot as plt
from matplotlib.patches import FancyBboxPatch
from PIL import Image, PngImagePlugin

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
CODIGO_DIR = os.path.dirname(SCRIPT_DIR)
ROOT_DIR = os.path.dirname(CODIGO_DIR)
DATA_DIR = os.path.join(CODIGO_DIR, "data")
FIG_DIR = os.path.join(ROOT_DIR, "Docs", "monografia", "figuras")
POC_RESULTS_PATH = os.path.join(DATA_DIR, "poc_results.json")

os.makedirs(FIG_DIR, exist_ok=True)
UNB_BLUE = "#1D4B7C"
UNB_GREEN = "#008C4F"
RED_ACCENT = "#B82E2E"
ORANGE_ACCENT = "#D97706"

plt.rcParams["font.sans-serif"] = "DejaVu Sans"
plt.rcParams["axes.edgecolor"] = "#333333"
FIGURE_DPI = 200


def save_rgb_png(fig, path):
    """Salva figura como PNG RGB opaco (8-bit, sem canal alfa) exclusivamente em Docs/monografia/figuras/."""
    out_path = os.path.join(FIG_DIR, os.path.basename(path))
    fig.savefig(out_path, dpi=FIGURE_DPI, facecolor="white", bbox_inches="tight", pad_inches=0.12)
    plt.close(fig)
    with Image.open(out_path) as img:
        bg = Image.new("RGB", img.size, (255, 255, 255))
        if img.mode in ("RGBA", "LA") or (img.mode == "P" and "transparency" in img.info):
            rgba = img.convert("RGBA")
            bg.paste(rgba, mask=rgba.split()[3])
        else:
            bg.paste(img.convert("RGB"))
        bg.save(out_path, format="PNG", optimize=True, pnginfo=PngImagePlugin.PngInfo(), dpi=(FIGURE_DPI, FIGURE_DPI))


def draw_box(ax, x, y, w, h, text, fc="#EBF2F9", ec=UNB_BLUE, fontsize=9, weight="normal"):
    box = FancyBboxPatch(
        (x, y),
        w,
        h,
        boxstyle="round,pad=0.03,rounding_size=0.08",
        facecolor=fc,
        edgecolor=ec,
        linewidth=1.6,
    )
    ax.add_patch(box)
    if text:
        ax.text(
            x + w / 2,
            y + h / 2,
            text,
            ha="center",
            va="center",
            fontsize=fontsize,
            weight=weight,
            color="#111111",
        )


def fig01():
    fig, ax = plt.subplots(figsize=(8.2, 4.2), dpi=FIGURE_DPI)
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 5)
    ax.axis("off")

    draw_box(ax, 0.5, 1.9, 2.4, 1.2, "T1: Modelagem\n(2 / 3 / 5) d.u.\n$\\mu=3{,}17$ | $\\sigma^2=0{,}31$", weight="bold")
    draw_box(ax, 3.8, 3.2, 2.4, 1.2, "T2: APIs\n(4 / 5 / 9) d.u.\n$\\mu=5{,}50$ | $\\sigma^2=0{,}75$", weight="bold")
    draw_box(
        ax,
        3.8,
        0.6,
        2.4,
        1.2,
        "T3: Mensageria\n(3 / 5 / 10) d.u.\n$\\mu=5{,}50$ | $\\sigma^2=1{,}61$",
        fc="#FDECEC",
        ec=RED_ACCENT,
        weight="bold",
    )
    draw_box(ax, 7.1, 1.9, 2.4, 1.2, "T4: Front-end\n(1 / 2 / 4) d.u.\n$\\mu=2{,}17$ | $\\sigma^2=0{,}31$", weight="bold")

    arrow_kw = dict(arrowstyle="->", color=UNB_BLUE, lw=1.8, mutation_scale=15)
    ax.annotate("", xy=(3.75, 3.8), xytext=(2.95, 2.7), arrowprops=arrow_kw)
    ax.annotate("", xy=(3.75, 1.2), xytext=(2.95, 2.3), arrowprops=arrow_kw)
    ax.annotate("", xy=(7.05, 2.7), xytext=(6.25, 3.8), arrowprops=arrow_kw)
    ax.annotate("", xy=(7.05, 2.3), xytext=(6.25, 1.2), arrowprops=arrow_kw)

    fig.tight_layout()
    save_rgb_png(fig, "figuras/fig01_dag_exemplo.png")


def beta_pdf(x, to, tm, tp, lam=4.0):
    diff = tp - to
    a = 1.0 + lam * (tm - to) / diff
    b = 1.0 + lam * (tp - tm) / diff
    u = np.clip((x - to) / diff, 0, 1)
    B = math.gamma(a) * math.gamma(b) / math.gamma(a + b)
    pdf = np.where((x >= to) & (x <= tp), (u ** (a - 1)) * ((1 - u) ** (b - 1)) / (B * diff), 0.0)
    return pdf


def fig02():
    fig, ax = plt.subplots(figsize=(8.2, 4.2), dpi=FIGURE_DPI)
    x = np.linspace(0.5, 10.5, 1000)
    tasks = [
        ("T1 (2, 3, 5)", 2, 3, 5, UNB_BLUE),
        ("T2 (4, 5, 9)", 4, 5, 9, ORANGE_ACCENT),
        ("T3 (3, 5, 10)", 3, 5, 10, RED_ACCENT),
        ("T4 (1, 2, 4)", 1, 2, 4, UNB_GREEN),
    ]
    for label, to, tm, tp, col in tasks:
        y = beta_pdf(x, to, tm, tp)
        ax.plot(x, y, label=label, color=col, lw=2.2)
        ax.axvline(tm, color=col, ls=":", lw=1.2, alpha=0.8)

    ax.set_xlabel("Duração da tarefa (dias úteis)")
    ax.set_ylabel("Densidade de probabilidade $f(t)$")
    ax.grid(True, linestyle="--", alpha=0.35)
    ax.legend(frameon=True)
    fig.tight_layout()
    save_rgb_png(fig, "figuras/fig02_pert_beta.png")


def fig03():
    fig, ax = plt.subplots(figsize=(10.8, 4.3), dpi=FIGURE_DPI)
    ax.set_xlim(0, 13.8)
    ax.set_ylim(0, 4.6)
    ax.axis("off")

    cols = [
        (0.3, "1. Modelo Digital", "Processo Real", "Modelo Virtual", "Manual", "Manual", "#F3F4F6", "#6B7280"),
        (
            4.8,
            "2. Sombra Digital Governada\n(Este Trabalho)",
            "Repositório Git / CI",
            "Estado + Monte Carlo + IA",
            "Auto\n(Webhooks)",
            "Fila +\nAprovação\nHumana",
            "#E6F4EA",
            UNB_GREEN,
        ),
        (9.3, "3. Gêmeo Digital Pleno", "Processo Real", "Modelo Virtual", "Automático", "Automático", "#EBF2F9", UNB_BLUE),
    ]
    pill = dict(boxstyle="round,pad=0.15", fc="white", ec="none", alpha=0.92)
    for x0, title, top_lbl, bot_lbl, f1, f2, fc, ec in cols:
        ax.text(x0 + 2.0, 4.15, title, ha="center", va="center", fontsize=9.5, weight="bold", color=ec)
        draw_box(ax, x0 + 0.3, 2.55, 3.4, 0.9, top_lbl, fc=fc, ec=ec, fontsize=8.5, weight="bold")
        draw_box(ax, x0 + 0.3, 0.45, 3.4, 0.9, bot_lbl, fc=fc, ec=ec, fontsize=8.5, weight="bold")
        # Seta descendente (Real -> Virtual) à direita do centro da coluna
        ax.annotate("", xy=(x0 + 2.55, 1.42), xytext=(x0 + 2.55, 2.48), arrowprops=dict(arrowstyle="->", color=UNB_BLUE, lw=1.8))
        ax.text(x0 + 2.68, 1.95, f1, fontsize=7.6, va="center", ha="left", color=UNB_BLUE, weight="bold", bbox=pill)
        # Seta ascendente (Virtual -> Real) à esquerda do centro da coluna
        ax.annotate("", xy=(x0 + 1.45, 2.48), xytext=(x0 + 1.45, 1.42), arrowprops=dict(arrowstyle="->", color=RED_ACCENT, lw=1.6, ls="--"))
        ax.text(x0 + 1.32, 1.95, f2, fontsize=7.6, va="center", ha="right", color=RED_ACCENT, weight="bold", bbox=pill)

    fig.tight_layout()
    save_rgb_png(fig, "figuras/fig03_niveis_gemeo_digital.png")


def fig04():
    fig, ax = plt.subplots(figsize=(9.5, 3.9), dpi=FIGURE_DPI)
    ax.set_xlim(0, 12.2)
    ax.set_ylim(0, 4.3)
    ax.axis("off")

    draw_box(
        ax,
        0.4,
        1.85,
        3.2,
        1.75,
        "Registro $k-1$\n$\\mathrm{payload}_{k-1}$\nprev: $H_{k-2}$\nhash: $H_{k-1}$",
        fontsize=8.8,
        weight="bold",
    )
    draw_box(
        ax,
        4.5,
        1.85,
        3.2,
        1.75,
        "Registro $k$\n$\\mathrm{payload}_k$\nprev: $H_{k-1}$\nhash: $H_k$",
        fontsize=8.8,
        weight="bold",
    )
    draw_box(
        ax,
        8.6,
        1.85,
        3.2,
        1.75,
        "Registro $k+1$\n$\\mathrm{payload}_{k+1}$\nprev: $H_k$\nhash: $H_{k+1}$",
        fontsize=8.8,
        weight="bold",
    )
    draw_box(
        ax,
        7.4,
        0.2,
        4.4,
        1.05,
        "Âncora Externa Assinada (HMAC / Ed25519)\nReplica $(k+1,\\,H_{k+1})$ fora do PostgreSQL ($K=100$ ou $15$ min)",
        fc="#E6F4EA",
        ec=UNB_GREEN,
        fontsize=8.0,
        weight="bold",
    )

    ax.annotate("", xy=(4.45, 2.72), xytext=(3.65, 2.72), arrowprops=dict(arrowstyle="->", color=UNB_BLUE, lw=2))
    ax.annotate("", xy=(8.55, 2.72), xytext=(7.75, 2.72), arrowprops=dict(arrowstyle="->", color=UNB_BLUE, lw=2))
    ax.annotate("", xy=(10.2, 1.30), xytext=(10.2, 1.80), arrowprops=dict(arrowstyle="->", color=UNB_GREEN, lw=2))

    fig.tight_layout()
    save_rgb_png(fig, "figuras/fig04_hash_chain.png")


def fig05():
    fig, ax = plt.subplots(figsize=(10.5, 4.7), dpi=FIGURE_DPI)
    ax.set_xlim(0, 14.0)
    ax.set_ylim(0, 5.4)
    ax.axis("off")

    # Duas sub-redes separadas; Core PM Service atua em modo dual-homed atravessando a fronteira
    draw_box(ax, 0.3, 0.35, 6.75, 4.65, "", fc="#F5F9FD", ec=UNB_BLUE)
    ax.text(0.55, 4.70, "Sub-rede: core-net (Isolada da IA)", fontsize=9.2, weight="bold", color=UNB_BLUE)

    draw_box(ax, 7.25, 0.35, 6.45, 4.65, "", fc="#FEF6F6", ec=RED_ACCENT)
    ax.text(13.45, 4.70, "Sub-rede: ai-isolated-net", fontsize=9.2, weight="bold", color=RED_ACCENT, ha="right")

    draw_box(ax, 0.55, 0.75, 2.75, 1.2, "PostgreSQL 16\nSKIP LOCKED + Audit", fc="#FFFFFF", ec=UNB_BLUE, weight="bold")
    draw_box(ax, 3.65, 0.75, 2.95, 1.2, "Simulation Worker\nCPM + Monte Carlo", fc="#FFFFFF", ec=UNB_BLUE, weight="bold")
    draw_box(
        ax,
        4.95,
        2.95,
        4.30,
        1.25,
        "Core PM Service (Go)\nDual-Homed: RBAC + Fail-Closed",
        fc="#FFFBEB",
        ec="#B45309",
        weight="bold",
    )
    draw_box(ax, 10.15, 2.95, 3.25, 1.25, "AI Agent Service\n(Sem Driver/Rota SQL)", fc="#FFFFFF", ec=RED_ACCENT, weight="bold")
    draw_box(ax, 10.15, 0.75, 3.25, 1.2, "API LLM Externa / Local\n(HTTPS / TLS)", fc="#F3F4F6", ec="#4B5563", weight="bold")

    ax.text(7.10, 2.35, "Fronteira\nDual-Homed", ha="center", va="center", fontsize=7.2, color="#B45309", weight="bold")

    # Setas internas à core-net e ponte dual-homed
    ax.annotate("", xy=(1.95, 2.00), xytext=(5.55, 2.92), arrowprops=dict(arrowstyle="<->", color=UNB_BLUE, lw=1.8))
    ax.annotate("", xy=(5.15, 2.00), xytext=(6.25, 2.92), arrowprops=dict(arrowstyle="<->", color=UNB_BLUE, lw=1.8))
    ax.annotate("", xy=(3.62, 1.35), xytext=(3.33, 1.35), arrowprops=dict(arrowstyle="<->", color=UNB_BLUE, lw=1.8))

    # Comunicação permitida via ai-isolated-net entre Core PM Service (NIC secundária) e AI Agent
    ax.annotate("", xy=(10.10, 3.57), xytext=(9.30, 3.57), arrowprops=dict(arrowstyle="<->", color=RED_ACCENT, lw=1.8))
    ax.text(
        9.70,
        4.32,
        "Allowlist JSON\n(HTTP/REST)",
        ha="center",
        va="center",
        fontsize=7.3,
        color=RED_ACCENT,
        weight="bold",
        bbox=dict(boxstyle="round,pad=0.16", fc="white", ec=RED_ACCENT, lw=0.8),
    )
    ax.annotate("", xy=(11.77, 2.00), xytext=(11.77, 2.92), arrowprops=dict(arrowstyle="<->", color="#4B5563", lw=1.8))

    fig.tight_layout()
    save_rgb_png(fig, "figuras/fig05_arquitetura.png")


def fig06():
    fig, ax = plt.subplots(figsize=(10.2, 4.7), dpi=FIGURE_DPI)
    ax.set_xlim(0, 12.8)
    ax.set_ylim(0, 5.3)
    ax.axis("off")

    actors = [
        (1.0, "GitHub"),
        (3.1, "Core PM"),
        (5.2, "PostgreSQL"),
        (7.3, "Sim Worker"),
        (9.4, "AI Service"),
        (11.6, "Gerente (UI)"),
    ]
    for x, name in actors:
        draw_box(ax, x - 0.85, 4.35, 1.7, 0.62, name, fontsize=8.5, weight="bold")
        ax.plot([x, x], [0.25, 4.35], ls="--", color="#9CA3AF", lw=1.1, zorder=1)

    steps = [
        (3.85, 1.0, 3.1, "1. Webhook assinado (HMAC)", UNB_BLUE),
        (3.38, 3.1, 5.2, "2. state_version++ e enfileira Job", UNB_BLUE),
        (2.91, 7.3, 5.2, "3. Consome Job via SKIP LOCKED", UNB_BLUE),
        (2.44, 7.3, 5.2, "4. Persiste $P_{50}, P_{85}, P_{95}$ e $IC_i$", UNB_BLUE),
        (1.97, 3.1, 9.4, "5. Envia Allowlist sanitizada", RED_ACCENT),
        (1.50, 9.4, 3.1, "6. JSON validado (Fail-Closed)", RED_ACCENT),
        (1.03, 3.1, 11.6, "7. SSE: Fila Cat. 2 com $\\Delta P_{85}$", UNB_GREEN),
        (0.52, 11.6, 5.2, "8. Aprovação humana $\\rightarrow$ SHA-256", UNB_BLUE),
    ]
    pill = dict(boxstyle="round,pad=0.14", fc="white", ec="none", alpha=0.94)
    for y, x1, x2, lbl, col in steps:
        ax.annotate("", xy=(x2, y), xytext=(x1, y), arrowprops=dict(arrowstyle="->", color=col, lw=1.7), zorder=3)
        ax.text((x1 + x2) / 2, y + 0.13, lbl, ha="center", va="center", fontsize=7.8, color=col, weight="bold", bbox=pill, zorder=4)

    fig.tight_layout()
    save_rgb_png(fig, "figuras/fig06_sequencia.png")


def fig07():
    """Plota o histograma e as ECDFs dos Cenarios A e B diretamente de Codigo/data/poc_results.json (FOR-03)."""
    with open(POC_RESULTS_PATH, "r", encoding="utf-8") as f:
        poc = json.load(f)

    ecdf = poc["ecdf_ab"]
    sa = poc["scenario_a"]
    sb = poc["scenario_b"]

    grid_x = np.array(ecdf["grid_x"])
    ecdf_a = np.array(ecdf["ecdf_a"])
    ecdf_b = np.array(ecdf["ecdf_b"])
    centers = np.array(ecdf["bin_centers"])
    dens_a = np.array(ecdf["density_a"])
    dens_b = np.array(ecdf["density_b"])
    width = (centers[1] - centers[0]) if len(centers) > 1 else 0.2

    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(10.0, 4.1), dpi=FIGURE_DPI)
    ax1.bar(centers, dens_a, width=width * 0.95, alpha=0.55, color=UNB_BLUE, label="Cenário A (Sem risco)")
    ax1.bar(centers, dens_b, width=width * 0.95, alpha=0.45, color=RED_ACCENT, label="Cenário B (Com R1)")
    ax1.axvline(sa["cpm_deterministic"], color=UNB_GREEN, ls="--", lw=1.8, label=f"CPM = {sa['cpm_deterministic']:.2f}".replace(".", ",") + " d.u.")
    ax1.set_xlabel("Prazo de término $F$ (dias úteis)")
    ax1.set_ylabel("Densidade empírica (PoC em Go)")
    ax1.grid(True, ls="--", alpha=0.3)
    ax1.legend(fontsize=8)

    lbl_a = f"CDF Cenário A ($P_{{50}}={sa['p50']:.2f}$ | $P_{{85}}={sa['p85']:.2f}$)".replace(".", ",")
    lbl_b = f"CDF Cenário B ($P_{{50}}={sb['p50']:.2f}$ | $P_{{85}}={sb['p85']:.2f}$)".replace(".", ",")
    ax2.plot(grid_x, ecdf_a, color=UNB_BLUE, lw=2.2, label=lbl_a)
    ax2.plot(grid_x, ecdf_b, color=RED_ACCENT, lw=2.2, label=lbl_b)
    ax2.axvline(sa["cpm_deterministic"], color=UNB_GREEN, ls="--", lw=1.8, label=f"CPM = {sa['cpm_deterministic']:.2f}".replace(".", ","))
    ax2.axhline(0.50, color="gray", ls=":", lw=1)
    ax2.axhline(0.85, color="gray", ls=":", lw=1)
    ax2.set_xlabel("Prazo de término $F$ (dias úteis)")
    ax2.set_ylabel("Probabilidade acumulada $P(F \\leq t)$")
    ax2.grid(True, ls="--", alpha=0.3)
    ax2.legend(fontsize=8, loc="lower right")

    fig.tight_layout()
    save_rgb_png(fig, "figuras/fig07_histograma_cdf_poc.png")


def fig08():
    with open(POC_RESULTS_PATH, "r", encoding="utf-8") as f:
        poc = json.load(f)
    tb = poc["tcc2_schedule_base"]
    tr = poc["tcc2_schedule_risks"]
    tc = poc["tcc2_schedule_cuts"]

    tasks = [
        ("E01 [Trilha B]: Extrator GitHub + Replay", 1, 2, UNB_GREEN),
        ("E02 [Trilha A]: Contrato + SQL + Âncora", 1, 2, UNB_BLUE),
        ("E03 [Trilha A]: CRUD + Paginação + RBAC", 3, 4, UNB_BLUE),
        ("E04 [Trilha B]: Ingestão + Outbox + DLQ", 3, 4, UNB_GREEN),
        ("E05 [Trilha A]: Motor Monte Carlo + CRN", 5, 6, UNB_BLUE),
        ("E06 [Trilha A]: Sim Worker + Persistência", 7, 8, UNB_BLUE),
        ("E08 [Trilha B]: Dim. 2 (10 Regimes, M=500)", 7, 8, UNB_GREEN),
        ("E07 [Trilha A]: Propostas + Guardrails", 9, 10, UNB_BLUE),
        ("E09 [Trilha B]: Dim. 3 (Backtesting Real)", 9, 11, UNB_GREEN),
        ("E11 [Trilha A]: UI Web Mínima Anti-Viés", 11, 12, UNB_BLUE),
        ("E10 [Trilha B]: Dim. 5 (300 Cenários + Abl.)", 12, 14, UNB_GREEN),
        ("E12 [Conjunto]: Consolidação Estatística", 15, 16, ORANGE_ACCENT),
        ("E13 [Conjunto]: Redação Final + Ameaças", 17, 18, ORANGE_ACCENT),
        ("E14 [Conjunto]: Revisão ABNT + Defesa + Buffer", 19, 20, ORANGE_ACCENT),
        ("D01 [Desejável]: Grafo Visual UI", 13, 14, "#6B7280"),
        ("D02 [Desejável]: Painel de Sensibilidade", 15, 16, "#6B7280"),
        ("D03 [Desejável]: 3º Juiz Externo Adicional", 16, 17, "#6B7280"),
    ]
    fig, ax = plt.subplots(figsize=(10.0, 5.4), dpi=FIGURE_DPI)
    y_pos = np.arange(len(tasks))[::-1]
    for idx, (label, start, end, col) in enumerate(tasks):
        y = y_pos[idx]
        ax.barh(y, end - start + 0.8, left=start - 0.4, height=0.6, color=col, alpha=0.85, edgecolor="#222222")

    ax.axhline(2.5, color=RED_ACCENT, ls="--", lw=1.6)
    ax.text(
        0.75,
        2.5,
        "Linha de Corte de Escopo (Abaixo: Desejáveis D01–D03)",
        color=RED_ACCENT,
        fontsize=8.0,
        weight="bold",
        va="center",
        bbox=dict(boxstyle="round,pad=0.2", fc="white", ec=RED_ACCENT, lw=0.9),
    )
    title_str = (
        f"Base: P50={tb['p50']:.2f} | P85={tb['p85']:.2f} sem  •  "
        f"Com Riscos: P50={tr['p50']:.2f} | P85={tr['p85']:.2f} sem (P≤20 = {tr['prob_target']:.1f}%)\n"
        f"Estresse (+25% O/M/P + Riscos) c/ Cortes Pré-Registrados: P85={tc['p85']:.2f} sem (P≤20 = {tc['prob_target']:.1f}%)"
    ).replace(".", ",")
    ax.set_title(title_str, fontsize=8.3, weight="bold", color="#1F2937", pad=8)
    ax.set_yticks(y_pos)
    ax.set_yticklabels([t[0] for t in tasks], fontsize=8)
    ax.set_xlim(0.5, 20.5)
    ax.set_xticks(range(1, 21))
    ax.set_xlabel(
        "Semanas Úteis do Cronograma do TCC 2 (01/03/2027 a 16/07/2027 • CEPE UnB nº 0151/2026: Aulas 15/03–16/07)\n"
        "Trilha A: Azul | Trilha B: Verde | Conjunto: Laranja | Desejáveis: Cinza",
        fontsize=8.3,
    )
    ax.grid(True, axis="x", ls="--", alpha=0.35)
    fig.tight_layout()
    save_rgb_png(fig, "figuras/fig08_gantt.png")


if __name__ == "__main__":
    fig01()
    fig02()
    fig03()
    fig04()
    fig05()
    fig06()
    fig07()
    fig08()
    print(f"8 figuras geradas com sucesso em {FIG_DIR}")
