// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_highlight_pieces.go
package wirepod_ttr

import (
	"strings"
	"unicode"
)

// gopodSpokenPieceTexts: the SayText pieces the speech loop in StreamingKGSim will say for these
// chunks, in order - the same steps, minus the animations (they add no words).
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

// gopodSegmentDisplay puts the segment marker into the bubble text where each spoken piece starts,
// and returns, for every spoken piece, the number of the bubble piece it lights (a spoken piece
// with no word in the bubble lights the piece before it).
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
		first, lead := -1, 0
		for i := range spoken {
			if pieceOf[i] != p {
				continue
			}
			if matchOf[i] >= 0 {
				first = matchOf[i]
				break
			}
			lead++
		}
		if first >= 0 && lead > 0 {
			if back := first - lead; back > lastWord {
				first = back
			} else {
				first = lastWord + 1
			}
		}
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
			c := shown[first].start
			for c > 0 && display[c-1] != ' ' && display[c-1] != '\n' &&
				!strings.HasSuffix(display[:c], "\u2014") && !strings.HasSuffix(display[:c], "\u2013") {
				c--
			}
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
			b.WriteString(piece + marker)
		}
		last = c
	}
	b.WriteString(display[last:])
	return b.String(), segOf
}

// ## EOF
