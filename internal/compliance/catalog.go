package compliance

import (
	"sort"
	"tactile-review/internal/domain"
)

type CatalogEntry struct {
	ID, Description string
	Threshold       float64
	Severity        domain.Severity
}

func Catalog(version string) []CatalogEntry {
	r := ruleSets[version]
	out := make([]CatalogEntry, 0, len(r))
	for _, x := range r {
		out = append(out, CatalogEntry{x.ID, x.Explain, x.Threshold, x.Severity})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func Thresholds(version string) map[string]float64 {
	m := map[string]float64{}
	for _, x := range ruleSets[version] {
		m[x.ID] = x.Threshold
	}
	return m
}
func Supports(version string) bool { _, ok := ruleSets[version]; return ok && Enabled(version) }
