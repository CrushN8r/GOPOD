// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_topic_clusters.go
package wirepod_ttr

// Topic-clusters for settings (2026-10-31, operator: "What else now, can be
// topic-clustered? * Topic-clusters please * Easter eggs * Golden answers * The
// Butler's lines * Nicknames"). A cluster is one md file in the master format
// (gopod_probes/topic_clusters/show/*.md): "# <title>", its head, "---", then sections. Each
// section's heading ends with the settings place it fills, in brackets -
// "## Run lines (session)" - and each point is "key: value":
//
//	a text        "started_text: {run} started. Right away, sir."
//	a list        "eggs_keep: photo | math"           (split on " | ")
//	a name map    "camera_feeds: Camera Feeds"        (one entry each)
//	a pair list   "math: Explain the Math"            (reminder_eggs: key, label)
//	              "what is ai wordplay | what's ai wordplay: AI wordplay: ..."
//	                                                  (golden_answers: queries, line)
//	a text place  "page_here:   chat page: here"      (gopodClusterTexts: " | " is
//	                                                  words, "\n" a break, spaces kept)
//	a line list   "3.   system-add ..."               (gopodClusterLines: one line each)
//	an exact text "cast_me: ` (me)`"                  (in backticks: spaces kept,
//	                                                  " | " words, "\n" a break)
//
// The cluster is the one source: its values are laid over the settings at
// Wire-Pod's start and for the chat page (gopodClustersOverlay), and the
// settings files no longer hold them. system-check names every point that
// lands nowhere (gopodClustersCheck).

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// gopodClustersDefaultFolder: where the clusters live (topic_clusters.folder overrides).
const gopodClustersDefaultFolder = "gopod_probes/topic_clusters"

// gopodClusterPairs: a list of two-field objects a point fills - its fields, left
// of the first ": " and right of it (the left one a list when it's "queries").
var gopodClusterPairs = map[string][2]string{
	"value_points.reminder_eggs": {"key", "label"},
	"golden_answers":             {"queries", "line"},
	// 2026-11-03 (step 5): the pre-show's bookends and the sign-off, a line each.
	"reporter_voices.opening": {"host", "text"},
	"reporter_voices.closing": {"host", "text"},
	"sign_off.lines":          {"who", "text"},
	// 2026-11-03 (step 7): the cast the here-and-now line names, "who: line".
	"here_now.cast": {"who", "line"},
	// 2026-11-03 (step 7b): the brand words, "from: to".
	"reply_tighten.brand_words": {"from", "to"},
}

// 2026-11-03 (operator: "the whole GOPOD system can be that ... let's do that as
// a planned group of steps"; step 1, the shell window's words):
// gopodClusterTexts: a place whose points are always texts - " | " in one is
// its words (system-help's "<brobots | reporters>"), never a list.
var gopodClusterTexts = map[string]bool{
	"system_text": true,
}

// gopodClusterLines: a list of lines a section fills - each point one line, as
// written after its number (its leading spaces kept: the help's columns).
var gopodClusterLines = map[string]bool{
	"system_help.lines":      true,
	"brain_chain.lines":      true, // step 3: the brain's template, a line each
	"brain_chain.intro_lines": true, // 2026-11-13: the comm check's (and an intro's) lines
	"prebuild_context.lines": true,
}

type gopodClusterSection struct {
	File, Title, Path string
	Points            []string
}

// gopodClustersFolder: the clusters' folder, from the merged settings' topic_clusters.folder.
func gopodClustersFolder(all map[string]json.RawMessage) string {
	var tc struct {
		Folder string `json:"folder"`
	}
	_ = json.Unmarshal(all["topic_clusters"], &tc)
	folder := strings.TrimSpace(tc.Folder)
	if folder == "" {
		folder = gopodClustersDefaultFolder
	}
	if !filepath.IsAbs(folder) {
		folder = filepath.Join(gopodChipperRoot(), folder)
	}
	return folder
}

