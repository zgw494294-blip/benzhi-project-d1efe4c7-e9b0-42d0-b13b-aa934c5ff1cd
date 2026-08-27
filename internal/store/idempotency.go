package store

import "sync"

type Idempotency struct {
	mu sync.Mutex
	m  map[string]any
}

func NewIdempotency() *Idempotency { return &Idempotency{m: map[string]any{}} }
func (i *Idempotency) Get(k string) (any, bool) {
	i.mu.Lock()
	defer i.mu.Unlock()
	v, ok := i.m[k]
	return v, ok
}
func (i *Idempotency) Put(k string, v any) { i.mu.Lock(); defer i.mu.Unlock(); i.m[k] = v }
