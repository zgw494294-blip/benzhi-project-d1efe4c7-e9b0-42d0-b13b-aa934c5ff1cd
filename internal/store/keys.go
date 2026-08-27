package store

import "strings"

func ValidKey(k string) bool                   { return len(k) >= 8 && !strings.ContainsAny(k, "/\\") }
func CaseKey(id string) []byte                 { return []byte("case:" + id) }
func IdempotencyKey(caseID, key string) []byte { return []byte("idem:" + caseID + ":" + key) }
