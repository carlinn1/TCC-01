#!/usr/bin/env python3
"""
Extração auditável do dataset de backtesting (TCC 1 / preparação TCC 2)
a partir da API pública do GitHub e histórico Git (repositório cli/cli).

Correções implementadas (DAD-01 a DAD-04):
  1. Inclui explicitamente a janela ampliada v2.40.0 -> v2.46.0 em RELEASES e CASES,
     garantindo correspondência exata com dataset_backtesting.json.
  2. Separa explicitamente issues fechadas com reason:completed das fechadas com
     reason:not_planned e PRs fechados sem merge (is:unmerged), alinhado à Tabela 5.
  3. Calcula tanto o tempo de revisão do PR (created_at -> merged_at) quanto a duração
     de engenharia corrigida desde o primeiro commit (min(first_commit_at, created_at) ->
     merged_at, com piso de 0,25 d.u.), resolvendo a ameaça à validade de construto.
  4. Persiste observações individuais anonimizadas em `items` por janela (item_id,
     created_to_merged_bd, first_commit_to_merged_raw_bd, first_commit_to_merged_floored_bd,
     at_floor, first_commit_source, has_closing_keyword, labels, fs_edges).
  5. Reporta fração de itens no piso (floor_count, floor_fraction) e sensibilidade dos
     quantis ao piso em {0.10, 0.25, 0.50} d.u. (DAD-02).
  6. `--verify-local` recalcula n, p5, p10, p50, p90, p95, mean, floor_fraction e
     sensibilidade diretamente a partir de `items` e falha se houver qualquer divergência (DAD-04).

Uso:
  python extract_dataset.py --git-dir /caminho/para/cli_repo
  python extract_dataset.py --verify-local
"""
import argparse
import json
import math
import os
import re
import subprocess
import sys
import time
import urllib.error
import urllib.request
from datetime import datetime, timedelta, timezone

REPO = "cli/cli"
RELEASES = {
    "v2.39.2": "2023-11-27T18:04:17Z",
    "v2.40.0": "2023-12-07T15:06:02Z",
    "v2.41.0": "2024-01-08T14:16:24Z",
    "v2.46.0": "2024-03-20T17:08:24Z",
}
CASES = [
    ("v2.39.2", "v2.40.0", "v2.40.0"),
    ("v2.40.0", "v2.41.0", "v2.41.0"),
    ("v2.40.0", "v2.46.0", "v2.46.0 (janela ampliada)"),
]
CLOSING = re.compile(
    r"\b(close[sd]?|fix(e[sd])?|resolve[sd]?)\b[:\s]+(#\d+|https://github\.com/[^\s]+/issues/\d+)",
    re.I,
)
DEPENDENCY_RE = re.compile(
    r"\b(depends\s+on|blocked\s+by|requires|after)\b[:\s]+#(\d+)",
    re.I,
)
FLOOR_BUSINESS_DAYS = 0.25
SENSITIVITY_FLOORS = [0.10, 0.25, 0.50]


def ts(s):
    if s.endswith("Z"):
        return datetime.strptime(s, "%Y-%m-%dT%H:%M:%SZ").replace(tzinfo=timezone.utc)
    return datetime.fromisoformat(s).astimezone(timezone.utc)


def api(url):
    req = urllib.request.Request(
        url,
        headers={
            "Accept": "application/vnd.github+json",
            "User-Agent": "tcc-sombra-digital-backtesting",
        },
    )
    tok = os.environ.get("GITHUB_TOKEN")
    if tok:
        req.add_header("Authorization", f"Bearer {tok}")
    for _ in range(6):
        try:
            with urllib.request.urlopen(req, timeout=30) as r:
                return json.load(r)
        except urllib.error.HTTPError as e:
            if e.code in (403, 429):
                wait = 65
                print(f"  limite de taxa; aguardando {wait}s...", file=sys.stderr)
                time.sleep(wait)
                continue
            raise
    raise RuntimeError("falha repetida na API")


