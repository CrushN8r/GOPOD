// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_highlight_corpus_test.go
package wirepod_ttr

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The highlight check on every real reply (operator 2026-10-08: "how to set up
// a correct proper mechanical pipeline process to generate more consistent
// predictable * stop this lazy patch work approach"). The golden test holds a
// few hand-picked replies; this one takes every Brobot reply this machine has
// kept - the chat memory (sessions/) and every draft's answers - and puts each
// through the live settings' split, the bubble cut and the piece numbers, the
// same path a reply takes live. deploy_overlay_build.sh runs it before a build,
// so a change that lights the wrong words never reaches the robots. The replies
// are private run files, read in place; with none on the machine, it's skipped.
func TestHighlightCorpus(t *testing.T) {
	replies := gopodHighlightCorpus()
	if len(replies) == 0 {
		t.Skip("no kept replies on this machine")
	}
	marker := gopodBrobotsSettings.ScriptedSplit.SegmentMarker
	bad := 0
	for _, raw := range replies {
		if problem := gopodHighlightProblem(raw, marker); problem != "" {
			bad++
			if bad <= 15 {
				t.Logf("%s\n  reply: %q", problem, raw)
			}
		}
	}
	// Words the robot says respelled ("Gowp-awd" for GOPOD) can't always be
	// lined up with the bubble yet. A ratchet: no more misses than the
	// choppy split had on 2026-10-08 (28 of 838, the operator's "choppy and
	// working"), so a change can only make it better.
	// 2026-10-13: respelled words checked as their bubble words - 2 left.
	const knownMisses = 2
	if bad > knownMisses {
		t.Errorf("%d of %d kept replies light the wrong words (more than the %d known)", bad, len(replies), knownMisses)
	}
	t.Logf("HIGHLIGHT_CORPUS %d replies checked, %d wrong", len(replies), bad)
}

// gopodHighlightProblem: what's wrong with this reply's highlight, or "".
func gopodHighlightProblem(raw, marker string) string {
	rawInternal := gopodAnimationMarkerToInternal(gopodFixSelfName(gopodBrobot1ESN, gopodCorrectNames(gopodStripLeadingAnimationWords(raw))))
	display := gopodDisplayReply(raw)
	pieces := gopodSpokenPieceTexts(buildGOPODStreamResult(rawInternal).AudibleChunks)
	if len(pieces) == 0 {
		return ""
	}
	marked, segOf := gopodSegmentDisplay(display, pieces)
	shown := strings.Split(marked, marker)
	if strings.Join(strings.Fields(strings.ReplaceAll(marked, marker, "")), " ") != strings.Join(strings.Fields(display), " ") {
		return "the bubble's words changed"
	}
	if segOf[0] != 0 || segOf[len(segOf)-1] != len(shown)-1 {
		return "the pieces don't run from the bubble's first piece to its last"
	}
	for i := 1; i < len(segOf); i++ {
		if segOf[i] < segOf[i-1] || segOf[i] > segOf[i-1]+1 {
			return "the piece numbers go back or skip"
		}
	}
	for i, p := range pieces {
		// 2026-10-13: a respelled word ("ay-emm") is checked as its bubble
		// word ("AM").
		for _, e := range loadedGOPODPronunciations {
			if e.To != "" && e.From != "" {
				p = strings.ReplaceAll(p, e.To, e.From)
			}
		}
		// 2026-10-15: a symbol said in words (spoken_symbols: "divided by",
		// "equals", "over", "repeating" - the math) is checked as its symbol,
		// and a repeating decimal's bars don't split its number.
		ss := gopodBrobotsSettings.SpokenSymbols
		for sym, said := range ss.Words {
			if strings.TrimSpace(said) != "" {
				p = strings.ReplaceAll(p, " "+said+" ", " "+sym+" ")
			}
		}
		if ss.RepeatingWord != "" {
			p = strings.ReplaceAll(p, " "+ss.RepeatingWord, "")
		}
		// 2026-10-17: a curly apostrophe in the bubble is checked as the spoken
		// straight one (gopodHighlightWords already makes them one).
		lit := strings.ToLower(strings.NewReplacer("\u0305", "", "\u0304", "", "\u2019", "'", "\u2018", "'").Replace(shown[segOf[i]]))
		words := gopodHighlightWords(p)
		in := 0
		for _, w := range words {
			if strings.Contains(lit, w.word) {
				in++
			}
		}
		if len(words) >= 3 && in*2 < len(words) {
			return "said " + strings.TrimSpace(p) + " | lit " + strings.TrimSpace(shown[segOf[i]])
		}
	}
	return ""
}

// gopodHighlightCorpus: every kept reply - the chat memory's raw replies and
// every draft's answers, each once.
func gopodHighlightCorpus() []string {
	seen := map[string]bool{}
	var out []string
	add := func(raw string) {
		raw = strings.TrimSpace(raw)
		if raw != "" && !seen[raw] {
			seen[raw] = true
			out = append(out, raw)
		}
	}
	sessions := filepath.Join(gopodChipperRoot(), "sessions")
	files, _ := filepath.Glob(filepath.Join(sessions, "archives", "*.jsonl"))
	files = append(files, filepath.Join(sessions, "current_session.jsonl"))
	for _, f := range files {
		fh, err := os.Open(f)
		if err != nil {
			continue
		}
		scan := bufio.NewScanner(fh)
		scan.Buffer(make([]byte, 1<<20), 1<<20)
		for scan.Scan() {
			var e struct {
				Raw string `json:"raw_response"`
			}
			if json.Unmarshal(scan.Bytes(), &e) == nil {
				add(e.Raw)
			}
		}
		fh.Close()
	}
	home, _ := os.UserHomeDir()
	drafts, _ := filepath.Glob(filepath.Join(home, "gopod_tts", "sessions", "harvest", "*", "*", "draft_performance.json"))
	for _, f := range drafts {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var d gopodDraft
		if json.Unmarshal(data, &d) != nil {
			continue
		}
		for _, q := range d.Questions {
			for _, a := range q.Answers {
				if strings.Contains(a.Say, "{weather}") {
					continue
				}
				if a.Cue != "" {
					add(a.Cue + "___ " + a.Say)
				} else {
					add(a.Say)
				}
			}
		}
	}
	return out
}
