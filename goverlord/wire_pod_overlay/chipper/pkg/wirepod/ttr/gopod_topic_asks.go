// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_topic_asks.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var gopodLastTopics = struct {
	sync.Mutex
	names []string
}{}

// gopodTopicAskKind: "detail" (a detail word), "summary" (a summary word, or a bare topic of up to
// bare_max_words words - "Mobile gear?"), else "point" - a pointed question gets the value point it
// asks about, as before.
func gopodTopicAskKind(question string) string {
	ta := gopodBrobotsSettings.TopicAsks
	lower := strings.ToLower(question)
	for _, w := range ta.DetailWords {
		if w = strings.ToLower(strings.TrimSpace(w)); w != "" && gopodHasWholeWords(lower, w) {
			return "detail"
		}
	}
	for _, w := range ta.SummaryWords {
		if w = strings.ToLower(strings.TrimSpace(w)); w != "" && gopodHasWholeWords(lower, w) {
			return "summary"
		}
	}
	if words := strings.Fields(gopodAddressStrip(lower)); len(words) > 0 && len(words) <= ta.BareMaxWords {
		return "summary"
	}
	if ta.TheTopic {
		for _, p := range gopodTopicRegistry() {
			for _, w := range p.Triggers {
				if w = strings.ToLower(strings.TrimSpace(w)); w != "" && gopodHasWholeWords(lower, "the "+w) {
					return "summary"
				}
			}
		}
	}
	return "point"
}

// gopodSummaryGiven: the points a summary ask scored for each Brobot - counted, never said word for
// word (he riffs on the summary instead).
var gopodSummaryGiven = struct {
	sync.Mutex
	points map[string]map[string]bool
}{points: map[string]map[string]bool{}}

func gopodSummaryGivenSet(esn string, points []string) {
	gopodSummaryGiven.Lock()
	defer gopodSummaryGiven.Unlock()
	m := map[string]bool{}
	for _, p := range points {
		m[p] = true
	}
	gopodSummaryGiven.points[esn] = m
}

func gopodSummaryGivenHas(esn, point string) bool {
	gopodSummaryGiven.Lock()
	defer gopodSummaryGiven.Unlock()
	return gopodSummaryGiven.points[esn][point]
}

// gopodAddressStrip: a question without a leading "Brobot 1," / "Brobot 2,".
func gopodAddressStrip(lower string) string {
	lower = strings.TrimSpace(lower)
	for _, b := range []string{"brobot 1", "brobot 2", "brobots"} {
		if strings.HasPrefix(lower, b) {
			return strings.TrimLeft(strings.TrimPrefix(lower, b), " ,:-")
		}
	}
	return lower
}

// gopodTopicFileByName: a topic's file (its summary and source), read now.
func gopodTopicFileByName(name string) (gopodTopicFile, bool) {
	chapters, always, _ := gopodReadChaptersFolder()
	for _, t := range always {
		if t.Name == name {
			return t, true
		}
	}
	for _, c := range chapters {
		for _, t := range c.Topics {
			if t.Name == name {
				return t, true
			}
		}
	}
	return gopodTopicFile{}, false
}

// gopodTopicPackets: the topics a question loads; a detail ask that loads none is about the last
// topics asked.
func gopodTopicPackets(question string) []loadedScenarioPacket {
	packets := loadScenarioPacketsForMessage(question)
	gopodLastTopics.Lock()
	defer gopodLastTopics.Unlock()
	if len(packets) == 0 && gopodTopicAskKind(question) == "detail" {
		for _, name := range gopodLastTopics.names {
			if t, ok := gopodTopicFileByName(name); ok {
				if content, err := gopodReadTopicFacts(t.Path); err == nil {
					packets = append(packets, loadedScenarioPacket{Name: t.Name, File: t.Path, Content: content})
				}
			}
		}
		return packets
	}
	if len(packets) > 0 {
		gopodLastTopics.names = nil
		for _, p := range packets {
			gopodLastTopics.names = append(gopodLastTopics.names, p.Name)
		}
	}
	return packets
}

// gopodTopicSummary: for a summary ask, the topic's summary - its notice line and the facts the
// Brobot riffs on.
func gopodTopicSummary(name, question string) (notice, facts string, ok bool) {
	if gopodTopicAskKind(question) != "summary" {
		return "", "", false
	}
	t, found := gopodTopicFileByName(name)
	if !found || strings.TrimSpace(t.Summary) == "" {
		return "", "", false
	}
	label := strings.TrimSpace(t.Label)
	if label == "" {
		label = name
	}
	ta := gopodBrobotsSettings.TopicAsks
	values := map[string]string{"topic": label, "summary": t.Summary}
	return gopodNoticeFill(ta.SummaryNotice, values), gopodNoticeFill(ta.SummaryFacts, values), true
}

// gopodTopicSourceProblems: topics whose source: file is missing from the repo, or changed after
// the topic was (the game no longer matches the repo).
func gopodTopicSourceProblems() []string {
	repo := gopodRepoRoot()
	chapters, always, _ := gopodReadChaptersFolder()
	all := append([]gopodTopicFile{}, always...)
	for _, c := range chapters {
		all = append(all, c.Topics...)
	}
	var out []string
	for _, t := range all {
		for _, src := range t.Sources {
			info, err := os.Stat(filepath.Join(repo, src))
			if err != nil {
				out = append(out, t.Name+": its source "+src+" isn't in the repo")
				continue
			}
			if ti, err := os.Stat(t.Path); err == nil && info.ModTime().After(ti.ModTime()) {
				out = append(out, t.Name+": "+src+" changed after the topic - check it still matches")
			}
		}
	}
	return out
}

