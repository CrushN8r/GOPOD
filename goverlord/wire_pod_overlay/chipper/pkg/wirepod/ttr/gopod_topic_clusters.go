// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_topic_clusters.go
package wirepod_ttr

// Topic-clusters for settings (2026-10-31, operator: "What else now, can be
// topic-clustered? * Topic-clusters please * Easter eggs * Golden answers * The
// Butler's lines * Nicknames"). A cluster is one md file in the master format
// (gopod_probes/clusters/*.md): "# <title>", its head, "---", then sections. Each
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
//
// The cluster is the one source: its values are laid over the settings at
// Wire-Pod's start and for the chat page (gopodClustersOverlay), and the
// settings files no longer hold them. system-check names every point that
// lands nowhere (gopodClustersCheck).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// gopodClustersDefaultFolder: where the clusters live (topic_clusters.folder overrides).
const gopodClustersDefaultFolder = "gopod_probes/clusters"

// gopodClusterPairs: a list of two-field objects a point fills - its fields, left
// of the first ": " and right of it (the left one a list when it's "queries").
var gopodClusterPairs = map[string][2]string{
	"value_points.reminder_eggs": {"key", "label"},
	"golden_answers":             {"queries", "line"},
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
	"system_help.lines": true,
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

// gopodClusterSections: every section of every cluster, in file then section order.
func gopodClusterSections(folder string) []gopodClusterSection {
	files, _ := filepath.Glob(filepath.Join(folder, "*.md"))
	sort.Strings(files)
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

// gopodClustersApply lays the clusters over the merged settings; the problems
// are the points that land nowhere.
func gopodClustersApply(data []byte) ([]byte, []string) {
	var all map[string]json.RawMessage
	if json.Unmarshal(data, &all) != nil {
		return data, nil
	}
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
		if s.Path == "" {
			bad(s, "no settings place in brackets")
			continue
		}
		parts := strings.Split(s.Path, ".")
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
		// The parent of the place, and the place's key in it.
		var parent map[string]any
		var key string
		node := blocks[top]
		if len(parts) > 1 {
			m, ok := node.(map[string]any)
			for _, p := range parts[1 : len(parts)-1] {
				if !ok {
					break
				}
				next, has := m[p].(map[string]any)
				if !has {
					next = map[string]any{}
					m[p] = next
				}
				m = next
			}
			if !ok {
				bad(s, s.Path+" isn't a group")
				continue
			}
			parent, key = m, parts[len(parts)-1]
			node = m[key]
		}
		if gopodClusterLines[s.Path] {
			var list []any
			for _, pt := range s.Points {
				list = append(list, pt)
			}
			if parent != nil {
				parent[key] = list
			} else {
				blocks[top] = list
			}
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
			if parent != nil {
				parent[key] = list
			} else {
				blocks[top] = list
			}
			continue
		}
		m, ok := node.(map[string]any)
		if !ok {
			m = map[string]any{}
			if node != nil {
				bad(s, s.Path+" isn't a group")
				continue
			}
			if parent != nil {
				parent[key] = m
			} else {
				blocks[top] = m
			}
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
			if old, has := m[k]; has {
				if _, isText := old.(string); !isText {
					bad(s, k+" isn't a text or a list")
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
		n += len(s.Points)
		if len(s.Points) == 0 {
			problems = append(problems, "cluster "+s.File+" \""+s.Title+"\": no points")
		}
	}
	return problems, n
}

// ## EOF
