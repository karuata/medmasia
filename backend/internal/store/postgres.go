package store

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"medmasia/backend/internal/core"
	"medmasia/backend/internal/id"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func Open(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	s := &Store{pool: pool}
	if err := s.Init(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() { s.pool.Close() }

func (s *Store) Init(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, schema)
	return err
}

func (s *Store) UpsertProspect(ctx context.Context, p core.Prospect) (core.Prospect, error) {
	if strings.TrimSpace(p.ID) == "" {
		p.ID = id.New()
	}
	if strings.TrimSpace(p.Source) == "" {
		p.Source = "manual"
	}
	if strings.TrimSpace(p.Status) == "" {
		p.Status = "new"
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO prospects (
			id, source, source_ref, name, crm, crm_uf, specialties_json, city, uf,
			emails_json, phones_json, linkedin_url, professional_metadata, status,
			score, grade, score_reasons_json, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,NOW(),NOW())
		ON CONFLICT (source, source_ref) DO UPDATE SET
			name=EXCLUDED.name,
			crm=EXCLUDED.crm,
			crm_uf=EXCLUDED.crm_uf,
			specialties_json=EXCLUDED.specialties_json,
			city=EXCLUDED.city,
			uf=EXCLUDED.uf,
			emails_json=EXCLUDED.emails_json,
			phones_json=EXCLUDED.phones_json,
			linkedin_url=EXCLUDED.linkedin_url,
			professional_metadata=prospects.professional_metadata || EXCLUDED.professional_metadata,
			status=EXCLUDED.status,
			score=EXCLUDED.score,
			grade=EXCLUDED.grade,
			score_reasons_json=EXCLUDED.score_reasons_json,
			updated_at=NOW()
		RETURNING id, source, source_ref, name, crm, crm_uf, specialties_json, city, uf,
			emails_json, phones_json, linkedin_url, professional_metadata, status,
			score, grade, score_reasons_json, created_at, updated_at
	`, p.ID, p.Source, p.SourceRef, p.Name, p.CRM, p.CRMUF, jsonString(p.Specialties), p.City, p.UF,
		jsonString(p.Emails), jsonString(p.Phones), p.LinkedInURL, jsonString(p.ProfessionalMetadata),
		p.Status, p.Score, p.Grade, jsonString(p.ScoreReasons))
	return scanProspect(row)
}

func (s *Store) UpdateProspectScore(ctx context.Context, prospectID string, score int, grade string, reasons []string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE prospects
		SET score=$2, grade=$3, score_reasons_json=$4, updated_at=NOW()
		WHERE id=$1
	`, prospectID, score, grade, jsonString(reasons))
	return err
}

func (s *Store) GetProspect(ctx context.Context, prospectID string) (core.Prospect, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, source, source_ref, name, crm, crm_uf, specialties_json, city, uf,
			emails_json, phones_json, linkedin_url, professional_metadata, status,
			score, grade, score_reasons_json, created_at, updated_at
		FROM prospects WHERE id=$1
	`, prospectID)
	return scanProspect(row)
}

func (s *Store) SearchProspects(ctx context.Context, query, grade string, limit int) ([]core.Prospect, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	query = strings.TrimSpace(query)
	grade = strings.ToUpper(strings.TrimSpace(grade))
	rows, err := s.pool.Query(ctx, `
		SELECT id, source, source_ref, name, crm, crm_uf, specialties_json, city, uf,
			emails_json, phones_json, linkedin_url, professional_metadata, status,
			score, grade, score_reasons_json, created_at, updated_at
		FROM prospects
		WHERE ($1='' OR name ILIKE '%' || $1 || '%' OR crm ILIKE '%' || $1 || '%' OR city ILIKE '%' || $1 || '%')
		  AND ($2='' OR grade=$2)
		ORDER BY score DESC, updated_at DESC
		LIMIT $3
	`, query, grade, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []core.Prospect
	for rows.Next() {
		p, err := scanProspect(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) UpsertLead(ctx context.Context, l core.Lead) (core.Lead, error) {
	if strings.TrimSpace(l.ID) == "" {
		l.ID = id.New()
	}
	if strings.TrimSpace(l.SessionID) == "" {
		l.SessionID = id.New()
	}
	if strings.TrimSpace(l.Language) == "" {
		l.Language = "pt"
	}
	if strings.TrimSpace(l.Status) == "" {
		l.Status = "new"
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO leads (
			id, session_id, site, name, company, role, email, phone, preferred_contact,
			language, main_pain, service_interest, urgency, consent, status, created_at, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,NOW(),NOW())
		ON CONFLICT (session_id) DO UPDATE SET
			site=EXCLUDED.site,
			name=COALESCE(NULLIF(EXCLUDED.name,''), leads.name),
			company=COALESCE(NULLIF(EXCLUDED.company,''), leads.company),
			role=COALESCE(NULLIF(EXCLUDED.role,''), leads.role),
			email=COALESCE(NULLIF(EXCLUDED.email,''), leads.email),
			phone=COALESCE(NULLIF(EXCLUDED.phone,''), leads.phone),
			preferred_contact=COALESCE(NULLIF(EXCLUDED.preferred_contact,''), leads.preferred_contact),
			language=COALESCE(NULLIF(EXCLUDED.language,''), leads.language),
			main_pain=COALESCE(NULLIF(EXCLUDED.main_pain,''), leads.main_pain),
			service_interest=COALESCE(NULLIF(EXCLUDED.service_interest,''), leads.service_interest),
			urgency=COALESCE(NULLIF(EXCLUDED.urgency,''), leads.urgency),
			consent=leads.consent OR EXCLUDED.consent,
			status=EXCLUDED.status,
			updated_at=NOW()
		RETURNING id, session_id, site, name, company, role, email, phone, preferred_contact,
			language, main_pain, service_interest, urgency, consent, status, created_at, updated_at
	`, l.ID, l.SessionID, l.Site, l.Name, l.Company, l.Role, l.Email, l.Phone, l.PreferredContact,
		l.Language, l.MainPain, l.ServiceInterest, l.Urgency, l.Consent, l.Status)
	return scanLead(row)
}

