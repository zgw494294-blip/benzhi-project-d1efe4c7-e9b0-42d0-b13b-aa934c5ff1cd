package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"tactile-review/internal/application"
	"tactile-review/internal/compliance"
	"tactile-review/internal/domain"
)

type Server struct {
	App *application.Service
	mux *http.ServeMux
}

func New(a *application.Service) *Server {
	s := &Server{App: a, mux: http.NewServeMux()}
	s.routes()
	return s
}
func (s *Server) routes() {
	s.mux.HandleFunc("/", s.index)
	s.mux.HandleFunc("/static/app.css", s.css)
	s.mux.HandleFunc("/static/app.js", s.js)
	s.mux.HandleFunc("/api/cases", s.cases)
	s.mux.HandleFunc("/api/cases/", s.caseAction)
	s.mux.HandleFunc("/api/standards", s.standards)
}
func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; script-src 'self'")
		s.mux.ServeHTTP(w, r)
	})
}
func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	template.Must(template.New("index").Parse(indexHTML)).Execute(w, nil)
}
func (s *Server) css(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	io.WriteString(w, cssText)
}
func (s *Server) js(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	io.WriteString(w, jsText)
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := decodeStrict(w, r, v); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "JSON 格式无效或包含未知字段："+err.Error())
		return false
	}
	return true
}
func (s *Server) standards(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		allow(w, "GET")
		return
	}
	writeJSON(w, compliance.VersionCatalog())
}

type createInput struct {
	BuildingZone         string `json:"buildingZone"`
	InstallationLocation string `json:"installationLocation"`
	AudienceProfile      string `json:"audienceProfile"`
	StandardVersion      string `json:"standardVersion"`
	DesignerID           string `json:"designerID"`
	MeasurerID           string `json:"measurerID"`
	IdempotencyKey       string `json:"idempotencyKey"`
}

