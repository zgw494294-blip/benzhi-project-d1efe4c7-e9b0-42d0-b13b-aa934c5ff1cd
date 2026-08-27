package domain

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

var EvidenceKinds = []string{"dot_spacing", "dot_height", "raised_text", "bevel", "contrast", "mounting_height"}

var evidenceLabels = map[string]string{
	"dot_spacing": "点距", "dot_height": "点高", "raised_text": "凸字",
	"bevel": "倒角", "contrast": "对比度", "mounting_height": "安装高度",
}

func EvidenceLabel(kind string) string { return evidenceLabels[kind] }

func NormalizeEvidence(items []EvidenceItem) []EvidenceItem {
	out := append([]EvidenceItem(nil), items...)
	for i := range out {
		out[i].Kind = strings.ToLower(NormalizeText(out[i].Kind))
		out[i].Digest = strings.ToLower(strings.TrimPrefix(NormalizeText(out[i].Digest), "sha256:"))
		out[i].Description = NormalizeText(out[i].Description)
	}
	sort.SliceStable(out, func(i, j int) bool { return evidenceOrder(out[i].Kind) < evidenceOrder(out[j].Kind) })
	return out
}

func evidenceOrder(kind string) int {
	for i, x := range EvidenceKinds {
		if kind == x {
			return i
		}
	}
	return len(EvidenceKinds)
}

func ValidateEvidence(items []EvidenceItem) []ValidationIssue {
	out := []ValidationIssue{}
	seenKind, seenDigest := map[string]bool{}, map[string]bool{}
	for i, e := range items {
		base := fmt.Sprintf("evidence[%d]", i)
		if EvidenceLabel(e.Kind) == "" {
			out = append(out, ValidationIssue{base + ".kind", "未知实测项目", true})
		}
		if seenKind[e.Kind] {
			out = append(out, ValidationIssue{base + ".kind", "同一实测项目只能登记一次", true})
		}
		seenKind[e.Kind] = true
		b, err := hex.DecodeString(e.Digest)
		if err != nil || len(b) != 32 {
			out = append(out, ValidationIssue{base + ".digest", "证据摘要必须是 SHA-256 十六进制值", true})
		} else if seenDigest[e.Digest] {
			out = append(out, ValidationIssue{base + ".digest", "同一修订内证据摘要不能重复", true})
		}
		seenDigest[e.Digest] = true
		if e.Description == "" {
			out = append(out, ValidationIssue{base + ".description", "证据说明不能为空", true})
		} else if len([]rune(e.Description)) > 500 {
			out = append(out, ValidationIssue{base + ".description", "证据说明不能超过 500 个字符", true})
		}
	}
	return out
}

func RevisionInputDigest(r PlateRevision) string {
	parts := []string{NormalizeText(r.BrailleCells), fmt.Sprintf("%.3f", r.DotSpacingMM), fmt.Sprintf("%.3f", r.DotHeightMM), fmt.Sprintf("%.3f", r.RaisedTextHeightMM), fmt.Sprintf("%.3f", r.BevelRadiusMM), NormalizeText(r.MaterialCode), fmt.Sprintf("%.3f", r.ContrastRatio), fmt.Sprintf("%.3f", r.MountingHeightMM), EvidenceCanonical(r.Evidence)}
	return HashText(strings.Join(parts, "\x1f"))
}

func ImpactedRules(changes []FieldDiff) []string {
	m := map[string]string{"dot_spacing_mm": "dot-spacing", "dot_height_mm": "dot-height", "raised_text_height_mm": "raised-text", "bevel_radius_mm": "bevel", "contrast_ratio": "contrast", "mounting_height_mm": "mounting-height"}
	set := map[string]bool{}
	for _, d := range changes {
		if d.Classification == "UNCHANGED" {
			continue
		}
		if strings.HasPrefix(d.Field, "evidence.") {
			evidenceRules := map[string]string{"dot_spacing": "dot-spacing", "dot_height": "dot-height", "raised_text": "raised-text", "bevel": "bevel", "contrast": "contrast", "mounting_height": "mounting-height"}
			if id := evidenceRules[strings.TrimPrefix(d.Field, "evidence.")]; id != "" {
				set[id] = true
			}
		} else if id := m[d.Field]; id != "" {
			set[id] = true
		}
	}
	out := make([]string, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
