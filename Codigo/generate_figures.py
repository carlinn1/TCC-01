#!/usr/bin/env python3
"""
Gera todas as 8 figuras do TCC 1 em PNG RGB opaco (200 DPI; > 315 DPI efetivos em A4)
na pasta ./figuras/ com os nomes exatos referenciados no main.tex:
  - fig01_dag_exemplo.png
  - fig02_pert_beta.png
  - fig03_niveis_gemeo_digital.png
  - fig04_hash_chain.png
  - fig05_arquitetura.png
  - fig06_sequencia.png
  - fig07_histograma_cdf_poc.png
  - fig08_gantt.png
"""
import os
import math
import numpy as np
import matplotlib.pyplot as plt
from matplotlib.patches import FancyBboxPatch
from PIL import Image, PngImagePlugin

os.makedirs("figuras", exist_ok=True)
UNB_BLUE = "#1D4B7C"
UNB_GREEN = "#008C4F"
RED_ACCENT = "#B82E2E"
ORANGE_ACCENT = "#D97706"

plt.rcParams["font.sans-serif"] = "DejaVu Sans"
plt.rcParams["axes.edgecolor"] = "#333333"
FIGURE_DPI = 200


def save_rgb_png(fig, path):
    """Salva figura como PNG RGB opaco (8-bit, sem canal alfa) para incorporacao rapida em pdflatex."""
    fig.savefig(path, dpi=FIGURE_DPI, facecolor="white")
    plt.close(fig)
    with Image.open(path) as img:
        bg = Image.new("RGB", img.size, (255, 255, 255))
        if img.mode in ("RGBA", "LA") or (img.mode == "P" and "transparency" in img.info):
            rgba = img.convert("RGBA")
            bg.paste(rgba, mask=rgba.split()[3])
        else:
            bg.paste(img.convert("RGB"))
        bg.save(path, format="PNG", optimize=True, pnginfo=PngImagePlugin.PngInfo(), dpi=(FIGURE_DPI, FIGURE_DPI))
        # Sincroniza automaticamente com ../Docs/figuras e ../figuras
        fname = os.path.basename(path)
        for target_dir in ("../Docs/figuras", "../figuras"):
            os.makedirs(target_dir, exist_ok=True)
            bg.save(os.path.join(target_dir, fname), format="PNG", optimize=True, pnginfo=PngImagePlugin.PngInfo(), dpi=(FIGURE_DPI, FIGURE_DPI))


def draw_box(ax, x, y, w, h, text, fc="#EBF2F9", ec=UNB_BLUE, fontsize=9, weight="normal"):
    box = FancyBboxPatch(
        (x, y), w, h, boxstyle="round,pad=0.03,rounding_size=0.08",
        facecolor=fc, edgecolor=ec, linewidth=1.6
    )
    ax.add_patch(box)
    ax.text(x + w / 2, y + h / 2, text, ha="center", va="center", fontsize=fontsize, weight=weight, color="#111111")


def fig01():
    fig, ax = plt.subplots(figsize=(8, 4.2), dpi=FIGURE_DPI)
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 5)
    ax.axis("off")

    draw_box(ax, 0.5, 1.9, 2.4, 1.2, "T1: Modelagem\n(2 / 3 / 5) d.u.\nμ=3,17 | σ²=0,31", weight="bold")
    draw_box(ax, 3.8, 3.2, 2.4, 1.2, "T2: APIs\n(4 / 5 / 9) d.u.\nμ=5,50 | σ²=0,75", weight="bold")
    draw_box(ax, 3.8, 0.6, 2.4, 1.2, "T3: Mensageria\n(3 / 5 / 10) d.u.\nμ=5,50 | σ²=1,61", fc="#FDECEC", ec=RED_ACCENT, weight="bold")
    draw_box(ax, 7.1, 1.9, 2.4, 1.2, "T4: Front-end\n(1 / 2 / 4) d.u.\nμ=2,17 | σ²=0,31", weight="bold")

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
    fig, ax = plt.subplots(figsize=(8, 4.2), dpi=FIGURE_DPI)
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
    ax.set_ylabel("Densidade de probabilidade f(t)")
    ax.grid(True, linestyle="--", alpha=0.35)
    ax.legend(frameon=True)
    fig.tight_layout()
    save_rgb_png(fig, "figuras/fig02_pert_beta.png")


