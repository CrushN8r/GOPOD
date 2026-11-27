// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_page_controls.go
package wirepod_ttr

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

const gopodSkipMarker = "skip:"

func init() {
	http.HandleFunc("/api-sdk/gopod_loop_volume", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		v, err := strconv.Atoi(strings.TrimSpace(r.FormValue("v")))
		if err != nil || v < 0 || v > 100 {
			http.Error(w, "v must be 0-100", http.StatusBadRequest)
			return
		}
		raw := fmt.Sprintf("%q", fmt.Sprintf("%d%%", v))
		if _, err := gopodSettingsWrite([]string{"background_loop"}, func(file string) (string, error) {
			return gopodJSONFileSet(file, []string{"background_loop", "volume"}, raw)
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logger.Println(fmt.Sprintf("GOPOD_LOOP_VOLUME %d%% (kept in the settings)", v))
		w.Write([]byte("ok"))
	})
}

// gopodSkipLine: the line is only "skip" - told to the page as a skip, logged.
func gopodSkipLine(text, how string) bool {
	t := strings.ToLower(strings.Trim(strings.TrimSpace(text), ".!"))
	if t != "skip" && t != "system-skip" {
		return false
	}
	logger.Println("GOPOD_SKIP skipped (" + how + ")")
	GopodChatFeedAdd(gopodBrobotsSettings.ReporterVoices.Role, gopodSkipMarker+how)
	return true
}

// ## EOF
