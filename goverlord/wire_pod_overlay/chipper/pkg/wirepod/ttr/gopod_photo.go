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

// The photo egg (2026-10-14, operator: "Get Photo in front of you ... implement
// anyways"; "robot to getPhoto, then look at photo ... then 'tell me what you
// saw' and robot should be able to continue seamlessly ... describe the image
// simply * Example: I see a table, chair and some book shelves ... * store the
// image to then display in a web page * keep/snapshot the 3-2-1 countdown +
// mirror-mode + shutter animation, then lock as a golden process"; "treat
// these as found easter eggs throughout a live chat"). In the live chat only
// (no draft, no run): "what's in front of you?" - the Brobot takes the photo
// the golden way (gopodPhotoCapture), it pops up in his reply bubble with
// Goverlord's congratulations, he thinks while the vision model
// (photo.vision_model, through Ollama) looks at it, and says what he sees. "Tell
// me what you saw" is answered from what he saw last; the photo is never sent
// again. Settings: photo in gopod_brobots_settings.json.

var gopodPhotoDir = gopodPathWithEnvOverride("GOPOD_PHOTO_DIR", "webroot/photos")

// gopodPhotoLast: what each Brobot saw last (his ESN -> the description).
var gopodPhotoLast = struct {
	sync.Mutex
	saw map[string]string
}{saw: map[string]string{}}

// gopodPhotoKind: "ask" (take a photo), "recall" (tell what he saw) or "" for
// this line - in the live chat only.
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

