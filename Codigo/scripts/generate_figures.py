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
        max_w = 1400
        if bg.width > max_w:
            new_h = round(bg.height * (max_w / bg.width))
            bg = bg.resize((max_w, new_h), Image.Resampling.LANCZOS)
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
    fig, ax = plt.subplots(figsize=(11.2, 4.5), dpi=FIGURE_DPI)
    ax.set_xlim(0, 14.2)
    ax.set_ylim(-0.15, 4.75)
    ax.axis("off")

    cols = [
        (0.3, "1. Modelo Digital", "Processo Real", "Modelo Virtual", "Manual", "Manual", "#F3F4F6", "#6B7280", None),
        (
            5.0,
            "2. Sombra Digital Governada\n(Este Trabalho)",
            "Repositório Git / CI",
            "Modelo Virtual + MC + Fila IA",
            "Auto\n(Webhooks)",
            "Manual Externo\n(Sem escrita Git)",
            "#E6F4EA",
            UNB_GREEN,
            "Aprovação humana atualiza somente a baseline virtual",
        ),
        (9.7, "3. Gêmeo Digital Pleno", "Processo Real", "Modelo Virtual", "Automático", "Automático", "#EBF2F9", UNB_BLUE, None),
    ]
    pill = dict(boxstyle="round,pad=0.15", fc="white", ec="none", alpha=0.92)
    for x0, title, top_lbl, bot_lbl, f1, f2, fc, ec, foot in cols:
        ax.text(x0 + 2.0, 4.25, title, ha="center", va="center", fontsize=9.5, weight="bold", color=ec)
        draw_box(ax, x0 + 0.25, 2.65, 3.5, 0.9, top_lbl, fc=fc, ec=ec, fontsize=8.4, weight="bold")
        draw_box(ax, x0 + 0.25, 0.55, 3.5, 0.9, bot_lbl, fc=fc, ec=ec, fontsize=8.4, weight="bold")
        # Seta descendente (Real -> Virtual) à direita do centro da coluna
        ax.annotate("", xy=(x0 + 2.65, 1.52), xytext=(x0 + 2.65, 2.58), arrowprops=dict(arrowstyle="->", color=UNB_BLUE, lw=1.8))
        ax.text(x0 + 2.78, 2.05, f1, fontsize=7.5, va="center", ha="left", color=UNB_BLUE, weight="bold", bbox=pill)
        # Seta ascendente (Virtual -> Real): automática no Pleno, manual externa nos demais
        is_auto_rev = foot is None and title.startswith("3.")
        rev_color = UNB_BLUE if is_auto_rev else "#6B7280"
        rev_ls = "-" if is_auto_rev else "--"
        ax.annotate("", xy=(x0 + 1.35, 2.58), xytext=(x0 + 1.35, 1.52), arrowprops=dict(arrowstyle="->", color=rev_color, lw=1.6, ls=rev_ls))
        ax.text(x0 + 1.22, 2.05, f2, fontsize=7.4, va="center", ha="right", color=rev_color, weight="bold", bbox=pill)
        if foot:
            ax.text(
                x0 + 2.0,
                0.15,
                foot,
                ha="center",
                va="center",
                fontsize=7.1,
                color=UNB_GREEN,
                weight="bold",
                bbox=dict(boxstyle="round,pad=0.16", fc="#F0FDF4", ec=UNB_GREEN, lw=0.7),
            )

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
        "Âncora Externa Assinada (Ed25519)\nReplica $(k+1,\\,H_{k+1})$ fora do PostgreSQL ($K=100$ ou $15$ min)",
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
    fig, ax = plt.subplots(figsize=(11.0, 5.4), dpi=FIGURE_DPI)
    ax.set_xlim(0, 14.5)
    ax.set_ylim(-0.1, 6.9)
    ax.axis("off")

    # Entradas externas de borda (GitHub Webhooks + Navegador com Cartão Mínimo CSP)
    draw_box(ax, 0.55, 5.55, 3.25, 1.05, "GitHub Webhooks\n(TLS + HMAC-SHA256)", fc="#F3F4F6", ec="#4B5563", fontsize=8.3, weight="bold")
    draw_box(ax, 4.35, 5.55, 3.70, 1.05, "Navegador / Cartão Mínimo\n(JWT + RBAC + CSP Estrita)", fc="#E6F4EA", ec=UNB_GREEN, fontsize=8.3, weight="bold")

    # Duas sub-redes separadas; Core PM Service atua em modo dual-homed atravessando a fronteira
    draw_box(ax, 0.3, 0.25, 6.95, 4.75, "", fc="#F5F9FD", ec=UNB_BLUE)
    ax.text(0.55, 4.72, "Sub-rede: core-net (Isolada da IA)", fontsize=9.2, weight="bold", color=UNB_BLUE)

    draw_box(ax, 7.50, 0.25, 6.70, 4.75, "", fc="#FEF6F6", ec=RED_ACCENT)
    ax.text(13.95, 4.72, "Sub-rede: ai-isolated-net", fontsize=9.2, weight="bold", color=RED_ACCENT, ha="right")

    draw_box(ax, 0.55, 0.65, 2.85, 1.2, "PostgreSQL 16\nSKIP LOCKED + Audit (Ed25519)", fc="#FFFFFF", ec=UNB_BLUE, fontsize=8.2, weight="bold")
    draw_box(ax, 3.75, 0.65, 3.10, 1.2, "Simulation Worker\nCPM + Monte Carlo (Go)", fc="#FFFFFF", ec=UNB_BLUE, fontsize=8.2, weight="bold")
    draw_box(
        ax,
        4.65,
        2.95,
        4.65,
        1.25,
        "Core PM Service (Go)\nDual-Homed: RBAC + Fail-Closed",
        fc="#FFFBEB",
        ec="#B45309",
        fontsize=8.4,
        weight="bold",
    )
    draw_box(ax, 10.45, 2.95, 3.45, 1.25, "AI Agent Service\n(Sem Driver/Rota SQL)", fc="#FFFFFF", ec=RED_ACCENT, fontsize=8.3, weight="bold")
    draw_box(
        ax,
        10.45,
        0.65,
        3.45,
        1.30,
        "Proxy de Egresso / Firewall\nAllowlist TLS -> Llama-3 (temp=0,0)",
        fc="#F3F4F6",
        ec="#4B5563",
        fontsize=8.0,
        weight="bold",
    )

    ax.text(7.37, 2.35, "Fronteira\nDual-Homed", ha="center", va="center", fontsize=7.2, color="#B45309", weight="bold")

    # Setas de entrada (Ingresso -> Core PM Service)
    ax.annotate("", xy=(5.25, 4.22), xytext=(2.20, 5.53), arrowprops=dict(arrowstyle="->", color="#4B5563", lw=1.7))
    ax.annotate("", xy=(6.45, 4.22), xytext=(6.20, 5.53), arrowprops=dict(arrowstyle="<->", color=UNB_GREEN, lw=1.8))

    # Setas internas à core-net e ponte dual-homed
    ax.annotate("", xy=(1.98, 1.88), xytext=(5.55, 2.93), arrowprops=dict(arrowstyle="<->", color=UNB_BLUE, lw=1.8))
    ax.annotate("", xy=(5.30, 1.88), xytext=(6.45, 2.93), arrowprops=dict(arrowstyle="<->", color=UNB_BLUE, lw=1.8))
    ax.annotate("", xy=(3.73, 1.25), xytext=(3.42, 1.25), arrowprops=dict(arrowstyle="<->", color=UNB_BLUE, lw=1.8))

    # Comunicação permitida via ai-isolated-net entre Core PM Service (NIC secundária) e AI Agent
    ax.annotate("", xy=(10.42, 3.57), xytext=(9.33, 3.57), arrowprops=dict(arrowstyle="<->", color=RED_ACCENT, lw=1.8))
    ax.text(
        9.87,
        4.30,
        "Allowlist JSON\n(HTTP/REST)",
        ha="center",
        va="center",
        fontsize=7.2,
        color=RED_ACCENT,
        weight="bold",
        bbox=dict(boxstyle="round,pad=0.16", fc="white", ec=RED_ACCENT, lw=0.8),
    )
    ax.annotate("", xy=(12.17, 1.98), xytext=(12.17, 2.93), arrowprops=dict(arrowstyle="<->", color="#4B5563", lw=1.8))
    ax.text(
        12.17,
        2.45,
        "Egresso Restrito",
        ha="center",
        va="center",
        fontsize=7.0,
        color="#374151",
        weight="bold",
        bbox=dict(boxstyle="round,pad=0.14", fc="white", ec="#9CA3AF", lw=0.7),
    )

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
        (1.03, 3.1, 11.6, "7. HTTP/UI: Fila Cat. 2 com $\\Delta P_{85}$", UNB_GREEN),
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
    ax1.axvline(11.0, color="#111827", ls="-.", lw=1.6, label="Meta = 11,00 d.u.")
    ax1.set_xlabel("Prazo de término $F$ (dias úteis)")
    ax1.set_ylabel("Densidade empírica (PoC em Go)")
    ax1.grid(True, ls="--", alpha=0.3)
    ax1.legend(fontsize=8)

    lbl_a = f"CDF Cenário A ($P_{{50}}={sa['p50']:.2f}$ | $P_{{85}}={sa['p85']:.2f}$)".replace(".", ",")
    lbl_b = f"CDF Cenário B ($P_{{50}}={sb['p50']:.2f}$ | $P_{{85}}={sb['p85']:.2f}$)".replace(".", ",")
    ax2.plot(grid_x, ecdf_a, color=UNB_BLUE, lw=2.2, label=lbl_a)
    ax2.plot(grid_x, ecdf_b, color=RED_ACCENT, lw=2.2, label=lbl_b)
    ax2.axvline(sa["cpm_deterministic"], color=UNB_GREEN, ls="--", lw=1.8, label=f"CPM = {sa['cpm_deterministic']:.2f}".replace(".", ","))
    ax2.axvline(11.0, color="#111827", ls="-.", lw=1.6, label="Meta = 11,00 d.u.")
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
    ts40 = poc["tcc2_schedule_stress_40"]

    # Janelas ES -> EF nominais (CAMINHO CRÍTICO CPM = 12,00 sem; zero itens desejáveis D01-D03)
    tasks = [
        ("E01 [Trilha B]: Extrator cli/cli + Replay (12 h)", 0.00, 1.00, UNB_GREEN),
        ("E02 [Trilha A]: Contrato + Proxy + SQL + Ed25519 (18 h)", 0.00, 1.50, UNB_BLUE),
        ("E03 [Trilha A]: REST + Paginação + RBAC 403 (15 h)", 1.50, 2.75, UNB_BLUE),
        ("E04 [Trilha B]: Webhooks HMAC + Audit Trail (15 h)", 1.50, 2.75, UNB_GREEN),
        ("E05 [Trilha A]: Motor Monte Carlo + CRN (12 h)", 2.75, 3.75, UNB_BLUE),
        ("E06 [Trilha A]: Persistência Sim Worker <2 s (12 h)", 3.75, 4.75, UNB_BLUE),
        ("E08 [Trilha B]: Dim. 2 Enxuta (6 Regimes, R9/R10) (18 h)", 3.75, 5.25, UNB_GREEN),
        ("E07 [Trilha A]: Proxy Llama-3 + Fail-Closed (18 h)", 4.75, 6.25, UNB_BLUE),
        ("E09 [Trilha B]: Dim. 3 Descritiva (2 Níveis cli/cli) (18 h)", 5.25, 6.75, UNB_GREEN),
        ("E11 [Trilha A]: Cartão Mínimo Anti-Viés + CSP (12 h)", 6.25, 7.25, UNB_BLUE),
        ("E10 [Trilha B]: Dim. 5 (Llama-3 t=0 + H2b Expl.) (18 h)", 6.75, 8.25, UNB_GREEN),
        ("E12 [Conjunto]: Consolidação + Zenodo (15 h/autor)", 8.25, 9.50, ORANGE_ACCENT),
        ("E13 [Conjunto]: Redação Final + Ameaças (18 h/autor)", 9.50, 11.00, ORANGE_ACCENT),
        ("E14 [Conjunto]: Revisão ABNT + Banca Estoc. (12 h/autor)", 11.00, 12.00, ORANGE_ACCENT),
        ("Reserva / Buffer de Contingência (Semanas 12–20)", 12.00, 20.00, "#9CA3AF"),
    ]
    fig, ax = plt.subplots(figsize=(10.5, 5.5), dpi=FIGURE_DPI)
    y_pos = np.arange(len(tasks))[::-1]
    for idx, (label, start, end, col) in enumerate(tasks):
        y = y_pos[idx]
        alpha_v = 0.45 if col == "#9CA3AF" else 0.86
        hatch_v = "//" if col == "#9CA3AF" else None
        ax.barh(y, end - start, left=start, height=0.62, color=col, alpha=alpha_v, edgecolor="#222222", hatch=hatch_v)

    ax.axvline(tr["p85"], color=UNB_BLUE, ls="--", lw=1.6, label=f"P85 Nominal c/ Riscos e ρ=0,50 ({tr['p85']:.2f} sem)".replace(".", ","))
    ax.axvline(ts40["p85"], color=RED_ACCENT, ls="-.", lw=1.7, label=f"P85 Estresse R10 (+40% + ρ=0,50: {ts40['p85']:.2f} sem)".replace(".", ","))
    ax.axvline(20.0, color="#111827", ls="-", lw=1.8, label="Prazo Limite do Semestre (20,0 sem)")

    title_str = (
        f"Cronograma Bottom-Up Enxuto (258 h totais • CPM={tb['cpm_deterministic']:.2f} sem • Escopo Pré-Cortado sem D01–D03)\n"
        f"Nominal (ρ=0,50): P85={tr['p85']:.2f} sem (P≤20={tr['prob_target']:.1f}%)  •  "
        f"Estresse R10 (+40% + ρ=0,50): P85={ts40['p85']:.2f} sem (P≤20={ts40['prob_target']:.1f}%)"
    ).replace(".", ",")
    ax.set_title(title_str, fontsize=8.3, weight="bold", color="#1F2937", pad=8)
    ax.set_yticks(y_pos)
    ax.set_yticklabels([t[0] for t in tasks], fontsize=7.9)
    ax.set_xlim(0.0, 20.8)
    ax.set_xticks(range(0, 21, 2))
    ax.set_xlabel(
        "Semanas Acumuladas do Semestre Letivo do TCC 2 (0 a 20 semanas • CEPE UnB nº 0151/2026: Aulas 15/03–16/07/2027)\n"
        "Trilha A: Azul (87 h) | Trilha B: Verde (81 h) | Conjunto: Laranja (45 h/autor) | Buffer: Cinza (8,0 sem nominais)",
        fontsize=8.2,
    )
    ax.grid(True, axis="x", ls="--", alpha=0.35)
    ax.legend(fontsize=7.6, loc="lower left", framealpha=0.95)
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
