package service

import "testing"

func TestProspectFromRowKeepsProfessionalCommercialSubset(t *testing.T) {
	headers := indexHeaders([]string{
		"nome",
		"crm",
		"especialidade_1",
		"email_1",
		"telefone_1",
		"cidade",
		"uf",
		"data_nascimento",
		"faturamento",
	})
	p := prospectFromRow(headers, []string{
		"Maria Silva",
		"12345",
		"Dermatologia",
		"MARIA@EXAMPLE.COM",
		"(11) 99999-9999",
		"Sao Paulo",
		"SP",
		"1970-01-01",
		"1200000",
	}, 2)
	if p.Name != "Maria Silva" || p.CRM != "12345" {
		t.Fatalf("unexpected prospect: %#v", p)
	}
	if _, ok := p.ProfessionalMetadata["data_nascimento"]; ok {
		t.Fatal("birth date must not be imported into professional metadata")
	}
	if len(p.Emails) != 1 || p.Emails[0] != "maria@example.com" {
		t.Fatalf("emails = %#v", p.Emails)
	}
	if len(p.Phones) != 1 || p.Phones[0] != "5511999999999" {
		t.Fatalf("phones = %#v", p.Phones)
	}
}

func TestDraftEmailPassesCompliance(t *testing.T) {
	p := prospectFromRow(indexHeaders([]string{"nome", "especialidade_1", "email_1"}), []string{"Maria Silva", "Dermatologia", "maria@example.com"}, 2)
	subject, body, err := DraftEmail(p, OutreachSettings{FromName: "Rodrigo", Company: "MedMasIA", BaseURL: "https://example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if err := ComplianceReview("email", subject, body, false); err != nil {
		t.Fatal(err)
	}
}
