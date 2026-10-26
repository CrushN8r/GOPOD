// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_own_files.go
package wirepod_ttr

// GOLDEN, one place (2026-10-26, operator: "Account for ANYTHING OWNED GENERATED
// BY/FOR ROOT, to migrate to OWNED by HOME"; "build the wire-pod helper"):
// Wire-Pod runs as root, but everything GOPOD writes in the operator's home is
// the operator's. Every overlay write goes through these (gopodWriteFile,
// gopodMkdirAll, gopodOwn), and every program Wire-Pod starts runs as the
// operator (gopodAsOperator, gopodRunAsOperator). The operator is the owner of
// GOPOD's chipper folder (gopodRecordRoot). Not root (a test, a dev run): nothing
// changes. ~/wire-pod (upstream) is left as it is.

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
)

var gopodOperatorIDs struct {
	once   sync.Once
	uid    int
	gid    int
	groups []uint32
}

// gopodOperatorUser: the operator's uid, gid and groups when Wire-Pod is root;
// uid -1 otherwise (nothing to hand over).
func gopodOperatorUser() (int, int, []uint32) {
	o := &gopodOperatorIDs
	o.once.Do(func() {
		o.uid, o.gid = -1, -1
		if os.Geteuid() != 0 {
			return
		}
		_, uid, gid := gopodRecordRoot()
		if uid <= 0 {
			return
		}
		o.uid, o.gid = uid, gid
		if u, err := user.LookupId(strconv.Itoa(uid)); err == nil {
			ids, _ := u.GroupIds()
			for _, id := range ids {
				if n, err := strconv.Atoi(id); err == nil {
					o.groups = append(o.groups, uint32(n))
				}
			}
		}
	})
	return o.uid, o.gid, o.groups
}

// gopodOwn hands paths to the operator (no-op when Wire-Pod isn't root).
func gopodOwn(paths ...string) {
	uid, gid, _ := gopodOperatorUser()
	if uid < 0 {
		return
	}
	for _, p := range paths {
		if p != "" {
			_ = os.Lchown(p, uid, gid)
		}
	}
}

// gopodWriteFile: os.WriteFile, the file the operator's.
func gopodWriteFile(path string, data []byte, perm os.FileMode) error {
	err := os.WriteFile(path, data, perm)
	if err == nil {
		gopodOwn(path)
	}
	return err
}

// gopodMkdirAll: os.MkdirAll, every folder it makes the operator's.
func gopodMkdirAll(path string, perm os.FileMode) error {
	var made []string
	for p := filepath.Clean(path); p != "/" && p != "."; p = filepath.Dir(p) {
		if _, err := os.Lstat(p); err == nil {
			break
		}
		made = append(made, p)
	}
	err := os.MkdirAll(path, perm)
	if err == nil {
		gopodOwn(made...)
	}
	return err
}

// gopodRunAsOperator: a program Wire-Pod starts runs as the operator, with his
// groups (audio, video, plugdev, input... for the mic, cameras and USB).
func gopodRunAsOperator(attr *syscall.SysProcAttr) *syscall.SysProcAttr {
	uid, gid, groups := gopodOperatorUser()
	if uid < 0 {
		return attr
	}
	if attr == nil {
		attr = &syscall.SysProcAttr{}
	}
	attr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid), Groups: groups}
	return attr
}

// ## EOF
