package application

import (
	"fmt"
	"tactile-review/internal/domain"
)

func EnsureExpected(actual, expected int) error {
	if actual != expected {
		return domain.ErrVersionConflict
	}
	return nil
}
func EnsureActor(actor string) error {
	if actor == "" {
		return fmt.Errorf("actor required")
	}
	return nil
}
func CanIssue(c *domain.ReleaseCase) error {
	if c.Status != domain.Frozen {
		return domain.ErrInvalidState
	}
	if c.Manifest == nil || !c.Manifest.VerifyDigest() {
		return fmt.Errorf("manifest integrity")
	}
	return nil
}
