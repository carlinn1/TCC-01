# Rastreabilidade de Execução da `SPEC_CORRECAO_TCC1.md` (Pareceres CL, CX e AR)

> **Status Global:** 100% Concluído e Verificado (Zero commits realizados conforme restrição do usuário)
> **Cobertura de Testes Go (`poc-engine`):** `94,2%` (`go test -count=1 -cover ./...`)
> **Verificação Numérica e LaTeX/BibTeX:** `PASS` (`check_numbers.py` e `check_latex.py`)

---

## Fase 0 — Baseline, ADRs e Governança de Qualidade (`BASE-01..05`)

- [x] **BASE-01:** Tag Git `pre-correcoes` criada sobre o commit base (`4ac2f8b`) e logs originais arquivados em `Docs/baseline/`.
- [x] **BASE-02:** Scripts automatizados `Codigo/check_numbers.py` e `Codigo/check_latex.py` criados e validados com saída `0`.
- [x] **BASE-03:** Sete Architecture Decision Records criados em `Docs/adr/ADR-001-poder-estatistico-h1b.md` a `Docs/adr/ADR-007-contribuicao-principal-e-custodia-ancora.md` documentando as decisões `DEC-1` a `DEC-7`.
- [x] **BASE-04:** Rascunhos históricos (`proposta_tcc1_*.md`, `avaliacao_tcc1_*.md`) movidos para `Docs/legado/`.
- [x] **BASE-05:** `CONSTRAINTS.md` atualizado com as restrições verificáveis `C1` a `C9`.

---

## Fase 1 — Motor Go (`poc-engine`) e Testes Automatizados (`COD-01..10`)

- [x] **COD-01:** Teste `TestBernoulliRiskComponent` auditado por injeção de mutação (`ImpactDays * 0` falha imediatamente com `11.44 < 11.80`).
- [x] **COD-02:** Comentário de `Risk.TaskID` corrigido em `Codigo/model.go` e validação de referência órfã implementada em `ValidateNetwork`.
- [x] **COD-03:** Funções `ValidateTasksAndDeps` e `ValidateNetwork` implementadas em `Codigo/model.go` e chamadas em `TopologicalSort` (`cpm.go`) e `RunMonteCarlo` (`simulator.go`), testadas por `TestInputValidationRejectsInvalidNetworks` (`engine_test.go`).
- [x] **COD-04:** Função `CalculateExpectedDurationWithLambda(to, tm, tp, lambda)` implementada em `Codigo/cpm.go` e integrada em `simulator.go`, testada por `TestLambdaConsistencyAcrossCPMAndMonteCarlo` (`engine_test.go`).
- [x] **COD-05:** Função `ComputeKSBetaFit` em `Codigo/random.go` calcula o supremo exato $D_n$ sobre todas as $n=10^6$ estatísticas de ordem (`DnFull`) além da malha de 2.000 pontos (`DnGrid2000`), e `TestRegularizedIncompleteBetaAnalytical` valida $I_x(\alpha,\beta)$ contra formas analíticas fechadas.
- [x] **COD-06:** Suíte `RunKSMultiSeedSuite` (100 sementes independentes $\times 10.000$ amostras) e teste `TestSplitMix64StreamIndependence` ($|r| \le 0{,}03$) implementados e verificados em `validation_test.go`.
- [x] **COD-07:** Função `ComputeCRNDiff` implementada em `Codigo/simulator.go`, quantificando o ganho de redução de variância CRN entre os Cenários C e C$'$ ($\Delta\bar{F} = -0{,}37 \pm 0{,}0059$ d.u. vs. $\mathrm{SE}_{\mathrm{indep}} = 0{,}0146$ d.u., redução de $6{,}1\times$).
- [x] **COD-08:** `RunAllPoCValidationAndExport` em `Codigo/main.go` unifica toda a exportação para `Codigo/poc_results.json`, `Docs/resultados_poc.tex` e `Docs/tabelas/*.tex`.
- [x] **COD-09:** `RunBenchmarkSuite(5)` em `Codigo/main.go` mede escalabilidade para 1, 2, 4 e 8 *workers* (`242,87 ms` $\rightarrow$ `122,35 ms` $\rightarrow$ `62,32 ms` $\rightarrow$ `59,07 ms`) após eliminação de *false sharing* via `workerCritCounts`.
- [x] **COD-10:** `AuditTCC2Schedule` em `Codigo/main.go` audita matematicamente as 14 tarefas do TCC2, soma das trilhas (`11,08`, `10,75`, `4,08`), caminho crítico (`16,75` sem.), folgas (`TF(E11)=1,58`, `TF(E06)=TF(E07)=0,50`) e simula os 4 cenários (Base, Com Riscos, Estresse `+25%` [$P_{85}=22{,}38$] e Estresse + Cortes Pré-Registrados [$P_{85}=19{,}82 \le 20{,}0$]).

