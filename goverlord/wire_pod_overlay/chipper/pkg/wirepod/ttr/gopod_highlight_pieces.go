// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_highlight_pieces.go
package wirepod_ttr

import (
	"strings"
	"unicode"
)

// GOLDEN - the chat bubble's speaking highlight (operator 2026-10-06: "it has
// bugs, we fix, its fixed, we update and same bugs come back ... fix once
// correct & proper, lock it as a golden process").
//
// Why the bugs came back: the rule (sentence first, then commas) lived twice -
// here, on the text the robot SAYS, and in the chat page, on the text the
// bubble SHOWS - and those two texts differ (pronunciations, emojis, dashes and
// "..." said as pauses, asterisks, snippets, merged-word hyphens). Every new
// speech change made the two splits disagree, and the page then guessed which
// piece was being said by matching words.
//
// The golden process - one split, made here, nowhere else:
//  1. The pieces the robot will say are worked out once, before it speaks
//     (gopodSpokenPieceTexts - the same steps the speech loop takes).
//  2. The bubble text gets an invisible marker (scripted_split.segment_marker)
//     between its pieces, at the words where each spoken piece starts
//     (gopodSegmentDisplay). Words are lined up by their longest common run,
//     so a changed word ("GOPOD" said "Gowp-awd") never moves a boundary.
//  3. While speaking, gopod_speaking_now.json names the piece by its number
//     ("piece": 2). The page shows the pieces as given and lights that number:
//     it never splits a marked bubble and never guesses.
// A new speech change needs nothing here or in the page. Locked by
// TestHighlightPiecesGolden (gopod_brobots_settings_test.go).

// gopodSpokenPieceTexts: the SayText pieces the speech loop in StreamingKGSim
// will say for these chunks, in order - the same steps, minus the animations
// (they add no words).
func gopodSpokenPieceTexts(chunks []string) []string {
	var out []string
	for _, chunk := range chunks {
		acts := GetActionsFromString(gopodPronounceCommandPacket(commandPacketFromFilteredChunk(chunk)))
		for i := range acts {
			if acts[i].Action == ActionSayText {
				acts[i].Parameter = gopodEndWithPunctuation(acts[i].Parameter)
			}
		}
		for _, act := range gopodSplitSpeechIntoSentences(acts) {
			if act.Action == ActionSayText {
				out = append(out, act.Parameter)
			}
		}
	}
	return out
}

type gopodWordAt struct {
	word  string
	start int // byte offset in its text
}

func gopodHighlightWords(text string) []gopodWordAt {
	var out []gopodWordAt
	start := -1
	for i, r := range text + " " {
		// A curly apostrophe (It’s) is one too (live 2026-10-06 14:43: "It’s" in
		// the bubble against "It's" spoken split the word and moved the piece).
		// 2026-10-15: a repeating decimal's bar (U+0305, the math) stays inside
		// its number ("1̅4̅2̅8̅5̅7̅" is the word "142857", said "142857 repeating").
		isWord := unicode.IsLetter(r) || unicode.IsDigit(r) || r == '\'' || r == '\u2019' || r == '\u2018' || (start >= 0 && unicode.Is(unicode.Mn, r))
		if isWord && start < 0 {
			start = i
		} else if !isWord && start >= 0 {
			word := strings.NewReplacer("\u2019", "'", "\u2018", "'", "\u0305", "", "\u0304", "").Replace(text[start:i])
			out = append(out, gopodWordAt{strings.ToLower(strings.Trim(word, "'")), start})
			start = -1
		}
	}
	return out
}

