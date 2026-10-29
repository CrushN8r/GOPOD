// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_chapters_folder.go
package wirepod_ttr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// The interview's topics live in a chapters folder (operator 2026-10-08: "read a
// 'chapters' folder of chapters, each chapter folder contains editable topic
// files ... topic files are to be loaded for the run"):
//
//	gopod_probes/chapters/
//	  summary.txt                 the whole interview in one line
//	  _always/01_what_else.md     topics that belong to no chapter
//	  01_warm_up/chapter.md       "# <title>" and "- **summary:** ..."
//	  01_warm_up/01_checks.md     one topic-cluster file per topic, in order
//
// A folder's number sets the chapter order; a file's number sets the topic order
// (the number isn't part of the topic's name). 2026-10-29 (operator: "Would
// turning topic clusters into readable md files, help organize? * template out a
// topic-cluster format for md files * then anything can be a topic-cluster for
// the GOPOD game, as a GOPOD formatted topic-cluster md file"): a topic is a
// readable markdown file - its label the heading, each setting a list line, a
// "---" line, then the facts the Brobots get (the numbered value points):
//
//	# the rig                                 its label: how a reminder names it (no heading = never named)
//	- **words:** rig, gear, dolly             the words in a question that load it
//	- **pictures:** gear/crate_2_display.jpg  shown in the chat with it
//	- **reminder:** Hey, that reminds me...   its reminder line
//	- **sentences:** 0                        its own sentence budget (left out = the usual)
//	- **say 1:** One dolly, two crates.       point 1's exact line, said first
//	- **start line:** ...                     a scripted line at the start (one per line)
//	- **end line:** ...                       a scripted line at the end (one per line)
//	- **summary:** Lower crate 1 ... | ...    what a summary ask riffs on (gopod_topic_asks.go)
//	- **source:** MOBILE_GEAR.md              the repo file it tells - the game matches the repo
//	- **ask:** Tell me about yourself         its ask on BROBOT_ASKS.md (left out = "Tell me about <label>")
//	- **pre:** ... / **post:** ...            its note before / after the room has covered it
//	- **closing:** ...                        a closing remark once it's covered (none = none)
//	---
//	The rig: ... Reveal these, in order: 1. ... 2. ...
//
// The template: gopod_probes/_topic_cluster/ (README.md, an example chapter and topic).
// One reader and one writer for the format (gopodTopicHeadKV, gopodTopicHeadLine):
// the game, system-topic and every check go through them.
//
// The facts are read from the folder each time a question comes in; the rest is
// read once, when Wire-Pod starts.
var canonicalChaptersDir = gopodPathWithEnvOverride("GOPOD_CHAPTERS_DIR", "gopod_probes/chapters")

// gopodTopicExt / gopodChapterFile: a topic-cluster file, and a chapter's own file.
const (
	gopodTopicExt    = ".md"
	gopodChapterFile = "chapter.md"
)

// gopodTopicHeadKV reads one settings line of a topic-cluster file: "# label"
// is the label (a chapter's title), "- **key:** value" a setting (a bare
// "key: value" still reads). ok false for any other line.
func gopodTopicHeadKV(line string) (key, value string, ok bool) {
	t := strings.TrimSpace(line)
	if strings.HasPrefix(t, "# ") {
		return "label", strings.TrimSpace(t[2:]), true
	}
	if strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* ") {
		t = strings.TrimSpace(t[2:])
	}
	if strings.HasPrefix(t, "**") {
		if end := strings.Index(t[2:], "**"); end >= 0 {
			k := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(t[2:2+end]), ":"))
			rest := strings.TrimSpace(t[2+end+2:])
			rest = strings.TrimSpace(strings.TrimPrefix(rest, ":"))
			if k != "" {
				return strings.ToLower(k), rest, true
			}
		}
		return "", "", false
	}
	if k, v, found := strings.Cut(t, ":"); found && strings.TrimSpace(k) != "" && !strings.ContainsAny(strings.TrimSpace(k), "#*") {
		return strings.ToLower(strings.TrimSpace(k)), strings.TrimSpace(v), true
	}
	return "", "", false
}

// gopodTopicHeadLine writes one settings line the format's way.
func gopodTopicHeadLine(key, value string) string {
	if key == "label" || key == "title" {
		return "# " + value
	}
	return "- **" + key + ":** " + value
}

