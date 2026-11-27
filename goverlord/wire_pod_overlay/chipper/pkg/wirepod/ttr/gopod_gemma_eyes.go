// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_gemma_eyes.go
package wirepod_ttr

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

const gopodGemmaEyesHost = "GEMMA"

// gopodGemmaRoom: what GEMMA saw last, and when.
var gopodGemmaRoom = struct {
	sync.Mutex
	line string
	at   time.Time
}{}

// gopodGemmaEyesCapture: a photo from the C920 (gemma_eyes.device, its first frames skipped while
// it settles), else Brobot 1's camera, then Brobot 2's; which one took it.
func gopodGemmaEyesCapture() ([]byte, string, error) {
	ge := gopodBrobotsSettings.GemmaEyes
	if dev := gopodGemmaEyesDevice(); dev != "" {
		out := filepath.Join(os.TempDir(), "gopod_gemma_eyes.jpg")
		skip := ge.SkipFrames
		if skip <= 0 {
			skip = 15
		}
		args := []string{"-loglevel", "error", "-f", "v4l2"}
		if ge.Size != "" {
			args = append(args, "-video_size", ge.Size)
		}
		args = append(args, "-i", dev, "-vf", "select=gte(n\\,"+strconv.Itoa(skip)+")", "-frames:v", "1", "-y", out)
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		cmd := gopodAsOperator(exec.CommandContext(ctx, "ffmpeg", args...))
		if err := cmd.Run(); err == nil {
			if data, err := os.ReadFile(out); err == nil && len(data) > 0 {
				os.Remove(out)
				return data, "C920", nil
			}
		} else {
			logger.Println("GOPOD_GEMMA_EYES C920: " + err.Error())
		}
	}
	for _, esn := range []string{gopodBrobot1ESN, gopodBrobot2ESN} {
		if esn == "" {
			continue
		}
		var data []byte
		var capErr error
		gopodRobotCall(esn, true, func(robot *vector.Vector, ctx context.Context) string {
			data, capErr = gopodPhotoCapture(robot, ctx)
			return ""
		})
		if capErr == nil && len(data) > 0 {
			return data, gopodBrobotName(esn), nil
		}
	}
	return nil, "", errors.New("no camera took a photo (C920, Brobot 1, Brobot 2)")
}

// gopodGemmaEyesDevice: the C920's video device (gemma_eyes.device, a glob).
func gopodGemmaEyesDevice() string {
	for _, pattern := range gopodBrobotsSettings.GemmaEyes.Device {
		if m, _ := filepath.Glob(strings.TrimSpace(pattern)); len(m) > 0 {
			return m[0]
		}
	}
	return ""
}

// gopodGemmaEyesLook: GEMMA looks - the photo, the look, the picture in the chat; what she says.
func gopodGemmaEyesLook() string {
	ge := gopodBrobotsSettings.GemmaEyes
	data, source, err := gopodGemmaEyesCapture()
	if err != nil {
		logger.Println("GOPOD_GEMMA_EYES " + err.Error())
		return strings.TrimSpace(ge.FailText)
	}
	name := time.Now().Format("2006-01-02_150405") + "_gemma_" + strings.ToLower(strings.ReplaceAll(source, " ", "")) + ".jpg"
	page := ""
	if !ge.Safeguard {
		if page, err = gopodPhotoSave(name, data); err != nil {
			logger.Println("GOPOD_GEMMA_EYES save: " + err.Error())
			page = ""
		}
	}
	if page != "" {
		GopodChatFeedAdd(gopodBrobotsSettings.TopicPictures.Role, page)
	}
	saw, err := gopodPhotoDescribeWith(data, ge.Prompt)
	if err != nil || strings.TrimSpace(saw) == "" {
		logger.Println("GOPOD_GEMMA_EYES look: " + gopodErrText(err))
		return strings.TrimSpace(ge.FailText)
	}
	line := gopodPhotoFirstSentence(strings.TrimSpace(saw))
	gopodGemmaRoom.Lock()
	gopodGemmaRoom.line, gopodGemmaRoom.at = line, time.Now()
	gopodGemmaRoom.Unlock()
	logger.Println("GOPOD_GEMMA_EYES " + source + " (" + page + "): " + saw)
	return strings.TrimSpace(strings.NewReplacer("{saw}", gopodPhotoLowerFirst(line), "{camera}", source).Replace(ge.SayText))
}

func gopodErrText(err error) string {
	if err == nil {
		return "nothing said"
	}
	return err.Error()
}

// gopodGemmaRoomLine: the brain's ROOM - what GEMMA saw, while it's fresh
// (gemma_eyes.keep_minutes); "" otherwise.
func gopodGemmaRoomLine() string {
	gopodGemmaRoom.Lock()
	defer gopodGemmaRoom.Unlock()
	keep := gopodBrobotsSettings.GemmaEyes.KeepMinutes
	if gopodGemmaRoom.line == "" || keep <= 0 || time.Since(gopodGemmaRoom.at) > time.Duration(keep*float64(time.Minute)) {
		return ""
	}
	return gopodGemmaRoom.line
}

// gopodGemmaEyesAsked: a line to GEMMA asking her to look (the photo's ask phrases).
func gopodGemmaEyesAsked(host, text string) bool {
	return host == gopodGemmaEyesHost && gopodPhotoKind(text) == "ask"
}

// ## EOF
