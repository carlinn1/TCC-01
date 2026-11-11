# SPEC de Correção — TCC1 "Sombra Digital Governada e IA para Apoio à Decisão em Projetos de Software"

**Data:** 2026-11-11 · **Documento avaliado:** `TCC_01.pdf` (68 p. + capa) · **Repositório:** `carlinn1/TCC-01`
**Fontes consolidadas:** `[CL]` parecer de Claude (1ª rodada) · `[AR]` parecer Argon (`gemini-code-*.md`) · `[CX]` parecer Codex (`correcoes_codex.md`) · `[NOVO]` achado surgido ao consolidar (não estava em nenhum parecer).
**Veredito consolidado:** **NÃO apto para entrega** até fechar todos os itens **P0** e registrar as decisões **DEC-1 a DEC-7**.

> **Aviso aos agentes.** Esta SPEC foi escrita **sem acesso ao repositório**. Caminhos e âncoras de linha (`Main.tex#L695` etc.) vêm de `[CX]` e `[AR]` e **devem ser revalidados contra o HEAD** antes de qualquer edição. Números do Go (73,9% de cobertura, ~247 ms, KS D=0,00131) são os relatados por `[CX]`; reproduza-os na BASE-01 antes de confiar neles.

---

## 0. Como executar (orquestração)

**Legenda de prioridade:** `P0` bloqueia a entrega · `P1` necessário · `P2` melhoria (pode ser rebaixado por decisão dos autores).

**Ordem e paralelismo**

```
BASE (Fase 0) ──► COD (Fase 1) ─┬─► EST (Fase 2, após DEC-1/2/3) ─┐
                 SEG (Fase 4) ──┤                                  ├─► TXT (Fase 7) ─► FOR (Fase 8) ─► REL (Fase 9)
                 BIB (Fase 6) ──┤   DAD (Fase 3, após DEC-5) ──────┤
                 CRO (Fase 5, após COD-10 e DEC-4/6) ──────────────┘
```

**Skill por fase**

| Fase | Skills a invocar | Observação |
|---|---|---|
| 0 BASE | `incremental-implementation`, `using-agent-skills` | Um commit por item. |
| 1 COD | `test-driven-development`, `debugging-and-error-recovery`, `code-review-and-quality` | **Teste vermelho primeiro**, depois correção. |
| 2 EST | `spec-driven-development`, `source-driven-development`, `documentation-and-adrs` | Citar fonte primária; uma ADR por decisão. |
| 3 DAD | `test-driven-development`, `source-driven-development` | Documentação da API do GitHub como fonte. |
| 4 SEG | `security-and-hardening`, `documentation-and-adrs` | Correções de **especificação**; implementação fica para o TCC2. |
| 5 CRO | `spec-driven-development`, `constraint-driven-development` | Re-simular com o próprio motor. |
| 6 BIB | `source-driven-development` | Nada de metadados "de memória". |
| 7 TXT | `documentation-and-adrs`, `code-review-and-quality` | Propagar mudanças estatísticas por `grep` (ver C5). |
| 8 FOR | `incremental-implementation`, `debugging-and-error-recovery` | Compilar e **olhar** as páginas afetadas. |
| 9 REL | `code-review-and-quality`, `constraint-driven-development` | Revisão por agente que **não escreveu** as correções. |

---

## 1. Restrições (adicionar ao `CONSTRAINTS.md`)

- **C1** Não reiniciar o trabalho. Preservar: tema, DSRM, distinção TCC1/TCC2, PoC, preocupação com governança. `[CX]`
- **C2** **Fonte única de números.** Nenhum número de resultado em tabela, figura, resumo ou conclusão pode ser digitado à mão; todos derivam do export do Go (COD-08). `[CX5]`
- **C3** Nenhuma afirmação nova sem fonte primária verificada (autoria, título, ano, veículo, URL/DOI). `[CX6]`
- **C4** Não implementar o sistema integrado (escopo do TCC2). Só são permitidos: correções no motor da PoC, testes, scripts de extração/figuras e ajustes de especificação. `[CX]`
- **C5** Mudança estatística ⇒ propagar para **todos** os pontos: Resumo/Abstract, §1.2 (RQ4, H1a, H1b), §3.4, Tabela 8, §3.5, §6.5. Verificar por `grep` e listar as ocorrências no PR. `[CX2][NOVO]`
- **C6** Alterou código, semente ou parâmetro ⇒ regenerar todos os resultados (`go test -count=1 -v ./... && go run .`) antes de tocar no texto.
- **C7** Decisões de escopo (DEC-1 a DEC-7) exigem aprovação dos autores/orientador **antes** de editar os Caps. 1, 3 e 6. `[CL]`
- **C8** Cobertura de testes não pode cair abaixo da linha de base (BASE-01); benchmark acima do orçamento **falha** o teste (BASE-02). `[CX12]`
- **C9** Linguagem: onde só há teste em escopo limitado, usar "demonstrado/testado em escopo limitado", não "verificado" nem "garantido". `[CL][CX9]`

---

## 2. Divergências entre os pareceres e decisão adotada

| # | Tema | `[AR]` | `[CX]` | `[CL]` (1ª rodada) | **Decisão** |
|---|---|---|---|---|---|
| D1 | Veredito | "Quase pronto" | Não apto | Não apto | **Não apto.** `[AR]` não viu os problemas estatísticos, de código e de bibliografia. |
| D2 | Pareamento F̄↔MAE e P50↔MdAE | Elogia como "rigoroso" | **Incorreto**: média↔perda quadrática; mediana↔perda absoluta | Não detectei | **Adotar `[CX]`** (EST-02). Gneiting (2011) é a referência. |
| D3 | Critério de H1b | Considera sólido | Matematicamente inviável | Estimei "60–65% de chance" (**errado**) | **Adotar `[CX]`**. Reproduzido por script (ver §3): 0 de 201 resultados satisfazem. |
| D4 | Eq. 3.2 (casamento de quantis) | Instável com massa no piso | Cai em *fallback* no piso | Mal-condicionada | **Convergência.** Parametrização robusta como primária (EST-03). |
| D5 | 3º juiz humano | Mover para D03; dois autores (κ de Cohen) | Confirmar participação externa | Reduzir escopo da Dim. 5 | **Combinar**: E10 com dois autores cegos; externo em D03 (EST-05). Registrar limite de independência. |
| D6 | Llama-3.3-70B local | Trocar por 8B local ou 70B via provedor | Falta infraestrutura | Usar **um** modelo só | **Combinar**: um modelo essencial (EST-05, CRO-02). |
| D7 | Cronograma auto-simulado | Elogia | Retirar do **resumo** | "Elegante", com ressalvas | **Manter no Cap. 6 com ressalvas; retirar do resumo e dos objetivos** (TXT-01). |
| D8 | Corte D01–D03 como contingência | Aceita | **Incoerente**: já estão fora da simulação | Insuficiente | **Adotar `[CX]`+`[CL]`** (CRO-04). |
| D9 | `temperature=0` | — | Não garante determinismo | Tratei como determinístico | **Adotar `[CX]`** (EST-04). |
| D10 | Numeração OWASP 2023/2025 | — | — | Conferi de memória: **consistente** | Manter; ainda exigir checagem na fonte oficial (BIB-07). |
| D11 | Defeitos visuais do PDF | Detalha 5 tabelas, Tab. 19, 4 figuras | **Não compilou o PDF** | Vi só parte | **Adotar `[AR]`**. O texto extraído do PDF corrobora: `Qualitativa/conceitualNão`, `CenárioCPM`, `Tarefaanalítica`, `payload_{k-1}`, e a Tab. 19 termina em D02 (sem D03). |

