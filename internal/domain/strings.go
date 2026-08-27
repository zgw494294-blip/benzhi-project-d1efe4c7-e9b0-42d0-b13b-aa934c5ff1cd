package domain

import (
	"fmt"
	"strings"
)

func (r PlateRevision) Label() string {
	return fmt.Sprintf("修订 %d · %s", r.RevisionNo, r.MaterialCode)
}
func (f ComplianceFinding) Label() string {
	return fmt.Sprintf("%s %s (%s)", f.RuleID, f.Severity, f.Status)
}
func JoinEvidence(v []string) string { return strings.Join(v, ",") }
func SplitEvidence(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := []string{}
	for _, p := range parts {
		if x := strings.TrimSpace(p); x != "" {
			out = append(out, x)
		}
	}
	return out
}
