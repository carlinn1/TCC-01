# Especificação Técnica de Correções e Ajustes (SPEC) — TCC 1
**Projeto:** Sombra Digital Governada para Gestão de Projetos de Software  
**Data da Especificação:** 11 de novembro de 2026  
**Status:** Aprovado para Execução  

---

## 1. Veredito e Diretriz Geral

Os pareceres do Codex e do Claude apresentam convergência superior a 90% nos pontos de falha do trabalho. O projeto possui excelente rigor metodológico e arquitetural, mas apresenta **4 bloqueadores de reprodutibilidade e consistência teórica** que impedem sua submissão imediata.

### Principais Diagnósticos Confirmados:
1. **Falha na Análise de Contingência do Cronograma:** O teste de estresse de $+25\%$ não foi aplicado aos pacotes reduzidos no código Go (`main.go`), gerando um valor artificial de $P85 = 19{,}82$ semanas. Com o estresse uniforme, obtém-se $P85 = 20{,}33$ semanas e $P(F \le 20 \text{ semanas}) = 74{,}18\%$.
2. **Incoerência Empírica e Validade do Piloto:** O extrator Python produziu 5 registros em que a duração de engenharia é inferior ao tempo de revisão do PR, violando a regra $t_{\text{start}} = \min(PR_{\text{open}}, \text{commit}_{\text{first}})$. Além disso, o script de piloto utiliza dados retrospectivos ($t_c = y_{\text{real}} / 3$) e ajusta hiperparâmetros com a própria janela testada.
3. **Colisão de Sementes no Motor de Monte Carlo:** O algoritmo de derivação `SplitMix64` em `random.go` gera sequências idênticas de amostras ordenadas para as sementes mestres 42 e 43.
4. **Divergência Numérica entre Artefatos:** Há desalinhamento entre benchmarks medidos no Resumo (PT), Abstract (EN), Seção 5.6 e tabelas (ex.: $6{,}24 \text{ ms}$ vs $5{,}31 \text{ ms}$).

---

## 2. Pacote de Trabalho 1 (PK-01): Extração de Dados Empíricos e Replay Causal

### 2.1 Reconciliação do Extrator (`extract_dataset.py`)
- **Problema:** Inconsistência de temporalidade em 5 registros de `dataset_backtesting.json` onde a duração bruta é menor que a revisão.
- **Ação:**
  1. Alterar `extract_dataset.py` para aplicar a restrição estrita:
     $$t_{\text{eng\_bruta}} \ge t_{\text{revisao\_pr}}$$
  2. Implementar a coleta formal do campo `issue_assigned_at` quando disponível na API do GitHub.
  3. Adicionar asserção em tempo de extração para descartar ou auditar commits de *squash* cujas datas distorçam o tempo de início da tarefa.
  4. Atualizar o comando de verificação local `--verify-local` para auditar a coerência temporal item a item (não apenas agregados).

### 2.2 Reclassificação e Validação Causal do Piloto (`pilot_backtesting.py`)
- **Problema:** O piloto retrospectivo atual consome a duração final $y_{\text{real}}$ e total de itens conhecidos no futuro.
- **Ação:**
  1. **Reclassificação no Texto:** Alterar o status do piloto no TCC 1 para *Demonstração Exploratória de Parametrização Estocástica*, reservando a execução do Replay Causal Integrado para o TCC 2.
  2. **Ajuste de Código:** Remover a contaminação de dados futuros no ajuste de parâmetros de $v2.40.0$. Garantir que os hiperparâmetros da distribuição PERT-Beta de uma janela $k$ dependam exclusivamente do histórico observado em $k-1, k-2, \dots$.
  3. **Correção de Métrica:** Substituir a utilização da duração total da janela $y_{\text{real}}$ por $y_{\text{congelado}}$, alinhando o script à definição metodológica da Seção 3.3.2.

---

## 3. Pacote de Trabalho 2 (PK-02): Motor Estocástico, PRNG e Dimensionamento Estatístico

### 3.1 Correção da Derivação de Sementes (`random.go` / `simulator.go`)
- **Problema:** As sementes mestres 42 e 43 resultam em amostragens idênticas devido à sobreposição de bits no `SplitMix64(S0 XOR k)`.
- **Ação:**
  1. Reformular o gerador em `random.go` para incorporar a chave de réplica $r$ e a chave do worker $w$ em fluxos não correlacionados:
     $$\text{Seed}_{\text{eff}} = \text{Hash64}(S_0 \mathbin{\Vert} r \mathbin{\Vert} w)$$
  2. Criar teste unitário em `simulator_test.go` que verifique que o vetor ordenado de términos para $S_0 = 42$ e $S_0 = 43$ possui distância euclidiana $\|X_{42} - X_{43}\| > \epsilon$.

