package domain

import (
	"sort"
	"strings"
)

func NormalizeText(v string) string { return strings.Join(strings.Fields(strings.TrimSpace(v)), " ") }

func NormalizeCase(c *ReleaseCase) {
	c.BuildingZone = NormalizeText(c.BuildingZone)
	c.InstallationLocation = NormalizeText(c.InstallationLocation)
	c.AudienceProfile = NormalizeText(c.AudienceProfile)
	c.StandardVersion = strings.ToUpper(NormalizeText(c.StandardVersion))
	c.DesignerID = NormalizeActor(c.DesignerID)
	c.MeasurerID = NormalizeActor(c.MeasurerID)
	c.ReviewerID = NormalizeActor(c.ReviewerID)
}

func NormalizeFindings(fs []ComplianceFinding) []ComplianceFinding {
	out := append([]ComplianceFinding(nil), fs...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