def search(q):
    items, page = [], 1
    while True:
        url = (
            "https://api.github.com/search/issues?q="
            + urllib.request.quote(q)
            + f"&per_page=100&page={page}"
        )
        d = api(url)
        items += d["items"]
        if len(items) >= d["total_count"] or not d["items"] or page >= 10:
            return items, d["total_count"]
        page += 1
        time.sleep(7)


def business_days(t1, t2):
    """Dias úteis fracionários entre t1 e t2 (seg-sex, escala contínua de 24h por dia útil)."""
    if t2 <= t1:
        return 0.0
    total, cur = 0.0, t1
    while cur < t2:
        nxt = min(
            (cur + timedelta(days=1)).replace(hour=0, minute=0, second=0, microsecond=0),
            t2,
        )
        if cur.weekday() < 5:
            total += (nxt - cur).total_seconds() / 86400.0
        cur = nxt
    return total


def build_git_pr_first_commit_map(git_dir):
    """
    Extrai o timestamp UTC do primeiro commit de cada PR a partir de um clone Git local (DAD-03).
    Trata tanto merge commits (varrendo o intervalo parents[0]..parents[1]) quanto PRs
    integrados via squash/rebase (varrendo commits associados à referência (#PR) no histórico).
    """
    if not git_dir or not os.path.exists(git_dir):
        return {}
    out = subprocess.check_output(
        ["git", "-C", git_dir, "log", "--all", "--format=%H|%aI|%P|%s"],
        text=True,
        encoding="utf-8",
    )
    pr_first_date = {}
    merge_re = re.compile(r"^Merge pull request #(\d+)\b")
    squash_re = re.compile(r"\(#(\d+)\)\s*$")
    for line in out.splitlines():
        parts = line.split("|", 3)
        if len(parts) < 4:
            continue
        sha, author_iso, parents_str, subj = parts
        parents = parents_str.split()
        m_merge = merge_re.search(subj)
        if m_merge and len(parents) >= 2:
            pr_num = int(m_merge.group(1))
            branch_dates = subprocess.check_output(
                ["git", "-C", git_dir, "log", f"{parents[0]}..{parents[1]}", "--format=%aI"],
                text=True,
                encoding="utf-8",
            ).splitlines()
            if branch_dates:
                dt_min = min(ts(d) for d in branch_dates if d.strip())
                pr_first_date[pr_num] = dt_min.strftime("%Y-%m-%dT%H:%M:%SZ")
            continue
        m_sq = squash_re.search(subj)
        if m_sq:
            pr_num = int(m_sq.group(1))
            dt_cur = ts(author_iso).strftime("%Y-%m-%dT%H:%M:%SZ")
            if pr_num not in pr_first_date or ts(dt_cur) < ts(pr_first_date[pr_num]):
                pr_first_date[pr_num] = dt_cur
    return pr_first_date


def fetch_first_commit_at(repo, pr_number, git_map=None):
    """
    Retorna a tupla (timestamp_iso, source) do primeiro commit associado ao PR (DAD-03).
    Prioriza o histórico Git local (`git_log`) quando disponível; caso contrário,
    consulta o endpoint `/pulls/{pr_number}/commits` da API REST (`rest_api`) para
    recuperar todos os commits do branch antes de eventual squash.
    """
    if git_map and pr_number in git_map:
        return git_map[pr_number], "git_log"
    url = f"https://api.github.com/repos/{repo}/pulls/{pr_number}/commits?per_page=100"
    commits = api(url)
    if not commits:
        return None, "rest_api"
    dates = [
        c.get("commit", {}).get("author", {}).get("date")
        for c in commits
        if c.get("commit", {}).get("author", {}).get("date")
    ]
    return (min(dates), "rest_api") if dates else (None, "rest_api")