### 3.2 Alinhamento da Tabela de Erro Tipo I e TOST (`Docs/Main.tex`)
- **Problema:** As colunas de erro Tipo I publicadas na Tabela 11 não correspondem à probabilidade binomial exata da região de aceitação de Wilson.
- **Ação:**
  1. Atualizar o script de dimensionamento estatístico para exportar as probabilidades binomiais exatas do teste TOST sob as fronteiras $p = 0{,}80$ e $p = 0{,}90$.
  2. Para $M = 500$, fixar o critério de decisão binomial exato aceito ($\alpha = 0{,}05$), onde a região de aceitação é $416 \le k \le 438$ sucessos, garantindo poder estatístico de $83{,}93\%$ para $p = 0{,}85$.

### 3.3 Reformulação do Teste de Hipótese H1a
- **Problema:** O teste de Wilcoxon assinado avalia postos de diferenças pareadas e não comprova diretamente a redução das perdas médias (MAE/RMSE) sob caudas pesadas.
- **Ação:**
  1. Redefinir a inferência de H1a para utilizar a diferença pareada das perdas absolutas e quadráticas médias com intervalo de confiança obtido via *Bootstrap Pareado* ($B = 10.000$ reamostragens por grafo).
  2. Manter o teste de Wilcoxon como análise descritiva secundária de localização de mediana.
  3. Aplicar a correção de multiplicidade de Holm-Bonferroni delimitada rigorosamente por regime sintético.

---

## 4. Pacote de Trabalho 3 (PK-03): Cronograma do TCC 2, Contingência e Calendário Institucional

### 4.1 Harmonização do Teste de Estresse Uniforme (`main.go` e Tabela 26)
- **Problema:** O fator de estresse $+25\%$ foi removido dos pacotes E08–E11 após a simulação de corte, criando uma falsa sensação de convergência para $P85 < 20$ semanas.
- **Ação:**
  1. Modificar `main.go` para aplicar o multiplicador de $1{,}25$ de forma uniforme sobre todas as estimativas pessimistas dos pacotes, inclusive após a incidência dos cortes de contingência.
  2. **Substituição de Corte:** Remover a proposta de corte do pacote E11 (que possui $IC = 0\%$ e folga de $1{,}58$ semanas) e incluir a redução do escopo de um pacote no caminho crítico (ex.: E06 ou E07).
  3. Recalcular os cenários de cronograma. Atualizar os valores no texto:
     - Cenário de Estresse com Contingência: refletir o valor real recalibrado ($P85 \approx 20{,}1 \text{ a } 20{,}3 \text{ semanas}$).
     - Atualizar Resumo, Abstract, Seção 6.2.1, Seção 6.5 e Figura 8.

### 4.2 Ajuste do DAG de Pacotes de Trabalho e Calendário UnB/FCTE
- **Problema:** A folga total $TF(E01) = TF(E04) = 0{,}42$ semanas no texto diverge da topologia apresentada na Tabela 25. O calendário letivo do texto (08/03/2027 a 23/07/2027) diverge do calendário oficial UnB (15/03/2027 a 16/07/2027).
- **Ação:**
  1. Documentar na Tabela 25 todas as arestas $FS$ do DAG utilizadas pelo motor Go para eliminar discrepâncias de folga calculada.
  2. Corrigir as datas institucionais do semestre 2027/1 na monografia segundo a Resolução CEPE nº 0151/2026:
     - Período de aulas: 15/03/2027 a 16/07/2027 (18 semanas úteis).
  3. Fixar o pacote de defesa (E13) como evento de data fixa ancorado no calendário acadêmico, e alargar as estimativas pessimistas dos pacotes de integração (E05/E06).

---

## 5. Pacote de Trabalho 4 (PK-04): Segurança, Arquitetura e Avaliação de LLM

### 5.1 Propagação dos Limites de Auditoria e Segurança (ADR-007 e CA2)
- **Problema:** O texto do TCC 1 promete detecção imediata de qualquer adulteração, enquanto o ADR-007 reconhece o risco residual de janelas não ancoradas ($K = 100$ eventos ou $15 \text{ min}$).
- **Ação:**
  1. Reformular o Critério de Aceitação CA2 e as conclusões da Seção 4.6 para restringir formalmente a garantia de integridade aos registros já ancorados externamente.
  2. Corrigir a terminologia na monografia: utilizar "Chave Secreta Simétrica" para HMAC e "Chave Privada de Assinatura" para Ed25519.
  3. Ajustar o esquema da tabela `audit_logs` no banco de dados para incluir a coluna `project_id`, alinhando o escopo do lock transacional (`pg_advisory_xact_lock`) ao cálculo da cadeia de hashes.

