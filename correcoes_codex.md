# Parecer final sobre o TCC1 — versão atualizada após o pull

## Aplicação das correções — 13/11/2026

O parecer de 12/11 abaixo é histórico. Suas correções técnicas e textuais foram aplicadas nesta árvore de trabalho; a avaliação antiga não descreve o estado corrigido.

- **2.1 e 2.2:** regra de dois níveis unificada no script, monografia e ADR-003. HSM condicional acima do piso explicitado; fallback para dispersão até 0,05 d.u., com tolerância de 1e-9. Casos constantes e de fronteira cobertos; testes legados atualizados.
- **2.3:** tabela e macros de poder geradas de `power_h1b_results.json`; cobertura exportada automaticamente por `export_coverage.py`, com log e metadados em `Docs/baseline/coverage_current.*`. A medição atual é 91,7%, acima da meta de 80%. O ambiente do benchmark é exportado pelo motor; o README referencia as medições locais, sem exigir igualdade de tempos entre máquinas. O auditor valida a tabela de poder integral e as macros/valores selecionados, sem alegar cobertura universal dos números.
- **2.4:** marginais, riscos e transformação da cópula de R7/R9 definidos; histórico sintético separado de 200 vetores, informações disponíveis e baseline de reamostragem explicitados. O motor nominal permanece com rho=0. Decisões registradas também no ADR-002.
- **2.5:** janela pós-âncora incluída na tabela STRIDE; alcance de NFKC corrigido e fundamentado nas normas Unicode. A conclusão passou a distinguir a parametrização exploratória da avaliação temporal futura.
- **2.6:** auditoria reconciliada com 68 entradas (66 anteriores e duas normas Unicode); removidas três chaves obsoletas e caminhos Windows; registradas versões dos preprints e publicação antecipada de Amin et al. As oito referências faltantes foram incorporadas com fontes primárias; as demais linhas históricas foram preservadas, sem alegar nova checagem externa integral.
- **3:** IDs de risco vazios e duplicados rejeitados antes da simulação. O teste novo falhou antes da correção e passou depois; uma mutação temporária do guard também foi detectada e restaurada.
- **4 e 5:** base e limites das estimativas de esforço descritos; confirmação de infraestrutura e prazo administrativo registrada como ação preparatória do TCC2, sem inventar confirmação dos autores. Comparações estatísticas descartadas saíram da tabela principal; conclusão resumida e mensagens de auditoria delimitadas.
- **6 e 7:** suíte Go com 23 testes aprovada; suíte Python com 12 testes aprovada; dataset local e auditores numérico e LaTeX/BibTeX aprovados. Resultados e oito figuras regenerados. PDF compilado com Tectonic 0.17.0; corrigidos extravasamentos e travessões incompatíveis. O verificador agora aceita `--log` e confere a correspondência entre auditoria e BibTeX. A checagem do log final não encontrou extravasamentos horizontais, referências ou citações indefinidas. Foram inspecionadas visualmente seis páginas: piloto, poder, ameaças, abertura do cronograma, tabela de entregas e apêndice de reprodução. Permanecem avisos de espaçamento `Underfull`, sem extravasamentos nas páginas examinadas. Um teste de adulteração confirmou que o auditor rejeita alterações manuais na tabela de poder.

A confirmação de prazos administrativos, recursos e avaliação institucional permanece com os autores/orientador. Os experimentos e serviços reservados ao TCC2 continuam como especificação.

---

Data: 12/11/2026. Revisão examinada: `acfafcefd57e67ec2b1b6c9144a1cc4eb141b9d6`.

**Parecer: o projeto tem estrutura e contribuição adequadas a um TCC1, mas a versão examinada ainda precisa de ajustes antes da entrega final. Não a considero apta para entrega sem as correções necessárias descritas abaixo.** As pendências são localizadas; não justificam reiniciar a pesquisa nem ampliar o escopo.

Esta avaliação considera a versão posterior ao pull informado pelos autores. Os resultados de verificações anteriores ao pull não fundamentam este parecer. Foram examinados a monografia, os ADRs, o contrato de qualidade, a bibliografia, os dados, os scripts e o motor da PoC. As reproduções que geram artefatos e os experimentos adicionais foram feitos em cópia temporária. Este parecer não altera a monografia nem o código do projeto.

