// goverlord/wire_pod_overlay/chipper/pkg/wirepod/config-ws/gopod_webroot.go
package webserver

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// gopodWebRoot serves Wire-Pod's web page with GOPOD's files on top (2026-10-02,
// operator: "reference from only overlay folder"): a file that exists in the
// overlay's webroot (GOPOD's index.html, the Brobots chat page files, KaTeX, the
// settings file and the run state the page polls) is served from there; anything
// else falls through to native ./webroot unchanged. With no overlay folder it is
// exactly the native file server.
func gopodWebRoot(native string) http.Handler {
	nativeFS := http.FileServer(http.Dir(native))
	dir := vars.GOPODOverlayChipper()
	if dir == "" {
		return nativeFS
	}
	overlay := filepath.Join(dir, "webroot")
	overlayFS := http.FileServer(http.Dir(overlay))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := path.Clean("/" + r.URL.Path)
		if name == "/" {
			name = "/index.html"
		}
		if info, err := os.Stat(filepath.Join(overlay, filepath.FromSlash(name))); err == nil && !info.IsDir() {
			overlayFS.ServeHTTP(w, r)
			return
		}
		// 2026-10-23: a folder's own page ("/sdkapp/" -> sdkapp/index.html),
		// when the overlay has one - GOPOD's copy of Bot Settings.
		if strings.HasSuffix(r.URL.Path, "/") {
			if info, err := os.Stat(filepath.Join(overlay, filepath.FromSlash(name), "index.html")); err == nil && !info.IsDir() {
				overlayFS.ServeHTTP(w, r)
				return
			}
		}
		nativeFS.ServeHTTP(w, r)
	})
}