func (s *Store) CreateCampaign(ctx context.Context, name, channel string, filter map[string]any) (core.Campaign, error) {
	c := core.Campaign{ID: id.New(), Name: name, Channel: channel, AudienceFilter: filter, Status: "draft"}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO campaigns (id, name, audience_filter, channel, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,NOW(),NOW())
		RETURNING id, name, audience_filter, channel, status, created_at, updated_at
	`, c.ID, c.Name, jsonString(c.AudienceFilter), c.Channel, c.Status)
	return scanCampaign(row)
}

func (s *Store) QueueTouchpoint(ctx context.Context, t core.Touchpoint) (core.Touchpoint, error) {
	if strings.TrimSpace(t.ID) == "" {
		t.ID = id.New()
	}
	if strings.TrimSpace(t.Status) == "" {
		t.Status = "draft"
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO touchpoints (
			id, prospect_id, lead_id, campaign_id, channel, subject, body, status,
			provider, provider_message_id, error, approved_by, scheduled_at, sent_at,
			metadata, created_at, updated_at
		)
		VALUES ($1,NULLIF($2,'')::uuid,NULLIF($3,'')::uuid,NULLIF($4,'')::uuid,$5,$6,$7,$8,$9,$10,$11,$12,
			NULLIF($13,'')::timestamptz,NULLIF($14,'')::timestamptz,$15,NOW(),NOW())
		RETURNING id, COALESCE(prospect_id::text,''), COALESCE(lead_id::text,''), COALESCE(campaign_id::text,''),
			channel, subject, body, status, provider, provider_message_id, error, approved_by,
			COALESCE(scheduled_at::text,''), COALESCE(sent_at::text,''), metadata, created_at, updated_at
	`, t.ID, t.ProspectID, t.LeadID, t.CampaignID, t.Channel, t.Subject, t.Body, t.Status,
		t.Provider, t.ProviderMessageID, t.Error, t.ApprovedBy, t.ScheduledAt, t.SentAt, jsonString(t.Metadata))
	return scanTouchpoint(row)
}

