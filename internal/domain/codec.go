package domain

import (
	"encoding/base64"
	"encoding/json"
)

func MarshalCase(c *ReleaseCase) (string, error) {
	b, e := json.Marshal(c)
	if e != nil {
		return "", e
	}
	return base64.RawStdEncoding.EncodeToString(b), nil
}
func UnmarshalCase(v string) (*ReleaseCase, error) {
	b, e := base64.RawStdEncoding.DecodeString(v)
	if e != nil {
		return nil, e
	}
	var c ReleaseCase
	e = json.Unmarshal(b, &c)
	return &c, e
}
