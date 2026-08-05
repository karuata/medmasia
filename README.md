# MedMasIA

Static GitHub Pages site for MedMasIA: AI mentorship and digital twin products for physicians, medical managers, entrepreneurs and clinic owners.

## Local

```powershell
.\scripts\run-local.ps1
```

Open `http://127.0.0.1:8020/`.

## Sales Backend

The Go/PostgreSQL backend for lead capture, private XLSX import, prospect scoring, outbox drafts and MCP lives in `backend/`.

GitHub Pages deploys only the static site files. Private spreadsheets and exports stay ignored by Git.
