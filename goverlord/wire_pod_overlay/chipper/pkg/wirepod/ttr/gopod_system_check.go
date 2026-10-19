// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_check.go
package wirepod_ttr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// system-check (2026-10-18, step 4 of the clean machine, operator: "all I want
// left to be done, is my input, flexible data sets * GOPOD should be robust &
// bullet proof enough to handle updates & additions to what's already there *
// audio/image/video, scriptable edits"): every file the data names - in the
// settings, the chapters' topic files and the songs - is looked for, and every
// song's story.md step is looked for in its knobs.json, so a missing picture,
// sound or step shows before a run, not in one. Nothing is changed. Run at
// Wire-Pod's start (GOPOD_CHECK lines in the journal) and typed as
// "system-check" on the chat page's keyboard.

// gopodCheckFileRE: a value that names a file.
var gopodCheckFileRE = regexp.MustCompile(`(?i)\.(wav|mp3|ogg|png|jpe?g|webp|svg|gif|mp4|webm|json|py|tflite|txt)$`)

// gopodCheckFind: the value is a file that's there - looked for where GOPOD
// would look: a page path under webroot, ~ (home), an absolute path, or by name
// in the chipper folder, webroot, the voice folder or the pictures folder.
func gopodCheckFind(value string, extra ...string) bool {
	root := gopodChipperRoot()
	web := filepath.Join(root, "webroot")
	v := strings.TrimSpace(strings.SplitN(value, "#", 2)[0])
	if v == "" {
		return true
	}
	var tries []string
	switch {
	case strings.HasPrefix(v, "~/"):
		tries = append(tries, gopodHomePath(v))
	case strings.HasPrefix(v, "/") && gopodFileExists(v):
		return true
	case strings.HasPrefix(v, "/"):
		tries = append(tries, filepath.Join(web, v))
	default:
		tries = append(tries, filepath.Join(root, v), filepath.Join(web, v), filepath.Join(gopodReporterAudioDir(), v),
			filepath.Join(web, gopodBrobotsSettings.TopicPictures.Folder, v))
		for _, dir := range extra {
			tries = append(tries, filepath.Join(dir, v))
		}
	}
	for _, t := range tries {
		if gopodFileExists(t) {
			return true
		}
	}
	return false
}

// gopodCheckWalk: every file named in a JSON value (keys ending "about" skipped).
func gopodCheckWalk(v any, path []string, found func(path, value string)) {
	switch t := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if !strings.HasSuffix(k, "about") {
				gopodCheckWalk(t[k], append(path, k), found)
			}
		}
	case []any:
		for i, x := range t {
			gopodCheckWalk(x, append(path, fmt.Sprint(i)), found)
		}
	case string:
		if gopodCheckFileRE.MatchString(strings.SplitN(t, "#", 2)[0]) {
			found(strings.Join(path, "."), t)
		}
	}
}

