# Arquitetura MedMasIA Sales

## Origem

O BridgeBot original foi mantido intacto. A adaptacao feita aqui copia a ideia central: uma camada backend auditavel atras do site estatico, com captura de leads, trilha de eventos/agentes e controle operacional. A diferenca e que esta versao remove SQLite, Claude CLI e Python, usando Go, PostgreSQL e MCP para uso com Codex/OpenAI.

O repo `ai-sales-team-claude` foi usado como referencia conceitual, nao como runtime. Ele oferece skills, agentes e templates de vendas; aqui isso virou pipeline Go com classificacao, compliance e outbox.

## Fluxo

1. O site chama `/api/chat` e `/api/lead`.
2. Leads vindos do site sao gravados em `leads`.
3. A planilha privada e importada localmente via `medmasia-import`.
4. Prospects sao classificados por contato, especialidade, complexidade operacional, capacidade comercial e sinal digital.
5. O pipeline agêntico cria rascunho, revisa compliance e grava `touchpoints` em `draft`.
6. Codex pode consultar/acionar o sistema via `medmasia-mcp`.

## Agentes

- `prospect_loader`: carrega o registro profissional/comercial.
- `privacy_guard`: bloqueia qualquer dado clinico ou de paciente.
- `classifier`: aplica score 0-100 e grade A/B/C/D.
- `outreach_strategist`: cria primeiro email com contexto profissional.
- `compliance_reviewer`: exige opt-out e bloqueia WhatsApp sem consentimento.
- `dispatcher`: grava outbox em modo draft/dry-run.

## Banco

Tabelas principais:

- `prospects`
- `leads`
- `campaigns`
- `touchpoints`
- `suppression_list`
- `consent_events`
- `agent_runs`
- `agent_steps`

## Fronteiras de LGPD

Este backend nao foi desenhado para dado clinico. A base aceita apenas informacao profissional/comercial para prospeccao e relacionamento B2B/B2P. Qualquer integracao futura de envio deve preservar supressao, opt-out, consentimento de WhatsApp e logs de auditoria.
