#!/usr/bin/env python3
"""
Extração auditável do dataset de backtesting (TCC 1 / preparação TCC 2)
a partir da API pública do GitHub (repositório cli/cli).

Correções implementadas:
  1. Inclui explicitamente a janela ampliada v2.40.0 -> v2.46.0 em RELEASES e CASES,
     garantindo correspondência exata com dataset_backtesting.json.
  2. Separa explicitamente issues fechadas com reason:completed das fechadas com
     reason:not_planned e PRs fechados sem merge (is:unmerged), alinhado à Tabela 6.
  3. Inclui função auxiliar de coleta do primeiro commit do PR (first_commit_at)
     para mitigar a ameaça à validade de construto no TCC 2.

Uso:   python3 extract_dataset.py
       GITHUB_TOKEN=... python3 extract_dataset.py
Saída: dataset_backtesting.json
"""
import argparse, json, os, re, sys, time, urllib.request, urllib.error
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


def ts(s):
    return datetime.strptime(s, "%Y-%m-%dT%H:%M:%SZ").replace(tzinfo=timezone.utc)


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
    """Dias úteis fracionários entre t1 e t2 (seg-sex)."""
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


def fetch_first_commit_at(repo, pr_number):
    """Retorna o timestamp ISO-8601 do primeiro commit associado ao PR (validade de construto TCC 2)."""
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


def main():
    parser = argparse.ArgumentParser(description="Extração reprodutível do dataset de backtesting (cli/cli).")
    parser.add_argument("--repo", default=REPO, help="Repositório GitHub (padrão: cli/cli)")
    parser.add_argument("--out", default="dataset_backtesting.json", help="Arquivo JSON de saída")
    args = parser.parse_args()

    out = {"repo": args.repo, "cases": []}
    for prev, tgt, label_target in CASES:
        a, b = RELEASES[prev], RELEASES[tgt]
        rng = f"{a}..{b}"
        print(f"== {label_target}: janela {a} -> {b}")
        prs, n_prs = search(f"repo:{args.repo} is:pr is:merged merged:{rng}")
        time.sleep(7)
        _, n_prs_unmerged = search(f"repo:{args.repo} is:pr is:closed is:unmerged closed:{rng}")
        time.sleep(7)
        iss_comp, n_iss_comp = search(
            f"repo:{args.repo} is:issue is:closed reason:completed closed:{rng}"
        )
        time.sleep(7)
        _, n_iss_not_planned = search(
            f"repo:{args.repo} is:issue is:closed reason:not_planned closed:{rng}"
        )
        time.sleep(7)

        human = [p for p in prs if p["user"]["type"] != "Bot"]
        cts = []
        linked = 0
        for p in human:
            merged = p["pull_request"]["merged_at"]
            cts.append(business_days(ts(p["created_at"]), ts(merged)))
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
            "issues_closed_completed": n_iss_comp,
            "issues_closed_not_planned_excluded": n_iss_not_planned,
            "cycle_time_business_days": {
                "n": len(cts),
                "p5": pct(cts, 0.05),
                "p10": pct(cts, 0.10),
                "p50": pct(cts, 0.50),
                "p90": pct(cts, 0.90),
                "p95": pct(cts, 0.95),
                "mean": sum(cts) / len(cts) if cts else None,
            },
            "top_labels": sorted(labels.items(), key=lambda kv: -kv[1])[:8],
        }
        out["cases"].append(case)
        print(json.dumps(case, indent=2, ensure_ascii=False))

    with open(args.out, "w", encoding="utf-8") as f:
        json.dump(out, f, indent=2, ensure_ascii=False)


if __name__ == "__main__":
    main()