// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_source_check.go
package wirepod_ttr

// The source check (2026-10-31, operator: "Please ensure all my system checks,
// check from the root of the single source of truth & check the builds up from
// there concurrently, so lower level bugs are caught to reduce this wasteful
// direction we're going in"). Part of system-check, at every start too. For every
// chat participant at once, from his master source up, layer by layer:
//
//	0 source   his profile (the master format) reads; its head has his label;
//	           his Self has every line of the intro (nickname, name, job title,
//	           role, who helps me) - a line still "(to come)" is named
//	1 labels   everything filled from his labels has nothing left unfilled:
//	           his profile, prompt, who-line, cast line, intro recipe, brain
//	2 voice    a Gomad has a voice (his own or voice_of's)
//	3 the chat he's in the cast - everyone else knows he's here
//	4 the keys every live numpad key reaches someone
//	5 headers  the system label fills from its voice persona's labels
//	6 roster   everyone in the chat has a line in the crew roster

import (
	"fmt"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var gopodUnfilledRE = regexp.MustCompile(`\{(name|nickname|job|self|cast|first|middle|last|fullname|task|asked)\}`)

// gopodSelfLines: the Self lines every profile has, for the intro (2026-11-06:
// the nickname is for banter, not the intro - it may still be to come).
var gopodSelfLines = []string{"My name", "My job title"}

// gopodSourceCheck: the problems, and how many things were checked.
func gopodSourceCheck() ([]string, int) {
	whos := map[string]bool{"Brobot 1": true, "Brobot 2": true}
	for k := range gopodBrobotsSettings.Managers.Personas {
		whos[k] = true
	}
	var mu sync.Mutex
	var problems []string
	checked := 0
	add := func(p ...string) {
		mu.Lock()
		problems = append(problems, p...)
		mu.Unlock()
	}
	count := func(n int) {
		mu.Lock()
		checked += n
		mu.Unlock()
	}
	var wg sync.WaitGroup
	for who := range whos {
		wg.Add(1)
		go func(who string) {
			defer wg.Done()
			p, n := gopodSourceCheckPersona(who)
			add(p...)
			count(n)
		}(who)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		p, n := gopodSourceCheckKeys()
		add(p...)
		count(n)
	}()
	// 7 schema (2026-11-06, operator: "Consistency would then get checked on every
	// restart instead of by you reading runs * please build that in"): every
	// persona file against persona-schema.md, every company file against
	// company-schema.md (gopodSchemaCheck).
	wg.Add(1)
	go func() {
		defer wg.Done()
		p, n := gopodSchemaCheck(whos)
		add(p...)
		count(n)
	}()
	wg.Wait()
	// 6 roster: everyone in the chat but the operator is on the crew roster (gopodCrewRoster).
	if strings.TrimSpace(gopodBrobotsSettings.BrainChain.Roster) != "" {
		roster := gopodCrewRoster()
		for _, c := range gopodBrobotsSettings.HereNow.Cast {
			if c.Who == gopodBrobotsSettings.BrainChain.OperatorWho {
				continue
			}
			count(1)
			if roster[strings.ToLower(gopodCallName(c.Who))] == "" && roster[strings.ToLower(c.Who)] == "" {
				add("roster: " + gopodCallName(c.Who) + " has no line in " + gopodBrobotsSettings.BrainChain.Roster)
			}
		}
	}
	// 5 headers: the system label from its voice persona's labels.
	count(1)
	if sb := gopodBrobotsSettings.SystemBubble; strings.TrimSpace(sb.VoicePersona) != "" {
		hp := gopodBrobotsSettings.HostPersonas[sb.VoicePersona]
		if strings.TrimSpace(hp.Name) == "" || (strings.Contains(sb.Label, "{job}") && strings.TrimSpace(hp.JobName) == "") {
			add("headers: system_bubble.label " + sb.Label + " can't fill from " + sb.VoicePersona + "'s labels (its profile head)")
		}
	}
	sort.Strings(problems)
	return problems, checked
}

// gopodSourceCheckPersona: one participant, from his source up.
func gopodSourceCheckPersona(who string) ([]string, int) {
	var out []string
	n := 0
	bad := func(layer, what string) { out = append(out, "source "+who+" ("+layer+"): "+what) }
	// 0 source
	n++
	head, sections := gopodPersonaProfileRead(gopodPersonaProfileFile(who))
	if head == nil {
		bad("0 source", "no profile ("+gopodPersonaProfileFile(who)+")")
		return out, n
	}
	if strings.TrimSpace(head["label"]) == "" {
		bad("0 source", "his profile has no \"# <full name>\" label")
	}
	raw := gopodPersonaProfileRawSelf(who)
	for _, label := range gopodSelfLines {
		n++
		v, ok := raw[label]
		switch {
		case !ok:
			bad("0 source", "his Self has no \""+label+"\" line (the intro needs it)")
		case strings.HasSuffix(v, "(to come)"):
			bad("0 source", "his Self \""+label+"\" is still to come")
		}
	}
	// 1 labels: nothing left unfilled, wherever his labels go.
	_, persona := gopodBrobotsSettings.Managers.Personas[who]
	filled := map[string]string{
		"profile":      gopodPersonaProfileText(who),
		"intro recipe": gopodIntroRecipe(who, "introduce yourself"),
	}
	if persona {
		filled["prompt"] = gopodPersonaPrompt(who)
		filled["who-line"] = gopodPersonaFill(who, gopodBrobotsSettings.Managers.Personas[who].WhoLine)
	}
	for _, c := range gopodBrobotsSettings.HereNow.Cast {
		if c.Who == who {
			filled["cast line"] = gopodPersonaFill(who, c.Line)
		}
	}
	for what, text := range filled {
		n++
		if m := gopodUnfilledRE.FindString(text); m != "" {
			bad("1 labels", "his "+what+" has "+m+" unfilled")
		}
	}
	_ = sections
	if persona {
		// 2 voice
		n++
		if gopodReporterVoice(who) == "" {
			bad("2 voice", "no voice (reporter_voices.hosts, or voice_of)")
		}
	}
	// 3 the chat
	n++
	inCast := false
	for _, c := range gopodBrobotsSettings.HereNow.Cast {
		inCast = inCast || c.Who == who
	}
	if !inCast {
		bad("3 the chat", "not in here_now.cast - nobody else knows he's here")
	}
	return out, n
}

// gopodPersonaProfileRawSelf: his Self lines by label, as written (to come too).
func gopodPersonaProfileRawSelf(who string) map[string]string {
	out := map[string]string{}
	path := gopodPersonaProfileFile(who)
	data := gopodReadFileString(path)
	// 2026-11-04: a robot's file and the persona it hosts (gopodPersonaProfileRead).
	if head, _ := gopodPersonaProfileReadOne(path); strings.TrimSpace(head["hosts"]) != "" {
		hosted := filepath.Join(filepath.Dir(path), strings.TrimSpace(head["hosts"]))
		if _, err := os.Stat(hosted); err != nil {
			if f := gopodTopicTreeFind(strings.TrimSpace(head["hosts"])); f != "" {
				hosted = f
			}
		}
		data += "\n" + gopodReadFileString(hosted)
	}
	inSelf := false
	for _, line := range strings.Split(data, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "## ") {
			inSelf = strings.TrimSpace(t[3:]) == "Self"
			continue
		}
		if !inSelf {
			continue
		}
		if m := gopodProfilePointRE.FindStringSubmatch(t); m != nil {
			if k, v, ok := strings.Cut(m[1], ":"); ok {
				out[strings.TrimSpace(k)] = strings.TrimSpace(v)
			}
		}
	}
	return out
}

