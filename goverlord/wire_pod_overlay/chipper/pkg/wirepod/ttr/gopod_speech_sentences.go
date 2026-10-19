// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_speech_sentences.go
package wirepod_ttr

import (
	"regexp"
	"strings"
)

// A reply chunk used to reach the robot as one SayText call. A SayText call can't
// be stopped once the robot starts it, so a touch interrupt only took effect after
// the whole chunk - often the whole reply. Splitting speech into sentences lets the
// chunk loop in StreamingKGSim check for an interrupt between sentences: a touch
// now stops the robot at the end of the sentence it is saying.

// The split, operator 2026-10-06: "aim for a hard split between sentences,
// even if punctuation is missing ... within each sentence, split by comma so
// highlighting and speech run together ... a sentence can be one word ...
// forget minimum word rules ... simplify and build up".
//  1. Sentences: a sentence ends at live_split.sentence_ends (. ! ?) before a
//     space (never at an ellipsis), or at a line break. Missing stops are put in before this, in
//     the speech clean-up: an emoji used as an ending (gopodEmojiSentenceEnd)
//     and a run of spaces before a capital (gopodMissingStopPattern).
//  2. Commas: inside each sentence, a piece ends at every live_split.comma_ends
//     (,) or ellipsis before a space.
// No length rules: "Wow!" is a sentence and a piece. The animation lead-in
// ("... ") stays at the front of the first piece. The chat bubble shows these
// same pieces (gopod_highlight_pieces.go); emojis after a piece's end stay with
// it there, and an animation cue plays before the piece it opens.

// gopodSplitSpeechIntoSentences returns the actions with every SayText action
// split into one SayText per sentence. Other actions are kept as they are.
func gopodSplitSpeechIntoSentences(actions []RobotAction) []RobotAction {
	var out []RobotAction
	for _, action := range actions {
		if action.Action != ActionSayText {
			out = append(out, action)
			continue
		}
		// 2026-10-08: a chat reply's short pieces are said as one run (gopodJoinPieces).
		for _, sentence := range gopodSplitPieces(action.Parameter, true) {
			out = append(out, RobotAction{Action: ActionSayText, Parameter: sentence})
		}
	}
	return out
}

// gopodSentences splits a reply into the pieces said one at a time - and lit one
// at a time in the chat bubble.
func gopodSentences(text string) []string {
	return gopodSplitPieces(text, false)
}

// gopodSplitPieces: the split above; join puts short pieces back together
// (live_split.join_up_to_words).
func gopodSplitPieces(text string, join bool) []string {
	text = strings.TrimSpace(text)
	body := strings.TrimLeft(text, ". \t\u2026")
	lead := text[:len(text)-len(body)] // the animation lead-in ("... "), if any
	most := gopodBrobotsSettings.LiveSplit.JoinUpToWords
	join = join && most > 0
	var pieces []string
	for _, sentence := range gopodWholeSentences(body) {
		// Joined, a sentence that fits stays whole, so a run breaks at a
		// sentence's end, not after a comma (live 20:24: "Doc," / "or DocBS").
		if join && len(strings.Fields(sentence)) <= most && !gopodHasMovePhrase(sentence) {
			pieces = append(pieces, sentence)
			continue
		}
		pieces = append(pieces, gopodCommaPieces(sentence)...)
	}
	if join {
		pieces = gopodJoinPieces(pieces)
	}
	if len(pieces) > 0 && lead != "" {
		pieces[0] = lead + pieces[0]
	}
	return pieces
}

// gopodHardEnd / gopodSoftEnd: the rune at i ends a sentence / a comma piece
// (it must also be followed by a space). The dots of an ellipsis are soft.
func gopodHardEnd(runes []rune, i int) bool {
	r := runes[i]
	if !strings.ContainsRune(gopodBrobotsSettings.LiveSplit.SentenceEnds, r) {
		return false
	}
	return !(r == '.' && i > 0 && runes[i-1] == '.')
}

func gopodSoftEnd(runes []rune, i int) bool {
	r := runes[i]
	return strings.ContainsRune(gopodBrobotsSettings.LiveSplit.CommaEnds, r) || r == '\u2026' || (r == '.' && i > 0 && runes[i-1] == '.')
}

