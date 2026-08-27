package application

import (
	"errors"
	"path/filepath"
	"tactile-review/internal/domain"
	"tactile-review/internal/store"
	"testing"
)

func testService(t *testing.T) *Service {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "workflow.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return New(s)
}

func completeEvidence(seed string) []domain.EvidenceItem {
	out := make([]domain.EvidenceItem, 0, len(domain.EvidenceKinds))
	for _, kind := range domain.EvidenceKinds {
		out = append(out, domain.EvidenceItem{Kind: kind, Digest: domain.HashText(seed + kind), Description: "现场实测 " + kind})
	}
	return out
}

func revision(spacing, contrast float64, seed string) domain.PlateRevision {
	return domain.PlateRevision{BrailleCells: "A1", DotSpacingMM: spacing, DotHeightMM: .7, RaisedTextHeightMM: 1, BevelRadiusMM: 1, MaterialCode: "PVC", ContrastRatio: contrast, MountingHeightMM: 900, Evidence: completeEvidence(seed), SubmittedBy: "designer"}
}

func createCase(t *testing.T, a *Service, key string) *domain.ReleaseCase {
	t.Helper()
	c, _, err := a.CreateWithCommand(CreateCaseCommand{BuildingZone: " A 区 ", InstallationLocation: " 东门 ", AudienceProfile: "视障人士", StandardVersion: "gb-2024", DesignerID: "Designer", MeasurerID: "Measurer", IdempotencyKey: key})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCreatePreflightAndPersistentIdempotency(t *testing.T) {
	a := testService(t)
	cmd := CreateCaseCommand{BuildingZone: " A  区 ", InstallationLocation: " 东 门 ", AudienceProfile: "视障人士", StandardVersion: "gb-2024", DesignerID: "Designer", MeasurerID: "Measurer", IdempotencyKey: "create-1"}
	c, reused, err := a.CreateWithCommand(cmd)
	if err != nil || reused {
		t.Fatalf("create: %v reused=%v", err, reused)
	}
	if c.BuildingZone != "A 区" || c.DesignerID != "designer" || c.Version != 1 || c.Status != domain.Draft {
		t.Fatalf("not normalized: %#v", c)
	}
	again, reused, err := a.CreateWithCommand(cmd)
	if err != nil || !reused || again.ID != c.ID {
		t.Fatalf("retry: %#v %v %v", again, reused, err)
	}
	cmd.InstallationLocation = "西门"
	_, _, err = a.CreateWithCommand(cmd)
	if !errors.Is(err, domain.ErrIdempotencyConflict) {
		t.Fatalf("want conflict, got %v", err)
	}
	cases, _ := a.Store.List()
	audit, _ := a.Store.Audit()
	if len(cases) != 1 || len(audit) != 1 {
		t.Fatalf("partial write: cases=%d audit=%d", len(cases), len(audit))
	}
	bad := a.PreflightCreate(CreateCaseCommand{StandardVersion: "GB-2023", DesignerID: "same", MeasurerID: "same"})
	if bad.Ready || len(bad.Issues) < 4 || len(bad.AvailableVersions) < 2 {
		t.Fatalf("preflight insufficient: %#v", bad)
	}
}

func TestRevisionEvidenceCheckHistoryAndMigration(t *testing.T) {
	a := testService(t)
	c := createCase(t, a, "case-2")
	first := revision(2.0, 4, "one")
	var err error
	c, err = a.AddRevision(c.ID, c.Version, first, "rev-1")
	if err != nil {
		t.Fatal(err)
	}
	oldEvidence := append([]domain.EvidenceItem(nil), c.CurrentRevision().Evidence...)
	c, err = a.Check(c.ID, c.Version)
	if err != nil {
		t.Fatal(err)
	}
	firstRun := c.CheckRuns[0]
	c, err = a.Check(c.ID, c.Version)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.CheckRuns) != 2 || c.CheckRuns[1].RetriedFrom != firstRun.ID || c.CheckRuns[1].ResultDigest != firstRun.ResultDigest || c.Version != 3 {
		t.Fatalf("retry history invalid: %#v", c.CheckRuns)
	}
	var spacingFinding string
	for _, f := range c.Findings {
		if f.RevisionID == c.Revisions[0].ID && f.RuleID == "dot-spacing" {
			spacingFinding = f.ID
		}
	}
	second := revision(2.5, 4, "two")
	second.RemediationItems = []domain.RemediationItem{{FindingID: spacingFinding, Explanation: "调整点距并复测", PlannedFields: []string{"dot_spacing_mm"}}}
	preview, err := a.PreviewRevision(c.ID, c.Version, second)
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.ChangedFields) != 7 || preview.ChangedFields[0].Field != "dot_spacing_mm" || len(preview.ImpactRuleIDs) != 6 {
		t.Fatalf("diff order/content: %#v", preview.ChangedFields)
	}
	c, err = a.AddRevision(c.ID, c.Version, second, "rev-2")
	if err != nil {
		t.Fatal(err)
	}
	if !domain.EvidenceEqual(oldEvidence, c.Revisions[0].Evidence) {
		t.Fatal("old evidence changed")
	}
	c, err = a.Check(c.ID, c.Version)
	if err != nil {
		t.Fatal(err)
	}
	migration, err := a.CompareAdjacent(c.ID, c.CurrentRevision().ID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range migration {
		if m.RuleID == "dot-spacing" && m.Before == "BLOCK" && m.After == "PASS" && m.Classification == "IMPROVED" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing migration: %#v", migration)
	}
	empty := *c.CurrentRevision()
	empty.ID = ""
	empty.RevisionNo++
	empty.SupersedesRevisionID = c.CurrentRevision().ID
	empty.SubmittedAt = empty.SubmittedAt.Add(1)
	if _, err = a.PreviewRevision(c.ID, c.Version, empty); !errors.Is(err, domain.ErrEmptyRevision) {
		t.Fatalf("want empty revision, got %v", err)
	}
}

