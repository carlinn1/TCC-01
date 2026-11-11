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
2. **Modelo de Custódia e Escopo da Auditoria (`SEG-03`):**
   - Corrigir a terminologia de `"chave privada HMAC"` para **chave secreta simétrica HMAC** (no MVP básico) ou **par de chaves assimétricas Ed25519** (onde o `Core` detém a chave privada de assinatura e o verificador externo retém apenas a chave pública);
   - Fixar o intervalo de ancoragem externa em **a cada $K = 100$ eventos ou a cada $\Delta t = 15$ minutos (e obrigatoriamente a cada aprovação de Categoria 2)**;
   - Delimitar o critério **CA2**: qualquer mutação ou exclusão retroativa (`UPDATE`/`DELETE`) em `audit_logs` sobre registros **anteriores à última âncora externa exportada** é detectada em 100% dos casos; registros gerados na janela ainda não ancorada ou sob comprometimento simultâneo do host do `Core` e do armazenamento externo são declarados como risco residual fora do escopo do controle.

## Consequences
- Posiciona com precisão científica a contribuição de Engenharia de Software do TCC e torna o modelo de ameaças da trilha de auditoria tecnicamente rigoroso.

