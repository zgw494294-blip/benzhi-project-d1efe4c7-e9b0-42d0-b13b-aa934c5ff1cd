package domain

import "strings"

type Policy struct {
	RequireEvidence, RequireReview, RequireContrast bool
	MinEvidence                                     int
}

var DefaultPolicy = Policy{true, true, true, 1}

func (p Policy) Check(c *ReleaseCase) []string {
	out := []string{}
	if c == nil {
		return []string{"case"}
	}
	if p.RequireEvidence {
		if r := c.CurrentRevision(); r == nil || len(r.EvidenceDigests) < p.MinEvidence {
			out = append(out, "evidence")
		}
	}
	if p.RequireReview && (c.Review == nil || c.Review.Decision != "APPROVE") {
		out = append(out, "review")
	}
	if p.RequireContrast {
		if r := c.CurrentRevision(); r != nil && r.ContrastRatio <= 0 {
			out = append(out, "contrast")
		}
	}
	return out
}
func NormalizeActor(v string) string { return strings.ToLower(strings.TrimSpace(v)) }
func IsIndependent(c *ReleaseCase, id string) bool {
	return c != nil && id != "" && id != c.DesignerID && id != c.MeasurerID
}
func IsTerminal(s Status) bool { return s == Authorized || s == Frozen }
