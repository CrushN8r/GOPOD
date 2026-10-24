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
// program GOPOD starts goes through here (2026-10-24: not only Python): GOPOD_HOME is the operator's home and
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
	// 2026-10-24 (operator: "RULE!! ANYTHING UNDER ROOT, FIX TO MY HOME"): its
	// home is the operator's too - "~", Path.home() and git's own settings in
	// whatever it starts never land in root's home. (The last HOME wins.)
	env = append(env, "HOME="+strings.TrimRight(home, "/"))
	// 2026-10-20 (live: "Edge TPU: the Edge TPU worker isn't ready" - the Python
	// 3.10 packages put in front of the coral env's own 3.9 ones broke its PIL):
	// a Python with its own environment (conda, venv) keeps its own packages;
	// only the system's python3 gets the operator's ~/.local ones.
	own := strings.Contains(cmd.Path, "/envs/") || strings.Contains(cmd.Path, "venv")
	if libs, _ := filepath.Glob(filepath.Join(home, ".local/lib/python3*/site-packages")); len(libs) > 0 && !own {
		if old := os.Getenv("PYTHONPATH"); old != "" {
			libs = append(libs, old)
		}
		env = append(env, "PYTHONPATH="+strings.Join(libs, ":"))
	}
	cmd.Env = env
	return cmd
}

// ## EOF
