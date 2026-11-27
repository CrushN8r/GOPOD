// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_animation_marker.go
package wirepod_ttr

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// the model marks its animation cue as "happy___ spoken thought" instead of "happy...
var gopodAnimationMarker = gopodBrobotsSettings.AnimationMarker

// gopodAnimationMarkerRegex matches the marker, with any extra repeats of its last character
// ("____" for "___").
func gopodAnimationMarkerRegex() string {
	last := gopodAnimationMarker[len(gopodAnimationMarker)-1:]
	return regexp.QuoteMeta(gopodAnimationMarker) + regexp.QuoteMeta(last) + `*`
}

// A word directly followed by the marker, and the spaces after it.
var gopodAnimationMarkerPattern = regexp.MustCompile(`([A-Za-z][A-Za-z0-9-]*)[ \t]*` + gopodAnimationMarkerRegex() + `[ \t]*`)

// The prompt's placeholder word, which gemma2 copies in front of its cue - it was shown in the
// bubble and said as "_thinking." It is removed here.
var gopodPlaceholderWordPattern = regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(gopodBrobotsSettings.PlaceholderWord) + `s?[ \t_:]*`)

// A line that is only the prompt's example text, placeholder_thought, with at most a cue word after
// it - it was shown in the bubble and said as "thinking".
var gopodPlaceholderThoughtLinePattern = regexp.MustCompile(`(?im)^[ \t]*` + regexp.QuoteMeta(gopodBrobotsSettings.PlaceholderThought) + `[ \t:]*(?:([a-z][A-Za-z]*)[ \t]*)?(?:\n|$)`)

// A stage direction in square brackets was not said at all and sat unlit in the bubble.
var gopodBracketStageDirectionPattern = regexp.MustCompile(`\[([^\[\]\n]*\S[ \t]+\S[^\[\]\n]*)\]`)

// A line of dashes ("---", a markdown rule) ends what the robot says (buildGOPODStreamResult keeps
// nothing after it); the bubble ends there too, so it never shows lines that were not said.
var gopodRuleLinePattern = regexp.MustCompile(`(?m)^[ \t]*(?:-{3,}|\*{3,}|_{3,})[ \t]*$`)

// A sentence the model ends with only a wide gap and no stop ran on in speech and in the bubble.
var gopodMissingStopPattern = regexp.MustCompile(`([\p{L}\p{N}])[ \t]{2,}(\p{Lu})`)

// A last line that ends in the marker with nothing after it cues nothing and was said as "brobot
// 1_." - it is dropped.
var gopodDanglingCueLinePattern = regexp.MustCompile(`\n[ \t]*[^\n]*?` + gopodAnimationMarkerRegex() + `[ \t*_]*\s*$`)

// The model explaining its own answer after it was shown and said.
var gopodReplyTailPattern = regexp.MustCompile(`(?im)^[ \t]*(?:#{1,6}[ \t]|[*_>]*[ \t]*explanation\b)`)

// gopodWaitNotePattern: "(Wait for Brobot 0 before replying)", "--- (waiting ...)".
var gopodWaitNotePattern = regexp.MustCompile(`(?i)[ \t]*(?:-{2,}[ \t]*)?\((?:wait|waiting|pause|pausing)\b[^)\n]*\)`)

// gopodTrailingRulePattern: a "---" left at the end of a line.
var gopodTrailingRulePattern = regexp.MustCompile(`(?m)[ \t]+-{3,}[ \t]*$`)

