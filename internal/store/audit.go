package store

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type AuditRecord struct {
	Seq                    uint64
	At                     time.Time
	Action, CaseID, Digest string
}

func AuditDigest(action, id string) string {
	h := sha256.Sum256([]byte(action + id))
	return hex.EncodeToString(h[:])
}
