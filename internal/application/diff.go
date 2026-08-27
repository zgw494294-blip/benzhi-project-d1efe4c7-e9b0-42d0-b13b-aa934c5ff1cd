package application

import (
	"fmt"
	"tactile-review/internal/domain"
)

func (a *Service) RevisionDiff(id string, from, to int) ([]domain.FieldDiff, error) {
	c, e := a.Store.Get(id)
	if e != nil {
		return nil, e
	}
	var x, y *domain.PlateRevision
	for i := range c.Revisions {
		if c.Revisions[i].RevisionNo == from {
			x = &c.Revisions[i]
		}
		if c.Revisions[i].RevisionNo == to {
			y = &c.Revisions[i]
		}
	}
	if x == nil || y == nil {
		return nil, fmt.Errorf("revision not found")
	}
	return domain.Diff(*x, *y), nil
}
func (a *Service) Findings(id string) ([]domain.ComplianceFinding, error) {
	c, e := a.Store.Get(id)
	if e != nil {
		return nil, e
	}
	return domain.NormalizeFindings(c.Findings), nil
}
