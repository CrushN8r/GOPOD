// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_animation_marker.go
package wirepod_ttr

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// 2026-09-30 (operator: "replace with "___" as the animation marker"): the
// model marks its animation cue as "happy___ spoken thought" instead of
// "happy... spoken thought". Three dots are also ordinary speech ("Ah...",
// "uh oh..."), so the old marker was mixed up with real pauses - a leading
// "Ah..." was dropped from the chat bubble. Three underscores never occur in
// speech.
//
// Wire-Pod's own reply pipeline still works in its internal "happy... text"
// form, so the reply is converted once where it arrives (kgsim.go); the chat
// display is built from the reply as written, with the cue removed.
// The marker itself is animation_marker in gopod_brobots_settings.json.
var gopodAnimationMarker = gopodBrobotsSettings.AnimationMarker

// gopodAnimationMarkerRegex matches the marker, with any extra repeats of its
// last character ("____" for "___").
func gopodAnimationMarkerRegex() string {
	last := gopodAnimationMarker[len(gopodAnimationMarker)-1:]
	return regexp.QuoteMeta(gopodAnimationMarker) + regexp.QuoteMeta(last) + `*`
}

// A word directly followed by the marker, and the spaces after it.
var gopodAnimationMarkerPattern = regexp.MustCompile(`([A-Za-z][A-Za-z0-9-]*)[ \t]*` + gopodAnimationMarkerRegex() + `[ \t]*`)

// The prompt's placeholder word, which gemma2 copies in front of its cue
// ("actionParameter_thinking___ Honestly…", live 2026-10-01 09:21) - it was shown
// in the bubble and said as "_thinking." It is removed here. The prompts keep the
// word: rewritten as "happy___ spoken thought", gemma2 stopped writing the marker
// at all (0 of 5 replies, live 11:5x, against 5 of 5 with the word).
// The word is placeholder_word in gopod_brobots_settings.json.
var gopodPlaceholderWordPattern = regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(gopodBrobotsSettings.PlaceholderWord) + `s?[ \t_:]*`)

// A line that is only the prompt's example text, placeholder_thought, with at
// most a cue word after it ("spoken thought  thinking", Brobot 1, live
// 2026-10-01 15:25) - it was shown in the bubble and said as "thinking". The
// line is dropped; a cue word on it becomes the cue of the line after it.
var gopodPlaceholderThoughtLinePattern = regexp.MustCompile(`(?im)^[ \t]*` + regexp.QuoteMeta(gopodBrobotsSettings.PlaceholderThought) + `[ \t:]*(?:([a-z][A-Za-z]*)[ \t]*)?(?:\n|$)`)

// A stage direction in square brackets ("[BROBOT turns heads, looking for the
// audience]", Brobot 1, live 2026-10-01 16:09) was not said at all and sat
// unlit in the bubble. It is read like one in asterisks: said out loud, and
// shown with a stage-direction emoji. Only a bracket with two or more words -
// a one-word "[happy]" is still an animation cue.
var gopodBracketStageDirectionPattern = regexp.MustCompile(`\[([^\[\]\n]*\S[ \t]+\S[^\[\]\n]*)\]`)

// A line of dashes ("---", a markdown rule) ends what the robot says
// (buildGOPODStreamResult keeps nothing after it); the bubble ends there too, so
// it never shows lines that were not said (live 2026-10-01 16:09). "***" and
// "___" lines are rules too (a "***" closed two bubbles, live 2026-10-03 00:58).
var gopodRuleLinePattern = regexp.MustCompile(`(?m)^[ \t]*(?:-{3,}|\*{3,}|_{3,})[ \t]*$`)

// A sentence the model ends with only a wide gap and no stop ("spread good
// vibes   This ain't", live 2026-10-02 23:31) ran on in speech and in the
// bubble. A word followed by two or more spaces and a capital gets a period.
var gopodMissingStopPattern = regexp.MustCompile(`([\p{L}\p{N}])[ \t]{2,}(\p{Lu})`)

// A last line that ends in the marker with nothing after it ("brobot 1___  **",
// live 2026-10-04 10:02) cues nothing and was said as "brobot 1_." - it is
// dropped.
var gopodDanglingCueLinePattern = regexp.MustCompile(`\n[ \t]*[^\n]*?` + gopodAnimationMarkerRegex() + `[ \t*_]*\s*$`)

