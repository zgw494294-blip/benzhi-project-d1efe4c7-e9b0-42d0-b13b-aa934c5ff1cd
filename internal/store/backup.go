package store

import (
	"go.etcd.io/bbolt"
	"os"
	"path/filepath"
	"time"
)

func (s *Store) Backup(dir string) (string, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	name := filepath.Join(dir, "tactile-"+time.Now().UTC().Format("20060102T150405.000000000")+".db")
	return name, s.db.View(func(tx *bbolt.Tx) error { return tx.CopyFile(name, 0600) })
}