// gopodSourceCheckKeys: 4 the keys - every live numpad key reaches someone.
func gopodSourceCheckKeys() ([]string, int) {
	m, _, err := gopodNumpadMap()
	if err != nil {
		return []string{"source numpad (4 the keys): the map can't be read: " + err.Error()}, 1
	}
	var out []string
	n := 0
	try := func(key string) {
		n++
		if _, _, err := gopodNumpadBrobot(key); err != nil {
			out = append(out, "source numpad (4 the keys): "+key+" is live in the map but reaches nobody: "+err.Error())
		}
	}
	for key, e := range m.NumlockOn {
		if e.Status == "live" {
			try(key)
		}
	}
	for key, e := range m.NumlockOff {
		if e.Status == "live" {
			try(key + gopodNumlockOffSuffix)
		}
	}
	try(gopodNumpadAllKey)
	return out, n
}

// gopodReadFileString: a file's text, "" when it can't be read.
func gopodReadFileString(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// gopodSchemaShape: a topic-cluster md's shape - its head's keys ("- **job:**"),
// its "## " sections in order, and its Self lines' labels ("My name").
type gopodSchemaShape struct {
	keys, sections, self []string
	points           map[string]int
}

func gopodSchemaShapeOf(path string) (gopodSchemaShape, bool) {
	var s gopodSchemaShape
	data, err := os.ReadFile(path)
	if err != nil {
		return s, false
	}
	s.points = map[string]int{}
	inHead, section := true, ""
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case inHead && line == "---":
			inHead = false
		case inHead && strings.HasPrefix(line, "- **"):
			if k, _, ok := strings.Cut(strings.TrimPrefix(line, "- **"), ":**"); ok {
				s.keys = append(s.keys, k)
			}
		case !inHead && strings.HasPrefix(line, "## "):
			section = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			s.sections = append(s.sections, section)
		case !inHead && gopodProfilePointRE.MatchString(line):
			s.points[section]++
			if section == "Self" {
				if label, _, ok := strings.Cut(gopodProfilePointRE.FindStringSubmatch(line)[1], ":"); ok {
					s.self = append(s.self, label)
				}
			}
		}
	}
	return s, true
}

