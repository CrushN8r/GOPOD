// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_picture_describe.go
package wirepod_ttr

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

var gopodPictureSeen = struct {
	sync.Mutex
	cache map[string]string
	saw   string        // the newest pop-up's description
	done  chan struct{} // closed once saw is set (or the look failed)
	at    time.Time
}{cache: map[string]string{}}

func gopodPictureDescriptionsPath() string {
	return filepath.Join(gopodChipperRoot(), "webroot", "gopod_picture_descriptions.json")
}

type gopodPictureEntry struct {
	MD5    string   `json:"md5"`
	Points []string `json:"points"`
}

type gopodPictureFile struct {
	About    string                       `json:"about"`
	Pictures map[string]gopodPictureEntry `json:"pictures"`
}

func gopodPictureLoad() gopodPictureFile {
	var f gopodPictureFile
	if data, err := os.ReadFile(gopodPictureDescriptionsPath()); err == nil {
		json.Unmarshal(data, &f)
	}
	if f.Pictures == nil {
		f.Pictures = map[string]gopodPictureEntry{}
	}
	return f
}

// gopodEggPictures: every egg's picture files (each topic's, and describe_on_ask's).
func gopodEggPictures() []string {
	tp := gopodBrobotsSettings.TopicPictures
	seen := map[string]bool{}
	var out []string
	add := func(f string) {
		if f != "" && !seen[f] {
			seen[f] = true
			out = append(out, f)
		}
	}
	for _, fs := range tp.Pictures {
		for _, f := range fs {
			add(f)
		}
	}
	for _, fs := range tp.DescribeOnAsk {
		for _, f := range fs {
			add(f)
		}
	}
	sort.Strings(out)
	return out
}

// gopodPicturePoints: the vision model's answer as points, one per line.
func gopodPicturePoints(answer string) []string {
	var out []string
	for _, l := range strings.Split(answer, "\n") {
		l = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(l), "-*•0123456789.) "))
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}

// gopodPictureDescribeAhead: points for every egg picture with none or changed.
func gopodPictureDescribeAhead() {
	tp := gopodBrobotsSettings.TopicPictures
	f := gopodPictureLoad()
	for file, e := range f.Pictures {
		gopodPictureCache(file, e.Points)
	}
	time.Sleep(gopodTiming("picture_describe_delay_seconds"))
	for _, file := range gopodEggPictures() {
		data, err := os.ReadFile(filepath.Join(gopodChipperRoot(), "webroot", tp.Folder, file))
		if err != nil {
			continue
		}
		sum := fmt.Sprintf("%x", md5.Sum(data))
		if e, ok := f.Pictures[file]; ok && e.MD5 == sum && len(e.Points) > 0 {
			continue
		}
		started := time.Now()
		answer, err := gopodVisionAsk(data, tp.DescribePrompt, 120)
		points := gopodPicturePoints(answer)
		if err != nil || len(points) == 0 {
			logger.Println(fmt.Sprintf("GOPOD_PICTURE_DESCRIBE %s not described: %v", file, err))
			continue
		}
		logger.Println(fmt.Sprintf("GOPOD_PICTURE_DESCRIBE saved %s (%.1fs): %s", file, time.Since(started).Seconds(), strings.Join(points, " | ")))
		f = gopodPictureLoad()
		f.Pictures[file] = gopodPictureEntry{MD5: sum, Points: points}
		if out, err := json.MarshalIndent(f, "", "  "); err == nil {
			gopodWriteFile(gopodPictureDescriptionsPath(), append(out, '\n'), 0o644)
		}
		gopodPictureCache(file, points)
	}
}

func gopodPictureCache(file string, points []string) {
	gopodPictureSeen.Lock()
	gopodPictureSeen.cache[file] = strings.Join(points, "; ")
	gopodPictureSeen.Unlock()
}

func init() {
	if strings.HasSuffix(os.Args[0], ".test") {
		return
	}
	go gopodPictureDescribeAhead()
}

// gopodPictureDescribeStart: the pictures of a pop-up that's showing now - every egg's.
func gopodPictureDescribeStart(files []string) {
	gopodPictureDescribeFiles(files)
}

