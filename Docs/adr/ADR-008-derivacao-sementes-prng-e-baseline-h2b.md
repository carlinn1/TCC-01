# ADR-008: Derivação Independente de Sementes PRNG (Mix64) e Baseline Determinístico da Hipótese H2b

## Status
Accepted

## Date
2026-11-11

## Context
1. **Colisão XOR no gerador SplitMix64 (`PK-02.1`):** Na versão anterior de `Codigo/poc-engine/random.go`, a função `DeriveIterationSeed(masterSeed, iter)` utilizava `masterSeed ^ uint64(iter)`. Essa combinação booleana possui simetria exata entre sementes adjacentes de paridade par/ímpar: como `42 ^ k == 43 ^ (k ^ 1)` para todo $k \in \{0, \dots, 2m-1\}$ (em particular $N = 10{.}000$), os conjuntos de sementes por iteração produzidos por $S_0 = 42$ e $S_0 = 43$ eram idênticos como multiconjuntos (permutados por troca de pares adjacentes), resultando na **mesma média, mesmos quantis e mesma probabilidade de meta** apesar da mudança de semente e contaminando `Seed: 42 + i` nas execuções de benchmark.
2. **Desacoplamento de rejeição na amostragem PERT-Beta (CRN):** Ao compartilhar uma única instância de PRNG por iteração entre tarefas e riscos, qualquer reparametrização em uma tarefa (`T3`) altera a contagem de rejeições do algoritmo de Marsaglia-Tsang (2000), dessincronizando os sub-fluxos das tarefas posteriores (`T4`).
3. **Ausência de baseline determinístico sem LLM em H2b (`PK-04`):** O protocolo experimental H2b comparava condições com diferentes níveis de IA sem incluir um grupo de controle determinístico puramente algorítmico gerado diretamente pelas métricas do motor Go, dificultando isolar o valor incremental real da síntese textual por LLM.

## Decision
1. **Derivação não comutativa por dupla mistura `Mix64` e sub-fluxos por entidade (`random.go` e `simulator.go`):**
   - Introduzir a função bijetora de avalanche SplitMix64 `Mix64(z uint64) uint64` com constante de Weyl áurea `goldenGamma = 0x9e3779b97f4a7c15`.
   - Definir `DeriveIterationSeed(masterSeed uint64, iteration int) uint64` como `Mix64(Mix64(masterSeed + goldenGamma) ^ (uint64(iteration) * goldenGamma))`, quebrando a bijetividade linear `s0 ^ k` entre sementes vizinhas.
   - Isolar cada entidade no laço de simulação com semente derivada por FNV-1a 64 bits + `Mix64` (`DeriveEntitySeedFromKey(iterSeed, key)`) para `"TASK:" + task.ID` e `"RISK:" + risk.ID` sobre `SplitMix64` alocado em *stack* (zero alocações no *heap*), garantindo pareamento CRN robusto mesmo sob alterações de parâmetros PERT e invariância completa à contagem de *workers* (`RNF08`).
   - Garantir espaçamento não colidente de sementes no conjunto de benchmarks (`Seed: uint64(42 + i*1000)`).
2. **Condição 0 em H2b (Baseline Determinístico sem LLM e Escopo Exploratório, `DEC-1`):**
   - Incorporar explicitamente no protocolo experimental da Dimensão 5 (`Cap. 2` e `Cap. 3`) a **Condição 0 (*Template Determinístico Estocástico*)**: um relatório estruturado gerado deterministicamente sem LLM contendo apenas $P_{50}$, $P_{85}$, $P(\le \text{meta})$, Índice de Criticidade ($IC$) e folgas reais do motor Go.
   - Reclassificar **H2b** como **análise exploratória descritiva** entre Condições C1/C2 (com LLM + Citação Âncora) e Condição 0 (sem LLM) na subamostra cega de 50 pares avaliada pelos dois autores ($\kappa \ge 0{,}80$), sem teste confirmatório de McNemar e declarando estudos de usabilidade, carga cognitiva ou viés de automação com usuários finais fora do escopo do TCC1/TCC2.

## Consequences
- `TestSeedIndependence42vs43` valida divergência estatística real entre `S0=42` e `S0=43` ($|r| \le 0{,}04$ e vetores distintos).
- `TestComputeCRNDiff` valida redução de variância CRN $> 3\times$ mesmo quando a reparametrização de `T3` altera sua taxa de aceitação Marsaglia-Tsang, além de redução $> 6\times$ no pareamento `C' - C` da Tabela 9.
- `RNF08` permanece determinístico para 1, 4 e 16 *workers*, com cobertura de testes Go em `94,1%`.
- O desenho experimental do TCC 2 adquire referência de controle sem IA para caracterizar qualitativamente e numericamente eventuais limitações do agente explicativo sem reivindicações confirmatórias sobre cognição humana.