// The model explaining its own answer after it ("### Explanation:", "**Explanation
// of Robot's Thought Process:**", then "1. Emotional signal: …", live 2026-10-04
// 10:3x) was shown and said. The reply ends at the first markdown heading or
// "Explanation" line.
var gopodReplyTailPattern = regexp.MustCompile(`(?im)^[ \t]*(?:#{1,6}[ \t]|[*_>]*[ \t]*explanation\b)`)

// gopodWaitNotePattern: "(Wait for Brobot 0 before replying)", "--- (waiting ...)".
var gopodWaitNotePattern = regexp.MustCompile(`(?i)[ \t]*(?:-{2,}[ \t]*)?\((?:wait|waiting|pause|pausing)\b[^)\n]*\)`)

// gopodTrailingRulePattern: a "---" left at the end of a line.
var gopodTrailingRulePattern = regexp.MustCompile(`(?m)[ \t]+-{3,}[ \t]*$`)

// gopodCutReplyTail ends the reply at that line, then drops what's left dangling
// at its end: lines that are only an animation word ("happy", said "happy."), or
// only asterisks, underscores or spaces.
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

// A cue word the model misspells at the start of a line ("dartedEyes... Did
// someone say", live 2026-10-04 14:00, said "dartedEyes,") is read as the
// animation it is close to (within two letters of a token name - three for a
// name of ten or more - and six letters or more, so short real words are left
// alone).
var gopodNearCuePattern = regexp.MustCompile(`(?m)^([ \t]*)([A-Za-z]{6,})([ \t]*(?:\.\.\.|\x{2026}|` + gopodAnimationMarkerRegex() + `))`)

// A cue written with two underscores instead of the marker ("thinking__ Okay,
// so", live 2026-10-04 14:54 - shown in the bubble) is read as the marker, when
// the word is an animation name. One underscore too (2026-10-07 dry run: "happy_
// My peeps" was shown and said).
var gopodShortMarkerPattern = regexp.MustCompile(`(?m)^([ \t]*)([A-Za-z]+)_{1,2}([^_A-Za-z0-9]|$)`)

// A reply that opens with a word in the cue's place - before "..." or the
// marker, then a new sentence - that isn't an animation name ("loving... Two
// crates on one dolly", "laughing... Pick an idea", 2026-10-07 dry run - shown
// and said) is read as the cue it stands for: an alias's animation
// ("laughing" -> happy), or the animation it's a form of ("loving" -> love,
// "celebrating" -> celebrate). Before the marker, a word that is none of them
// is dropped - the marker is the cue's place. Mid-sentence ("party... time!!")
// a word stays a word the Brobot says (gopod_word_animations.go).
// 2026-10-25 (live: "happy** BINGO EXPERIENCE!!", "veryHappy** Shmengy bala ..."
// shown in the bubble): "**" after the opening cue word is a cue mark too.
var gopodOpeningCueSlotPattern = regexp.MustCompile(`\A(\s*)([A-Za-z][A-Za-z*]*?)([ \t]*)(\.\.\.|\x{2026}|-{2,}|_+|\*{2,})([ \t]*)`)

// Speech text rules (kgsim.go's speech cleanup, 2026-10-07).
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

// gopodFixOpeningCueSlot reads the reply's opening cue (2026-10-07 dry runs:
// "loving... Two crates", "happy--- Awaken", "darlingeyes---- Is that you?",
// "dar*ingEyes... You've got" - each shown and said). Before "..." or
// underscores, only when a new sentence follows: "party___ time!!" and "Bro___
// ok then" stay words the Brobot says. Before underscores, a word that is no
// animation is dropped; before dashes, a word that is no animation stays
// ("Well-- listen").
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
		// Not an animation: the word stays, the stray "**" goes ("Wow** That").
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

