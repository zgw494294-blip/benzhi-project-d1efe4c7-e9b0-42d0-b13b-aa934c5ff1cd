package store

import (
	"encoding/json"
	"go.etcd.io/bbolt"
	"sync"
	"tactile-review/internal/domain"
)

var bucket = []byte("cases")

type Store struct {
	db *bbolt.DB
	mu sync.Mutex
}

func Open(path string) (*Store, error) {
	db, e := bbolt.Open(path, 0600, nil)
	if e != nil {
		return nil, e
	}
	if e = db.Update(func(tx *bbolt.Tx) error { _, e := tx.CreateBucketIfNotExists(bucket); return e }); e != nil {
		db.Close()
		return nil, e
	}
	s := &Store{db: db}
	if e = s.Migrate(); e != nil {
		db.Close()
		return nil, e
	}
	return s, nil
}
func (s *Store) Close() error { return s.db.Close() }
func (s *Store) Save(c *domain.ReleaseCase) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		v, _ := json.Marshal(c)
		return b.Put([]byte(c.ID), v)
	})
}
func (s *Store) Get(id string) (*domain.ReleaseCase, error) {
	var c domain.ReleaseCase
	e := s.db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket(bucket).Get([]byte(id))
		if v == nil {
			return domain.ErrNotFound
		}
		return json.Unmarshal(v, &c)
	})
	if e != nil {
		return nil, e
	}
	return &c, nil
}
func (s *Store) List() ([]*domain.ReleaseCase, error) {
	out := []*domain.ReleaseCase{}
	e := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucket).ForEach(func(_, v []byte) error {
			var c domain.ReleaseCase
			if err := json.Unmarshal(v, &c); err != nil {
				return err
			}
			out = append(out, &c)
			return nil
		})
	})
	return out, e
}
