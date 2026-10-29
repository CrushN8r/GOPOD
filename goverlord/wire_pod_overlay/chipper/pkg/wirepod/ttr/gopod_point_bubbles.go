// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_point_bubbles.go
package wirepod_ttr

import (
	"math/rand"
	"strconv"
	"strings"
	"unicode"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// The point bubbles (2026-10-23, operator: "thought bubbles are not reiterating
// their respective value points * brobots to reiterate their value points brobot
// party style * I need a templated answer-chain mechanism * either the machine
// (game-mechanism) f*ck*ng works or it doesn't, regardless of the editable
// data"): the answer chain's point_bubbles part. A value point with notes ("title |
// note | note") becomes his thought bubbles, in order - a hype line for its title,
// then one party line per note. The model writes them (point_bubbles.prompt asks
// for exactly that); a line that's missing, or doesn't share a word with its note,
// is made from the note itself (point_bubbles.fallback). So the shape always
// holds, whatever the data. A line of his before them that's no note's (the
// weather comment) is kept, first.

// gopodPointBubblesNotes: the notes of the point given to this Brobot, if any.
func gopodPointBubblesNotes(esn string) (string, []string) {
	for _, given := range gopodGivenPoints(esn) {
		f := strings.Fields(given)
		if len(f) != 2 {
			continue
		}
		n, err := strconv.Atoi(strings.SplitN(f[1], "/", 2)[0])
		if err != nil {
			continue
		}
		if title, notes := gopodPointNotes(f[0], n); title != "" {
			return title, notes
		}
	}
	return "", nil
}

// gopodPointBubblesPrompt: what the model is asked for (point_bubbles.prompt).
func gopodPointBubblesPrompt(esn string) string {
	title, notes := gopodPointBubblesNotes(esn)
	pb := gopodBrobotsSettings.PointBubbles
	if title == "" || strings.TrimSpace(pb.Prompt) == "" {
		return ""
	}
	var numbered []string
	for i, n := range notes {
		numbered = append(numbered, strconv.Itoa(i+1)+") "+n)
	}
	return strings.NewReplacer("{title}", title, "{notes}", strings.Join(notes, "; "),
		"{count}", strconv.Itoa(len(notes)+1), "{numbered}", strings.Join(numbered, " ")).Replace(pb.Prompt)
}

// gopodShareWord: two texts share a telling word (gopodPickWords).
func gopodShareWord(a, b string) bool {
	wa := gopodPickWords(a)
	for w := range gopodPickWords(b) {
		if wa[w] {
			return true
		}
	}
	return false
}

// gopodCapFirst: "wild party" -> "Wild party".
func gopodCapFirst(s string) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) == 0 {
		return ""
	}
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// gopodPointBubbles shapes his answer (raw, the model's) into the point bubbles.
func gopodPointBubbles(esn, question, raw string) string {
	title, notes := gopodPointBubblesNotes(esn)
	if title == "" {
		return raw
	}
	pb := gopodBrobotsSettings.PointBubbles
	cue, body := "", raw
	if loc := gopodAnimationMarkerPattern.FindStringIndex(raw); loc != nil &&
		strings.TrimFunc(raw[:loc[0]], func(r rune) bool { return !unicode.IsLetter(r) }) == "" {
		cue, body = raw[:loc[1]], raw[loc[1]:]
	}
	var lines []string
	for _, l := range strings.Split(body, "\n") {
		for _, s := range gopodWholeSentences(l) {
			if s = strings.TrimSpace(s); s != "" {
				lines = append(lines, s)
			}
		}
	}
	used := make([]bool, len(lines))
	take := func(want string) string {
		for i, l := range lines {
			if !used[i] && gopodShareWord(l, want) {
				used[i] = true
				return l
			}
		}
		return ""
	}
	items := append([]string{title}, notes...)
	var out []string
	filled := 0
	first := len(lines)
	for k, item := range items {
		var line string
		if k == 0 {
			// The title line says the title itself (answer_shapes: the point's line).
			for i, l := range lines {
				if !used[i] && strings.Contains(strings.ToLower(l), strings.ToLower(title)) {
					used[i], line = true, l
					break
				}
			}
		} else {
			line = take(item)
		}
		if line == "" {
			filled++
			if k == 0 {
				line = strings.ReplaceAll(pb.Fallback.Title, "{title}", title)
				if hype := pb.Hype[gopodBrobotName(esn)]; len(hype) > 0 {
					line += " " + hype[rand.Intn(len(hype))]
				}
			} else {
				line = strings.NewReplacer("{note}", item, "{Note}", gopodCapFirst(item)).Replace(pb.Fallback.Note)
			}
		}
		out = append(out, strings.TrimSpace(line))
	}
	for i := range lines {
		if used[i] && i < first {
			first = i
		}
	}
	// His own line before the point's stays first - only the weather comment
	// (answer_shapes: a topic answer opens with its point).
	var lead []string
	weather := gopodHasWholeWords(strings.ToLower(question), "weather") || gopodNativeTakeIs(gopodBrobotName(esn), "weather", question)
	for i := 0; weather && i < first && i < len(lines) && len(lead) < pb.KeepLeadLines; i++ {
		if !used[i] {
			lead = append(lead, lines[i])
		}
	}
	logger.Println("BROBOT_POINT_BUBBLES " + gopodBrobotName(esn) + ": " + title + " - " + strconv.Itoa(len(items)-filled) + " of " + strconv.Itoa(len(items)) + " from the model, " + strconv.Itoa(filled) + " from the notes")
	all := append(lead, out...)
	result := strings.Join(all, gopodThoughtGap())
	if cue != "" {
		return strings.TrimRight(cue, " \t") + " " + result
	}
	return result
}

// ## EOF
