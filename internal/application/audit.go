package application

import "time"

type AuditEvent struct {
	Action, CaseID string
	At             time.Time
}
