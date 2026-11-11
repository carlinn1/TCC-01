> **AVISO HISTÓRICO (LEGADO):** Este documento é uma versão histórica inicial da proposta e **diverge do texto principal oficial (Docs/Main.tex)** em título, autoria e numeração de requisitos. Mantido em Docs/legado/ apenas para registro de evolução.

# SPEC DE CORREÇÃO INTEGRAL DO TCC 1 (Código Go + Python + Figuras + LaTeX/BibTeX)

**Repositório Alvo:** `https://github.com/carlinn1/TCC-01`  
**Objetivo:** Corrigir 100% dos erros de compilação LaTeX/Overleaf, sincronizar os códigos Go/Python da PoC com o texto e fechar todas as 19 pendências apontadas pela banca revisora.

---

## PARTE 1 — Correção Imediata dos Erros de Compilação LaTeX (`main.tex`)

Aplique obrigatoriamente as 6 correções abaixo no `main.tex` para eliminar o timeout do Overleaf e zerar todos os erros e warnings de compilação:

### 1.1. Erro fatal no TikZ da Figura 8 (`No shape named 'Slides'`, `Undefined color '19'`, `Undefined color 'gray '` e `Float too large by 499pt`)
* **Causa:** No `\foreach \row/\lbl/\s/\e/\col in { ... }` da Figura 8 (`fig:gantt`), o texto `E13 (Buffer/Slides)` possui uma barra `/` que corrompe o parser de tuplas do TikZ, e o último item `gray` possui espaço residual antes de `}`.
* **Correção:**
  1. Nunca usar `/` dentro do campo `\lbl` do `\foreach` (substituir `Buffer/Slides` por `Buffer e Slides` ou proteger com chaves `{E13 (Buffer e Slides)}`).
  2. Colocar `%` imediatamente após o último item do `\foreach` (`0.5/D01 (Grafo UI)/14/17/gray%`) para evitar espaço no nome da cor.
  3. Bloco corrigido para o fallback da Figura 8:
  ```latex
  \figura{fig08_gantt.png}{Diagrama de Gantt do TCC2 ao longo das 20 semanas, destacando a antecipação do piloto do DAG (E01) nas semanas 1--3 e a escrita contínua (E12) das semanas 4 a 18.}{fig:gantt}{0.95}{%
  \begin{tikzpicture}[x=0.50cm,y=0.34cm, font=\tiny]
    \draw[->, thick] (0,0) -- (21,0) node[right]{Semanas};
    \foreach \w in {1,2,...,20} {
      \draw[gray!30] (\w,0) -- (\w,17.5);
      \node[below] at (\w-0.5, 0) {\w};
    }
    \foreach \row/\lbl/\s/\e/\col in {
      17/{E01 (Piloto DAG)}/0/3/unbgreen,
      16/{E02 (Docker+SQL)}/0/3/unbblue,
      15/{E03 (CRUD+RBAC)}/2/5/unbblue,
      14/{E04 (Webhooks)}/3/6/unbblue,
      13/{E05 (Sim Worker)}/4/7/unbblue,
      12/{E06 (AI Service)}/6/9/unbblue,
      11/{E07 (Fail-Closed)}/7/10/unbblue,
      10/{E08 (Dim 2: Sint.)}/6/10/orange!85!black,
      9/{E09 (Dim 3: Real)}/8/12/orange!85!black,
      8/{E10 (Dim 4-5: Seg.)}/10/13/orange!85!black,
      7/{E11 (UI React)}/10/14/unbblue,
      6/{E12 (Redacao Cont.)}/3/18/unbgreen,
      5/{E13 (Buffer e Slides)}/17/19/unbblue,
      4/{E14 (Defesa)}/19/20/unbblue,
      2.5/{D02 (Sensibilidade)}/13/15/gray,
      1.5/{D03 (Juizes Kappa)}/14/16/gray,
      0.5/{D01 (Grafo UI)}/14/17/gray%
    } {
      \node[left, font=\tiny] at (0, \row) {\lbl};
      \fill[\col!75, draw=\col!95!black, rounded corners=1pt] (\s, \row-0.30) rectangle (\e, \row+0.30);
    }
    \draw[dashed, thick, red!75!black] (0, 3.25) -- (20.5, 3.25) node[right, font=\bfseries\tiny]{Linha de Corte};
  \end{tikzpicture}%
  }
  ```

### 1.2. Erro `Command \' invalid in math mode` (Linhas 624, 773 e similares)
* **Causa:** Uso de caracteres acentuados (`í`, `á`, `é`) dentro de `\mathrm{...}` em modo matemático (ex.: `$t_{\mathrm{início}}$`).
* **Correção:** Buscar todas as ocorrências de `\mathrm{início}` (ou qualquer `\mathrm{...}` com acento) no `main.tex` e substituir por `\text{início}` (do pacote `amsmath`) ou `\mathrm{inicio}`.
  * Exemplo: trocar `$t_{\mathrm{início}}$` por `$t_{\text{início}}$` em todo o arquivo.

