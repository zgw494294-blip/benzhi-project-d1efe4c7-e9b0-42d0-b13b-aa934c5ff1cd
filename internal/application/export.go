package application

import (
	"encoding/json"
	"tactile-review/internal/domain"
)

func ExportManifest(c *domain.ReleaseCase) ([]byte, error) {
	if c.Manifest == nil {
		return nil, domain.ErrInvalidState
	}
	return json.MarshalIndent(c.Manifest, "", "  ")
}
func ExportCredential(c *domain.ReleaseCase) ([]byte, error) {
	if c.Credential == nil {
		return nil, domain.ErrInvalidState
	}
	return json.MarshalIndent(c.Credential, "", "  ")
}