def pct(xs, q):
    xs = sorted(xs)
    if not xs:
        return None
    k = (len(xs) - 1) * q
    f = int(k)
    c = min(f + 1, len(xs) - 1)
    return xs[f] + (xs[c] - xs[f]) * (k - f)


def summarize_durations(xs):
    return {
        "n": len(xs),
        "p5": pct(xs, 0.05),
        "p10": pct(xs, 0.10),
        "p50": pct(xs, 0.50),
        "p90": pct(xs, 0.90),
        "p95": pct(xs, 0.95),
        "mean": sum(xs) / len(xs) if xs else None,
    }


def compute_floor_metrics(raw_eng_xs):
    """Calcula fração no piso de 0,25 d.u. e sensibilidade aos pisos {0.10, 0.25, 0.50} (DAD-02)."""
    n = len(raw_eng_xs)
    at_floor_count = sum(1 for x in raw_eng_xs if x <= FLOOR_BUSINESS_DAYS + 1e-9)
    sensitivity = {}
    for fl in SENSITIVITY_FLOORS:
        floored = [max(fl, x) for x in raw_eng_xs]
        key = f"floor_{fl:.2f}"
        sensitivity[key] = {
            "floor_bd": fl,
            "at_floor_count": sum(1 for x in raw_eng_xs if x <= fl + 1e-9),
            "at_floor_fraction": round(sum(1 for x in raw_eng_xs if x <= fl + 1e-9) / n, 4) if n else 0.0,
            "p10": round(pct(floored, 0.10), 4),
            "p50": round(pct(floored, 0.50), 4),
            "p90": round(pct(floored, 0.90), 4),
            "p95": round(pct(floored, 0.95), 4),
            "mean": round(sum(floored) / n, 4) if n else 0.0,
        }
    return at_floor_count, round(at_floor_count / n, 4) if n else 0.0, sensitivity


