package main

import "tactile-review/internal/domain"

func domainRevision() domain.PlateRevision {
	return domain.PlateRevision{RevisionNo: 1, BrailleCells: "A1", DotSpacingMM: 2.5, DotHeightMM: .7, RaisedTextHeightMM: 1, BevelRadiusMM: 1, MaterialCode: "PVC", ContrastRatio: 4, MountingHeightMM: 900, EvidenceDigests: []string{"digest"}, Measurement: &domain.Measurement{EvidenceSummary: "现场测量记录"}}
}