def fig03():
    fig, ax = plt.subplots(figsize=(9, 4.0), dpi=FIGURE_DPI)
    ax.set_xlim(0, 12)
    ax.set_ylim(0, 4.5)
    ax.axis("off")

    cols = [
        (0.4, "1. Modelo Digital", "Processo Real", "Modelo Virtual", "Manual", "Manual", "#F3F4F6", "#6B7280"),
        (4.2, "2. Sombra Digital Governada\n(Este Trabalho)", "Repositório Git/CI", "Estado + Monte Carlo + IA", "Auto (Webhooks)", "Fila + Aprovação Humana", "#E6F4EA", UNB_GREEN),
        (8.2, "3. Gêmeo Digital Pleno", "Processo Real", "Modelo Virtual", "Automático", "Automático", "#EBF2F9", UNB_BLUE),
    ]
    for x0, title, top_lbl, bot_lbl, f1, f2, fc, ec in cols:
        ax.text(x0 + 1.7, 4.05, title, ha="center", va="center", fontsize=9.5, weight="bold", color=ec)
        draw_box(ax, x0 + 0.3, 2.5, 2.8, 0.9, top_lbl, fc=fc, ec=ec, fontsize=8.5, weight="bold")
        draw_box(ax, x0 + 0.3, 0.5, 2.8, 0.9, bot_lbl, fc=fc, ec=ec, fontsize=8.5, weight="bold")
        ax.annotate("", xy=(x0 + 2.4, 1.45), xytext=(x0 + 2.4, 2.45), arrowprops=dict(arrowstyle="->", color=UNB_BLUE, lw=1.8))
        ax.text(x0 + 2.5, 1.95, f1, fontsize=7.5, va="center", color=UNB_BLUE)
        ax.annotate("", xy=(x0 + 0.9, 2.45), xytext=(x0 + 0.9, 1.45), arrowprops=dict(arrowstyle="->", color=RED_ACCENT, lw=1.6, ls="--"))
        ax.text(x0 + 0.8, 1.95, f2, fontsize=7.5, va="center", ha="right", color=RED_ACCENT)

    fig.tight_layout()
    save_rgb_png(fig, "figuras/fig03_niveis_gemeo_digital.png")


def fig04():
    fig, ax = plt.subplots(figsize=(9, 3.8), dpi=FIGURE_DPI)
    ax.set_xlim(0, 12)
    ax.set_ylim(0, 4.2)
    ax.axis("off")

    draw_box(ax, 0.4, 1.8, 3.1, 1.7, "Registro k-1\npayload_{k-1}\nprev: H_{k-2}\nhash: H_{k-1}", fontsize=8.5, weight="bold")
    draw_box(ax, 4.4, 1.8, 3.1, 1.7, "Registro k\npayload_k\nprev: H_{k-1}\nhash: H_k", fontsize=8.5, weight="bold")
    draw_box(ax, 8.4, 1.8, 3.1, 1.7, "Registro k+1\npayload_{k+1}\nprev: H_k\nhash: H_{k+1}", fontsize=8.5, weight="bold")
    draw_box(ax, 7.9, 0.2, 3.8, 1.0, "Âncora Externa Assinada (WORM)\nReplica (k+1, H_{k+1}) fora do PostgreSQL", fc="#E6F4EA", ec=UNB_GREEN, fontsize=8.0, weight="bold")

    ax.annotate("", xy=(4.35, 2.65), xytext=(3.55, 2.65), arrowprops=dict(arrowstyle="->", color=UNB_BLUE, lw=2))
    ax.annotate("", xy=(8.35, 2.65), xytext=(7.55, 2.65), arrowprops=dict(arrowstyle="->", color=UNB_BLUE, lw=2))
    ax.annotate("", xy=(9.8, 1.25), xytext=(9.8, 1.75), arrowprops=dict(arrowstyle="->", color=UNB_GREEN, lw=2))

    fig.tight_layout()
    save_rgb_png(fig, "figuras/fig04_hash_chain.png")


