// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_lookup.go
package wirepod_ttr

// The look-up (2026-11-16, operator: "Rather than repeat brain bubble bloat, do a
// swap in/out of topic clusters for the task at hand (based on key value
// words/phrases) * do look ups from the repo read me files * in the brain bubble,
// list which topic clusters were aggregated"; "matching topic clusters to the
// question's key words * repo's md files in topic-cluster format as the search
// pool"): every md file in the pool (brain_chain.lookup_files - topic_clusters/,
// the repo's own pages, the scrolls, the studio) is a card: its words (its
// "- **words:**" head line; a page without one is named by its file name), its
// first lines. A question picks the cards whose words it says (the longest word
// first, whole words); the brain's FACT is their first lines, and the bubble
// names them (LOOKED UP). Nothing is copied - each card is read where it lives.

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type gopodLookupCard struct {
	Path  string
	Words []string
	Lines []string
}

var gopodLookupCache = struct {
	sync.Mutex
	at    time.Time
	cards []gopodLookupCard
}{}

var gopodLookupWordRE = regexp.MustCompile(`[^a-z0-9]+`)

// gopodLookupCards: the pool's cards (read again after a minute).
func gopodLookupCards() []gopodLookupCard {
	gopodLookupCache.Lock()
	defer gopodLookupCache.Unlock()
	if gopodLookupCache.cards != nil && time.Since(gopodLookupCache.at) < time.Minute {
		return gopodLookupCache.cards
	}
	var cards []gopodLookupCard
	add := func(path string) {
		base := filepath.Base(path)
		if !strings.HasSuffix(base, ".md") || strings.HasPrefix(base, "z") || strings.HasSuffix(base, ".local.md") {
			return
		}
		if c, ok := gopodLookupRead(path); ok {
			cards = append(cards, c)
		}
	}
	for _, f := range gopodBrobotsSettings.BrainChain.LookupFiles {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		path := filepath.Join(gopodRepoRoot(), f)
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			add(path)
			continue
		}
		_ = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
			if err == nil && d.IsDir() && (strings.HasPrefix(d.Name(), "_") || d.Name() == "locks") {
				return fs.SkipDir
			}
			if err == nil && !d.IsDir() {
				add(p)
			}
			return nil
		})
	}
	gopodLookupCache.cards, gopodLookupCache.at = cards, time.Now()
	return cards
}

// gopodLookupRead: a card - its words, then its first lines (a topic cluster's
// first section's points; a page's first paragraph).
func gopodLookupRead(path string) (gopodLookupCard, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return gopodLookupCard{}, false
	}
	c := gopodLookupCard{Path: path}
	inBody, sections := false, 0
	// A topic cluster opens "# title" then its head lines ("- **words:** ...");
	// anything else is a page (its "---" is a rule, not a head's end).
	var heads []string
	for _, l := range strings.Split(string(data), "\n") {
		if t := strings.TrimSpace(l); t != "" {
			heads = append(heads, t)
			if len(heads) == 2 {
				break
			}
		}
	}
	cluster := len(heads) == 2 && strings.HasPrefix(heads[0], "# ") && strings.HasPrefix(heads[1], "- **")
	var para []string
	summary := false
	for _, line := range strings.Split(string(data), "\n") {
		t := strings.TrimSpace(line)
		if !inBody {
			if k, _, ok := gopodTopicHeadKV(t); ok && k == "summary" && cluster {
				summary = true
			}
			// A card's summary head line is its lines (a chapter card's).
			if k, v, ok := gopodTopicHeadKV(t); ok && k == "summary" && cluster && strings.TrimSpace(v) != "" {
				c.Lines = []string{strings.TrimSpace(v)}
			}
			if k, v, ok := gopodTopicHeadKV(t); ok && k == "words" && cluster {
				for _, w := range strings.Split(v, ",") {
					if w = strings.TrimSpace(strings.ToLower(w)); w != "" {
						c.Words = append(c.Words, w)
					}
				}
			}
			if t == "---" && cluster {
				inBody = true
			}
			if t != "" && !strings.HasPrefix(t, "#") && !strings.HasPrefix(t, "-") && !strings.HasPrefix(t, ">") && !strings.HasPrefix(t, "|") && !strings.HasPrefix(t, "<") && !strings.HasPrefix(t, "!") {
				para = append(para, t)
			} else if len(para) > 0 && c.Lines == nil {
				c.Lines = []string{strings.Join(para, " ")}
			}
			continue
		}
		if strings.HasPrefix(t, "## ") {
			sections++
			continue
		}
		// Its first section's points (or those before any section), unless a summary said it.
		if m := gopodProfilePointRE.FindStringSubmatch(t); m != nil && sections <= 1 && !summary {
			c.Lines = append(c.Lines, m[1])
		}
	}
	if !cluster {
		// A page (no topic-cluster head): named by its file name too
		// ("GOPOD_YOURSELF.md" -> "gopod yourself"), its first paragraph its lines.
		name := strings.TrimSuffix(filepath.Base(path), ".md")
		if strings.EqualFold(name, "README") {
			name = filepath.Base(filepath.Dir(path))
		}
		c.Words = append([]string{strings.TrimSpace(gopodLookupWordRE.ReplaceAllString(strings.ToLower(name), " "))}, c.Words...)
		if len(c.Lines) == 0 && len(para) > 0 {
			c.Lines = []string{strings.Join(para, " ")}
		}
	}
	return c, len(c.Words) > 0
}

