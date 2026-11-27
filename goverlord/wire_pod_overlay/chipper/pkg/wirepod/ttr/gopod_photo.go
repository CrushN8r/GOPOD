// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_photo.go
package wirepod_ttr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

var gopodPhotoDir = gopodPathWithEnvOverride("GOPOD_PHOTO_DIR", "webroot/photos")

// gopodPhotoLast: what each Brobot saw last (his ESN -> the description).
var gopodPhotoLast = struct {
	sync.Mutex
	saw map[string]string
}{saw: map[string]string{}}

// gopodPhotoKind: "ask" (take a photo), "recall" (tell what he saw) or "" for this line - in the
// live chat only.
func gopodPhotoKind(text string) string {
	p := gopodBrobotsSettings.Photo
	if gopodScriptPlaying.Load() || gopodRunActive() {
		return ""
	}
	lower := strings.ToLower(text)
	for _, phrase := range p.RecallPhrases {
		if phrase = strings.ToLower(strings.TrimSpace(phrase)); phrase != "" && strings.Contains(lower, phrase) {
			return "recall"
		}
	}
	for _, phrase := range p.AskPhrases {
		if phrase = strings.ToLower(strings.TrimSpace(phrase)); phrase != "" && strings.Contains(lower, phrase) {
			return "ask"
		}
	}
	return ""
}

// gopodPhotoCapture: GOLDEN, kept as Wire-Pod's own getImage takes it (DoGetImage, kgsim_cmds.go) -
// mirror mode on, "3", "2", "1", the photo (high resolution), mirror mode off, the shutter
// animation.
func gopodPhotoCapture(robot *vector.Vector, ctx context.Context) ([]byte, error) {
	time.Sleep(gopodSecondsToDuration(gopodBrobotsSettings.Photo.AwakeSeconds))
	if cam := gopodBrobotsSettings.Photo.Camera; cam.Manual {
		if err := gopodSetCameraSettings(ctx, robot.Conn, cam.Gain, cam.ExposureMs, false); err != nil {
			logger.Println("GOPOD_PHOTO camera settings: " + err.Error())
		}
		time.Sleep(gopodSecondsToDuration(cam.SettleSeconds))
		defer func() {
			if err := gopodSetCameraSettings(context.Background(), robot.Conn, 0, 0, true); err != nil {
				logger.Println("GOPOD_PHOTO camera back to auto: " + err.Error())
			}
		}()
	}
	robot.Conn.EnableMirrorMode(ctx, &vectorpb.EnableMirrorModeRequest{Enable: true})
	for i := 3; i > 0; i-- {
		time.Sleep(time.Millisecond * 300)
		gopodLogSayText(robot, ctx, &vectorpb.SayTextRequest{
			Text:           fmt.Sprint(i),
			UseVectorVoice: true,
			DurationScalar: 1.05,
		})
	}
	resp, err := robot.Conn.CaptureSingleImage(ctx, &vectorpb.CaptureSingleImageRequest{
		EnableHighResolution: true,
	})
	robot.Conn.EnableMirrorMode(ctx, &vectorpb.EnableMirrorModeRequest{Enable: false})
	go func() {
		gopodLogPlayAnimation(robot, context.Background(), &vectorpb.PlayAnimationRequest{
			Animation: &vectorpb.Animation{Name: "anim_photo_shutter_01"},
			Loops:     1,
		})
	}()
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Data) == 0 {
		return nil, errors.New("empty photo")
	}
	return resp.Data, nil
}

// gopodPhotoSave keeps the photo for the chat page (webroot/photos, private, not in the repo) and
// returns its page path.
func gopodPhotoSave(name string, data []byte) (string, error) {
	if err := gopodMkdirAll(gopodPhotoDir, 0o755); err != nil {
		return "", err
	}
	if err := gopodWriteFile(filepath.Join(gopodPhotoDir, name), data, 0o644); err != nil {
		return "", err
	}
	return gopodPhotoPage(name), nil
}

// gopodPhotoName: the photo's file name, known before it's taken.
func gopodPhotoName(esn string) string {
	return time.Now().Format("2006-01-02_150405") + "_" + strings.ToLower(strings.ReplaceAll(gopodBrobotName(esn), " ", "")) + ".jpg"
}

// gopodPhotoPage: the page path of a photo (or a picture made beside it).
func gopodPhotoPage(name string) string {
	return "/" + filepath.Base(gopodPhotoDir) + "/" + filepath.Base(name)
}

// gopodPhotoDescribe asks the vision model what's in the photo, simply.
var gopodPhotoDescribe = func(data []byte) (string, error) {
	return gopodPhotoDescribeWith(data, gopodBrobotsSettings.Photo.Prompt)
}

