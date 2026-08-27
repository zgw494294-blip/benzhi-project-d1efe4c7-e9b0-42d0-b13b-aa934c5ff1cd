package store

import (
	"encoding/json"
	"go.etcd.io/bbolt"
	"strconv"
	"tactile-review/internal/domain"
)

var auditBucket = []byte("audit")

func (s *Store) appendAudit(e domain.AuditEntry) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b, er := tx.CreateBucketIfNotExists(auditBucket)
		if er != nil {
			return er
		}
		v, er := json.Marshal(e)
		if er != nil {
			return er
		}
		return b.Put([]byte(strconv.FormatUint(e.Seq, 10)), v)
	})
}
func (s *Store) Audit() ([]domain.AuditEntry, error) {
	out := []domain.AuditEntry{}
	e := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(auditBucket)
		if b == nil {
			return nil
		}
		return b.ForEach(func(_, v []byte) error {
			var x domain.AuditEntry
			if er := json.Unmarshal(v, &x); er != nil {
				return er
			}
			out = append(out, x)
			return nil
		})
	})
	return out, e
}
