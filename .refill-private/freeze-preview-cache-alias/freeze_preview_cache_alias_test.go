package freeze_preview_cache_alias_test

import (
	"path/filepath"
	"testing"

	"tactile-review/internal/application"
	"tactile-review/internal/domain"
	"tactile-review/internal/store"
)

func TestFreezePreviewCacheIsolation(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "preview.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	app := application.New(s)

	c, err := app.Create("A区", "东门", "视障人士", "GB-2024", "designer", "measurer")
	if err != nil {
		t.Fatal(err)
	}
	r := domain.PlateRevision{
		BrailleCells:       "A1",
		DotSpacingMM:       2.5,
		DotHeightMM:        0.7,
		RaisedTextHeightMM: 1,
		BevelRadiusMM:      1,
		MaterialCode:       "PVC",
		ContrastRatio:      4,
		MountingHeightMM:   900,
		EvidenceDigests:    []string{"private-evidence"},
		Measurement:        &domain.Measurement{EvidenceSummary: "现场实测"},
		SubmittedBy:        "designer",
	}
	c, err = app.AddRevision(c.ID, c.Version, r, "private-revision")
	if err != nil {
		t.Fatal(err)
	}
	c, err = app.Check(c.ID, c.Version)
	if err != nil {
		t.Fatal(err)
	}
	c, err = app.Review(c.ID, c.Version, "reviewer", "APPROVE", "证据充分")
	if err != nil {
		t.Fatal(err)
	}

	first, err := app.PreviewFreeze(c.ID)
	if err != nil {
		t.Fatal(err)
	}
	originalZone := first.Fields["building_zone"]
	originalEvidence := first.Evidence[0].Description
	originalCount := first.Counts["fields"]
	first.Fields["building_zone"] = "外部污染"
	first.Evidence[0].Description = "外部污染"
	first.Counts["fields"] = -1

	second, err := app.PreviewFreeze(c.ID)
	if err != nil {
		t.Fatal(err)
	}
	if second.Fields["building_zone"] != originalZone || second.Evidence[0].Description != originalEvidence || second.Counts["fields"] != originalCount {
		t.Fatalf("preview cache leaked caller mutation: zone=%q evidence=%q fields=%d", second.Fields["building_zone"], second.Evidence[0].Description, second.Counts["fields"])
	}
}
