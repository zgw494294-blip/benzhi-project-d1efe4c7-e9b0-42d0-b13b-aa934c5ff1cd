package application

import "tactile-review/internal/domain"

func (a *Service) Checklist(id string) ([]domain.ChecklistItem, error) {
	c, e := a.Store.Get(id)
	if e != nil {
		return nil, e
	}
	return c.Checklist(), nil
}
