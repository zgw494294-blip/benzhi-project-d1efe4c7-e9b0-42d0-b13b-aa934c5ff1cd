package application

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"tactile-review/internal/compliance"
	"tactile-review/internal/domain"
	"tactile-review/internal/store"
	"time"
)

type Service struct {
	Store  *store.Store
	Secret string
	Idem   *store.Idempotency
}

type CreateCaseCommand struct {
	BuildingZone         string `json:"building_zone"`
	InstallationLocation string `json:"installation_location"`
	AudienceProfile      string `json:"audience_profile"`
	StandardVersion      string `json:"standard_version"`
	DesignerID           string `json:"designer_id"`
	MeasurerID           string `json:"measurer_id"`
	IdempotencyKey       string `json:"idempotency_key"`
}
type CreatePreflight struct {
	Normalized        CreateCaseCommand        `json:"normalized"`
	Issues            []domain.ValidationIssue `json:"issues"`
	AvailableVersions []compliance.VersionInfo `json:"available_versions"`
	Ready             bool                     `json:"ready"`
}
type RevisionPreview struct {
	Revision      domain.PlateRevision  `json:"revision"`
	Fields        []domain.FieldDiff    `json:"fields"`
	ChangedFields []domain.FieldDiff    `json:"changed_fields"`
	ImpactRuleIDs []string              `json:"impact_rule_ids"`
	Coverage      []domain.CoverageItem `json:"coverage"`
}
type ReviewCommand struct {
	ExpectedVersion              int
	ReviewerID, Decision, Reason string
	Items                        []domain.ReviewReturnItem
	ConfirmItemIDs               []string
}
type VerificationResult struct {
	Valid            bool                       `json:"valid"`
	Status           string                     `json:"status"`
	CredentialID     string                     `json:"credential_id"`
	MismatchedFields []string                   `json:"mismatched_fields"`
	Checks           map[string]bool            `json:"checks"`
	Recent           []store.VerificationRecord `json:"recent"`
}

func New(s *store.Store) *Service {
	return &Service{Store: s, Secret: "tactile-review-local-secret", Idem: store.NewIdempotency()}
}

func (a *Service) PreflightCreate(cmd CreateCaseCommand) CreatePreflight {
	c := &domain.ReleaseCase{ID: "preflight", BuildingZone: cmd.BuildingZone, InstallationLocation: cmd.InstallationLocation, AudienceProfile: cmd.AudienceProfile, StandardVersion: cmd.StandardVersion, DesignerID: cmd.DesignerID, MeasurerID: cmd.MeasurerID, Version: 1}
	domain.NormalizeCase(c)
	cmd.BuildingZone, cmd.InstallationLocation, cmd.AudienceProfile = c.BuildingZone, c.InstallationLocation, c.AudienceProfile
	cmd.StandardVersion, cmd.DesignerID, cmd.MeasurerID = c.StandardVersion, c.DesignerID, c.MeasurerID
	issues := c.Validate()
	if !compliance.Supports(c.StandardVersion) {
		message := "标准版本未知"
		if _, ok := compliance.Version(c.StandardVersion); ok {
			message = "标准版本已停用"
		}
		issues = append(issues, domain.ValidationIssue{Field: "standard_version", Message: message, Blocking: true})
	}
	if strings.TrimSpace(cmd.IdempotencyKey) == "" {
		issues = append(issues, domain.ValidationIssue{Field: "idempotency_key", Message: "建档幂等键不能为空", Blocking: true})
	} else if len([]rune(cmd.IdempotencyKey)) > 128 {
		issues = append(issues, domain.ValidationIssue{Field: "idempotency_key", Message: "建档幂等键不能超过 128 个字符", Blocking: true})
	}
	return CreatePreflight{Normalized: cmd, Issues: issues, AvailableVersions: compliance.VersionCatalog(), Ready: len(issues) == 0}
}
func createDigest(cmd CreateCaseCommand) string {
	return domain.HashText(strings.Join([]string{cmd.BuildingZone, cmd.InstallationLocation, cmd.AudienceProfile, cmd.StandardVersion, cmd.DesignerID, cmd.MeasurerID}, "\x1f"))
}
func (a *Service) CreateWithCommand(cmd CreateCaseCommand) (*domain.ReleaseCase, bool, error) {
	p := a.PreflightCreate(cmd)
	if err := domain.IssuesError(p.Issues); err != nil {
		return nil, false, err
	}
	cmd = p.Normalized
	now := time.Now().UTC()
	c := &domain.ReleaseCase{ID: domain.NewID("case", 0), BuildingZone: cmd.BuildingZone, InstallationLocation: cmd.InstallationLocation, AudienceProfile: cmd.AudienceProfile, StandardVersion: cmd.StandardVersion, Status: domain.Draft, Version: 1, DesignerID: cmd.DesignerID, MeasurerID: cmd.MeasurerID, CreatedAt: now, UpdatedAt: now}
	return a.Store.CreateIdempotent(c, cmd.IdempotencyKey, createDigest(cmd))
}
func (a *Service) Create(zone, loc, aud, std, designer, measurer string) (*domain.ReleaseCase, error) {
	c, _, err := a.CreateWithCommand(CreateCaseCommand{BuildingZone: zone, InstallationLocation: loc, AudienceProfile: aud, StandardVersion: std, DesignerID: designer, MeasurerID: measurer, IdempotencyKey: domain.NewID("create", 0)})
	return c, err
}

