package addrevisionidempotency

import (
	"path/filepath"
	"testing"

	"tactile-review/internal/application"
	"tactile-review/internal/domain"
	"tactile-review/internal/store"
)

func evidence(seed string) []domain.EvidenceItem {
	out := make([]domain.EvidenceItem, 0, len(domain.EvidenceKinds))
	for _, kind := range domain.EvidenceKinds {
		out = append(out, domain.EvidenceItem{
			Kind: kind, Digest: domain.HashText(seed + kind), Description: "现场实测 " + kind,
		})
	}
	return out
}

func TestAddRevisionIdempotencyReplay(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "workflow.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	a := application.New(s)
	c, err := a.Create("A区", "东门", "视障人士", "GB-2024", "designer", "measurer")
	if err != nil {
		t.Fatal(err)
	}
	r := domain.PlateRevision{
		BrailleCells: "A1", DotSpacingMM: 2.5, DotHeightMM: .7,
		RaisedTextHeightMM: 1, BevelRadiusMM: 1, MaterialCode: "PVC",
		ContrastRatio: 4, MountingHeightMM: 900, Evidence: evidence("revision-"),
		SubmittedBy: "designer",
	}
	first, err := a.AddRevision(c.ID, c.Version, r, "revision-key")
	if err != nil {
		t.Fatal(err)
	}
	// A retry carries the original expected version and must replay the committed result.
	replayed, err := a.AddRevision(c.ID, c.Version, r, "revision-key")
	if err != nil {
		t.Fatalf("idempotent replay failed: %v", err)
	}
	if replayed.ID != first.ID || len(replayed.Revisions) != 1 || replayed.Version != first.Version {
		t.Fatalf("replay changed state: first=%#v replayed=%#v", first, replayed)
	}
}
