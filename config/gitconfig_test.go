package config

import (
	"testing"
)

func TestGetString(t *testing.T) {
	t.Run("returns default when nothing set", func(t *testing.T) {
		got := GetString("nonexistent-key-12345", "default-value")
		if got != "default-value" {
			t.Errorf("GetString() = %q, want %q", got, "default-value")
		}
	})
}

func TestGetInt(t *testing.T) {
	t.Run("returns default when nothing set", func(t *testing.T) {
		got := GetInt("nonexistent-key-12345", 42)
		if got != 42 {
			t.Errorf("GetInt() = %d, want %d", got, 42)
		}
	})
}

func TestGetFloat(t *testing.T) {
	t.Run("returns default when nothing set", func(t *testing.T) {
		got := GetFloat("nonexistent-key-12345", 0.5)
		if got != 0.5 {
			t.Errorf("GetFloat() = %f, want %f", got, 0.5)
		}
	})
}

func TestGetBool(t *testing.T) {
	t.Run("returns default when nothing set", func(t *testing.T) {
		got := GetBool("nonexistent-key-12345", true)
		if got != true {
			t.Errorf("GetBool() = %v, want %v", got, true)
		}
	})

	t.Run("returns false as default", func(t *testing.T) {
		got := GetBool("nonexistent-key-12345", false)
		if got != false {
			t.Errorf("GetBool() = %v, want %v", got, false)
		}
	})
}