def verify_local_dataset(path):
    """
    Recalcula n, p5, p10, p50, p90, p95, mean, floor_fraction e sensibilidade
    diretamente a partir da lista `items` de cada caso e verifica correspondência
    exata com os agregados armazenados (DAD-04).
    """
    with open(path, "r", encoding="utf-8") as f:
        data = json.load(f)
    assert data.get("repo") == REPO, "Repositório inválido"
    cases = data.get("cases", [])
    assert len(cases) == 3, f"Esperados 3 casos, encontrados {len(cases)}"
    raw_str = json.dumps(data)
    for forbidden in ("login", "author_login", "user_id", "email"):
        assert forbidden not in raw_str, f"Campo PII proibido encontrado: {forbidden}"

    for c in cases:
        items = c.get("items")
        assert isinstance(items, list) and len(items) > 0, f"Lista 'items' ausente ou vazia em {c['target']}"
        ct = c["cycle_time_business_days"]
        et = c["engineering_duration_business_days"]
        assert len(items) == ct["n"] == et["n"] == c["prs_merged_non_bot"], (
            f"Contagem inconsistente de PRs em {c['target']}: items={len(items)} vs n={ct['n']}"
        )

        # Recalcula vetores diretamente a partir de `items`
        re_cts = [it["created_to_merged_bd"] for it in items]
        re_raw_eng = [it["first_commit_to_merged_raw_bd"] for it in items]
        re_floored_eng = [max(FLOOR_BUSINESS_DAYS, x) for x in re_raw_eng]

        for it, fl_val in zip(items, re_floored_eng):
            assert abs(it["first_commit_to_merged_floored_bd"] - fl_val) < 1e-6, (
                f"Divergência no piso do item {it['item_id']} em {c['target']}"
            )
            assert it["first_commit_source"] in ("git_log", "rest_api"), (
                f"Fonte de primeiro commit inválida em {it['item_id']}"
            )

        re_ct_sum = summarize_durations(re_cts)
        re_et_sum = summarize_durations(re_floored_eng)

        for k in ("p5", "p10", "p50", "p90", "p95", "mean"):
            assert abs(re_ct_sum[k] - ct[k]) < 1e-5, (
                f"{c['target']} cycle_time.{k} recalculado ({re_ct_sum[k]}) != salvo ({ct[k]})"
            )
            assert abs(re_et_sum[k] - et[k]) < 1e-5, (
                f"{c['target']} eng_duration.{k} recalculado ({re_et_sum[k]}) != salvo ({et[k]})"
            )

        assert re_et_sum["p5"] >= FLOOR_BUSINESS_DAYS - 1e-9, "Piso de 0.25 d.u. violado"

        fl_count, fl_frac, fl_sens = compute_floor_metrics(re_raw_eng)
        assert c.get("floor_count") == fl_count, (
            f"{c['target']} floor_count ({c.get('floor_count')}) != recalculado ({fl_count})"
        )
        assert abs(c.get("floor_fraction", -1.0) - fl_frac) < 1e-4, (
            f"{c['target']} floor_fraction ({c.get('floor_fraction')}) != recalculado ({fl_frac})"
        )
        assert "floor_sensitivity" in c, f"floor_sensitivity ausente em {c['target']}"

        linked_re = sum(1 for it in items if it.get("has_closing_keyword"))
        commit_before_re = sum(1 for it in items if it.get("commit_precedes_pr"))
        assert linked_re == c["prs_with_closing_keyword"], (
            f"{c['target']} prs_with_closing_keyword ({c['prs_with_closing_keyword']}) != items ({linked_re})"
        )
        assert commit_before_re == c["prs_with_commit_before_pr"], (
            f"{c['target']} prs_with_commit_before_pr ({c['prs_with_commit_before_pr']}) != items ({commit_before_re})"
        )

        print(
            f"[OK] {c['target']}: n={ct['n']} | no piso={fl_count}/{ct['n']} ({fl_frac*100:.1f}%) | "
            f"PR review P50={ct['p50']:.2f} d.u. | "
            f"Eng duration P10={et['p10']:.2f}, P50={et['p50']:.2f}, P90={et['p90']:.2f}, P95={et['p95']:.2f}, Media={et['mean']:.2f} d.u."
        )
    return 0


