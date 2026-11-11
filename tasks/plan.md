# Implementation Plan (v3): Adequação Integral ao Parecer Crítico (TCC 1)

## Overview
Adequar integralmente a monografia (`Docs/Main.tex` e `Docs/Referencias.bib`), o motor em Go (`Codigo/`), o dataset e gerador de figuras (`Codigo/extract_dataset.py`, `Codigo/dataset_backtesting.json`, `Codigo/generate_figures.py`) e o `README.md` às observações metodológicas, estatísticas, de segurança e de forma do parecer crítico (v1–v3), preservando intactos os sorteios homologados dos Cenários A–C′ da PoC no TCC1.

## Architecture & Methodological Decisions
- **Escopo vs. Prazo (Entrega 27/11):** Mantidos intactos os sorteios de `DeriveIterationSeed(S0, k)` na PoC do TCC1 (onde CRN já vale para C vs. C′ porque o risco R1 é sorteado após todas as durações PERT-Beta), especificando *substreams* encadeadas `SplitMix64(SplitMix64(S0 ^ k) ^ id)` para alterações de estimativas (RF09) nos pacotes E05/E07 do TCC2.
- **Kahan, MCSE e Exportação Automática (`Docs/resultados_poc.tex`):** Adicionada média $\bar{F}$ via soma compensada de Kahan pós-`wg.Wait()` e Erro Padrão de Monte Carlo (MCSE da média e dos quantis por estatísticas de ordem binomiais), declarando o MCSE como erro exclusivo de simulação (sem incerteza paramétrica).
- **Parametrização $(t_o, t_m, t_p)$ e Riscos de Bernoulli sem Dupla Contagem:** Ajuste da PERT-Beta no subconjunto basal sem eventos adversos, casamento de quantis $Q_{10}/Q_{90}$ com truncamento superior residual reconhecido, moda HSM de Bickel & Frühwirth (2006), $n_{\text{classe}} \ge 15$ com agrupamento hierárquico e *fallback* para *bootstrap* empírico quando $\lambda=4$ não ajustar ou $t_o < 0{,}25$ d.u.
- **Cronograma do TCC2 Pré-registrado e Rebalanceado (Trilhas A e B):** Duas trilhas equilibradas (~11,1 e ~12,2 semanas esperadas antes do fecho conjunto E12–E14) serializadas por arestas $FS$ condicionais à atribuição dos 2 autores, com escrita contínua embutida nas estimativas de E01–E11 e simuladas no `poc-engine` (`TestTCC2ScheduleMonteCarlo`).
- **Dimensões 2, 3, 4 e 5:** Dimensão 2 estruturada em 10 regimes ($M=200$ grafos/regime) com família de Holm-Bonferroni ($m=3$ baselines); Dimensão 3 descritiva ($n_{\mathrm{ef}}=10$ janelas) com 4º baseline ($k=3$ releases anteriores); Dimensão 4 com Guarda Simétrica de Deriva e teto acumulado; Dimensão 5 separando taxa bruta do modelo (H2a, Clopper-Pearson unilateral em $n=300$), *recall* do verificador por precisão do literal citado e ablação contra extrator por palavras-chave pré-fixadas com sinônimos, distratores e 3º avaliador (H2b).

## Task List

### Phase 0: Quality Bar & Constraints
- [x] **Task 0:** Criar `CONSTRAINTS.md` e atualizar `tasks/plan.md` e `tasks/todo.md`

### Phase 1: Go Engine (`poc-engine`), Dataset Corrigido e Figuras PNG
- [ ] **Task 1:** Implementar soma de Kahan pós-join, MCSE por estatísticas de ordem, exportação `Docs/resultados_poc.tex` e `TestTCC2ScheduleMonteCarlo` em `Codigo/`
- [ ] **Task 2:** Atualizar `Codigo/extract_dataset.py` e `Codigo/dataset_backtesting.json` (duração desde 1º commit) e regenerar as 8 figuras PNG (incluindo Figura 8 com Trilhas A e B rebalanceadas)

### Phase 2: Monografia (`Docs/Main.tex` e `Docs/Referencias.bib`) e `README.md`
- [ ] **Task 3:** Atualizar `Docs/Referencias.bib` (12 novas referências verificadas) e Capítulos 1 e 2 de `Docs/Main.tex` (Folha de Aprovação, Siglas, H1a/H1b, H2a/H2b, CA1/CA2, Revisão e Matriz de Ferramentas)
- [ ] **Task 4:** Atualizar Capítulos 3 e 4 de `Docs/Main.tex` (RNF06/RNF09, Tabela 5 com duração corrigida, $y(t_c)$, parametrização sem dupla contagem, 10 regimes na Dim. 2, Dim. 3 descritiva, Dim. 5 e Guarda Simétrica com teto acumulado)
- [ ] **Task 5:** Atualizar Capítulos 5 e 6 de `Docs/Main.tex` e `README.md` (verificação do amostrador, MCSE, Kahan, "$H_0$ não rejeitada", simulação do cronograma do TCC2, limitações e limpeza global de linguagem)
- [ ] **Task 6:** Executar suíte completa de verificação (`go test`, `go run .` e `check_latex.py` expandido)
