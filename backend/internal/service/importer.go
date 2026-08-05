package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"medmasia/backend/internal/core"
	"medmasia/backend/internal/store"

	"github.com/xuri/excelize/v2"
)

type ImportResult struct {
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors"`
}

func ImportXLSX(ctx context.Context, st *store.Store, path string, maxRows int) (ImportResult, error) {
	if strings.ToLower(filepath.Ext(path)) != ".xlsx" {
		return ImportResult{}, fmt.Errorf("only .xlsx import is supported")
	}
	f, err := excelize.OpenFile(path)
	if err != nil {
		return ImportResult{}, err
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return ImportResult{}, fmt.Errorf("workbook has no sheets")
	}
	rows, err := f.Rows(sheets[0])
	if err != nil {
		return ImportResult{}, err
	}
	defer rows.Close()

	if !rows.Next() {
		return ImportResult{}, fmt.Errorf("workbook has no header row")
	}
	headerRow, err := rows.Columns()
	if err != nil {
		return ImportResult{}, err
	}
	headers := indexHeaders(headerRow)
	result := ImportResult{}
	rowNumber := 1
	for rows.Next() {
		rowNumber++
		if maxRows > 0 && result.Imported >= maxRows {
			break
		}
		cols, err := rows.Columns()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: %v", rowNumber, err))
			result.Skipped++
			continue
		}
		p := prospectFromRow(headers, cols, rowNumber)
		if p.Name == "" {
			result.Skipped++
			continue
		}
		score := ScoreProspect(p)
		p.Score, p.Grade, p.ScoreReasons = score.Score, score.Grade, score.Reasons
		if err := RejectClinicalContent(map[string]string{
			"name":        p.Name,
			"specialty":   strings.Join(p.Specialties, " "),
			"city":        p.City,
			"crm":         p.CRM,
			"linkedinUrl": p.LinkedInURL,
		}); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: %v", rowNumber, err))
			result.Skipped++
			continue
		}
		if _, err := st.UpsertProspect(ctx, p); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: %v", rowNumber, err))
			result.Skipped++
			continue
		}
		result.Imported++
	}
	return result, rows.Error()
}

func prospectFromRow(headers map[string]int, row []string, rowNumber int) core.Prospect {
	name := CleanText(cell(headers, row, "nome", "name", "medico", "médico"))
	crm := CleanText(cell(headers, row, "crm", "crm_medico", "numero_crm", "n_crm"))
	crmUF := strings.ToUpper(CleanText(cell(headers, row, "crm_uf", "uf_crm", "uf_do_crm")))
	uf := strings.ToUpper(CleanText(cell(headers, row, "uf", "estado")))
	specialties := UniqueNonEmpty(
		cell(headers, row, "especialidade", "especialidade_1", "especialidade1", "specialty", "specialty_1"),
		cell(headers, row, "especialidade_2", "especialidade2", "specialty_2"),
		cell(headers, row, "especialidade_3", "especialidade3", "specialty_3"),
	)
	emails := UniqueNonEmpty(
		NormalizeEmail(cell(headers, row, "email", "email_1", "email1")),
		NormalizeEmail(cell(headers, row, "email_2", "email2")),
		NormalizeEmail(cell(headers, row, "email_3", "email3")),
	)
	phones := UniqueNonEmpty(
		NormalizePhone(cell(headers, row, "telefone", "telefone_1", "telefone1", "phone", "phone1", "celular")),
		NormalizePhone(cell(headers, row, "telefone_2", "telefone2", "phone2")),
		NormalizePhone(cell(headers, row, "telefone_3", "telefone3", "phone3")),
	)
	meta := professionalMetadata(headers, row)
	sourceRef := crm
	if sourceRef == "" {
		sourceRef = fmt.Sprintf("row-%d-%s", rowNumber, strings.ToLower(strings.ReplaceAll(name, " ", "-")))
	}
	return core.Prospect{
		Source:               "medicos_enriquecida_xlsx",
		SourceRef:            sourceRef,
		Name:                 name,
		CRM:                  crm,
		CRMUF:                crmUF,
		Specialties:          specialties,
		City:                 CleanText(cell(headers, row, "cidade", "municipio", "município", "city")),
		UF:                   uf,
		Emails:               emails,
		Phones:               phones,
		LinkedInURL:          CleanText(cell(headers, row, "linkedin", "linkedin_url", "url_linkedin")),
		ProfessionalMetadata: meta,
		Status:               "new",
	}
}

func professionalMetadata(headers map[string]int, row []string) map[string]any {
	meta := map[string]any{}
	for key := range headers {
		switch key {
		case "n_locais_trabalho", "locais_trabalho", "n_proprietario_cooperado", "consultorio_isolado",
			"clinica_centro_de_especialidade", "policlinica", "funcionarios", "funcionarios_estimados",
			"faturamento", "faturamento_estimado", "renda", "renda_presumida", "cnpjs_ativos":
			if value := CleanText(cell(headers, row, key)); value != "" {
				meta[key] = parseMetaValue(value)
			}
		}
	}
	return meta
}

func indexHeaders(headers []string) map[string]int {
	out := map[string]int{}
	for i, header := range headers {
		key := normalizeHeader(header)
		if key != "" {
			out[key] = i
		}
	}
	return out
}

func cell(headers map[string]int, row []string, aliases ...string) string {
	for _, alias := range aliases {
		if idx, ok := headers[normalizeHeader(alias)]; ok && idx < len(row) {
			return row[idx]
		}
	}
	return ""
}

func normalizeHeader(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "ã", "a", "â", "a",
		"é", "e", "ê", "e",
		"í", "i",
		"ó", "o", "õ", "o", "ô", "o",
		"ú", "u",
		"ç", "c",
		" ", "_", "-", "_", ".", "_", "/", "_",
	)
	value = replacer.Replace(value)
	for strings.Contains(value, "__") {
		value = strings.ReplaceAll(value, "__", "_")
	}
	return strings.Trim(value, "_")
}

func parseMetaValue(value string) any {
	lower := strings.ToLower(strings.TrimSpace(value))
	switch lower {
	case "sim", "true", "yes":
		return true
	case "nao", "não", "false", "no":
		return false
	}
	normalized := strings.ReplaceAll(value, ".", "")
	normalized = strings.ReplaceAll(normalized, ",", ".")
	if n, err := strconv.ParseFloat(normalized, 64); err == nil {
		return n
	}
	return value
}