// gopodCutReplyTail ends the reply at that line, then drops what's left dangling at its end: lines
// that are only an animation word ("happy", said "happy."), or only asterisks, underscores or
// spaces.
func gopodCutReplyTail(text string) string {
	if loc := gopodReplyTailPattern.FindStringIndex(text); loc != nil && loc[0] > 0 {
		text = text[:loc[0]]
	}
	lines := strings.Split(text, "\n")
	for len(lines) > 1 {
		last := strings.Trim(lines[len(lines)-1], " \t\r*_")
		if last != "" && !(gopodIsAnimationWord(last) && !strings.Contains(last, " ")) {
			break
		}
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

// A cue word the model misspells at the start of a line is read as the animation it is close to
// (within two letters of a token name - three for a name of ten or more - and six letters or more,
// so short real words are left alone).
var gopodNearCuePattern = regexp.MustCompile(`(?m)^([ \t]*)([A-Za-z]{6,})([ \t]*(?:\.\.\.|\x{2026}|` + gopodAnimationMarkerRegex() + `))`)

// A cue written with two underscores instead of the marker is read as the marker, when the word is
// an animation name.
var gopodShortMarkerPattern = regexp.MustCompile(`(?m)^([ \t]*)([A-Za-z]+)_{1,2}([^_A-Za-z0-9]|$)`)

// A reply that opens with a word in the cue's place - before "..." or the marker, then a new
// sentence - that isn't an animation name is read as the cue it stands for: an alias's animation
// ("laughing" -> happy), or the animation it's a form of ("loving" -> love, "celebrating" ->
// celebrate).
var gopodOpeningCueSlotPattern = regexp.MustCompile(`\A(\s*)([A-Za-z][A-Za-z*]*?)([ \t]*)(\.\.\.|\x{2026}|-{2,}|_+|\*{2,})([ \t]*)`)

// Speech text rules.
var (
	gopodAtSignRe     = regexp.MustCompile(`@[ \t]*`)
	gopodPlusSignRe   = regexp.MustCompile(`[ \t]*\+[ \t]*(\w?)`)
	gopodColonPauseRe = regexp.MustCompile(`([A-Za-z\)"'])[ \t]*:[ \t]+`)
	gopodShortYearRe  = regexp.MustCompile(`(^|[\s(])'(\d{2})\b`)
)

var gopodNewSentencePattern = regexp.MustCompile(`\A["'\x{201C}]?[A-Z]`)

// gopodCueSlotToken is the animation a word in the cue's place stands for, or "".
func gopodCueSlotToken(word string) string {
	key := normalizeAnimationToken(word)
	if target, ok := animationAliases[key]; ok {
		return target
	}
	if token := gopodNearAnimationToken(word); token != "" {
		return token
	}
	for _, animThing := range animationMap {
		token := animThing[0]
		t := strings.ToLower(token)
		n := 0
		for n < len(t) && n < len(key) && t[n] == key[n] {
			n++
		}
		shorter := len(t)
		if len(key) < shorter {
			shorter = len(key)
		}
		if n >= 3 && n*4 >= shorter*3 {
			return token
		}
	}
	return ""
}

// gopodFixOpeningCueSlot reads the reply's opening cue.
func gopodFixOpeningCueSlot(text string) string {
	g := gopodOpeningCueSlotPattern.FindStringSubmatchIndex(text)
	if g == nil {
		return text
	}
	written, mark, rest := text[g[4]:g[5]], text[g[8]:g[9]], text[g[1]:]
	word := strings.ReplaceAll(written, "*", "")
	dots := mark == "..." || mark == "\u2026"
	dashes := strings.HasPrefix(mark, "-")
	stars := strings.HasPrefix(mark, "*")
	if !dashes && (g[11] == g[10] || !gopodNewSentencePattern.MatchString(rest)) {
		return text
	}
	if gopodIsAnimationTokenName(word) && word == written && !dashes && !stars {
		return text
	}
	token := word
	if !gopodIsAnimationTokenName(word) {
		token = gopodCueSlotToken(word)
	}
	switch {
	case token != "" && dots:
		return text[:g[3]] + token + text[g[6]:g[1]] + rest
	case token != "":
		return text[:g[3]] + token + gopodAnimationMarker + " " + rest
	case stars:
		return text[:g[3]] + word + " " + rest
	case !dots && !dashes:
		return text[:g[3]] + rest
	}
	return text
}

func gopodFixNearCues(text string) string {
	text = gopodFixOpeningCueSlot(text)
	text = gopodShortMarkerPattern.ReplaceAllStringFunc(text, func(m string) string {
		g := gopodShortMarkerPattern.FindStringSubmatch(m)
		if gopodIsAnimationTokenName(g[2]) {
			return g[1] + g[2] + gopodAnimationMarker + g[3]
		}
		return m
	})
	return gopodNearCuePattern.ReplaceAllStringFunc(text, func(m string) string {
		g := gopodNearCuePattern.FindStringSubmatch(m)
		if token := gopodNearAnimationToken(g[2]); token != "" {
			return g[1] + token + g[3]
		}
		return m
	})
}

// gopodNearAnimationToken: the token name a misspelled cue word is within two letters of, or ""
// (also "" for an exact token name - nothing to fix).
func gopodNearAnimationToken(word string) string {
	lower := strings.ToLower(word)
	for _, anim := range animationMap {
		name := strings.ToLower(anim[0])
		if lower == name {
			return ""
		}
	}
	if token, ok := animationMisspellings[normalizeAnimationToken(word)]; ok {
		return token
	}
	for _, anim := range animationMap {
		name := strings.ToLower(anim[0])
		limit := 2
		if len(name) >= 10 {
			limit = 3 // "dartedEyes" -> "dartingEyes"
		}
		if gopodEditDistance(lower, name) <= limit {
			return anim[0]
		}
	}
	return ""
}

// Tidy-ups for leftovers the model sometimes adds: - a label in front of a line, "Brotobot moment:
// ..." -> "..."; - a markdown line with no words in it, "* ...
var gopodMomentLabelPattern = regexp.MustCompile(`(?im)^([ \t]*)\**[A-Za-z-]+[ \t]+moment[ \t]*\**[ \t]*:\**[ \t]*`)
var gopodCommaBeforeEndPattern = regexp.MustCompile(`,[ \t]*([?!.])`)

func gopodTidyLines(text string) string {
	text = gopodMomentLabelPattern.ReplaceAllString(text, "$1")
	text = gopodCommaBeforeEndPattern.ReplaceAllString(text, "$1")
	lines := strings.Split(text, "\n")
	kept := lines[:0]
	for _, line := range lines {
		if strings.ContainsAny(line, "*#") {
			words := strings.FieldsFunc(line, func(r rune) bool {
				return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r == '\'')
			})
			allCues := len(words) > 0
			for _, w := range words {
				if !gopodIsAnimationWord(w) {
					allCues = false
					break
				}
			}
			if len(words) == 0 || allCues {
				continue
			}
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

// gopodJoinSplitCues: a two-word cue name split by a space ("darting Eyes", "very happy") is joined
// back into its token when a cue marker, dots or a colon follow it.
func gopodJoinSplitCues(text string) string {
	for _, anim := range animationMap {
		name := anim[0]
		cut := strings.IndexFunc(name, func(r rune) bool { return r >= 'A' && r <= 'Z' })
		if cut <= 0 {
			continue
		}
		re := regexp.MustCompile(`(?i)(^|[\s|*(])(` + regexp.QuoteMeta(name[:cut]) + `)[ \t]+(` + regexp.QuoteMeta(name[cut:]) + `)([ \t]*(?:_{2,}|\.{2,}|\x{2026}|:))`)
		text = re.ReplaceAllString(text, "${1}"+name+"${4}")
	}
	return text
}

// gopodHTMLBreak: an HTML line break the model writes ("<br>", "<br/>").
var gopodHTMLBreak = regexp.MustCompile(`(?i)\s*<br\s*/?>\s*`)

func gopodPrepareReply(text string) string {
	text = gopodHTMLBreak.ReplaceAllString(text, "\n")
	text = gopodJoinSplitCues(text)
	text = gopodWaitNotePattern.ReplaceAllString(text, "")
	text = gopodTrailingRulePattern.ReplaceAllString(text, "")
	text = gopodTidyEllipses(text)
	text = gopodFixCueList(gopodStripCueWordNotes(text))
	text = gopodTidyLines(gopodFixNearCues(gopodUnboldCues(gopodCutReplyTail(text))))
	text = gopodDanglingCueLinePattern.ReplaceAllString(text, "")
	text = gopodMissingStopPattern.ReplaceAllString(text, "$1. $2")
	text = gopodDropPlaceholderThoughtLines(gopodPlaceholderWordPattern.ReplaceAllString(text, ""))
	return gopodBracketStageDirections(text)
}

// gopodBracketStageDirections: "[two words]" -> "*two words*", but never a link.
func gopodBracketStageDirections(text string) string {
	var b strings.Builder
	last := 0
	for _, m := range gopodBracketStageDirectionPattern.FindAllStringSubmatchIndex(text, -1) {
		if m[1] < len(text) && text[m[1]] == '(' {
			continue
		}
		b.WriteString(text[last:m[0]] + "*" + text[m[2]:m[3]] + "*")
		last = m[1]
	}
	b.WriteString(text[last:])
	return b.String()
}

// gopodMidEllipsisPattern: "..." / "…" between two words, however it's spaced ("me…I", "me
// ...I", "me...
var gopodMidEllipsisPattern = regexp.MustCompile(`([\pL\pN"'\x{2019})])[ \t]*(?:\.{3,}|\x{2026})[ \t]*`)

// gopodTidyEllipses: every such "..." becomes "word...
func gopodTidyEllipses(text string) string {
	var b strings.Builder
	last := 0
	for _, m := range gopodMidEllipsisPattern.FindAllStringSubmatchIndex(text, -1) {
		next, _ := utf8.DecodeRuneInString(text[m[1]:])
		if !unicode.IsLetter(next) && !unicode.IsDigit(next) {
			continue
		}
		b.WriteString(text[last:m[3]])
		b.WriteString("... ")
		last = m[1]
	}
	b.WriteString(text[last:])
	return b.String()
}

// gopodCueListPattern: a reply opening with cue words in a list, the last one marked with two or
// more underscores ("verySad, frustrated, angry__").
var gopodCueListPattern = regexp.MustCompile(`^\s*((?:[A-Za-z]+\s*,\s*)*)([A-Za-z]+)\s*_{2,}\s*`)

// gopodFixCueList keeps the last cue word of such a list, properly marked; the words are only
// dropped when every one is an animation word.
func gopodFixCueList(text string) string {
	m := gopodCueListPattern.FindStringSubmatch(text)
	if m == nil || !gopodIsAnimationWord(m[2]) {
		return text
	}
	for _, w := range strings.Split(m[1], ",") {
		if w = strings.TrimSpace(w); w != "" && !gopodIsAnimationWord(w) {
			return text
		}
	}
	return m[2] + gopodAnimationMarker + " " + text[len(m[0]):]
}

func gopodDropPlaceholderThoughtLines(text string) string {
	return gopodPlaceholderThoughtLinePattern.ReplaceAllStringFunc(text, func(line string) string {
		word := gopodPlaceholderThoughtLinePattern.FindStringSubmatch(line)[1]
		if word != "" && gopodIsAnimationWord(word) {
			return word + gopodAnimationMarker + " "
		}
		return ""
	})
}

// The model's own "..." / "…" in speech.
var gopodSpokenEllipsisPattern = regexp.MustCompile(`[ \t]*(?:(?:\.[ \t]*){3,}|\x{2026}[ \t]*)`)

// GopodEllipsisToComma is gopodEllipsisToComma for the song lines sdkapp says.
func GopodEllipsisToComma(text string) string {
	return gopodEllipsisToComma(text)
}

// gopodEllipsisToComma: each "..." / "…" after a word or a closing quote becomes ", "; one
// anywhere else (opening a line, after other punctuation) or closing a line just goes.
func gopodEllipsisToComma(text string) string {
	var b strings.Builder
	last := 0
	for _, m := range gopodSpokenEllipsisPattern.FindAllStringIndex(text, -1) {
		b.WriteString(text[last:m[0]])
		before := strings.TrimRight(text[:m[0]], " \t")
		after := text[m[1]:]
		endsLine := after == "" || after[0] == '\n' || after[0] == '\r'
		prev := ""
		if before != "" {
			prev = before[len(before)-1:]
		}
		if !endsLine && prev != "" && (strings.ContainsAny(prev, "\"')") || prev[0] >= 0x80 ||
			(prev[0] >= 'a' && prev[0] <= 'z') || (prev[0] >= 'A' && prev[0] <= 'Z') || (prev[0] >= '0' && prev[0] <= '9')) {
			b.WriteString(", ")
		} else if !endsLine {
			b.WriteString(" ")
		}
		last = m[1]
	}
	b.WriteString(text[last:])
	return b.String()
}

// A reply that opens with a bare lowercase animation word and no marker - the word is the cue,
// shown in the bubble and spoken otherwise.
var gopodBareLeadingCuePattern = regexp.MustCompile(`^[ \t]*([a-z][A-Za-z]*)[ \t]+`)

// gopodBareLeadingCue returns the leading bare cue word and the rest, or "" when the reply doesn't
// open with one.
func gopodBareLeadingCue(text string) (string, string) {
	m := gopodBareLeadingCuePattern.FindStringSubmatchIndex(text)
	if m == nil || gopodAnimationMarkerPattern.MatchString(text) {
		return "", text
	}
	word := text[m[2]:m[3]]
	if !gopodIsAnimationWord(word) {
		return "", text
	}
	return word, text[m[1]:]
}

// gopodAnimationMarkerToInternal: "happy___ text" -> "happy...
func gopodAnimationMarkerToInternal(text string) string {
	text = gopodPrepareReply(text)
	if word, rest := gopodBareLeadingCue(text); word != "" {
		text = word + gopodAnimationMarker + " " + rest
	}
	if gopodAnimationMarkerPattern.MatchString(text) {
		text = gopodEllipsisToComma(text)
	}
	return gopodAnimationMarkerPattern.ReplaceAllString(text, "$1... ")
}

// gopodDisplayReply: the reply as the chat bubble shows it.
func gopodDisplayReply(text string) string {
	text = gopodPrepareReply(text)
	if loc := gopodRuleLinePattern.FindStringIndex(text); loc != nil {
		text = strings.TrimRight(text[:loc[0]], " \t\r\n")
	}
	if word, rest := gopodBareLeadingCue(text); word != "" && gopodIsAnimationTokenName(word) {
		text = rest
	}
	out := gopodAnimationMarkerPattern.ReplaceAllStringFunc(text, func(cue string) string {
		word := gopodAnimationMarkerPattern.FindStringSubmatch(cue)[1]
		if gopodIsAnimationTokenName(word) {
			return ""
		}
		return word + "... "
	})
	return gopodDecorateStageDirections(gopodStripAnimationCueForDisplay(strings.TrimLeft(out, " \t\r\n")))
}

// Stage directions in the bubble get an emoji in front instead of their asterisks - "*circuits
// buzzing!*" shows as "🎬 circuits buzzing!".
var gopodStageDirectionSpanPattern = regexp.MustCompile(`\*([^*\n]+)\*`)

var gopodStageDirectionEmojiNext int

func gopodDecorateStageDirections(text string) string {
	matches := gopodStageDirectionSpanPattern.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return text
	}
	var b strings.Builder
	last := 0
	prevEnd := -1
	for _, m := range matches {
		inner := strings.TrimSpace(text[m[2]:m[3]])
		before := strings.TrimRight(text[:m[0]], " \t")
		opens := before == "" || strings.HasSuffix(before, ".") || strings.HasSuffix(before, "!") || strings.HasSuffix(before, "?") ||
			(prevEnd >= 0 && strings.TrimSpace(text[prevEnd:m[0]]) == "")
		b.WriteString(text[last:m[0]])
		if inner != "" && (len(strings.Fields(inner)) >= 2 || opens) {
			emojis := gopodBrobotsSettings.StageDirectionEmojis
			b.WriteString("\n" + emojis[gopodStageDirectionEmojiNext%len(emojis)] + " " + inner + "\n")
			gopodStageDirectionEmojiNext++
			prevEnd = m[1]
		} else {
			b.WriteString(text[m[0]:m[1]])
			prevEnd = -1
		}
		last = m[1]
	}
	b.WriteString(text[last:])
	out := gopodLineSpacePattern.ReplaceAllString(b.String(), "\n")
	return strings.Trim(gopodBlankLinesPattern.ReplaceAllString(out, "\n"), "\n")
}

var gopodLineSpacePattern = regexp.MustCompile(`[ \t]*\n[ \t]*`)
var gopodBlankLinesPattern = regexp.MustCompile(`\n{2,}`)

// gopodInternalToAnimationMarker turns a saved reply ("happy...
func gopodInternalToAnimationMarker(text string) string {
	units := parseBrobotPhraseUnits(text)
	if len(units) == 0 {
		if strings.TrimSpace(text) == "" {
			return text
		}
		return "thinking" + gopodAnimationMarker + " " + strings.TrimSpace(text)
	}
	parts := make([]string, 0, len(units))
	for _, u := range units {
		parts = append(parts, u.ActionParameter+gopodAnimationMarker+" "+u.Text)
	}
	return strings.Join(parts, " ")
}

// A cue written in markdown bold was missed: the bold was stripped later, after the cue was read,
// so the default animation played and the robot said "confused_".
var gopodBoldCuePattern = regexp.MustCompile(`[*_]{1,3}([A-Za-z]+)[*]{1,3}[ \t]*(` + gopodAnimationMarkerRegex() + `)`)

func gopodUnboldCues(text string) string {
	return gopodBoldCuePattern.ReplaceAllStringFunc(text, func(m string) string {
		g := gopodBoldCuePattern.FindStringSubmatch(m)
		if !gopodIsAnimationWord(g[1]) {
			return m
		}
		return g[1] + g[2]
	})
}

// ## EOF

// gopodSayPlus says every "+" as "plus": "18+." -> "18 plus.", "2+2" -> "2 plus 2".
func gopodSayPlus(text string) string {
	return gopodPlusSignRe.ReplaceAllStringFunc(text, func(m string) string {
		if next := gopodPlusSignRe.FindStringSubmatch(m)[1]; next != "" {
			return " plus " + next
		}
		return " plus"
	})
}