func TestReviewFreezeAndCredentialVerification(t *testing.T) {
	a := testService(t)
	c := createCase(t, a, "case-3")
	var err error
	c, err = a.AddRevision(c.ID, c.Version, revision(2.5, 4, "a"), "r1")
	if err != nil {
		t.Fatal(err)
	}
	c, err = a.Check(c.ID, c.Version)
	if err != nil {
		t.Fatal(err)
	}
	revID := c.CurrentRevision().ID
	items := []domain.ReviewReturnItem{{RevisionID: revID, Category: "EVIDENCE_INSUFFICIENT", Description: "点高证据拍摄角度不足", ResponsibleRole: "MEASURER"}, {RevisionID: revID, Category: "DESCRIPTION_UNCLEAR", Description: "材质说明不清", ResponsibleRole: "DESIGNER"}}
	c, err = a.ReviewStructured(c.ID, ReviewCommand{ExpectedVersion: c.Version, ReviewerID: "reviewer", Decision: "RETURN", Items: items})
	if err != nil {
		t.Fatal(err)
	}
	next := revision(2.5, 4, "b")
	next.ReviewResponses = []domain.ReviewResponse{{ReturnItemID: c.ReturnItems[0].ID, Response: "已补拍", EvidenceDigest: domain.HashText("response-1")}}
	c, err = a.AddRevision(c.ID, c.Version, next, "r2")
	if err != nil {
		t.Fatal(err)
	}
	c, err = a.Check(c.ID, c.Version)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.ReviewStructured(c.ID, ReviewCommand{ExpectedVersion: c.Version, ReviewerID: "reviewer", Decision: "APPROVE", ConfirmItemIDs: []string{c.ReturnItems[0].ID}}); err == nil {
		t.Fatal("unanswered return item was approved")
	}
	last := revision(2.5, 4, "c")
	last.ReviewResponses = []domain.ReviewResponse{{ReturnItemID: c.ReturnItems[1].ID, Response: "已补充材质说明", EvidenceDigest: domain.HashText("response-2")}}
	c, err = a.AddRevision(c.ID, c.Version, last, "r3")
	if err != nil {
		t.Fatal(err)
	}
	c, err = a.Check(c.ID, c.Version)
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{c.ReturnItems[0].ID, c.ReturnItems[1].ID}
	c, err = a.ReviewStructured(c.ID, ReviewCommand{ExpectedVersion: c.Version, ReviewerID: "reviewer", Decision: "APPROVE", ConfirmItemIDs: ids})
	if err != nil {
		t.Fatal(err)
	}
	preview, err := a.PreviewFreeze(c.ID)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Digest == "" || preview.Counts["evidence"] != 6 {
		t.Fatalf("bad preview: %#v", preview)
	}
	c, err = a.FreezeConfirm(c.ID, c.Version, preview.Digest)
	if err != nil {
		t.Fatal(err)
	}
	cred, err := a.Issue(c.ID, c.Version, "publisher")
	if err != nil {
		t.Fatal(err)
	}
	result, err := a.VerifyCredential(c.ID, *cred)
	if err != nil || result.Status != "VALID" {
		t.Fatalf("valid verify: %#v %v", result, err)
	}
	tampered := *cred
	tampered.VerificationCode = "00"
	result, err = a.VerifyCredential(c.ID, tampered)
	if err != nil || result.Status != "TAMPERED" {
		t.Fatalf("tampered: %#v %v", result, err)
	}
	unknown := *cred
	unknown.CredentialID = "missing"
	result, err = a.VerifyCredential(c.ID, unknown)
	if err != nil || result.Status != "UNKNOWN" {
		t.Fatalf("unknown: %#v %v", result, err)
	}
	result, err = a.VerifyCredential("another-case", *cred)
	if err != nil || result.Status != "MISMATCHED" || len(result.MismatchedFields) == 0 {
		t.Fatalf("mismatched: %#v %v", result, err)
	}
	history, _ := a.Store.VerificationHistory(cred.CredentialID, 10)
	if len(history) != 3 {
		t.Fatalf("history=%d", len(history))
	}
}
