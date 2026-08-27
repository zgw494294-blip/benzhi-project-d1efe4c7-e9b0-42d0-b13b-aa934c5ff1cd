package compliance

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"tactile-review/internal/domain"
)

func EvidenceDigest(parts ...string) string {
	h := sha256.Sum256([]byte(strings.Join(parts, "\x1f")))
	return hex.EncodeToString(h[:])
}
func EvidencePresent(values []string) bool {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return true
		}
	}
	return false
}

var evidenceRule = map[string]string{
	"dot_spacing": "dot-spacing", "dot_height": "dot-height", "raised_text": "raised-text",
	"bevel": "bevel", "contrast": "contrast", "mounting_height": "mounting-height",
}

func Coverage(r domain.PlateRevision, version string) []domain.CoverageItem {
	byKind := map[string]domain.EvidenceItem{}
	for _, e := range domain.NormalizeEvidence(r.Evidence) {
		byKind[e.Kind] = e
	}
	if len(byKind) == 0 && len(r.EvidenceDigests) > 0 && r.Measurement != nil && strings.TrimSpace(r.Measurement.EvidenceSummary) != "" {
		for _, kind := range domain.EvidenceKinds {
			byKind[kind] = domain.EvidenceItem{Kind: kind, Digest: r.EvidenceDigests[0], Description: r.Measurement.EvidenceSummary}
		}
	}
	out := make([]domain.CoverageItem, 0, len(domain.EvidenceKinds))
	for _, kind := range domain.EvidenceKinds {
		ruleID := evidenceRule[kind]
		status := "MISSING"
		if _, exists := Thresholds(version)[ruleID]; !exists {
			status = "NOT_APPLICABLE"
		}
		e := byKind[kind]
		if status != "NOT_APPLICABLE" && e.Digest != "" && e.Description != "" {
			status = "COVERED"
		}
		out = append(out, domain.CoverageItem{Kind: kind, RuleID: ruleID, Label: domain.EvidenceLabel(kind), Status: status, Digest: e.Digest, Description: e.Description})
	}
	return out
}

func CoveredForRule(coverage []domain.CoverageItem, ruleID string) bool {
	for _, c := range coverage {
		if c.RuleID == ruleID {
			return c.Status == "COVERED" || c.Status == "NOT_APPLICABLE"
		}
	}
	return false
}
