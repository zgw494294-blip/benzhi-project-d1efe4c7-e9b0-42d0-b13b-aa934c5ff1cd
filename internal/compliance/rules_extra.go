package compliance

import (
	"sort"
	"tactile-review/internal/domain"
)

type RuleResult struct {
	RuleID           string
	Severity         domain.Severity
	Value, Threshold float64
	Passed           bool
}

func EvaluateResults(c *domain.ReleaseCase, r domain.PlateRevision) []RuleResult {
	fs := Evaluate(c, r)
	out := []RuleResult{}
	for _, f := range fs {
		if f.RuleID == "evidence" {
			continue
		}
		var v, t float64
		for _, x := range Measure(r) {
			if x.Field == f.RuleID {
				v = x.Value
				t = x.Threshold
			}
		}
		out = append(out, RuleResult{f.RuleID, domain.Severity(f.Severity), v, t, f.Severity == string(domain.Pass)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].RuleID < out[j].RuleID })
	return out
}
func Blocking(results []RuleResult) int {
	n := 0
	for _, r := range results {
		if r.Severity == domain.Block && !r.Passed {
			n++
		}
	}
	return n
}
func Warnings(results []RuleResult) int {
	n := 0
	for _, r := range results {
		if r.Severity == domain.Warn && !r.Passed {
			n++
		}
	}
	return n
}
