package domain

import "testing"

func TestDigestStable(t *testing.T) {
	a := map[string]string{"b": "2", "a": "1"}
	if Digest(a, nil, nil) != Digest(map[string]string{"a": "1", "b": "2"}, nil, nil) {
		t.Fatal("unstable")
	}
}
