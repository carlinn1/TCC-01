# Constraints & Quality Contract (TCC 1 — Sombra Digital Governada)

Last reviewed: 2026-11-13 — caminhos de reprodução atualizados; limites C1–C9 preservados.

## 1. Floor & Project Restrictions (C1–C9 — Always Enforced)

- **C1 (Continuidade científica):** Não reiniciar o trabalho. Preservar tema, DSRM, distinção clara entre TCC1 (especificação + PoC do núcleo estocástico + piloto/extração de dados) e TCC2 (implementação integrada dos serviços e avaliação empírica das Dimensões 2–5), PoC em Go e governança de IA.
- **C2 (Fonte única de números):** Nenhum número de resultado experimental em tabela, figura, resumo ou conclusão pode ser digitado à mão; todos derivam do export automatizado do motor em Go (`Codigo/data/poc_results.json`, `Docs/monografia/resultados_poc.tex`, `Docs/monografia/tabelas/*.tex`), do extrator de dados (`Codigo/data/dataset_backtesting.json`), do estudo de poder (`Codigo/scripts/power_h1b.py`) e do registro de cobertura (`Codigo/scripts/export_coverage.py`), verificados por `Codigo/checks/check_numbers.py`.
- **C3 (Verificação em fonte primária):** Nenhuma afirmação ou citação nova sem fonte primária verificada (autoria completa, título, ano, veículo, URL/DOI/ISBN documentados em `Docs/bib_audit.md`).
- **C4 (Fronteira de escopo TCC1/TCC2):** Não implementar o sistema web integrado no TCC1 (escopo do TCC2). No TCC1 são permitidos: correções e testes no motor da PoC (`Codigo/poc-engine/*.go`), scripts de extração/piloto/poder/figuras (`Codigo/scripts/*.py`), verificadores (`Codigo/checks/*.py`) e ajustes de especificação e texto (`Docs/`).
- **C5 (Propagação integral de mudanças estatísticas):** Qualquer alteração estatística deve ser propagada para todos os pontos do documento: Resumo/Abstract, §1.2 (RQ4, H1a, H1b, H2a, H2b), §3.3, §3.4, Tabela 8, §3.5 e §6.5, verificada por busca textual automatizada em `Codigo/checks/check_latex.py`.
- **C6 (Regeneração obrigatória):** Alterou código, semente ou parâmetro $\implies$ regenerar todos os resultados (`cd Codigo/poc-engine && go test -count=1 -v ./... && go run .`) e figuras (`python Codigo/scripts/generate_figures.py`) antes de concluir alterações no texto.
- **C7 (Governança de decisões de escopo):** As decisões `DEC-1` a `DEC-7` são formalizadas como Architecture Decision Records (`Docs/adr/ADR-001..007.md`).
- **C8 (Não-regressão de cobertura e desempenho):** A cobertura de testes de `Codigo/poc-engine/` não pode cair abaixo da linha de base registrada em `Docs/baseline/baseline_cover.txt` (`>= 64.0%`, meta `>= 80.0%`), e o teste de desempenho reprova caso a mediana da rede sintética de 200 tarefas ultrapasse o orçamento de 500 ms do RNF06(i).
- **C9 (Precisão terminológica de validação):** Onde há apenas verificação em escopo de Prova de Conceito, utilizar *"verificado/demonstrado no escopo da PoC"* para o motor e *"especificado/projetado para o TCC2"* para os serviços integrados, jamais atribuindo garantias universais irrestritas.
- **LGPD/GDPR:** Zero logins (`author_login`, `merged_by`), nomes pessoais de terceiros ou e-mails em arquivos JSON versionados (`Codigo/data/dataset_backtesting.json`).

## 2. Automated Gates (Executados automaticamente a cada tarefa)

| Dimensão | Regra / Limiar | Comando Verificador (`Checked by`) | Tipo |
| :--- | :--- | :--- | :--- |
| **Motor PoC (Go)** | 100% dos testes unitários, analíticos, KS exato ($10^6$), mutação Bernoulli, CRN e reprodutibilidade (1, 4 e 16 workers) verdes; cobertura $\ge 80\%$ | `python Codigo/scripts/export_coverage.py` (executa `go test -count=1 -v -cover ./...` e registra ambiente/cobertura) | Automatizado |
| **Orçamento RNF06(i)** | Mediana de 5 execuções para 200 tarefas / 300 dependências ($N=10.000$) $< 500\text{ ms}$ | `cd Codigo/poc-engine && go test -count=1 -run TestBenchmarkDispersion5Runs ./...` | Automatizado |
| **Export Único (C2)** | `Codigo/data/poc_results.json`, `Docs/monografia/resultados_poc.tex` e `Docs/monografia/tabelas/*.tex` sincronizados com `Docs/monografia/Main.tex` e `generate_figures.py` | `python Codigo/checks/check_numbers.py` | Automatizado |
| **Dataset & Piloto (LGPD)** | `Codigo/data/dataset_backtesting.json` contém observações minimizadas por item, recalcula a Tabela 7 offline e possui zero PII | `python Codigo/scripts/extract_dataset.py --verify-local && python Codigo/scripts/pilot_backtesting.py` | Automatizado |
| **Poder Estatístico H1b** | Script reproduz exatamente a tabela de poder para $M \in [200, 1000]$ e sensibilidade $p \in \{0{,}83; 0{,}85; 0{,}87\}$ | `python Codigo/scripts/power_h1b.py` | Automatizado |
| **Integridade LaTeX, BibTeX & Figuras** | Zero chaves/ambientes desbalanceados, 100% de citações/labels válidas, zero frases proibidas (C5/C9), pasta Overleaf isolada e 8 figuras PNG RGB opacas (200 DPI) | `python Codigo/checks/check_latex.py` | Automatizado |

## 3. Manual / Visual Inspections (Quando o PDF compilado estiver disponível)

| Dimensão | Critério de Inspeção | Como Verificar | Tipo |
| :--- | :--- | :--- | :--- |
| **Inspeção Visual de Páginas Críticas** | Verificar ausência de cabeçalhos colados ou tabelas cortadas nas Tabelas 2, 7, 8, 13, 15, 17, 19 e 20 e nas Figuras 3, 4, 5, 6, 7 e 8 | Enviar `Docs/monografia/` no Overleaf / TeX Live e inspecionar visualmente as páginas das tabelas e figuras (além da checagem de `.log` por `python Codigo/checks/check_latex.py --log output.log`) | Manual / Visual |
