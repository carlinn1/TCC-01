# ADR-005: Escopo da Dimensão 3, Observações por Item e Piloto de Backtesting no TCC1

## Status
Accepted

## Date
2026-11-11

## Context
O vínculo `"PR fecha issue"` no GitHub (`closes #N`) representa associação entre artefatos, e não necessariamente uma relação de precedência técnica Término-a-Início ($FS$) entre tarefas distintas. Além disso, o arquivo `Codigo/data/dataset_backtesting.json` precisa preservar as observações individuais minimizadas por item (sem PII) para permitir recalcular todas as estatísticas da Tabela 7 e executar um piloto rastreável offline.

## Decision
1. **Observações por item sem PII (`DAD-01`):** estender `Codigo/scripts/extract_dataset.py` e `Codigo/data/dataset_backtesting.json` para armazenar o vetor `items` com registros pseudonimizados (`item_id`, timestamps UTC incluindo `issue_assigned_at`, durações de revisão e de engenharia desde o 1º commit pré-squash ou atribuição, fonte do 1º commit, flag `at_floor`, vínculo a issue e rótulos técnicos), sem logins ou dados pessoais;
2. **Regras explícitas para inferência de arestas $FS$ (`DAD-04`):** inferir dependências técnicas combinando: (i) menções explícitas de bloqueio no corpo de issues/PRs (`depends on #`, `blocked by #`, `requires #`), (ii) cadeias sequenciais de PRs que modificam o mesmo subsistema/rótulo com sobreposição zero ($t_{\mathrm{merge},u} \le t_{\mathrm{início},v}$) e (iii) validação manual de uma amostra de arestas pelos dois autores;
3. **Demonstração descritiva de viabilidade e parametrização causal (`DEC-3`, `DAD-05`):** no TCC1, `Codigo/scripts/pilot_backtesting.py` executa uma demonstração descritiva offline de parametrização estocástica em **2 níveis** sobre as janelas reais do `dataset_backtesting.json` (sem `y_real` sintético e sem vazamento temporal), reservando a Dimensão 3 do TCC2 ($n_{\mathrm{ef}} \ge 10$ releases no `cli/cli`) estritamente como **demonstração descritiva de viabilidade e parametrização causal** (sem pretensão de inferência preditiva/causal generalista; declara-se explicitamente o tempo ocioso entre *timestamps* de *commit*/PR e a validade de construto de *cycle time* como ameaça à validade), mantendo a Dimensão 2 como o núcleo confirmatório.

## Consequences
- Permite reprodução 100% offline da Tabela 7 (`cli/cli`) e da parametrização exploratória (Tabela 8) a partir de `Codigo/data/dataset_backtesting.json`, eliminando qualquer vazamento temporal ou erro de inferência de *cycle time* no TCC1.

