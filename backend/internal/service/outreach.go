package service

import (
	"fmt"
	"strings"

	"medmasia/backend/internal/core"
)

type OutreachSettings struct {
	FromName string
	Company  string
	BaseURL  string
	DryRun   bool
}

func DraftEmail(p core.Prospect, settings OutreachSettings) (subject, body string, err error) {
	firstName := firstToken(p.Name)
	specialty := firstNonEmpty(p.Specialties...)
	if specialty == "" {
		specialty = "gestao do consultorio"
	}
	city := firstNonEmpty(p.City, p.UF)
	subject = fmt.Sprintf("%s: IA aplicada a leads e contatos comerciais", firstName)
	body = fmt.Sprintf(`Ola %s,

Vi seu perfil profissional%s e estou estruturando a MedMasIA para ajudar medicos e clinicas a organizar contatos comerciais, responder interessados no site e priorizar oportunidades usando somente dados profissionais e comerciais.

Para perfis de %s, o objetivo e simples: reduzir perda de oportunidades, registrar conversas e deixar a equipe com uma fila clara de proximos contatos.

Faz sentido eu te enviar uma analise curta de como isso funcionaria no seu contexto comercial?

%s
%s

Para nao receber novas mensagens, responda "remover". Politica de privacidade: %s/privacidade.html
`, firstName, locationSuffix(city), specialty, settings.FromName, settings.Company, strings.TrimRight(settings.BaseURL, "/"))
	err = RejectClinicalContent(map[string]string{"subject": subject, "body": body})
	return subject, body, err
}

func ComplianceReview(channel, subject, body string, hasConsent bool) error {
	if err := RejectClinicalContent(map[string]string{"subject": subject, "body": body}); err != nil {
		return err
	}
	lower := strings.ToLower(body)
	if !strings.Contains(lower, "remover") && !strings.Contains(lower, "descadastrar") {
		return fmt.Errorf("missing_opt_out")
	}
	if channel == "whatsapp" && !hasConsent {
		return fmt.Errorf("whatsapp_requires_prior_opt_in")
	}
	return nil
}

func firstToken(value string) string {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return "Doutor(a)"
	}
	return fields[0]
}

func locationSuffix(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return " em " + CleanText(value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if v := CleanText(value); v != "" {
			return v
		}
	}
	return ""
}