// gopodSchemaCheck: the persona files (each participant's, and the persona a
// robot hosts) against persona-schema.md - its head keys ("why" aside: a nickname
// may have no story), its sections and its Self lines; the company files (a file
// with "## Our company") against company-schema.md - the same sections, as many
// points. A missing one is a problem to fix.
func gopodSchemaCheck(whos map[string]bool) ([]string, int) {
	var out []string
	n := 0
	has := func(list []string, v string) bool {
		for _, x := range list {
			if strings.EqualFold(x, v) {
				return true
			}
		}
		return false
	}
	if schemaPath := gopodTopicTreeFind("persona-schema.md"); schemaPath != "" {
		schema, _ := gopodSchemaShapeOf(schemaPath)
		files := map[string]bool{}
		for who := range whos {
			p := gopodPersonaProfileFile(who)
			files[p] = true
			if head, _ := gopodPersonaProfileReadOne(p); head != nil {
				if h := strings.TrimSpace(head["hosts"]); h != "" {
					if f := gopodTopicTreeFind(h); f != "" {
						files[f] = true
					}
				}
			}
		}
		for p := range files {
			shape, ok := gopodSchemaShapeOf(p)
			n++
			if !ok {
				continue
			}
			name := filepath.Base(p)
			for _, k := range schema.keys {
				if k == "why" || k == "about" || k == "hosts" { // a story may be blank; only robots host
					continue
				}
				n++
				if !has(shape.keys, k) {
					out = append(out, "schema "+name+": no \"- **"+k+":**\" head line (persona-schema.md)")
				}
			}
			for _, sec := range schema.sections {
				n++
				if !has(shape.sections, sec) {
					out = append(out, "schema "+name+": no \"## "+sec+"\" section (persona-schema.md)")
				}
			}
			for _, label := range schema.self {
				n++
				if !has(shape.self, label) {
					out = append(out, "schema "+name+": its Self has no \""+label+"\" line (persona-schema.md)")
				}
			}
		}
	}
	if schemaPath := gopodTopicTreeFind("company-schema.md"); schemaPath != "" {
		schema, _ := gopodSchemaShapeOf(schemaPath)
		_ = filepath.WalkDir(filepath.Dir(schemaPath), func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(p, ".md") || p == schemaPath {
				return nil
			}
			shape, ok := gopodSchemaShapeOf(p)
			if !ok || !has(shape.sections, "Our company") {
				return nil
			}
			for _, sec := range schema.sections {
				n++
				switch {
				case !has(shape.sections, sec):
					out = append(out, "schema "+filepath.Base(p)+": no \"## "+sec+"\" section (company-schema.md)")
				case shape.points[sec] != schema.points[sec]:
					out = append(out, "schema "+filepath.Base(p)+": \"## "+sec+"\" has "+strconv.Itoa(shape.points[sec])+" points, company-schema.md has "+strconv.Itoa(schema.points[sec]))
				}
			}
			return nil
		})
	}
	more, m := gopodSchemaCheckGame()
	return append(out, more...), n + m
}

