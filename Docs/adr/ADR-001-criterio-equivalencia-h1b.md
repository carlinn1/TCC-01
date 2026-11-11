# ADR-001: Redesenho do Critério de Equivalência de H1b (TOST / IC Wilson 90% com M = 500)

## Status
Accepted

## Date
2026-11-11

## Context
Na formulação anterior de H1b, exigia-se que a cobertura empírica do quantil $P_{85}$ sobre $M=200$ grafos independentes por regime tivesse seu intervalo de confiança de score de Wilson a 95% inteiramente contido na margem de equivalência $[80\%, 90\%]$ ($\pm 5$ pontos percentuais). Conforme demonstrado analiticamente e verificado por `Codigo/power_h1b.py`, para $M=200$ nenhum número de sucessos $k \in \{0,\dots,200\}$ satisfaz simultaneamente $L_{95\%}(k, 200) \ge 0{,}80$ e $U_{95\%}(k, 200) \le 0{,}90$ (em $k=170$, o IC 95% é $[79{,}39\%;\; 89{,}29\%]$; em $k=171$, $[79{,}95\%;\; 89{,}71\%]$; em $k=172$, $[80{,}51\%;\; 90{,}13\%]$), resultando em poder estatístico de **0,0%**.

## Decision
Adotar o **teste de equivalência de dois testes unilaterais (TOST) ao nível $\alpha = 0{,}05$**, que é matematicamente equivalente à inclusão do **intervalo de confiança de score de Wilson a 90%** ($1 - 2\alpha = 90\%$) dentro da faixa de equivalência prática $[80\%, 90\%]$ ($\pm 5$ p.p.), dimensionando **$M = 500$ grafos independentes por regime** (total de $5.000$ grafos nos 10 regimes da Dimensão 2).

## Alternatives Considered

### Manter IC 95% com M = 200 e ampliar margem para ±7,5 p.p. ([77,5%, 92,5%])
- **Prós:** Mantém $M=200$.
- **Contras:** Atinge poder de apenas 69,5% (abaixo do limiar convencional de 80%) e afrouxa a precisão de calibração.

### Manter IC 95% com ±5 p.p. e elevar M para 550–600
- **Prós:** Mantém $1-\alpha = 95\%$ bilateral.
- **Contras:** Na literatura de testes de equivalência (Schuirmann, 1987), usar IC $1-2\alpha$ (90%) já controla o erro Tipo I de ambos os testes unilaterais em $\alpha=0{,}05$; usar IC 95% corresponde a $\alpha=0{,}025$ unilateral conservador.

## Consequences
- Com $M = 500$ e TOST ($\alpha = 0{,}05$, IC de Wilson a 90% em $[80\%, 90\%]$), o poder estatístico sob calibração perfeita ($p_{\text{real}} = 0{,}85$) é de **86,2%** (e sob IC 95% atinge 75,4%, superando 80% já em $M=550$).
- O custo computacional adicional no pacote `E08` do TCC2 é desprezível ($5.000$ grafos $\times 60\text{ ms} \approx 5\text{ minutos}$).