// gopodPhotoCapture: GOLDEN, kept as Wire-Pod's own getImage takes it
// (DoGetImage, kgsim_cmds.go) - mirror mode on, "3", "2", "1", the photo (high
// resolution), mirror mode off, the shutter animation. A failed or empty photo
// is an error.
func gopodPhotoCapture(robot *vector.Vector, ctx context.Context) ([]byte, error) {
	// 2026-10-14 (operator: "make sure brobots are awake, then 3-2-1, ...
	// otherwise photo is taken as they're waking, head not in position"; live:
	// "3" went 0 ms after control was granted): he wakes and settles first
	// (golden rule 3), photo.awake_seconds.
	time.Sleep(gopodSecondsToDuration(gopodBrobotsSettings.Photo.AwakeSeconds))
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

// gopodPhotoSave keeps the photo for the chat page (webroot/photos, private,
// not in the repo) and returns its page path.
func gopodPhotoSave(esn string, data []byte) (string, error) {
	if err := os.MkdirAll(gopodPhotoDir, 0o755); err != nil {
		return "", err
	}
	name := time.Now().Format("2006-01-02_150405") + "_" + strings.ToLower(strings.ReplaceAll(gopodBrobotName(esn), " ", "")) + ".jpg"
	if err := os.WriteFile(filepath.Join(gopodPhotoDir, name), data, 0o644); err != nil {
		return "", err
	}
	return "/" + filepath.Base(gopodPhotoDir) + "/" + name, nil
}

// gopodPhotoDescribe asks the vision model what's in the photo, simply.
var gopodPhotoDescribe = func(data []byte) (string, error) {
	return gopodPhotoDescribeWith(data, gopodBrobotsSettings.Photo.Prompt)
}

// gopodPhotoDetailPrompt (2026-10-16, operator: "do not repeat without more
// detail * I see a person, I see a person with glasses * build 2 more details
// from 1, no repeats"): with statement 1's objects, the vision model is asked
// for two more details about them (photo.detail_prompt), else photo.prompt.
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

// gopodPhotoTake takes the photo, shows it, and returns the Brobot's first answer
// (statement 1) and the rest, said after his reply. 2026-10-16 (operator: "do
// this in 2 rounds * round 1 is from TPU quick look (while moondream looks) *
// round 2 is from moondream"; then: "thinking please wait > photo > TPU >
// Brobot N: I see [TPU labels] * (if background moondream look still running,
// brobot says gathering details) * Brobot optional statement to stall for time *
// No bubble for notes ... this is background internal stuff * build brobot
// response as statement-2 \"I see ... ...\" from statement-1 & what moondream
// found"). Statement 1 is the Edge TPU's objects, at once ("I see a person, a
// chair and a keyboard."); the vision model looks meanwhile, in the background.
// After his reply: photo.gathering_text if it's still looking, then one of
// photo.stall_lines if it's still looking; then statement 2 (gopodPhotoDetail).
// Failures give photo.fail_text / photo.unclear_text. The photo pops up as the
// easter egg photo.egg_name (up to topic_pictures.max_per_session times), else
// it shows as a picture; Goverlord's congratulations is heard before this
// returns.
func gopodPhotoTake(robot *vector.Vector, ctx context.Context, esn string) (string, func()) {
	p := gopodBrobotsSettings.Photo
	name := gopodBrobotName(esn)
	data, err := gopodPhotoCapture(robot, ctx)
	if err != nil {
		logger.Println("GOPOD_PHOTO " + name + " no photo: " + err.Error())
		return p.FailText, nil
	}
	page, err := gopodPhotoSave(esn, data)
	if err != nil {
		logger.Println("GOPOD_PHOTO save failed: " + err.Error())
	}
	// Statement 1 first: the Edge TPU's quick look (well under a second), so the
	// vision model can be asked for more than it named.
	var objects []string
	if page != "" {
		objects, err = gopodQuickLookObjects(filepath.Join(gopodPhotoDir, filepath.Base(page)))
		if err != nil {
			logger.Println("GOPOD_QUICK_LOOK " + name + ": " + err.Error())
		}
	}
	// 2026-10-17 (operator: "3 images displayed * [raw snapshot][snapshot
	// enhancement attempt][TPU object recognition on enhancement] * I see ...
	// [TPU read on enhancement] * I see ... [moondream read of TPU]"): the
	// enhanced snapshot and its boxes pop up beside the photo, and the vision
	// model looks at the enhanced one, told the TPU's objects.
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
		// The notes stay in the log (background, no bubble).
		logger.Println(fmt.Sprintf("GOPOD_PHOTO %s saw (%s, %.1fs): %s", name, p.VisionModel, time.Since(started).Seconds(), saw))
		gopodPhotoLast.Lock()
		gopodPhotoLast.saw[esn] = saw
		gopodPhotoLast.Unlock()
	}()
	if page != "" {
		// 2026-10-24 (operator: "isn't the photo thing an easter egg also? * if not,
		// make it an easter egg please"): the first photo of a session (live chat)
		// is the "photo" Easter egg - Goverlord's congratulations, the photo the
		// prize; after that, the pictures in their own picture bubble.
		if name := strings.TrimSpace(p.EggName); name != "" && !gopodNotEgg("photo") && !gopodScriptPlaying.Load() && !gopodRunActive() && gopodPopupAllowed("photo") {
			gopodShowEgg(gopodBrobotName(esn), name, pictures)
		} else {
			GopodChatFeedAdd(gopodBrobotsSettings.TopicPictures.Role, pictures)
		}
	}
	first := strings.TrimSpace(p.QuickNoneText)
	if list := gopodQuickLookList(objects); list != "" {
		first = strings.TrimSpace(strings.ReplaceAll(p.QuickText, "{objects}", list))
	}
	// After his reply (kgsim.go): the stalls while it still looks, then statement 2.
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

// gopodPhotoDetail: statement 2, built up from statement 1 - what the vision
// model saw, its own lead-in ("The image shows ...") taken off, as "I see ..."
// (photo.detail_text). 2026-10-16 (operator: "do not repeat without more detail
// * build 2 more details from 1, no repeats"): statement 1's objects aren't
// added again, and a look that names nothing new (only statement 1's objects)
// gives photo.detail_unclear_text, as does one that saw nothing.
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

// gopodPhotoFirstSentence: the description's first sentence only - simply
// ("I see a table, chair and some book shelves", operator), never a paragraph.
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