// gopodAsksPage: BROBOT_ASKS.md at the repo root, made from the chapters' topics - the three kinds
// of ask, then each topic's ask, its summary and the repo file it tells.
func gopodAsksPage() string {
	chapters, _, _ := gopodReadChaptersFolder()
	ta := gopodBrobotsSettings.TopicAsks
	asked := map[string][]string{}
	for _, q := range gopodBrobotsSettings.GameQuestions {
		asked[q.Topic] = append(asked[q.Topic], "\""+gopodAskWithoutAddress(q.Ask)+"\"")
	}
	probed := map[string][]string{}
	if probers, _, err := gopodProbers(); err == nil {
		for _, p := range probers {
			if len(p.want) > 0 {
				probed[p.want[0]] = append(probed[p.want[0]], "\""+p.text+"\"")
			}
		}
	}
	quote := func(words []string, n int) string {
		var out []string
		for i, w := range words {
			if i == n {
				break
			}
			out = append(out, "\""+w+"\"")
		}
		return strings.Join(out, ", ")
	}
	var b strings.Builder
	b.WriteString("# Brobot Asks\n\n")
	b.WriteString("The basic asks Brobot 1 and Brobot 2 know how to answer - type them in the chat, or hold a numpad key and say them. ")
	b.WriteString("This page is made from the game's own topics - the folded game cards in the repo docs among them - so it changes when they do.\n\n")
	b.WriteString("## Three ways to ask\n\n")
	b.WriteString("| Ask | Example | You get |\n|---|---|---|\n")
	b.WriteString("| A topic on its own, or " + quote(ta.SummaryWords, 1) + " a topic | \"Mobile gear?\" · \"Tell me about the rig\" | The summary, riffed on - it shows in the Brobot's notice bubble. |\n")
	b.WriteString("| " + quote(ta.DetailWords, 3) + " | \"Tell me all about the rig\" · \"Tell me more\" | The details, one value point at a time. \"Tell me more\" stays on the last topic. |\n")
	b.WriteString("| A pointed question | \"What's in the upper crate?\" | The one value point it asks about. |\n\n")
	b.WriteString("Every answer scores in the game: topics, value points and easter eggs, shown in the notice bubble.\n\n")
	b.WriteString("**The probers** are Ron's own questions from real sessions - examples of how he thought, kept in the probers list (`goverlord/wire_pod_overlay/chipper/gopod_probes/probers_list/`, checked by `system-test-probers-list`).\n\n")
	b.WriteString("## The topics\n")
	for _, c := range chapters {
		var rows []string
		for _, t := range c.Topics {
			if strings.TrimSpace(t.Summary) == "" {
				continue
			}
			ask := strings.TrimSpace(t.Ask)
			if ask == "" {
				ask = "Tell me about " + t.Label
			}
			var docs []string
			for _, src := range t.Sources {
				docs = append(docs, "["+filepath.Base(src)+"]("+src+")")
			}
			rows = append(rows, "| \""+ask+"\" | "+strings.ReplaceAll(t.Summary, " | ", " · ")+" | "+strings.Join(asked[t.Name], " · ")+" | "+strings.Join(probed[t.Name], " · ")+" | "+strings.Join(docs, " ")+" |")
		}
		if len(rows) == 0 {
			continue
		}
		b.WriteString("\n### " + c.Title + "\n\n| Ask | What you get | Also ask | The probers | More in |\n|---|---|---|---|---|\n")
		b.WriteString(strings.Join(rows, "\n") + "\n")
	}
	b.WriteString("\n---\n\nGOPOD — [README](README.md) · [mobile gear](MOBILE_GEAR.md) · [WIRED-POD](WIRED-POD.md)\n")
	return b.String()
}

// gopodAskWithoutAddress: a game question without its "Brobot 2, " - the ask itself.
func gopodAskWithoutAddress(q string) string {
	q = strings.TrimSpace(q)
	for _, b := range []string{"Brobot 1, ", "Brobot 2, ", "Brobots, "} {
		if strings.HasPrefix(q, b) {
			q = strings.TrimPrefix(q, b)
			if q != "" {
				q = strings.ToUpper(q[:1]) + q[1:]
			}
		}
	}
	return q
}

// gopodTopicTouched: a value point of the topic has been given this session.
func gopodTopicTouched(name string) bool {
	gopodCoverage.Lock()
	defer gopodCoverage.Unlock()
	return gopodCoverage.revealed[name] > 0
}

// gopodTopicPrePost: the topic's note for this answer - pre or post.
func gopodTopicPrePost(name string, covered bool) string {
	t, ok := gopodTopicFileByName(name)
	if !ok {
		return ""
	}
	label := strings.TrimSpace(t.Label)
	if label == "" {
		label = name
	}
	ta := gopodBrobotsSettings.TopicAsks
	note := t.Pre
	if note == "" {
		note = ta.PreDefault
	}
	if covered {
		note = t.Post
		if note == "" {
			note = ta.PostDefault
		}
	}
	return strings.TrimSpace(strings.ReplaceAll(note, "{topic}", label))
}

// gopodTopicClosing: a covered topic's closing remark for a question about it.
func gopodTopicClosing(question string) string {
	for _, p := range loadScenarioPacketsForMessage(question) {
		if t, ok := gopodTopicFileByName(p.Name); ok && strings.TrimSpace(t.Closing) != "" && gopodTopicTouched(p.Name) {
			return strings.TrimSpace(t.Closing)
		}
	}
	return ""
}

// gopodRepoRoot: the GOPOD repo, three folders above the overlay's chipper.
func gopodRepoRoot() string {
	return filepath.Clean(filepath.Join(gopodChipperRoot(), "..", "..", ".."))
}

// ## EOF
