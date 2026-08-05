package service

import (
	"context"
	"fmt"

	"medmasia/backend/internal/config"
	"medmasia/backend/internal/core"
	"medmasia/backend/internal/store"
)

type AgentRunner struct {
	store    *store.Store
	settings config.Settings
}

func NewAgentRunner(st *store.Store, settings config.Settings) *AgentRunner {
	return &AgentRunner{store: st, settings: settings}
}

func (r *AgentRunner) RunProspectOutreach(ctx context.Context, prospectID, requestedBy string) (core.AgentRun, error) {
	runID, err := r.store.CreateAgentRun(ctx, requestedBy, "prospect_outreach")
	if err != nil {
		return core.AgentRun{}, err
	}
	terminal := "completed"

	p, err := r.store.GetProspect(ctx, prospectID)
	if err != nil {
		_ = r.store.AddAgentStep(ctx, runID, "prospect_loader", "failed", "Prospect nao encontrado.", map[string]any{"prospect_id": prospectID}, err.Error())
		return r.store.FinishAgentRun(ctx, runID, "failed", "Pipeline falhou ao carregar prospect.")
	}
	_ = r.store.AddAgentStep(ctx, runID, "prospect_loader", "done", "Prospect carregado do PostgreSQL.", map[string]any{"prospect_id": p.ID, "name": p.Name}, "")

	if err := RejectClinicalContent(map[string]string{"name": p.Name, "specialties": fmt.Sprint(p.Specialties)}); err != nil {
		_ = r.store.AddAgentStep(ctx, runID, "privacy_guard", "blocked", "Conteudo clinico/paciente bloqueado.", map[string]any{"prospect_id": p.ID}, err.Error())
		return r.store.FinishAgentRun(ctx, runID, "blocked", "Pipeline bloqueado por politica de dados.")
	}
	_ = r.store.AddAgentStep(ctx, runID, "privacy_guard", "done", "Somente dados profissionais/comerciais aprovados.", map[string]any{"policy": "no_clinical_or_patient_data"}, "")

	score := ScoreProspect(p)
	if err := r.store.UpdateProspectScore(ctx, p.ID, score.Score, score.Grade, score.Reasons); err != nil {
		_ = r.store.AddAgentStep(ctx, runID, "classifier", "failed", "Falha ao gravar score.", map[string]any{"score": score}, err.Error())
		return r.store.FinishAgentRun(ctx, runID, "failed", "Pipeline falhou no classificador.")
	}
	p.Score, p.Grade, p.ScoreReasons = score.Score, score.Grade, score.Reasons
	_ = r.store.AddAgentStep(ctx, runID, "classifier", "done", "Prospect classificado por fit comercial.", map[string]any{"score": score.Score, "grade": score.Grade, "reasons": score.Reasons}, "")

	subject, body, err := DraftEmail(p, OutreachSettings{
		FromName: r.settings.OutreachFromName,
		Company:  r.settings.OutreachCompany,
		BaseURL:  r.settings.OutreachBaseURL,
		DryRun:   r.settings.OutreachDryRun,
	})
	if err != nil {
		_ = r.store.AddAgentStep(ctx, runID, "outreach_strategist", "failed", "Rascunho reprovado por privacidade.", map[string]any{}, err.Error())
		return r.store.FinishAgentRun(ctx, runID, "failed", "Pipeline falhou no rascunho.")
	}
	_ = r.store.AddAgentStep(ctx, runID, "outreach_strategist", "done", "Email inicial personalizado criado.", map[string]any{"channel": "email", "subject": subject}, "")

	hasConsent := false
	if err := ComplianceReview("email", subject, body, hasConsent); err != nil {
		terminal = "blocked"
		_ = r.store.AddAgentStep(ctx, runID, "compliance_reviewer", "blocked", "Mensagem bloqueada por compliance.", map[string]any{"channel": "email"}, err.Error())
		return r.store.FinishAgentRun(ctx, runID, terminal, "Pipeline bloqueado por compliance.")
	}
	_ = r.store.AddAgentStep(ctx, runID, "compliance_reviewer", "done", "Mensagem contem opt-out e nao contem termos clinicos.", map[string]any{"channel": "email"}, "")

	status := "draft"
	if r.settings.OutreachDryRun {
		status = "draft"
	}
	touchpoint, err := r.store.QueueTouchpoint(ctx, core.Touchpoint{
		ProspectID: p.ID,
		Channel:    "email",
		Subject:    subject,
		Body:       body,
		Status:     status,
		Metadata: map[string]any{
			"dry_run":      r.settings.OutreachDryRun,
			"agent_run_id": runID,
			"grade":        p.Grade,
			"score":        p.Score,
		},
	})
	if err != nil {
		_ = r.store.AddAgentStep(ctx, runID, "dispatcher", "failed", "Falha ao gravar outbox.", map[string]any{}, err.Error())
		return r.store.FinishAgentRun(ctx, runID, "failed", "Pipeline falhou no dispatcher.")
	}
	_ = r.store.AddAgentStep(ctx, runID, "dispatcher", "done", "Touchpoint salvo na outbox; nenhum envio real foi executado.", map[string]any{"touchpoint_id": touchpoint.ID, "status": touchpoint.Status}, "")

	return r.store.FinishAgentRun(ctx, runID, terminal, "Pipeline de classificacao e contato concluido em modo seguro.")
}
