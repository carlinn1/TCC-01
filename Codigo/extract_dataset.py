#!/usr/bin/env python3
"""
Extração auditável do dataset de backtesting (TCC 1 / preparação TCC 2)
a partir da API pública do GitHub e histórico Git (repositório cli/cli).

Correções implementadas:
  1. Inclui explicitamente a janela ampliada v2.40.0 -> v2.46.0 em RELEASES e CASES,
     garantindo correspondência exata com dataset_backtesting.json.
  2. Separa explicitamente issues fechadas com reason:completed das fechadas com
     reason:not_planned e PRs fechados sem merge (is:unmerged), alinhado à Tabela 5.
  3. Calcula tanto o tempo de revisão do PR (created_at -> merged_at) quanto a duração
     de engenharia corrigida desde o primeiro commit (min(first_commit_at, created_at) ->
     merged_at, com piso de 0,25 d.u.), resolvendo a ameaça à validade de construto.
  4. Pseudonimização integral: nenhum login ou identificador pessoal (PII) é persistido.

Uso:
  python3 extract_dataset.py --git-dir /caminho/para/cli_repo
  python3 extract_dataset.py --verify-local
"""
import argparse, json, os, re, subprocess, sys, time, urllib.request, urllib.error
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
FLOOR_BUSINESS_DAYS = 0.25


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
    """Extrai o timestamp UTC do primeiro commit de cada PR a partir de um clone Git local."""
    if not git_dir or not os.path.exists(git_dir):
        return {}
    out = subprocess.check_output(
        ["git", "-C", git_dir, "log", "--format=%H|%aI|%P|%s"],
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
    """Retorna o timestamp ISO-8601 do primeiro commit associado ao PR."""
    if git_map and pr_number in git_map:
        return git_map[pr_number]
    url = f"https://api.github.com/repos/{repo}/pulls/{pr_number}/commits?per_page=100"
    commits = api(url)
    if not commits:
        return None
    dates = [
        c.get("commit", {}).get("author", {}).get("date")
        for c in commits
        if c.get("commit", {}).get("author", {}).get("date")
    ]
    return min(dates) if dates else None


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


def verify_local_dataset(path):
    with open(path, "r", encoding="utf-8") as f:
        data = json.load(f)
    assert data.get("repo") == REPO, "Repositório inválido"
    cases = data.get("cases", [])
    assert len(cases) == 3, f"Esperados 3 casos, encontrados {len(cases)}"
    raw_str = json.dumps(data)
    for forbidden in ("login", "author_login", "user_id", "email"):
        assert forbidden not in raw_str, f"Campo PII proibido encontrado: {forbidden}"
    for c in cases:
        ct = c["cycle_time_business_days"]
        et = c["engineering_duration_business_days"]
        assert ct["n"] == et["n"] == c["prs_merged_non_bot"], "Contagem inconsistente de PRs"
        assert et["p5"] >= FLOOR_BUSINESS_DAYS - 1e-9, "Piso de 0.25 d.u. violado"
        print(
            f"[OK] {c['target']}: n={ct['n']} | PR review P50={ct['p50']:.2f} d.u. | "
            f"Eng duration (1st commit) P10={et['p10']:.2f}, P50={et['p50']:.2f}, P90={et['p90']:.2f}, Media={et['mean']:.2f} d.u."
        )
    return 0


def main():
    parser = argparse.ArgumentParser(description="Extração reprodutível do dataset de backtesting (cli/cli).")
    parser.add_argument("--repo", default=REPO, help="Repositório GitHub (padrão: cli/cli)")
    parser.add_argument("--out", default="dataset_backtesting.json", help="Arquivo JSON de saída")
    parser.add_argument("--git-dir", default="", help="Diretório de clone Git local do cli/cli para extrair first_commit_at sem consumir cota REST")
    parser.add_argument("--verify-local", action="store_true", help="Apenas verifica a integridade e ausência de PII do arquivo JSON local")
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

        human = [p for p in prs if p["user"]["type"] != "Bot"]
        cts = []
        eng_cts = []
        linked = 0
        first_commit_Precedes_pr = 0
        for p in human:
            pr_num = p["number"]
            created_dt = ts(p["created_at"])
            merged_dt = ts(p["pull_request"]["merged_at"])
            cts.append(business_days(created_dt, merged_dt))

            fc_iso = fetch_first_commit_at(args.repo, pr_num, git_map)
            start_dt = created_dt
            if fc_iso:
                fc_dt = ts(fc_iso)
                if fc_dt < created_dt:
                    start_dt = fc_dt
                    first_commit_Precedes_pr += 1
            raw_eng = business_days(start_dt, merged_dt)
            eng_cts.append(max(FLOOR_BUSINESS_DAYS, raw_eng))

            if CLOSING.search(p.get("body") or ""):
                linked += 1

        labels = {}
        for p in human:
            for lb in p["labels"]:
                labels[lb["name"]] = labels.get(lb["name"], 0) + 1

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
            "prs_with_commit_before_pr": first_commit_Precedes_pr,
            "issues_closed_completed": n_iss_comp,
            "issues_closed_not_planned_excluded": n_iss_not_planned,
            "cycle_time_business_days": summarize_durations(cts),
            "engineering_duration_business_days": summarize_durations(eng_cts),
            "top_labels": sorted(labels.items(), key=lambda kv: -kv[1])[:8],
        }
        out["cases"].append(case)
        print(json.dumps(case, indent=2, ensure_ascii=False))

    with open(args.out, "w", encoding="utf-8") as f:
        json.dump(out, f, indent=2, ensure_ascii=False)


if __name__ == "__main__":
    main()
