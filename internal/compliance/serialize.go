package compliance

import (
	"sort"
	"strings"
	"tactile-review/internal/domain"
)

func CanonicalFindings(fs []domain.ComplianceFinding) string {
	cp := append([]domain.ComplianceFinding(nil), fs...)
	sort.Slice(cp, func(i, j int) bool { return cp[i].ID < cp[j].ID })
	parts := []string{}
	for _, f := range cp {
		parts = append(parts, strings.Join([]string{f.ID, f.RuleID, f.RuleVersion, f.Severity, f.Status, f.Threshold, f.ActualValue, f.Explanation, f.InputDigest}, "|"))
	}
	return strings.Join(parts, "\n")
}
func Summary(fs []domain.ComplianceFinding) map[string]int {
	m := map[string]int{"PASS": 0, "WARN": 0, "BLOCK": 0, "OPEN": 0, "SUPERSEDED": 0, "CLOSED": 0}
	for _, f := range fs {
		m[f.Severity]++
		m[f.Status]++
	}
	return m
}
