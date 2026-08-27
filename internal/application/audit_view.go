package application

import (
	"sort"
	"tactile-review/internal/domain"
	"tactile-review/internal/store"
)

func AuditSort(v []domain.AuditEntry) []domain.AuditEntry {
	out := append([]domain.AuditEntry(nil), v...)
	sort.Slice(out, func(i, j int) bool { return out[i].Seq < out[j].Seq })
	return out
}
func (a *Service) AuditTrail() ([]domain.AuditEntry, error) {
	v, e := a.Store.Audit()
	return AuditSort(v), e
}
func AuditAction(e domain.AuditEntry) string { return e.Action + "/" + e.Actor }

var _ = store.FormatVersion
