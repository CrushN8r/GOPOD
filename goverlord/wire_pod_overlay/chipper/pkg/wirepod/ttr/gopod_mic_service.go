// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_mic_service.go
package wirepod_ttr

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

var gopodMic = struct {
	sync.Mutex
	stdin   io.WriteCloser
	next    int
	waiting map[int]chan map[string]any
}{waiting: map[int]chan map[string]any{}}

var errGopodMicGone = errors.New("mic service ended")

// gopodMicCall sends one command and waits for its reply.
func gopodMicCall(cmd map[string]any, timeout time.Duration) (map[string]any, error) {
	gopodMic.Lock()
	if gopodMic.stdin == nil {
		if err := gopodMicStartLocked(); err != nil {
			gopodMic.Unlock()
			return nil, err
		}
	}
	gopodMic.next++
	id := gopodMic.next
	reply := make(chan map[string]any, 1)
	gopodMic.waiting[id] = reply
	cmd["id"] = id
	data, _ := json.Marshal(cmd)
	_, err := gopodMic.stdin.Write(append(data, '\n'))
	gopodMic.Unlock()
	if err != nil {
		gopodMicForget(id)
		return nil, err
	}
	select {
	case r, ok := <-reply:
		if !ok {
			return nil, errGopodMicGone
		}
		return r, nil
	case <-time.After(timeout):
		gopodMicForget(id)
		return nil, errors.New("mic service did not answer in time")
	}
}

func gopodMicForget(id int) {
	gopodMic.Lock()
	delete(gopodMic.waiting, id)
	gopodMic.Unlock()
}

func gopodMicStartLocked() error {
	script := filepath.Join(gopodChipperRoot(), gopodBrobotsSettings.MicService.Script)
	cmd := gopodAsOperator(exec.Command("python3", "-u", script))
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		return err
	}
	gopodMic.stdin = stdin
	logger.Println("GOPOD_MIC service started")
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "{") {
				var r map[string]any
				if json.Unmarshal([]byte(line), &r) == nil {
					id, _ := r["id"].(float64)
					gopodMic.Lock()
					ch := gopodMic.waiting[int(id)]
					delete(gopodMic.waiting, int(id))
					gopodMic.Unlock()
					if ch != nil {
						ch <- r
					}
					continue
				}
			}
			logger.Println(line)
		}
		err := cmd.Wait()
		gopodMic.Lock()
		gopodMic.stdin = nil
		for id, ch := range gopodMic.waiting {
			close(ch)
			delete(gopodMic.waiting, id)
		}
		gopodMic.Unlock()
		reason := "exit"
		if err != nil {
			reason = err.Error()
		}
		logger.Println("GOPOD_MIC service ended: " + reason)
	}()
	return nil
}

// gopodMicReplyError: a reply's error ("no_mic", ...), or "" when it went fine.
func gopodMicReplyError(r map[string]any) string {
	if ok, _ := r["ok"].(bool); ok {
		return ""
	}
	if e, _ := r["error"].(string); e != "" {
		return e
	}
	return "failed"
}

// ## EOF
