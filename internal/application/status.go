package application

import "tactile-review/internal/domain"

type StatusCount map[domain.Status]int

func (a *Service) StatusCounts() (StatusCount, error) {
	cs, e := a.Store.List()
	if e != nil {
		return nil, e
	}
	m := StatusCount{}
	for _, c := range cs {
		m[c.Status]++
	}
	return m, nil
}
func (a *Service) CanEdit(id string) bool {
	c, e := a.Store.Get(id)
	return e == nil && c.CanWrite() == nil
}
func (a *Service) PolicyIssues(id string) []string {
	c, e := a.Store.Get(id)
	if e != nil {
		return []string{"not_found"}
	}
	return domain.DefaultPolicy.Check(c)
}
