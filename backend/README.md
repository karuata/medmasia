# MedMasIA Sales Backend

Backend Go para capturar leads do site, importar a planilha privada de medicos, classificar prospects e criar uma outbox de contato auditavel. Ele adapta as ideias do BridgeBot e do `ai-sales-team-claude` para OpenAI Codex/MCP, PostgreSQL e operacao sem Python.

## Componentes

- `medmasia-server`: API HTTP para o chatbot do site e rotas admin.
- `medmasia-import`: importador `.xlsx` em Go, usando apenas campos profissionais/comerciais.
- `medmasia-mcp`: servidor MCP stdio para o Codex consultar e operar prospects.
- PostgreSQL: unico banco suportado.

## Variaveis

```powershell
$env:MEDMASIA_DATABASE_URL="postgres://user:password@localhost:5432/medmasia?sslmode=disable"
$env:MEDMASIA_ADMIN_TOKEN="troque-este-token"
$env:MEDMASIA_ADDR="127.0.0.1:8092"
$env:MEDMASIA_OUTREACH_DRY_RUN="true"
$env:MEDMASIA_OUTREACH_FROM_NAME="Rodrigo Masini"
$env:MEDMASIA_OUTREACH_COMPANY="MedMasIA"
$env:MEDMASIA_OUTREACH_BASE_URL="https://karuata.github.io/medmasia"
```

## Rodar

```powershell
go run .\cmd\medmasia-server
```

Health:

```powershell
Invoke-RestMethod http://127.0.0.1:8092/health
```

## Importar a Planilha Privada

A planilha deve continuar fora do Git. O `.gitignore` ja bloqueia `.xlsx`, `data/`, `private/`, `exports/` e `campaigns/`.

Validacao curta:

```powershell
go run .\cmd\medmasia-import -file ..\medicos_enriquecida_rmasini.xlsx -max-rows 25
```

Importacao completa:

```powershell
go run .\cmd\medmasia-import -file ..\medicos_enriquecida_rmasini.xlsx
```

## MCP Para Codex

Configure um servidor MCP local apontando para:

```powershell
go run C:\Users\rmasini\Personal\Experiments\medmasia\backend\cmd\medmasia-mcp
```

Ferramentas expostas:

- `medmasia.prospect_search`
- `medmasia.prospect_score`
- `medmasia.draft_outreach`
- `medmasia.campaign_summary`

## Guardrails

- Dados clinicos e dados de pacientes sao rejeitados.
- WhatsApp exige opt-in previo antes de qualquer envio real.
- Email exige opt-out no corpo.
- A outbox nasce em `draft`; nenhum provedor externo e acionado por padrao.
- GitHub Pages publica somente `index.html`, `privacidade.html`, `.nojekyll` e `assets/`.