func normalizeRevision(r domain.PlateRevision) domain.PlateRevision {
	r.BrailleCells = domain.NormalizeText(r.BrailleCells)
	r.MaterialCode = domain.NormalizeText(r.MaterialCode)
	r.SubmittedBy = domain.NormalizeActor(r.SubmittedBy)
	if len(r.Evidence) == 0 && len(r.EvidenceDigests) > 0 && r.Measurement != nil && domain.NormalizeText(r.Measurement.EvidenceSummary) != "" {
		for _, kind := range domain.EvidenceKinds {
			r.Evidence = append(r.Evidence, domain.EvidenceItem{Kind: kind, Digest: domain.HashText(r.EvidenceDigests[0] + ":" + kind), Description: domain.NormalizeText(r.Measurement.EvidenceSummary)})
		}
	}
	r.Evidence = domain.NormalizeEvidence(r.Evidence)
	r.EvidenceDigests = nil
	for _, e := range r.Evidence {
		r.EvidenceDigests = append(r.EvidenceDigests, e.Digest)
	}
	return r
}
func (a *Service) PreviewRevision(id string, expected int, r domain.PlateRevision) (RevisionPreview, error) {
	c, err := a.Store.Get(id)
	if err != nil {
		return RevisionPreview{}, err
	}
	if err = EnsureExpected(c.Version, expected); err != nil {
		return RevisionPreview{}, err
	}
	if err = c.CanWrite(); err != nil {
		return RevisionPreview{}, err
	}
	r = normalizeRevision(r)
	r.CaseID = id
	current := c.CurrentRevision()
	if current == nil {
		if r.RevisionNo == 0 {
			r.RevisionNo = 1
		}
		if r.RevisionNo != 1 || r.SupersedesRevisionID != "" {
			return RevisionPreview{}, fmt.Errorf("首个修订的修订号必须为 1 且不能替代其他修订")
		}
	} else {
		if r.RevisionNo == 0 {
			r.RevisionNo = current.RevisionNo + 1
		}
		if r.SupersedesRevisionID == "" {
			r.SupersedesRevisionID = current.ID
		}
		if r.RevisionNo != current.RevisionNo+1 {
			return RevisionPreview{}, fmt.Errorf("修订号必须连续")
		}
		if r.SupersedesRevisionID != current.ID {
			return RevisionPreview{}, fmt.Errorf("supersedes_revision_id 必须指向当前修订")
		}
	}
	if err = domain.IssuesError(r.Validate()); err != nil {
		return RevisionPreview{}, err
	}
	if current != nil && !domain.Changed(*current, r) && domain.EvidenceEqual(current.Evidence, r.Evidence) {
		return RevisionPreview{}, domain.ErrEmptyRevision
	}
	if err = validateRemediation(c, r); err != nil {
		return RevisionPreview{}, err
	}
	if err = validateResponses(c, r.ReviewResponses); err != nil {
		return RevisionPreview{}, err
	}
	fields := []domain.FieldDiff{}
	if current != nil {
		fields = domain.RevisionDiff(*current, r, true)
	}
	changed := []domain.FieldDiff{}
	for _, d := range fields {
		if d.Classification != "UNCHANGED" {
			changed = append(changed, d)
		}
	}
	r.Changes = changed
	r.ImpactRuleIDs = domain.ImpactedRules(changed)
	return RevisionPreview{Revision: r, Fields: fields, ChangedFields: changed, ImpactRuleIDs: r.ImpactRuleIDs, Coverage: compliance.Coverage(r, c.StandardVersion)}, nil
}
func validateRemediation(c *domain.ReleaseCase, r domain.PlateRevision) error {
	if c.HasOpenBlocking() && len(r.RemediationItems) == 0 {
		return fmt.Errorf("存在 OPEN BLOCK 时必须选择整改目标")
	}
	seen := map[string]bool{}
	for i, item := range r.RemediationItems {
		if seen[item.FindingID] {
			return fmt.Errorf("remediation_items[%d] 重复引用发现", i)
		}
		seen[item.FindingID] = true
		if domain.NormalizeText(item.Explanation) == "" {
			return fmt.Errorf("remediation_items[%d].explanation 不能为空", i)
		}
		var found *domain.ComplianceFinding
		for j := range c.Findings {
			if c.Findings[j].ID == item.FindingID {
				found = &c.Findings[j]
				break
			}
		}
		if found == nil || c.CurrentRevision() == nil || found.RevisionID != c.CurrentRevision().ID || found.Severity != string(domain.Block) || found.Status != string(domain.Open) {
			return fmt.Errorf("remediation_items[%d] 必须引用当前修订的 OPEN BLOCK", i)
		}
		if len(item.PlannedFields) == 0 {
			return fmt.Errorf("remediation_items[%d].planned_fields 不能为空", i)
		}
	}
	return nil
}
func validateResponses(c *domain.ReleaseCase, responses []domain.ReviewResponse) error {
	seen := map[string]bool{}
	for i, response := range responses {
		if seen[response.ReturnItemID] {
			return fmt.Errorf("review_responses[%d] 重复", i)
		}
		seen[response.ReturnItemID] = true
		if domain.NormalizeText(response.Response) == "" || response.EvidenceDigest == "" {
			return fmt.Errorf("review_responses[%d] 必须包含回应和证据", i)
		}
		found := false
		for _, item := range c.ReturnItems {
			if item.ID == response.ReturnItemID && item.Status != "CLOSED" {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("review_responses[%d] 未关联待关闭退回项", i)
		}
	}
	return nil
}
func (a *Service) AddRevision(id string, expected int, r domain.PlateRevision, key string) (*domain.ReleaseCase, error) {
	if strings.TrimSpace(key) == "" {
		return nil, fmt.Errorf("idempotency_key 不能为空")
	}
	// Normalize the input first so the request digest is stable across retries.
	// PreviewRevision normalizes again, but computing the digest from the
	// normalized form keeps the idempotency boundary independent of the
	// aggregate version, which has already advanced after the first successful
	// submission.
	nr := normalizeRevision(r)
	nr.CaseID = id
	requestDigest := domain.HashText(strings.Join([]string{id, strconv.Itoa(expected), domain.RevisionInputDigest(nr)}, "\x1f"))
	// Replay path: a prior submission under the same key already succeeded.
	// This check must run before any version-sensitive validation so a retry
	// using the original expectedVersion returns the first result instead of a
	// version conflict, and so a distinct request under the same key is rejected
	// rather than being treated as a new revision.
	if replay, reused, err := a.Store.RevisionIdempotencyReplay(id, key, requestDigest); err != nil {
		return nil, err
	} else if reused {
		return replay, nil
	}
	p, err := a.PreviewRevision(id, expected, r)
	if err != nil {
		return nil, err
	}
	c, err := a.Store.Get(id)
	if err != nil {
		return nil, err
	}
	r = p.Revision
	r.ID = domain.NewID("rev", r.RevisionNo)
	r.SubmittedAt = time.Now().UTC()
	if err = c.AddRevision(r); err != nil {
		return nil, err
	}
	checkRunID := ""
	if len(r.RemediationItems) > 0 {
		findings := compliance.Evaluate(c, r)
		inputDigest := domain.RevisionInputDigest(r)
		run := domain.CheckRun{ID: domain.NewID("check", len(c.CheckRuns)+1), RevisionID: r.ID, StandardVersion: c.StandardVersion, RuleSetVersion: "rules-v1", InputDigest: inputDigest, ResultDigest: domain.HashText(compliance.CanonicalFindings(findings)), Findings: findings, Coverage: compliance.Coverage(r, c.StandardVersion), CreatedAt: time.Now().UTC()}
		c.CheckRuns = append(c.CheckRuns, run)
		c.Findings = append(c.Findings, findings...)
		applyRemediationClosures(c, r, findings, run.Coverage)
		if hasOpenBlockForRevision(c, r.ID) {
			c.Status = domain.ReworkRequired
		} else {
			c.Status = domain.Checked
		}
		checkRunID = run.ID
	}
	for _, response := range r.ReviewResponses {
		for i := range c.ReturnItems {
			if c.ReturnItems[i].ID == response.ReturnItemID {
				response.At = r.SubmittedAt
				c.ReturnItems[i].Responses = append(c.ReturnItems[i].Responses, response)
				c.ReturnItems[i].Status = "RESPONDED"
			}
		}
	}
	result, _, err := a.Store.SaveRevisionIdempotent(c, expected, key, requestDigest, "REVISION_CREATED", r.SubmittedBy, map[string]string{"revision_id": r.ID, "input_digest": domain.RevisionInputDigest(r), "idempotency_key": key, "check_run_id": checkRunID})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Service) Check(id string, expected int) (*domain.ReleaseCase, error) {
	c, err := a.Store.Get(id)
	if err != nil {
		return nil, err
	}
	if err = EnsureExpected(c.Version, expected); err != nil {
		return nil, err
	}
	if err = c.CanWrite(); err != nil {
		return nil, err
	}
	r := c.CurrentRevision()
	if r == nil {
		return nil, fmt.Errorf("尚无可校核修订")
	}
	findings := compliance.Evaluate(c, *r)
	inputDigest := domain.RevisionInputDigest(*r)
	resultDigest := domain.HashText(compliance.CanonicalFindings(findings))
	run := domain.CheckRun{ID: domain.NewID("check", len(c.CheckRuns)+1), RevisionID: r.ID, StandardVersion: c.StandardVersion, RuleSetVersion: "rules-v1", InputDigest: inputDigest, ResultDigest: resultDigest, Findings: findings, Coverage: compliance.Coverage(*r, c.StandardVersion), CreatedAt: time.Now().UTC()}
	first := true
	for _, old := range c.CheckRuns {
		if old.RevisionID == run.RevisionID && old.StandardVersion == run.StandardVersion && old.RuleSetVersion == run.RuleSetVersion && old.InputDigest == run.InputDigest {
			if old.ResultDigest != run.ResultDigest {
				return nil, domain.ErrDeterminism
			}
			run.RetriedFrom = old.ID
			first = false
			break
		}
	}
	c.CheckRuns = append(c.CheckRuns, run)
	if first {
		c.Findings = append(c.Findings, findings...)
		applyRemediationClosures(c, *r, findings, run.Coverage)
		if hasOpenBlockForRevision(c, r.ID) {
			c.Status = domain.ReworkRequired
		} else {
			c.Status = domain.Checked
		}
		c.Version++
		c.UpdatedAt = time.Now().UTC()
	}
	action := "CHECK_COMPLETED"
	if !first {
		action = "CHECK_RETRIED"
	}
	if err = a.Store.SaveWithAudit(c, expected, action, r.SubmittedBy, map[string]string{"run_id": run.ID, "input_digest": inputDigest, "result_digest": resultDigest, "retry_of": run.RetriedFrom}); err != nil {
		return nil, err
	}
	return c, nil
}
func hasOpenBlockForRevision(c *domain.ReleaseCase, revisionID string) bool {
	for _, f := range c.Findings {
		if f.RevisionID == revisionID && f.Severity == string(domain.Block) && f.Status == string(domain.Open) {
			return true
		}
	}
	return false
}
func applyRemediationClosures(c *domain.ReleaseCase, r domain.PlateRevision, newFindings []domain.ComplianceFinding, coverage []domain.CoverageItem) {
	pass := map[string]domain.ComplianceFinding{}
	for _, f := range newFindings {
		if f.Severity == string(domain.Pass) {
			pass[f.RuleID] = f
		}
	}
	targets := map[string]bool{}
	for _, item := range r.RemediationItems {
		targets[item.FindingID] = true
	}
	for i := range c.Findings {
		old := &c.Findings[i]
		if !targets[old.ID] || old.Status != string(domain.Open) || old.Severity != string(domain.Block) {
			continue
		}
		ruleID := strings.TrimPrefix(old.RuleID, "evidence-")
		newFinding, passed := pass[ruleID]
		covered := compliance.CoveredForRule(coverage, ruleID)
		closure := domain.RemediationClosure{FindingID: old.ID, RuleID: old.RuleID, FromRevisionID: old.RevisionID, ToRevisionID: r.ID, Status: "OPEN", Reason: "新结果尚未通过或新实测证据不足", FieldChanges: r.Changes}
		if passed && covered {
			old.Status = string(domain.Superseded)
			old.SupersededByRevisionID = r.ID
			closure.Status, closure.Reason, closure.NewFindingID = "SUPERSEDED", "新修订规则已通过且相关实测证据完整", newFinding.ID
		}
		c.RemediationClosures = append(c.RemediationClosures, closure)
	}
}
func (a *Service) CompareAdjacent(id, revisionID string) ([]domain.RuleMigration, error) {
	c, err := a.Store.Get(id)
	if err != nil {
		return nil, err
	}
	idx := -1
	for i, r := range c.Revisions {
		if r.ID == revisionID {
			idx = i
		}
	}
	if idx <= 0 {
		return []domain.RuleMigration{}, nil
	}
	before, after := latestRun(c, c.Revisions[idx-1].ID), latestRun(c, revisionID)
	if before == nil || after == nil {
		return nil, fmt.Errorf("相邻修订尚未完成校核")
	}
	return migrate(before.Findings, after.Findings), nil
}
func latestRun(c *domain.ReleaseCase, revisionID string) *domain.CheckRun {
	for i := len(c.CheckRuns) - 1; i >= 0; i-- {
		if c.CheckRuns[i].RevisionID == revisionID {
			return &c.CheckRuns[i]
		}
	}
	return nil
}
func migrate(before, after []domain.ComplianceFinding) []domain.RuleMigration {
	bm, am := map[string]string{}, map[string]string{}
	for _, f := range before {
		bm[f.RuleID] = f.Severity
	}
	for _, f := range after {
		am[f.RuleID] = f.Severity
	}
	ids := map[string]bool{}
	for id := range bm {
		ids[id] = true
	}
	for id := range am {
		ids[id] = true
	}
	keys := []string{}
	for id := range ids {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	rank := map[string]int{"BLOCK": 0, "WARN": 1, "PASS": 2}
	out := make([]domain.RuleMigration, 0, len(keys))
	for _, id := range keys {
		kind := "UNCHANGED"
		if bm[id] == "" {
			kind = "ADDED"
		} else if rank[am[id]] > rank[bm[id]] {
			kind = "IMPROVED"
		} else if rank[am[id]] < rank[bm[id]] {
			kind = "WORSENED"
		}
		out = append(out, domain.RuleMigration{RuleID: id, Before: bm[id], After: am[id], Classification: kind})
	}
	return out
}

func (a *Service) ReviewStructured(id string, cmd ReviewCommand) (*domain.ReleaseCase, error) {
	c, err := a.Store.Get(id)
	if err != nil {
		return nil, err
	}
	if err = EnsureExpected(c.Version, cmd.ExpectedVersion); err != nil {
		return nil, err
	}
	if err = c.CanWrite(); err != nil {
		return nil, err
	}
	reviewer := domain.NormalizeActor(cmd.ReviewerID)
	if !domain.IsIndependent(c, reviewer) {
		return nil, domain.ErrUnauthorized
	}
	r := c.CurrentRevision()
	if r == nil {
		return nil, domain.ErrInvalidState
	}
	decision := strings.ToUpper(domain.NormalizeText(cmd.Decision))
	if decision == "REJECT" {
		decision = "RETURN"
	}
	d := domain.ReviewDecision{ID: domain.NewID("review", len(c.Reviews)+1), ReviewerID: reviewer, RevisionID: r.ID, Decision: decision, Reason: domain.NormalizeText(cmd.Reason), ConfirmedItemIDs: append([]string(nil), cmd.ConfirmItemIDs...), At: time.Now().UTC()}
	switch decision {
	case "RETURN":
		if len(cmd.Items) == 0 {
			return nil, fmt.Errorf("退回决定至少包含一个结构化原因")
		}
		for i, item := range cmd.Items {
			item.ID = domain.NewID("return", i+1)
			item.DecisionID = d.ID
			item.Status = "PENDING"
			item.Description = domain.NormalizeText(item.Description)
			item.ResponsibleRole = strings.ToUpper(domain.NormalizeText(item.ResponsibleRole))
			item.Category = strings.ToUpper(domain.NormalizeText(item.Category))
			if err = validateReturnItem(c, r.ID, item, i); err != nil {
				return nil, err
			}
			d.Items = append(d.Items, item)
			c.ReturnItems = append(c.ReturnItems, item)
		}
		c.Status = domain.ReworkRequired
	case "APPROVE":
		if latestRun(c, r.ID) == nil {
			return nil, fmt.Errorf("当前修订尚未完成校核")
		}
		if c.HasOpenBlocking() || hasOpenBlockForRevision(c, r.ID) {
			return nil, fmt.Errorf("存在 OPEN BLOCK，不能批准")
		}
		confirm := map[string]bool{}
		for _, x := range cmd.ConfirmItemIDs {
			confirm[x] = true
		}
		for i := range c.ReturnItems {
			if c.ReturnItems[i].Status == "RESPONDED" && confirm[c.ReturnItems[i].ID] {
				c.ReturnItems[i].Status = "CLOSED"
			}
		}
		for _, item := range c.ReturnItems {
			if item.Status != "CLOSED" {
				return nil, fmt.Errorf("退回项 %s 尚未确认关闭", item.ID)
			}
		}
		c.Status = domain.Approved
	default:
		return nil, fmt.Errorf("未知复核决定")
	}
	c.Reviews = append(c.Reviews, d)
	c.Review = &c.Reviews[len(c.Reviews)-1]
	c.ReviewerID = reviewer
	c.Version++
	c.UpdatedAt = time.Now().UTC()
	if err = a.Store.SaveWithAudit(c, cmd.ExpectedVersion, "REVIEW_"+decision, reviewer, map[string]string{"decision_id": d.ID}); err != nil {
		return nil, err
	}
	return c, nil
}
func validateReturnItem(c *domain.ReleaseCase, currentRevisionID string, item domain.ReviewReturnItem, i int) error {
	validCategory := map[string]bool{"EVIDENCE_INSUFFICIENT": true, "DESCRIPTION_UNCLEAR": true, "MEASUREMENT_QUESTION": true, "OTHER": true}
	validRole := map[string]bool{"DESIGNER": true, "MEASURER": true}
	if !validCategory[item.Category] {
		return fmt.Errorf("items[%d].category 未知", i)
	}
	if item.Description == "" {
		return fmt.Errorf("items[%d].description 不能为空", i)
	}
	if !validRole[item.ResponsibleRole] {
		return fmt.Errorf("items[%d].responsible_role 未知", i)
	}
	if item.RevisionID == "" && item.FindingID == "" {
		return fmt.Errorf("items[%d] 必须关联修订或发现", i)
	}
	if item.RevisionID != "" && item.RevisionID != currentRevisionID {
		return fmt.Errorf("items[%d] 修订不属于当前修订", i)
	}
	if item.FindingID != "" {
		found := false
		for _, f := range c.Findings {
			if f.ID == item.FindingID && f.RevisionID == currentRevisionID {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("items[%d] 发现不属于当前案当前修订", i)
		}
	}
	return nil
}
func (a *Service) Review(id string, expected int, reviewer, decision, reason string) (*domain.ReleaseCase, error) {
	return a.ReviewStructured(id, ReviewCommand{ExpectedVersion: expected, ReviewerID: reviewer, Decision: decision, Reason: reason})
}

func (a *Service) PreviewFreeze(id string) (*domain.Manifest, error) {
	c, err := a.Store.Get(id)
	if err != nil {
		return nil, err
	}
	if c.Status != domain.Approved {
		return nil, domain.ErrInvalidState
	}
	if c.HasOpenBlocking() {
		return nil, fmt.Errorf("存在 OPEN BLOCK")
	}
	for _, item := range c.ReturnItems {
		if item.Status != "CLOSED" {
			return nil, fmt.Errorf("存在未关闭退回项")
		}
	}
	return domain.BuildManifest(c, time.Time{}), nil
}
func (a *Service) FreezeConfirm(id string, expected int, previewDigest string) (*domain.ReleaseCase, error) {
	c, err := a.Store.Get(id)
	if err != nil {
		return nil, err
	}
	if c.Status == domain.Frozen || c.Status == domain.Authorized {
		if c.Manifest != nil && c.Manifest.Digest == previewDigest {
			return c, nil
		}
		return nil, domain.ErrDigestConflict
	}
	if err = EnsureExpected(c.Version, expected); err != nil {
		return nil, err
	}
	preview, err := a.PreviewFreeze(id)
	if err != nil {
		return nil, err
	}
	if previewDigest == "" || preview.Digest != previewDigest {
		return nil, domain.ErrDigestConflict
	}
	manifest := domain.BuildManifest(c, time.Now().UTC())
	if manifest.Digest != previewDigest {
		return nil, domain.ErrDigestConflict
	}
	c.Manifest = manifest
	c.Status = domain.Frozen
	c.Version++
	c.UpdatedAt = time.Now().UTC()
	if err = a.Store.SaveWithAudit(c, expected, "MANIFEST_FROZEN", c.ReviewerID, map[string]string{"manifest_digest": manifest.Digest}); err != nil {
		return nil, err
	}
	return c, nil
}
func (a *Service) Freeze(id string, expected int) (*domain.ReleaseCase, error) {
	p, err := a.PreviewFreeze(id)
	if err != nil {
		return nil, err
	}
	return a.FreezeConfirm(id, expected, p.Digest)
}
func (a *Service) Issue(id string, expected int, issuer string) (*domain.FabricationCredential, error) {
	c, err := a.Store.Get(id)
	if err != nil {
		return nil, err
	}
	if c.Status == domain.Authorized && c.Credential != nil {
		return c.Credential, nil
	}
	if err = EnsureExpected(c.Version, expected); err != nil {
		return nil, err
	}
	if err = CanIssue(c); err != nil {
		return nil, err
	}
	cred := &domain.FabricationCredential{CredentialID: domain.NewID("cred", 0), CaseID: id, RevisionID: c.Manifest.RevisionID, ManifestDigest: c.Manifest.Digest, IssuedBy: domain.NormalizeActor(issuer), IssuedAt: time.Now().UTC(), SchemaVersion: "v1"}
	cred.VerificationCode = domain.Sign(a.Secret, domain.CredentialPayload(*cred))
	c.Credential = cred
	c.Status = domain.Authorized
	c.Version++
	c.UpdatedAt = time.Now().UTC()
	if err = a.Store.SaveWithAudit(c, expected, "CREDENTIAL_ISSUED", cred.IssuedBy, map[string]string{"credential_id": cred.CredentialID}); err != nil {
		return nil, err
	}
	return cred, nil
}

func (a *Service) VerifyCredential(targetCaseID string, input domain.FabricationCredential) (VerificationResult, error) {
	result := VerificationResult{Status: "TAMPERED", CredentialID: input.CredentialID, Checks: map[string]bool{}}
	raw, _ := json.Marshal(input)
	inputDigest := domain.HashText(string(raw))
	formatOK := input.SchemaVersion == "v1" && input.CredentialID != "" && input.CaseID != "" && input.RevisionID != "" && input.ManifestDigest != "" && !input.IssuedAt.IsZero()
	if b, err := hex.DecodeString(input.VerificationCode); err != nil || len(b) != 32 {
		formatOK = false
	}
	result.Checks["schema_and_format"] = formatOK
	if !formatOK {
		return a.recordVerification(targetCaseID, inputDigest, result)
	}
	issuedCase, stored, lookupErr := a.Store.FindCredential(input.CredentialID)
	if errors.Is(lookupErr, domain.ErrNotFound) {
		result.Status = "UNKNOWN"
		return a.recordVerification(targetCaseID, inputDigest, result)
	}
	if lookupErr != nil {
		return result, lookupErr
	}
	hmacOK := formatOK && domain.Verify(a.Secret, domain.CredentialPayload(input), input.VerificationCode)
	result.Checks["hmac"] = hmacOK
	if !hmacOK {
		return a.recordVerification(targetCaseID, inputDigest, result)
	}
	checks := map[string]bool{"credential_id": input.CredentialID == stored.CredentialID, "case_id": input.CaseID == stored.CaseID && input.CaseID == targetCaseID, "revision_id": input.RevisionID == stored.RevisionID, "manifest_digest": input.ManifestDigest == stored.ManifestDigest}
	for k, v := range checks {
		result.Checks[k] = v
		if !v {
			result.MismatchedFields = append(result.MismatchedFields, k)
		}
	}
	if issuedCase.Manifest == nil || issuedCase.Manifest.Digest != stored.ManifestDigest {
		result.MismatchedFields = append(result.MismatchedFields, "stored_manifest_digest")
	}
	sort.Strings(result.MismatchedFields)
	if len(result.MismatchedFields) > 0 {
		result.Status = "MISMATCHED"
	} else {
		result.Status = "VALID"
		result.Valid = true
	}
	return a.recordVerification(targetCaseID, inputDigest, result)
}
func (a *Service) recordVerification(targetCaseID, inputDigest string, result VerificationResult) (VerificationResult, error) {
	err := a.Store.AppendVerification(store.VerificationRecord{TargetCaseID: targetCaseID, CredentialID: result.CredentialID, Status: result.Status, InputDigest: inputDigest, MismatchedFields: result.MismatchedFields, At: time.Now().UTC()})
	if err != nil {
		return result, err
	}
	result.Recent, _ = a.Store.VerificationHistory(result.CredentialID, 10)
	return result, nil
}
func (a *Service) Verify(id, code string) (bool, string, error) {
	c, err := a.Store.Get(id)
	if err != nil {
		return false, "UNKNOWN", err
	}
	if c.Credential == nil {
		return false, "UNKNOWN", nil
	}
	input := *c.Credential
	input.VerificationCode = code
	r, err := a.VerifyCredential(id, input)
	return r.Valid, r.Status, err
}