type gopodTopicFile struct {
	Name      string
	Path      string
	Label     string
	Words     []string
	Pictures  []string
	Reminder  string
	Sentences *int
	SayLines  map[int]string
	// 2026-10-22: a point in note form, "notes N: title | note | note" (gopodPointNotes).
	Notes      map[int]string
	StartLines []string
	EndLines   []string
	// 2026-10-26 (the topic ask template, gopod_topic_asks.go): the summary a
	// summary ask riffs on, and the repo file(s) the topic tells.
	Summary string
	Sources []string
	// 2026-10-26: its ask on the BROBOT_ASKS.md page (left out = "Tell me about <label>").
	Ask string
	// 2026-10-26 (the topic pre/post template): a note before the room has
	// covered it, one after, and a closing remark after (topic_asks defaults).
	Pre, Post, Closing string
}

type gopodChapterFolder struct {
	Title   string
	Summary string
	Topics  []gopodTopicFile
}

var gopodTopicNumberPrefix = regexp.MustCompile(`^\d+_`)

// gopodReadChaptersFolder reads every chapter folder in order, and the topics
// that belong to no chapter (folders starting with "_", read first).
func gopodReadChaptersFolder() (chapters []gopodChapterFolder, always []gopodTopicFile, summary string) {
	if data, err := os.ReadFile(filepath.Join(canonicalChaptersDir, "summary.txt")); err == nil {
		summary = strings.TrimSpace(string(data))
	}
	entries, _ := os.ReadDir(canonicalChaptersDir)
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(canonicalChaptersDir, e.Name())
		topics := gopodReadTopicFiles(dir)
		if strings.HasPrefix(e.Name(), "_") {
			always = append(always, topics...)
			continue
		}
		c := gopodChapterFolder{Topics: topics}
		if data, err := os.ReadFile(filepath.Join(dir, gopodChapterFile)); err == nil {
			head, _ := gopodSplitTopicFile(string(data))
			for _, kv := range head {
				switch kv[0] {
				case "title", "label":
					c.Title = kv[1]
				case "summary":
					c.Summary = kv[1]
				}
			}
		}
		chapters = append(chapters, c)
	}
	return chapters, always, summary
}

func gopodReadTopicFiles(dir string) []gopodTopicFile {
	entries, _ := os.ReadDir(dir)
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	var out []gopodTopicFile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), gopodTopicExt) || e.Name() == gopodChapterFile {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		t := gopodTopicFile{
			Name:     gopodTopicNumberPrefix.ReplaceAllString(strings.TrimSuffix(e.Name(), gopodTopicExt), ""),
			Path:     path,
			SayLines: map[int]string{},
			Notes:    map[int]string{},
		}
		head, _ := gopodSplitTopicFile(string(data))
		for _, kv := range head {
			key, value := kv[0], kv[1]
			switch {
			case key == "label":
				t.Label = value
			case key == "words":
				t.Words = gopodCommaList(value)
			case key == "pictures":
				t.Pictures = gopodCommaList(value)
			case key == "reminder":
				t.Reminder = value
			case key == "sentences":
				if n, err := strconv.Atoi(value); err == nil {
					t.Sentences = &n
				}
			case key == "summary":
				t.Summary = value
			case key == "source":
				t.Sources = gopodCommaList(value)
			case key == "ask":
				t.Ask = value
			case key == "pre":
				t.Pre = value
			case key == "post":
				t.Post = value
			case key == "closing":
				t.Closing = value
			case key == "start line":
				t.StartLines = append(t.StartLines, value)
			case key == "end line":
				t.EndLines = append(t.EndLines, value)
			case strings.HasPrefix(key, "notes "):
				if n, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(key, "notes "))); err == nil {
					t.Notes[n] = value
				}
			case strings.HasPrefix(key, "say "):
				if n, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(key, "say "))); err == nil {
					t.SayLines[n] = value
				}
			}
		}
		out = append(out, t)
	}
	return out
}

// gopodSplitTopicFile splits a file into its settings lines (gopodTopicHeadKV) (before a "---"
// line) and the facts (after it). A file with no "---" is all settings.
func gopodSplitTopicFile(text string) (head [][2]string, body string) {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			body = strings.Join(lines[i+1:], "\n")
			break
		}
		if k, v, ok := gopodTopicHeadKV(line); ok {
			head = append(head, [2]string{k, v})
		}
	}
	return head, body
}

func gopodCommaList(value string) []string {
	var out []string
	for _, w := range strings.Split(value, ",") {
		if w = strings.TrimSpace(w); w != "" {
			out = append(out, w)
		}
	}
	return out
}

