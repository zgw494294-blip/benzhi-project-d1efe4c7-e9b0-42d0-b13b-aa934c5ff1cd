package domain

import "time"

type AuditEntry struct {
	Seq                                              uint64
	CaseID, Action, Actor, BeforeDigest, AfterDigest string
	At                                               time.Time
	Metadata                                         map[string]string
}

func NewAudit(seq uint64, id, action, actor, before, after string) AuditEntry {
	return AuditEntry{Seq: seq, CaseID: id, Action: action, Actor: actor, BeforeDigest: before, AfterDigest: after, At: time.Now().UTC(), Metadata: map[string]string{}}
}
func (e AuditEntry) IsTransition() bool { return e.BeforeDigest != e.AfterDigest }
