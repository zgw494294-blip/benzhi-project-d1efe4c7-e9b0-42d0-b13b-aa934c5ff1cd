package compliance

import "tactile-review/internal/domain"

func SeverityFor(rule string) domain.Severity {
	for _, r := range ruleSets["GB-2024"] {
		if r.ID == rule {
			return r.Severity
		}
	}
	return domain.Warn
}
func RuleIDs(version string) []string {
	r := ruleSets[version]
	out := make([]string, 0, len(r))
	for _, x := range r {
		out = append(out, x.ID)
	}
	return out
}
func IsBlocking(version, rule string) bool { return SeverityFor(rule) == domain.Block && version != "" }
