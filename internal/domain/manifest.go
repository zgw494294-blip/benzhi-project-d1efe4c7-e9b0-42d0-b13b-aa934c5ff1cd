package domain

func (m *Manifest) VerifyDigest() bool {
	return m != nil && HashText(m.Canonical()) == m.Digest
}
