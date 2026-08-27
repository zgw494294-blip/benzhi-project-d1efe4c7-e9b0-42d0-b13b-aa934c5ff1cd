package store

import (
	"fmt"
	"go.etcd.io/bbolt"
)

func (s *Store) Check() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("database unavailable")
	}
	return s.db.View(func(tx *bbolt.Tx) error {
		if tx.Bucket(bucket) == nil {
			return fmt.Errorf("cases bucket missing")
		}
		return nil
	})
}
func (s *Store) Count() int {
	n := 0
	s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b != nil {
			b.ForEach(func(k, v []byte) error {
				if string(k) != "_last_write" {
					n++
				}
				return nil
			})
		}
		return nil
	})
	return n
}
