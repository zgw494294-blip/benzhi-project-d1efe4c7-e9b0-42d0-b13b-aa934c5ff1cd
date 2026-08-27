package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Status string

const (
	Draft          Status = "DRAFT"
	Checked        Status = "CHECKED"
	ReworkRequired Status = "REWORK_REQUIRED"
	ReadyForReview Status = "READY_FOR_REVIEW"
	Approved       Status = "APPROVED"
	Frozen         Status = "FROZEN"
	Authorized     Status = "AUTHORIZED"
)

type Severity string

const (
	Pass  Severity = "PASS"
	Warn  Severity = "WARN"
	Block Severity = "BLOCK"
)

type FindingStatus string

const (
	Open       FindingStatus = "OPEN"
	Superseded FindingStatus = "SUPERSEDED"
	Closed     FindingStatus = "CLOSED"
)

type ReleaseCase struct {
	ID                   string                 `json:"id"`
	BuildingZone         string                 `json:"building_zone"`
	InstallationLocation string                 `json:"installation_location"`
	AudienceProfile      string                 `json:"audience_profile"`
	StandardVersion      string                 `json:"standard_version"`
	Status               Status                 `json:"status"`
	Version              int                    `json:"version"`
	DesignerID           string                 `json:"designer_id"`
	MeasurerID           string                 `json:"measurer_id"`
	ReviewerID           string                 `json:"reviewer_id,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
	Revisions            []PlateRevision        `json:"revisions"`
	Findings             []ComplianceFinding    `json:"findings"`
	CheckRuns            []CheckRun             `json:"check_runs"`
	RemediationClosures  []RemediationClosure   `json:"remediation_closures"`
	Reviews              []ReviewDecision       `json:"reviews"`
	ReturnItems          []ReviewReturnItem     `json:"return_items"`
	Review               *ReviewDecision        `json:"review,omitempty"`
	Manifest             *Manifest              `json:"manifest,omitempty"`
	Credential           *FabricationCredential `json:"credential,omitempty"`
}
type PlateRevision struct {
	ID                   string            `json:"id"`
	CaseID               string            `json:"case_id"`
	SupersedesRevisionID string            `json:"supersedes_revision_id"`
	RevisionNo           int               `json:"revision_no"`
	BrailleCells         string            `json:"braille_cells"`
	DotSpacingMM         float64           `json:"dot_spacing_mm"`
	DotHeightMM          float64           `json:"dot_height_mm"`
	RaisedTextHeightMM   float64           `json:"raised_text_height_mm"`
	BevelRadiusMM        float64           `json:"bevel_radius_mm"`
	ContrastRatio        float64           `json:"contrast_ratio"`
	MountingHeightMM     float64           `json:"mounting_height_mm"`
	MaterialCode         string            `json:"material_code"`
	EvidenceDigests      []string          `json:"evidence_digests,omitempty"`
	Evidence             []EvidenceItem    `json:"evidence"`
	ImpactRuleIDs        []string          `json:"impact_rule_ids,omitempty"`
	Changes              []FieldDiff       `json:"changes,omitempty"`
	RemediationItems     []RemediationItem `json:"remediation_items,omitempty"`
	ReviewResponses      []ReviewResponse  `json:"review_responses,omitempty"`
	SubmittedBy          string            `json:"submitted_by"`
	SubmittedAt          time.Time         `json:"submitted_at"`
	Measurement          *Measurement      `json:"measurement,omitempty"`
}
type EvidenceItem struct {
	Kind        string `json:"kind"`
	Digest      string `json:"digest"`
	Description string `json:"description"`
}
type CoverageItem struct {
	Kind        string `json:"kind"`
	RuleID      string `json:"rule_id"`
	Label       string `json:"label"`
	Status      string `json:"status"`
	Digest      string `json:"digest,omitempty"`
	Description string `json:"description,omitempty"`
}
type Measurement struct {
	DotSpacingMM, DotHeightMM, RaisedTextHeightMM, BevelRadiusMM, ContrastRatio, MountingHeightMM float64
	EvidenceSummary                                                                               string
}
type ComplianceFinding struct {
	ID                     string    `json:"id"`
	CaseID                 string    `json:"case_id"`
	RevisionID             string    `json:"revision_id"`
	RuleID                 string    `json:"rule_id"`
	RuleVersion            string    `json:"rule_version"`
	Severity               string    `json:"severity"`
	Status                 string    `json:"status"`
	Threshold              string    `json:"threshold"`
	ActualValue            string    `json:"actual_value"`
	Explanation            string    `json:"explanation"`
	InputDigest            string    `json:"input_digest"`
	SupersededByRevisionID string    `json:"superseded_by_revision_id,omitempty"`
	EvaluatedAt            time.Time `json:"evaluated_at"`
}
type CheckRun struct {
	ID              string              `json:"id"`
	RevisionID      string              `json:"revision_id"`
	StandardVersion string              `json:"standard_version"`
	RuleSetVersion  string              `json:"rule_set_version"`
	InputDigest     string              `json:"input_digest"`
	ResultDigest    string              `json:"result_digest"`
	RetriedFrom     string              `json:"retried_from,omitempty"`
	Findings        []ComplianceFinding `json:"findings"`
	Coverage        []CoverageItem      `json:"coverage"`
	CreatedAt       time.Time           `json:"created_at"`
}
type RuleMigration struct {
	RuleID         string `json:"rule_id"`
	Before         string `json:"before"`
	After          string `json:"after"`
	Classification string `json:"classification"`
}
type RemediationItem struct {
	FindingID     string   `json:"finding_id"`
	Explanation   string   `json:"explanation"`
	PlannedFields []string `json:"planned_fields"`
}
type RemediationClosure struct {
	FindingID      string      `json:"finding_id"`
	RuleID         string      `json:"rule_id"`
	FromRevisionID string      `json:"from_revision_id"`
	ToRevisionID   string      `json:"to_revision_id"`
	NewFindingID   string      `json:"new_finding_id,omitempty"`
	Status         string      `json:"status"`
	Reason         string      `json:"reason"`
	FieldChanges   []FieldDiff `json:"field_changes"`
}
type ReviewResponse struct {
	ReturnItemID   string    `json:"return_item_id"`
	Response       string    `json:"response"`
	EvidenceDigest string    `json:"evidence_digest"`
	At             time.Time `json:"at"`
}
type ReviewReturnItem struct {
	ID              string           `json:"id"`
	DecisionID      string           `json:"decision_id"`
	RevisionID      string           `json:"revision_id,omitempty"`
	FindingID       string           `json:"finding_id,omitempty"`
	Category        string           `json:"category"`
	Description     string           `json:"description"`
	ResponsibleRole string           `json:"responsible_role"`
	Status          string           `json:"status"`
	Responses       []ReviewResponse `json:"responses"`
}
type ReviewDecision struct {
	ID               string             `json:"id"`
	ReviewerID       string             `json:"reviewer_id"`
	RevisionID       string             `json:"revision_id"`
	Decision         string             `json:"decision"`
	Reason           string             `json:"reason"`
	Items            []ReviewReturnItem `json:"items"`
	ConfirmedItemIDs []string           `json:"confirmed_item_ids,omitempty"`
	At               time.Time          `json:"at"`
}
type Manifest struct {
	CaseID      string              `json:"case_id"`
	RevisionID  string              `json:"revision_id"`
	Fields      map[string]string   `json:"fields"`
	Evidence    []EvidenceItem      `json:"evidence"`
	Findings    []ComplianceFinding `json:"findings"`
	Reviews     []ReviewDecision    `json:"reviews"`
	ReturnItems []ReviewReturnItem  `json:"return_items"`
	Review      ReviewDecision      `json:"review"`
	Counts      map[string]int      `json:"counts"`
	Digest      string              `json:"digest"`
	FrozenAt    time.Time           `json:"frozen_at,omitempty"`
}
type FabricationCredential struct {
	CredentialID     string    `json:"credential_id"`
	CaseID           string    `json:"case_id"`
	RevisionID       string    `json:"revision_id"`
	ManifestDigest   string    `json:"manifest_digest"`
	IssuedBy         string    `json:"issued_by"`
	IssuedAt         time.Time `json:"issued_at"`
	VerificationCode string    `json:"verification_code"`
	SchemaVersion    string    `json:"schema_version"`
}

func (c *ReleaseCase) CanWrite() error {
	if c.Status == Frozen || c.Status == Authorized {
		return ErrFrozen
	}
	return nil
}
func (c *ReleaseCase) CurrentRevision() *PlateRevision {
	if len(c.Revisions) == 0 {
		return nil
	}
	return &c.Revisions[len(c.Revisions)-1]
}
func (c *ReleaseCase) AddRevision(r PlateRevision) error {
	if err := c.CanWrite(); err != nil {
		return err
	}
	if r.RevisionNo != len(c.Revisions)+1 {
		return fmt.Errorf("revision number")
	}
	if r.DotSpacingMM <= 0 || r.DotHeightMM <= 0 || r.RaisedTextHeightMM < 0 || r.BevelRadiusMM < 0 || r.ContrastRatio < 0 || r.MountingHeightMM < 0 {
		return fmt.Errorf("invalid measurement")
	}
	if len(c.Revisions) > 0 && r.SupersedesRevisionID != c.Revisions[len(c.Revisions)-1].ID {
		return fmt.Errorf("supersedes mismatch")
	}
	if len(c.Revisions) > 0 && !Changed(c.Revisions[len(c.Revisions)-1], r) && EvidenceEqual(c.Revisions[len(c.Revisions)-1].Evidence, r.Evidence) {
		return ErrEmptyRevision
	}
	c.Revisions = append(c.Revisions, r)
	c.Version++
	c.UpdatedAt = time.Now().UTC()
	return nil
}
func (c *ReleaseCase) ApplyFindings(fs []ComplianceFinding) error {
	if err := c.CanWrite(); err != nil {
		return err
	}
	c.Findings = append(c.Findings, fs...)
	hasBlock := false
	for _, f := range fs {
		if f.Severity == string(Block) && f.Status == string(Open) {
			hasBlock = true
		}
	}
	if hasBlock {
		c.Status = ReworkRequired
	} else {
		c.Status = Checked
	}
	c.Version++
	c.UpdatedAt = time.Now().UTC()
	return nil
}
func (c *ReleaseCase) Supersede(revID string) {
	for i := range c.Findings {
		if c.Findings[i].Status == string(Open) {
			c.Findings[i].Status = string(Superseded)
			c.Findings[i].SupersededByRevisionID = revID
		}
	}
}
func (c *ReleaseCase) ReviewBy(id, decision, reason string) error {
	if id == c.DesignerID || id == c.MeasurerID {
		return ErrUnauthorized
	}
	if c.Status == ReworkRequired {
		return ErrInvalidState
	}
	if decision == "REJECT" {
		c.Status = ReworkRequired
	} else if decision == "APPROVE" {
		for _, f := range c.Findings {
			if f.Severity == string(Block) && f.Status == string(Open) {
				return fmt.Errorf("blocking findings")
			}
		}
		c.Status = Approved
	} else {
		return fmt.Errorf("decision")
	}
	d := ReviewDecision{ID: NewID("review", len(c.Reviews)+1), ReviewerID: id, RevisionID: c.CurrentRevision().ID, Decision: decision, Reason: reason, At: time.Now().UTC()}
	c.Reviews = append(c.Reviews, d)
	c.Review = &c.Reviews[len(c.Reviews)-1]
	c.Version++
	return nil
}
func (c *ReleaseCase) Freeze() error {
	if c.Status != Approved {
		return ErrInvalidState
	}
	r := c.CurrentRevision()
	if r == nil {
		return ErrInvalidState
	}
	c.Manifest = BuildManifest(c, time.Now().UTC())
	c.Status = Frozen
	c.Version++
	return nil
}
func Digest(fields map[string]string, findings []ComplianceFinding, review *ReviewDecision) string {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(fields[k])
		b.WriteByte('\n')
	}
	fs := append([]ComplianceFinding(nil), findings...)
	sort.Slice(fs, func(i, j int) bool { return fs[i].ID < fs[j].ID })
	for _, f := range fs {
		b.WriteString(strings.Join([]string{f.ID, f.RuleID, f.Severity, f.Status, f.Threshold, f.ActualValue, f.InputDigest}, "|"))
		b.WriteByte('\n')
	}
	if review != nil {
		b.WriteString(review.ReviewerID + "|" + review.Decision + "|" + review.Reason)
	}
	h := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(h[:])
}
func Sign(secret, digest string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(digest))
	return hex.EncodeToString(h.Sum(nil))
}
func Verify(secret, digest, code string) bool {
	return hmac.Equal([]byte(Sign(secret, digest)), []byte(code))
}
func NewID(prefix string, n int) string {
	return prefix + "-" + strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + strconv.Itoa(n)
}
func Encode(v any) []byte { b, _ := json.Marshal(v); return b }