---

## Fase 2 — Metodologia e Estatística (`EST-01..06`)

- [x] **EST-01 (`DEC-1` / `ADR-001`):** Script `Codigo/power_h1b.py` calcula o poder binomial exato de H1b, demonstrando poder `0,0%` para $M=200$ (IC Wilson 95%) e **`86,2%`** para $M=500$ sob teste de equivalência TOST ($\alpha=0{,}05 \equiv$ IC Wilson 90% em $[80\%,90\%]$). Tabela `\ref{tab:poder_h1b}` incorporada ao Cap. 3.
- [x] **EST-02 (`DEC-2` / `ADR-002`):** Pareamento funcional-perda corrigido segundo Gneiting (2011) em RQ4, H1a, H1b e Seção 3.4: $P_{50} \leftrightarrow \mathrm{MAE}$, $\bar{F} \leftrightarrow \mathrm{RMSE}$, $P_{85} \leftrightarrow L_{0{,}85}$ e $\hat{F}_N \leftrightarrow \mathrm{CRPS}$.
- [x] **EST-03 (`DEC-3` / `ADR-003`):** Regra de parametrização em 3 níveis especificada na Seção 3.3 (Nível 1: $Q_{05}/\hat{m}_{\mathrm{HSM}}/Q_{95}$ quando $Q_{50} > 0{,}25$; Nível 2: casamento $Q_{10}/Q_{90}$ quando $Q_{10} > 0{,}30$; Nível 3: *fallback* de *bootstrap* empírico quando $Q_{50} \le 0{,}25$), removendo a expressão `"extensão metodológica proposta"`.
- [x] **EST-04:** Equações paramétricas explícitas dos Regimes R1 a R10 (camadas topológicas, Weibull, Erlang, Log-Normal, cópula gaussiana com fator latente $Z_0$ e deriva $+40\%$) detalhadas na Seção 3.4.
- [x] **EST-05:** Desenho estratificado $5 \times 3 \times 4 \times 5 = 300$ cenários de H2a, bateria balanceada de 200 casos (100 corrompidos + 100 íntegros) para o verificador determinístico e política de até 2 regenerações antes de *fallback* especificados na Seção 3.4.
- [x] **EST-06 (`DEC-6` / `ADR-006`):** Extrator léxico/regex de H2b formalizado e auditoria cega dimensionada com 2 juízes internos em E10 ($\kappa \ge 0{,}80$) e 3º juiz externo em D03.

---

## Fase 3 — Dados e Piloto da Dimensão 3 (`DAD-01..05`)

- [x] **DAD-01 (`DEC-5` / `ADR-005`):** Vetores itemizados pseudonimizados (`PR_01..PR_n`) adicionados a `Codigo/dataset_backtesting.json` por `Codigo/extract_dataset.py`.
- [x] **DAD-02:** Contagem e fração de truncamento no piso de $0{,}25$ d.u. (`2/11 = 18,2%`, `7/13 = 53,8%`, `30/62 = 48,4%`) reportadas na Tabela `\ref{tab:backtesting}`.
- [x] **DAD-03:** Regras hierárquicas de inferência de dependências $FS$ e condicionamento por trilhas quando densidade $< 0{,}2$ arestas/tarefa especificadas na Seção 3.3.
- [x] **DAD-04:** Sensibilidade de $P_{50}$ e da média ao piso $\{0{,}10;\,0{,}25;\,0{,}50\}$ d.u. calculada e incluída na Tabela `\ref{tab:backtesting}`.
- [x] **DAD-05:** Script `Codigo/pilot_backtesting.py` executado, gerando `Codigo/pilot_backtesting_results.json` e a Tabela `\ref{tab:piloto_backtesting}` (`Docs/tabelas/tab_piloto_backtesting.tex`).

