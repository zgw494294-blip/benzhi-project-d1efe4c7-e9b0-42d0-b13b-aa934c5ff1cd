package domain

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrVersionConflict     = errors.New("version conflict")
	ErrInvalidState        = errors.New("invalid state")
	ErrFrozen              = errors.New("case is frozen")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrValidation          = errors.New("validation failed")
	ErrIdempotencyConflict = errors.New("idempotency key conflict")
	ErrEmptyRevision       = errors.New("empty replacement revision")
	ErrDeterminism         = errors.New("deterministic result changed")
	ErrDigestConflict      = errors.New("preview digest conflict")
)
