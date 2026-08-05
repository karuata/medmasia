package core

type Prospect struct {
	ID                   string         `json:"id"`
	Source               string         `json:"source"`
	SourceRef            string         `json:"source_ref"`
	Name                 string         `json:"name"`
	CRM                  string         `json:"crm"`
	CRMUF                string         `json:"crm_uf"`
	Specialties          []string       `json:"specialties"`
	City                 string         `json:"city"`
	UF                   string         `json:"uf"`
	Emails               []string       `json:"emails"`
	Phones               []string       `json:"phones"`
	LinkedInURL          string         `json:"linkedin_url"`
	ProfessionalMetadata map[string]any `json:"professional_metadata"`
	Status               string         `json:"status"`
	Score                int            `json:"score"`
	Grade                string         `json:"grade"`
	ScoreReasons         []string       `json:"score_reasons"`
	CreatedAt            string         `json:"created_at"`
	UpdatedAt            string         `json:"updated_at"`
}

type Lead struct {
	ID               string `json:"id"`
	SessionID        string `json:"session_id"`
	Site             string `json:"site"`
	Name             string `json:"name"`
	Company          string `json:"company"`
	Role             string `json:"role"`
	Email            string `json:"email"`
	Phone            string `json:"phone"`
	PreferredContact string `json:"preferred_contact"`
	Language         string `json:"language"`
	MainPain         string `json:"main_pain"`
	ServiceInterest  string `json:"service_interest"`
	Urgency          string `json:"urgency"`
	Consent          bool   `json:"consent"`
	Status           string `json:"status"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type Campaign struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	AudienceFilter map[string]any `json:"audience_filter"`
	Channel        string         `json:"channel"`
	Status         string         `json:"status"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
}

type Touchpoint struct {
	ID                string         `json:"id"`
	ProspectID        string         `json:"prospect_id"`
	LeadID            string         `json:"lead_id"`
	CampaignID        string         `json:"campaign_id"`
	Channel           string         `json:"channel"`
	Subject           string         `json:"subject"`
	Body              string         `json:"body"`
	Status            string         `json:"status"`
	Provider          string         `json:"provider"`
	ProviderMessageID string         `json:"provider_message_id"`
	Error             string         `json:"error"`
	ApprovedBy        string         `json:"approved_by"`
	ScheduledAt       string         `json:"scheduled_at"`
	SentAt            string         `json:"sent_at"`
	Metadata          map[string]any `json:"metadata"`
	CreatedAt         string         `json:"created_at"`
	UpdatedAt         string         `json:"updated_at"`
}

type AgentRun struct {
	ID          string      `json:"id"`
	Status      string      `json:"status"`
	RequestedBy string      `json:"requested_by"`
	Mode        string      `json:"mode"`
	Summary     string      `json:"summary"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
	Steps       []AgentStep `json:"steps,omitempty"`
}

type AgentStep struct {
	ID          string         `json:"id"`
	AgentName   string         `json:"agent_name"`
	Status      string         `json:"status"`
	Summary     string         `json:"summary"`
	Artifact    map[string]any `json:"artifact"`
	Error       string         `json:"error"`
	StartedAt   string         `json:"started_at"`
	CompletedAt string         `json:"completed_at"`
}

type ChatResult struct {
	Answer       string   `json:"answer"`
	CollectLead  bool     `json:"collect_lead"`
	QuickActions []string `json:"quick_actions"`
	Confidence   string   `json:"confidence"`
}