## 1. Correções anteriores que considero atendidas

| Ponto | Situação na versão atual |
| --- | --- |
| Delimitação do TCC1 | A especificação e a PoC estão distinguidas da integração e dos experimentos futuros do TCC2. Não é necessário entregar a aplicação web completa neste estágio. |
| Contribuição científica | A contribuição é apresentada como integração arquitetural governada de técnicas existentes. Essa delimitação é defensável e evita reivindicar originalidade sobre CPM, PERT ou cadeias de hash. |
| Piloto com dados reais | O piloto foi reclassificado como estudo exploratório de parametrização. O antigo replay que utilizava o resultado real foi removido; a avaliação preditiva temporal fica para o TCC2. |
| Extração de durações | A verificação local passou e não identificou a antiga inconsistência de duração de engenharia menor que o tempo de revisão do PR. A consulta ao histórico anterior ao squash está contemplada. |
| Sementes e CRN | A derivação das sementes foi revista e existem subfluxos por entidade. Os testes atuais de sementes 42/43, reprodução entre workers e CRN passaram. |
| Equivalência H1b | O procedimento adotado é binomial exato. Reproduzi, para M=500 e p=0,85, poder de 83,93%, região de aceitação de 416 a 438 sucessos e erros Tipo I de 3,94% e 4,65% nas fronteiras. |
| H1a e medida de efeito | O bootstrap pareado passou a sustentar a comparação de perdas. Wilcoxon e Hodges–Lehmann aparecem como análises complementares. |
| Cronograma | O escopo reduzido está incorporado à linha de base; o estresse incide sobre todas as 14 entregas, inclusive E14. A distinção entre preparação e período de aulas está explicitada. |
| Experimento com LLM | Há modelo único, condição sem LLM e caráter exploratório declarado para H2b. Não é necessário reintroduzir comparações com diversos provedores. |
| Auditoria ancorada | CA2 e a seção de auditoria delimitam corretamente a proteção aos registros anteriores à última âncora externa. Resta uma inconsistência na tabela de ameaças, detalhada adiante. |

Esses pontos não devem voltar a ser tratados como pendências gerais. As ressalvas abaixo dizem respeito ao que ainda permanece na versão atual.

## 2. Correções necessárias antes da entrega

### 2.1. Prioridade alta — unificar a regra estatística escrita e a implementada

**Evidências:** `Docs/monografia/Main.tex:646`, `Main.tex:647`, `Codigo/scripts/pilot_backtesting.py:71`, `pilot_backtesting.py:97` e `Docs/adr/ADR-003-parametrizacao-pert-beta-piso.md:14`.

A monografia define a moda por HSM, mas o código calcula essa moda apenas sobre observações maiores que o piso de 0,25, quando existem pelo menos três delas. Essa seleção não é explicitada na definição matemática. O efeito é material nos próprios dados utilizados:

| Janela | HSM sobre toda a amostra com piso | HSM utilizado pelo script |
| --- | ---: | ---: |
| v2.40.0 | 0,25 | 5,2351 |
| v2.41.0 | 0,25 | 0,8629 |
| v2.46.0, janela ampliada | 0,25 | 0,8629 |

Não se trata de diferença de arredondamento. O leitor que aplicar HSM diretamente à amostra descrita obterá outra parametrização. A exclusão da massa no piso pode ser uma decisão metodológica, mas exige definição e justificativa: o resultado passa a representar a moda estimada no subconjunto acima do piso.

Há uma segunda divergência. O texto exige fallback quando `tp <= to`; o ADR exige dispersão superior a 0,05; o script força `tp = max(to + 0.05, Q95)`, sem executar esse fallback. Para uma entrada de quinze durações iguais a 1,0, o resultado observado foi:

```text
regra: primaria_quantis_regularizados_Q05_HSM_Q95
to = 1.0; tm = 1.0; tp = 1.05; esperança PERT = 1.0083
```