**Retratações de `[CL]`** (não devem aparecer no backlog): (a) a estimativa de 60–65% para H1b; (b) "Trilha B = 10,74" — o cálculo exato dá **10,75**; (c) "inconsistência temporal de datas" — só permanece a necessidade de datas-calendário (CRO-02).

---

## 3. Verificações executadas na consolidação (reprodutíveis)

Script Python (somente biblioteca padrão), resultados:

1. **H1b, n=200, Wilson 95%, janela [80%, 90%]:** nenhum k de 0 a 200 satisfaz. k=170 (85,0%) → [79,39%; 89,29%]; k=171 → [79,95%; 89,71%]; k=172 → [80,51%; 90,13%]. **Confirma `[CX1]`.**
2. **Poder** (probabilidade de o critério ser satisfeito quando a cobertura verdadeira é exatamente 85%):

| M (grafos/regime) | IC 95%, ±5 p.p. | IC 90% (≡ TOST α=0,05), ±5 p.p. | IC 95%, ±7,5 p.p. | IC 95%, ±10 p.p. |
|---|---|---|---|---|
| 200 | 0,0% | 30,1% | 69,5% | 92,6% |
| 300 | 36,6% | 57,3% | 90,1% | 99,2% |
| 400 | 62,3% | 75,2% | 96,6% | 99,9% |
| 500 | 75,4% | 86,2% | 99,1% | 100% |
| 600 | 84,8% | 91,6% | 99,7% | 100% |
| 800 | 94,8% | 97,8% | 100% | 100% |
| 1000 | 98,5% | 99,4% | 100% | 100% |

   Menor M (passo 10) com poder ≥ 80%: **550** (95%, ±5), **440** (90%/TOST, ±5), **240** (95%, ±7,5), **150** (95%, ±10). Poder calculado **só** para p=0,85; incluir sensibilidade (p∈{0,83; 0,87}) na EST-01.
3. **Cronograma TCC2** (Tab. 19 + arestas de trilha): `T_CPM = 16,75` semanas; caminho crítico `E02→E03→E05→E08→E09→E10→E12→E13→E14`; somas esperadas: Trilha A 11,08, Trilha B 10,75, Conjunto 4,08; folga determinística de E11 = 1,58 e de E06/E07 = 0,50. **Confirma `[AR4.1]` e `[CL1.3]`.**

**Não verificado por mim (reproduzir na BASE-01):** achados de `[CX]` que dependem do repositório (cobertura, 247 ms, KS completo, testes com mutação, divergência P85 da Fig. 7, autoria no `.bib`). Defeitos visuais de `[AR]` foram corroborados apenas pelo texto extraído do PDF; é preciso compilar e olhar as páginas.

---

## 4. Decisões que dependem dos autores/orientador (bloqueiam fases)

Para cada uma, registrar uma ADR (`documentation-and-adrs`) em `Docs/adr/`.

| ID | Decisão | Opções (recomendação em **negrito**) | Bloqueia |
|---|---|---|---|
| **DEC-1** | Desenho de H1b | (a) **Teste de equivalência (TOST ≡ IC 90%) com M ≥ 500 por regime, margem ±5 p.p.** (poder 86%); (b) margem ±7,5 p.p. com M ≈ 300–400; (c) rebaixar H1b a análise descritiva. Custo computacional de M=500 é baixo (estimativa a partir dos benchmarks da PoC; confirmar). | EST-01, CRO-03 |
| **DEC-2** | Pares métrica↔funcional | **P50↔MAE; F̄↔RMSE/MSE; P85↔Pinball; CDF↔CRPS**; MdAE e NMAE apenas descritivos. | EST-02 |
| **DEC-3** | Parametrização primária | **Suporte por quantis regularizados (t_o=max(piso, Q05), t_m=HSM, t_p=Q95) como primária; Eq. 3.2 só se Q10 > piso+δ.** | EST-03 |
| **DEC-4** | Cortes de contingência **dentro** dos essenciais | Candidatos: um único LLM (já proposto); regimes de 10 para 6–7; E11 só tabular; E09 como piloto; M reduzido (recalcular poder). Definir ordem e gatilho. | CRO-04 |
| **DEC-5** | Dimensão 3 / pacotes E01 e E09 | (a) **Piloto rastreável de 1–2 janelas + validação manual de amostra de arestas** e E09 enxuto; (b) manter 10 releases; (c) cortar a Dim. 3. | DAD-04, DAD-05 |
| **DEC-6** | Calendário e capacidade | Data da semana 1, horas/semana por autor, disciplinas concorrentes, infraestrutura do modelo local, confirmação (ou não) de avaliador externo. | CRO-02, CRO-03 |
| **DEC-7** | Contribuição principal | **"Arquitetura de integração governada entre estado do projeto, simulação e assistência interpretativa, com verificação do núcleo estocástico."** | TXT-01, TXT-02 |

---

## 5. Backlog

### Fase 0 — Linha de base e salvaguardas

#### BASE-01 — Congelar o estado atual e reproduzir os números
- **Origem:** `[CX]` (números), `[CL]` · **P0** · **Arquivos:** todo o repositório
- **Ação:** `git tag pre-correcoes`. Rodar `cd Codigo && go test -count=1 -v ./... && go test -cover ./... && go run . && go test -bench=Benchmark -benchmem -count=5`. Salvar saídas em `Docs/baseline/`.
- **Aceite:** `Docs/baseline/` contém as saídas; cobertura, mediana de 1 worker e tabelas 13–17/20 conferem com o PDF (ou a divergência é registrada como achado novo).