// gopodSystemCheckProblems: what's missing or doesn't match, and how many things were checked.
func gopodSystemCheckProblems() ([]string, int) {
	var problems []string
	checked := 0
	root := gopodChipperRoot()
	web := filepath.Join(root, "webroot")
	// 1. The settings: every file they name.
	var settings map[string]any
	if data, err := os.ReadFile(filepath.Join(web, "gopod_brobots_settings.json")); err != nil || json.Unmarshal(data, &settings) != nil {
		problems = append(problems, "gopod_brobots_settings.json can't be read")
	} else {
		gopodCheckWalk(settings, nil, func(path, value string) {
			checked++
			if !gopodCheckFind(value) {
				problems = append(problems, "settings "+path+": no file "+value)
			}
		})
		// The weather faces and sounds the page draws and plays.
		if wc, ok := settings["weather_card"].(map[string]any); ok {
			folder, _ := wc["folder"].(string)
			faces, _ := wc["faces"].([]any)
			for _, f := range faces {
				face, _ := f.(map[string]any)
				for key, dir := range map[string]string{"face": "faces", "night_face": "faces", "sound": "sounds", "night_sound": "sounds"} {
					name, _ := face[key].(string)
					if name == "" {
						continue
					}
					ext := ".webp"
					if dir == "sounds" {
						ext = ".mp3"
					}
					checked++
					if !gopodFileExists(filepath.Join(web, folder, dir, name+ext)) {
						problems = append(problems, "settings weather_card: no "+dir+" file "+name+ext)
					}
				}
			}
		}
	}
	// 2. The chapters: each topic's words and pictures; egg names and reminder eggs.
	topics := map[string]bool{}
	for _, p := range gopodTopicRegistry() {
		topics[p.Name] = true
		checked++
		if len(p.Triggers) == 0 {
			problems = append(problems, "topic "+p.Name+": no words: line (nothing finds it)")
		}
		for _, f := range gopodBrobotsSettings.TopicPictures.Pictures[p.Name] {
			checked++
			if !gopodFileExists(filepath.Join(web, gopodBrobotsSettings.TopicPictures.Folder, f)) {
				problems = append(problems, "topic "+p.Name+": no picture "+f+" in webroot/"+gopodBrobotsSettings.TopicPictures.Folder)
			}
		}
	}
	// 2026-10-18 (operator: "system-check: it would warn when a topic's trigger
	// words include one of Wire-Pod's own phrases * we set up a wire-pod intents
	// list or something similar, didn't we? Use that?"): a topic word that is, or
	// holds, one of Wire-Pod's own intent phrases (intent-data/en-US.json) or a
	// GOPOD command phrase (native_cube.intents) - the line would go to Vector's
	// own answer, not the egg.
	phrases := gopodNativePhrases()
	for _, p := range gopodTopicRegistry() {
		for _, w := range p.Triggers {
			w = strings.ToLower(strings.TrimSpace(w))
			checked++
			for _, ph := range phrases {
				if gopodPhraseOverlap(w, ph[0]) {
					problems = append(problems, "topic "+p.Name+": word \""+w+"\" overlaps Wire-Pod's "+ph[1]+" \""+ph[0]+"\"")
					break
				}
			}
		}
	}
	for name := range gopodBrobotsSettings.TopicPictures.EggNames {
		checked++
		if !topics[name] {
			problems = append(problems, "topic_pictures.egg_names: no topic "+name)
		}
	}
	for _, c := range gopodBrobotsSettings.InterviewChapters.Chapters {
		for _, t := range c.Topics {
			checked++
			if !topics[t] {
				problems = append(problems, "chapter "+c.Title+": no topic file for "+t)
			}
		}
	}
	// 3. The songs: every story.md step in knobs.json, every file the knobs name.
	songs, _ := gopodAllSongs()
	names := make([]string, 0, len(songs))
	for n := range songs {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		dir := songs[n]
		song := filepath.Base(dir)
		data, err := os.ReadFile(filepath.Join(dir, "knobs.json"))
		if err != nil {
			continue
		}
		steps, _, _ := gopodSystemSetKnobsFileData(data)
		have := map[string]bool{}
		for _, s := range steps {
			have[s.StepID] = true
		}
		// 2026-10-18 (operator: "files match check for story.md, zKnobs.json,
		// knob.json"): zKnobs.json has the same steps as knobs.json.
		if zdata, err := os.ReadFile(filepath.Join(dir, "zKnobs.json")); err == nil {
			zsteps, _, _ := gopodSystemSetKnobsFileData(zdata)
			zhave := map[string]bool{}
			for _, s := range zsteps {
				zhave[s.StepID] = true
				checked++
				if !have[s.StepID] {
					problems = append(problems, "song "+song+": zKnobs.json step "+s.StepID+" isn't in knobs.json")
				}
			}
			for _, s := range steps {
				if !zhave[s.StepID] {
					problems = append(problems, "song "+song+": knobs.json step "+s.StepID+" isn't in zKnobs.json")
				}
			}
		}
		if story, err := os.ReadFile(filepath.Join(dir, "story.md")); err == nil {
			for _, line := range strings.Split(string(story), "\n") {
				if id := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "## STEP ")); strings.HasPrefix(strings.TrimSpace(line), "## STEP ") && id != "" {
					checked++
					if !have[id] {
						problems = append(problems, "song "+song+": story.md step "+id+" isn't in knobs.json")
					}
				}
			}
		}
		var knobs any
		if json.Unmarshal(data, &knobs) == nil {
			gopodCheckWalk(knobs, nil, func(path, value string) {
				checked++
				if !gopodCheckFind(value, dir) {
					problems = append(problems, "song "+song+" knobs "+path+": no file "+value)
				}
			})
		}
	}
	return problems, checked
}

// gopodNativePhrases: Wire-Pod's own intent phrases and GOPOD's command
// phrases, each with where it's from.
func gopodNativePhrases() [][2]string {
	var out [][2]string
	intents := vars.IntentList
	if len(intents) == 0 {
		if data, err := os.ReadFile(gopodHomePath("~/wire-pod/chipper/intent-data/en-US.json")); err == nil {
			json.Unmarshal(data, &intents)
		}
	}
	for _, in := range intents {
		for _, k := range in.Keyphrases {
			if k = strings.ToLower(strings.TrimSpace(k)); k != "" {
				out = append(out, [2]string{k, in.Name})
			}
		}
	}
	for _, c := range gopodBrobotsSettings.NativeCube.Intents {
		for _, k := range c.Phrases {
			if k = strings.ToLower(strings.TrimSpace(k)); k != "" {
				out = append(out, [2]string{k, "GOPOD command " + c.Intent})
			}
		}
	}
	return out
}

// gopodPhraseOverlap: the same words, or one holds the other as whole words
// (the one held at least two words long, so "name" alone isn't "my name is").
func gopodPhraseOverlap(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	if a == b {
		return true
	}
	short, long := a, b
	if len(short) > len(long) {
		short, long = long, short
	}
	if len(strings.Fields(short)) < 2 {
		return false
	}
	return strings.Contains(" "+long+" ", " "+short+" ")
}

// gopodSystemCheck: the answer to "system-check".
func gopodSystemCheck() string {
	problems, checked := gopodSystemCheckProblems()
	if len(problems) == 0 {
		return fmt.Sprintf("system-check: all good - %d things checked (settings, chapters, songs).", checked)
	}
	out := []string{fmt.Sprintf("system-check: %d to fix (%d things checked):", len(problems), checked)}
	for i, p := range problems {
		if i == 25 {
			out = append(out, fmt.Sprintf("  ... %d more (system-find GOPOD_CHECK)", len(problems)-25))
			break
		}
		out = append(out, "  "+p)
	}
	return gopodJoinLines(out)
}

// gopodSystemCheckAtStart writes every problem to the journal at Wire-Pod's start.
func gopodSystemCheckAtStart() {
	problems, checked := gopodSystemCheckProblems()
	for _, p := range problems {
		logger.Println("GOPOD_CHECK " + p)
	}
	logger.Println(fmt.Sprintf("GOPOD_CHECK %d things checked, %d to fix", checked, len(problems)))
}

// ## EOF
