package web

import (
	"tactile-review/internal/domain"
	"time"
)

type TimelineItem struct {
	Label string
	At    time.Time
	Done  bool
}
type ViewModel struct {
	CaseID, Status   string
	Version, Pending int
	Timeline         []TimelineItem
	Credential       *domain.FabricationCredential
}

func makeTimeline(c *domain.ReleaseCase) []TimelineItem {
	items := []TimelineItem{{"建档", c.CreatedAt, c.Version >= 1}, {"修订", c.UpdatedAt, len(c.Revisions) > 0}, {"校核", c.UpdatedAt, len(c.Findings) > 0}, {"复核", time.Time{}, c.Review != nil}, {"冻结", time.Time{}, c.Manifest != nil}, {"授权", time.Time{}, c.Credential != nil}}
	return items
}
func model(c *domain.ReleaseCase) ViewModel {
	pending := 0
	for _, f := range c.Findings {
		if f.Status == string(domain.Open) {
			pending++
		}
	}
	return ViewModel{c.ID, string(c.Status), c.Version, pending, makeTimeline(c), c.Credential}
}
