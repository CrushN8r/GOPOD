// goverlord/wire_pod_overlay/chipper/pkg/vars/gopod_overlay_root.go
package vars

import (
	"os"
	"path/filepath"
	"strings"
)

// GOPOD 2026-10-02 (operator: "keep ~/wire-pod as native as possible ... reference
// from only overlay folder"): every GOPOD file - prompts, content, settings, the
// Brobots chat page files, the custom intents, run state - lives in the GOPOD repo's
// wire_pod_overlay/chipper folder, and Wire-Pod reads it from there. The build sets
// the folder (deploy_overlay_build.sh / build_with_overlay.sh, -ldflags -X); the
// GOPOD_OVERLAY_CHIPPER environment variable overrides it. Unset (a plain native
// build), nothing changes.
var GOPODOverlayChipperBuild string

// GOPODOverlayChipper is the overlay chipper folder, or "" when there is none.
func GOPODOverlayChipper() string {
	if dir := strings.TrimSpace(os.Getenv("GOPOD_OVERLAY_CHIPPER")); dir != "" {
		return dir
	}
	return strings.TrimSpace(GOPODOverlayChipperBuild)
}

// The custom intents (the GOPOD voice commands) are read from, and saved by the
// web config page to, the overlay copy instead of chipper/customIntents.json.
func init() {
	if dir := GOPODOverlayChipper(); dir != "" {
		CustomIntentsPath = filepath.Join(dir, "customIntents.json")
	}
}
