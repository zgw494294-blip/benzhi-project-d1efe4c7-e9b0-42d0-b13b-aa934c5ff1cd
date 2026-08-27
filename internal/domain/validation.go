package domain

import (
	"math"
	"strings"
)

type ValidationIssue struct {
	Field    string `json:"field"`
	Message  string `json:"message"`
	Blocking bool   `json:"blocking"`
}

type ValidationErrors struct{ Issues []ValidationIssue }

func (e *ValidationErrors) Error() string { return "请求字段校验失败" }

func IssuesError(issues []ValidationIssue) error {
	if len(issues) == 0 {
		return nil
	}
	return &ValidationErrors{Issues: issues}
}

func (r PlateRevision) Validate() []ValidationIssue {
	out := []ValidationIssue{}
	add := func(f, m string, b bool) { out = append(out, ValidationIssue{f, m, b}) }
	if strings.TrimSpace(r.BrailleCells) == "" {
		add("braille_cells", "盲文编码不能为空", true)
	}
	if r.DotSpacingMM < 1 || r.DotSpacingMM > 10 {
		add("dot_spacing_mm", "点距应在 1 至 10mm", true)
	}
	if r.DotHeightMM < 0.1 || r.DotHeightMM > 5 {
		add("dot_height_mm", "点高应在 0.1 至 5mm", true)
	}
	if r.RaisedTextHeightMM > 10 {
		add("raised_text_height_mm", "凸字高度超出范围", true)
	}
	if r.BevelRadiusMM > 20 {
		add("bevel_radius_mm", "倒角半径超出范围", true)
	}
	if r.ContrastRatio > 30 {
		add("contrast_ratio", "对比度超出范围", true)
	}
	if r.MountingHeightMM > 3000 {
		add("mounting_height_mm", "安装高度超出范围", true)
	}
	if strings.TrimSpace(r.MaterialCode) == "" {
		add("material_code", "材质代码不能为空", true)
	}
	out = append(out, ValidateEvidence(r.Evidence)...)
	return out
}
func (c *ReleaseCase) Validate() []ValidationIssue {
	out := []ValidationIssue{}
	add := func(f, m string, b bool) { out = append(out, ValidationIssue{f, m, b}) }
	if strings.TrimSpace(c.ID) == "" {
		add("id", "案号不能为空", true)
	}
	if strings.TrimSpace(c.BuildingZone) == "" {
		add("building_zone", "建筑区域不能为空", true)
	}
	if strings.TrimSpace(c.InstallationLocation) == "" {
		add("installation_location", "安装位置不能为空", true)
	}
	if strings.TrimSpace(c.AudienceProfile) == "" {
		add("audience_profile", "目标人群不能为空", true)
	}
	for field, value := range map[string]string{"building_zone": c.BuildingZone, "installation_location": c.InstallationLocation, "audience_profile": c.AudienceProfile} {
		if len([]rune(value)) > 120 {
			add(field, "长度不能超过 120 个字符", true)
		}
	}
	for field, value := range map[string]string{"designer_id": c.DesignerID, "measurer_id": c.MeasurerID} {
		if strings.TrimSpace(value) == "" {
			add(field, "人员标识不能为空", true)
		} else if len([]rune(value)) > 64 {
			add(field, "长度不能超过 64 个字符", true)
		}
	}
	if strings.TrimSpace(c.StandardVersion) == "" {
		add("standard_version", "标准版本不能为空", true)
	}
	if c.DesignerID != "" && c.DesignerID == c.MeasurerID {
		add("measurer_id", "设计人员与检测人员不能相同", true)
	}
	if c.Version < 1 {
		add("version", "版本必须为正数", true)
	}
	return out
}
func NearlyEqual(a, b, eps float64) bool { return math.Abs(a-b) <= eps }
