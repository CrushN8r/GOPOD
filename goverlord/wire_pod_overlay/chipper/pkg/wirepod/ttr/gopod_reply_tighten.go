// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_reply_tighten.go
package wirepod_ttr

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

var (
	gopodCodeBlockPattern = regexp.MustCompile("(?s)```.*?(```|$)")
	gopodListLinePattern  = regexp.MustCompile(`(?m)^[ \t]*(?:\d+[.)]|[-•*])[ \t]+.*$`)
	gopodHeadingPattern   = regexp.MustCompile(`(?m)^[ \t]*#+[ \t]+.*$`)
	gopodBoldPattern      = regexp.MustCompile(`\*\*([^*\n]+?)\*\*`)
	gopodStarSpanPattern  = regexp.MustCompile(`\*([^*\n]+?)\*`)
)

// gopodResponseBudget is how many sentences a question gets: base_sentences, plus
// per_topic_sentences per topic it loads (or the topic's own topic_sentences), plus more_sentences
// when it asks for more (more_words), never over max_sentences.
func gopodResponseBudget(question string) int {
	t := gopodBrobotsSettings.ReplyTighten
	n := t.BaseSentences
	for _, p := range loadScenarioPacketsForMessage(question) {
		if own, ok := t.TopicSentences[p.Name]; ok {
			n += own
		} else {
			n += t.PerTopicSentences
		}
	}
	lower := strings.ToLower(question)
	for _, w := range t.MoreWords {
		if gopodHasWholeWords(lower, strings.ToLower(w)) {
			n += t.MoreSentences
			break
		}
	}
	if t.MaxSentences > 0 && n > t.MaxSentences {
		n = t.MaxSentences
	}
	return n
}

// Anything in square brackets is a placeholder or a note, never speech.
var gopodStarRunPattern = regexp.MustCompile(`(^|\s)\*{2,}(\s|$)`)

var gopodBracketPlaceholderPattern = regexp.MustCompile(`[ \t]*\[[^\[\]\n]*\]`)

// gopodTightenReply tightens the LLM's reply (marker form, "thinking___ ...") to the question's
// budget.
var gopodEmphasisPattern = regexp.MustCompile(`\*([^*\s]+)\*`)

func gopodTightenReply(reply, question, esn string) string {
	t := gopodBrobotsSettings.ReplyTighten
	cue, body := "", reply
	if loc := gopodAnimationMarkerPattern.FindStringIndex(reply); loc != nil &&
		strings.TrimFunc(reply[:loc[0]], func(r rune) bool { return !unicode.IsLetter(r) }) == "" {
		cue, body = reply[:loc[1]], reply[loc[1]:]
	}
	before := len(strings.Fields(body))
	body = gopodStripLeaks(body)
	body = gopodBracketPlaceholderPattern.ReplaceAllString(body, "")
	body = gopodStarRunPattern.ReplaceAllString(body, " ")
	for _, phrase := range t.DropSentencesWith {
		body = regexp.MustCompile(`(?i)[^.!?\n]*\b`+regexp.QuoteMeta(phrase)+`\b[^.!?\n]*[.!?]*`).ReplaceAllString(body, "")
	}
	if t.DropLists {
		body = gopodCodeBlockPattern.ReplaceAllString(body, "")
		body = gopodListLinePattern.ReplaceAllString(body, "")
		body = gopodHeadingPattern.ReplaceAllString(body, "")
		body = gopodBoldPattern.ReplaceAllString(body, "$1")
	}
	if t.DropStageDirections {
		body = gopodStarSpanPattern.ReplaceAllStringFunc(body, func(span string) string {
			inner := strings.TrimSpace(gopodStarSpanPattern.FindStringSubmatch(span)[1])
			if len(strings.Fields(inner)) >= 2 || gopodIsStageDirection(inner) {
				return ""
			}
			return span
		})
		for _, e := range gopodBrobotsSettings.StageDirectionEmojis {
			body = regexp.MustCompile(`(?m)^[ \t]*`+regexp.QuoteMeta(e)+`.*$`).ReplaceAllString(body, "")
		}
	}
	body = gopodEmphasisPattern.ReplaceAllString(body, "$1")
	for _, w := range t.BrandWords {
		re := regexp.QuoteMeta(w.From)
		if r := []rune(w.From); len(r) > 0 && (unicode.IsLetter(r[0]) || unicode.IsDigit(r[0])) {
			re = `\b` + re
		}
		if r := []rune(w.From); len(r) > 0 && (unicode.IsLetter(r[len(r)-1]) || unicode.IsDigit(r[len(r)-1])) {
			re += `\b`
		}
		body = regexp.MustCompile(`(?i)`+re).ReplaceAllString(body, w.To)
	}
	if strings.Count(body, "\"")+strings.Count(body, "“")+strings.Count(body, "”") > 0 &&
		(strings.Count(body, "\"")%2 == 1 || strings.Count(body, "“") != strings.Count(body, "”")) {
		body = strings.NewReplacer("\"", "", "“", "", "”", "").Replace(body)
		body = gopodPersonaGapRE.ReplaceAllString(body, "$1")
	}
	body = strings.Join(strings.Fields(body), " ")
	if strings.TrimFunc(body, func(r rune) bool { return !unicode.IsLetter(r) }) == "" {
		body = strings.Join(strings.Fields(strings.NewReplacer("*", "", "`", "").Replace(reply[len(cue):])), " ")
	}
	budget := gopodResponseBudget(question)
	if t.SayLineSentences > 0 && gopodSayLine(esn) != "" && budget > t.SayLineSentences {
		budget = t.SayLineSentences
	}
	if t.BudgetOff {
		budget = 0
	}
	pointBubbles := gopodGivenHasNotes(esn) && gopodChainHas(esn, "point_bubbles")
	if c := gopodBrobotsSettings.ColourSlots.MaxSentences; c > 0 && !pointBubbles && (gopodSayLine(esn) != "" || gopodGivenHasNotes(esn)) && gopodChainHas(esn, "colour:value_point") && (budget == 0 || budget > c) {
		budget = c
	}
	if f := t.FreeMaxSentences; f > 0 && !pointBubbles && gopodSayLine(esn) == "" && !gopodGivenHasNotes(esn) && gopodAnswerType(esn, question) == "free" && (budget == 0 || budget > f) {
		budget = f
	}
	if budget > 0 {
		body = gopodFirstSentences(body, budget)
	}
	if after := len(strings.Fields(body)); after < before {
		logger.Println(fmt.Sprintf("BROBOT_REPLY_TIGHTENED budget=%d words=%d->%d", budget, before, after))
	}
	if cue == "" {
		return body
	}
	return strings.TrimRight(cue, " \t") + " " + body
}

