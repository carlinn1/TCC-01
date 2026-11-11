# ADR-004: Política de Contingência e Cortes Ordenados Dentro dos Pacotes Essenciais do TCC2

## Status
Accepted

## Date
2026-11-11

## Context
Os pacotes desejáveis `D01`, `D02` e `D03` já estão situados abaixo da linha de corte e fora da rede de simulação dos pacotes essenciais `E01` a `E14`. Portanto, declarar apenas o descarte de `D01–D03` como contingência não reduz o prazo simulado do caminho crítico caso ocorram atrasos severos durante o TCC2.

## Decision
Definir um gatilho objetivo no monitoramento quinzenal do TCC2 ($P_{85} > 20{,}0$ semanas úteis) que aciona uma **sequência ordenada de cortes de escopo dentro dos pacotes essenciais**, preservando o **mínimo científico inegociável** (Dimensão 1 + Dimensão 2 núcleo + Dimensão 4 + Dimensão 5 com 1 LLM):
1. **Corte 1 (Fixado já na linha de base essencial):** operar o pacote `E10` com **1 único modelo aberto essencial da família Llama-3** (8B local ou 70B via provedor com pesos fixados) e **2 juízes internos cegos** (os dois autores, $\kappa$ de Cohen $\ge 0{,}80$), mantendo o 2º LLM e o 3º juiz externo em `D03`;
2. **Corte 2 (Interface `E11`):** substituir a SPA gráfica em React por inspeção tabular estática + aprovação via CLI/endpoint HTTP autenticado (reduzindo `E11` de `1,0/1,5/2,0` para `0,5/1,0/1,5` sem.);
3. **Corte 3 (Backtesting `E09`):** reduzir a coleta histórica do `cli/cli` de 10 para 5 *releases* mantendo o replay completo sobre o piloto já validado (reduzindo `E09` de `1,5/2,5/3,5` para `1,0/1,5/2,0` sem.);
4. **Corte 4 (Regimes Sintéticos `E08`):** concentrar a Dimensão 2 nos 6 regimes nucleares (R1 controle, R3–R5 confirmatórios e R7–R8 fronteira de estresse, reduzindo `E08` de `1,5/2,0/3,0` para `1,0/1,5/2,0` sem.).

## Consequences
- O motor `poc-engine` passa a simular e exportar na Tabela 20 quatro configurações: (1) Base PERT-Beta, (2) Com Riscos Discretos, (3) Cenário de Estresse (+25% nas estimativas O/M/P + riscos ativos, no qual $P_{85} > 20$ sem.), e (4) Cenário de Contingência Cortado sob Estresse (comprovando que os cortes internos reduzem o caminho crítico e restauram $P_{85} \le 20{,}0$ semanas).