// gopodTopicRoot (2026-11-04, operator: "make the code search the whole
// topic_clusters/ tree"): topic_clusters.folder - every topic-cluster file's one folder.
func gopodTopicRoot() string {
	f, _ := gopodTopicRootSetting.Load().(string)
	if f = strings.TrimSpace(f); f == "" {
		f = gopodClustersDefaultFolder
	}
	if !filepath.IsAbs(f) {
		f = filepath.Join(gopodChipperRoot(), f)
	}
	return f
}

// gopodTopicRootSetting: topic_clusters.folder, kept once the settings are loaded
// (the chapters are read while they load - the default folder until then).
var gopodTopicRootSetting atomic.Value

// gopodTopicRootDocsSetting: topic_clusters.root_docs, kept the same way.
var gopodTopicRootDocsSetting atomic.Value

func init() {
	gopodTopicRootSetting.Store(gopodBrobotsSettings.TopicClusters.Folder)
	gopodTopicRootDocsSetting.Store(gopodBrobotsSettings.TopicClusters.RootDocs)
}

// gopodTopicIndex: every file and folder under topic_clusters/ by name, rebuilt
// at most every few seconds (files can be moved while Wire-Pod runs).
var gopodTopicIndex = struct {
	sync.Mutex
	at     time.Time
	root   string
	byName map[string]string
}{}

// gopodTopicTreeFind (operator: "a persona found by its file name wherever it sits";
// "DocBS pools found by folder name"): a file or folder by its name anywhere
// under topic_clusters/ - the first in path order; "" when there's none. A "_"
// folder (the template) is found itself, not looked into.
func gopodTopicTreeFind(name string) string {
	root := gopodTopicRoot()
	gopodTopicIndex.Lock()
	defer gopodTopicIndex.Unlock()
	if gopodTopicIndex.byName == nil || gopodTopicIndex.root != root || time.Since(gopodTopicIndex.at) > 5*time.Second {
		idx := map[string]string{}
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || path == root {
				return nil
			}
			if _, seen := idx[strings.ToLower(d.Name())]; !seen {
				idx[strings.ToLower(d.Name())] = path
			}
			if d.IsDir() && strings.HasPrefix(d.Name(), "_") {
				return fs.SkipDir
			}
			return nil
		})
		gopodTopicIndex.byName, gopodTopicIndex.root, gopodTopicIndex.at = idx, root, time.Now()
	}
	return gopodTopicIndex.byName[strings.ToLower(strings.TrimSpace(name))]
}

// gopodTopicPath: a settings path where it is; moved, found by its last name
// under topic_clusters/; neither, the path as written.
func gopodTopicPath(rel string) string {
	p := strings.TrimSpace(rel)
	if p == "" {
		return ""
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(gopodChipperRoot(), p)
	}
	if _, err := os.Stat(p); err == nil {
		return p
	}
	if f := gopodTopicTreeFind(filepath.Base(p)); f != "" {
		return f
	}
	return p
}