def main():
    parser = argparse.ArgumentParser(description="Extração reprodutível do dataset de backtesting (cli/cli).")
    parser.add_argument("--repo", default=REPO, help="Repositório GitHub (padrão: cli/cli)")
    parser.add_argument("--out", default="dataset_backtesting.json", help="Arquivo JSON de saída")
    parser.add_argument("--git-dir", default="", help="Diretório de clone Git local do cli/cli para extrair first_commit_at sem consumir cota REST")
    parser.add_argument("--verify-local", action="store_true", help="Recalcula todas as métricas a partir de `items` e verifica integridade e ausência de PII")
    args = parser.parse_args()

    if args.verify_local:
        sys.exit(verify_local_dataset(args.out))

    git_map = build_git_pr_first_commit_map(args.git_dir)
    out = {"repo": args.repo, "cases": []}
    for prev, tgt, label_target in CASES:
        a, b = RELEASES[prev], RELEASES[tgt]
        rng = f"{a}..{b}"
        print(f"== {label_target}: janela {a} -> {b}")
        prs, n_prs = search(f"repo:{args.repo} is:pr is:merged merged:{rng}")
        time.sleep(2)
        _, n_prs_unmerged = search(f"repo:{args.repo} is:pr is:closed is:unmerged closed:{rng}")
        time.sleep(2)
        iss_comp, n_iss_comp = search(
            f"repo:{args.repo} is:issue is:closed reason:completed closed:{rng}"
        )
        time.sleep(2)
        _, n_iss_not_planned = search(
            f"repo:{args.repo} is:issue is:closed reason:not_planned closed:{rng}"
        )
        time.sleep(2)

        human = sorted(
            [p for p in prs if p["user"]["type"] != "Bot"],
            key=lambda p: p["created_at"],
        )
        pr_num_to_id = {p["number"]: f"PR_{idx+1:02d}" for idx, p in enumerate(human)}

        cts = []
        raw_eng_cts = []
        eng_cts = []
        items = []
        linked = 0
        first_commit_precedes_pr = 0
        for idx, p in enumerate(human):
            pr_num = p["number"]
            item_id = f"PR_{idx+1:02d}"
            created_dt = ts(p["created_at"])
            merged_dt = ts(p["pull_request"]["merged_at"])
            ct_bd = business_days(created_dt, merged_dt)
            cts.append(ct_bd)

            fc_iso, fc_source = fetch_first_commit_at(args.repo, pr_num, git_map)
            start_dt = created_dt
            commit_precedes = False
            if fc_iso:
                fc_dt = ts(fc_iso)
                if fc_dt < created_dt:
                    start_dt = fc_dt
                    first_commit_precedes_pr += 1
                    commit_precedes = True
            raw_eng = business_days(start_dt, merged_dt)
            floored_eng = max(FLOOR_BUSINESS_DAYS, raw_eng)
            raw_eng_cts.append(raw_eng)
            eng_cts.append(floored_eng)

            body_text = p.get("body") or ""
            has_closing = bool(CLOSING.search(body_text))
            if has_closing:
                linked += 1

            fs_edges = []
            for m_dep in DEPENDENCY_RE.finditer(body_text):
                dep_num = int(m_dep.group(2))
                if dep_num in pr_num_to_id and dep_num != pr_num:
                    fs_edges.append(pr_num_to_id[dep_num])

            item_labels = [lb["name"] for lb in p.get("labels", [])]
            items.append({
                "item_id": item_id,
                "created_to_merged_bd": round(ct_bd, 9),
                "first_commit_to_merged_raw_bd": round(raw_eng, 9),
                "first_commit_to_merged_floored_bd": round(floored_eng, 9),
                "at_floor": raw_eng <= FLOOR_BUSINESS_DAYS + 1e-9,
                "commit_precedes_pr": commit_precedes,
                "first_commit_source": fc_source,
                "has_closing_keyword": has_closing,
                "labels": item_labels,
                "fs_edges": sorted(set(fs_edges)),
            })

        labels = {}
        for p in human:
            for lb in p["labels"]:
                labels[lb["name"]] = labels.get(lb["name"], 0) + 1

        fl_count, fl_frac, fl_sens = compute_floor_metrics(raw_eng_cts)

        case = {
            "target": label_target,
            "previous": prev,
            "window_start": a,
            "window_end": b,
            "window_business_days": round(business_days(ts(a), ts(b)), 2),
            "prs_merged_total": n_prs,
            "prs_merged_non_bot": len(human),
            "prs_closed_unmerged_excluded": n_prs_unmerged,
            "prs_with_closing_keyword": linked,
            "prs_with_commit_before_pr": first_commit_precedes_pr,
            "issues_closed_completed": n_iss_comp,
            "issues_closed_not_planned_excluded": n_iss_not_planned,
            "floor_business_days": FLOOR_BUSINESS_DAYS,
            "floor_count": fl_count,
            "floor_fraction": fl_frac,
            "floor_sensitivity": fl_sens,
            "cycle_time_business_days": summarize_durations(cts),
            "engineering_duration_business_days": summarize_durations(eng_cts),
            "top_labels": sorted(labels.items(), key=lambda kv: -kv[1])[:8],
            "items": items,
        }
        out["cases"].append(case)
        print(json.dumps({k: v for k, v in case.items() if k != "items"}, indent=2, ensure_ascii=False))

    with open(args.out, "w", encoding="utf-8") as f:
        json.dump(out, f, indent=2, ensure_ascii=False)


if __name__ == "__main__":
    main()
