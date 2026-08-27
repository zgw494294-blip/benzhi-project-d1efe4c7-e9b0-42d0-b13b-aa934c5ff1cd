package domain

import (
	"fmt"
	"sort"
	"strings"
)

type FieldDiff struct {
	Field          string `json:"field"`
	Before         string `json:"before"`
	After          string `json:"after"`
	Classification string `json:"classification"`
}

func RevisionDiff(a, b PlateRevision, includeUnchanged bool) []FieldDiff {
	d := []FieldDiff{}
	add := func(f, x, y string) {
		kind := "UNCHANGED"
		if x != y {
			kind = "CHANGED"
			if x == "" {
				kind = "ADDED"
			}
		}
		if includeUnchanged || kind != "UNCHANGED" {
			d = append(d, FieldDiff{f, x, y, kind})
		}
	}
	add("braille_cells", NormalizeText(a.BrailleCells), NormalizeText(b.BrailleCells))
	add("dot_spacing_mm", fmt.Sprintf("%.3fmm", a.DotSpacingMM), fmt.Sprintf("%.3fmm", b.DotSpacingMM))
	add("dot_height_mm", fmt.Sprintf("%.3fmm", a.DotHeightMM), fmt.Sprintf("%.3fmm", b.DotHeightMM))
	add("raised_text_height_mm", fmt.Sprintf("%.3fmm", a.RaisedTextHeightMM), fmt.Sprintf("%.3fmm", b.RaisedTextHeightMM))
	add("bevel_radius_mm", fmt.Sprintf("%.3fmm", a.BevelRadiusMM), fmt.Sprintf("%.3fmm", b.BevelRadiusMM))
	add("material_code", NormalizeText(a.MaterialCode), NormalizeText(b.MaterialCode))
	add("contrast_ratio", fmt.Sprintf("%.3f", a.ContrastRatio), fmt.Sprintf("%.3f", b.ContrastRatio))
	add("mounting_height_mm", fmt.Sprintf("%.3fmm", a.MountingHeightMM), fmt.Sprintf("%.3fmm", b.MountingHeightMM))
	beforeEvidence, afterEvidence := map[string]EvidenceItem{}, map[string]EvidenceItem{}
	for _, e := range NormalizeEvidence(a.Evidence) {
		beforeEvidence[e.Kind] = e
	}
	for _, e := range NormalizeEvidence(b.Evidence) {
		afterEvidence[e.Kind] = e
	}
	for _, kind := range EvidenceKinds {
		x, y := beforeEvidence[kind], afterEvidence[kind]
		add("evidence."+kind, x.Digest+":"+x.Description, y.Digest+":"+y.Description)
	}
	return d
}

func Diff(a, b PlateRevision) []FieldDiff {
	return RevisionDiff(a, b, false)
}
func Changed(a, b PlateRevision) bool { return len(Diff(a, b)) > 0 }

func EvidenceCanonical(items []EvidenceItem) string {
	cp := append([]EvidenceItem(nil), items...)
	sort.Slice(cp, func(i, j int) bool { return cp[i].Kind < cp[j].Kind })
	parts := make([]string, 0, len(cp))
	for _, e := range cp {
		parts = append(parts, e.Kind+"="+strings.ToLower(e.Digest)+":"+NormalizeText(e.Description))
	}
	return strings.Join(parts, "|")
}

func EvidenceEqual(a, b []EvidenceItem) bool { return EvidenceCanonical(a) == EvidenceCanonical(b) }