---

## Fase 4 — Segurança, Arquitetura e Governança de IA (`SEG-01..06`)

- [x] **SEG-01 (`DEC-7` / `ADR-007`):** Custódia criptográfica da âncora externa corrigida na Seção 4.5 (autenticação simétrica `HMAC-SHA256` com chave secreta em cofre isolado ou assinatura assimétrica `Ed25519` com chave pública somente-leitura, a cada 100 eventos ou 15 min).
- [x] **SEG-02:** Pipeline determinístico de sanitização em 5 estágios (normalização Unicode NFKC, remoção de caracteres invisíveis/Trojan Source, regex scrubbing, truncamento UTF-8 em 120/280 caracteres e *spotlighting* `<untrusted_git_metadata>`) especificado na Seção 4.4.1.
- [x] **SEG-03:** Máquina de estados completa (`DRAFT -> PENDING_APPROVAL -> {APPROVED, REJECTED, STALE, EXPIRED}`) e trava de concorrência otimista por `base_state_version` especificadas na Seção 4.4.2.
- [x] **SEG-04:** Tabela completa de ameaças STRIDE + OWASP LLM 2023/2025 (`\ref{tab:stride_owasp}`) adicionada ao Cap. 4.
- [x] **SEG-05:** Riscos residuais declarados explicitamente na Tabela `\ref{tab:stride_owasp}` e na Seção 6.4.
- [x] **SEG-06:** Serialização transacional no topo da cadeia `audit_logs` (`UNIQUE(prev_hash)` + `pg_advisory_xact_lock` / `FOR UPDATE` em *trigger* `BEFORE INSERT`) especificada na Seção 4.5.

---

## Fase 5 — Cronograma do TCC2 e Viabilidade Executiva (`CRO-01..06`)

- [x] **CRO-01 & CRO-02:** Somas por trilha (`11,08`, `10,75`, `4,08`), caminho crítico determinístico (`16,75` sem.) e folgas (`TF(E11)=1,58`, `TF(E06)=TF(E07)=0,50`) demonstrados e reportados na Seção 6.2.1.
- [x] **CRO-03 (`DEC-4` / `ADR-004`):** Quatro cortes ordenados de contingência dentro dos pacotes essenciais (`E10`, `E11`, `E09`, `E08`) especificados e simulados sob estresse de $+25\%$, reduzindo $P_{85}$ de $22{,}38$ para $19{,}82 \le 20{,}0$ semanas (Tabela `\ref{tab:sim_tcc2}`).
- [x] **CRO-04 & CRO-05 & CRO-06 (`DEC-6` / `ADR-006`):** Calendário real (`08/03/2027 a 23/07/2027`, 12 h/sem/autor) e dimensionamento de hardware/custo de LLM (`< US$ 5,00`) incorporados à Seção 6.2.

---

## Fase 6, 7 e 8 — Bibliografia, Coerência Textual, Tabelas e Figuras (`BIB-01..08`, `TXT-01..10`, `FOR-01..03`)

- [x] **BIB-01..08:** 50 referências auditadas em `Docs/bib_audit.md` e `Docs/Referencias.bib` (100% citadas em `Docs/Main.tex`), incluindo `gneiting2011`, `grubbs1962`, `buehler1994`, `goldratt1997`, `mcconnell2006`, `boehm1981` e `schuirmann1987`, além dos critérios de seleção/avaliação das Tabelas 2 e 3.
- [x] **TXT-01..10:** Terminologia padronizada (`Sombra Digital Governada`), seção de Contribuição Principal e Fronteira TCC1 vs. TCC2 (`Seção 1.4`, Tabelas `\ref{tab:contribuicao}` e `\ref{tab:fronteira_tcc1_tcc2}`), esquema relacional no Cap. 4 (Tabela `\ref{tab:esquema_er}`), discussão integrada da PoC (`Seção 5.7`), Apêndice B e `README.md` sincronizados.
- [x] **FOR-01..03:** 7 fragmentos de tabelas LaTeX gerados automaticamente e incluídos via `\input{tabelas/...}`, `\emergencystretch=1.5em` adicionado e todas as 8 figuras PNG regeneradas em 200 DPI RGB opaco a partir de `poc_results.json` sem sobreposições.
