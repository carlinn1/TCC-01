#!/usr/bin/env python3
"""
Extração do dataset de backtesting (TCC 1) a partir da API pública do GitHub.

Definição operacional (documentada na Seção 5 do texto):
  - "Janela de release" = intervalo entre a publicação da release anterior
    e a publicação da release-alvo (timestamps lidos das páginas de release).
  - Itens de trabalho = Pull Requests MESCLADOS (merged) dentro da janela,
    excluindo autores do tipo Bot (ex.: dependabot); contagem com bots é
    reportada à parte.
  - Cycle time de um PR = merged_at - created_at, em dias úteis fracionários
    (seg-sex, sem feriados).
  - Issues concluídas = issues fechadas com motivo "completed" na janela.
  - Vínculo PR->issue (aproximação): PR cujo corpo contém palavra-chave de
    fechamento (close/closes/closed/fix/fixes/fixed/resolve/resolves/resolved #N).

Uso:   python3 extract_dataset.py            (60 req/h sem token)
       GITHUB_TOKEN=... python3 extract_dataset.py   (limite maior)
Saída: dataset_backtesting.json
"""
import json, os, re, sys, time, urllib.request, urllib.error
from datetime import datetime, timedelta, timezone

REPO = "cli/cli"
# Timestamps de publicação lidos de https://github.com/cli/cli/releases/tag/<tag>
RELEASES = {
    "v2.39.2": "2023-11-27T18:04:17Z",
    "v2.40.0": "2023-12-07T15:06:02Z",
    "v2.41.0": "2024-01-08T14:16:24Z",
}
CASES = [("v2.39.2", "v2.40.0"), ("v2.40.0", "v2.41.0")]  # (anterior, alvo)
CLOSING = re.compile(r"\b(close[sd]?|fix(e[sd])?|resolve[sd]?)\b[:\s]+(#\d+|https://github\.com/[^\s]+/issues/\d+)", re.I)


def ts(s):
    return datetime.strptime(s, "%Y-%m-%dT%H:%M:%SZ").replace(tzinfo=timezone.utc)


def api(url):
    req = urllib.request.Request(url, headers={"Accept": "application/vnd.github+json", "User-Agent": "tcc1-backtesting"})
    tok = os.environ.get("GITHUB_TOKEN")
    if tok:
        req.add_header("Authorization", f"Bearer {tok}")
    for attempt in range(6):
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
        url = "https://api.github.com/search/issues?q=" + urllib.request.quote(q) + f"&per_page=100&page={page}"
        d = api(url)
        items += d["items"]
        if len(items) >= d["total_count"] or not d["items"] or page >= 10:
            return items, d["total_count"]
        page += 1
        time.sleep(7)


def business_days(t1, t2):
    """Dias úteis fracionários entre t1 e t2 (seg-sex)."""
    total, cur = 0.0, t1
    while cur < t2:
        nxt = min((cur + timedelta(days=1)).replace(hour=0, minute=0, second=0, microsecond=0), t2)
        if cur.weekday() < 5:
            total += (nxt - cur).total_seconds() / 86400.0
        cur = nxt
    return total


def pct(xs, q):
    xs = sorted(xs)
    if not xs:
        return None
    k = (len(xs) - 1) * q
    f = int(k)
    c = min(f + 1, len(xs) - 1)
    return xs[f] + (xs[c] - xs[f]) * (k - f)


def main():
    out = {"repo": REPO, "cases": []}
    for prev, tgt in CASES:
        a, b = RELEASES[prev], RELEASES[tgt]
        rng = f"{a}..{b}"
        print(f"== {tgt}: janela {a} -> {b}")
        prs, n_prs = search(f"repo:{REPO} is:pr is:merged merged:{rng}")
        time.sleep(7)
        iss, n_iss = search(f"repo:{REPO} is:issue is:closed reason:completed closed:{rng}")
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
            "target": tgt, "previous": prev, "window_start": a, "window_end": b,
            "window_business_days": round(business_days(ts(a), ts(b)), 2),
            "prs_merged_total": n_prs, "prs_merged_non_bot": len(human),
            "prs_with_closing_keyword": linked,
            "issues_closed_completed": n_iss,
            "cycle_time_business_days": {
                "n": len(cts), "p5": pct(cts, .05), "p10": pct(cts, .10), "p50": pct(cts, .50),
                "p90": pct(cts, .90), "p95": pct(cts, .95), "mean": sum(cts) / len(cts) if cts else None},
            "top_labels": sorted(labels.items(), key=lambda kv: -kv[1])[:8],
        }
        out["cases"].append(case)
        print(json.dumps(case, indent=2, ensure_ascii=False))
    json.dump(out, open("dataset_backtesting.json", "w"), indent=2, ensure_ascii=False)


if __name__ == "__main__":
    main()
