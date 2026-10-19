package wirepod_ttr

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

const gopodChipperRootEnvVar = "GOPOD_CHIPPER_ROOT"

// gopodChipperRoot is where GOPOD's own files live: the overlay chipper folder
// (vars.GOPODOverlayChipper, set at build - 2026-10-02, operator: "reference from
// only overlay folder"), overridable via GOPOD_CHIPPER_ROOT. Without either it
// falls back to the chipper/ directory relative to this Go source file's own
// location (runtime.Caller(0), the same technique animation_vocab.go used).
func gopodChipperRoot() string {
	if override := strings.TrimSpace(os.Getenv(gopodChipperRootEnvVar)); override != "" {
		return override
	}
	if dir := vars.GOPODOverlayChipper(); dir != "" {
		return dir
	}
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("gopod_paths: unable to resolve source file location via runtime.Caller")
	}
	// thisFile: .../chipper/pkg/wirepod/ttr/gopod_paths.go
	return filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(thisFile))))
}

// gopodPathWithEnvOverride joins gopodChipperRoot() with relPath, unless
// envVar is set in the environment, in which case that value wins outright
// (an absolute replacement path, not a further join).
func gopodPathWithEnvOverride(envVar, relPath string) string {
	if override := strings.TrimSpace(os.Getenv(envVar)); override != "" {
		return override
	}
	return filepath.Join(gopodChipperRoot(), relPath)
}