#### BASE-02 — Criar os verificadores declarados e *gates*
- **Origem:** `[CX12]` · **P1** · **Arquivos:** `CONSTRAINTS.md`, `Codigo/check_latex.py` (ausente), testes
- **Problema:** o contrato exige `check_latex.py`, que não existe, e o checklist o marca como feito; benchmark não reprova ao estourar orçamento.
- **Ação:** criar `check_latex.py` (erros/avisos do log, `Overfull`, `Float too large`, referências indefinidas). Adicionar teste de orçamento com margem documentada para o benchmark do RNF06. Marcar no checklist o que é **automatizado** e o que é **medição manual**.
- **Aceite:** `python3 Codigo/check_latex.py` retorna ≠0 quando há `Float too large`/`Overfull` relevante; checklist distingue os dois tipos.

#### BASE-03 — Identificar a versão Markdown antiga como histórica
- **Origem:** `[CX12]` · **P2** · **Aceite:** arquivo movido para `Docs/legado/` com cabeçalho "versão histórica; diverge do texto principal em título, autoria e requisitos".

---

### Fase 1 — Motor em Go (TDD: teste vermelho → correção → teste verde)

#### COD-01 — Testes que realmente detectam a componente Bernoulli
- **Origem:** `[CX4]` · **P0** · **Arquivos:** `Codigo/simulator.go`, `engine_test.go`
- **Problema:** zerar o impacto de todos os riscos **não** quebra a suíte.
- **Ação:** testes com resultado conhecido: risco com p=0 (duração inalterada), p=1 (soma exata de Δ·(1−cᵢ)), p=0,5 (média dentro de MCSE). Rodar uma **mutação manual** (impacto=0) e confirmar que ≥1 teste falha.
- **Aceite:** a mutação "impacto=0" faz a suíte falhar; documentar no PR.

#### COD-02 — Validar entradas (to ≤ tm ≤ tp, não negativos, sem NaN)
- **Origem:** `[CX4]` · **P1** · **Arquivos:** `model.go`, `cpm.go`, `simulator.go`
- **Ação:** validação explícita com erro descritivo; teste com `(5,2,3)`, valores negativos, NaN/Inf, `ci∉[0,1]`, `pr∉[0,1]`, `Δr<0`.
- **Aceite:** entradas inválidas retornam erro; nenhuma chega ao CPM ou ao amostrador.

#### COD-03 — Risco sem `TaskID`
- **Origem:** `[CX4]` · **P1** · **Ação:** ou **rejeitar** na validação (recomendado, por simplicidade) ou implementar o efeito global descrito no contrato. Alinhar o contrato (comentário/Doc) à escolha.
- **Aceite:** teste cobre o caso; contrato e comportamento coincidem.

#### COD-04 — λ coerente entre CPM e simulador
- **Origem:** `[CX4]` · **P0** (afeta a análise OFAT λ∈{2,4,6}) · **Arquivos:** `cpm.go` (L61), `simulator.go`, `model.go`
- **Ação:** CPM deve usar μ = (to + λ·tm + tp)/(λ+2) com o **mesmo** λ do amostrador. Testes: λ∈{2,4,6} comparando μ analítica, média simulada (dentro de MCSE) e T_CPM.
- **Aceite:** para λ≠4, média simulada de cada tarefa ≈ μ do CPM; teste verde. Resultados publicados (λ=4) permanecem idênticos (diff vazio na BASE-01).

#### COD-05 — KS com estatística completa
- **Origem:** `[CX3]` · **P0** · **Arquivos:** `validation_test.go` (L187)
- **Problema:** o teste usa 2.000 das 10⁶ posições; a tabela publicada descreve o supremo completo.
- **Ação:** ordenar a amostra e calcular `D = max_i max(i/n − F(x₍ᵢ₎), F(x₍ᵢ₎) − (i−1)/n)` sobre **todas** as posições; p-valor pela distribuição de Kolmogorov; verificar a Beta regularizada contra uma implementação independente. Caso degenerado (5,5,5): verificação **determinística**, sem p-valor.
- **Aceite:** Tabela 15 regenerada do export (COD-08); `p=1,000` do degenerado removido; texto da §5.4 descreve exatamente o procedimento. `[CX]` relatou T3: D=0,00131, p=0,065 — confirmar.

#### COD-06 — KS com múltiplas sementes e independência de fluxos
- **Origem:** `[CL2.6]` · **P1** · **Ação:** repetir o KS com ≥100 sementes (n reduzido se necessário) e reportar taxa de rejeição (~5%) e uniformidade dos p-valores. Testar correlação entre fluxos de índices vizinhos (`SplitMix64(S0⊕k)` e `S0⊕(k+1)`) e entre tarefas.
- **Aceite:** tabela/figura com taxa de rejeição e p-valores; ausência de correlação serial significativa ou limitação declarada.

#### COD-07 — MCSE da diferença sob CRN
- **Origem:** `[CL2.5]` · **P1** · **Ação:** *bootstrap* pareado do ΔP85 (C vs. C′) usando as mesmas realizações; exportar ΔP50/ΔP85/ΔP95 com erro-padrão.
- **Aceite:** §5.3 item 3 reporta ΔP85 = −0,57 com seu **próprio** MCSE, não a raiz da soma dos quadrados.

#### COD-08 — Export único e versionado
- **Origem:** `[CX5]` · **P0** · **Arquivos:** `Codigo/main.go`, `Docs/resultados_poc.tex`
- **Ação:** `go run .` gera `poc_results.json` (esquema versionado, hash do commit, semente, N) com: cenários A/B/C/C′ (CPM, F̄, quantis, MCSE, P(F≤11), ICᵢ), `finishDates` brutos (ou ECDF/histograma exatos), estatísticas KS, benchmarks (com ambiente), cronograma TCC2 (inclui ICs e estresse) e COD-07. Gera também `Docs/tabelas/*.tex`.
- **Aceite:** rodar duas vezes produz arquivos idênticos (exceto timestamp); nenhum consumidor lê dado fora deste export.

#### COD-09 — Declarar o hardware e refazer o benchmark multi-worker
- **Origem:** `[CL2.8]` · **P1** · **Problema:** 259,6 ms foi medido em 1 vCPU e 62,3 ms com 8 workers, sem declarar a máquina; 1 vCPU não paraleliza 8 workers. `[CX]` mediu ~247 ms (1 worker), compatível.
- **Ação:** o benchmark imprime `GOMAXPROCS`, núcleos lógicos/físicos, CPU e kernel; medir 1, 2, 4, 8 workers **na mesma máquina**; reportar aceleração real.
- **Aceite:** Tabela 17 e §5.6 citam a máquina de cada linha; aceleração calculada a partir de medições comparáveis.

#### COD-10 — Cronograma do TCC2 como teste
- **Origem:** `[AR4.1][CL1.3][CL1.4]` · **P1** · **Arquivos:** `main.go` (`buildTCC2ScheduleNetwork`)
- **Ação:** teste afirmando `T_CPM=16,75`, caminho crítico `E02→E03→E05→E08→E09→E10→E12→E13→E14` e as arestas de trilha. Adicionar **cenário de estresse** (O/M/P +40% e/ou cauda pessimista) e o cenário de **corte** da DEC-4; exportar.
- **Aceite:** teste verde; export inclui P50/P85/P95 e P(F≤20) para base, riscos e estresse.

