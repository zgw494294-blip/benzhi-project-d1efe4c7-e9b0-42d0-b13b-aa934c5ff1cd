package store

func (s *Store) Healthy() bool { return s != nil && s.db != nil }