// gopodTopicRegistry: every topic, in load order (no-chapter topics first, then
// chapter by chapter), read fresh from the folder.
func gopodTopicRegistry() []scenarioPacketEntry {
	chapters, always, _ := gopodReadChaptersFolder()
	var out []scenarioPacketEntry
	add := func(t gopodTopicFile) {
		out = append(out, scenarioPacketEntry{Name: t.Name, File: t.Path, Triggers: t.Words})
	}
	for _, t := range always {
		add(t)
	}
	for _, c := range chapters {
		for _, t := range c.Topics {
			add(t)
		}
	}
	return out
}

// gopodReadTopicFacts returns the facts part of a topic file (after "---").
func gopodReadTopicFacts(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	_, body := gopodSplitTopicFile(string(data))
	return []byte(strings.TrimSpace(body) + "\n"), nil
}

// gopodApplyChaptersFolder fills the settings' topic parts from the folder: the
// chapters and labels, the pictures, the say-lines and reminders, the sentence
// budgets, and the topics' scripted lines (ahead of the settings' own lines).
func gopodApplyChaptersFolder(s *gopodBrobotsSettingsFile) {
	chapters, always, summary := gopodReadChaptersFolder()
	if len(chapters) == 0 && len(always) == 0 {
		return
	}
	if summary != "" {
		s.InterviewChapters.Summary = summary
	}
	type chapterJSON struct {
		Title   string   `json:"title"`
		Summary string   `json:"summary"`
		Topics  []string `json:"topics"`
	}
	type snippetJSON struct {
		When  string   `json:"when"`
		At    string   `json:"at,omitempty"`
		Lines []string `json:"lines"`
	}
	var chapterList []chapterJSON
	var snippets []snippetJSON
	labels := map[string]string{}
	pictures := map[string][]string{}
	sayLines := map[string]string{}
	reminders := map[string]string{}
	notes := map[string]string{}
	sentences := map[string]int{}
	topic := func(t gopodTopicFile) {
		if t.Label != "" {
			labels[t.Name] = t.Label
		}
		if len(t.Pictures) > 0 {
			pictures[t.Name] = t.Pictures
		}
		for n, line := range t.SayLines {
			sayLines[t.Name+" "+strconv.Itoa(n)] = line
		}
		for n, line := range t.Notes {
			notes[t.Name+" "+strconv.Itoa(n)] = line
		}
		if t.Reminder != "" {
			reminders[t.Name] = t.Reminder
		}
		if t.Sentences != nil {
			sentences[t.Name] = *t.Sentences
		}
		if len(t.StartLines) > 0 {
			snippets = append(snippets, snippetJSON{When: t.Name, Lines: t.StartLines})
		}
		if len(t.EndLines) > 0 {
			snippets = append(snippets, snippetJSON{When: t.Name, At: "end", Lines: t.EndLines})
		}
	}
	topicNames := map[string]bool{}
	for _, t := range always {
		topic(t)
		topicNames[t.Name] = true
	}
	for _, c := range chapters {
		cj := chapterJSON{Title: c.Title, Summary: c.Summary}
		for _, t := range c.Topics {
			topic(t)
			topicNames[t.Name] = true
			cj.Topics = append(cj.Topics, t.Name)
		}
		chapterList = append(chapterList, cj)
	}
	// The settings' own scripted lines that aren't about a topic stay, after the topics'.
	own, _ := json.Marshal(s.ScriptedSnippets.Lines)
	var ownLines []map[string]any
	json.Unmarshal(own, &ownLines)
	var lines []any
	for _, l := range snippets {
		lines = append(lines, l)
	}
	for _, l := range ownLines {
		if when, _ := l["when"].(string); !topicNames[when] {
			lines = append(lines, l)
		}
	}
	fill := func(v any, into any) {
		data, _ := json.Marshal(v)
		json.Unmarshal(data, into)
	}
	s.InterviewChapters.Chapters = nil
	fill(chapterList, &s.InterviewChapters.Chapters)
	s.ScriptedSnippets.Lines = nil
	fill(lines, &s.ScriptedSnippets.Lines)
	s.InterviewChapters.TopicLabels = labels
	s.TopicPictures.Pictures = pictures
	s.ValuePoints.SayLines = sayLines
	s.ValuePoints.Notes = notes
	s.ValuePoints.Reminders = reminders
	s.ReplyTighten.TopicSentences = sentences
}