---

### Fase 2 — Estatística e protocolo (texto de §1.2, §3.3–§3.5, Tab. 8)

#### EST-01 — Redesenhar o critério de H1b
- **Origem:** `[CX1]` (verificado em §3) · **P0** · **Depende de:** DEC-1
- **Problema:** com M=200 e IC de Wilson 95% contido em [80%, 90%], o critério é inalcançável mesmo para modelo perfeitamente calibrado.
- **Ação:** aplicar a DEC-1. Incluir análise de poder reprodutível (`Codigo/power_h1b.py` ou teste em Go) com a tabela da §3 e sensibilidade a p∈{0,83; 0,85; 0,87}. Declarar a regra de decisão (TOST/IC 90% ou margem alternativa). Corrigir a frase "semi-largura ±4,9 p.p." (§3.4), a justificativa de M=200 (§3.5) e o pacote E08 (duração e M).
- **Aceite:** `grep -n "200 grafos\|M = 200\|\[80%, 90%\]"` não deixa ocorrência inconsistente; script de poder reproduz a tabela; ADR-001 registrada.

#### EST-02 — Corrigir a teoria dos pares métrica↔funcional
- **Origem:** `[CX2]` · **P0** · **Depende de:** DEC-2
- **Problema:** o texto atribui à média a otimalidade para MAE e à mediana a otimalidade para MdAE. Correto: média↔erro quadrático; mediana↔erro absoluto (MAE). `[AR]` elogiou o pareamento errado.
- **Ação:** reescrever RQ4, H1a, §3.4 (1º item), Tab. 8, §3.5 (3.i) e Resumo. Citar Gneiting (2011), *Making and Evaluating Point Forecasts*, JASA 106(494) (arXiv:0912.0902) — **verificar metadados na fonte** (BIB-01). Manter NMAE/MdAE só como descritivos, sem propriedade ótima atribuída.
- **Aceite:** `grep -n "minimiza.*MdAE\|preditor natural.*MAE"` sem resultados; H1a e Tab. 8 usam os pares da DEC-2; ADR-002.

#### EST-03 — Parametrização (to, tm, tp) robusta ao piso de 0,25 d.u.
- **Origem:** `[CL2.4][AR3.1][CX7]` · **P0** · **Depende de:** DEC-3
- **Problema:** no Caso 2, Q10=Q50=0,25 ⇒ F_Beta(0)=0≠0,10 ⇒ a Eq. 3.2 não tem raiz; chamar o procedimento de "extensão metodológica" sem prova de existência/unicidade é frágil.
- **Ação:** reordenar a §3.3.2 (item 3): parametrização primária por quantis regularizados; Eq. 3.2 apenas quando Q10 > piso+δ; *fallback* (bootstrap) quando há massa no piso. Reportar a fração de itens no piso por classe. **Não** adotar t_m=Q50 como padrão (o próprio texto distingue moda de mediana; usar só como *fallback* declarado). Retirar "extensão metodológica proposta" ou demonstrar existência. Corrigir a Eq. 3.2 truncada no PDF (lado direito "= 0,90" ausente — ver FOR-05).
- **Aceite:** árvore de decisão (pseudocódigo) na §3.3.2; teste que aplica a regra aos quantis da Tab. 7 (Caso 2 aciona o *fallback* de forma determinística); ADR-003.

#### EST-04 — Dimensão 5: determinismo, repetibilidade e intervalo exato
- **Origem:** `[CX8][CL2.2]` · **P0** · **Arquivos:** H2a, §3.4, Tab. 8
- **Problema:** o texto trata `temperature=0` como determinístico (daí "n_ef=1 por cenário"); a documentação do provedor admite variação.
- **Ação:** remover a premissa; registrar modelo+versão, configuração, *prompts* e respostas; medir repetibilidade em subamostra (ex.: 30 cenários × 3 execuções). Especificar se o limite de Clopper-Pearson é **unilateral** (0 falhas em 300 ⇒ 0,99%) ou **bilateral** (⇒ 1,22%). Pré-registrar (tag git) gerador de cenários e extrator antes da execução.
- **Aceite:** H2a e §3.4 sem a afirmação de determinismo; limite citado com a sua natureza; protocolo de repetibilidade escrito.

#### EST-05 — Dimensão 5: valor do LLM, juízes e escopo das conclusões
- **Origem:** `[CX8][AR3.2][CL2.2]` · **P1** · **Depende de:** DEC-4
- **Problema:** identificar a tarefa de maior ICᵢ é regra determinística; a ablação mede recuperação de um indicador, não explicabilidade. Juiz externo e Llama-70B local criam risco no caminho crítico.
- **Ação:** (1) incluir *baseline* **sem LLM** (relatório por *template* a partir das métricas); (2) limitar conclusões a **fidelidade numérica e identificação de gargalo** (eficácia sobre decisão humana → trabalho futuro); (3) E10 essencial com **dois autores cegos à condição** (κ de Cohen ≥ 0,80) e juízes externos (κ de Fleiss) em D03, declarando que autores avaliando o próprio extrator reduz a independência; (4) um modelo essencial: "modelo aberto da família Llama-3 (8B local ou 70B por provedor com pesos fixados)".
- **Aceite:** H2b, §3.4, Tab. 8 e Tab. 21 coerentes entre si; nenhum "três juízes" no escopo essencial.

#### EST-06 — Regimes, multiplicidade e baseline de Vacanti
- **Origem:** `[CL2.2]`, `[NOVO]` (família de testes) · **P1**
- **Ação:** (a) tratar **R1–R2 como controle interno** (esperados a favor do modelo), não como evidência confirmatória; confirmatório = R3–R6. (b) Definir a **família** do Holm: hoje m=3 por métrica, mas há 4–6 regimes × várias métricas — declarar se a correção é por regime, por hipótese ou global. (c) Admitir que, em redes convergentes, o baseline de Vacanti perde quase por construção e dar mais peso a R7 (quase-serial).
- **Aceite:** §3.4 e Tab. 8 definem a família e o nível de α; R1–R2 rotulados como controle.

---

### Fase 3 — Dados e extração (`Codigo/extract_dataset.py`, `dataset_backtesting.json`)

#### DAD-01 — Preservar observações por item (minimizadas)
- **Origem:** `[CX7]` · **P1** · **Ação:** o JSON mantém registros pseudonimizados por item (id sintético, durações, rótulos, flags), sem logins/e-mails (LGPD); `--verify-local` recalcula **todas** as estatísticas da Tab. 7 offline.
- **Aceite:** Tab. 7 reproduzida só a partir do JSON; teste confirma ausência de PII.

