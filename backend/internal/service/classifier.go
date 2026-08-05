package service

import (
	"fmt"
	"math"
	"strings"

	"medmasia/backend/internal/core"
)

type ScoreResult struct {
	Score   int      `json:"score"`
	Grade   string   `json:"grade"`
	Reasons []string `json:"reasons"`
}

func ScoreProspect(p core.Prospect) ScoreResult {
	score := 0
	var reasons []string

	contact := 0
	if len(p.Emails) > 0 {
		contact += 12
	}
	if len(p.Phones) > 0 {
		contact += 8
	}
	score += contact
	if contact > 0 {
		reasons = append(reasons, fmt.Sprintf("contactability:%d", contact))
	}

	fit := specialtyFit(p.Specialties)
	score += fit
	if fit > 0 {
		reasons = append(reasons, fmt.Sprintf("specialty_fit:%d", fit))
	}

	operator := operatorComplexity(p.ProfessionalMetadata)
	score += operator
	if operator > 0 {
		reasons = append(reasons, fmt.Sprintf("operator_complexity:%d", operator))
	}

	capacity := commercialCapacity(p.ProfessionalMetadata)
	score += capacity
	if capacity > 0 {
		reasons = append(reasons, fmt.Sprintf("commercial_capacity:%d", capacity))
	}

	if strings.TrimSpace(p.LinkedInURL) != "" {
		score += 10
		reasons = append(reasons, "digital_signal:10")
	}

	if score > 100 {
		score = 100
	}
	return ScoreResult{Score: score, Grade: grade(score), Reasons: reasons}
}

func specialtyFit(values []string) int {
	highFit := []string{
		"dermat", "oftalmo", "gineco", "obst", "ortop", "cardio", "urolo",
		"otorrino", "cirurgia plast", "endocrino", "neurolog", "psiquiatr",
		"radiolog", "estet", "vascular", "reumatolog", "pediatr",
	}
	text := strings.ToLower(strings.Join(values, " "))
	for _, needle := range highFit {
		if strings.Contains(text, needle) {
			return 25
		}
	}
	if strings.TrimSpace(text) != "" {
		return 14
	}
	return 0
}

func operatorComplexity(meta map[string]any) int {
	points := 0
	points += boundedInt(metaNumber(meta, "n_locais_trabalho")*5, 0, 15)
	points += boundedInt(metaNumber(meta, "n_proprietario_cooperado")*5, 0, 10)
	for _, key := range []string{"consultorio_isolado", "clinica_centro_de_especialidade", "policlinica"} {
		if metaBool(meta, key) || metaNumber(meta, key) > 0 {
			points += 5
		}
	}
	return boundedInt(points, 0, 25)
}

func commercialCapacity(meta map[string]any) int {
	points := 0
	for _, key := range []string{"faturamento", "faturamento_estimado", "renda", "renda_presumida"} {
		value := metaNumber(meta, key)
		switch {
		case value >= 1000000:
			points += 12
		case value >= 300000:
			points += 8
		case value > 0:
			points += 4
		}
	}
	points += boundedInt(metaNumber(meta, "cnpjs_ativos")*4, 0, 8)
	return boundedInt(points, 0, 20)
}

func grade(score int) string {
	switch {
	case score >= 75:
		return "A"
	case score >= 55:
		return "B"
	case score >= 35:
		return "C"
	default:
		return "D"
	}
}

func metaBool(meta map[string]any, key string) bool {
	switch v := meta[key].(type) {
	case bool:
		return v
	case string:
		v = strings.ToLower(strings.TrimSpace(v))
		return v == "sim" || v == "true" || v == "1" || v == "yes"
	case float64:
		return v > 0
	case int:
		return v > 0
	default:
		return false
	}
}

func metaNumber(meta map[string]any, key string) int {
	v, ok := meta[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(math.Round(n))
	case string:
		return parseLooseInt(n)
	default:
		return 0
	}
}

func parseLooseInt(value string) int {
	value = strings.ReplaceAll(value, ".", "")
	value = strings.ReplaceAll(value, ",", ".")
	value = strings.TrimSpace(value)
	var n float64
	_, _ = fmt.Sscanf(value, "%f", &n)
	return int(math.Round(n))
}

func boundedInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