// gopodLookupStem: a word with its ending off ("booking" -> "book", "gigs" -> "gig").
func gopodLookupStem(w string) string {
	for _, end := range []string{"ing", "es", "ed", "s"} {
		if len(w) > len(end)+2 && strings.HasSuffix(w, end) {
			return strings.TrimSuffix(w, end)
		}
	}
	return w
}

var gopodLookupLittle = map[string]bool{"the": true, "and": true, "for": true, "you": true, "your": true, "our": true, "with": true, "who": true, "what": true, "how": true}

// gopodLookupTokens: text's words, stemmed.
func gopodLookupTokens(text string) []string {
	var out []string
	for _, w := range strings.Fields(gopodLookupWordRE.ReplaceAllString(strings.ToLower(text), " ")) {
		out = append(out, gopodLookupStem(w))
	}
	return out
}

// gopodLookup: the cards a question asks about - a card's phrase counts when its
// words come in the question in a row (endings aside), by its meaningful letters
// (not the little words, not lookup_skip's common ones); up to lookup_max cards,
// the best, and any as near it.
func gopodLookup(question string) []gopodLookupCard {
	q := " " + strings.Join(gopodLookupTokens(question), " ") + " "
	skip := map[string]bool{}
	for _, w := range gopodBrobotsSettings.BrainChain.LookupSkip {
		skip[gopodLookupStem(strings.ToLower(strings.TrimSpace(w)))] = true
	}
	type hit struct {
		card  gopodLookupCard
		score int
	}
	var hits []hit
	for _, c := range gopodLookupCards() {
		best := 0
		for _, w := range c.Words {
			toks := gopodLookupTokens(w)
			if len(toks) == 0 || !strings.Contains(q, " "+strings.Join(toks, " ")+" ") {
				continue
			}
			score := 0
			for _, t := range toks {
				if len(t) >= 3 && !skip[t] && !gopodLookupLittle[t] {
					score += len(t)
				}
			}
			if score >= 4 && score > best {
				best = score
			}
		}
		if best > 0 {
			hits = append(hits, hit{c, best})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].score > hits[j].score })
	most := gopodBrobotsSettings.BrainChain.LookupMax
	var out []gopodLookupCard
	for _, h := range hits {
		// The best card, and any as near its words (within 2 letters of it).
		if most > 0 && len(out) >= most || h.score < hits[0].score-2 {
			break
		}
		out = append(out, h.card)
	}
	return out
}

// gopodLookupFact: what the brain's FACT says from the cards - their first lines,
// at most brain_chain.lookup_chars each; and their names, for LOOKED UP.
func gopodLookupFact(question string) (fact string, sources []string) {
	chars := gopodBrobotsSettings.BrainChain.LookupChars
	var facts []string
	for _, c := range gopodLookup(question) {
		sources = append(sources, gopodLookupName(c.Path))
		t := strings.Join(c.Lines, " ")
		if chars > 0 && len(t) > chars {
			t = gopodCut(t, chars)
		}
		if t = strings.TrimSpace(gopodPersonaFill("", t)); t != "" {
			facts = append(facts, t)
		}
	}
	return strings.Join(facts, " "), sources
}

// gopodLookupName: a card as the bubble names it - its file, or its folder for a README.
func gopodLookupName(path string) string {
	if strings.EqualFold(filepath.Base(path), "README.md") && filepath.Dir(path) != gopodRepoRoot() {
		return filepath.Base(filepath.Dir(path)) + "/README.md"
	}
	return filepath.Base(path)
}

// ## EOF
