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
//	  _always/01_what_else.txt    topics that belong to no chapter
//	  01_warm_up/chapter.txt      "title: ..." and "summary: ..."
//	  01_warm_up/01_checks.txt    one topic file per topic, in order
//
// A folder's number sets the chapter order; a file's number sets the topic order
// (the number isn't part of the topic's name). A topic file is its settings, a
// "---" line, then the facts the Brobots get (the numbered value points):
//
//	label: the rig                    how a reminder names it
//	words: rig, gear, dolly           the words in a question that load it
//	pictures: gear/crate_2_display.jpg         shown in the chat with it
//	reminder: Hey, that reminds me... its reminder line
//	sentences: 0                      its own sentence budget (left out = the usual)
//	say 1: One dolly, two crates.     point 1's exact line, said first
//	start line: ...                   a scripted line at the start (one per line)
//	end line: ...                     a scripted line at the end (one per line)
//	summary: Lower crate 1 ... | ...  what a summary ask riffs on (gopod_topic_asks.go)
//	source: MOBILE_GEAR.md            the repo file it tells - the game matches the repo
//	ask: Tell me about yourself       its ask on BROBOT_ASKS.md (left out = "Tell me about <label>")
//	pre: ... / post: ...              its note before / after the room has covered it (topic_asks defaults)
//	closing: ...                      a closing remark once it's covered (none = none)
//	---
//	The rig: ... Reveal these, in order: 1. ... 2. ...
//
// The facts are read from the folder each time a question comes in; the rest is
// read once, when Wire-Pod starts.
var canonicalChaptersDir = gopodPathWithEnvOverride("GOPOD_CHAPTERS_DIR", "gopod_probes/chapters")

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
		if data, err := os.ReadFile(filepath.Join(dir, "chapter.txt")); err == nil {
			head, _ := gopodSplitTopicFile(string(data))
			for _, kv := range head {
				switch kv[0] {
				case "title":
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
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".txt") || e.Name() == "chapter.txt" {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		t := gopodTopicFile{
			Name:     gopodTopicNumberPrefix.ReplaceAllString(strings.TrimSuffix(e.Name(), ".txt"), ""),
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

// gopodSplitTopicFile splits a file into its "key: value" lines (before a "---"
// line) and the facts (after it). A file with no "---" is all settings.
func gopodSplitTopicFile(text string) (head [][2]string, body string) {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			body = strings.Join(lines[i+1:], "\n")
			break
		}
		if k, v, ok := strings.Cut(line, ":"); ok && strings.TrimSpace(k) != "" {
			head = append(head, [2]string{strings.ToLower(strings.TrimSpace(k)), strings.TrimSpace(v)})
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
