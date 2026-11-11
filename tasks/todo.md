# Task Breakdown — Correção Integral do TCC 1

## Task 1: Restaurar pacote Go completo em `Codigo/` e validar TDD (Caso Degenerado, Momentos, KS e Reprodutibilidade)

**Description:** Restaurar os módulos Go ausentes em `Codigo/` (`go.mod`, `model.go`, `random.go`, `main.go`, `engine_test.go`), adicionar a função explícita `SamplePERTBeta(to, tm, tp, lambda float64) float64` em `random.go` com proteção contra divisão por zero no caso degenerado (`tp - to <= 1e-12`), e garantir que todos os testes unitários e estatísticos (`TestPertBetaMomentsMatchAnalytical`, `TestPertBetaKSGoodnessOfFit`, `TestDegeneratePertBetaZeroVariance`, `TestDeterministicCPMScenarios`, `TestReproducibilityAcrossWorkers`) passem sem erros.

**Acceptance criteria:**
- [x] `Codigo/` contém o pacote Go autocontido (`go.mod`, `model.go`, `random.go`, `cpm.go`, `simulator.go`, `main.go`, `engine_test.go`, `validation_test.go`).
- [x] Caso degenerado `to = tm = tp` retorna `to` com variância zero e sem `NaN` tanto em `SamplePERTBeta` quanto em `RunMonteCarlo`.
- [x] Teste de aderência de Kolmogorov-Smirnov (`TestPertBetaKSGoodnessOfFit`) passa com $D_n < D_{\mathrm{crit}, 0.05} = 0.001358$ para T1, T2, T3 e T4 ($n = 10^6$).

**Verification:**
- [x] Tests pass: `cd Codigo && go test -v ./...`
- [x] Build succeeds: `cd Codigo && go build .`

**Dependencies:** None

**Files likely touched:**
- `Codigo/go.mod`
- `Codigo/model.go`
- `Codigo/random.go`
- `Codigo/main.go`
- `Codigo/engine_test.go`
- `Codigo/validation_test.go`

**Estimated scope:** Medium (4–5 files)

---

## Task 2: Unificar topologia do Benchmark de 200 tarefas / 300 dependências e medir dispersão de 5 repetições

**Description:** Atualizar `BenchmarkMonteCarloSynthetic200Tasks` em `Codigo/engine_test.go` para usar `buildSynthetic200Network()` (10 camadas de 20 tarefas, 180 arestas base + 120 arestas cruzadas = 300 dependências FS, 40 riscos de Bernoulli, semente 99), garantindo paridade com `TestBenchmarkDispersion5Runs` e com a Seção 5.6 de `Docs/Main.tex`. Executar 5 repetições do benchmark e sincronizar estatísticas.

**Acceptance criteria:**
- [x] `BenchmarkMonteCarloSynthetic200Tasks` e `TestBenchmarkDispersion5Runs` utilizam exatamente a mesma topologia `buildSynthetic200Network()`.
- [x] A descrição da topologia na Seção 5.6 de `Docs/Main.tex` corresponde exatamente às arestas geradas por `buildSynthetic200Network()`.
- [x] Os resultados das 5 repetições cumprem RNF06 (tempo mediano $< 1{,}0$ s) e estão sincronizados na Tabela de Momentos/KS e na Tabela de Benchmark de `Docs/Main.tex`.

**Verification:**
- [x] Tests pass: `cd Codigo && go test -v -run TestBenchmarkDispersion5Runs`
- [x] Benchmarks pass: `cd Codigo && go test -bench=Benchmark -benchmem -count=5`

**Dependencies:** Task 1

**Files likely touched:**
- `Codigo/engine_test.go`
- `Codigo/validation_test.go`
- `Docs/Main.tex`

**Estimated scope:** Small (2–3 files)

---

## Checkpoint 1: After Tasks 1–2 (Go Engine & Benchmarks)
- [x] `cd Codigo && go test -v ./...` passa 100% dos testes
- [x] `cd Codigo && go test -bench=Benchmark -benchmem -count=5` conclui com sucesso
- [x] Valores estatísticos (Momentos, KS, Reprodutibilidade e Benchmark) anotados e sincronizados no LaTeX

---

## Task 3: Sincronizar `Codigo/extract_dataset.py` e validar `Codigo/dataset_backtesting.json`

**Description:** Revisar `Codigo/extract_dataset.py` para garantir que inclui `v2.46.0` em `RELEASES` e `CASES`, contabiliza separadamente `reason:completed` vs `reason:not_planned` e PRs não mesclados, inclui a função auxiliar `fetch_first_commit_at` documentada no cabeçalho/Apêndice B e aceita argumentos CLI (`--repo`, `--out`) mantendo valores padrão compatíveis com `dataset_backtesting.json`.