A amostra constante ganha artificialmente variância e média maior que 1,0, contrariando a regra de fallback descrita.

**Correção recomendada:** decidir e registrar uma única regra no texto, no ADR e no script. Explicitar o conjunto usado pelo HSM e a condicional para amostras com pouca dispersão; testar amostra constante acima do piso e os limites de dispersão. Se a regra implementada mudar, regenerar o piloto e as tabelas afetadas. Não basta alterar o nome do procedimento.

**Critério de encerramento:** a parametrização pode ser reproduzida apenas com a definição da monografia, e os casos de fronteira respeitam essa definição.

### 2.2. Prioridade alta — atualizar o teste que ainda exige a árvore antiga

**Evidência:** `Codigo/scripts/test_pilot_backtesting.py:19`, especialmente linhas 30 e 38.

A suíte Python executada apresentou **8 testes aprovados e 1 erro**, entre 9 testes. O teste `test_classify_three_level_decision_tree` ainda acessa `q10_q90_applicable`, chave removida na simplificação para dois níveis:

```text
KeyError: 'q10_q90_applicable'
```

**Correção recomendada:** atualizar o teste para o contrato de dois níveis, incluindo os casos de fronteira do item 2.1. Não restaurar o ramo obsoleto apenas para satisfazer o teste, nem remover ou ignorar o teste sem substituir suas verificações úteis.

**Critério de encerramento:** `python -m unittest discover -s Codigo/scripts -p 'test_*.py' -v` termina sem erros e verifica a regra adotada no texto.

### 2.3. Prioridade média — tornar a reprodução e a rastreabilidade numérica coerentes

**Evidências:** `Codigo/checks/check_numbers.py:241`, `README.md:130`, `Docs/monografia/Main.tex:981`, `Main.tex:1294` e a tabela de poder em `Main.tex:699`.

Após regenerar os resultados do código atual em cópia temporária, o auditor numérico falhou com:

```text
[ERRO] Valor chave '75,4' ausente em README.md
```

O auditor exige que a mediana de desempenho recém-medida apareça literalmente no README. Tempos de execução variam entre máquinas e execuções; a monografia informa medições em Windows/i5-8400, enquanto a reprodução foi feita em Linux. Essa diferença, por si só, não invalida o benchmark original. O problema é exigir igualdade textual entre medições distintas, sem distinguir o registro publicado de uma nova reprodução.

Também obtive **91,6% de cobertura**, enquanto README, monografia e ADRs afirmam **94,1%**. A cobertura observada supera a meta de 80%; não há evidência de reprovação desse requisito. Entretanto, falta vincular o percentual publicado a um comando, ambiente e revisão reproduzíveis, ou atualizá-lo se estiver desatualizado.

Finalmente, a tabela de poder contém valores escritos diretamente em `Main.tex`, apesar de `CONSTRAINTS.md`, C2, determinar exportação automatizada de todos os resultados. Os valores de poder conferidos estão corretos nesta revisão; a pendência é de rastreabilidade e prevenção de divergências futuras.

**Correção recomendada:** separar o benchmark publicado, identificado por ambiente e revisão, da verificação de desempenho de uma nova execução; comparar cada artefato com sua fonte correspondente. Registrar a evidência da cobertura anunciada. Exportar também os valores da tabela de poder e limitar a mensagem de sucesso do auditor às verificações que ele efetivamente realiza.

**Critério de encerramento:** o procedimento documentado funciona em uma cópia limpa, sem exigir a transcrição manual de tempos locais para o README; números publicados têm origem identificável.

### 2.4. Prioridade média — completar a definição dos regimes sintéticos R7 e R9

**Evidência:** `Docs/monografia/Main.tex:711`, principalmente linhas 716 e 717.

R7 define a topologia quase serial, mas não explicita a distribuição das durações. R9 define o fator latente normal e a correlação, mas não especifica a distribuição marginal nem sua transformação para duração. Um fator latente normal, isoladamente, não determina uma duração PERT, log-normal ou de outra família.

O protocolo também precisa explicitar a origem, o tamanho e a separação dos dados históricos usados pelo baseline de reamostragem na Dimensão 2. A informação disponível para cada método deve estar definida antes dos resultados. A separação de módulos evita compartilhamento de implementação, mas não basta para assegurar igualdade de informação entre métodos.