func (s *Store) CountSuppressed(ctx context.Context, email, phone string) (int, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM suppression_list
		WHERE ($1<>'' AND email=$1) OR ($2<>'' AND phone=$2)
	`, strings.ToLower(strings.TrimSpace(email)), onlyDigits(phone))
	var n int
	return n, row.Scan(&n)
}

func (s *Store) CampaignSummary(ctx context.Context) (map[string]any, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT 'prospects' AS name, COUNT(*)::int FROM prospects
		UNION ALL SELECT 'grade_a', COUNT(*)::int FROM prospects WHERE grade='A'
		UNION ALL SELECT 'grade_b', COUNT(*)::int FROM prospects WHERE grade='B'
		UNION ALL SELECT 'leads', COUNT(*)::int FROM leads
		UNION ALL SELECT 'touchpoints_draft', COUNT(*)::int FROM touchpoints WHERE status='draft'
		UNION ALL SELECT 'touchpoints_approved', COUNT(*)::int FROM touchpoints WHERE status='approved'
		UNION ALL SELECT 'touchpoints_sent', COUNT(*)::int FROM touchpoints WHERE status='sent'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]any{}
	for rows.Next() {
		var name string
		var count int
		if err := rows.Scan(&name, &count); err != nil {
			return nil, err
		}
		out[name] = count
	}
	return out, rows.Err()
}

func (s *Store) CreateAgentRun(ctx context.Context, requestedBy, mode string) (string, error) {
	runID := id.New()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO agent_runs (id, status, requested_by, mode, summary, created_at, updated_at)
		VALUES ($1,'running',$2,$3,'',NOW(),NOW())
	`, runID, requestedBy, mode)
	return runID, err
}