**Acceptance criteria:**
- [x] `RELEASES` contém `v2.39.2`, `v2.40.0`, `v2.41.0` e `v2.46.0`, e `CASES` inclui `("v2.40.0", "v2.46.0", "v2.46.0 (janela ampliada)")`.
- [x] `extract_dataset.py` implementa o suporte CLI (`--repo`, `--out`) citado na linha 1233 de `Docs/Main.tex` e a função auxiliar de consulta do primeiro commit (`fetch_first_commit_at`).
- [x] `Codigo/dataset_backtesting.json` é JSON válido e seus números coincidem exatamente com a Tabela 5 (`tab:backtesting`) de `Docs/Main.tex`.

**Verification:**
- [x] Syntax & CLI check: `python Codigo/extract_dataset.py --help`
- [x] JSON validation: `Codigo/dataset_backtesting.json` corresponde aos 3 casos da Tabela 5 de `Docs/Main.tex`.

**Dependencies:** None

**Files likely touched:**
- `Codigo/extract_dataset.py`
- `Codigo/dataset_backtesting.json`

**Estimated scope:** Small (1–2 files)

---

## Task 4: Gerar estaticamente as 8 figuras PNG (300 DPI) via `generate_figures.py`

**Description:** Instalar `matplotlib` e `numpy` no ambiente Python local e executar `Codigo/generate_figures.py` para gerar todas as 8 figuras PNG em alta resolução (300 DPI) tanto em `Docs/figuras/` (ao lado de `Main.tex`) quanto em `./figuras/` e `Codigo/figuras/`, prevenindo timeout de compilação no Overleaf (Parte 1.5 e Parte 3 Item 1).

**Acceptance criteria:**
- [x] Os 8 arquivos PNG (`fig01_dag_exemplo.png` a `fig08_gantt.png`) são gerados sem erros com 300 DPI.
- [x] Os arquivos estão disponíveis em `Docs/figuras/`, `./figuras/` e `Codigo/figuras/` para carregamento imediato pela macro `\IfFileExists{figuras/#1}` no LaTeX.

**Verification:**
- [x] Build succeeds: `python Codigo/generate_figures.py`
- [x] File verification: 8 arquivos PNG verificados em `Docs/figuras/` e `./figuras/`.

**Dependencies:** None

**Files likely touched:**
- `Codigo/generate_figures.py`
- `Docs/figuras/fig01_dag_exemplo.png` .. `Docs/figuras/fig08_gantt.png`
- `figuras/fig01_dag_exemplo.png` .. `figuras/fig08_gantt.png`

**Estimated scope:** Small (1 script + 8 imagens geradas)

---

## Checkpoint 2: After Tasks 3–4 (Python & Visual Assets)
- [x] `python Codigo/generate_figures.py` gera as 8 imagens PNG com sucesso
- [x] `dataset_backtesting.json` e `extract_dataset.py` verificados

---

## Task 5: Aplicar correções de compilação LaTeX (Parte 1) e fechar checklist dos 19 itens da banca (Parte 3) em `Docs/Main.tex` e `Docs/Referencias.bib`

**Description:** Aplicar todas as correções de compilação da Parte 1 em `Docs/Main.tex` e auditar os 19 itens da Parte 3 em `Docs/Main.tex` e `Docs/Referencias.bib`.

**Acceptance criteria:**
- [x] Zero ocorrências de `\mathrm{` com acentos em `Docs/Main.tex`.
- [x] Zero ocorrências de `brazil` isolado (substituído por `brazilian`) em `Docs/Main.tex`.
- [x] Zero ocorrências de math mode cru em títulos de `\chapter`, `\section`, `\subsection`.
- [x] Bloco TikZ da Figura 8 idêntico à especificação da Parte 1.1.
- [x] URL `https://github.com/carlinn1/TCC-01` presente no Cap. 5 e Apêndice B, e todos os 19 itens da Parte 3 conferidos.

**Verification:**
- [x] Static audit script verifica 0 erros de regex para `\mathrm` acentuado, headings com `$`, barra em `\foreach` da Figura 8 e referências cruzadas `\ref`/`\cite`.

**Dependencies:** Tasks 1, 2, 3, 4

**Files likely touched:**
- `Docs/Main.tex`
- `Docs/Referencias.bib`
- `referencias.bib`

**Estimated scope:** Small (2–3 files)

---

## Task 6: Executar bateria de validação final automatizada (Parte 4)

**Description:** Executar na ordem todos os passos de verificação da Parte 4 de `Correçoes.md` (geração de figuras, testes Go, benchmarks Go com 5 repetições e auditoria estática de sintaxe LaTeX/BibTeX), confirmando ausência de erros e regressões.

**Acceptance criteria:**
- [x] `python Codigo/generate_figures.py` executa sem erros e popula `./figuras/` e `Docs/figuras/`.
- [x] `go test -v ./...` passa 100% dos testes.
- [x] `go test -bench=Benchmark -benchmem -count=5` executa as 5 repetições sem falhas.
- [x] Auditoria completa de LaTeX/BibTeX concluída com zero erros.

**Verification:**
- [x] Todos os comandos da Parte 4 executados e documentados.

**Dependencies:** Task 5

**Files likely touched:**
- `tasks/todo.md`
- `tasks/plan.md`

**Estimated scope:** XS (validação e relatório)