#### DAD-02 — `issue_assigned_at` e *squash*
- **Origem:** `[CX7]` · **P1** · **Problema:** o texto inclui `issue_assigned_at` na definição de dᵢ, mas o código não o coleta; o caminho Git usa a data do commit de *squash* como "primeiro commit".
- **Ação:** coletar a atribuição (eventos da issue) **ou** remover do texto; obter o primeiro commit pelos commits do PR (API) e contabilizar casos com dado ausente/aproximado.
- **Aceite:** definição de dᵢ no texto = definição no código; contagem de aproximações reportada.

#### DAD-03 — Tratar o piso de 0,25 d.u. como censura
- **Origem:** `[CX7][CL2.4]` · **P1** · **Ação:** reportar a fração de itens no piso por janela; sensibilidade do piso (0,10/0,25/0,50); sinalizar na Tab. 7 que P10 (e a mediana do Caso 2) são o próprio piso.
- **Aceite:** Tab. 7 com coluna/nota de fração no piso.

#### DAD-04 — Regras para inferir dependências FS
- **Origem:** `[CX7][CL2.3]` · **P1** · **Depende de:** DEC-5
- **Problema:** "PR fecha issue" é associação entre artefatos, não precedência Finish-to-Start; só ~46–55% dos PRs têm vínculo explícito.
- **Ação:** definir regras explícitas (ex.: links "depends on/blocked by", referências cruzadas) e validar manualmente uma amostra de arestas (dois autores, precisão reportada). Se a densidade for baixa, declarar e manter a Dim. 2 como confirmatória.
- **Aceite:** regras + resultado da validação manual documentados.

#### DAD-05 — Piloto rastreável do backtesting
- **Origem:** `[CX7]` · **P1** · **Depende de:** DEC-5 · **Ação:** 1–2 janelas ponta a ponta (parametrização causal com tfim ≤ tc, replay sem vazamento) como prova de viabilidade; não concluir o backtesting no TCC1.
- **Aceite:** notebook/script que reproduz o piloto a partir do JSON.

#### DAD-06 — Tabela 7: base dos percentuais e notação
- **Origem:** `[CL2.3]` · **P2** · **Ação:** informar se "% com 1º commit anterior ao PR" usa total ou sem bots; padronizar cabeçalhos "release anterior → alvo" (a coluna "Ampliada" repete "v2.40.0 →").
- **Aceite:** nota de tabela e cabeçalhos consistentes com §3.3.1.

---

### Fase 4 — Segurança e arquitetura (correções de **especificação**)

#### SEG-01 — Limites da sanitização por regex (RNF02)
- **Origem:** `[CX9]` · **P1** · **Ação:** trocar "bloqueio total de PII" por redução de risco; incluir na bateria da Dim. 4 um conjunto de conteúdo sensível que **não** casa com os padrões (nomes próprios, hosts internos, IDs de ticket, formatos de chave não previstos) e relatar a taxa de vazamento.
- **Aceite:** RNF02 e §4.4.1 sem garantia absoluta; teste planejado descrito.

#### SEG-02 — "Isolamento físico" → isolamento lógico/de rede
- **Origem:** `[CX9]`, `[NOVO]` (egress) · **P0 (redação)** · **Ação:** substituir "físico" em RNF07, §4.2, legenda da Fig. 4 e resumo, conforme a documentação do Docker (*bridge* é mecanismo de software). Declarar o risco residual: o serviço de IA tem saída HTTPS para o provedor; descrever restrição de *egress* (destino fixo) e que a allowlist é a única barreira de dados.
- **Aceite:** `grep -n "isolamento físico\|isolamento fisico"` sem ocorrências.

#### SEG-03 — Auditoria: escopo da detecção e tipo de chave
- **Origem:** `[CX9]`, `[NOVO]` (custódia) · **P0 (redação)** · **Arquivos:** §4.5, RNF03, RF10, CA2
- **Ação:** (a) "chave privada" → **chave secreta simétrica** (HMAC); (b) delimitar CA2 ao trecho ancorado e **definir o intervalo entre âncoras**; declarar o que não é detectável entre âncoras; (c) registrar que, se o atacante controla o Core (e a chave HMAC), a âncora simétrica não prova integridade — avaliar assinatura assimétrica (ex.: Ed25519) com chave fora do Core como alternativa; decidir em ADR-005.
- **Aceite:** CA2 reescrito com escopo e intervalo; ADR-005.

#### SEG-04 — Rejeição transacional de aprovação sobre estado obsoleto
- **Origem:** `[CX9]` · **P1** · **Ação:** especificar que a aprovação carrega `base_state_version` e é **rejeitada na mesma transação** (HTTP 409) se divergir do estado corrente, exigindo nova simulação; `is_stale` sozinho não basta. Atualizar RNF05/RF09.
- **Aceite:** regra escrita em §4.2/§4.4 e RNF05; caso de teste planejado na Dim. 4.

#### SEG-05 — Relato da bateria de ataques
- **Origem:** `[CX9]` · **P1** · **Ação:** CA1 declarado como resultado **daquela bateria** de 50 payloads, sem extrapolar para "segurança"; mencionar ataques adaptativos (CaMeL/AgentDojo) como limitação.
- **Aceite:** §3.4, Tab. 8 e §6.4 sem "segurança universal".

#### SEG-06 — Modelo de ameaças formal e diagrama ER
- **Origem:** `[CL4.2]` · **P2** · **Ação:** modelo de ameaças estruturado (STRIDE ou árvore de ataque: ativos, atores, fronteiras de confiança, ameaças **não** cobertas) e ER simplificado (`audit_logs`, `audit_anchors`, `simulation_jobs`, `simulation_scenarios`, `draft_risks`, `project_events`).
- **Aceite:** figura/tabela no Cap. 4 referenciada no texto.

---

### Fase 5 — Cronograma e escopo do TCC2

#### CRO-01 — Tabela 19: arestas de trilha, caminho crítico e ordem
- **Origem:** `[AR4.1][CL1.2][CL1.3]` · **P0** · **Ação:** nota explicitando a serialização por autor (A: E02→E03→E05→E06→E07→E11; B: E01→E04→E08→E09→E10) e mostrando a conta do caminho crítico de 16,75 semanas (ver §3); ordenar linhas por semana de início; distinguir **janelas semanais** (E09 = sem. 9–11) de **durações esperadas** (2,50) e declarar o *buffer*.
- **Aceite:** o leitor reproduz 16,75 a partir da tabela e da nota.

#### CRO-02 — Datas, disponibilidade e infraestrutura
- **Origem:** `[CX10][CL1.1]` · **P0** · **Depende de:** DEC-6 · **Ação:** semana 1 = data real; horas/semana por autor; disciplinas concorrentes; GPU/VRAM ou provedor para o modelo; situação do avaliador externo.
- **Aceite:** Seção 6.2 com datas-calendário e premissas de capacidade.

