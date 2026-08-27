package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func HashText(v string) string { h := sha256.Sum256([]byte(v)); return hex.EncodeToString(h[:]) }
func SafeText(v string, max int) string {
	v = strings.TrimSpace(v)
	if len(v) > max {
		return v[:max]
	}
	return v
}
func CredentialStatus(c *FabricationCredential, m *Manifest, secret string) string {
	if c == nil || m == nil {
		return "UNKNOWN"
	}
	if c.ManifestDigest != m.Digest {
		return "MISMATCH"
	}
	if !Verify(secret, CredentialPayload(*c), c.VerificationCode) {
		return "TAMPERED"
	}
	return "VALID"
}
