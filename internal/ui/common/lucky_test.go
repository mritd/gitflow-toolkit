package common

import (
	"testing"
)

func TestGenerateFakeHash(t *testing.T) {
	for i := 0; i < 10; i++ {
		hash := generateFakeHash()

		// Check length is 40
		if len(hash) != 40 {
			t.Errorf("generateFakeHash() length = %d, want 40", len(hash))
		}

		// Check all chars are hex
		for _, c := range hash {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("generateFakeHash() contains non-hex char: %c", c)
			}
		}
	}

	// Check randomness (two calls should produce different results)
	hash1 := generateFakeHash()
	hash2 := generateFakeHash()
	if hash1 == hash2 {
		t.Errorf("generateFakeHash() produced same hash twice: %s", hash1)
	}
}