#### CRO-03 — Estimativas calibradas e cenário de estresse
- **Origem:** `[CL1.4][CX10]` · **P1** · **Ação:** explicar a origem das O/M/P (ex.: horas reais gastas no TCC1 como classe de referência); reportar P(F≤20) também no cenário de estresse (COD-10); reescrever "99,6%" como **condicional** às estimativas e aos três riscos, não como prova de viabilidade.
- **Aceite:** Tab. 20 com linha de estresse; texto condicional.

#### CRO-04 — Contingência que alcança o que está no cronograma
- **Origem:** `[CX10][CL1.5]` · **P0** · **Depende de:** DEC-4 · **Problema:** cortar D01–D03 não reduz o prazo simulado (já estão fora).
- **Ação:** lista **ordenada** de cortes dentro dos essenciais, gatilho (P85 > 20 sem. no monitoramento quinzenal) e **mínimo científico** (Dim. 1 + Dim. 2 núcleo + Dim. 4 + Dim. 5 reduzida). Re-simular o cenário cortado (COD-10).
- **Aceite:** §6.2.1 item 3 reescrito; P85 do cenário cortado reportado.

#### CRO-05 — Antecipar provas de viabilidade
- **Origem:** `[CX10]` · **P1** · **Ação:** *spikes* nas semanas 1–2 para (i) viabilidade do DAG/dados (DAD-04) e (ii) infraestrutura/integração do LLM, com critério de reformulação.
- **Aceite:** pacote ou subtarefa explícita na Tab. 19 com DoD.

#### CRO-06 — Re-simular e propagar
- **Origem:** `[CL]` · **P1** · **Ação:** após CRO-01 a CRO-05, regenerar Tab. 19/20 e Fig. 8 **a partir do export** (COD-08/10) e atualizar §6.5.
- **Aceite:** nenhum número do cronograma digitado à mão.

---

### Fase 6 — Bibliografia e fundamentação (`source-driven-development`)

#### BIB-01 — Auditar autoria/metadados de **todas** as entradas do `.bib`
- **Origem:** `[CX6]` · **P0** · **Arquivos:** `Docs/Referencias.bib` (L133 Lehner; L409 CaMeL; L417 *Design Patterns*)
- **Ação:** importar metadados das fontes primárias (arXiv/editora/DOI) e conferir autores, título, ano, veículo, páginas; entregar `Docs/bib_audit.md` com URL por entrada. Corrigir ao menos as 3 já apontadas.
- **Aceite:** `bib_audit.md` cobre 100% das entradas; zero divergência remanescente.

#### BIB-02 — Caixa alta ABNT nas citações
- **Origem:** `[AR2.5]` · **P2** · **Ação:** `{{BROOKS JR.}, Frederick P.}`, `{{VAN SLYKE}, Richard M.}`, `{{STEELE JR.}, Guy L.}` etc., e organizações em caixa alta conforme NBR 10520.
- **Aceite:** citações parentéticas uniformes no PDF.

#### BIB-03 — Método de seleção dos trabalhos relacionados
- **Origem:** `[CX6][CL3.1]` · **P1** · **Ação:** parágrafo curto na §2.6: bases, strings, período, critérios de inclusão/exclusão (não é preciso revisão sistemática completa) e como a lacuna foi inferida.
- **Aceite:** a §2.6 cita o procedimento; a Tab. 2 tem narrativa de lacuna ligada às RQs.

#### BIB-04 — Matriz de ferramentas (Tab. 3) com evidência por célula
- **Origem:** `[CX6][CL3.2]` · **P1** · **Ação:** cada "Não"/"Limitado" cita página, versão e data; "não encontrado na documentação" ≠ "não possui" (usar "Não documentado"). Checar apps de Monte Carlo no *marketplace* do Jira. Trocar "estado em out. 2026" pela data real da checagem.
- **Aceite:** cada afirmação negativa tem nota com URL e data.

#### BIB-05 — Ampliar a fundamentação (verificar tudo na fonte antes de citar)
- **Origem:** `[CL3.1]` · **P1** · **Candidatos (a partir da memória do avaliador — confirmar existência e metadados):** Goldratt (*Critical Chain*); McConnell (*Software Estimation*, cone da incerteza); Boehm (*Software Engineering Economics*); Buehler, Griffin & Ross (1994) sobre a falácia do planejamento; Grubbs (1962) sobre validação das estatísticas do PERT; literatura acadêmica de previsão probabilística em ágil para complementar Vacanti/Magennis.
- **Aceite:** ≥1 parágrafo sobre Critical Chain e outro sobre críticas ao PERT-Beta/λ=4; cada nova referência em `bib_audit.md`.

#### BIB-06 — Sustentar (ou rotular) as escolhas autorais
- **Origem:** `[CX11]`, `[NOVO]` (limiares) · **P1** · **Ação:** para c_rev=0,80 ("decomposição empírica 80/20" sem referência), teto de 0,75 do checklist, proporcionalidade (1−cᵢ), deriva ±30%/±20%, n_classe ≥ 15 e n_global ≥ 30: dar fonte **ou** declarar como **hipótese de projeto parametrizável** com análise de sensibilidade.
- **Aceite:** nenhum limiar apresentado como "empírico" sem referência.

#### BIB-07 — Conferir a numeração OWASP na fonte oficial
- **Origem:** `[CL3.3]` · **P1** · **Ação:** confrontar com os documentos oficiais 2023 (v1.1) e 2025 as correspondências LLM01, LLM02, LLM05, LLM06, LLM10. A conferência de memória do avaliador não encontrou erro.
- **Aceite:** nota de verificação (URL/versão) na §2.5.

#### BIB-08 — Ressalvas de fonte
- **Origem:** `[CL3.2][CL3.3]` · **P2** · **Ação:** declarar que CaMeL e *Design Patterns* são *preprints*; manter o CHAOS Report com a ressalva já presente ou substituir por estudo revisado por pares; incluir Buehler et al. (1994) ao lado de Kahneman & Tversky para a falácia do planejamento.
- **Aceite:** textos ajustados.

---

### Fase 7 — Estrutura e redação

#### TXT-01 — Alinhar Resumo, Abstract, Introdução e Objetivos ao que foi realmente entregue
- **Origem:** `[CL2.1][CX11]` · **P0** · **Depende de:** DEC-7
- **Ação:** deixar explícito: "verificação limitada ao **motor estocástico**; governança, integração GitHub, agente LLM e auditoria são **projetados**". Retirar do Resumo/Abstract/objetivos a simulação do próprio cronograma. Trocar "verificada(o)" por "demonstrada/testada em escopo limitado" onde cabível (C9). Reduzir a repetição de números (P50, P85, 62,3 ms, 259,6 ms, KS) a Resumo e Conclusões.
- **Aceite:** Resumo/Abstract coerentes entre si e com o §6.5; `grep` de números sem duplicação desnecessária.

