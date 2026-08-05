package service

import (
	"fmt"
	"regexp"
	"strings"
)

var blockedClinicalTerms = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(paciente|pacientes|prontuario|prontu[aá]rio|diagn[oó]stico|cid|exame|receita|prescri[cç][aã]o)\b`),
	regexp.MustCompile(`(?i)\b(tratamento|cirurgia realizada|consulta m[eé]dica|intern[aã]o|hist[oó]rico cl[ií]nico)\b`),
}

func RejectClinicalContent(fields map[string]string) error {
	for key, value := range fields {
		text := strings.TrimSpace(value)
		if text == "" {
			continue
		}
		for _, re := range blockedClinicalTerms {
			if re.MatchString(text) {
				return fmt.Errorf("clinical_or_patient_data_rejected: field %s", key)
			}
		}
	}
	return nil
}

func NormalizeEmail(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if !strings.Contains(value, "@") {
		return ""
	}
	return value
}

func NormalizePhone(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if len(out) < 10 {
		return ""
	}
	if len(out) == 10 || len(out) == 11 {
		out = "55" + out
	}
	return out
}

func CleanText(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func UniqueNonEmpty(values ...string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		value = CleanText(value)
		if value == "" || seen[strings.ToLower(value)] {
			continue
		}
		seen[strings.ToLower(value)] = true
		out = append(out, value)
	}
	return out
}