// gopodNearAnimationToken: the token name a misspelled cue word is within two
// letters of, or "" (also "" for an exact token name - nothing to fix).
func gopodNearAnimationToken(word string) string {
	lower := strings.ToLower(word)
	for _, anim := range animationMap {
		name := strings.ToLower(anim[0])
		if lower == name {
			return ""
		}
	}
	// The known misspellings first (animation_vocab.json "misspellings").
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

// Tidy-ups for leftovers the model sometimes adds (live 2026-10-04 15:4x,
// operator: "see what you can do about these lines"):
//   - a label in front of a line, "Brotobot moment: ..." -> "...";
//   - a markdown line with no words in it, "* ... *", "* 🤖 **0)**:" -> dropped;
//   - a markdown line of only animation words, "**Happy:** Celebrate." -> dropped;
//   - a comma right before an end mark, "to do,?" -> "to do?".
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

func gopodPrepareReply(text string) string {
	// 2026-10-17 (live: "--- (Wait for Brobot 0 before replying)" was shown and
	// said): a "(wait ...)" stage note, and a "---" at a line's end, are dropped.
	text = gopodWaitNotePattern.ReplaceAllString(text, "")
	text = gopodTrailingRulePattern.ReplaceAllString(text, "")
	text = gopodTidyEllipses(text)
	// 2026-10-11 (live 23:2x): "(Cue word: thinking)" notes, and "verySad,
	// frustrated, angry__ Haha" - a list of cue words with a short marker.
	text = gopodFixCueList(gopodStripCueWordNotes(text))
	text = gopodTidyLines(gopodFixNearCues(gopodUnboldCues(gopodCutReplyTail(text))))
	text = gopodDanglingCueLinePattern.ReplaceAllString(text, "")
	text = gopodMissingStopPattern.ReplaceAllString(text, "$1. $2")
	text = gopodDropPlaceholderThoughtLines(gopodPlaceholderWordPattern.ReplaceAllString(text, ""))
	return gopodBracketStageDirections(text)
}

// gopodBracketStageDirections: "[two words]" -> "*two words*", but never a link
// (2026-10-16, live: "[GOPOD Weather Kit](/gopod_weather_kit.html)" lost its
// brackets and showed as plain words and an address).
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

// gopodMidEllipsisPattern: "..." / "…" between two words, however it's spaced
// ("me…I", "me ...I", "me... I", "me ... I").
var gopodMidEllipsisPattern = regexp.MustCompile(`([\pL\pN"'\x{2019})])[ \t]*(?:\.{3,}|\x{2026})[ \t]*`)

// gopodTidyEllipses (2026-10-12, operator: 'split error on "..." must handle
// "[word]...[word]", "[word] ...[word]", "[word]... [word]", "[word] ... [word]"';
// live: "The answer for me…I sleep" lit "me…I" with the wrong piece): every such
// "..." becomes "word... word", one form for the speech and the bubble alike.
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

// gopodCueListPattern: a reply opening with cue words in a list, the last one
// marked with two or more underscores ("verySad, frustrated, angry__").
var gopodCueListPattern = regexp.MustCompile(`^\s*((?:[A-Za-z]+\s*,\s*)*)([A-Za-z]+)\s*_{2,}\s*`)

// gopodFixCueList keeps the last cue word of such a list, properly marked; the
// words are only dropped when every one is an animation word.
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

// The model's own "..." / "…" in speech (operator 2026-10-01: "is there anywhere
// GOPOD can further absolve itself of using "..." in any meaningful way?"). In a
// reply that uses the "___" marker they are dropped from the speech path before
// the cues are converted: they no longer end a sentence (a new say, where an
// animation could land - "…or is it the test?" was not heard, live 09:32), and an
// animation word before one ("Wow! Happy... birthday") can no longer be taken for
// a cue. The bubble keeps them (gopodDisplayReply reads the reply as written).
// Since 2026-10-01 17:1x (operator: '"..." = ","') one after a word is said as a
// comma pause instead of nothing ("put something in... and it got" was said
// "in and it got"); see gopodEllipsisToComma.
var gopodSpokenEllipsisPattern = regexp.MustCompile(`[ \t]*(?:(?:\.[ \t]*){3,}|\x{2026}[ \t]*)`)

// GopodEllipsisToComma is gopodEllipsisToComma for the song lines sdkapp says
// (2026-10-03, operator: '"..." = "," (comma for the pause)').
func GopodEllipsisToComma(text string) string {
	return gopodEllipsisToComma(text)
}

// gopodEllipsisToComma: each "..." / "…" after a word or a closing quote becomes
// ", "; one anywhere else (opening a line, after other punctuation) or closing a
// line just goes.
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

// A reply that opens with a bare lowercase animation word and no marker ("happy
// circuits buzzing! Hey there", live 2026-10-01 11:5x) - the word is the cue,
// shown in the bubble and spoken otherwise. Lowercase only, so "Happy to help!"
// stays speech.
var gopodBareLeadingCuePattern = regexp.MustCompile(`^[ \t]*([a-z][A-Za-z]*)[ \t]+`)

// gopodBareLeadingCue returns the leading bare cue word and the rest, or "" when
// the reply doesn't open with one.
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

// gopodAnimationMarkerToInternal: "happy___ text" -> "happy... text", for every
// marked cue in the reply; the reply's own "..." are dropped first (above). A bare
// leading cue word ("happy circuits…") is taken as the cue too.
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

// gopodDisplayReply: the reply as the chat bubble shows it. A marked animation
// cue is removed ("happy___ Ah... so" -> "Ah... so"); a marked word that is not
// an animation keeps the word as a pause ("Bro___ ok" -> "Bro... ok"). An
// old-style opening cue ("happy." / "happy...") is removed too.
func gopodDisplayReply(text string) string {
	text = gopodPrepareReply(text)
	if loc := gopodRuleLinePattern.FindStringIndex(text); loc != nil {
		text = strings.TrimRight(text[:loc[0]], " \t\r\n")
	}
	if word, rest := gopodBareLeadingCue(text); word != "" && gopodIsAnimationTokenName(word) {
		text = rest
	}
	// 2026-10-02: only a token name is a cue the bubble hides; an alias is a
	// word the Brobot says, so it is shown (gopod_word_animations.go).
	out := gopodAnimationMarkerPattern.ReplaceAllStringFunc(text, func(cue string) string {
		word := gopodAnimationMarkerPattern.FindStringSubmatch(cue)[1]
		if gopodIsAnimationTokenName(word) {
			return ""
		}
		return word + "... "
	})
	// Leading blank lines go too: a reply that opens on a new line ("\nI'm Brobot
	// 1", live 2026-10-01 15:54) left the log's "LLM response for <ESN>:" line
	// empty, and the chat page showed no bubble.
	return gopodDecorateStageDirections(gopodStripAnimationCueForDisplay(strings.TrimLeft(out, " \t\r\n")))
}

// Stage directions in the bubble get an emoji in front instead of their
// asterisks (operator 2026-10-01: "decorate the stage directions, like prepend
// with emojis? the megaphone and movie scene emojis?") - "*circuits buzzing!*"
// shows as "🎬 circuits buzzing!". The emojis are stage_direction_emojis in
// gopod_brobots_settings.json, taken in turn. A span counts as a stage direction
// when it has two or more words, or opens the reply, a sentence or follows
// another one ("*thinking* *sarcastic*"); a single stressed word mid-sentence
// ("my *brain* switch") is left as written. Bubble only - speech is unchanged.
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
			// On its own line (operator 2026-10-01 17:1x: "make director messages
			// and similar in their own paragraph, for easier visual grasping").
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

// gopodInternalToAnimationMarker turns a saved reply ("happy... text", Wire-Pod's
// internal form) back into the "happy___ text" form the prompt asks for, so the
// shared chat memory shows the model its own past answers in the same form it
// must write - not the old "..." one. A reply with no cue gets Wire-Pod's default
// "thinking" cue.
func gopodInternalToAnimationMarker(text string) string {
	units := parseBrobotPhraseUnits(text)
	if len(units) == 0 {
		// No cue written: Wire-Pod played its default "thinking" for it, so memory
		// shows that - every remembered answer is an example of the marker.
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

// A cue written in markdown bold ("**confused**___ Say ...", live 2026-10-04
// 19:26) was missed: the bold was stripped later, after the cue was read, so the
// default animation played and the robot said "confused_". The bold is taken off
// a cue word right before the marker first. A GOPOD glitch, not a reply to patch:
// the cue is the page's and the robot's own format.
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
