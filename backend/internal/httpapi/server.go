package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"medmasia/backend/internal/config"
	"medmasia/backend/internal/core"
	"medmasia/backend/internal/id"
	"medmasia/backend/internal/service"
	"medmasia/backend/internal/store"
)

type Server struct {
	settings config.Settings
	store    *store.Store
	agents   *service.AgentRunner
	mux      *http.ServeMux
}

func New(settings config.Settings, st *store.Store, agents *service.AgentRunner) *Server {
	s := &Server{settings: settings, store: st, agents: agents, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, origin := range s.settings.CORSOrigins {
		if r.Header.Get("Origin") == origin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			break
		}
	}
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Admin-Token")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/health", s.health)
	s.mux.HandleFunc("/api/chat", s.chat)
	s.mux.HandleFunc("/api/lead", s.lead)
	s.mux.HandleFunc("/api/admin/summary", s.admin(s.summary))
	s.mux.HandleFunc("/api/admin/prospects", s.admin(s.prospects))
	s.mux.HandleFunc("/api/admin/prospects/", s.admin(s.prospectDetail))
	s.mux.HandleFunc("/api/admin/campaigns", s.admin(s.campaigns))
}

func (s *Server) admin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.settings.AdminToken == "" || r.Header.Get("X-Admin-Token") != s.settings.AdminToken {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r)
	}
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":               true,
		"service":          "medmasia-sales",
		"database":         "postgresql",
		"outreach_dry_run": s.settings.OutreachDryRun,
		"clinical_data":    "rejected",
	})
}

func (s *Server) chat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var body struct {
		Message   string `json:"message"`
		SessionID string `json:"session_id"`
		Language  string `json:"language"`
		Site      string `json:"site"`
	}
	if !decode(w, r, &body) {
		return
	}
	if err := service.RejectClinicalContent(map[string]string{"message": body.Message}); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"session_id":    body.SessionID,
			"answer":        "Nao envie dados clinicos ou dados de pacientes por aqui. Posso ajudar com duvidas comerciais sobre automacao, captacao e organizacao de contatos.",
			"collect_lead":  false,
			"quick_actions": []string{"Falar sobre automacao comercial", "Ver politica de privacidade"},
			"confidence":    "high",
		})
		return
	}
	answer := "Posso te ajudar a entender como a MedMasIA organiza contatos comerciais, qualifica oportunidades e registra interessados vindos do site. Se quiser, deixe nome, email ou telefone e o melhor canal de contato."
	if strings.TrimSpace(body.Message) != "" {
		answer = "Entendi. A MedMasIA pode registrar o interesse, classificar a oportunidade e encaminhar o proximo contato sem usar dados clinicos de pacientes. Quer deixar seus dados para retorno?"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"session_id":    defaultString(body.SessionID, "site-"+id.New()),
		"answer":        answer,
		"collect_lead":  true,
		"quick_actions": []string{"Quero uma demonstracao", "Tenho uma clinica", "Prefiro email"},
		"confidence":    "medium",
	})
}

func (s *Server) lead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var body map[string]any
	if !decode(w, r, &body) {
		return
	}
	lead := core.Lead{
		SessionID:        stringMap(body, "session_id"),
		Site:             stringMap(body, "site"),
		Name:             stringMap(body, "name"),
		Company:          stringMap(body, "company"),
		Role:             stringMap(body, "role"),
		Email:            service.NormalizeEmail(stringMap(body, "email")),
		Phone:            service.NormalizePhone(stringMap(body, "phone")),
		PreferredContact: stringMap(body, "preferred_contact"),
		Language:         defaultString(stringMap(body, "language"), "pt"),
		MainPain:         stringMap(body, "main_pain"),
		ServiceInterest:  stringMap(body, "service_interest"),
		Urgency:          stringMap(body, "urgency"),
		Consent:          boolMap(body, "consent"),
		Status:           "new",
	}
	if err := service.RejectClinicalContent(map[string]string{"name": lead.Name, "company": lead.Company, "main_pain": lead.MainPain, "service_interest": lead.ServiceInterest}); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if s.settings.RequireLeadConsent && !lead.Consent {
		writeError(w, http.StatusBadRequest, "consent_required")
		return
	}
	saved, err := s.store.UpsertLead(r.Context(), lead)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"lead": saved})
}

func (s *Server) summary(w http.ResponseWriter, r *http.Request) {
	result, err := s.store.CampaignSummary(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) prospects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	items, err := s.store.SearchProspects(r.Context(), r.URL.Query().Get("q"), r.URL.Query().Get("grade"), intQuery(r, "limit", 50))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"prospects": items})
}

func (s *Server) prospectDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/admin/prospects/"), "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	prospectID := parts[0]
	if r.Method == http.MethodPost && len(parts) > 1 && parts[1] == "run-outreach" {
		var body struct {
			RequestedBy string `json:"requested_by"`
		}
		if !decode(w, r, &body) {
			return
		}
		run, err := s.agents.RunProspectOutreach(r.Context(), prospectID, defaultString(body.RequestedBy, "operator"))
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"run": run})
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	p, err := s.store.GetProspect(r.Context(), prospectID)
	if err != nil {
		status := http.StatusInternalServerError
		if store.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, "prospect not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"prospect": p})
}

func (s *Server) campaigns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var body struct {
		Name           string         `json:"name"`
		Channel        string         `json:"channel"`
		AudienceFilter map[string]any `json:"audience_filter"`
	}
	if !decode(w, r, &body) {
		return
	}
	if body.Channel == "" {
		body.Channel = "email"
	}
	campaign, err := s.store.CreateCampaign(r.Context(), body.Name, body.Channel, body.AudienceFilter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"campaign": campaign})
}

func decode(w http.ResponseWriter, r *http.Request, out any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, detail any) {
	writeJSON(w, status, map[string]any{"detail": detail})
}

func stringMap(values map[string]any, key string) string {
	if value, ok := values[key].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func boolMap(values map[string]any, key string) bool {
	if value, ok := values[key].(bool); ok {
		return value
	}
	return false
}

func intQuery(r *http.Request, key string, fallback int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
