package compliance

type VersionInfo struct {
	ID             string `json:"id"`
	Label          string `json:"label"`
	RuleSetVersion string `json:"rule_set_version"`
	Enabled        bool   `json:"enabled"`
}

var versionCatalog = []VersionInfo{
	{ID: "GB-2024", Label: "GB-2024（现行）", RuleSetVersion: "rules-v1", Enabled: true},
	{ID: "GB-2023", Label: "GB-2023（已停用）", RuleSetVersion: "rules-v0", Enabled: false},
}

func VersionCatalog() []VersionInfo { return append([]VersionInfo(nil), versionCatalog...) }
func Versions() []string {
	out := []string{}
	for _, v := range versionCatalog {
		if v.Enabled {
			out = append(out, v.ID)
		}
	}
	return out
}
func Version(version string) (VersionInfo, bool) {
	for _, v := range versionCatalog {
		if v.ID == version {
			return v, true
		}
	}
	return VersionInfo{}, false
}
func Enabled(version string) bool { v, ok := Version(version); return ok && v.Enabled }