// gopodPhotoDetailPrompt: with statement 1's objects, the vision model is asked for two more
// details about them (photo.detail_prompt), else photo.prompt.
func gopodPhotoDetailPrompt(objects []string) string {
	p := gopodBrobotsSettings.Photo
	if list := gopodQuickLookList(objects); list != "" && strings.TrimSpace(p.DetailPrompt) != "" {
		return strings.ReplaceAll(p.DetailPrompt, "{objects}", list)
	}
	return p.Prompt
}

// gopodPhotoDescribeWith asks the vision model about the photo with this prompt.
var gopodPhotoDescribeWith = func(data []byte, prompt string) (string, error) {
	out, err := gopodVisionAsk(data, prompt, 60)
	if err != nil {
		return "", err
	}
	saw := gopodPhotoFirstSentence(out)
	if saw == "" {
		return "", errors.New("the vision model said nothing")
	}
	return saw, nil
}

// gopodVisionAsk: the vision model's whole answer about a picture.
var gopodVisionAsk = func(data []byte, prompt string, most int) (string, error) {
	p := gopodBrobotsSettings.Photo
	base, _, ok := gopodWarmupOllama()
	if !ok || p.VisionModel == "" {
		return "", errors.New("no Ollama endpoint or photo.vision_model")
	}
	body, _ := json.Marshal(map[string]any{
		"model": p.VisionModel, "prompt": prompt, "stream": false, "keep_alive": p.KeepAlive,
		"options": map[string]any{"num_predict": most},
		"images":  []string{base64.StdEncoding.EncodeToString(data)},
	})
	client := &http.Client{Timeout: gopodSecondsToDuration(p.DescribeMaxSeconds)}
	resp, err := client.Post(base+"/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.Response), nil
}

// gopodPhotoTake takes the photo, shows it, and returns the Brobot's first answer (statement 1) and
// the rest, said after his reply.
func gopodPhotoTake(robot *vector.Vector, ctx context.Context, esn string) (string, func()) {
	p := gopodBrobotsSettings.Photo
	name := gopodBrobotName(esn)
	file := gopodPhotoName(esn)
	egg := strings.TrimSpace(p.EggName) != "" && !gopodNotEgg("photo") && !gopodScriptPlaying.Load() && !gopodRunActive() && gopodPopupAllowed("photo")
	eggSaid := make(chan struct{})
	if egg {
		planned := gopodPhotoPage(file)
		if p.QuickLook.Pictures {
			enhanced, boxes := gopodQuickLookPictures(file)
			planned += "|" + gopodPhotoPage(enhanced) + "|" + gopodPhotoPage(boxes)
		}
		go func() {
			defer close(eggSaid)
			gopodShowEgg(name, strings.TrimSpace(p.EggName), planned)
		}()
	} else {
		close(eggSaid)
	}
	defer func() { <-eggSaid }()
	data, err := gopodPhotoCapture(robot, ctx)
	if err != nil {
		logger.Println("GOPOD_PHOTO " + name + " no photo: " + err.Error())
		return p.FailText, nil
	}
	page, err := gopodPhotoSave(file, data)
	if err != nil {
		logger.Println("GOPOD_PHOTO save failed: " + err.Error())
	}
	// Statement 1 first: the Edge TPU's quick look (well under a second), so the vision model can be
	// asked for more than it named.
	var objects []string
	if page != "" {
		objects, err = gopodQuickLookObjects(filepath.Join(gopodPhotoDir, filepath.Base(page)))
		if err != nil {
			logger.Println("GOPOD_QUICK_LOOK " + name + ": " + err.Error())
		}
	}
	pictures, look := page, data
	if page != "" && gopodBrobotsSettings.Photo.QuickLook.Pictures {
		enhanced, boxes := gopodQuickLookPictures(filepath.Join(gopodPhotoDir, filepath.Base(page)))
		dir := filepath.Dir(page)
		if b, err := os.ReadFile(enhanced); err == nil && len(b) > 0 {
			pictures += "|" + dir + "/" + filepath.Base(enhanced)
			look = b
		}
		if gopodFileExists(boxes) {
			pictures += "|" + dir + "/" + filepath.Base(boxes)
		}
	}
	looked := make(chan struct{})
	var notes string
	go func() {
		defer close(looked)
		started := time.Now()
		saw, err := gopodPhotoDescribeWith(look, gopodPhotoDetailPrompt(objects))
		if err != nil {
			logger.Println("GOPOD_PHOTO describe failed: " + err.Error())
			return
		}
		notes = saw
		logger.Println(fmt.Sprintf("GOPOD_PHOTO %s saw (%s, %.1fs): %s", name, p.VisionModel, time.Since(started).Seconds(), saw))
		gopodPhotoLast.Lock()
		gopodPhotoLast.saw[esn] = saw
		gopodPhotoLast.Unlock()
	}()
	if page != "" && !egg {
		GopodChatFeedAdd(gopodBrobotsSettings.TopicPictures.Role, pictures)
	}
	first := strings.TrimSpace(p.QuickNoneText)
	if list := gopodQuickLookList(objects); list != "" {
		first = strings.TrimSpace(strings.ReplaceAll(p.QuickText, "{objects}", list))
	}
	second := func() {
		stalls := []string{p.GatheringText}
		if len(p.StallLines) > 0 {
			stalls = append(stalls, p.StallLines[rand.Intn(len(p.StallLines))])
		}
		for _, line := range stalls {
			select {
			case <-looked:
			default:
				if line = strings.TrimSpace(line); line != "" {
					gopodRecordSay(name, line)
				}
			}
		}
		<-looked
		gopodRecordSay(name, gopodPhotoDetail(notes, objects))
	}
	return first, second
}

// gopodPhotoDetail: statement 2, built up from statement 1 - what the vision model saw, its own
// lead-in ("The image shows ...") taken off, as "I see ..." (photo.detail_text).
func gopodPhotoDetail(saw string, objects []string) string {
	p := gopodBrobotsSettings.Photo
	saw = strings.TrimSpace(saw)
	for _, prefix := range p.DetailStripPrefixes {
		prefix = strings.TrimSpace(prefix)
		if prefix != "" && len(saw) > len(prefix) && strings.EqualFold(saw[:len(prefix)], prefix) {
			saw = strings.TrimLeft(strings.TrimSpace(saw[len(prefix):]), ",: ")
			break
		}
	}
	if saw == "" {
		if len(objects) > 0 {
			return strings.TrimSpace(p.DetailUnclearText)
		}
		return strings.TrimSpace(p.UnclearText)
	}
	if len(objects) > 0 && !gopodPhotoSaysMore(saw, objects) {
		return strings.TrimSpace(p.DetailUnclearText)
	}
	saw = strings.TrimRight(gopodPhotoLowerFirst(saw), ".!? ")
	return strings.TrimSpace(strings.ReplaceAll(p.DetailText, "{saw}", saw+"."))
}

// gopodPhotoPlainWords: words that add no detail of their own.
var gopodPhotoPlainWords = map[string]bool{
	"a": true, "an": true, "the": true, "and": true, "of": true, "in": true, "on": true, "is": true,
	"are": true, "with": true, "there": true, "this": true, "that": true, "it": true, "some": true,
	"image": true, "photo": true, "picture": true, "shows": true, "see": true, "i": true,
}

// gopodPhotoSaysMore: the look names something beyond statement 1's objects.
func gopodPhotoSaysMore(saw string, objects []string) bool {
	known := map[string]bool{}
	for _, o := range objects {
		for _, w := range gopodHighlightWords(o) {
			known[w.word] = true
			known[w.word+"s"] = true
		}
	}
	for _, w := range gopodHighlightWords(saw) {
		if !known[w.word] && !gopodPhotoPlainWords[w.word] {
			return true
		}
	}
	return false
}

// gopodPhotoRecall: the Brobot's answer to "tell me what you saw".
func gopodPhotoRecall(esn string) string {
	p := gopodBrobotsSettings.Photo
	gopodPhotoLast.Lock()
	saw := gopodPhotoLast.saw[esn]
	gopodPhotoLast.Unlock()
	if saw == "" {
		return p.NothingText
	}
	return strings.TrimSpace(strings.ReplaceAll(p.RecallText, "{saw}", gopodPhotoLowerFirst(saw)))
}

// gopodPhotoFirstSentence: the description's first sentence only - simply, never a paragraph.
func gopodPhotoFirstSentence(s string) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	for i := 0; i < len(s); i++ {
		if (s[i] == '.' || s[i] == '!' || s[i] == '?') && (i+1 == len(s) || s[i+1] == ' ') {
			return strings.TrimSpace(s[:i+1])
		}
	}
	return s
}

// gopodPhotoLowerFirst: "A table and..." -> "a table and..." (it follows "I see").
func gopodPhotoLowerFirst(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 1 && s[0] >= 'A' && s[0] <= 'Z' && !(s[1] >= 'A' && s[1] <= 'Z') {
		return strings.ToLower(s[:1]) + s[1:]
	}
	return s
}

// ## EOF