def fig05():
    fig, ax = plt.subplots(figsize=(9.5, 4.5), dpi=FIGURE_DPI)
    ax.set_xlim(0, 13)
    ax.set_ylim(0, 5.2)
    ax.axis("off")

    draw_box(ax, 0.3, 0.4, 7.5, 4.4, "", fc="#F5F9FD", ec=UNB_BLUE)
    ax.text(0.5, 4.5, "Sub-rede Docker: core-net (Isolada da IA)", fontsize=9.5, weight="bold", color=UNB_BLUE)

    draw_box(ax, 8.2, 0.4, 4.5, 4.4, "", fc="#FEF6F6", ec=RED_ACCENT)
    ax.text(8.4, 4.5, "Sub-rede: ai-isolated-net", fontsize=9.5, weight="bold", color=RED_ACCENT)

    draw_box(ax, 3.2, 2.8, 3.8, 1.2, "Core PM Service (Go)\nRBAC / Governança Fail-Closed", fc="#FFFFFF", ec=UNB_BLUE, weight="bold")
    draw_box(ax, 0.7, 0.8, 3.0, 1.2, "PostgreSQL 16\nSKIP LOCKED + Audit", fc="#FFFFFF", ec=UNB_BLUE, weight="bold")
    draw_box(ax, 4.3, 0.8, 3.1, 1.2, "Simulation Worker\nCPM + Monte Carlo", fc="#FFFFFF", ec=UNB_BLUE, weight="bold")
    draw_box(ax, 8.7, 2.8, 3.5, 1.2, "AI Agent Service\n(Sem Driver/Rota SQL)", fc="#FFFFFF", ec=RED_ACCENT, weight="bold")
    draw_box(ax, 8.7, 0.8, 3.5, 1.2, "API LLM Externa\n(HTTPS / TLS)", fc="#F3F4F6", ec="#4B5563", weight="bold")

    ax.annotate("", xy=(2.2, 2.05), xytext=(4.2, 2.75), arrowprops=dict(arrowstyle="<->", color=UNB_BLUE, lw=1.8))
    ax.annotate("", xy=(5.8, 2.05), xytext=(2.8, 1.4), arrowprops=dict(arrowstyle="<->", color=UNB_BLUE, lw=1.8))
    ax.annotate("", xy=(8.65, 3.4), xytext=(7.05, 3.4), arrowprops=dict(arrowstyle="<->", color=RED_ACCENT, lw=1.8))
    ax.text(7.85, 3.65, "Allowlist JSON", ha="center", fontsize=7.5, color=RED_ACCENT, weight="bold")
    ax.annotate("", xy=(10.45, 2.05), xytext=(10.45, 2.75), arrowprops=dict(arrowstyle="<->", color="#4B5563", lw=1.8))

    fig.tight_layout()
    save_rgb_png(fig, "figuras/fig05_arquitetura.png")


def fig06():
    fig, ax = plt.subplots(figsize=(9.5, 4.5), dpi=FIGURE_DPI)
    ax.set_xlim(0, 12.5)
    ax.set_ylim(0, 5.2)
    ax.axis("off")

    actors = [
        (1.0, "GitHub"),
        (3.0, "Core PM"),
        (5.0, "PostgreSQL"),
        (7.0, "Sim Worker"),
        (9.2, "AI Service"),
        (11.4, "Gerente (UI)"),
    ]
    for x, name in actors:
        draw_box(ax, x - 0.8, 4.3, 1.6, 0.6, name, fontsize=8.5, weight="bold")
        ax.plot([x, x], [0.3, 4.3], ls="--", color="#9CA3AF", lw=1.2)

    steps = [
        (3.8, 1.0, 3.0, "1. Webhook assinado (HMAC)", UNB_BLUE),
        (3.35, 3.0, 5.0, "2. state_version++ e enfileira Job", UNB_BLUE),
        (2.9, 7.0, 5.0, "3. Consome via SKIP LOCKED", UNB_BLUE),
        (2.45, 7.0, 3.0, "4. Persiste P50, P85, P95 e IC_i", UNB_BLUE),
        (1.95, 3.0, 9.2, "5. Envia Allowlist sanitizada", RED_ACCENT),
        (1.5, 9.2, 3.0, "6. JSON validado (Fail-Closed)", RED_ACCENT),
        (1.0, 3.0, 11.4, "7. SSE: Fila Cat. 2 com ΔP85", UNB_GREEN),
        (0.5, 11.4, 5.0, "8. Aprovação humana -> SHA-256", UNB_BLUE),
    ]
    for y, x1, x2, lbl, col in steps:
        ax.annotate("", xy=(x2, y), xytext=(x1, y), arrowprops=dict(arrowstyle="->", color=col, lw=1.7))
        ax.text((x1 + x2) / 2, y + 0.1, lbl, ha="center", fontsize=7.8, color=col, weight="bold")

    fig.tight_layout()
    save_rgb_png(fig, "figuras/fig06_sequencia.png")


