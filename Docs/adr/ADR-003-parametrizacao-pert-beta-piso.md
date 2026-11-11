# ADR-003: Parametrização PERT-Beta Robusta à Censura no Piso de 0,25 d.u.

## Status
Accepted

## Date
2026-11-11

## Context
Na caracterização empírica do repositório `cli/cli` (Tabela 7), a imposição do piso operacional de $0{,}25$ d.u. gera acúmulo de massa no limite inferior em janelas curtas (por exemplo, no Caso 2, `v2.40.0 -> v2.41.0`, tanto $Q_{10}$ quanto $Q_{50}$ valem $0{,}25$ d.u.). Quando $Q_{10} = t_o = 0{,}25$, a razão $(Q_{10}-t_o)/(t_p-t_o)$ anula-se e $F_{\mathrm{Beta}}(0) = 0 \ne 0{,}10$, tornando o sistema não linear da Equação 3.2 sem solução real em $t_o \ge 0{,}25$.

## Decision
Estabelecer uma árvore de decisão determinística em três níveis para a parametrização causal $(t_o, t_m, t_p)$ em cada corte $t_c$ sobre o subconjunto limpo $\mathcal{H}_{\text{limpo}}(t_c)$:
1. **Regra Primária (Quantis Regularizados + Moda HSM):** se a fração de itens retidos no piso for $< 50\%$ ($Q_{50} > 0{,}25$ d.u.) e $n \ge 15$ (na classe ou após *shrinkage* para o pool global $n_{\text{global}} \ge 30$), definir $t_o = \max(0{,}25;\; Q_{05})$, $t_p = Q_{95}$ e $t_m = \min(\max(\hat{m}_{\mathrm{HSM}}, t_o), t_p)$, onde $\hat{m}_{\mathrm{HSM}}$ é o estimador *Half-Sample Mode* (Bickel & Frühwirth, 2006);
2. **Variante de Comparação por Casamento de Quantis (Eq. 3.2):** avaliada na análise de sensibilidade de `E09` somente quando $Q_{10} > 0{,}25 + \delta$ ($\delta = 0{,}05$ d.u.) e o sistema admitir solução única monótona com $t_o \ge 0{,}25$;
3. **Fallback Determinístico para Massa no Piso:** quando $Q_{50} = 0{,}25$ d.u. (ex.: Caso 2 da Tabela 7) ou $t_p - t_o \le \delta$, acionar automaticamente a amostragem por *bootstrap* empírico direto com reposição sobre $\mathcal{H}_{\text{limpo}}(t_c)$.

Remove-se do texto a expressão *"extensão metodológica proposta"* para a Eq. 3.2, tratando as três regras como variantes empíricas comparadas no pacote `E09`.

## Consequences
- Garante que o algoritmo de parametrização nunca falhe numericamente diante de censura à esquerda no piso de $0{,}25$ d.u.
- Comprovado por teste automatizado em `Codigo/pilot_backtesting.py`.