### 1.3. Warning `Package babel Warning: Name 'brazil' is deprecated`
* **Correção:**
  * Na linha 1 do documento, trocar `\documentclass[12pt,openany,oneside,a4paper,brazil,sumario=tradicional]{abntex2}` por `\documentclass[12pt,openany,oneside,a4paper,brazilian,sumario=tradicional]{abntex2}`.
  * Logo após `\begin{document}`, trocar `\selectlanguage{brazil}` por `\selectlanguage{brazilian}`.

### 1.4. Warning `Package hyperref Warning: Token not allowed in a PDF string`
* **Causa:** Uso de `$\rightarrow$` diretamente dentro de `\section{...}` (ex.: `\section{Matriz de Rastreabilidade (Requisitos $\rightarrow$ Pacotes de Trabalho)}`).
* **Correção:** Substituir `$\rightarrow$` dentro de comandos `\chapter`, `\section`, `\subsection` por `\texorpdfstring{$\rightarrow$}{->}`.

### 1.5. Prevenção de Timeout no Overleaf (Geração Estática das Figuras PNG)
* O plano gratuito do Overleaf tem limite curto de CPU. Executar o script `generate_figures.py` para gerar os 8 arquivos PNG dentro da pasta `./figuras/` do repositório:
  * `figuras/fig01_dag_exemplo.png`
  * `figuras/fig02_pert_beta.png`
  * `figuras/fig03_niveis_gemeo_digital.png`
  * `figuras/fig04_hash_chain.png`
  * `figuras/fig05_arquitetura.png`
  * `figuras/fig06_sequencia.png`
  * `figuras/fig07_histograma_cdf_poc.png`
  * `figuras/fig08_gantt.png`
* Com os arquivos `.png` presentes na pasta `figuras/`, a macro `\IfFileExists{figuras/#1}{\includegraphics...}{...}` carrega diretamente as imagens pré-renderizadas em milissegundos, evitando compilar TikZ pesado no Overleaf.

---

## PARTE 2 — Ajustes nos Códigos do Repositório (Go e Python)

### 2.1. Inspecionar e atualizar o pacote Go da PoC
1. Verificar todos os arquivos `.go` (`random.go`, `cpm.go`, `simulator.go`, `engine_test.go`, `validation_test.go` e arquivo de structs/modelos).
2. **Caso degenerado ($t_o = t_m = t_p$):**
   * Garantir que `SampleBeta` / `RunMonteCarlo` trate `tp - to <= 1e-12` retornando duração constante `to` (sem divisão por zero e sem `NaN`).
   * Garantir que `TestDegeneratePertBetaZeroVariance` esteja em `validation_test.go` e passe no `go test`.
3. **Teste de Aderência de Kolmogorov-Smirnov (KS):**
   * Garantir que `TestPertBetaKSGoodnessOfFit` esteja implementado em `validation_test.go` (comparando $n=10^6$ amostras de T1, T2, T3 e T4 contra a CDF Beta regularizada exata $I_x(\alpha, \beta)$ via frações contínuas de Lentz, verificando $D_n < D_{\mathrm{crit}, 0.05} = 0.001358$).
4. **Topologia da Rede Sintética (200 tarefas, 300 dependências) e Dispersão do Benchmark (5 repetições):**
   * Verificar como o benchmark de 200 tarefas e 300 dependências está implementado em `engine_test.go` / `validation_test.go`.
   * Garantir que a descrição da topologia no `main.tex` (Seção de Benchmark: número de camadas/tarefas, distribuição das durações e semente) reflita **exatamente** o código de benchmark do repositório.
   * Executar `go test -v ./...` e `go test -bench=Benchmark -benchmem -count=5` e atualizar na Tabela de Momentos/KS e na Tabela de Benchmark do `main.tex` os números exatos medidos (mediana, média $\pm$ desvio-padrão e intervalo mín–máx das 5 repetições).

### 2.2. Sincronizar `extract_dataset.py` e `dataset_backtesting.json`
1. No `extract_dataset.py` original, faltava a release `"v2.46.0": "2024-03-20T17:08:24Z"` em `RELEASES` e a tupla `("v2.40.0", "v2.46.0")` em `CASES`, embora ela estivesse no `dataset_backtesting.json`.
2. Atualizar `extract_dataset.py` para incluir `v2.46.0` e documentar/contabilizar o filtro de `reason:completed` (excluindo `reason:not_planned` e PRs fechados sem merge), mantendo total consistência com a Tabela de Eventos GitHub/CI do texto.

---

## PARTE 3 — Checklist das 19 Correções da Banca para Verificação Final

