package compliance

import (
	"sort"
	"strconv"
	"tactile-review/internal/domain"
	"time"
)

type Rule struct {
	ID        string
	Threshold float64
	Op        string
	Severity  domain.Severity
	Explain   string
}

var ruleSets = map[string][]Rule{"GB-2024": {{"dot-spacing", 2.3, ">=", domain.Block, "盲文点距应达到 2.3mm"}, {"dot-height", 0.6, ">=", domain.Block, "盲文点高应达到 0.6mm"}, {"raised-text", 0.8, ">=", domain.Warn, "凸字高度建议达到 0.8mm"}, {"bevel", 1.0, ">=", domain.Warn, "边缘倒角半径建议达到 1.0mm"}, {"contrast", 3.0, ">=", domain.Block, "颜色对比度应达到 3.0"}, {"mounting-height", 800, ">=", domain.Warn, "安装高度应不低于 800mm"}}}

func Evaluate(c *domain.ReleaseCase, r domain.PlateRevision) []domain.ComplianceFinding {
	rules := append([]Rule(nil), ruleSets[c.StandardVersion]...)
	sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })
	now := time.Now().UTC()
	out := make([]domain.ComplianceFinding, 0, len(rules)*2)
	inputDigest := domain.RevisionInputDigest(r)
	vals := map[string]float64{"dot-spacing": r.DotSpacingMM, "dot-height": r.DotHeightMM, "raised-text": r.RaisedTextHeightMM, "bevel": r.BevelRadiusMM, "contrast": r.ContrastRatio, "mounting-height": r.MountingHeightMM}
	for _, x := range rules {
		v := vals[x.ID]
		sev := x.Severity
		if v >= x.Threshold && x.ID != "mounting-height" || x.ID == "mounting-height" && v >= x.Threshold {
			sev = domain.Pass
		}
		id := x.ID + "-" + r.ID
		out = append(out, domain.ComplianceFinding{ID: id, CaseID: c.ID, RevisionID: r.ID, RuleID: x.ID, RuleVersion: "rules-v1", Severity: string(sev), Status: string(domain.Open), Threshold: strconv.FormatFloat(x.Threshold, 'f', 3, 64), ActualValue: strconv.FormatFloat(v, 'f', 3, 64), Explanation: x.Explain, InputDigest: inputDigest, EvaluatedAt: now})
	}
	for _, coverage := range Coverage(r, c.StandardVersion) {
		if coverage.Status == "MISSING" {
			out = append(out, domain.ComplianceFinding{ID: "evidence-" + coverage.Kind + "-" + r.ID, CaseID: c.ID, RevisionID: r.ID, RuleID: "evidence-" + coverage.RuleID, RuleVersion: "rules-v1", Severity: string(domain.Block), Status: string(domain.Open), Threshold: "required", ActualValue: "missing", Explanation: coverage.Label + "缺少实测证据摘要与说明", InputDigest: inputDigest, EvaluatedAt: now})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].RuleID < out[j].RuleID })
	return out
}
