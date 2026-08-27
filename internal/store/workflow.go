package store

import (
	"encoding/json"
	"fmt"
	"go.etcd.io/bbolt"
	"strconv"
	"tactile-review/internal/domain"
	"time"
)

var idempotencyBucket = []byte("idempotency")
var verificationBucket = []byte("verifications")

type IdempotencyRecord struct {
	Key, RequestDigest, CaseID string
	CreatedAt                  time.Time
}

type VerificationRecord struct {
	ID               string    `json:"id"`
	TargetCaseID     string    `json:"target_case_id"`
	CredentialID     string    `json:"credential_id"`
	Status           string    `json:"status"`
	InputDigest      string    `json:"input_digest"`
	MismatchedFields []string  `json:"mismatched_fields"`
	At               time.Time `json:"at"`
}

func putAudit(tx *bbolt.Tx, entry domain.AuditEntry) error {
	b := tx.Bucket(auditBucket)
	seq, err := b.NextSequence()
	if err != nil {
		return err
	}
	entry.Seq = seq
	if entry.At.IsZero() {
		entry.At = time.Now().UTC()
	}
	v, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return b.Put([]byte(fmt.Sprintf("%020d", seq)), v)
}

func (s *Store) CreateIdempotent(c *domain.ReleaseCase, key, requestDigest string) (*domain.ReleaseCase, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result *domain.ReleaseCase
	reused := false
	err := s.db.Update(func(tx *bbolt.Tx) error {
		ib := tx.Bucket(idempotencyBucket)
		if raw := ib.Get([]byte("create:" + key)); raw != nil {
			var rec IdempotencyRecord
			if err := json.Unmarshal(raw, &rec); err != nil {
				return err
			}
			if rec.RequestDigest != requestDigest {
				return domain.ErrIdempotencyConflict
			}
			var existing domain.ReleaseCase
			rawCase := tx.Bucket(bucket).Get([]byte(rec.CaseID))
			if rawCase == nil {
				return domain.ErrNotFound
			}
			if err := json.Unmarshal(rawCase, &existing); err != nil {
				return err
			}
			result, reused = &existing, true
			return nil
		}
		encoded, err := json.Marshal(c)
		if err != nil {
			return err
		}
		if err = tx.Bucket(bucket).Put([]byte(c.ID), encoded); err != nil {
			return err
		}
		rec := IdempotencyRecord{Key: key, RequestDigest: requestDigest, CaseID: c.ID, CreatedAt: time.Now().UTC()}
		raw, err := json.Marshal(rec)
		if err != nil {
			return err
		}
		if err = ib.Put([]byte("create:"+key), raw); err != nil {
			return err
		}
		entry := domain.NewAudit(0, c.ID, "CASE_CREATED", c.DesignerID, "", requestDigest)
		if err = putAudit(tx, entry); err != nil {
			return err
		}
		copyCase := *c
		result = &copyCase
		return nil
	})
	return result, reused, err
}

func (s *Store) SaveWithAudit(c *domain.ReleaseCase, expectedVersion int, action, actor string, metadata map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Update(func(tx *bbolt.Tx) error {
		cb := tx.Bucket(bucket)
		raw := cb.Get([]byte(c.ID))
		if raw == nil {
			return domain.ErrNotFound
		}
		var before domain.ReleaseCase
		if err := json.Unmarshal(raw, &before); err != nil {
			return err
		}
		if expectedVersion >= 0 && before.Version != expectedVersion {
			return domain.ErrVersionConflict
		}
		encoded, err := json.Marshal(c)
		if err != nil {
			return err
		}
		if err = cb.Put([]byte(c.ID), encoded); err != nil {
			return err
		}
		entry := domain.NewAudit(0, c.ID, action, actor, domain.HashText(string(raw)), domain.HashText(string(encoded)))
		entry.Metadata = metadata
		return putAudit(tx, entry)
	})
}

func (s *Store) FindCredential(id string) (*domain.ReleaseCase, *domain.FabricationCredential, error) {
	var foundCase *domain.ReleaseCase
	var found *domain.FabricationCredential
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucket).ForEach(func(k, v []byte) error {
			if string(k) == "_last_write" {
				return nil
			}
			var c domain.ReleaseCase
			if err := json.Unmarshal(v, &c); err != nil {
				return err
			}
			if c.Credential != nil && c.Credential.CredentialID == id {
				cc := c
				cred := *c.Credential
				foundCase, found = &cc, &cred
			}
			return nil
		})
	})
	if err != nil {
		return nil, nil, err
	}
	if found == nil {
		return nil, nil, domain.ErrNotFound
	}
	return foundCase, found, nil
}

func (s *Store) AppendVerification(v VerificationRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(verificationBucket)
		seq, err := b.NextSequence()
		if err != nil {
			return err
		}
		v.ID = "verify-" + strconv.FormatUint(seq, 10)
		if v.At.IsZero() {
			v.At = time.Now().UTC()
		}
		raw, err := json.Marshal(v)
		if err != nil {
			return err
		}
		return b.Put([]byte(fmt.Sprintf("%020d", seq)), raw)
	})
}

func (s *Store) VerificationHistory(credentialID string, limit int) ([]VerificationRecord, error) {
	out := []VerificationRecord{}
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(verificationBucket)
		if b == nil {
			return nil
		}
		return b.ForEach(func(_, raw []byte) error {
			var v VerificationRecord
			if err := json.Unmarshal(raw, &v); err != nil {
				return err
			}
			if credentialID == "" || v.CredentialID == credentialID {
				out = append(out, v)
			}
			return nil
		})
	})
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out, err
}
