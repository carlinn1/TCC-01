# ADR-003: Parametrização PERT-Beta Robusta à Censura no Piso de 0,25 d.u.

## Status
Accepted

## Date
2026-11-13

## Context
Na caracterização empírica do repositório `cli/cli` (Tabela 7), a imposição do piso operacional de $0{,}25$ d.u. gera acúmulo de massa no limite inferior em janelas curtas (por exemplo, no Caso 2, `v2.40.0 -> v2.41.0`, tanto $Q_{10}$ quanto $Q_{50}$ valem $0{,}25$ d.u.). Quando $Q_{10} = t_o = 0{,}25$, a razão $(Q_{10}-t_o)/(t_p-t_o)$ anula-se e $F_{\mathrm{Beta}}(0) = 0 \ne 0{,}10$, tornando o sistema não linear da Equação 3.2 sem solução real em $t_o \ge 0{,}25$.

## Decision
Estabelecer uma **regra de decisão determinística simplificada em dois níveis** para a parametrização causal $(t_o, t_m, t_p)$ em cada corte $t_c$ sobre o subconjunto limpo $\mathcal{H}_{\text{limpo}}(t_c)$ (conforme DEC-3, removendo o antigo nível intermediário de casamento de quantis da Eq. 3.2):
1. **Nível 1 — Primária (Quantis Regularizados + Moda HSM):** quando a massa no piso for inferior a $50\%$ ($Q_{50} > 0{,}25$ d.u.) e $n \ge 15$ (na classe ou após *shrinkage* para o pool global $n_{\text{global}} \ge 30$), com dispersão $Q_{95} - \max(0{,}25;\; Q_{05}) > \delta$ ($\delta = 0{,}05$ d.u.), definir $t_o = \max(0{,}25;\; Q_{05})$, $t_p = Q_{95}$ e $t_m = \min(\max(\hat{m}_{\mathrm{HSM}}, t_o), t_p)$, onde $\hat{m}_{\mathrm{HSM}}$ é o estimador *Half-Sample Mode* (Bickel & Frühwirth, 2006) aplicado ao subconjunto acima do piso (maior que piso + 1e-9) se houver pelo menos três observações, ou à amostra completa caso contrário. Os quantis usam sempre a amostra completa; a moda condicional acima do piso evita que a massa criada pelo truncamento determine sozinha a moda;
2. **Nível 2 — Fallback Determinístico para Massa no Piso:** quando $Q_{50} \le 0{,}25$ d.u. (ex.: Caso 2 da Tabela 7) ou $t_p - t_o \le \delta$, com tolerância numérica de 1e-9 d.u., acionar automaticamente a amostragem por *bootstrap* empírico direto com reposição sobre $\mathcal{H}_{\text{limpo}}(t_c)$.

O estudo sobre `cli/cli` é posicionado explicitamente como **demonstração descritiva de viabilidade e parametrização causal**, declarando o tempo ocioso (*idle time*) entre *timestamps* de *commit*/PR versus tempo ativo de engenharia como ameaça à validade de construto.

## Consequences
- Simplifica a árvore de decisão para dois ramos mutuamente exclusivos e garante que o algoritmo nunca falhe numericamente diante de censura à esquerda no piso de $0{,}25$ d.u.
- Comprovado por teste automatizado em `Codigo/scripts/pilot_backtesting.py`.


## Verificação de 13/11/2026
O suporte não é artificialmente ampliado. Amostras constantes e dispersões até 0,05 d.u. usam bootstrap. O piloto TCC1 sinaliza n < 15, sem fingir agrupamento histórico causal; o agrupamento por classe é especificação do TCC2. Testes em `test_pilot_backtesting.py` cobrem os dois ramos, a fronteira de dispersão e a escolha da amostra para HSM.
