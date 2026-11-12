# ADR-004: Política de Contingência e Cortes Ordenados Dentro dos Pacotes Essenciais do TCC2

## Status
Accepted

## Date
2026-11-11

## Context
Os pacotes desejáveis `D01`, `D02` e `D03` já estão situados abaixo da linha de corte e fora da rede de simulação dos pacotes essenciais `E01` a `E14`. Portanto, declarar apenas o descarte de `D01–D03` como contingência não reduz o prazo simulado do caminho crítico caso ocorram atrasos severos durante o TCC2.

## Decision
Definir um gatilho objetivo no monitoramento quinzenal do TCC2 ($P_{85} > 20{,}0$ semanas úteis) que aciona uma **sequência ordenada de cortes de escopo dentro dos pacotes essenciais**, preservando o **mínimo científico inegociável** (Dimensão 1 + Dimensão 2 núcleo + Dimensão 4 + Dimensão 5 com 1 LLM):
1. **Corte 1 (Agente LLM `E10`):** operar o pacote `E10` com **1 único modelo aberto essencial da família Llama-3** (8B local ou 70B via provedor com pesos fixados) e **2 juízes internos cegos** (os dois autores, $\kappa$ de Cohen $\ge 0{,}80$), mantendo o 2º LLM e o 3º juiz externo em `D03` (reduzindo `E10` de `2,0/2,5/3,5` para `1,5/2,0/2,5` sem. antes do estresse);
2. **Corte 2 (Worker de Simulação `E06` — em substituição ao corte inócuo anterior em `E11`, que possuía $\IC = 0\%$):** simplificar o *Simulation Worker* para modo síncrono *in-process* ou fila enxuta sem particionamento avançado de *workers*, preservando integralmente o contrato do motor Go (`1,5/2,0/3,0` $\rightarrow$ `1,0/1,5/2,0` sem. antes do estresse, mantendo `E11` inalterado);
3. **Corte 3 (Backtesting `E09`):** reduzir a coleta histórica do `cli/cli` de 10 para 5 *releases* mantendo o replay completo com congelamento causal (`1,5/2,5/3,5` $\rightarrow$ `1,0/1,5/2,0` sem. antes do estresse);
4. **Corte 4 (Regimes Sintéticos `E08`):** concentrar a Dimensão 2 nos 6 regimes nucleares R1, R3--R5 e R7--R8 (`1,5/2,0/3,0` $\rightarrow$ `1,0/1,5/2,0` sem. antes do estresse).

**Regra de Estresse Uniforme (`PK-03`):** na modelagem de estresse, o fator `+25%` sobre $(t_o, t_m, t_p)$ incide uniformemente sobre **todos os pacotes de execução `E01` a `E13`** *após* a aplicação dos cortes (preservando apenas `E14`, Defesa/Entrega final, fixo em `1,0/1,0/1,0` semana e ampliando a descrição de revisão/folga de integração em `E05/E06`).

## Consequences
- O motor `poc-engine` exporta em `tab:sim_tcc2` quatro cenários validados: Base (`CPM=16,75`, `P85=17,53` sem.), Riscos (`P85=18,31` sem.), Estresse +25% (`CPM=20,69`, `P85=22,40`, `P(F<=20)=11,4%`) e Estresse +25% com Cortes em pacotes críticos (`CPM=18,60`, `P85=19,57` sem., `P(F<=20)=94,7%`), comprovando viabilidade abaixo do teto de $20{,}0$ semanas.

