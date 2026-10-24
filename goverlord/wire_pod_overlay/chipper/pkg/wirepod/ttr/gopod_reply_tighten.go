// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_reply_tighten.go
package wirepod_ttr

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// Tighter replies (operator 2026-10-06, from the three harvest takes:
// "Tighten the responses" - a typical answer was 145 words, about a minute of
// speech; TIGHTEN_RESPONSES_FROM_TAKES_001.md). The code does it, the prompt is
// unchanged (reply_tighten in gopod_brobots_settings.json):
//   - drop_lists: numbered and bulleted list lines and ``` blocks are cut, and
//     **bold** keeps its words only (the LLM dumped packs as pages);
//   - drop_stage_directions: *stage directions* are cut - a span of two words
//     or more, or one on stage_directions.json - so the bubble has no "🎬 ..."
//     lines (speech already had none); a single stressed *word* keeps its word;
//   - the response budget (operator 2026-10-07, "Can you hear me? vs Tell me
//     all about yourself"): the LLM's part keeps its first N sentences, N
//     worked out per question by gopodResponseBudget - not one cap for all.
// It runs on the LLM's reply only, before the scripted lines are stitched in,
// so a say line or flavour line is never cut or counted.

var (
	gopodCodeBlockPattern = regexp.MustCompile("(?s)```.*?(```|$)")
	gopodListLinePattern  = regexp.MustCompile(`(?m)^[ \t]*(?:\d+[.)]|[-•*])[ \t]+.*$`)
	gopodHeadingPattern   = regexp.MustCompile(`(?m)^[ \t]*#+[ \t]+.*$`)
	gopodBoldPattern      = regexp.MustCompile(`\*\*([^*\n]+?)\*\*`)
	gopodStarSpanPattern  = regexp.MustCompile(`\*([^*\n]+?)\*`)
)

// gopodResponseBudget is how many sentences a question gets: base_sentences,
// plus per_topic_sentences per topic it loads (or the topic's own
// topic_sentences), plus more_sentences when it asks
// for more (more_words), never over max_sentences.
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

// Anything in square brackets is a placeholder or a note, never speech
// ("[Name]", "[end]", 2026-10-07 dry run - said out loud). Code drops it.
var gopodStarRunPattern = regexp.MustCompile(`(^|\s)\*{2,}(\s|$)`)

var gopodBracketPlaceholderPattern = regexp.MustCompile(`[ \t]*\[[^\[\]\n]*\]`)

// gopodTightenReply tightens the LLM's reply (marker form, "thinking___ ...")
// to the question's budget. After a point's scripted line (gopodSayLine) the
// LLM's part gets say_line_sentences instead (2026-10-07: its talk after the
// line argued with it - "No time for that right now" after the websites line).
func gopodTightenReply(reply, question, esn string) string {
	t := gopodBrobotsSettings.ReplyTighten
	cue, body := "", reply
	if loc := gopodAnimationMarkerPattern.FindStringIndex(reply); loc != nil &&
		strings.TrimFunc(reply[:loc[0]], func(r rune) bool { return !unicode.IsLetter(r) }) == "" {
		cue, body = reply[:loc[1]], reply[loc[1]:]
	}
	before := len(strings.Fields(body))
	body = gopodBracketPlaceholderPattern.ReplaceAllString(body, "")
	// A run of stars on its own ("***", 2026-10-07 dry run) is never speech.
	body = gopodStarRunPattern.ReplaceAllString(body, " ")
	// A sentence where the LLM talks about its own answer ("The response
	// reflects each part of what makes a great Goverlord teaching answer",
	// 2026-10-07 dry run - said out loud) is cut: drop_sentences_with.
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
	// 2026-10-09 (operator: "every 'robot' possible = 'brobot' ... for
	// branding"): brand_words, whole words, any case.
	for _, w := range t.BrandWords {
		// 2026-10-23 (operator: "Keep exact test string: 'MAth' * If I say Keep
		// 'MAth' update the template"): a phrase that starts or ends with a mark
		// ("... Explain the MAth!") has no word edge there.
		re := regexp.QuoteMeta(w.From)
		if r := []rune(w.From); len(r) > 0 && (unicode.IsLetter(r[0]) || unicode.IsDigit(r[0])) {
			re = `\b` + re
		}
		if r := []rune(w.From); len(r) > 0 && (unicode.IsLetter(r[len(r)-1]) || unicode.IsDigit(r[len(r)-1])) {
			re += `\b`
		}
		body = regexp.MustCompile(`(?i)`+re).ReplaceAllString(body, w.To)
	}
	body = strings.Join(strings.Fields(body), " ")
	if strings.TrimFunc(body, func(r rune) bool { return !unicode.IsLetter(r) }) == "" {
		// All list or stage direction: keep the words, flattened.
		body = strings.Join(strings.Fields(strings.NewReplacer("*", "", "`", "").Replace(reply[len(cue):])), " ")
	}
	budget := gopodResponseBudget(question)
	if t.SayLineSentences > 0 && gopodSayLine(esn) != "" && budget > t.SayLineSentences {
		budget = t.SayLineSentences
	}
	// 2026-10-08 (operator: "remove 2-sentence clamps & stuff"): with
	// budget_off, no reply is cut to a sentence count; the clean-up stays.
	if t.BudgetOff {
		budget = 0
	}
	// 2026-10-22: a colour slot's colour is capped all the same (colour_slots).
	// The point bubbles bound his notes' answer themselves (gopod_point_bubbles.go).
	pointBubbles := gopodGivenHasNotes(esn) && gopodChainHas(esn, "point_bubbles")
	if c := gopodBrobotsSettings.ColourSlots.MaxSentences; c > 0 && !pointBubbles && (gopodSayLine(esn) != "" || gopodGivenHasNotes(esn)) && gopodChainHas(esn, "colour:value_point") && (budget == 0 || budget > c) {
		budget = c
	}
	// 2026-10-24 (operator: "Brobot 1 response * is amazing, lol, but too long"):
	// a free answer (no value point; a greeting too) keeps at most
	// reply_tighten.free_max_sentences, budget_off or not.
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

// gopodFirstSentences keeps the first n sentences. A sentence ends at ! or ?,
// or a single . (not "..."), followed by a space or the end; the closing quotes
// and emoji right after it stay with it.
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
			// Emoji standing after the last sentence stay with it ("bro! 🚀").
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

// ## EOF