// gopodFirstSentences keeps the first n sentences.
func gopodFirstSentences(text string, n int) string {
	r := []rune(text)
	count := 0
	for i := 0; i < len(r); i++ {
		if r[i] != '.' && r[i] != '!' && r[i] != '?' {
			continue
		}
		j := i
		for j < len(r) && (r[j] == '.' || r[j] == '!' || r[j] == '?') {
			j++
		}
		run := string(r[i:j])
		i = j - 1
		if !strings.ContainsAny(run, "!?") && len(run) != 1 {
			continue // "..." carries on
		}
		if run == "." && gopodNotSentenceEnd(r[:i+1]) {
			continue // "Dr."
		}
		for j < len(r) && !unicode.IsSpace(r[j]) && !unicode.IsLetter(r[j]) && !unicode.IsDigit(r[j]) {
			j++ // closing quotes, emoji
		}
		if j < len(r) && !unicode.IsSpace(r[j]) {
			continue // "crushn8r.ca", "4.5"
		}
		count++
		if count == n {
			for k := j; k < len(r); {
				for k < len(r) && unicode.IsSpace(r[k]) {
					k++
				}
				e := k
				for e < len(r) && !unicode.IsSpace(r[e]) && !unicode.IsLetter(r[e]) && !unicode.IsDigit(r[e]) && !strings.ContainsRune(`*"'(`, r[e]) {
					e++
				}
				if e == k || (e < len(r) && !unicode.IsSpace(r[e])) {
					break
				}
				j, k = e, e
			}
			return strings.TrimSpace(string(r[:j]))
		}
	}
	return text
}

// gopodNotSentenceEnd: the text so far ends in a not_sentence_ends word ("Dr.").
func gopodNotSentenceEnd(upTo []rune) bool {
	text := string(upTo)
	for _, w := range gopodBrobotsSettings.ReplyTighten.NotSentenceEnds {
		if strings.HasSuffix(text, w) {
			start := len(text) - len(w)
			if start == 0 || !unicode.IsLetter([]rune(text[:start])[len([]rune(text[:start]))-1]) {
				return true
			}
		}
	}
	return false
}

// gopodLeakTagRE / gopodLeakLabelRE: a tag or a working label in a reply is the model's
// scaffolding, not speech - it goes.
var gopodLeakTagRE = regexp.MustCompile(`</?[a-zA-Z_]{2,20}>`)
var gopodLeakLabelRE = regexp.MustCompile(`(?i)(^|[.!?\s])(?:note|notes|emoji|emojis|response|answer|reply|output)\s*:\s*`)

func gopodStripLeaks(s string) string {
	s = strings.ReplaceAll(s, " | ", ", ")
	s = gopodDropStrangeLinks(s)
	s = gopodLeakTagRE.ReplaceAllString(s, "")
	s = gopodLeakLabelRE.ReplaceAllString(s, "$1")
	return gopodLeakSpacesRE.ReplaceAllString(s, " ")
}

// gopodStrangeLinkRE: a web link in a reply, with the brackets around it when it has them.
var gopodStrangeLinkRE = regexp.MustCompile(`\(?\b(?:https?://|www\.)[^\s)]+\)?`)

// gopodDropStrangeLinks: links to anywhere but reply_tighten.link_hosts go.
func gopodDropStrangeLinks(s string) string {
	hosts := gopodBrobotsSettings.ReplyTighten.LinkHosts
	return gopodStrangeLinkRE.ReplaceAllStringFunc(s, func(link string) string {
		l := strings.ToLower(strings.Trim(link, "()"))
		l = strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(l, "https://"), "http://"), "www.")
		host := strings.SplitN(l, "/", 2)[0]
		for _, h := range hosts {
			if h = strings.ToLower(strings.TrimSpace(h)); h != "" && (host == h || strings.HasSuffix(host, "."+h)) {
				return link
			}
		}
		return ""
	})
}

// gopodLeakSpacesRE: runs of spaces a removed tag leaves - line breaks stay.
var gopodLeakSpacesRE = regexp.MustCompile(`[ \t]{2,}`)

// ## EOF
