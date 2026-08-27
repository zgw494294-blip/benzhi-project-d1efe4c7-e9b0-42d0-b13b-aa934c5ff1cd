package domain

import "strings"

func CredentialPayload(c FabricationCredential) string {
	return strings.Join([]string{c.SchemaVersion, c.CredentialID, c.CaseID, c.RevisionID, c.ManifestDigest, c.IssuedBy, c.IssuedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00")}, "|")
}
func (c *FabricationCredential) Valid(secret string, manifest *Manifest) bool {
	return c != nil && manifest != nil && c.ManifestDigest == manifest.Digest && Verify(secret, CredentialPayload(*c), c.VerificationCode)
}
