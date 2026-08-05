package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"medmasia/backend/internal/config"
	"medmasia/backend/internal/core"
	"medmasia/backend/internal/service"
	"medmasia/backend/internal/store"
)

type Server struct {
	settings config.Settings
	store    *store.Store
	in       io.Reader
	out      io.Writer
}

func New(settings config.Settings, st *store.Store, in io.Reader, out io.Writer) *Server {
	return &Server{settings: settings, store: st, in: in, out: out}
}

func (s *Server) Run(ctx context.Context) error {
	scanner := bufio.NewScanner(s.in)
	encoder := json.NewEncoder(s.out)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		var req request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			_ = encoder.Encode(response{JSONRPC: "2.0", Error: &rpcError{Code: -32700, Message: "parse error"}})
			continue
		}
		if req.Method == "notifications/initialized" {
			continue
		}
		result, rpcErr := s.handle(ctx, req)
		resp := response{JSONRPC: "2.0", ID: req.ID}
		if rpcErr != nil {
			resp.Error = rpcErr
		} else {
			resp.Result = result
		}
		if err := encoder.Encode(resp); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (s *Server) handle(ctx context.Context, req request) (any, *rpcError) {
	switch req.Method {
	case "initialize":
		return map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{"name": "medmasia-sales", "version": "0.1.0"},
		}, nil
	case "tools/list":
		return map[string]any{"tools": tools()}, nil
	case "tools/call":
		var params toolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, &rpcError{Code: -32602, Message: "invalid params"}
		}
		return s.callTool(ctx, params)
	default:
		return nil, &rpcError{Code: -32601, Message: "method not found"}
	}
}

func (s *Server) callTool(ctx context.Context, params toolCallParams) (any, *rpcError) {
	switch params.Name {
	case "medmasia.prospect_search":
		query := stringArg(params.Arguments, "query")
		grade := stringArg(params.Arguments, "grade")
		limit := intArg(params.Arguments, "limit", 20)
		items, err := s.store.SearchProspects(ctx, query, grade, limit)
		if err != nil {
			return nil, &rpcError{Code: -32000, Message: err.Error()}
		}
		return textResult(map[string]any{"prospects": items}), nil
	case "medmasia.prospect_score":
		prospectID := stringArg(params.Arguments, "prospect_id")
		p, err := s.store.GetProspect(ctx, prospectID)
		if err != nil {
			return nil, &rpcError{Code: -32004, Message: "prospect not found"}
		}
		score := service.ScoreProspect(p)
		if boolArg(params.Arguments, "persist", true) {
			if err := s.store.UpdateProspectScore(ctx, p.ID, score.Score, score.Grade, score.Reasons); err != nil {
				return nil, &rpcError{Code: -32000, Message: err.Error()}
			}
		}
		return textResult(score), nil
	case "medmasia.draft_outreach":
		prospectID := stringArg(params.Arguments, "prospect_id")
		p, err := s.store.GetProspect(ctx, prospectID)
		if err != nil {
			return nil, &rpcError{Code: -32004, Message: "prospect not found"}
		}
		subject, body, err := service.DraftEmail(p, service.OutreachSettings{
			FromName: s.settings.OutreachFromName,
			Company:  s.settings.OutreachCompany,
			BaseURL:  s.settings.OutreachBaseURL,
			DryRun:   s.settings.OutreachDryRun,
		})
		if err != nil {
			return nil, &rpcError{Code: -32010, Message: err.Error()}
		}
		if err := service.ComplianceReview("email", subject, body, false); err != nil {
			return nil, &rpcError{Code: -32011, Message: err.Error()}
		}
		payload := map[string]any{"prospect_id": p.ID, "channel": "email", "subject": subject, "body": body, "dry_run": true}
		if boolArg(params.Arguments, "queue", false) {
			t, err := s.store.QueueTouchpoint(ctx, core.Touchpoint{
				ProspectID: p.ID,
				Channel:    "email",
				Subject:    subject,
				Body:       body,
				Status:     "draft",
				Metadata:   map[string]any{"via": "mcp", "dry_run": true},
			})
			if err != nil {
				return nil, &rpcError{Code: -32000, Message: err.Error()}
			}
			payload["touchpoint_id"] = t.ID
		}
		return textResult(payload), nil
	case "medmasia.campaign_summary":
		summary, err := s.store.CampaignSummary(ctx)
		if err != nil {
			return nil, &rpcError{Code: -32000, Message: err.Error()}
		}
		return textResult(summary), nil
	default:
		return nil, &rpcError{Code: -32601, Message: "unknown tool"}
	}
}

func tools() []map[string]any {
	return []map[string]any{
		{
			"name":        "medmasia.prospect_search",
			"description": "Search imported physician prospects by name, CRM, city, and grade. Returns professional/commercial fields only.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{"type": "string"},
					"grade": map[string]any{"type": "string", "enum": []string{"", "A", "B", "C", "D"}},
					"limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 200},
				},
			},
		},
		{
			"name":        "medmasia.prospect_score",
			"description": "Score one prospect with the MedMasIA commercial classifier and optionally persist the result.",
			"inputSchema": map[string]any{
				"type":     "object",
				"required": []string{"prospect_id"},
				"properties": map[string]any{
					"prospect_id": map[string]any{"type": "string"},
					"persist":     map[string]any{"type": "boolean"},
				},
			},
		},
		{
			"name":        "medmasia.draft_outreach",
			"description": "Draft a compliant first email for one prospect. Queues as draft only when requested.",
			"inputSchema": map[string]any{
				"type":     "object",
				"required": []string{"prospect_id"},
				"properties": map[string]any{
					"prospect_id": map[string]any{"type": "string"},
					"queue":       map[string]any{"type": "boolean"},
				},
			},
		},
		{
			"name":        "medmasia.campaign_summary",
			"description": "Return counts for prospects, leads, and outbox statuses.",
			"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
		},
	}
}

func textResult(payload any) map[string]any {
	buf, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		buf = []byte(fmt.Sprintf("%v", payload))
	}
	return map[string]any{
		"content": []map[string]any{{"type": "text", "text": string(buf)}},
	}
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type toolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

func stringArg(values map[string]any, key string) string {
	if value, ok := values[key].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func boolArg(values map[string]any, key string, fallback bool) bool {
	if value, ok := values[key].(bool); ok {
		return value
	}
	return fallback
}

func intArg(values map[string]any, key string, fallback int) int {
	if value, ok := values[key].(float64); ok {
		return int(value)
	}
	return fallback
}
