# Implementation Plan v4: Execução Integral da `SPEC_CORRECAO_TCC1.md`

## Overview
Execução completa das Fases 0 a 9 da `SPEC_CORRECAO_TCC1.md` (consolidação dos pareceres de Claude `[CL]`, Codex `[CX]` e Argon `[AR]`), cobrindo as restrições `C1–C9`, as decisões `DEC-1` a `DEC-7` (`Docs/adr/ADR-001..007.md`), o motor em Go via TDD (`COD-01..10`), o protocolo estatístico e poder de H1b (`EST-01..06`), o dataset item-a-item e piloto de backtesting (`DAD-01..06`), segurança/STRIDE/ER (`SEG-01..06`), cronograma/estresse/contingência do TCC2 (`CRO-01..06`), auditoria bibliográfica 100% (`BIB-01..08`), reestruturação textual (`TXT-01..11`), tabelas/figuras geradas do export (`FOR-01..05`) e gates automatizados (`REL-01..04`).

## Architecture Decisions (ADRs)
- **ADR-001 (`DEC-1`):** H1b avaliado por TOST ($\alpha=0{,}05 \equiv$ IC de Wilson a 90% contido em $[80\%, 90\%]$) com $M=500$ grafos por regime (poder de 86,2% em $p=0{,}85$).
- **ADR-002 (`DEC-2`):** Pares métrica $\leftrightarrow$ funcional consistentes segundo Gneiting (2011): $P_{50} \leftrightarrow \mathrm{MAE}$, $\bar{F} \leftrightarrow \mathrm{RMSE}$, $P_{85} \leftrightarrow \text{Pinball Loss}$, $\hat{F}_N \leftrightarrow \mathrm{CRPS}$.
- **ADR-003 (`DEC-3`):** Árvore de decisão de parametrização PERT-Beta com suporte primário $[Q_{05}, Q_{95}]$ + moda HSM, variante Eq. 3.2 para $Q_{10} > 0{,}25+\delta$ e *fallback* de *bootstrap* para massa no piso.
- **ADR-004 (`DEC-4`):** Cortes de contingência ordenados dentro dos pacotes essenciais (1 LLM essencial, `E11` tabular, `E09` 5 *releases*, `E08` 6 regimes nucleares), validados por simulação de estresse e corte.
- **ADR-005 (`DEC-5`):** Preservação de observações minimizadas por item em `dataset_backtesting.json`, regras explícitas de inferência de arestas $FS$ e piloto rastreável `Codigo/pilot_backtesting.py`.
- **ADR-006 (`DEC-6`):** Calendário letivo 2027/1 (08/03/2027 a 23/07/2027), 12 h/semana por autor, modelo aberto Llama-3 (8B local ou 70B via API com pesos fixados) e 2 juízes cegos ($\kappa$ de Cohen) em `E10`.
- **ADR-007 (`DEC-7`):** Contribuição principal única centrada na arquitetura de integração governada e custódia de auditoria com chave secreta simétrica HMAC (ancoragem a cada $K=100$ eventos ou 15 min) e alternativa Ed25519.