#### TXT-02 — Uma contribuição principal
- **Origem:** `[CX11][CL4.2]` · **P1** · **Ação:** seção "Contribuições" com a contribuição principal (DEC-7) e o que é **conhecido** (CPM, Monte Carlo, PERT-Beta, aprovação humana) vs. **novo** (combinação governada, trava simétrica, parametrização sem dupla contagem, reprodutibilidade por índice). Diferenciar de CaMeL.
- **Aceite:** seção presente; demais "frentes" tratadas como componentes.

#### TXT-03 — RQs, objetivos e matriz RQ → TCC1/TCC2
- **Origem:** `[CL5][CX11]` · **P1** · **Ação:** reescrever a pergunta central em forma simples; objetivo geral com um verbo principal; desdobrar a RQ2 (4 critérios); quadro "RQ → o que o TCC1 responde → o que fica para o TCC2"; nas Conclusões, "objetivos atingidos, com qual evidência, e o que permanece aberto".
- **Aceite:** quadro presente; §6.5 lista objetivo × evidência.

#### TXT-04 — Mover estatística para o Cap. 3 e subdividir a §3.3.2
- **Origem:** `[CL4.1]` · **P1** · **Ação:** na §1.2, hipóteses em 1–2 linhas com remissão à §3.4; dividir a §3.3.2 em subseções (validade de construto; variável-alvo; parametrização; dupla contagem; tamanho amostral).
- **Aceite:** §1.2 sem nomes de testes estatísticos; subseções numeradas.

#### TXT-05 — Remover duplicações
- **Origem:** `[CL4.1][CX]` · **P1** · **Ação:** Apêndice A repete §4.1/§4.4 (debounce, SKIP LOCKED, CSP, rate limiting, idempotência) — manter uma vez; reduzir repetição de requisitos/controles entre introdução, metodologia, arquitetura e conclusão; mover detalhes operacionais para apêndice.
- **Aceite:** diff mostra remoção, sem perder referências cruzadas.

#### TXT-06 — Discussão do Cap. 5 e critérios de aceitação do TCC1
- **Origem:** `[CL4.2]` · **P1** · **Ação:** seção "Discussão": interpretação, o que os resultados **não** mostram, relação com a literatura; lista objetiva dos critérios que a Dim. 1 satisfaz.
- **Aceite:** seção presente.

#### TXT-07 — Terminologia e redação
- **Origem:** `[CL4.3]` · **P2** · **Ação:** padronizar "Sombra Digital Governada de Processo"; reduzir siglas no resumo; variar "arquitetura proposta"; "Gêmeo Digital de Processo" só como horizonte.
- **Aceite:** busca por variantes não encontra formas inconsistentes.

#### TXT-08 — Ajustes pontuais de texto
- **Origem:** `[CL2.5][CL2.7][CL7]` · **P2** · **Ação:** (a) Clark (§5.3 item 1) vale para médias iguais e independência — dizer; (b) comentar a diferença entre Tab. 13 (S0=42, N=10.000: P(F≤11)=29,4%; IC T3=61,3%) e Tab. 16 (S0=12345, N=5.000: 28,26%; 59,62%); (c) encurtar a conversão "1,45 h/0,48 h"; (d) padronizar Trilha B = 10,75.
- **Aceite:** notas presentes.

#### TXT-09 — Limites da transposição de Kritzinger
- **Origem:** `[CL3.2]` · **P2** · **Ação:** parágrafo discutindo que o "objeto físico" aqui é um processo sociotécnico e que a "sombra" é de estado de tarefas.
- **Aceite:** parágrafo na §2.4.

#### TXT-10 — Justificativa tecnológica e custo do LLM
- **Origem:** `[CL4.2]` · **P2** · **Ação:** comparar brevemente Go/PostgreSQL/Docker Compose com alternativas; estimar tokens, latência e custo por ciclo (ligado a LLM10).
- **Aceite:** subseção curta no Cap. 4.

#### TXT-11 — Capa e folha de aprovação
- **Origem:** `[AR4.2][CL4.3]` · **P2** · **Ação:** preencher nomes na folha de aprovação (ou marcar "versão submetida à banca de TCC1"); mês na capa conforme norma da FCTE/UnB.
- **Aceite:** páginas conferidas no PDF.

---

### Fase 8 — Forma (figuras, tabelas, LaTeX)

#### FOR-01 — Tabelas e legendas geradas do export
- **Origem:** `[CX5]` · **P0** · **Depende de:** COD-08 · **Ação:** Tabelas 13, 14, 15, 16, 17 e 20 via `\input{Docs/tabelas/*.tex}`; macros numéricas usadas também no texto corrido; legendas de figura vindas do export.
- **Aceite:** apagar um número do `.tex` e regenerar restaura o valor; `check_numbers.py` (REL-02) verde.

#### FOR-02 — Layout das tabelas
- **Origem:** `[AR2.1][AR2.2]` · **P0** · **Arquivos:** `Docs/Main.tex`
- **Ação:** Tabelas 2, 13, 15, 17, 20 — primeira coluna com largura explícita (`>{\raggedright\arraybackslash}p{3.4cm}`), `\setlength{\tabcolsep}{3.5pt}`; Tabela 19 — `\scriptsize`, `\tabcolsep` 3 pt, `\arraystretch` 0,85 ou `longtable`, dividindo essencial/desejável; **restaurar D03**.
- **Aceite:** `check_latex.py` sem `Float too large`/`Overfull`; páginas 24, 47, 50, 51, 55 e 57 inspecionadas visualmente; D01–D03 presentes.

#### FOR-03 — Figuras
- **Origem:** `[AR2.3][CX5][CL4.3]` · **P0** · **Arquivos:** `Codigo/generate_figures.py`
- **Ação:** **Fig. 7** lê as amostras do export (não reamostra com NumPy; `[CX]` achou P85≈12,80/13,49 contra 12,75/13,43 da legenda); **Fig. 3** rótulo central/quebrado e `figsize=(10.5, 4.2)`; **Figs. 5 e 6** mathtext (`r"$\mathrm{payload}_{k-1}$"`, `r"$H_{k-2}$"`, `r"$\mathrm{IC}_i$"`, `$\rightarrow$`); **Fig. 4** seta que cruza o texto do PostgreSQL e rótulo "Allowlist JSON" sobreposto; **Fig. 8** texto da linha de corte sem passar por cima de E14 (ou movido ao título).
- **Aceite:** as quatro (cinco) figuras inspecionadas a olho; legenda da Fig. 7 = export.

