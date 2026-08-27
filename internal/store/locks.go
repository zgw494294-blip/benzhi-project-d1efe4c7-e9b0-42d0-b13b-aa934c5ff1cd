package store

import "sync"

type CaseLock struct {
	mu   sync.Mutex
	held map[string]bool
}

func NewCaseLock() *CaseLock { return &CaseLock{held: map[string]bool{}} }
func (l *CaseLock) Acquire(id string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.held[id] {
		return false
	}
	l.held[id] = true
	return true
}
func (l *CaseLock) Release(id string) { l.mu.Lock(); delete(l.held, id); l.mu.Unlock() }
