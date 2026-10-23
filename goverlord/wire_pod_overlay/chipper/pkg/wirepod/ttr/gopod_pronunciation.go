// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_pronunciation.go
package wirepod_ttr

import (
	"encoding/json"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// The Brobot chat path speaks with the same pronunciation list as the songs and
// go search: SPEECH_PRONUNCIATION_REGISTRY_001 in the interview template (the one
// source of truth, read by run_section1_full_live_001.py's
// apply_pronunciation_safety()). Same rules: longest entry first, case-insensitive,
// whole words only. Read at startup; edit the template and restart wire-pod.service.

// An absolute path in the GOPOD repo, overridable via GOPOD_PRONUNCIATION_REGISTRY_PATH
// (not gopodPathWithEnvOverride, which joins onto the chipper folder).
var gopodPronunciationRegistryPath = func() string {
	if override := strings.TrimSpace(os.Getenv("GOPOD_PRONUNCIATION_REGISTRY_PATH")); override != "" {
		return override
	}
	return "/home/goverlord/crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/gopod_probes/section_packets/brobots_wirepod_interview_section_card_template_1_001.md"
}()

type gopodPronunciationEntry struct {
	From string `json:"from"`
	To   string `json:"to"`
	// MatchCase (2026-10-06): "LED" but not "led" (he led the way).
	MatchCase bool `json:"match_case"`
	// WordStart (2026-10-17, operator: "brobot lingo: choco[text] = chalk[new
	// text]"): From at the start of a word, the rest of the word kept
	// ("chocoyum" -> "chalkyum"); Except: whole words it leaves alone.
	WordStart bool     `json:"word_start"`
	Except    []string `json:"except"`
	pattern   *regexp.Regexp
}

var gopodPronunciationBlockPattern = regexp.MustCompile("(?s)```json SPEECH_PRONUNCIATION_REGISTRY_001[ \\t]*\\r?\\n(.*?)\\n```")

func loadGOPODPronunciations() []gopodPronunciationEntry {
	data, err := os.ReadFile(gopodPronunciationRegistryPath)
	if err != nil {
		logger.Println("Unable to load GOPOD pronunciation registry at " + gopodPronunciationRegistryPath + ": " + err.Error())
		return nil
	}
	block := gopodPronunciationBlockPattern.FindSubmatch(data)
	if block == nil {
		logger.Println("GOPOD pronunciation registry block not found in " + gopodPronunciationRegistryPath)
		return nil
	}
	var registry struct {
		Entries []gopodPronunciationEntry `json:"pronunciation_entries"`
	}
	if err := json.Unmarshal(block[1], &registry); err != nil {
		logger.Println("Unable to parse GOPOD pronunciation registry: " + err.Error())
		return nil
	}
	var entries []gopodPronunciationEntry
	for _, e := range registry.Entries {
		if e.From == "" || e.To == "" {
			continue
		}
		flags := "(?i)"
		if e.MatchCase {
			flags = ""
		}
		e.pattern = regexp.MustCompile(flags + regexp.QuoteMeta(e.From))
		entries = append(entries, e)
	}
	sort.SliceStable(entries, func(i, j int) bool { return len(entries[i].From) > len(entries[j].From) })
	return entries
}

var loadedGOPODPronunciations = loadGOPODPronunciations()

func gopodIsWordRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// gopodApplyPronunciations replaces whole-word matches, like the Python side's
// (?<!\w)...(?!\w) pattern.
func gopodApplyPronunciations(text string) string {
	return gopodSplitMergedWords(gopodReplaceWholeWords(gopodSayStretchedWords(gopodSayDegrees(text)), loadedGOPODPronunciations))
}

// gopodSayDegrees (2026-10-12, operator: "rich display = robot safe say * ° =
// degrees"): the bubble keeps "16°C"; the robot says "16 degrees Celsius".
var gopodDegreesSaid = strings.NewReplacer("°C", " degrees Celsius", "°F", " degrees Fahrenheit", "°", " degrees")

// 2026-10-23 (operator: "So it's a 57° day (degrees is singular here, how to
// handle?)"): "°" before a word that isn't its unit is one word with it - "a 57°
// day" is said "a 57 degree day"; "57° Fahrenheit" stays "degrees".
var gopodDegreeBeforeWordRE = regexp.MustCompile(`°(\s+)(\p{L}+)`)

func gopodSayDegrees(text string) string {
	text = strings.NewReplacer("°C", " degrees Celsius", "°F", " degrees Fahrenheit").Replace(text)
	text = gopodDegreeBeforeWordRE.ReplaceAllStringFunc(text, func(m string) string {
		sub := gopodDegreeBeforeWordRE.FindStringSubmatch(m)
		switch strings.ToLower(sub[2]) {
		case "celsius", "fahrenheit", "c", "f", "and", "or", "at", "in", "with", "outside", "today", "out":
			return " degrees" + sub[1] + sub[2]
		}
		return " degree" + sub[1] + sub[2]
	})
	return gopodDegreesSaid.Replace(text)
}

// gopodSayStretchedWords (stretched_words in gopod_brobots_settings.json;
// operator 2026-10-06: "Long 'brooooo' = 'bro-o-o'"): a word ending in one
// vowel repeated min_repeat+ times is said with that vowel said_repeats times,
// hyphen-joined - brooooo -> bro-o-o. Runs before the registry, so "bro-o-o"
// is not respelled.
func gopodSayStretchedWords(text string) string {
	minRepeat, said := gopodBrobotsSettings.StretchedWords.MinRepeat, gopodBrobotsSettings.StretchedWords.SaidRepeats
	return gopodLetterWordPattern.ReplaceAllStringFunc(text, func(word string) string {
		rs := []rune(word)
		last := rs[len(rs)-1]
		if !strings.ContainsRune("aeiouAEIOU", last) {
			return word
		}
		run := 0
		for i := len(rs) - 1; i >= 0 && unicode.ToLower(rs[i]) == unicode.ToLower(last); i-- {
			run++
		}
		if run < minRepeat || run == len(rs) {
			return word
		}
		parts := []string{string(rs[:len(rs)-run+1])}
		for i := 1; i < said; i++ {
			parts = append(parts, string(rs[len(rs)-1]))
		}
		return strings.Join(parts, "-")
	})
}

// Merged words (merged_words in gopod_brobots_settings.json; operator
// 2026-10-06: "snoozefest = snooze-fest"): a long word that isn't in the
// dictionary but splits cleanly into two dictionary words is spoken with a
// hyphen. The registry above runs first and its words are never split, so a
// wrong split is fixed with a registry entry or keep_whole.
var gopodDictionaryWords, gopodKeepWholeWords = loadGOPODMergedWordLists()

func loadGOPODMergedWordLists() (map[string]bool, map[string]bool) {
	keep := map[string]bool{}
	for _, w := range gopodBrobotsSettings.MergedWords.KeepWhole {
		keep[strings.ToLower(w)] = true
	}
	for _, e := range loadedGOPODPronunciations {
		for _, w := range strings.FieldsFunc(e.From+" "+e.To, func(r rune) bool { return !unicode.IsLetter(r) }) {
			keep[strings.ToLower(w)] = true
		}
	}
	dict := map[string]bool{}
	for _, path := range gopodBrobotsSettings.MergedWords.Dictionaries {
		data, err := os.ReadFile(path)
		if err != nil {
			logger.Println("GOPOD merged words: dictionary skipped: " + err.Error())
			continue
		}
		for _, w := range strings.Split(string(data), "\n") {
			if w = strings.TrimSpace(w); w != "" && !strings.Contains(w, "'") {
				dict[strings.ToLower(w)] = true
			}
		}
	}
	if len(dict) == 0 {
		logger.Println("GOPOD merged words: no dictionary, no splitting")
		return nil, keep
	}
	// Known words count as dictionary words, endings and all (rockstars).
	for w := range keep {
		dict[w] = true
	}
	return dict, keep
}

// gopodSplitScore ranks a split at i of an n-letter word: the shorter part's
// length first, then the second part's.
func gopodSplitScore(i, n int) int {
	short := i
	if n-i < short {
		short = n - i
	}
	return short*100 + (n - i)
}

var gopodLetterWordPattern = regexp.MustCompile(`\pL+`)

func gopodSplitMergedWords(text string) string {
	if gopodDictionaryWords == nil {
		return text
	}
	minLetters, minPart := gopodBrobotsSettings.MergedWords.MinLetters, gopodBrobotsSettings.MergedWords.MinPart
	return gopodLetterWordPattern.ReplaceAllStringFunc(text, func(word string) string {
		lower := strings.ToLower(word)
		rs := []rune(word)
		// Names written in capitals inside the word (CamelCase, ALL CAPS) and
		// known words stay whole.
		if len(rs) < minLetters || gopodDictionaryWords[lower] || gopodKeepWholeWords[lower] || strings.ToLower(string(rs[1:])) != string(rs[1:]) {
			return word
		}
		// A dictionary word with an ending (robotic+ally, blabber+ing) or a
		// dropped g (somethin) is a real word, not two merged ones.
		if gopodDictionaryWords[lower+"g"] {
			return word
		}
		for _, end := range gopodBrobotsSettings.MergedWords.Endings {
			if base := strings.TrimSuffix(lower, end); base != lower && (gopodDictionaryWords[base] || gopodDictionaryWords[base+"e"]) {
				return word
			}
		}
		// The most even split wins, then the longer second part:
		// kickstart -> kick-start, not kicks-tart.
		lr := []rune(lower)
		best := 0
		for i := minPart; i <= len(rs)-minPart; i++ {
			if !gopodDictionaryWords[string(lr[:i])] || !gopodDictionaryWords[string(lr[i:])] {
				continue
			}
			if best == 0 || gopodSplitScore(i, len(rs)) > gopodSplitScore(best, len(rs)) {
				best = i
			}
		}
		if best == 0 {
			return word
		}
		return string(rs[:best]) + "-" + string(rs[best:])
	})
}

// gopodReplaceWholeWords applies a from/to list (already sorted longest first)
// to whole-word matches only; shared by the pronunciation and name-spelling lists.
func gopodReplaceWholeWords(text string, list []gopodPronunciationEntry) string {
	for _, e := range list {
		var b strings.Builder
		last := 0
		for _, m := range e.pattern.FindAllStringIndex(text, -1) {
			before, after := []rune(text[:m[0]]), []rune(text[m[1]:])
			if len(before) > 0 && gopodIsWordRune(before[len(before)-1]) {
				continue
			}
			if !e.WordStart && len(after) > 0 && gopodIsWordRune(after[0]) {
				continue
			}
			// 2026-10-17 (operator: "chalk-o- is intended for brobot made up
			// lingo ... try to not include normal words"): a word in the
			// dictionary (chocolate) is left alone, and so is one in Except.
			if e.WordStart && gopodPronunciationExcept(e, text[m[0]:]) {
				continue
			}
			b.WriteString(text[last:m[0]])
			b.WriteString(e.To)
			last = m[1]
		}
		b.WriteString(text[last:])
		text = b.String()
	}
	return text
}

// gopodPronunciationExcept: the word starting here is one of e.Except.
func gopodPronunciationExcept(e gopodPronunciationEntry, from string) bool {
	end := 0
	for i, r := range from {
		if !gopodIsWordRune(r) {
			break
		}
		end = i + len(string(r))
	}
	word := from[:end]
	if gopodDictionaryWords[strings.ToLower(word)] && !strings.EqualFold(word, e.From) {
		return true
	}
	for _, x := range e.Except {
		if strings.EqualFold(x, word) {
			return true
		}
	}
	return false
}

// gopodPronounceCommandPacket applies the list to the spoken part of a
// "{{playAnimationWI||x}}... words" packet only, never to the command.
func gopodPronounceCommandPacket(packet string) string {
	if i := strings.Index(packet, "}}"); i >= 0 {
		return packet[:i+2] + gopodApplyPronunciations(packet[i+2:])
	}
	return gopodApplyPronunciations(packet)
}

// ## EOF
