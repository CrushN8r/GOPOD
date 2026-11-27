// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_speech_sentences.go
package wirepod_ttr

import (
	"regexp"
	"strings"
)

// gopodSplitSpeechIntoSentences returns the actions with every SayText action split into one
// SayText per sentence.
func gopodSplitSpeechIntoSentences(actions []RobotAction) []RobotAction {
	var out []RobotAction
	for _, action := range actions {
		if action.Action != ActionSayText {
			out = append(out, action)
			continue
		}
		for _, sentence := range gopodSplitPieces(action.Parameter, true) {
			out = append(out, RobotAction{Action: ActionSayText, Parameter: sentence})
		}
	}
	return out
}

// gopodSentences splits a reply into the pieces said one at a time - and lit one at a time in the
// chat bubble.
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

// gopodHardEnd / gopodSoftEnd: the rune at i ends a sentence / a comma piece (it must also be
// followed by a space).
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

// gopodIsEmojiRune: an emoji or its joiners/variation marks (gopodPersonaEmojiRE's set).
func gopodIsEmojiRune(r rune) bool {
	return gopodPersonaEmojiRE.MatchString(string(r))
}

// gopodCutAt cuts text after every rune isEnd marks that is followed by a space or a line break,
// and at every line break.
func gopodCutAt(text string, isEnd func([]rune, int) bool) []string {
	var out []string
	var current strings.Builder
	runes := []rune(text)
	inside := make([]bool, len(runes))
	low := []rune(strings.ToLower(text))
	for _, p := range gopodBrobotsSettings.LiveSplit.KeepWhole {
		pr := []rune(strings.ToLower(strings.TrimSpace(p)))
		if len(pr) == 0 {
			continue
		}
		for i := 0; i+len(pr) <= len(low); i++ {
			if string(low[i:i+len(pr)]) == string(pr) {
				for k := i; k < i+len(pr)-1; k++ {
					inside[k] = true
				}
			}
		}
	}
	flush := func() {
		if piece := strings.TrimSpace(current.String()); piece != "" {
			out = append(out, piece)
		}
		current.Reset()
	}
	skip := -1
	for i, r := range runes {
		if i <= skip {
			continue
		}
		if r == '\n' {
			flush()
			continue
		}
		current.WriteRune(r)
		if i+1 < len(runes) && (runes[i+1] == ' ' || runes[i+1] == '\n') && isEnd(runes, i) && !inside[i] {
			j := i + 1
			for j < len(runes) && runes[j] == ' ' {
				j++
			}
			end := i
			for j < len(runes) && (gopodIsEmojiRune(runes[j]) || (runes[j] == ' ' && end > i)) {
				if runes[j] != ' ' {
					end = j
				}
				j++
			}
			if end > i {
				current.WriteString(string(runes[i+1 : end+1]))
				skip = end
			}
			flush()
		}
	}
	flush()
	return out
}

// gopodJoinPieces: a piece takes in the next one while the two together are at most
// live_split.join_up_to_words words (0: never), so short sentences and comma pieces are said as one
// run.
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

// An emoji the model uses as a sentence ending ("in each sentence 😉 You know") is removed before
// speech, which left the two sentences run together with no period.
var gopodEmojiSentenceEndPattern = regexp.MustCompile(`([\p{L}\p{N})"'])([ \t]*(?:[\x{1FA00}-\x{1FAFF}\x{1F600}-\x{1F64F}\x{1F300}-\x{1F5FF}\x{1F680}-\x{1F6FF}\x{1F1E0}-\x{1F1FF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}\x{1F900}-\x{1F9FF}\x{1F004}-\x{1F0CF}\x{1F191}-\x{1F251}\x{2B50}\x{FE0F}\x{200D}][ \t]*)+[ \t]+)(\p{Lu})`)

// gopodEmojiJoinersRe: what's left of a joined emoji ("🙅‍♂️") once its pictures are taken
// out - the zero-width joiner, the style marks.
var gopodEmojiJoinersRe = regexp.MustCompile(`[\x{200D}\x{FE0E}\x{FE0F}]`)

func gopodEmojiSentenceEnd(text string) string {
	return gopodEmojiSentenceEndPattern.ReplaceAllString(text, "$1.$2$3")
}

// gopodAsteriskWordGapPattern: asterisks with a word (or its ending punctuation) on one side and a
// word on the other.
var gopodAsteriskWordGapPattern = regexp.MustCompile(`([\p{L}\p{N}.,!?;:'")])\*+([\p{L}\p{N}"'(])`)

func gopodAsteriskWordGap(text string) string {
	return gopodAsteriskWordGapPattern.ReplaceAllString(text, "$1 $2")
}

// ## EOF
