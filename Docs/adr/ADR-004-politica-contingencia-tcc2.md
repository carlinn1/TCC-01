# ADR-004: Planejamento Enxuto (Pre-Trimmed Bottom-Up) e Reserva de Contingência do TCC2

## Status
Accepted

## Date
2026-11-12

## Context
Na formulação inicial do cronograma do TCC2, os cortes de escopo eram acionados apenas reativamente após eventual estouro de prazo, e o caminho crítico nominal (`16,75` semanas) deixava pouca margem até o limite letivo de 20 semanas, especialmente sob correlação positiva entre pacotes ($\rho > 0$) e risco de viés otimista de estimativa. Adicionalmente, os antigos itens desejáveis `D01`–`D03` permaneceram fora da rede simulada e foram integralmente removidos do escopo (DEC-1, DEC-5).

## Decision
Adotar desde a linha de base um **cronograma enxuto pré-cortado (*pre-trimmed*) e reestimado *bottom-up* em horas-pessoa** (DEC-6), dentro do período de 20 semanas úteis (**01/03/2027 a 16/07/2027**, Resolução CEPE/UnB nº 0151/2026, sob teto de **480 h-pessoa** com 12 h semanais por autor):
1. **Reduções estruturais aplicadas a priori no planejamento (`E01`–`E14`):**
   - Remoção integral de `D01`–`D03` e estudos com usuários finais / carga cognitiva;
   - **Trilha A (`87 h` / `7,25` sem. esperadas):** `E02` (`18 h`, âncora única Ed25519), `E03` (`15 h`, REST+RBAC), `E05` (`12 h`, motor MC+CRN+cópula), `E06` (`12 h`, *Sim Worker* `SKIP LOCKED`), `E07` (`18 h`, proxy Llama-3 `temp=0.0` + alerta $\pm 30\%$) e `E11` (`12 h`, interface mínima funcional de cartão com CSP estrita);
   - **Trilha B (`81 h` / `6,75` sem. esperadas):** `E01` (`12 h`, extrator `cli/cli`), `E04` (`15 h`, *webhooks* HMAC + auditoria), `E08` (`18 h`, Dim. 2 enxuta com 6 regimes preservando $R_9/R_{10}$, $M=500$), `E09` (`18 h`, Dim. 3 descritiva com regra em 2 níveis) e `E10` (`18 h`, Dim. 5 com modelo único Llama-3 `t=0.0` e H2b exploratória);
   - **Fechamento Conjunto (`90 h` conjuntas / `45 h/autor` / `3,75` sem. esperadas):** `E12` (`30 h*`), `E13` (`36 h*`) e `E14` estocástica (`0,75/1,00/1,25` sem., incorporando variância de agenda de banca externa).
2. **Total de esforço e folga de defesa:** **`258 h-pessoa`** (**`53,8%`** do teto de `480 h`), resultando em caminho crítico determinístico de **`12,00` semanas** e deixando as semanas **12 a 20** como reserva explícita de contingência (sem qualquer corte reativo necessário para caber no calendário).
3. **Auditoria de estresse no motor Go (`poc-engine`):** avaliar quatro cenários com modelagem de correlação via cópula gaussiana ($\rho = 0{,}50$ por padrão nos cenários com risco): Base independente (`base`), Nominal com cópula $\rho=0{,}50$ (`with_risks`), Estresse `+25%` + $\rho=0{,}50$ (`stress_25`) e Estresse severo R10 `+40%` + $\rho=0{,}50$ (`stress_40`).

## Consequences
- O motor `poc-engine` exporta em `tab:sim_tcc2` quatro cenários validados sobre o cronograma de `12,00` semanas: Base (`CPM=12,00`, `P85=12,49` sem.), Nominal com riscos e cópula $\rho=0{,}50$ (`P50=12,52`, `P85=13,76` sem., `P(F<=20)=100,0%`), Estresse +25% + $\rho=0{,}50$ (`CPM=15,00`, `P85=17,00` sem., `P(F<=20)=100,0%`) e Estresse R10 +40% + $\rho=0{,}50$ (`CPM=16,80`, `P85=18,93` sem., `P(F<=20)=95,9%`), sustentando o cumprimento condicional do prazo de 20 semanas mesmo sob degradação combinada.

