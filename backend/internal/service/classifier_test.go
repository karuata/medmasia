package service

import (
	"testing"

	"medmasia/backend/internal/core"
)

func TestScoreProspectGradesHighFitCommercialProspect(t *testing.T) {
	result := ScoreProspect(core.Prospect{
		Name:        "Maria Silva",
		Specialties: []string{"Dermatologia"},
		Emails:      []string{"maria@example.com"},
		Phones:      []string{"5511999999999"},
		LinkedInURL: "https://linkedin.com/in/maria",
		ProfessionalMetadata: map[string]any{
			"n_locais_trabalho":        3,
			"n_proprietario_cooperado": 1,
			"faturamento":              1200000,
		},
	})
	if result.Grade != "A" {
		t.Fatalf("grade = %s, want A; score=%d reasons=%v", result.Grade, result.Score, result.Reasons)
	}
}

func TestComplianceRejectsClinicalTerms(t *testing.T) {
	err := ComplianceReview("email", "Contato", "Tenho solucao para prontuario e pacientes. responder remover", false)
	if err == nil {
		t.Fatal("expected clinical content rejection")
	}
}

func TestComplianceRejectsWhatsAppWithoutConsent(t *testing.T) {
	err := ComplianceReview("whatsapp", "Contato", "Mensagem comercial. responder remover", false)
	if err == nil || err.Error() != "whatsapp_requires_prior_opt_in" {
		t.Fatalf("err = %v, want whatsapp_requires_prior_opt_in", err)
	}
}