// gopodSegmentDisplay puts the segment marker into the bubble text where each
// spoken piece starts, and returns, for every spoken piece, the number of the
// bubble piece it lights (a spoken piece with no word in the bubble lights the
// piece before it). With no pieces, or nothing lined up, the text comes back
// unmarked and every number is -1 (the page then lights nothing by number).
func gopodSegmentDisplay(display string, pieces []string) (string, []int) {
	segOf := make([]int, len(pieces))
	for i := range segOf {
		segOf[i] = -1
	}
	marker := gopodBrobotsSettings.ScriptedSplit.SegmentMarker
	if len(pieces) == 0 || marker == "" || strings.Contains(display, marker) {
		return display, segOf
	}
	// The spoken words, each with the number of its piece.
	var spoken []string
	var pieceOf []int
	for p, piece := range pieces {
		for _, w := range gopodHighlightWords(piece) {
			spoken = append(spoken, w.word)
			pieceOf = append(pieceOf, p)
		}
	}
	shown := gopodHighlightWords(display)
	if len(spoken) == 0 || len(shown) == 0 {
		return display, segOf
	}
	// Longest common run of words (LCS), then each spoken word's matched
	// bubble word.
	n, m := len(spoken), len(shown)
	table := make([][]int, n+1)
	for i := range table {
		table[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if spoken[i] == shown[j].word {
				table[i][j] = table[i+1][j+1] + 1
			} else if table[i+1][j] >= table[i][j+1] {
				table[i][j] = table[i+1][j]
			} else {
				table[i][j] = table[i][j+1]
			}
		}
	}
	matchOf := make([]int, n)
	for i := range matchOf {
		matchOf[i] = -1
	}
	for i, j := 0, 0; i < n && j < m; {
		if spoken[i] == shown[j].word {
			matchOf[i] = j
			i, j = i+1, j+1
		} else if table[i+1][j] >= table[i][j+1] {
			i++
		} else {
			j++
		}
	}
	// Each piece starts at its first matched word; boundaries only move forward.
	var cuts []int // byte offsets in display
	seg, lastWord := 0, -1
	for p := range pieces {
		first := -1
		for i := range spoken {
			if pieceOf[i] == p && matchOf[i] >= 0 {
				first = matchOf[i]
				break
			}
		}
		// 2026-10-08: a piece whose words are all said differently ("PipTA" said
		// "Pip-T-A", "crushn8r.ca" said "Crush-Ehnaydir-dotseeay") matches no
		// bubble word, and it merged into the piece before it. It gets the
		// unmatched bubble words between the piece before and the next match.
		if first < 0 && p > 0 {
			next := m
			for i := range spoken {
				if pieceOf[i] > p && matchOf[i] >= 0 {
					next = matchOf[i]
					break
				}
			}
			if lastWord+1 < next {
				first = lastWord + 1
			}
		}
		if p > 0 && first > lastWord && (lastWord >= 0 || first > 0) {
			// The cut goes before the whole token ("*blink*", "\"Hey"), so a
			// marker never lands inside one.
			// A long dash ends a token too: the robot pauses there
			// ("cameras—literally" is said "cameras, literally", 2026-10-07).
			c := shown[first].start
			for c > 0 && display[c-1] != ' ' && display[c-1] != '\n' &&
				!strings.HasSuffix(display[:c], "\u2014") && !strings.HasSuffix(display[:c], "\u2013") {
				c--
			}
			// 2026-10-15: a bullet ("• ", math.bullet) goes with its own line,
			// not the end of the piece before it.
			if b := gopodBrobotsSettings.Math.Bullet; b != "" && strings.HasSuffix(display[:c], b) {
				c -= len(b)
			}
			if len(cuts) == 0 || c > cuts[len(cuts)-1] {
				cuts = append(cuts, c)
				seg++
			}
		}
		segOf[p] = seg
		if first > lastWord {
			lastWord = first
		}
		if first >= 0 {
			for i := range spoken {
				if pieceOf[i] == p && matchOf[i] > lastWord {
					lastWord = matchOf[i]
				}
			}
		}
	}
	var b strings.Builder
	last := 0
	for _, c := range cuts {
		piece := display[last:c]
		if strings.HasSuffix(piece, " ") || strings.HasSuffix(piece, "\n") {
			b.WriteString(strings.TrimRight(piece, " "))
			b.WriteString(" " + marker)
		} else {
			// At a long dash: the marker alone, no space added.
			b.WriteString(piece + marker)
		}
		last = c
	}
	b.WriteString(display[last:])
	return b.String(), segOf
}

// ## EOF
