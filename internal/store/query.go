package store

import "tactile-review/internal/domain"

func FilterStatus(cs []*domain.ReleaseCase, status domain.Status) []*domain.ReleaseCase {
	out := []*domain.ReleaseCase{}
	for _, c := range cs {
		if c.Status == status {
			out = append(out, c)
		}
	}
	return out
}
