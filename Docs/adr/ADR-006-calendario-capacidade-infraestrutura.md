# ADR-006: Calendário Real, Capacidade de Dedicação e Infraestrutura Computacional do TCC2

## Status
Accepted

## Date
2026-11-11

## Context
Para que a simulação de Monte Carlo do cronograma do TCC2 (Seção 6.2) seja auditável pela banca, é necessário fixar as datas-calendário das 20 semanas, a disponibilidade semanal de horas de cada autor frente às demais disciplinas da graduação na FCTE/UnB e os requisitos de hardware para execução do modelo de linguagem.

## Decision
1. **Calendário Letivo alinhado à Resolução CEPE/UnB nº 0151/2026 (20 semanas úteis):** Semana 1 iniciando em **01/03/2027** (segunda semana do período letivo geral oficial `23/02/2027 a 20/07/2027`, aproveitando 2 semanas preparatórias pré-aulas para `E01/E02`) e Semana 20 encerrando em **16/07/2027** (término exato das aulas de 2027/1 segundo a Resolução CEPE/UnB nº 0151/2026: **15/03/2027 a 16/07/2027**, 18 semanas / 100 dias letivos), com marcos de monitoramento quinzenal nas Semanas 4, 8, 12 e 16;
2. **Capacidade por Autor:** dedicação de **12 horas úteis semanais por autor** (24 h/semana combinadas; $\approx 480$ homens-hora totais ao longo das 20 semanas), compatível com a matrícula simultânea em disciplinas obrigatórias/optativas de final de curso;
3. **Infraestrutura de Modelo Único em LLM (`E10`, `DEC-2`):** execução exclusiva com **um único modelo de pesos abertos da família Llama-3** (`Llama-3.1-8B-Instruct` quantizado em GPU local de 12–16 GB VRAM via Ollama/vLLM, ou `Llama-3.3-70B-Instruct` via servidor dedicado imutável com cache em disco) em regime determinístico de decodificação (`temperature = 0.0`), excluindo comparações com APIs comerciais externas e o lote exploratório adicional em `temperature = 0.2`;
4. **Reclassificação Exploratória de `H2b` e Escopo Humano (`DEC-1`):** no pacote `E10`, a análise qualitativa cega da subamostra de 50 cenários é conduzida pelos dois autores ($\kappa$ de Cohen $\ge 0{,}80$) como **análise exploratória descritiva**, sem teste confirmatório de McNemar, sem 3º avaliador externo (`D03` removido) e com estudos de usabilidade, carga cognitiva ou viés de automação com usuários finais declarados explicitamente fora do escopo do TCC1 e do TCC2.

## Consequences
- Elimina o risco de gargalo de hardware (VRAM para 70B local) e de dependência externa no caminho crítico do TCC2.