Antes de finalizar o commit, faça uma varredura nos arquivos `main.tex`, `referencias.bib` e códigos garantindo que **nenhum** dos 19 pontos abaixo tenha regredido:

1. **[ ] Acabamento visual (Figuras 1–8 e Folha de Aprovação):**
   * Pasta `./figuras/` populada com os 8 PNGs gerados por `generate_figures.py` + fallback TikZ sem erros de sintaxe.
   * Folha de aprovação sem placeholders (`DIA de MÊS` ou `Convidado 1/2`), preenchida com data e nomes de professores da UnB/FCTE.
2. **[ ] Seções mudas e citação de tabelas:**
   * Nenhuma seção pode conter apenas título seguido imediatamente de tabela sem texto introdutório/analítico.
   * Todas as tabelas do documento (especialmente RF, RNF e Eventos GitHub/CI) devem ser explicitamente citadas com `\ref{...}` e discutidas no corpo do texto.
3. **[ ] BibTeX (`referencias.bib`) 100% compatível com `abntex2cite`:**
   * `{Brooks Jr.}, Frederick P.` (gera `BROOKS JR.`).
   * `{Van Slyke}, Richard M.` (gera `VAN SLYKE`).
   * Organizações protegidas com chaves duplas: `{{The Standish Group}}`, `{{Project Management Institute}}`, `{{International Organization for Standardization}}`, `{{Open Web Application Security Project}}`, `{{Vite Consortium}}`, `{{World Wide Web Consortium}}`.
   * Nenhum uso de `et al.` literal no campo `author` (todos os coautores listados nominalmente para eliminar o erro `"e. a."` na lista de referências).
   * Citação oficial do PostgreSQL 16 (`postgresql2024`) no lugar de `momjian2001` para sustentar `FOR UPDATE SKIP LOCKED`.
   * Citação do AgentDojo (`debenedetti2024`) atualizada para **NeurIPS 2024** (*Datasets and Benchmarks Track*) com `arXiv:2406.13352`.
4. **[ ] Estrutura de Capítulos (DSRM):**
   * Cap. 1: Introdução; Cap. 2: Revisão Bibliográfica; Cap. 3: Metodologia (com seção explícita mapeando as **6 atividades da DSRM** de Peffers et al., requisitos, caracterização do backtesting, protocolo de 5 dimensões e Ameaças à Validade); Cap. 4: Arquitetura de Referência e Governança de IA; Cap. 5: Prova de Conceito e Resultados Preliminares; Cap. 6: Cronograma do TCC2 e Considerações Finais.
5. **[ ] Reprodutibilidade Real:**
   * Link oficial `https://github.com/carlinn1/TCC-01` citado no Capítulo 5 e no Apêndice B, acompanhado do hash real do commit de entrega.
6. **[ ] Coerência Taxonômica (Sombra Digital Governada vs. Gêmeo Digital):**
   * Título e texto alinhados à classificação de Kritzinger et al. (2018): o sistema opera como **Sombra Digital Governada de Processo** devido à barreira humana intencional (*human-in-the-loop*) no fluxo reverso.
7. **[ ] Regra *Fail-Closed* explícita:**
   * Declaração formal no texto e na Tabela de Governança de que qualquer `action_type` desconhecido, omitido ou fora do JSON Schema é negado por padrão (`REJECTED`), sem mutação de banco, gerando alerta e registro de violação na auditoria.
8. **[ ] Especificação da *Allowlist* e Coerência RNF02 $\leftrightarrow$ H2:**
   * Tabela explícita definindo os campos que chegam ao LLM (métricas numéricas + `task_title` sanitizado até 120 chars e `issue_pr_summary` até 280 chars) e os bloqueados (código-fonte, diffs, PII, e-mails, corpos brutos de commit), explicando que H2 testa injeção indireta exatamente nos títulos/resumos permitidos.
9. **[ ] Modelo de Ameaças (Persuasão do Aprovador e `CREATE_DRAFT_RISK`):**
   * Mitigação contra engenharia social/persuasão do gerente (cartão determinístico do servidor em destaque primário na UI mostrando $\Delta P_{50}, \Delta P_{85}, \Delta\mathrm{IC}_i$ e texto do LLM recolhido com aviso de fonte não confiável).
   * Contenção de `CREATE_DRAFT_RISK` em tabela segregada (`draft_risks`), cota máxima de 3 por versão de estado, texto puro sem links/HTML e peso zero no cálculo até aprovação de Categoria 2.