func (in createInput) command() application.CreateCaseCommand {
	return application.CreateCaseCommand{BuildingZone: in.BuildingZone, InstallationLocation: in.InstallationLocation, AudienceProfile: in.AudienceProfile, StandardVersion: in.StandardVersion, DesignerID: in.DesignerID, MeasurerID: in.MeasurerID, IdempotencyKey: in.IdempotencyKey}
}
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
func (s *Server) cases(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		v, e := s.App.Store.List()
		if e != nil {
			http.Error(w, e.Error(), 500)
			return
		}
		writeJSON(w, v)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "method", 405)
		return
	}
	var in createInput
	if !decode(w, r, &in) {
		return
	}
	c, reused, e := s.App.CreateWithCommand(in.command())
	if e != nil {
		writeAppError(w, e)
		return
	}
	if reused {
		w.Header().Set("X-Idempotent-Replay", "true")
	}
	writeJSON(w, c)
}
func (s *Server) caseAction(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/cases/"), "/")
	if len(parts) == 0 {
		return
	}
	id := parts[0]
	if id == "preflight" && len(parts) == 1 && r.Method == "POST" {
		var in createInput
		if !decode(w, r, &in) {
			return
		}
		writeJSON(w, s.App.PreflightCreate(in.command()))
		return
	}
	if len(parts) == 1 && r.Method == "GET" {
		c, e := s.App.Store.Get(id)
		if e != nil {
			http.Error(w, e.Error(), 404)
			return
		}
		writeJSON(w, c)
		return
	}
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}
	action := parts[1]
	if r.Method == "GET" {
		s.caseQuery(w, r, id, action)
		return
	}
	if r.Method != "POST" {
		allow(w, "GET, POST")
		return
	}
	var in struct {
		ExpectedVersion int                           `json:"expectedVersion"`
		IdempotencyKey  string                        `json:"idempotencyKey"`
		ReviewerID      string                        `json:"reviewerID"`
		Decision        string                        `json:"decision"`
		Reason          string                        `json:"reason"`
		Issuer          string                        `json:"issuer"`
		Code            string                        `json:"code"`
		PreviewDigest   string                        `json:"previewDigest"`
		Revision        domain.PlateRevision          `json:"revision"`
		Items           []domain.ReviewReturnItem     `json:"items"`
		ConfirmItemIDs  []string                      `json:"confirmItemIDs"`
		Credential      *domain.FabricationCredential `json:"credential"`
		CredentialJSON  json.RawMessage               `json:"credentialJSON"`
	}
	if !decode(w, r, &in) {
		return
	}
	var out any
	var e error
	switch action {
	case "revisions":
		if len(parts) > 2 && parts[2] == "preview" {
			out, e = s.App.PreviewRevision(id, in.ExpectedVersion, in.Revision)
		} else {
			out, e = s.App.AddRevision(id, in.ExpectedVersion, in.Revision, in.IdempotencyKey)
		}
	case "rework-preview":
		out, e = s.App.PreviewRevision(id, in.ExpectedVersion, in.Revision)
	case "check":
		out, e = s.App.Check(id, in.ExpectedVersion)
	case "review":
		out, e = s.App.ReviewStructured(id, application.ReviewCommand{ExpectedVersion: in.ExpectedVersion, ReviewerID: in.ReviewerID, Decision: in.Decision, Reason: in.Reason, Items: in.Items, ConfirmItemIDs: in.ConfirmItemIDs})
	case "freeze":
		out, e = s.App.FreezeConfirm(id, in.ExpectedVersion, in.PreviewDigest)
	case "issue":
		out, e = s.App.Issue(id, in.ExpectedVersion, in.Issuer)
	case "verify":
		if in.Credential != nil {
			out, e = s.App.VerifyCredential(id, *in.Credential)
		} else if len(in.CredentialJSON) > 0 {
			var credential domain.FabricationCredential
			raw := in.CredentialJSON
			if len(raw) > 0 && raw[0] == '"' {
				var encoded string
				if json.Unmarshal(raw, &encoded) == nil {
					raw = []byte(encoded)
				}
			}
			if err := json.Unmarshal(raw, &credential); err != nil {
				e = fmt.Errorf("凭据 JSON 格式无效")
			} else {
				out, e = s.App.VerifyCredential(id, credential)
			}
		} else {
			var ok bool
			var status string
			ok, status, e = s.App.Verify(id, in.Code)
			out = map[string]any{"valid": ok, "status": status}
		}
	default:
		http.NotFound(w, r)
		return
	}
	if e != nil {
		writeAppError(w, e)
		return
	}
	writeJSON(w, out)
}

func (s *Server) caseQuery(w http.ResponseWriter, r *http.Request, id, action string) {
	switch action {
	case "checks":
		c, e := s.App.Store.Get(id)
		if e != nil {
			writeAppError(w, e)
			return
		}
		writeJSON(w, c.CheckRuns)
	case "compare":
		revisionID := r.URL.Query().Get("revision_id")
		v, e := s.App.CompareAdjacent(id, revisionID)
		if e != nil {
			writeAppError(w, e)
			return
		}
		writeJSON(w, v)
	case "freeze-preview":
		v, e := s.App.PreviewFreeze(id)
		if e != nil {
			writeAppError(w, e)
			return
		}
		writeJSON(w, v)
	case "verifications":
		v, e := s.App.Store.VerificationHistory(r.URL.Query().Get("credential_id"), 10)
		if e != nil {
			writeAppError(w, e)
			return
		}
		writeJSON(w, v)
	default:
		http.NotFound(w, r)
	}
}

func writeAppError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	code := "BUSINESS_RULE_VIOLATION"
	if errors.Is(err, domain.ErrNotFound) {
		status = http.StatusNotFound
		code = "NOT_FOUND"
	}
	if errors.Is(err, domain.ErrVersionConflict) || errors.Is(err, domain.ErrIdempotencyConflict) || errors.Is(err, domain.ErrDigestConflict) {
		status = http.StatusConflict
		code = "CONFLICT"
	}
	var validation *domain.ValidationErrors
	if errors.As(err, &validation) {
		status = http.StatusUnprocessableEntity
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]any{"code": "VALIDATION_FAILED", "message": err.Error(), "issues": validation.Issues, "availableVersions": compliance.VersionCatalog()})
		return
	}
	respondError(w, status, code, err.Error())
}