def fig07():
    rng = np.random.default_rng(42)
    n = 10000

    def sample_pert(to, tm, tp):
        diff = tp - to
        a = 1.0 + 4.0 * (tm - to) / diff
        b = 1.0 + 4.0 * (tp - tm) / diff
        return to + rng.beta(a, b, size=n) * diff

    t1 = sample_pert(2, 3, 5)
    t2 = sample_pert(4, 5, 9)
    t3 = sample_pert(3, 5, 10)
    t4 = sample_pert(1, 2, 4)
    r1 = (rng.random(n) < 0.30) * 2.0

    f_a = np.sort(t1 + np.maximum(t2, t3) + t4)
    f_b = np.sort(t1 + np.maximum(t2, t3 + r1) + t4)
    cdf = np.arange(1, n + 1) / n

    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(9.5, 4.0), dpi=FIGURE_DPI)
    ax1.hist(f_a, bins=40, density=True, alpha=0.55, color=UNB_BLUE, label="Cenário A (Sem risco)")
    ax1.hist(f_b, bins=40, density=True, alpha=0.45, color=RED_ACCENT, label="Cenário B (Com R1)")
    ax1.axvline(10.8333, color=UNB_GREEN, ls="--", lw=1.8, label="CPM = 10,83 d.u.")
    ax1.set_xlabel("Prazo de término F (dias úteis)")
    ax1.set_ylabel("Densidade empírica")
    ax1.grid(True, ls="--", alpha=0.3)
    ax1.legend(fontsize=8)

    ax2.plot(f_a, cdf, color=UNB_BLUE, lw=2.2, label="CDF Cenário A (P50=11,39 | P85=12,75)")
    ax2.plot(f_b, cdf, color=RED_ACCENT, lw=2.2, label="CDF Cenário B (P50=11,79 | P85=13,43)")
    ax2.axvline(10.8333, color=UNB_GREEN, ls="--", lw=1.8, label="CPM = 10,83")
    ax2.axhline(0.50, color="gray", ls=":", lw=1)
    ax2.axhline(0.85, color="gray", ls=":", lw=1)
    ax2.set_xlabel("Prazo de término F (dias úteis)")
    ax2.set_ylabel("Probabilidade acumulada P(F ≤ t)")
    ax2.grid(True, ls="--", alpha=0.3)
    ax2.legend(fontsize=8, loc="lower right")

    fig.tight_layout()
    save_rgb_png(fig, "figuras/fig07_histograma_cdf_poc.png")


def fig08():
    tasks = [
        ("E01 [Trilha B]: Extrator GitHub + Replay", 1, 2, UNB_GREEN),
        ("E02 [Trilha A]: Contrato + SQL + Âncora", 1, 2, UNB_BLUE),
        ("E03 [Trilha A]: CRUD + Paginação + RBAC", 3, 4, UNB_BLUE),
        ("E04 [Trilha B]: Ingestão + Outbox + DLQ", 3, 4, UNB_GREEN),
        ("E05 [Trilha A]: Motor Monte Carlo + CRN", 5, 6, UNB_BLUE),
        ("E06 [Trilha A]: Sim Worker + Persistência", 7, 8, UNB_BLUE),
        ("E08 [Trilha B]: Dim. 2 (10 Regimes, M=200)", 7, 8, UNB_GREEN),
        ("E07 [Trilha A]: Propostas + Guardrails", 9, 10, UNB_BLUE),
        ("E09 [Trilha B]: Dim. 3 (Backtesting Real)", 9, 11, UNB_GREEN),
        ("E11 [Trilha A]: UI Web Mínima Anti-Viés", 11, 12, UNB_BLUE),
        ("E10 [Trilha B]: Dim. 5 (300 Cenários + Abl.)", 12, 14, UNB_GREEN),
        ("E12 [Conjunto]: Consolidação Estatística", 15, 16, ORANGE_ACCENT),
        ("E13 [Conjunto]: Redação Final + Ameaças", 17, 18, ORANGE_ACCENT),
        ("E14 [Conjunto]: Revisão ABNT + Defesa + Buffer", 19, 20, ORANGE_ACCENT),
        ("D01 [Desejável]: Grafo Visual UI", 13, 14, "#6B7280"),
        ("D02 [Desejável]: Painel de Sensibilidade", 15, 16, "#6B7280"),
        ("D03 [Desejável]: Juízes Humanos Adicionais", 16, 17, "#6B7280"),
    ]
    fig, ax = plt.subplots(figsize=(9.5, 5.2), dpi=FIGURE_DPI)
    y_pos = np.arange(len(tasks))[::-1]
    for idx, (label, start, end, col) in enumerate(tasks):
        y = y_pos[idx]
        ax.barh(y, end - start + 0.8, left=start - 0.4, height=0.6, color=col, alpha=0.85, edgecolor="#222222")

    ax.axhline(2.5, color=RED_ACCENT, ls="--", lw=1.6)
    ax.text(0.8, 2.65, "Linha de Corte (Abaixo: D01–D03) | Monte Carlo c/ Riscos: P50=17,29 | P85=18,30 | P(F≤20)=99,6%", color=RED_ACCENT, fontsize=8.0, weight="bold")
    ax.set_yticks(y_pos)
    ax.set_yticklabels([t[0] for t in tasks], fontsize=8)
    ax.set_xlim(0.5, 20.5)
    ax.set_xticks(range(1, 21))
    ax.set_xlabel("Semanas Úteis do Cronograma do TCC 2 (Trilha A: Azul | Trilha B: Verde | Conjunto: Laranja)")
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
    print("8 figuras geradas com sucesso em ./figuras/")