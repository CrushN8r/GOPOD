// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_as_operator.go
package wirepod_ttr

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// gopodAsOperator: Wire-Pod runs as root, but GOPOD's Python libraries (qrcode, vosk) and models
// live in the operator's home.
func gopodAsOperator(cmd *exec.Cmd) *exec.Cmd {
	home := gopodHomePath("~/")
	env := cmd.Env
	if env == nil {
		env = os.Environ()
	}
	env = append(env, "GOPOD_HOME="+home)
	env = append(env, "HOME="+strings.TrimRight(home, "/"))
	own := strings.Contains(cmd.Path, "/envs/") || strings.Contains(cmd.Path, "venv")
	if libs, _ := filepath.Glob(filepath.Join(home, ".local/lib/python3*/site-packages")); len(libs) > 0 && !own {
		if old := os.Getenv("PYTHONPATH"); old != "" {
			libs = append(libs, old)
		}
		env = append(env, "PYTHONPATH="+strings.Join(libs, ":"))
	}
	if uid, _, _ := gopodOperatorUser(); uid >= 0 {
		env = append(env, "XDG_RUNTIME_DIR=/run/user/"+strconv.Itoa(uid))
		cmd.SysProcAttr = gopodRunAsOperator(cmd.SysProcAttr)
	}
	cmd.Env = env
	return cmd
}

// ## EOF
