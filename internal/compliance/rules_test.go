package compliance

import (
	"tactile-review/internal/domain"
	"testing"
)

func TestEvaluate(t *testing.T) {
	c := &domain.ReleaseCase{ID: "c", StandardVersion: "GB-2024"}
	r := domain.PlateRevision{ID: "r", DotSpacingMM: 2.5, DotHeightMM: 0.7, ContrastRatio: 4, MountingHeightMM: 900, EvidenceDigests: []string{"e"}, Measurement: &domain.Measurement{EvidenceSummary: "ok"}}
	f := Evaluate(c, r)
	if len(f) != 6 {
		t.Fatalf("%d", len(f))
	}
}
