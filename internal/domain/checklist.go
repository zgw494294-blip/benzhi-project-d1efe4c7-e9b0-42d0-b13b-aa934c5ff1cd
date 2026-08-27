package domain

type ChecklistItem struct {
	Key, Label string
	Complete   bool
}

func (c *ReleaseCase) Checklist() []ChecklistItem {
	r := c.CurrentRevision()
	return []ChecklistItem{{"case", "发布案建档", c != nil && c.ID != ""}, {"revision", "版样修订", r != nil}, {"evidence", "实测证据", r != nil && len(r.EvidenceDigests) > 0}, {"compliance", "规则校核", len(c.Findings) > 0}, {"review", "独立复核", c.Review != nil}, {"manifest", "冻结清单", c.Manifest != nil}, {"credential", "制作授权", c.Credential != nil}}
}
func Completed(items []ChecklistItem) int {
	n := 0
	for _, i := range items {
		if i.Complete {
			n++
		}
	}
	return n
}