#### FOR-04 — Caminhos de arquivo
- **Origem:** `[AR2.4]` · **P1** · **Ação:** `\texttt{...}` → `\path{...}` (p. 45 e p. 67) para `generate_figures.py`, `resultados_poc.tex`, `extract_dataset.py`, `dataset_backtesting.json`.
- **Aceite:** nenhum caminho cortado na margem.

#### FOR-05 — Equação 3.2 truncada
- **Origem:** `[CL2.4]` · **P1** · **Ação:** conferir no `.tex` e no PDF (p. 31–32) que o lado direito "= 0,90" aparece; revisar outras equações longas.
- **Aceite:** equação íntegra no PDF.

---

### Fase 9 — Verificação final e release

#### REL-01 — Build completo e inspeção do PDF
- **Origem:** `[CX]` (não compilou), `[AR]` · **P0** · **Ação:** compilar (ambiente com LaTeX), `check_latex.py`, `pdftotext` buscando colisões (`Degener`, `conce`, `Configuc`, `MaFila`) e conferir visualmente as páginas 21, 24, 39, 44, 47, 50, 51, 55, 56, 57.
- **Aceite:** zero colisão; contagem de páginas e sumário coerentes.

#### REL-02 — Consistência numérica automática
- **Origem:** `[CX5]` · **P0** · **Ação:** `Codigo/check_numbers.py` extrai os números de `Main.tex`/PDF e compara com o export; falha em divergência.
- **Aceite:** verde; incluído nos *gates* (BASE-02).

#### REL-03 — Re-revisão independente e ADRs
- **Origem:** `[CL]` · **P0** · **Ação:** agente que **não escreveu** as correções revisa contra esta SPEC (`code-review-and-quality`); ADR-001 a ADR-007 completas.
- **Aceite:** checklist com todos os P0/P1 fechados e evidência por item.

#### REL-04 — Clone limpo e *tag* de entrega
- **Origem:** `[CL7]` · **P1** · **Ação:** clonar em ambiente limpo e rodar os comandos do Apêndice B; confirmar repositório público; `git tag tcc1-entrega`.
- **Aceite:** reprodução do zero sem intervenção manual.

---

## 6. Critério de pronto da SPEC

1. Todos os itens **P0** e **P1** fechados com evidência; **P2** fechados ou rebaixados por decisão registrada.
2. DEC-1 a DEC-7 com ADR aprovada.
3. REL-01 a REL-04 verdes.
4. Parecer final de re-revisão: **"apto para entrega"** declarado de forma explícita.

---

## 7. Matriz de rastreabilidade (nenhum achado fica órfão)

| Fonte | Item do parecer | Itens da SPEC |
|---|---|---|
| `[AR]` | 2.1 Tabelas 2, 13, 15, 17, 20 | FOR-02 |
| `[AR]` | 2.2 Tabela 19 cortada / D03 | FOR-02 |
| `[AR]` | 2.3 Figuras 3, 5, 6, 8 | FOR-03 |
| `[AR]` | 2.4 `\path` | FOR-04 |
| `[AR]` | 2.5 caixa alta ABNT | BIB-02 |
| `[AR]` | 3.1 Eq. 3.2 | EST-03 |
| `[AR]` | 3.2 juiz externo / Llama-70B | EST-05, CRO-02, DEC-4 |
| `[AR]` | 4.1 arestas de trilha | CRO-01, COD-10 |
| `[AR]` | 4.2 folha de aprovação | TXT-11 |
| `[CX]` | 1 H1b | EST-01, DEC-1 |
| `[CX]` | 2 métricas | EST-02, DEC-2 |
| `[CX]` | 3 KS | COD-05 |
| `[CX]` | 4 testes / entradas / λ | COD-01, COD-02, COD-03, COD-04 |
| `[CX]` | 5 figuras e tabelas da mesma execução | COD-08, FOR-01, FOR-03 |
| `[CX]` | 6 bibliografia | BIB-01, BIB-03, BIB-04 |
| `[CX]` | 7 duração/DAG | DAD-01…DAD-05, EST-03 |
| `[CX]` | 8 valor do LLM, `temperature=0` | EST-04, EST-05 |
| `[CX]` | 9 segurança | SEG-01…SEG-05 |
| `[CX]` | 10 cronograma | CRO-02…CRO-05 |
| `[CX]` | 11 contribuição, 80/20 | TXT-02, TXT-03, BIB-06 |
| `[CX]` | 12 artefatos/verificadores | BASE-02, BASE-03 |
| `[CX]` | "remover/reduzir" | TXT-01, TXT-05 |
| `[CX]` | números relatados (73,9%, ~247 ms, 10 testes) | BASE-01 |
| `[CX]` | consistência numérica texto × resultados | REL-02, CRO-06 |
| `[CL]` | re-revisão independente do conjunto | REL-03 |
| `[CL]` | 1.1 datas | CRO-02 |
| `[CL]` | 1.2–1.3 Tab. 19, soma das trilhas, 16,75 | CRO-01, COD-10 |
| `[CL]` | 1.4 estimativas/estresse | CRO-03, COD-10 |
| `[CL]` | 1.5 escopo | CRO-04, DEC-4 |
| `[CL]` | 2.1 promessa vs. entrega | TXT-01 |
| `[CL]` | 2.2 hipóteses (circularidade, IC, CP) | EST-01, EST-04, EST-06 |
| `[CL]` | 2.3 Dimensão 3 | DAD-04…DAD-06, DEC-5 |
| `[CL]` | 2.4 Eq. 3.2 | EST-03, FOR-05 |
| `[CL]` | 2.5 CRN | COD-07 |
| `[CL]` | 2.6 KS / semente | COD-05, COD-06 |
| `[CL]` | 2.7 Tab. 13 × 16 | TXT-08 |
| `[CL]` | 2.8 hardware | COD-09 |
| `[CL]` | 3.1–3.3 revisão/fontes/OWASP | BIB-03, BIB-05, BIB-07, BIB-08, TXT-09 |
| `[CL]` | 4.1 remoções | TXT-04, TXT-05 |
| `[CL]` | 4.2 faltas (discussão, ER, ameaças, tecnologia) | TXT-02, TXT-06, TXT-10, SEG-06 |
| `[CL]` | 4.3 redação/formatação | TXT-07, FOR-03, TXT-11 |
| `[CL]` | 5 RQs/objetivos | TXT-03 |
| `[CL]` | 7 verificações pontuais | TXT-08, REL-01, REL-04 |
| `[NOVO]` | Família do Holm | EST-06 |
| `[NOVO]` | Limiares de projeto (n_classe, n_global, ±30%/±20%) | BIB-06 |
| `[NOVO]` | *Egress* do serviço de IA; custódia da chave HMAC | SEG-02, SEG-03 |
| `[NOVO]` | Propagação por `grep` (C5) | todos os itens estatísticos |
