package compliance

import (
	"sort"
	"tactile-review/internal/domain"
)

type Report struct {
	RevisionID                 string
	Findings                   []domain.ComplianceFinding
	Blocking, Warnings, Passed int
}

func BuildReport(revision string, fs []domain.ComplianceFinding) Report {
	out := append([]domain.ComplianceFinding(nil), fs...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	r := Report{RevisionID: revision, Findings: out}
	for _, f := range out {
		switch f.Severity {
		case string(domain.Block):
			r.Blocking++
		case string(domain.Warn):
			r.Warnings++
		case string(domain.Pass):
			r.Passed++
		}
	}
	return r
}