### 5.2 Redes Docker e Protocolo de Avaliação de LLM (H2a / H2b)
- **Problema:** A topologia de rede Docker (`ai-isolated-net`) apresenta inconsistência de roteamento na documentação, e a afirmação de que $\text{temperature} = 0$ garante $100\%$ de determinismo é conceitualmente incorreta para APIs comerciais.
- **Ação:**
  1. Corrigir o diagrama e o texto da topologia Docker: explicitar que o serviço *Core* atua como ponte de aplicação (dual-homed) entre a rede interna de isolamento do banco e a rede de saída com controle de egress.
  2. Remover a alegação de determinismo absoluto para $\text{temperature} = 0$. Adicionar nota explicitando a variabilidade residual do provedor.
  3. **Reformulação de H2b:** Alterar o baseline de comparação de H2b para incluir uma regra determinística baseada estritamente nas métricas estocásticas sem LLM (ex.: ranking por Índice de Criticidade), mensurando o ganho efetivo de explicabilidade.

---

## 6. Pacote de Trabalho 5 (PK-05): Automação de Verificação e Sincronização de Texto

### 6.1 Correção e Extensão do Script `check_numbers.py`
- **Problema:** O script atual valida apenas matrizes de tabelas LaTeX e não detecta incoerências entre o Abstract, Resumo e Seção 5.6.
- **Ação:**
  1. Expandir `check_numbers.py` para extrair e validar tokens numéricos presentes nas seguintes seções:
     - Abstract em Inglês e Resumo em Português.
     - Texto narrativo da Seção 5.6 (ganhos de velocidade e tempos em milissegundos).
     - Tabelas 19, 23 e 26.
  2. Padronizar os tempos médios do motor em todas as menções do texto para os valores oficiais da PoC:
     $$\text{Benchmark PoC:} \quad 5{,}31 \text{ ms (10 tasks)}, \quad 61{,}76 \text{ ms (50 tasks)}, \quad 61{,}50 \text{ ms (100 tasks)}, \quad 241{,}49 \text{ ms (200 tasks)}$$
  3. Atualizar a referência da versão do Go para Go 1.27 uniformemente em todo o documento.

### 6.2 Ajustes Visuais e de Leitura
- **Ação:**
  1. Ajustar o texto explicativo da Figura 5 para refletir com exatidão a sequência de 8 etapas representadas.
  2. Atualizar a legenda da Figura 7 para mencionar explicitamente a presença do Histogram e da ECDF.

---

## 7. Pacote de Trabalho 6 (PK-06): Normas ABNT, Siglas e Bibliografia

### 7.1 Padronização ABNT de Referências Bibliográficas (`Docs/Referencias.bib`)
- **Ação:**
  1. Normalizar os nomes de autores em caixa-alta/caixa-baixa no BibTeX (ex.: `Brooks Jr.`, `Van Slyke`, `Beurer-Kellner`).
  2. Adicionar as entradas bibliográficas faltantes citadas no texto:
     - Wilson (1927), Clopper-Pearson (1934), McNemar (1947), Holm (1979), Hodges-Lehmann (1963), Cohen (1960), Kahan (1965), RFC 8032, RFC 2104.
  3. Atualizar as citações de motivação de mercado, incorporando dados recentes de gestão de projetos de software (ex.: relatórios Flyvbjerg / Reference Class Forecasting).

### 7.2 Lista de Siglas e Redução de Escopo Narrativo
- **Ação:**
  1. Incluir na Lista de Siglas do documento: STRIDE, XSS, CSP, DoD, DLQ, SGBD, LGPD, MdAE e CRN.
  2. Condensar o Resumo/Abstract para no máximo 250 palavras, removendo a detalhação excessiva do cronograma dos autores.
  3. Remover a expressão "Gêmeo Digital de Processo" das palavras-chave, mantendo a padronização em "Sombra Digital de Processo".

---

## 8. Matriz de Aceite e Rastreabilidade das Correções

| Item | Descrição da Correção | Pacote | Critério de Validação |
|---|---|---|---|
| **B1** | Uniformizar estresse de $+25\%$ no cronograma Go | PK-03 | `go run main.go` gera $P85$ coerente nos cenários com e sem corte |
| **B2** | Sincronizar benchmarks no Abstract, PT e Seção 5.6 | PK-05 | `python3 check_numbers.py` retorna 0 erros em 100% dos arquivos |
| **B3** | Reclassificar piloto e alinhar métrica $y_{\text{congelado}}$ | PK-01 | Script Python sem vazamento de dados futuros; texto alinhado |
| **B4** | Reduzir escopo TCC 2 e ajustar calendário UnB | PK-03 / PK-06 | Calendário 15/03/2027 a 16/07/2027; pacotes com faixas PERT realistas |
| **P1** | Corrigir derivação de sementes `SplitMix64` | PK-02 | Teste unitário confirma divergência de trajetórias para $S_0=42$ e $S_0=43$ |
| **P2** | Ajustar tabela de erro Tipo I do TOST | PK-02 | Valores binomiais exatos recalculados e exportados para LaTeX |
| **P3** | Delimitar auditoria aos registros ancorados (CA2) | PK-04 | Texto do TCC 1 e ADR-007 com premissas de segurança idênticas |
| **P4** | Normalizar bibliografia ABNT e siglas | PK-06 | 100% das referências com metadados auditados e siglas compiladas |
