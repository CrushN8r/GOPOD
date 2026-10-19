package wirepod_ttr

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGopodChipperRootFollowsOverlayFolder pins GOPOD's file root to the overlay
// chipper folder when one is set (2026-10-02: Wire-Pod reads GOPOD's files only
// from the overlay), so a refactor can't silently send them back into ~/wire-pod.
func TestGopodChipperRootFollowsOverlayFolder(t *testing.T) {
	t.Setenv("GOPOD_OVERLAY_CHIPPER", "/tmp/gopod-overlay-chipper")
	if got := gopodChipperRoot(); got != "/tmp/gopod-overlay-chipper" {
		t.Fatalf("gopodChipperRoot() = %q, want the overlay folder", got)
	}
	t.Setenv(gopodChipperRootEnvVar, "/tmp/explicit-root")
	if got := gopodChipperRoot(); got != "/tmp/explicit-root" {
		t.Fatalf("gopodChipperRoot() = %q, want GOPOD_CHIPPER_ROOT to win", got)
	}
}

func TestGopodPathWithEnvOverrideDefaultsToChipperRootJoin(t *testing.T) {
	envVar := "GOPOD_TEST_PATH_OVERRIDE_UNUSED_001"
	os.Unsetenv(envVar)
	got := gopodPathWithEnvOverride(envVar, "some/relative/file.txt")
	want := filepath.Join(gopodChipperRoot(), "some/relative/file.txt")
	if got != want {
		t.Fatalf("gopodPathWithEnvOverride() = %q, want %q", got, want)
	}
}

func TestGopodPathWithEnvOverrideHonorsEnvVar(t *testing.T) {
	envVar := "GOPOD_TEST_PATH_OVERRIDE_001"
	t.Setenv(envVar, "/tmp/overridden-path.txt")
	got := gopodPathWithEnvOverride(envVar, "some/relative/file.txt")
	if got != "/tmp/overridden-path.txt" {
		t.Fatalf("gopodPathWithEnvOverride() = %q, want the env override value verbatim", got)
	}
}
