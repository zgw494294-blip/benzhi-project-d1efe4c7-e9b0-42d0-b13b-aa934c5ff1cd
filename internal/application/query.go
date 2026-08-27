package application

import (
	"sort"
	"strings"
	"tactile-review/internal/domain"
)

type Query struct {
	Status         domain.Status
	Standard, Text string
}

func (a *Service) Search(q Query) ([]*domain.ReleaseCase, error) {
	cs, e := a.Store.List()
	if e != nil {
		return nil, e
	}
	out := []*domain.ReleaseCase{}
	for _, c := range cs {
		if q.Status != "" && c.Status != q.Status {
			continue
		}
		if q.Standard != "" && c.StandardVersion != q.Standard {
			continue
		}
		if q.Text != "" && !strings.Contains(c.BuildingZone, q.Text) && !strings.Contains(c.InstallationLocation, q.Text) {
			continue
		}
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}
func (a *Service) Snapshot(id string) (CaseView, error) {
	c, e := a.Store.Get(id)
	if e != nil {
		return CaseView{}, e
	}
	return View(c), nil
}
