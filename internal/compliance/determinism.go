package compliance

import "tactile-review/internal/domain"

func Stable(c *domain.ReleaseCase, r domain.PlateRevision) []domain.ComplianceFinding {
	return Evaluate(c, r)
}
