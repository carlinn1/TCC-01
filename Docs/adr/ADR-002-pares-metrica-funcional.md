# ADR-002: Pareamento entre Funcionais Preditivos e Funções de Perda Consistentes (Gneiting, 2011)

## Status
Accepted

## Date
2026-11-11

## Context
Na versão preliminar da monografia, a hipótese H1a e a Seção 3.4 pareavam a média amostral $\bar{F}$ ao Erro Médio Absoluto ($\mathrm{MAE}$) e a mediana $P_{50}$ ao Erro Mediano Absoluto ($\mathrm{MdAE}$). Conforme demonstrado por Gneiting (2011, *Making and Evaluating Point Forecasts*, JASA, 106(494), 746–762), uma função de perda é consistente para um funcional estatístico se e somente se esse funcional minimiza a perda esperada sob a distribuição verdadeira:
- A **mediana** ($P_{50}$) é o funcional Bayesiano ótimo que minimiza a **perda absoluta esperada ($\mathrm{MAE} = E[|Y - \hat{y}|]$)**;
- A **esperança matemática** ($\bar{F} = \hat{E}[F]$) é o funcional ótimo que minimiza a **perda quadrática esperada ($\mathrm{MSE} = E[(Y - \hat{y})^2]$ e sua raiz $\mathrm{RMSE}$)**;
- O **quantil de nível $\tau$** ($P_{85}$, $\tau=0{,}85$) minimiza a **perda quantílica assimétrica (*Pinball Loss* $L_\tau$)**.

## Decision
Parear estritamente em todo o protocolo experimental (Resumo, §1.2 RQ4/H1a, §3.4, Tabela 8 e §3.5):
1. **Mediana $P_{50} \leftrightarrow \mathrm{MAE}$** (acurácia pontual sob perda absoluta);
2. **Média $\bar{F} \leftrightarrow \mathrm{RMSE}$** (acurácia pontual sob perda quadrática);
3. **Quantil $P_{85} \leftrightarrow \text{Pinball Loss } L_{0{,}85}$** (compromisso tático);
4. **Distribuição acumulada $\hat{F}_N \leftrightarrow \mathrm{CRPS}$** (regra de pontuação estritamente própria global).

As métricas $\mathrm{MdAE}$ e $\mathrm{NMAE}$ são reportadas exclusivamente como estatísticas descritivas complementares de escala e robustez, sem atribuição de propriedade de elicitação ótima.

## Consequences
- Elimina a inconsistência teórica de funções de perda apontada na revisão técnica.
- Mantém compatibilidade total com a inferência primária da Dimensão 2 via *bootstrap* pareado unilateral com correção de Holm-Bonferroni ($m=2$ por regime e funcional contra CPM e regressão linear em H1a; $m=3$ entre os regimes R1, R3 e R5 para *Pinball Loss* em H1b) sobre $\Delta\overline{\mathrm{AE}}$ e $\Delta\overline{\mathrm{SE}}$ (reportando Wilcoxon pareado e estimador de Hodges-Lehmann como análise complementar).


## Complemento operacional — 13/11/2026
R7 preserva as marginais e riscos de R1, alterando a largura para k=2. R9 preserva as marginais PERT-Beta e riscos de R1, com transformação inversa da Beta sobre a cópula gaussiana rho=0,50. O motor preditivo permanece nominalmente independente (rho=0). Para cada grafo, o baseline recebe um histórico independente de 200 vetores completos; cada tarefa é reamostrada pela respectiva coluna, com riscos já incorporados e sem novo incremento. Todos os métodos têm acesso ao mesmo histórico e às estimativas visíveis. Em R10, o histórico permanece nominal e o viés de +40% incide somente sobre o desfecho futuro. A especificação integral está na seção de protocolo da monografia; sua execução pertence ao TCC2.