// gopodSchemaRequired: a schema file's listed lines - the points of its
// "## Required..." and "## Known..." sections, each its name (before ":"), by section.
func gopodSchemaRequired(path string) map[string][]string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	out := map[string][]string{}
	section := ""
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## ") {
			section = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			continue
		}
		if m := gopodProfilePointRE.FindStringSubmatch(line); m != nil && (strings.HasPrefix(section, "Required") || strings.HasPrefix(section, "Known")) {
			name := m[1]
			if strings.HasPrefix(section, "Required") {
				name, _, _ = strings.Cut(m[1], ":") // "label: what it is"; a known part is its whole name
			}
			out[section] = append(out[section], strings.TrimSpace(name))
		}
	}
	return out
}

// gopodSchemaCheckGame (2026-11-06, operator: "build schemas for the other places
// in GOPOD where a schema would earn its keep, most useful first"): every game
// card against _topic_cluster/topic-schema.md, every chapter against
// chapter-schema.md (and every card's chapter is one of them), every song folder
// against songs/_new_song/song-schema.md.
func gopodSchemaCheckGame() ([]string, int) {
	var out []string
	n := 0
	tpl := filepath.Join(gopodTopicRoot(), "_topic_cluster")
	// Chapters: their titles, each with its required lines.
	chapters := map[string]bool{}
	if req := gopodSchemaRequired(filepath.Join(tpl, "chapter-schema.md"))["Required"]; req != nil {
		files, _ := filepath.Glob(filepath.Join(gopodTopicRoot(), "chapters", "*", "chapter.md"))
		for _, f := range files {
			head := gopodSchemaHead(f)
			chapters[strings.ToLower(head["label"])] = true
			for _, k := range req {
				n++
				if strings.TrimSpace(head[k]) == "" {
					out = append(out, "schema chapter "+filepath.Base(filepath.Dir(f))+": no "+k+" line (chapter-schema.md)")
				}
			}
		}
	}
	// Game cards.
	if req := gopodSchemaRequired(filepath.Join(tpl, "topic-schema.md"))["Required"]; req != nil {
		for _, c := range gopodTopicDocCardsAll() {
			head := map[string]string{}
			if at := strings.Index(c.Text, "\n---\n"); at >= 0 {
				for _, line := range strings.Split(c.Text[:at], "\n") {
					if k, v, ok := gopodTopicHeadKV(line); ok {
						head[k] = v
					}
				}
			}
			where := "topic " + c.File + " (" + filepath.Base(c.Path) + ")"
			for _, k := range req {
				n++
				switch {
				case strings.TrimSpace(head[k]) == "":
					out = append(out, "schema "+where+": no "+k+" line (topic-schema.md)")
				case k == "chapter" && len(chapters) > 0 && !chapters[strings.ToLower(head[k])]:
					out = append(out, "schema "+where+": its chapter \""+head[k]+"\" isn't a chapter (chapter-schema.md)")
				}
			}
		}
	}
	songs := filepath.Join(gopodRepoRoot(), "goverlord", "runtime", "songs")
	// Run cards (game.runs) against run-schema.md; a run's song must be a song folder.
	if req := gopodSchemaRequired(filepath.Join(tpl, "run-schema.md")); req != nil {
		for _, run := range gopodBrobotsSettings.Game.Runs {
			raw, _ := json.Marshal(run)
			var m map[string]any
			_ = json.Unmarshal(raw, &m)
			pending, _ := m["pending"].(bool)
			keys := append([]string{}, req["Required"]...)
			if !pending {
				keys = append(keys, req["Required unless pending"]...)
			}
			for _, k := range keys {
				n++
				v := strings.TrimSpace(fmt.Sprint(m[k]))
				switch {
				case m[k] == nil || v == "":
					out = append(out, "schema run "+run.Name+": no "+k+" (run-schema.md)")
				case k == "song":
					if _, err := os.Stat(filepath.Join(songs, v)); err != nil {
						out = append(out, "schema run "+run.Name+": its song "+v+" isn't a song folder (run-schema.md)")
					}
				}
			}
		}
	}
	// Answer chains against answer-chain-schema.md: every part known, every "from" a chain.
	if req := gopodSchemaRequired(filepath.Join(tpl, "answer-chain-schema.md")); req != nil {
		known := map[string]bool{}
		for _, part := range req["Known parts"] {
			known[part] = true
		}
		for name, c := range gopodBrobotsSettings.AnswerChains {
			if c.From != "" {
				n++
				if _, ok := gopodBrobotsSettings.AnswerChains[c.From]; !ok {
					out = append(out, "schema answer chain "+name+": from "+c.From+", no such chain (answer-chain-schema.md)")
				}
			}
			for _, part := range c.Parts {
				n++
				if !known[part] {
					out = append(out, "schema answer chain "+name+": part "+part+" isn't a known part (answer-chain-schema.md)")
				}
			}
		}
	}
	// Songs.
	req := gopodSchemaRequired(filepath.Join(songs, "_new_song", "song-schema.md"))
	if req != nil {
		dirs, _ := os.ReadDir(songs)
		for _, d := range dirs {
			if !d.IsDir() || strings.HasPrefix(d.Name(), "_") || d.Name() == "tools" {
				continue
			}
			dir := filepath.Join(songs, d.Name())
			var knobs map[string]json.RawMessage
			data, err := os.ReadFile(filepath.Join(dir, "knobs.json"))
			if err == nil && json.Unmarshal(data, &knobs) != nil {
				out = append(out, "schema song "+d.Name()+": knobs.json doesn't read (song-schema.md)")
				continue
			}
			if _, wip := knobs["_comment"]; wip && len(knobs) == 1 {
				continue // a song still being written
			}
			// The opener warm-up (song_type llm_warm): its steps, in order (llm-warm-schema.md).
			var kind string
			_ = json.Unmarshal(knobs["song_type"], &kind)
			if warm := gopodSchemaRequired(filepath.Join(songs, "_new_song", "_llm_warm", "llm-warm-schema.md"))["Required steps, in order"]; kind == "llm_warm" && warm != nil {
				var steps []struct {
					Note string `json:"note"`
				}
				_ = json.Unmarshal(knobs["steps"], &steps)
				n++
				var got []string
				for _, s := range steps {
					got = append(got, s.Note)
				}
				if strings.Join(got, " ") != strings.Join(warm, " ") {
					out = append(out, "schema song "+d.Name()+": its steps "+strings.Join(got, ", ")+" aren't the opener's "+strings.Join(warm, ", ")+" (llm-warm-schema.md)")
				}
			}
			for _, f := range req["Required files"] {
				n++
				if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
					out = append(out, "schema song "+d.Name()+": no "+f+" (song-schema.md)")
				}
			}
			for _, k := range req["Required in knobs.json"] {
				n++
				if knobs != nil && knobs[k] == nil {
					out = append(out, "schema song "+d.Name()+": its knobs.json has no "+k+" (song-schema.md)")
				}
			}
		}
	}
	return out, n
}

// gopodSchemaHead: a topic-cluster md's head lines (its "# " label and "- **k:**" lines).
func gopodSchemaHead(path string) map[string]string {
	head := map[string]string{}
	data, err := os.ReadFile(path)
	if err != nil {
		return head
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "---" {
			break
		}
		if k, v, ok := gopodTopicHeadKV(line); ok {
			head[k] = v
		}
	}
	return head
}

// ## EOF
