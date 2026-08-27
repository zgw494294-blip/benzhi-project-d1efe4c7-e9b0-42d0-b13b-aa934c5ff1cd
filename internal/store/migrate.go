package store

import (
	"go.etcd.io/bbolt"
	"time"
)

func (s *Store) Migrate() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		for _, name := range [][]byte{[]byte("cases"), []byte("audit"), []byte("meta"), []byte("idempotency"), []byte("verifications")} {
			if _, e := tx.CreateBucketIfNotExists(name); e != nil {
				return e
			}
		}
		return tx.Bucket([]byte("meta")).Put([]byte("format"), []byte(FormatVersion))
	})
}
func (s *Store) LastWrite() time.Time {
	var t time.Time
	s.db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket(bucket).Get([]byte("_last_write"))
		if len(v) > 0 {
			t, _ = time.Parse(time.RFC3339Nano, string(v))
		}
		return nil
	})
	return t
}