func (s *Store) AddAgentStep(ctx context.Context, runID, agentName, status, summary string, artifact map[string]any, stepErr string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO agent_steps (id, run_id, agent_name, status, summary, artifact, error, started_at, completed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,NOW(),NOW())
	`, id.New(), runID, agentName, status, summary, jsonString(artifact), stepErr)
	if err == nil {
		_, _ = s.pool.Exec(ctx, "UPDATE agent_runs SET updated_at=NOW() WHERE id=$1", runID)
	}
	return err
}

func (s *Store) FinishAgentRun(ctx context.Context, runID, status, summary string) (core.AgentRun, error) {
	_, err := s.pool.Exec(ctx, "UPDATE agent_runs SET status=$2, summary=$3, updated_at=NOW() WHERE id=$1", runID, status, summary)
	if err != nil {
		return core.AgentRun{}, err
	}
	return s.GetAgentRun(ctx, runID)
}

func (s *Store) GetAgentRun(ctx context.Context, runID string) (core.AgentRun, error) {
	row := s.pool.QueryRow(ctx, "SELECT id, status, requested_by, mode, summary, created_at, updated_at FROM agent_runs WHERE id=$1", runID)
	var r core.AgentRun
	var created, updated time.Time
	if err := row.Scan(&r.ID, &r.Status, &r.RequestedBy, &r.Mode, &r.Summary, &created, &updated); err != nil {
		return core.AgentRun{}, err
	}
	r.CreatedAt = formatTime(created)
	r.UpdatedAt = formatTime(updated)
	rows, err := s.pool.Query(ctx, `
		SELECT id, agent_name, status, summary, artifact, error, started_at, completed_at
		FROM agent_steps WHERE run_id=$1 ORDER BY started_at ASC
	`, runID)
	if err != nil {
		return core.AgentRun{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var step core.AgentStep
		var artifact []byte
		var started, completed time.Time
		if err := rows.Scan(&step.ID, &step.AgentName, &step.Status, &step.Summary, &artifact, &step.Error, &started, &completed); err != nil {
			return core.AgentRun{}, err
		}
		step.Artifact = mapStringAny(artifact)
		step.StartedAt = formatTime(started)
		step.CompletedAt = formatTime(completed)
		r.Steps = append(r.Steps, step)
	}
	return r, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanProspect(row rowScanner) (core.Prospect, error) {
	var p core.Prospect
	var specialties, emails, phones, metadata, reasons []byte
	var created, updated time.Time
	err := row.Scan(&p.ID, &p.Source, &p.SourceRef, &p.Name, &p.CRM, &p.CRMUF, &specialties, &p.City, &p.UF,
		&emails, &phones, &p.LinkedInURL, &metadata, &p.Status, &p.Score, &p.Grade, &reasons, &created, &updated)
	if err != nil {
		return core.Prospect{}, err
	}
	p.Specialties = stringSlice(specialties)
	p.Emails = stringSlice(emails)
	p.Phones = stringSlice(phones)
	p.ProfessionalMetadata = mapStringAny(metadata)
	p.ScoreReasons = stringSlice(reasons)
	p.CreatedAt = formatTime(created)
	p.UpdatedAt = formatTime(updated)
	return p, nil
}

func scanLead(row rowScanner) (core.Lead, error) {
	var l core.Lead
	var created, updated time.Time
	err := row.Scan(&l.ID, &l.SessionID, &l.Site, &l.Name, &l.Company, &l.Role, &l.Email, &l.Phone, &l.PreferredContact,
		&l.Language, &l.MainPain, &l.ServiceInterest, &l.Urgency, &l.Consent, &l.Status, &created, &updated)
	l.CreatedAt = formatTime(created)
	l.UpdatedAt = formatTime(updated)
	return l, err
}

func scanCampaign(row rowScanner) (core.Campaign, error) {
	var c core.Campaign
	var filter []byte
	var created, updated time.Time
	err := row.Scan(&c.ID, &c.Name, &filter, &c.Channel, &c.Status, &created, &updated)
	c.AudienceFilter = mapStringAny(filter)
	c.CreatedAt = formatTime(created)
	c.UpdatedAt = formatTime(updated)
	return c, err
}

func scanTouchpoint(row rowScanner) (core.Touchpoint, error) {
	var t core.Touchpoint
	var metadata []byte
	var created, updated time.Time
	err := row.Scan(&t.ID, &t.ProspectID, &t.LeadID, &t.CampaignID, &t.Channel, &t.Subject, &t.Body, &t.Status,
		&t.Provider, &t.ProviderMessageID, &t.Error, &t.ApprovedBy, &t.ScheduledAt, &t.SentAt, &metadata, &created, &updated)
	t.Metadata = mapStringAny(metadata)
	t.CreatedAt = formatTime(created)
	t.UpdatedAt = formatTime(updated)
	return t, err
}

func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func jsonString(value any) string {
	if value == nil {
		return "{}"
	}
	buf, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(buf)
}

func stringSlice(raw []byte) []string {
	var out []string
	_ = json.Unmarshal(raw, &out)
	return out
}

func mapStringAny(raw []byte) map[string]any {
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	return out
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}

func onlyDigits(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

const schema = `
CREATE TABLE IF NOT EXISTS prospects (
	id uuid PRIMARY KEY,
	source text NOT NULL DEFAULT 'manual',
	source_ref text NOT NULL DEFAULT '',
	name text NOT NULL DEFAULT '',
	crm text NOT NULL DEFAULT '',
	crm_uf text NOT NULL DEFAULT '',
	specialties_json jsonb NOT NULL DEFAULT '[]'::jsonb,
	city text NOT NULL DEFAULT '',
	uf text NOT NULL DEFAULT '',
	emails_json jsonb NOT NULL DEFAULT '[]'::jsonb,
	phones_json jsonb NOT NULL DEFAULT '[]'::jsonb,
	linkedin_url text NOT NULL DEFAULT '',
	professional_metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
	status text NOT NULL DEFAULT 'new',
	score integer NOT NULL DEFAULT 0 CHECK (score >= 0 AND score <= 100),
	grade text NOT NULL DEFAULT 'D',
	score_reasons_json jsonb NOT NULL DEFAULT '[]'::jsonb,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	UNIQUE (source, source_ref)
);

CREATE INDEX IF NOT EXISTS prospects_score_idx ON prospects(score DESC);
CREATE INDEX IF NOT EXISTS prospects_grade_idx ON prospects(grade);
CREATE INDEX IF NOT EXISTS prospects_name_idx ON prospects USING gin (to_tsvector('simple', name));

CREATE TABLE IF NOT EXISTS leads (
	id uuid PRIMARY KEY,
	session_id text NOT NULL UNIQUE,
	site text NOT NULL DEFAULT '',
	name text NOT NULL DEFAULT '',
	company text NOT NULL DEFAULT '',
	role text NOT NULL DEFAULT '',
	email text NOT NULL DEFAULT '',
	phone text NOT NULL DEFAULT '',
	preferred_contact text NOT NULL DEFAULT '',
	language text NOT NULL DEFAULT 'pt',
	main_pain text NOT NULL DEFAULT '',
	service_interest text NOT NULL DEFAULT '',
	urgency text NOT NULL DEFAULT '',
	consent boolean NOT NULL DEFAULT false,
	status text NOT NULL DEFAULT 'new',
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS campaigns (
	id uuid PRIMARY KEY,
	name text NOT NULL,
	audience_filter jsonb NOT NULL DEFAULT '{}'::jsonb,
	channel text NOT NULL CHECK (channel IN ('email','whatsapp','linkedin','phone')),
	status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','approved','active','paused','completed','cancelled')),
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS touchpoints (
	id uuid PRIMARY KEY,
	prospect_id uuid REFERENCES prospects(id) ON DELETE CASCADE,
	lead_id uuid REFERENCES leads(id) ON DELETE CASCADE,
	campaign_id uuid REFERENCES campaigns(id) ON DELETE SET NULL,
	channel text NOT NULL CHECK (channel IN ('email','whatsapp','linkedin','phone')),
	subject text NOT NULL DEFAULT '',
	body text NOT NULL DEFAULT '',
	status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','approved','queued','sent','failed','suppressed','cancelled')),
	provider text NOT NULL DEFAULT '',
	provider_message_id text NOT NULL DEFAULT '',
	error text NOT NULL DEFAULT '',
	approved_by text NOT NULL DEFAULT '',
	scheduled_at timestamptz,
	sent_at timestamptz,
	metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	CHECK (prospect_id IS NOT NULL OR lead_id IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS touchpoints_status_idx ON touchpoints(status, scheduled_at);

CREATE TABLE IF NOT EXISTS suppression_list (
	id uuid PRIMARY KEY,
	email text NOT NULL DEFAULT '',
	phone text NOT NULL DEFAULT '',
	reason text NOT NULL,
	source text NOT NULL DEFAULT '',
	created_at timestamptz NOT NULL,
	CHECK (email <> '' OR phone <> '')
);

CREATE UNIQUE INDEX IF NOT EXISTS suppression_email_idx ON suppression_list(email) WHERE email <> '';
CREATE UNIQUE INDEX IF NOT EXISTS suppression_phone_idx ON suppression_list(phone) WHERE phone <> '';

CREATE TABLE IF NOT EXISTS consent_events (
	id uuid PRIMARY KEY,
	prospect_id uuid REFERENCES prospects(id) ON DELETE CASCADE,
	lead_id uuid REFERENCES leads(id) ON DELETE CASCADE,
	channel text NOT NULL,
	event_type text NOT NULL,
	source text NOT NULL DEFAULT '',
	occurred_at timestamptz NOT NULL,
	CHECK (prospect_id IS NOT NULL OR lead_id IS NOT NULL)
);

CREATE TABLE IF NOT EXISTS agent_runs (
	id uuid PRIMARY KEY,
	status text NOT NULL DEFAULT 'running',
	requested_by text NOT NULL DEFAULT '',
	mode text NOT NULL DEFAULT 'preflight',
	summary text NOT NULL DEFAULT '',
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS agent_steps (
	id uuid PRIMARY KEY,
	run_id uuid NOT NULL REFERENCES agent_runs(id) ON DELETE CASCADE,
	agent_name text NOT NULL,
	status text NOT NULL,
	summary text NOT NULL DEFAULT '',
	artifact jsonb NOT NULL DEFAULT '{}'::jsonb,
	error text NOT NULL DEFAULT '',
	started_at timestamptz NOT NULL,
	completed_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS agent_steps_run_idx ON agent_steps(run_id, started_at ASC);
`
