package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Settings struct {
	Addr                string
	DatabaseURL         string
	AdminToken          string
	CORSOrigins         []string
	OutreachDryRun      bool
	OutreachFromName    string
	OutreachCompany     string
	OutreachBaseURL     string
	MaxImportRows       int
	RequestTimeout      time.Duration
	SendingDailyLimit   int
	RequireLeadConsent  bool
	AllowProspectOutbox bool
}

func Load() Settings {
	return Settings{
		Addr:                env("MEDMASIA_ADDR", "127.0.0.1:8092"),
		DatabaseURL:         env("MEDMASIA_DATABASE_URL", ""),
		AdminToken:          env("MEDMASIA_ADMIN_TOKEN", "dev-admin-token"),
		CORSOrigins:         splitCSV(env("MEDMASIA_CORS_ORIGINS", "http://localhost:8000,http://127.0.0.1:8000,https://karuata.github.io")),
		OutreachDryRun:      boolEnv("MEDMASIA_OUTREACH_DRY_RUN", true),
		OutreachFromName:    env("MEDMASIA_OUTREACH_FROM_NAME", "Rodrigo Masini"),
		OutreachCompany:     env("MEDMASIA_OUTREACH_COMPANY", "MedMasIA"),
		OutreachBaseURL:     strings.TrimRight(env("MEDMASIA_OUTREACH_BASE_URL", "https://karuata.github.io/medmasia/"), "/"),
		MaxImportRows:       intEnv("MEDMASIA_MAX_IMPORT_ROWS", 0),
		RequestTimeout:      durationSeconds("MEDMASIA_TIMEOUT_SECONDS", 10),
		SendingDailyLimit:   intEnv("MEDMASIA_SENDING_DAILY_LIMIT", 150),
		RequireLeadConsent:  boolEnv("MEDMASIA_REQUIRE_LEAD_CONSENT", true),
		AllowProspectOutbox: boolEnv("MEDMASIA_ALLOW_PROSPECT_OUTBOX", true),
	}
}

func (s Settings) Validate() error {
	if strings.TrimSpace(s.DatabaseURL) == "" {
		return ErrMissingDatabaseURL
	}
	return nil
}

var ErrMissingDatabaseURL = settingsError("MEDMASIA_DATABASE_URL is required")

type settingsError string

func (e settingsError) Error() string { return string(e) }

func env(name, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	return fallback
}

func boolEnv(name string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	if v == "" {
		return fallback
	}
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func intEnv(name string, fallback int) int {
	v, err := strconv.Atoi(strings.TrimSpace(os.Getenv(name)))
	if err != nil {
		return fallback
	}
	return v
}

func durationSeconds(name string, fallback int) time.Duration {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return time.Duration(fallback) * time.Second
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return time.Duration(fallback) * time.Second
	}
	return time.Duration(v * float64(time.Second))
}

func splitCSV(value string) []string {
	var out []string
	for _, part := range strings.Split(value, ",") {
		if item := strings.TrimSpace(part); item != "" {
			out = append(out, item)
		}
	}
	return out
}
