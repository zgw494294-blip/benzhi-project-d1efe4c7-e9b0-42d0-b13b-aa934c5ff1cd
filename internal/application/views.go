package application

import "tactile-review/internal/domain"

type CaseView struct {
	Case    *domain.ReleaseCase
	Pending int
}

func View(c *domain.ReleaseCase) CaseView {
	p := 0
	for _, f := range c.Findings {
		if f.Status == string(domain.Open) {
			p++
		}
	}
	return CaseView{c, p}
}