**Correção recomendada:** acrescentar uma definição compacta por regime com distribuição marginal, riscos, correlação e parâmetros visíveis aos estimadores. Manter explícita a independência nominal já declarada na seção de limitações. Definir como o histórico sintético do baseline é gerado sem consultar o desfecho avaliado.

**Critério de encerramento:** outro pesquisador consegue implementar os seis regimes e as entradas dos baselines sem escolher premissas estatísticas adicionais. Isso exige completar a especificação do TCC1, não executar antecipadamente os experimentos do TCC2.

### 2.5. Prioridade média — ajustar afirmações residuais de segurança e de alcance empírico

**Evidências:** `Docs/monografia/Main.tex:888`, `Main.tex:950` e `Main.tex:1294`.

A tabela STRIDE ainda apresenta como risco residual da auditoria apenas o comprometimento simultâneo do SGBD e da chave de ancoragem. Ela deve incluir também alterações no trecho posterior à última âncora, conforme já reconhecido em CA2 e na seção de auditoria. Essa janela existe mesmo sem comprometimento da chave externa.

O texto também chama NFKC de normalização “canônica” e lhe atribui o colapso de homógrafos visuais. NFKC é uma normalização por compatibilidade; não elimina todos os caracteres visualmente confundíveis. Por exemplo, o `a` latino e o `а` cirílico continuam diferentes após NFKC. A distinção está documentada no [Unicode UAX #15](https://www.unicode.org/reports/tr15/) e os mecanismos específicos para caracteres confundíveis são tratados no [Unicode UTS #39](https://www.unicode.org/reports/tr39/).

Na conclusão, a expressão “demonstrou a regra causal em dois níveis” é mais forte que a evidência do piloto atual. O piloto parametriza amostras já extraídas, sem executar o replay temporal por cortes que ficou reservado ao TCC2.

**Correção recomendada:** sincronizar a tabela STRIDE com CA2; descrever o alcance real de NFKC; substituir a frase da conclusão por “demonstrou exploratoriamente a parametrização em dois níveis; a avaliação temporal causal permanece prevista para o TCC2”. São correções de precisão, sem necessidade de ampliar a arquitetura.

### 2.6. Prioridade média — atualizar a auditoria bibliográfica

**Evidências:** `Docs/bib_audit.md:5` e `Docs/monografia/Referencias.bib:600` em diante.

A bibliografia possui 66 entradas citadas, mas a auditoria contém 61 registros, anuncia 50 no título e ainda relaciona três chaves removidas: `mcnemar1947`, `vite2024` e `w3c2015`. Oito referências atuais não constam dessa auditoria:

```text
abdelhamid1991       amin2027llmspm       herroelen2005
kellner1999          li2024finese        richardeau2024graphs
shrestha2025gsmranges                    tang2024grapharena
```

Isso não demonstra que as referências sejam falsas; demonstra que o registro de verificação está incompleto em relação à exigência C3 do projeto. Nesta revisão, confirmei título e autoria das três novas referências do arXiv: [GraphArena](https://arxiv.org/abs/2407.00379), [LLMs Prompted for Graphs](https://arxiv.org/abs/2409.00159) e [GSM-Ranges](https://arxiv.org/abs/2502.08680).

A referência de Amin et al. com ano 2027 também aparece na [página da editora, volume 244, artigo 113129](https://www.sciencedirect.com/science/article/pii/S0164121226003626). Portanto, o ano do fascículo, isoladamente, não é motivo para classificá-la como inexistente ou removê-la. Convém registrar a publicação antecipada e a versão consultada.

**Correção recomendada:** atualizar a auditoria para as 66 entradas atuais, verificar os registros faltantes em suas fontes primárias e corrigir os links locais Windows obsoletos. Para preprints revisados, identificar a versão consultada. Não alegar verificação integral apenas porque todas as chaves BibTeX são resolvidas.

## 3. Defeito adicional da PoC a corrigir antes da reutilização do motor

**Prioridade média — IDs de risco duplicados induzem dependência não declarada.**

Em `Codigo/poc-engine/model.go:155`, `ValidateNetwork` não rejeita IDs de risco vazios ou duplicados. Em `Codigo/poc-engine/simulator.go:136`, a semente do risco depende de `RISK:` mais seu ID. Dois riscos com o mesmo ID, ambos aceitos pela validação, recebem o mesmo sorteio.

Reproduzi o caso em cópia temporária: uma tarefa constante de duração 10 e dois riscos de impacto 1, probabilidade 0,5, com 10.000 iterações e semente 42.

| Identificadores | F=10 | F=11 | F=12 |
| --- | ---: | ---: | ---: |
| Duplicados: R e R | 4.929 | 0 | 5.071 |
| Distintos: R e R2 | 2.415 | 5.048 | 2.537 |

O primeiro caso transforma dois riscos supostamente independentes em eventos perfeitamente dependentes. A solução é exigir IDs únicos e não vazios na validação do motor e cobrir essa condição em teste. **Não identifiquei esse defeito nos resultados publicados**, cujas redes usam identificadores distintos; por isso, não o trato como invalidação dos números da monografia. Ele deve ser corrigido antes de aceitar redes externas no TCC2, preferencialmente ainda na PoC.

## 4. Avaliação acadêmica e de gestão do projeto

**Coerência e objetivos.** A integração entre simulação de cronogramas, sincronização de eventos e assistência de IA com aprovação humana tem uma justificativa identificável. As questões de pesquisa, hipóteses e entregas estão melhor vinculadas. O resultado do TCC1 deve continuar sendo apresentado como arquitetura especificada e núcleo demonstrado; melhoria efetiva da gestão de projetos permanece uma hipótese a investigar.

**Fundamentação teórica.** A combinação de literatura clássica de planejamento e simulação com trabalhos recentes sobre LLMs é pertinente. A comparação com trabalhos científicos ajuda a sustentar a contribuição. A seleção declarada é representativa; não precisa ser convertida em revisão sistemática para justificar este projeto. A pendência concreta é completar a rastreabilidade bibliográfica e manter as afirmações proporcionais ao que cada fonte estudou.

**Metodologia.** A DSRM e a divisão em dimensões são apropriadas. O dimensionamento exato de H1b e a classificação exploratória de H2b são avanços. O principal ponto estatístico restante é alinhar a regra de parametrização ao código. O poder de 83,93% corresponde ao cenário p=0,85 por regime; a própria sensibilidade mostra 47,96% em p=0,83 e 66,85% em p=0,87. Não apresentar 83,93% como poder garantido em toda a faixa de equivalência.

**Escopo.** Preservaria os cortes já feitos: dois níveis de parametrização, seis regimes, um modelo LLM, uma estratégia de ancoragem e interface mínima. Não recomendo adicionar estudos de UX, múltiplos modelos ou integração completa ao TCC1. O risco atual está mais na execução e integração do TCC2 do que na falta de funcionalidades previstas.

**Viabilidade.** A reprodução do planejamento resultou nos seguintes valores, coerentes com os números publicados:

| Cenário | P50, semanas | P85, semanas | P95, semanas | Frequência simulada F ≤ 20 |
| --- | ---: | ---: | ---: | ---: |
| Nominal, riscos e rho=0,50 | 12,52 | 13,76 | 14,53 | 100,0% |
| Estresse +25%, riscos e rho=0,50 | 15,52 | 17,00 | 17,86 | 100,0% |
| Estresse +40%, riscos e rho=0,50 | 17,32 | 18,93 | 19,88 | 95,9% |

Esses valores demonstram consistência do modelo implementado; não comprovam empiricamente a duração futura das entregas. Em especial, 100,0% observado em 10.000 simulações não significa ausência de risco real.

Como melhoria de gestão, recomendo registrar a base das estimativas de 12–18 horas nos pacotes mais exigentes, especialmente E02, E07, E08, E09 e E10: componentes reaproveitados, trabalho remanescente e critério de pronto. Não é necessário produzir outra planilha extensa; algumas linhas por pacote permitem avaliar se a reserva resulta de estimativas fundamentadas. Confirmar também o prazo administrativo de defesa e a infraestrutura efetivamente disponível antes do início do TCC2.

## 5. O que pode ser removido ou simplificado

- Remover do teste as expectativas do antigo terceiro nível e do casamento Q10/Q90, preservando verificações relevantes para a regra atual.
- Eliminar registros bibliográficos obsoletos e links para caminhos de outra máquina na auditoria.
- Considerar mover para um apêndice o histórico de alternativas estatísticas descartadas. No corpo, preservar a regra escolhida, sua justificativa e a análise de sensibilidade; o leitor precisa compreender o método final.
- Reduzir repetições extensas de tecnologias e controles entre resumo, contribuição e conclusão. Na conclusão, priorizar o que foi demonstrado, o que continua pendente e os limites da evidência.
- Evitar frases absolutas como “100% dos números sincronizados” quando o auditor verifica apenas um conjunto de condições. Enumerar o alcance da verificação é mais defensável.

Essas simplificações editoriais são recomendações, não novas exigências de pesquisa.

## 6. Verificações efetivamente realizadas

Ambiente local: Linux/amd64, Go `1.27.3-X:nodwarf5` e Python `3.14.9`. A diferença para o ambiente de benchmark informado na monografia deve ser considerada ao comparar tempos.

| Verificação | Resultado |
| --- | --- |
| `go test -count=1 -v -cover ./...`, em `Codigo/poc-engine` | 22 testes aprovados; cobertura 91,6%. |
| Suíte Python, `unittest discover` | 9 testes executados: 8 aprovados e 1 erro por contrato antigo do piloto. |
| `extract_dataset.py --verify-local` | Aprovado; estatísticas locais e invariantes de duração conferidas. |
| `go run .`, em cópia temporária | Executado; resultados centrais de CPM, Monte Carlo, CRN e cronograma reproduzidos. |
| `power_h1b.py`, em cópia temporária | Executado; tabela de poder e erros Tipo I conferidos. |
| `pilot_backtesting.py`, em cópia temporária | Executado; classificação exploratória reproduzida, com as divergências metodológicas do item 2.1. |
| `check_numbers.py`, após regeneração | Reprovado pela exigência de transcrever no README o benchmark local de 75,4 ms. |
| `check_latex.py` | Aprovado nas verificações estáticas: 69 labels, 66 referências citadas, resumo de 244 palavras e abstract de 220 palavras. |
| Mutação temporária da condição que ativa a cópula | Detectada por `TestTCC2StressModesUniformIncludingE14`; arquivo temporário restaurado. Evidência favorável à capacidade de detectar essa regressão. |
| Casos adicionais de parametrização e IDs de risco | Divergências reproduzidas, conforme itens 2.1 e 3. |
| Inspeção das imagens do Gantt e histograma/CDF | Realizada nos PNGs existentes; não substitui inspeção dessas figuras diagramadas no PDF. |

## 7. Condições para encerrar a revisão

1. Resolver a definição e os casos de fronteira da parametrização, com texto, ADR e código consistentes.
2. Atualizar os testes Python e obter a suíte aprovada.
3. Corrigir a distinção entre resultados publicados e reprodução local no auditor; documentar a cobertura e automatizar os números de poder.
4. Completar as definições dos regimes e das entradas dos baselines.
5. Aplicar os ajustes pontuais de segurança, conclusão e auditoria bibliográfica.
6. Compilar e inspecionar o PDF final, verificando tabelas, figuras, referências, quebras e exigências institucionais.

**Decisão final: adequado como projeto de TCC1 e próximo da entrega, mas ainda não apto para entrega final na revisão examinada.** A liberação depende dessas correções e da conferência do documento compilado; a existência de pendências não exige reformular o tema ou aumentar o escopo.

Limites da revisão: não havia PDF compilado nem ferramenta LaTeX disponível neste ambiente; não foi possível avaliar a diagramação final. Não refiz a coleta remota integral do GitHub nem a auditoria externa das 66 fontes. Os controles do sistema integrado e os experimentos reservados ao TCC2 não foram executados.
