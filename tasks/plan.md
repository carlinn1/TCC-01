# Implementation Plan: Correção Integral do TCC 1 (LaTeX, Go PoC, Python Scripts & 19 Pendências da Banca)

## Overview
Executar o plano de correção integral especificado em `Correçoes.md` para o repositório `https://github.com/carlinn1/TCC-01`, aplicando os workflows de `using-agent-skills`, `planning-and-task-breakdown`, `test-driven-development`, `code-review-and-quality`, `security-and-hardening` e `performance-optimization`. O trabalho abrange: (1) eliminação de 100% dos erros e warnings de compilação LaTeX/BibTeX em `Docs/Main.tex` e `Docs/Referencias.bib`; (2) geração estática das 8 figuras PNG (300 DPI) via `generate_figures.py` para evitar timeout no Overleaf; (3) restauração, unificação e validação TDD/Benchmark do pacote Go (`Codigo/`) com caso degenerado ($t_o=t_m=t_p$), teste KS ($n=10^6$), reprodutibilidade concorrente (1/4/16 workers) e benchmark da rede de 200 tarefas / 300 dependências (5 repetições); (4) sincronização de `Codigo/extract_dataset.py` com `dataset_backtesting.json` e Apêndice B; e (5) auditoria final dos 19 itens da banca revisora (incluindo correção da URL oficial `https://github.com/carlinn1/TCC-01` e hash do commit).

## Architecture & Review Decisions (Five-Axis Audit)
- **Correctness (LaTeX & TikZ):** Em `Docs/Main.tex`, o fallback TikZ da Figura 8 (`fig:gantt`) foi atualizado protegendo todos os rótulos com `{...}`, substituindo `Buffer/Slides` por `Buffer e Slides` e adicionando `%` após `gray`.
- **Correctness (Math Mode, Babel, Hyperref & Case-Sensitivity):**
  - Substituídas as 3 ocorrências de `\mathrm{início}` (linhas 624, 761, 765 de `Docs/Main.tex`) por `\text{início}` para eliminar `Command \' invalid in math mode`.
  - Atualizada a opção `brazil` para `brazilian` em `\documentclass` (linha 9) e `\selectlanguage{brazilian}` (linha 82).
  - Envolvido `$\rightarrow$` em `\texorpdfstring{$\rightarrow$}{->}` no título da seção da Matriz de Rastreabilidade (linha 1057).
  - Adicionado `\IfFileExists{referencias.bib}{\bibliography{referencias}}{\bibliography{Referencias}}` e cópia em minúsculo para compatibilidade total em sistemas case-sensitive (Linux/Overleaf).
- **Test-Driven Development & Code Quality (Pacote Go em `Codigo/`):**
  - Restaurado o pacote Go completo em `Codigo/` (`go.mod`, `model.go`, `random.go`, `cpm.go`, `simulator.go`, `main.go`, `engine_test.go`, `validation_test.go`).
  - Adicionada `SamplePERTBeta(to, tm, tp, lambda float64) float64` em `random.go` com guarda explícita para `tp - to <= 1e-12`.
  - Alinhado `BenchmarkMonteCarloSynthetic200Tasks` em `engine_test.go` para invocar `buildSynthetic200Network()`.
- **Performance Optimization (RNF06 & Benchmarks):**
  - Mantida a referência homologada de 1 vCPU (5,8 ms para 4 tarefas e 259,6 ms para 200 tarefas/300 dependências) consistente entre Resumo, Abstract, Tabela 6 (`tab:protocolo`), Seção 5.6 (`tab:bench`) e Considerações Finais, e atualizados os valores reais do Teste KS ($D_n$ e $p$-valor) na Tabela 11 (`tab:momentos`).
- **Security & Hardening + Reproducibility (Checklist de 19 Itens):**
  - Verificados os 19 itens da banca revisora, incluindo atualização da URL oficial `https://github.com/carlinn1/TCC-01` e caminhos reais (`Codigo/`, `Docs/`) no Capítulo 5 e Apêndice B.

## Task List

### Phase 1: Código Go, Testes TDD e Benchmarks (Parte 2.1)
- [x] **Task 1:** Restaurar pacote Go completo em `Codigo/` e validar TDD (Caso Degenerado, Momentos, KS e Reprodutibilidade)
- [x] **Task 2:** Unificar topologia do Benchmark de 200 tarefas / 300 dependências e medir dispersão de 5 repetições

### Checkpoint 1: Go Engine & Statistical Validation
- [x] `go test -v ./...` passa 100% dos testes em `Codigo/`
- [x] `go test -bench=Benchmark -benchmem -count=5` executa sem erros e abaixo de 1,0 s

### Phase 2: Scripts Python e Geração Estática das 8 Figuras PNG (Partes 1.5 e 2.2)
- [x] **Task 3:** Sincronizar `Codigo/extract_dataset.py` (incluindo `v2.46.0`, filtro `completed` vs `not_planned`, `first_commit_at` e CLI args) com `dataset_backtesting.json`
- [x] **Task 4:** Executar `generate_figures.py` para gerar as 8 figuras PNG em `./figuras/` e `Docs/figuras/`

### Checkpoint 2: Python & Visual Assets
- [x] Todas as 8 figuras (`fig01` a `fig08`) existem em `./figuras/` e `Docs/figuras/` com resolução de 300 DPI
- [x] `dataset_backtesting.json` e `extract_dataset.py` estão 100% consistentes com a Tabela 5 e o Apêndice B

### Phase 3: Correções LaTeX/BibTeX e Auditoria dos 19 Itens da Banca (Partes 1, 3 e 4)
- [x] **Task 5:** Aplicar as correções de compilação LaTeX (Itens 1.1 a 1.4) e fechar lacunas dos 19 itens da banca (Item 5: URL `https://github.com/carlinn1/TCC-01`, caminhos e métricas) em `Docs/Main.tex` e `Docs/Referencias.bib`
- [x] **Task 6:** Executar bateria de validação final (Parte 4) e verificar compilação limpa do LaTeX/BibTeX

### Checkpoint 3: Complete Verification
- [x] Zero erros de sintaxe TikZ, math mode, babel ou hyperref em `Docs/Main.tex`
- [x] Todos os 19 itens da Parte 3 verificados e atendidos
