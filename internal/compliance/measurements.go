package compliance

import (
	"fmt"
	"tactile-review/internal/domain"
)

type MeasurementResult struct {
	Field            string
	Value, Threshold float64
	Pass             bool
	Message          string
}

func Measure(r domain.PlateRevision) []MeasurementResult {
	return []MeasurementResult{{"dot_spacing_mm", r.DotSpacingMM, 2.3, r.DotSpacingMM >= 2.3, fmt.Sprintf("点距 %.3fmm", r.DotSpacingMM)}, {"dot_height_mm", r.DotHeightMM, .6, r.DotHeightMM >= .6, fmt.Sprintf("点高 %.3fmm", r.DotHeightMM)}, {"raised_text_height_mm", r.RaisedTextHeightMM, .8, r.RaisedTextHeightMM >= .8, fmt.Sprintf("凸字 %.3fmm", r.RaisedTextHeightMM)}, {"bevel_radius_mm", r.BevelRadiusMM, 1, r.BevelRadiusMM >= 1, fmt.Sprintf("倒角 %.3fmm", r.BevelRadiusMM)}, {"contrast_ratio", r.ContrastRatio, 3, r.ContrastRatio >= 3, fmt.Sprintf("对比度 %.3f", r.ContrastRatio)}, {"mounting_height_mm", r.MountingHeightMM, 800, r.MountingHeightMM >= 800, fmt.Sprintf("安装高度 %.3fmm", r.MountingHeightMM)}}
}
func Failing(r domain.PlateRevision) []MeasurementResult {
	all := Measure(r)
	out := []MeasurementResult{}
	for _, m := range all {
		if !m.Pass {
			out = append(out, m)
		}
	}
	return out
}
