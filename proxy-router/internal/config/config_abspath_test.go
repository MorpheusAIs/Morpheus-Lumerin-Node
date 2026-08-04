package config

import (
	"path/filepath"
	"testing"
)

func TestMustAbsPathLeavesAbsoluteUnchanged(t *testing.T) {
	abs := filepath.Join(t.TempDir(), ".cookie")
	got := mustAbsPath(abs)
	if got != filepath.Clean(abs) {
		t.Fatalf("got %q want %q", got, filepath.Clean(abs))
	}
}

func TestMustAbsPathResolvesRelative(t *testing.T) {
	got := mustAbsPath("./.cookie")
	if !filepath.IsAbs(got) {
		t.Fatalf("expected absolute path, got %q", got)
	}
	if filepath.Base(got) != ".cookie" {
		t.Fatalf("expected basename .cookie, got %q", got)
	}
}

func TestMustAbsPathEmpty(t *testing.T) {
	if got := mustAbsPath(""); got != "." && got != "" {
		// Clean("") == "." on most platforms
		if filepath.Clean("") != got {
			t.Fatalf("unexpected empty handling: %q", got)
		}
	}
}
