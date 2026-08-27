package application

import (
	"os"
	"tactile-review/internal/domain"
	"tactile-review/internal/store"
	"testing"
)

func TestFlow(t *testing.T) {
	f := "a.db"
	defer os.Remove(f)
	s, _ := store.Open(f)
	defer s.Close()
	a := New(s)
	c, err := a.Create("A", "L", "all", "GB-2024", "d", "m")
	if err != nil {
		t.Fatal(err)
	}
	r := domain.PlateRevision{RevisionNo: 1, BrailleCells: "abc", DotSpacingMM: 2.5, DotHeightMM: .7, RaisedTextHeightMM: 1, BevelRadiusMM: 1, MaterialCode: "PVC", ContrastRatio: 4, MountingHeightMM: 900, EvidenceDigests: []string{"e"}, Measurement: &domain.Measurement{EvidenceSummary: "ok"}}
	c, err = a.AddRevision(c.ID, c.Version, r, "k")
	if err != nil {
		t.Fatal(err)
	}
	c, err = a.Check(c.ID, c.Version)
	if err != nil {
		t.Fatal(err)
	}
	c, err = a.Review(c.ID, c.Version, "r", "APPROVE", "")
	if err != nil {
		t.Fatal(err)
	}
	c, err = a.Freeze(c.ID, c.Version)
	if err != nil {
		t.Fatal(err)
	}
	_, e := a.Issue(c.ID, c.Version, "pub")
	if e != nil {
		t.Fatal(e)
	}
}