// gopodClusterFiles (2026-11-04, operator: "Build clusters/gopod-system/ as the
// one gopod-system topic clusters source folder"): every cluster md in folder and
// its sub-folders (clusters/gopod-system/), in path order.
func gopodClusterFiles(folder string) []string {
	var files []string
	_ = filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
		// A folder whose name starts with "_" (the _topic_cluster template) isn't read.
		if err == nil && d.IsDir() && strings.HasPrefix(d.Name(), "_") && path != folder {
			return fs.SkipDir
		}
		if err == nil && !d.IsDir() && strings.HasSuffix(path, ".md") && !strings.EqualFold(d.Name(), "README.md") {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	return files
}

// gopodClusterPathRE: a settings place - "." or a lower-case name, then its parts.
var gopodClusterPathRE = regexp.MustCompile(`^(\.|[a-z][a-z0-9_]*(\.[^()]+)?)(\[\])?$`)

// gopodClusterSections: every section of every cluster, in file then section order.
func gopodClusterSections(folder string) []gopodClusterSection {
	files := gopodClusterFiles(folder)
	var out []gopodClusterSection
	for _, f := range files {
		if strings.EqualFold(filepath.Base(f), "README.md") {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		inBody := false
		var cur *gopodClusterSection
		for _, line := range strings.Split(string(data), "\n") {
			t := strings.TrimSpace(line)
			switch {
			case t == "---":
				inBody = true
			case !inBody:
			case strings.HasPrefix(t, "## "):
				head := strings.TrimSpace(t[3:])
				path := ""
				if open, close := strings.LastIndex(head, "("), strings.LastIndex(head, ")"); open >= 0 && close > open {
					path = strings.TrimSpace(head[open+1 : close])
				}
				// 2026-11-04 (the whole topic_clusters/ tree is read): brackets are a
				// settings place only when they look like one - "." or a name like
				// docbs.chains.0 - never ordinary words ("(mandatory order)", a date).
				if path != "" && !gopodClusterPathRE.MatchString(path) {
					path = ""
				}
				out = append(out, gopodClusterSection{File: filepath.Base(f), Title: head, Path: path})
				cur = &out[len(out)-1]
			case cur != nil && t != "":
				if dot := strings.Index(t, ". "); dot > 0 && strings.Trim(t[:dot], "0123456789") == "" {
					// Leading spaces after the number stay (gopodClusterLines).
					cur.Points = append(cur.Points, strings.TrimRight(strings.TrimLeft(line, " \t")[dot+2:], " \t\r"))
				}
			}
		}
	}
	return out
}

// gopodProfileSettingsSections: the persona profiles' sections that fill a
// settings place ("## My settings (managers.personas.CHALK)"); the profile's own
// sections (Self, Voice, ...) have no place and are left to it.
func gopodProfileSettingsSections(all map[string]json.RawMessage) []gopodClusterSection {
	var pp struct {
		Folder string `json:"folder"`
	}
	_ = json.Unmarshal(all["persona_profiles"], &pp)
	folder := strings.TrimSpace(pp.Folder)
	if folder == "" {
		return nil
	}
	if !filepath.IsAbs(folder) {
		folder = filepath.Join(gopodChipperRoot(), folder)
	}
	var out []gopodClusterSection
	for _, s := range gopodClusterSections(folder) {
		if s.Path != "" {
			out = append(out, s)
		}
	}
	return out
}

// gopodClusterHolds (2026-11-03, the clusters plan: system-set wrote the settings
// file while a cluster's value won, so an edit did nothing): the cluster file that
// fills dotted, or a part of it ("" none).
func gopodClusterHolds(dotted string) string {
	data, err := gopodSettingsMerged()
	if err != nil {
		return ""
	}
	var all map[string]json.RawMessage
	if json.Unmarshal(data, &all) != nil {
		return ""
	}
	dotted = strings.Trim(dotted, ".")
	sections := gopodClusterSections(gopodClustersFolder(all))
	for _, s := range sections {
		if s.Path == "" {
			continue
		}
		if dotted == s.Path || strings.HasPrefix(s.Path, dotted+".") {
			return s.File
		}
		if !strings.HasPrefix(dotted, s.Path+".") {
			continue
		}
		if gopodClusterLines[s.Path] {
			return s.File
		}
		rest := dotted[len(s.Path)+1:]
		for _, pt := range s.Points {
			if at := strings.Index(pt, ": "); at > 0 {
				k := strings.TrimSpace(pt[:at])
				if rest == k || strings.HasPrefix(rest, k+".") {
					return s.File
				}
			}
		}
	}
	return ""
}

// gopodClusterTyped: a bare number, true, false, null, [] (an empty list) or {} (an empty group), as its value.
func gopodClusterTyped(v string) (any, bool) {
	switch v {
	case "true":
		return true, true
	case "false":
		return false, true
	case "null":
		return nil, true
	case "[]":
		return []any{}, true // an empty list
	case "{}":
		return map[string]any{}, true // an empty group
	}
	if !gopodClusterNumberRE.MatchString(v) {
		return nil, false
	}
	var n any
	if json.Unmarshal([]byte(v), &n) != nil {
		return nil, false
	}
	return n, true
}

var gopodClusterNumberRE = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?$`)

// gopodClusterPlace: the node at parts under blocks[top], made as it goes - a
// group for a name, a list for a number (the next entry, or one already there) -
// and the setter that puts a new node in its place.
func gopodClusterPlace(blocks map[string]any, top string, parts []string) (any, func(any), string) {
	node := blocks[top]
	set := func(v any) { blocks[top] = v }
	for i, p := range parts {
		// A number is a list's entry - unless the place is already a group, where
		// it's a key like any other (the numpad's "000" lane).
		_, inGroup := node.(map[string]any)
		if n, err := strconv.Atoi(p); err == nil && n >= 0 && !inGroup {
			list, ok := node.([]any)
			if node != nil && !ok {
				return nil, nil, strings.Join(parts[:i], ".") + " isn't a list"
			}
			if n > len(list) {
				return nil, nil, "list entry " + p + " comes before entry " + strconv.Itoa(len(list))
			}
			if n == len(list) {
				list = append(list, nil)
				set(list)
			}
			parentSet, idx, l := set, n, list
			node = l[idx]
			set = func(v any) { l[idx] = v; parentSet(l) }
			continue
		}
		m, ok := node.(map[string]any)
		if !ok {
			if node != nil {
				return nil, nil, strings.Join(parts[:i], ".") + " isn't a group"
			}
			m = map[string]any{}
			set(m)
		}
		key, mm := p, m
		node = m[p]
		set = func(v any) { mm[key] = v }
	}
	return node, set, ""
}

// gopodWordsLeft (2026-11-03, step 8 of GOPOD_CLUSTERS_PLAN_002.md, operator: "the
// whole GOPOD system can be that"): the texts still in the settings files, not in
// a topic-cluster - every text of three words or more (a note's about/why aside),
// but those topic_clusters.words_left_not starts or words_left_not_words name (a
// path, a colour, a sound chain). system-check names them; more than
// topic_clusters.words_left_allowed is a problem, so the number only comes down.
func gopodWordsLeft() (left []string, allowed int) {
	data, err := gopodSettingsMerged()
	if err != nil {
		return nil, 0
	}
	var all map[string]any
	if json.Unmarshal(data, &all) != nil {
		return nil, 0
	}
	var cfg struct {
		Allowed  int      `json:"words_left_allowed"`
		NotStart []string `json:"words_left_not"`
		NotWords []string `json:"words_left_not_words"`
		NotPaths []string `json:"words_left_not_paths"`
	}
	if raw, err := json.Marshal(all["topic_clusters"]); err == nil {
		_ = json.Unmarshal(raw, &cfg)
	}
	isWords := func(v string) bool {
		f := strings.Fields(v)
		if len(f) < 3 || strings.IndexFunc(v, func(r rune) bool { return r >= 'a' && r <= 'z' }) < 0 {
			return false
		}
		for _, p := range cfg.NotStart {
			if strings.HasPrefix(v, p) {
				return false
			}
		}
		for _, w := range f {
			for _, n := range cfg.NotWords {
				if w == n {
					return false
				}
			}
		}
		return true
	}
	var walk func(path string, v any)
	walk = func(path string, v any) {
		switch x := v.(type) {
		case map[string]any:
			keys := make([]string, 0, len(x))
			for k := range x {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				// A note is no spoken line: about, why, where ("intro_recipe.where"), about_2.
				if k == "about" || k == "why" || k == "where" || strings.HasSuffix(k, "_about") || strings.HasSuffix(k, "_where") || strings.HasPrefix(k, "about_") {
					continue
				}
				p := k
				if path != "" {
					p = path + "." + k
				}
				if gopodWordsLeftKept(cfg.NotPaths, p) {
					continue
				}
				walk(p, x[k])
			}
		case []any:
			for i, it := range x {
				walk(path+"."+strconv.Itoa(i), it)
			}
		case string:
			if isWords(x) {
				left = append(left, path)
			}
		}
	}
	walk("", all)
	return left, cfg.Allowed
}

// gopodWordsLeftKept: path is one the settings keep on purpose (words_left_not_paths), or inside one.
func gopodWordsLeftKept(kept []string, path string) bool {
	for _, k := range kept {
		if path == k || strings.HasPrefix(path, k+".") {
			return true
		}
	}
	return false
}

// gopodClustersApply lays the clusters over the merged settings; the problems
// are the points that land nowhere.
func gopodClustersApply(data []byte) ([]byte, []string) {
	var all map[string]json.RawMessage
	if json.Unmarshal(data, &all) != nil {
		return data, nil
	}
	// 2026-11-04 (operator: "make the code search the whole topic_clusters/ tree"):
	// every md under topic_clusters/ - the clusters, the persona files (their
	// "## My settings (...)" sections) and all - wherever it sits.
	sections := gopodClusterSections(gopodClustersFolder(all))
	if len(sections) == 0 {
		return data, nil
	}
	var problems []string
	bad := func(s gopodClusterSection, what string) {
		problems = append(problems, "cluster "+s.File+" \""+s.Title+"\": "+what)
	}
	blocks := map[string]any{}
	for _, s := range sections {
		// A section without a settings place is the file's own (a persona's Self,
		// a company note) - not a setting.
		if s.Path == "" {
			continue
		}
		// 2026-11-03 (step 8b): "(.)" is the settings' top level - each point a
		// top-level text, number or list ("off_topic_note: ...").
		if s.Path == "." {
			for _, pt := range s.Points {
				at := strings.Index(pt, ": ")
				if at <= 0 {
					bad(s, "\""+pt+"\" has no \": \"")
					continue
				}
				k, v := strings.TrimSpace(pt[:at]), strings.TrimSpace(pt[at+2:])
				switch {
				case len(v) >= 2 && strings.HasPrefix(v, "`") && strings.HasSuffix(v, "`"):
					blocks[k] = strings.ReplaceAll(v[1:len(v)-1], `\n`, "\n")
				case strings.Contains(v, " | ") || strings.HasSuffix(v, " |"):
					var items []any
					for _, it := range strings.Split(strings.TrimSuffix(v, " |"), " | ") {
						if it = strings.TrimSpace(it); it != "" {
							items = append(items, it)
						}
					}
					blocks[k] = items
				default:
					if typed, ok := gopodClusterTyped(v); ok {
						blocks[k] = typed
					} else {
						blocks[k] = v
					}
				}
			}
			continue
		}
		// 2026-11-03 (step 7, the format step): "(a.b[])" is a line list, a
		// point a line; a number in the place is a list's entry ("managers.takes.0").
		path, lines := s.Path, gopodClusterLines[s.Path]
		if strings.HasSuffix(path, "[]") {
			path, lines = strings.TrimSuffix(path, "[]"), true
		}
		parts := strings.Split(path, ".")
		top := parts[0]
		if _, ok := blocks[top]; !ok {
			var v any
			if raw, ok := all[top]; ok {
				_ = json.Unmarshal(raw, &v)
			}
			if v == nil {
				if pair, ok := gopodClusterPairs[s.Path]; ok && len(parts) == 1 && pair[0] != "" {
					v = []any{}
				} else {
					v = map[string]any{}
				}
			}
			blocks[top] = v
		}
		node, set, werr := gopodClusterPlace(blocks, top, parts[1:])
		if werr != "" {
			bad(s, werr)
			continue
		}
		if lines {
			var list []any
			for _, pt := range s.Points {
				list = append(list, pt)
			}
			set(list)
			continue
		}
		if pair, ok := gopodClusterPairs[s.Path]; ok {
			var list []any
			for _, pt := range s.Points {
				at := strings.Index(pt, ": ")
				if at <= 0 {
					bad(s, "\""+pt+"\" has no \": \"")
					continue
				}
				left, right := strings.TrimSpace(pt[:at]), strings.TrimSpace(pt[at+2:])
				var l any = left
				if pair[0] == "queries" {
					var qs []any
					for _, q := range strings.Split(left, " | ") {
						if q = strings.TrimSpace(q); q != "" {
							qs = append(qs, q)
						}
					}
					l = qs
				}
				list = append(list, map[string]any{pair[0]: l, pair[1]: right})
			}
			set(list)
			continue
		}
		m, ok := node.(map[string]any)
		if !ok {
			m = map[string]any{}
			if node != nil {
				bad(s, s.Path+" isn't a group")
				continue
			}
			set(m)
		}
		for _, pt := range s.Points {
			at := strings.Index(pt, ": ")
			if at <= 0 {
				bad(s, "\""+pt+"\" has no \": \"")
				continue
			}
			k, v := strings.TrimSpace(pt[:at]), strings.TrimSpace(pt[at+2:])
			// A text place keeps its leading spaces (the status lines' column) and
			// reads "\n" as a line break.
			if gopodClusterTexts[s.Path] {
				v = strings.ReplaceAll(pt[at+2:], `\n`, "\n")
			}
			// 2026-11-03 (step 3: the brain's " (me)" and " Reminders: ..."): a
			// value in backticks is that text exactly - its spaces, " | " as words,
			// "\n" a break.
			if len(v) >= 2 && strings.HasPrefix(v, "`") && strings.HasSuffix(v, "`") {
				m[k] = strings.ReplaceAll(v[1:len(v)-1], `\n`, "\n")
				continue
			}
			if old, isList := m[k].([]any); !gopodClusterTexts[s.Path] && (isList || (old == nil && (strings.Contains(v, " | ") || strings.HasSuffix(v, " |")))) {
				var items []any
				for _, it := range strings.Split(strings.TrimSuffix(v, " |"), " | ") {
					if it = strings.TrimSpace(it); it != "" {
						items = append(items, it)
					}
				}
				m[k] = items
				continue
			}
			// 2026-11-03 (step 7): a bare number, true, false or null is that value
			// (a dial, a switch) - unless the place already holds a text there.
			if _, wasText := m[k].(string); !wasText && !gopodClusterTexts[s.Path] {
				if typed, ok := gopodClusterTyped(v); ok {
					m[k] = typed
					continue
				}
			}
			if old, has := m[k]; has {
				switch old.(type) {
				case string, float64, bool:
				default:
					bad(s, k+" isn't a text, a number or a list")
					continue
				}
			}
			m[k] = v
		}
	}
	for top, v := range blocks {
		raw, err := json.Marshal(v)
		if err != nil {
			continue
		}
		all[top] = raw
	}
	out, err := json.Marshal(all)
	if err != nil {
		return data, problems
	}
	return out, problems
}

// gopodSettingsFull (2026-11-03, step 7: the Torah read the menu without its
// cluster and found none): the merged settings with the clusters laid over them -
// what Wire-Pod and the page run on. Anything reading the settings as data uses this.
func gopodSettingsFull() ([]byte, error) {
	data, err := gopodSettingsMerged()
	if err != nil {
		return nil, err
	}
	return gopodClustersOverlay(data), nil
}

// gopodSettingsSnapshotFile: the settings Wire-Pod runs on, written at its start
// (the settings folder with the topic-clusters laid over) - for the song engine
// (run_golden_song_001.py) and any tool that reads the settings itself, so none
// of them misses a line that lives in a topic-cluster. Generated, not tracked.
const gopodSettingsSnapshotFile = "webroot/gopod_settings_full.json"

// It's written at Wire-Pod's start, once the settings are loaded - not in a test run.
func init() {
	if strings.HasSuffix(os.Args[0], ".test") {
		return
	}
	if full, err := gopodSettingsFull(); err == nil {
		_ = gopodWriteFile(filepath.Join(gopodChipperRoot(), gopodSettingsSnapshotFile), gopodPersonaProfileLabelsJSON(full), 0o644)
	}
}

// gopodClustersOverlay: the merged settings with the clusters laid over them.
func gopodClustersOverlay(data []byte) []byte {
	out, _ := gopodClustersApply(data)
	return out
}

// gopodClustersCheck: system-check's problems with the clusters, and how many points.
func gopodClustersCheck() ([]string, int) {
	data, err := gopodSettingsMerged()
	if err != nil {
		return nil, 0
	}
	_, problems := gopodClustersApply(data)
	var all map[string]json.RawMessage
	_ = json.Unmarshal(data, &all)
	n := 0
	for _, s := range gopodClusterSections(gopodClustersFolder(all)) {
		if s.Path == "" {
			continue // a note section (a persona's Self, a reference doc), not a setting
		}
		n += len(s.Points)
		if len(s.Points) == 0 {
			problems = append(problems, "cluster "+s.File+" \""+s.Title+"\": no points")
		}
	}
	return problems, n
}

// ## EOF
