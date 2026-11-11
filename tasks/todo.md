# Task Breakdown (v3) — Adequação ao Parecer Crítico

## Task 0: Formalizar `CONSTRAINTS.md` e plano de tarefas
- [x] Criar `CONSTRAINTS.md` na raiz e referenciar em `AGENTS.md`
- [x] Atualizar `tasks/plan.md` e `tasks/todo.md`

## Task 1: Soma de Kahan, MCSE, `Docs/resultados_poc.tex` e Simulação do Cronograma TCC2 em `Codigo/`
- [x] Adicionar `Mean`, `MeanSE`, `P50SE`, `P85SE`, `P95SE` em `Codigo/model.go` e `Codigo/simulator.go` (Kahan pós-`wg.Wait()` e estatísticas de ordem binomiais, mantendo sorteios intactos)
- [x] Adicionar `TestTCC2ScheduleMonteCarlo` em `Codigo/validation_test.go` com as estimativas pré-registradas das Trilhas A e B (E01–E14) e riscos da Tabela 16
- [x] Atualizar `TestPrintReproducibilityValues` para verificar variação zero em `Mean` ($\bar{F}$)
- [x] Atualizar `Codigo/main.go` para imprimir MCSE, simular o cronograma do TCC2 e exportar `Docs/resultados_poc.tex`

## Task 2: Duração corrigida desde 1º commit no Dataset e Regeneração das 8 Figuras PNG
- [x] Atualizar `Codigo/extract_dataset.py` e `Codigo/dataset_backtesting.json` com quantis de duração desde o 1º commit (`engineering_duration_business_days`) sem PII/logins
- [x] Atualizar `Codigo/generate_figures.py` (Figura 8 com o Gantt rebalanceado das Trilhas A e B) e regenerar as 8 figuras PNG RGB opacas (200 DPI) em `Docs/figuras/`, `figuras/` e `Codigo/figuras/`

## Task 3: Bibliografia (`Docs/Referencias.bib`) e Capítulos 1 e 2 (`Docs/Main.tex`)
- [x] Inserir as 12 referências verificadas por DOI/ISBN em `Docs/Referencias.bib` e sincronizar `referencias.bib`
- [x] Remover nomes dos examinadores e data da Folha de Aprovação em `Docs/Main.tex`
- [x] Atualizar Resumo, Abstract, Lista de Siglas (`CA`, `CRN`, `HSM`, `MCSE`, `NMAE`, `OFAT`, `PRNG`), Introdução (H1a/H1b, H2a/H2b, CA1/CA2, descompasso motivação vs. `cli/cli`, ressalva sobre persuasão) e Capítulo 2 (MacCrimmon/Magennis/Vacanti, Vose/Davis, Clark 1961, Williams 1992, `LLM05`/`LLM10`, CaMeL/Beurer-Kellner/NIST, Choetkiertikul 2018, Primavera Risk Analysis/Acumen/@RISK + URLs específicas, Figuras 2 e 3)

## Task 4: Metodologia e Arquitetura (Capítulos 3 e 4 de `Docs/Main.tex`)
- [x] Atualizar RNF06 (latência motor + ponta a ponta) e RNF09 (comparação na precisão do literal citado)
- [x] Atualizar Seção 3.3 e Tabela 5 (duração corrigida desde 1º commit, nota sobre 24h vs. 8h para `0,06 d.u.`, $y_{\mathrm{congelado}}$ vs. $y_{\mathrm{release}}$, casamento quantílico com truncamento residual, HSM de Bickel & Frühwirth 2006, $n_{\text{classe}}\ge 15$, agrupamento hierárquico, fallback bootstrap, ajuste da Beta no subconjunto sem eventos adversos e condicionamento em $t_c$)
- [x] Atualizar Seção 3.4 e Tabela 6 (10 regimes com $M=200$/regime e família de Holm $m=3$ na Dim. 2; Dim. 3 descritiva com 4º baseline de $k=3$ releases; Dim. 5 com taxa bruta, recall do verificador e ablação com sinônimos/distratores/3º avaliador)
- [x] Atualizar Capítulo 4 (mover citações de Vite/Docker/W3C para Apêndice A, `LLM05`/`LLM10`, Guarda Simétrica de Deriva com teto acumulado e ressalva de persuasão não avaliada com humanos)

## Task 5: Resultados da PoC, Cronograma Simulado do TCC2 e Limpeza Global (Capítulos 5 e 6 + `README.md`)
- [x] Atualizar Capítulo 5 ("verificação do amostrador", MCSE como erro de simulação, CRN atual em C vs. C′ + substreams encadeadas para E05/E07, média $\bar{F}$ via Kahan na Tabela 12, "$H_0$ não rejeitada" na Tabela 11, "determinística em função do índice de iteração k", fallback da Figura 7)
- [x] Atualizar Capítulo 6 (Tabela 15 e Figura 8 rebalanceadas para 2 autores, subseção e tabela com os resultados do `poc-engine` sobre o cronograma do TCC2, Limitações Declaradas expandidas, promoção de `\scriptsize` para `\footnotesize` onde couber e limpeza global de metatexto de revisão/superlativos)
- [x] Sincronizar `README.md`

## Task 6: Quality Gate Final
- [x] Atualizar e executar `check_latex.py` (falha dura se faltar qualquer PNG, 0 citações/referências quebradas, 0 frases de revisão/superlativos, 0 "Aceito" na Tabela 11)
- [x] Executar `cd Codigo && go test -count=1 -v ./... && go run .`
