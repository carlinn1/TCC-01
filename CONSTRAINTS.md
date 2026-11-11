# Constraints

Last reviewed: 2026-11-11 by @carlinn1

## Floor (always enforced, no setup required)

- No new suppression comments or disabled checks
- No unimplemented stubs or silent error swallowing
- No skipped or deleted tests in `Codigo/*_test.go`
- No secrets, tokens, contributor logins, or PII in source or `Codigo/dataset_backtesting.json`
- Zero revision metatext ("Em resposta direta...", "vetores sutis apontados...", "Para fechar qualquer lacuna...") or promotional superlatives ("contenção comprovada", "degrau arquitetural seguro", "livre de espantalhos") in `Docs/Main.tex`
- This file does not get weakened to make a change pass

## Enforced with numbers

| Dimension | Rule | Checked by | Runs at |
|-----------|------|-----------|---------|
| Go Unit & Statistical Suite | 100% pass, 0 skipped, KS $D_n < 0.001358$ ($n=10^6$), 0 worker drift (`1/4/16`) | `cd Codigo && go test -count=1 -v ./...` | every edit, task end |
| Go PoC Build & LaTeX Macro Export | Clean build, exports `Docs/resultados_poc.tex` matching `poc_results.json` | `cd Codigo && go run .` | task end |
| Performance Budget (RNF06) | Median $< 500\text{ ms}$ (homologated 1 vCPU: $259.6\text{ ms}$) for 200 tasks / 300 deps ($N=10,000$) | `cd Codigo && go test -bench=Benchmark -benchmem -count=5` | task end |
| LaTeX & Bibliography Integrity | Balanced braces, 0 accented `\mathrm`, 0 literal `et al.`, all `\cite`/`\ref` resolved, 0 `"Aceito"` in Table 11, all 8 RGB PNGs present and $\ge 10\text{ KB}$ | `python check_latex.py` | every edit, task end |
| Dataset Privacy & Schema | Valid JSON, 0 contributor `login`/email fields, includes both PR review latency and corrected engineering duration (`first_commit_at`) | `python Codigo/extract_dataset.py --verify-local` | task end |

## Measured, not yet enforced

| Metric | Today | Direction |
|--------|-------|-----------|
| Go PoC test count | 10 tests | must not fall |
| Total size of 8 RGB PNG figures (`Docs/figuras/`) | ~727 KB (opaque RGB, 200 DPI) | must stay < 1.5 MB for Overleaf Free |