// gopodCutAt cuts text after every rune isEnd marks that is followed by a space
// or a line break, and at every line break.
func gopodCutAt(text string, isEnd func([]rune, int) bool) []string {
	var out []string
	var current strings.Builder
	runes := []rune(text)
	flush := func() {
		if piece := strings.TrimSpace(current.String()); piece != "" {
			out = append(out, piece)
		}
		current.Reset()
	}
	for i, r := range runes {
		if r == '\n' {
			flush()
			continue
		}
		current.WriteRune(r)
		if i+1 < len(runes) && (runes[i+1] == ' ' || runes[i+1] == '\n') && isEnd(runes, i) {
			flush()
		}
	}
	flush()
	return out
}

// gopodJoinPieces (operator 2026-10-08: "less choppiness in responses ... very
// short comma filled sentences ... sound choppy and annoying ... a few sentences
// in a row sound smoother"): a piece takes in the next one while the two
// together are at most live_split.join_up_to_words words (0: never), so short
// sentences and comma pieces are said as one run. A piece holding a playing
// draft's move phrase stays on its own, so its move still lands on it.
func gopodJoinPieces(pieces []string) []string {
	most := gopodBrobotsSettings.LiveSplit.JoinUpToWords
	if most <= 0 || len(pieces) < 2 {
		return pieces
	}
	var out []string
	words := 0
	alone := false
	for _, p := range pieces {
		n := len(strings.Fields(p))
		move := gopodHasMovePhrase(p)
		if len(out) > 0 && !alone && !move && words+n <= most {
			out[len(out)-1] += " " + p
			words += n
			continue
		}
		out = append(out, p)
		words, alone = n, move
	}
	return out
}

// gopodHasMovePhrase: the text holds one of the playing draft's move phrases.
func gopodHasMovePhrase(text string) bool {
	lower := strings.ToLower(text)
	for _, at := range gopodScriptMovePhrases() {
		if strings.Contains(lower, at) {
			return true
		}
	}
	return false
}

func gopodWholeSentences(text string) []string { return gopodCutAt(text, gopodHardEnd) }

func gopodCommaPieces(sentence string) []string { return gopodCutAt(sentence, gopodSoftEnd) }

// An emoji the model uses as a sentence ending ("in each sentence 😉 You know")
// is removed before speech, which left the two sentences run together with no
// period (live 2026-09-29). gopodEmojiSentenceEnd adds a period before the emoji
// ("in each sentence. 😉 You know") when the emoji follows a word and the next word
// starts a new sentence (a capital). The emoji stays for the chat bubble. Emojis
// after "!", "?" or "." already have their ending and are left alone.
var gopodEmojiSentenceEndPattern = regexp.MustCompile(`([\p{L}\p{N})"'])([ \t]*(?:[\x{1FA00}-\x{1FAFF}\x{1F600}-\x{1F64F}\x{1F300}-\x{1F5FF}\x{1F680}-\x{1F6FF}\x{1F1E0}-\x{1F1FF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}\x{1F900}-\x{1F9FF}\x{1F004}-\x{1F0CF}\x{1F191}-\x{1F251}\x{2B50}\x{FE0F}\x{200D}][ \t]*)+[ \t]+)(\p{Lu})`)

// gopodEmojiJoinersRe: what's left of a joined emoji ("🙅‍♂️") once its
// pictures are taken out - the zero-width joiner, the style marks.
var gopodEmojiJoinersRe = regexp.MustCompile(`[\x{200D}\x{FE0E}\x{FE0F}]`)

func gopodEmojiSentenceEnd(text string) string {
	return gopodEmojiSentenceEndPattern.ReplaceAllString(text, "$1.$2$3")
}

// gopodAsteriskWordGapPattern: asterisks with a word (or its ending punctuation)
// on one side and a word on the other. Speech drops every asterisk, so these get
// a space instead: "fun!*Giggles merrily*" is said "fun! Giggles merrily".
var gopodAsteriskWordGapPattern = regexp.MustCompile(`([\p{L}\p{N}.,!?;:'")])\*+([\p{L}\p{N}"'(])`)

func gopodAsteriskWordGap(text string) string {
	return gopodAsteriskWordGapPattern.ReplaceAllString(text, "$1 $2")
}

// ## EOF