10. **[ ] Auditoria Resistente a Adulterações (*Tamper-Evident*) com Ancoragem Externa:**
    * Substituição do termo absoluto "imutável no mesmo banco" por trilha *tamper-evident* combinada à exportação periódica assinada do `head_hash` para armazenamento externo WORM/*append-only* fora do alcance do superusuário do PostgreSQL.
11. **[ ] Alinhamento OWASP LLM Top 10 (v1.1 2023 vs. v2025):**
    * Citação conjunta de `owasp2023` e `owasp2025`, explicando que *Prompt Injection* é LLM01 em ambas, *Sensitive Information Disclosure* passou de LLM06 (2023) para LLM02 (2025), e *Excessive Agency* passou de LLM08 (2023) para LLM06 (2025).
12. **[ ] Tabela de Eventos GitHub/CI e Progresso $c_i$:**
    * Tratamento explícito de `issues: closed (not_planned)` ($\rightarrow$ `CANCELLED`, fora da vazão) e `pull_request: closed (merged = false)` ($\rightarrow$ retorna a `IN_PROGRESS` com $c_i = \min(c_i; 0{,}50)$).
    * Explicação da fórmula de progresso por checklist $c_{\mathrm{chk}} = n_{\mathrm{concl}}/n_{\mathrm{total}}$ (que gera $c=0{,}30$ e $0{,}40$ no Cenário C) e justificativa/sensibilidade de $c_{\mathrm{rev}} = 0{,}80 \in \{0{,}70; 0{,}80; 0{,}90\}$.
13. **[ ] Matriz Comparativa de Ferramentas (Tabela 3):**
    * Atualizada para 2026 (Jira Plans + Rovo, Azure DevOps + Copilot, MS Project + Copilot), com fontes e datas de acesso em nota de rodapé, e linha do artefato separando o que foi **Validado em PoC (TCC1)** do que é **Projetado (TCC2)**.
14. **[ ] H1 Justa e Baseline de Vacanti (2015):**
    * Erro absoluto pontual (MAE/MdAE) avaliado com $P_{50}$ e média $E[F]$; calibração probabilística e quantil $P_{85}$ avaliados por cobertura empírica, *Pinball Loss* e CRPS; inclusão do Monte Carlo por vazão histórica (Vacanti, 2015) como baseline competitivo.
15. **[ ] Correções Metodológicas do Backtesting e Simulação Sintética:**
    * Zero vazamento temporal: $(t_o, t_m, t_p)$ parametrizados exclusivamente com itens concluídos antes do corte ($t_{\mathrm{fim}} \le t_c$).
    * Validade de construto: duração real medida desde $\min(\texttt{first\_commit\_at}, \texttt{assigned\_at}, \texttt{pr\_created\_at})$ até `\texttt{merged\_at}`.
    * Dados reais tratados como exploratórios (10 releases $\times$ 3 cortes = 30 avaliações) e dados sintéticos ($M=500$ DAGs) como confirmatórios não circulares (verdade gerada por Log-Normal, Weibull, Erlang e cópulas gaussianas).
16. **[ ] Dimensão 5 (Regra dos Três, Modelo/Temperatura) e Ameaças à Validade:**
    * 30 cenários $\times$ 10 repetições = 300 execuções (limite superior de alucinação $\le 1{,}0\%$ a 95% de confiança pela regra dos três, $3/300$), com verificador numérico automático e especificação de modelo/versão e `temperature` (`0.0` e `0.2`).
    * Seção dedicada a Ameaças à Validade (Construto, Interna, Conclusão Estatística e Externa).
17. **[ ] Validação Completa da PoC (Cap. 5):**
    * Teste KS ($D_n$ e *p-valor*), caso degenerado $t_o=t_p$, topologia detalhada da rede de 200 tarefas / 300 dependências e dispersão das 5 repetições do benchmark.
18. **[ ] Matriz de Rastreabilidade RF/RNF $\rightarrow$ Pacotes:**
    * Tabela explícita no Cap. 6 mapeando RF01–RF12 e RNF01–RNF09 aos pacotes E01–E14 (incluindo RF01/RF02, RF12 e RNF01 JWT/RBAC).
19. **[ ] Cronograma Reestruturado (Riscos nas Semanas 1–3 e Escrita Contínua):**
    * Pacote E01 nas Semanas 1–3 executando o piloto de construção do DAG e consolidando as 10 releases; Pacote E12 executando a redação contínua da monografia das Semanas 4 a 18.

---

## PARTE 4 — Validação Final Automatizada (Comandos que o Antigravity deve rodar)

Após aplicar todas as alterações, execute na ordem:

```bash
# 1. Gerar todas as 8 figuras PNG em ./figuras/
python3 generate_figures.py

# 2. Executar todos os testes unitários e estatísticos em Go
go test -v ./...

# 3. Executar os benchmarks com 5 repetições
go test -bench=Benchmark -benchmem -count=5

# 4. Compilar o PDF LaTeX e verificar ausência de erros
pdflatex -interaction=nonstopmode main.tex
bibtex main
pdflatex -interaction=nonstopmode main.tex
pdflatex -interaction=nonstopmode main.tex
```
