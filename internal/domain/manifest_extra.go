package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func (m *Manifest) Canonical() string {
	if m == nil {
		return ""
	}
	parts := []string{"case_id=" + m.CaseID, "revision_id=" + m.RevisionID}
	keys := make([]string, 0, len(m.Fields))
	for k := range m.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		parts = append(parts, "field:"+k+"="+m.Fields[k])
	}
	for _, e := range NormalizeEvidence(m.Evidence) {
		parts = append(parts, "evidence:"+e.Kind+"="+e.Digest+":"+e.Description)
	}
	for _, f := range NormalizeFindings(m.Findings) {
		parts = append(parts, "finding:"+strings.Join([]string{f.ID, f.RuleID, f.RuleVersion, f.Severity, f.Status, f.Threshold, f.ActualValue, f.Explanation, f.InputDigest, f.SupersededByRevisionID}, "|"))
	}
	for _, r := range m.Reviews {
		parts = append(parts, "review:"+strings.Join([]string{r.ID, r.ReviewerID, r.RevisionID, r.Decision, r.Reason}, "|"))
		for _, id := range r.ConfirmedItemIDs {
			parts = append(parts, "confirmed:"+id)
		}
	}
	items := append([]ReviewReturnItem(nil), m.ReturnItems...)
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	for _, item := range items {
		parts = append(parts, "return:"+strings.Join([]string{item.ID, item.DecisionID, item.RevisionID, item.FindingID, item.Category, item.Description, item.ResponsibleRole, item.Status}, "|"))
		for _, response := range item.Responses {
			parts = append(parts, "response:"+strings.Join([]string{response.ReturnItemID, response.Response, response.EvidenceDigest}, "|"))
		}
	}
	countKeys := make([]string, 0, len(m.Counts))
	for k := range m.Counts {
		countKeys = append(countKeys, k)
	}
	sort.Strings(countKeys)
	for _, k := range countKeys {
		parts = append(parts, fmt.Sprintf("count:%s=%d", k, m.Counts[k]))
	}
	return strings.Join(parts, "\n")
}
func (m *Manifest) Recompute() { m.Digest = HashText(m.Canonical()) }

func BuildManifest(c *ReleaseCase, frozenAt time.Time) *Manifest {
	r := c.CurrentRevision()
	if c == nil || r == nil {
		return nil
	}
	fields := map[string]string{"case_id": c.ID, "revision_id": r.ID, "building_zone": c.BuildingZone, "installation_location": c.InstallationLocation, "audience_profile": c.AudienceProfile, "standard_version": c.StandardVersion, "braille_cells": r.BrailleCells, "dot_spacing_mm": fmt.Sprintf("%.3f", r.DotSpacingMM), "dot_height_mm": fmt.Sprintf("%.3f", r.DotHeightMM), "raised_text_height_mm": fmt.Sprintf("%.3f", r.RaisedTextHeightMM), "bevel_radius_mm": fmt.Sprintf("%.3f", r.BevelRadiusMM), "material_code": r.MaterialCode, "contrast_ratio": fmt.Sprintf("%.3f", r.ContrastRatio), "mounting_height_mm": fmt.Sprintf("%.3f", r.MountingHeightMM)}
	findings := NormalizeFindings(c.FindingsForRevision(r.ID))
	reviews := append([]ReviewDecision(nil), c.Reviews...)
	items := append([]ReviewReturnItem(nil), c.ReturnItems...)
	m := &Manifest{CaseID: c.ID, RevisionID: r.ID, Fields: fields, Evidence: append([]EvidenceItem(nil), r.Evidence...), Findings: findings, Reviews: reviews, ReturnItems: items, Counts: map[string]int{"fields": len(fields), "evidence": len(r.Evidence), "findings": len(findings), "reviews": len(reviews), "return_items": len(items)}, FrozenAt: frozenAt}
	if c.Review != nil {
		m.Review = *c.Review
	}
	m.Recompute()
	return m
}
