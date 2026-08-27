package domain

import "fmt"

type Transition struct {
	From, To Status
	Actor    string
}

func (t Transition) Validate(c *ReleaseCase) error {
	if c == nil {
		return ErrNotFound
	}
	if c.Status != t.From {
		return ErrInvalidState
	}
	if t.Actor == "" {
		return ErrUnauthorized
	}
	if !Allowed(t.From, t.To) {
		return fmt.Errorf("transition %s to %s forbidden", t.From, t.To)
	}
	return nil
}
func (c *ReleaseCase) TransitionTo(to Status, actor string) error {
	t := Transition{c.Status, to, actor}
	if e := t.Validate(c); e != nil {
		return e
	}
	c.Status = to
	c.Version++
	return nil
}
func (c *ReleaseCase) HasOpenBlocking() bool {
	for _, f := range c.Findings {
		if f.Severity == string(Block) && f.Status == string(Open) {
			return true
		}
	}
	return false
}
func (c *ReleaseCase) CloseFinding(id string) bool {
	for i := range c.Findings {
		if c.Findings[i].ID == id && c.Findings[i].Status == string(Open) {
			c.Findings[i].Status = string(Closed)
			return true
		}
	}
	return false
}
func (c *ReleaseCase) FindingsForRevision(id string) []ComplianceFinding {
	out := []ComplianceFinding{}
	for _, f := range c.Findings {
		if f.RevisionID == id {
			out = append(out, f)
		}
	}
	return out
}
