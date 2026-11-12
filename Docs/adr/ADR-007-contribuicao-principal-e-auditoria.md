# ADR-007: Declaração da Contribuição Principal e Modelo de Custódia da Âncora de Auditoria

## Status
Accepted

## Date
2026-11-11

## Context
Dois pontos de posicionamento arquitetural exigem formalização explícita:
1. O texto não deve sugerir que técnicas clássicas (CPM, PERT-Beta, aprovação humana ou *hash chaining*) são invenções do trabalho, devendo concentrar a tese em uma contribuição principal clara;
2. Na trilha de auditoria (`SEG-03`), assinar a âncora externa com uma chave simétrica HMAC armazenada no próprio `Core PM Service` protege contra adulteração restrita ao banco PostgreSQL, mas não contra um atacante que comprometa simultaneamente o processo do `Core` e sua chave em memória.

## Decision
1. **Contribuição Principal Única (`DEC-7`):** definir na Seção 1.3 a contribuição central do trabalho como a **"Arquitetura de integração governada entre estado do projeto, simulação estocástica em rede e assistência interpretativa por IA, com verificação experimental do núcleo estocástico em Go"**, acompanhada de um quadro contrastando o que é **conhecido na literatura** versus o que é **contribuição de integração deste trabalho**;
2. **Modelo de Custódia e Escopo da Auditoria (`DEC-4`, `SEG-03`):**
   - Padronizar a ancoragem externa de auditoria (`audit_anchors`) exclusivamente por assinatura assimétrica **Ed25519 (RFC 8032)** (onde o `Core PM Service` detém a chave privada de assinatura e o verificador externo independente retém apenas a chave pública, eliminando a variante simétrica HMAC para ancoragem e reservando **HMAC-SHA256 / RFC 2104** estritamente para autenticação de entrada de *webhooks* do GitHub);
   - Fixar o intervalo de ancoragem externa em **a cada $K = 100$ eventos ou a cada $\Delta t = 15$ minutos (e obrigatoriamente a cada aprovação de Categoria 2)**;
   - Isolar cadeias por projeto incluindo `\texttt{project\_id}_k` no encadeamento $H_k$, na mensagem assinada da âncora $M_k = (\texttt{project\_id}_k \,\|\, k \,\|\, \texttt{timestamp}_k \,\|\, H_k)$ e na restrição `UNIQUE (project_id, prev_hash)` do PostgreSQL;
   - Delimitar o critério **CA2**: qualquer mutação ou exclusão retroativa (`UPDATE`/`DELETE`) em `audit_logs` sobre registros **anteriores à última âncora externa exportada** é detectada em 100% dos casos; registros gerados na janela ainda não ancorada ou sob comprometimento simultâneo do host do `Core` e do armazenamento externo são declarados como risco residual fora do escopo do controle.

## Consequences
- Posiciona com precisão científica a contribuição de Engenharia de Software do TCC e torna o modelo de ameaças da trilha de auditoria tecnicamente rigoroso.

