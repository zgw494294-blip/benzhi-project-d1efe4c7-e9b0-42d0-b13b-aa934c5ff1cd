package store

import (
	"os"
	"tactile-review/internal/domain"
	"testing"
)

func TestStore(t *testing.T) {
	f := "test.db"
	defer os.Remove(f)
	s, e := Open(f)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	c := &domain.ReleaseCase{ID: "x"}
	if e = s.Save(c); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Get("x"); e != nil {
		t.Fatal(e)
	}
}
