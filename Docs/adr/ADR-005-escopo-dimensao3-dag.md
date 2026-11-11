# ADR-005: Escopo da Dimensão 3, Observações por Item e Piloto de Backtesting no TCC1

## Status
Accepted

## Date
2026-11-11

## Context
O vínculo `"PR fecha issue"` no GitHub (`closes #N`) representa associação entre artefatos, e não necessariamente uma relação de precedência técnica Término-a-Início ($FS$) entre tarefas distintas. Além disso, o arquivo `Codigo/data/dataset_backtesting.json` precisa preservar as observações individuais minimizadas por item (sem PII) para permitir recalcular todas as estatísticas da Tabela 7 e executar um piloto rastreável offline.

## Decision
1. **Observações por item sem PII (`DAD-01`):** estender `Codigo/scripts/extract_dataset.py` e `Codigo/data/dataset_backtesting.json` para armazenar o vetor `items` com registros pseudonimizados (`item_id`, timestamps UTC, durações de revisão e de engenharia desde o 1º commit, fonte do 1º commit, flag `at_floor`, vínculo a issue e rótulos técnicos), sem logins ou dados pessoais;
2. **Regras explícitas para inferência de arestas $FS$ (`DAD-04`):** inferir dependências técnicas combinando: (i) menções explícitas de bloqueio no corpo de issues/PRs (`depends on #`, `blocked by #`, `requires #`), (ii) cadeias sequenciais de PRs que modificam o mesmo subsistema/rótulo com sobreposição zero ($t_{\mathrm{merge},u} \le t_{\mathrm{início},v}$) e (iii) validação manual de uma amostra de arestas pelos dois autores;
3. **Piloto rastreável no TCC1 (`DAD-05`):** fornecer `Codigo/scripts/pilot_backtesting.py` que executa offline o replay causal nas janelas do `dataset_backtesting.json` sem vazamento temporal, mantendo a Dimensão 3 como exploratória/descritiva ($n_{\mathrm{ef}}=10$ releases) e a Dimensão 2 como confirmatória.

## Consequences
- Permite reprodução 100% offline da Tabela 7 e do piloto de backtesting a partir de `Codigo/data/dataset_backtesting.json`.