// gopodPictureDescribeOnAsk: a question with one of topic_pictures.describe_on_ask's words
// ("chocolate") has its pictures looked at for the reply, pop-up or not.
func gopodPictureDescribeOnAsk(question string) {
	lower := strings.ToLower(question)
	for word, files := range gopodBrobotsSettings.TopicPictures.DescribeOnAsk {
		if gopodHasWholeWords(lower, strings.ToLower(word)) {
			gopodPictureDescribeFiles(files)
			return
		}
	}
}

// gopodPictureDescribeFiles looks at each file (once per Wire-Pod start each) in the background;
// gopodPictureNote waits for all of them.
func gopodPictureDescribeFiles(files []string) {
	if len(files) == 0 {
		return
	}
	tp := gopodBrobotsSettings.TopicPictures
	done := make(chan struct{})
	gopodPictureSeen.Lock()
	gopodPictureSeen.saw, gopodPictureSeen.done, gopodPictureSeen.at = "", done, time.Now()
	gopodPictureSeen.Unlock()
	fresh := 0
	gopodPictureSeen.Lock()
	for _, file := range files {
		if _, ok := gopodPictureSeen.cache[file]; !ok {
			fresh++
		}
	}
	gopodPictureSeen.Unlock()
	if text := gopodPictureWaitText(fresh); text != "" {
		go GopodSystemSayHeard(text, 0)
	}
	go func() {
		var seen []string
		for _, file := range files {
			gopodPictureSeen.Lock()
			saw, ok := gopodPictureSeen.cache[file]
			gopodPictureSeen.Unlock()
			if !ok {
				started := time.Now()
				data, err := os.ReadFile(filepath.Join(gopodChipperRoot(), "webroot", tp.Folder, file))
				if err != nil {
					logger.Println("GOPOD_PICTURE_DESCRIBE " + file + " unreadable: " + err.Error())
					continue
				}
				saw, err = gopodPhotoDescribe(data)
				if err != nil {
					logger.Println("GOPOD_PICTURE_DESCRIBE " + file + " failed: " + err.Error())
					continue
				}
				logger.Println(fmt.Sprintf("GOPOD_PICTURE_DESCRIBE %s (%.1fs): %s", file, time.Since(started).Seconds(), saw))
				gopodPictureSeen.Lock()
				gopodPictureSeen.cache[file] = saw
				gopodPictureSeen.Unlock()
			}
			seen = append(seen, saw)
		}
		gopodPictureSeen.Lock()
		if gopodPictureSeen.done == done {
			gopodPictureSeen.saw = strings.Join(seen, " Also: ")
		}
		gopodPictureSeen.Unlock()
		close(done)
	}()
}

// gopodPictureWaitText: "Looking at 2 pictures.
func gopodPictureWaitText(n int) string {
	tp := gopodBrobotsSettings.TopicPictures
	if n <= 0 || strings.TrimSpace(tp.DescribeWaitText) == "" {
		return ""
	}
	word := "pictures"
	if n == 1 {
		word = "picture"
	}
	return strings.NewReplacer("{n}", fmt.Sprint(n), "{pictures}", word).Replace(tp.DescribeWaitText)
}

// gopodPictureForget: a new question - the last pop-up's description is dropped (what was seen
// stays cached for the next time it pops up).
func gopodPictureForget() {
	gopodPictureSeen.Lock()
	gopodPictureSeen.done, gopodPictureSeen.saw = nil, ""
	gopodPictureSeen.Unlock()
}

// gopodPictureNote: the note for a reply to the newest pop-up (within two minutes of it), waiting
// for the description at most describe_wait_seconds.
func gopodPictureNote() string {
	tp := gopodBrobotsSettings.TopicPictures
	gopodPictureSeen.Lock()
	done, at := gopodPictureSeen.done, gopodPictureSeen.at
	gopodPictureSeen.Unlock()
	if done == nil || time.Since(at) > 2*time.Minute || strings.TrimSpace(tp.DescribeNote) == "" {
		return ""
	}
	select {
	case <-done:
	case <-time.After(gopodSecondsToDuration(tp.DescribeWaitSeconds)):
		logger.Println("GOPOD_PICTURE_DESCRIBE not ready in time - the reply goes on without it")
		return ""
	}
	gopodPictureSeen.Lock()
	saw := gopodPictureSeen.saw
	gopodPictureSeen.Unlock()
	if saw == "" {
		return ""
	}
	return strings.ReplaceAll(tp.DescribeNote, "{saw}", saw)
}

// ## EOF
