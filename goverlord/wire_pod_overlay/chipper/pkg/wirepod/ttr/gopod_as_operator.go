// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_as_operator.go
package wirepod_ttr

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// gopodAsOperator (2026-10-18, operator: "re: Wire-Pod runs as root * harden in
// code, so as not to have to remember"): Wire-Pod runs as root, but GOPOD's
// Python libraries (qrcode, vosk) and models live in the operator's home. Every
// Python GOPOD starts goes through here: GOPOD_HOME is the operator's home and
// PYTHONPATH adds the operator's own ~/.local packages. TestPythonRunsAsOperator
// fails on a Python start that skips it. (Live 2026-10-18: the QR codes failed
// and 101 of 103 voices lost their Vosk timing - both this.)
func gopodAsOperator(cmd *exec.Cmd) *exec.Cmd {
	home := gopodHomePath("~/")
	env := cmd.Env
	if env == nil {
		env = os.Environ()
	}
	env = append(env, "GOPOD_HOME="+home)
	if libs, _ := filepath.Glob(filepath.Join(home, ".local/lib/python3*/site-packages")); len(libs) > 0 {
		if old := os.Getenv("PYTHONPATH"); old != "" {
			libs = append(libs, old)
		}
		env = append(env, "PYTHONPATH="+strings.Join(libs, ":"))
	}
	cmd.Env = env
	return cmd
}

// ## EOF
