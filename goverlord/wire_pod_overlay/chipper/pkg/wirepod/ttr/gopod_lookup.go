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
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

type gopodLookupCard struct {
	Path  string
	Words []string
	Lines []string
	Owner string   // its "- **owner:**" head line - whose job its topic is (gopod_any_persona.go)
	Body  []string // every point and paragraph line, for the depth a question asks (gopodLookupFact)
	// 2026-11-18 (operator: "build the 'see also' multi-look-up"): its "see also"
	// cards' file names - looked up with it, one step out.
	SeeAlso []string
	// 2026-11-20 (operator: "Build 2 indexes ... index-2 an actual repo lookup of
	// whatever wordage is there"): its headings - the repo file's own topics.
	Headings []string
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
	// 2026-11-17 (operator: "Gap 3, privacy? Fix first"): only the public repo -
	// the files git tracks; a work file on disk (untracked, ignored) is never a
	// card. No list from git (git missing, not a repo): no cards, said in the log.
	tracked := gopodLookupTracked()
	add := func(path string) {
		base := filepath.Base(path)
		// 2026-11-20: a "-schema.md" is a blank template ("(to come)"), not a fact.
		if !strings.HasSuffix(base, ".md") || strings.HasPrefix(base, "z") || strings.HasSuffix(base, ".local.md") || strings.HasSuffix(base, "-schema.md") || !tracked[filepath.Clean(path)] {
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

// gopodLookupTracked: the repo's tracked files (git ls-files), by full path.
func gopodLookupTracked() map[string]bool {
	root := gopodRepoRoot()
	out, err := gopodAsOperator(exec.Command("git", "-C", root, "ls-files", "-z")).Output()
	tracked := map[string]bool{}
	if err != nil {
		logger.Println("GOPOD_LOOKUP no tracked-file list, no cards: " + err.Error())
		return tracked
	}
	for _, f := range strings.Split(string(out), "\x00") {
		if f != "" {
			tracked[filepath.Clean(filepath.Join(root, f))] = true
		}
	}
	return tracked
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
	// 2026-11-18: a plain-topped card ("- nickname: DOC") is a card too.
	cluster := len(heads) == 2 && strings.HasPrefix(heads[0], "# ") && (strings.HasPrefix(heads[1], "- **") || (strings.HasPrefix(heads[1], "- ") && strings.Contains(heads[1], ":")))
	var para []string
	summary := false
	for _, line := range strings.Split(string(data), "\n") {
		t := strings.TrimSpace(line)
		if h := strings.TrimSpace(strings.TrimLeft(t, "#")); strings.HasPrefix(t, "#") && h != "" && len(c.Headings) < 40 {
			c.Headings = append(c.Headings, strings.NewReplacer("**", "", "`", "").Replace(h))
		}
		if !inBody {
			if k, _, ok := gopodTopicHeadKV(t); ok && k == "summary" && cluster {
				summary = true
			}
			// A card's summary head line is its lines (a chapter card's).
			if k, v, ok := gopodTopicHeadKV(t); ok && k == "summary" && cluster && strings.TrimSpace(v) != "" {
				c.Lines = []string{strings.TrimSpace(v)}
			}
			// A cluster's title and its head's own lines (not its words, its ask or
			// its build notes) are lines too - a settings card says what it is.
			if cluster && strings.HasPrefix(t, "# ") {
				c.Body = append(c.Body, strings.TrimPrefix(t, "# "))
			}
			if k, v, ok := gopodTopicHeadKV(t); ok && cluster && k != "words" && k != "ask" && k != "about" && k != "owner" && k != "see also" && strings.TrimSpace(v) != "" {
				c.Body = append(c.Body, strings.TrimSpace(v))
			}
			if k, v, ok := gopodTopicHeadKV(t); ok && k == "see also" && cluster {
				for _, part := range strings.Split(v, ",") {
					name := strings.TrimSpace(strings.SplitN(strings.TrimSpace(part), " ", 2)[0])
					name = strings.TrimSuffix(strings.Trim(name, "()`"), ".md")
					if name = filepath.Base(name); name != "" && name != "." {
						c.SeeAlso = append(c.SeeAlso, name+".md")
					}
				}
			}
			if k, v, ok := gopodTopicHeadKV(t); ok && k == "owner" && cluster {
				c.Owner = strings.TrimSpace(v)
			}
			if k, v, ok := gopodTopicHeadKV(t); ok && (k == "people call me" || k == "people call us") && cluster {
				v = strings.ReplaceAll(v, " or ", ", ")
				for _, w := range strings.Split(v, ",") {
					if w = strings.TrimSpace(strings.ToLower(w)); w != "" {
						c.Words = append(c.Words, w)
					}
				}
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
			if !cluster && t != "" && !strings.HasPrefix(t, "#") && !strings.HasPrefix(t, "|") && !strings.HasPrefix(t, "<") && !strings.HasPrefix(t, "!") && t != "---" {
				c.Body = append(c.Body, strings.TrimLeft(t, "-*> "))
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
		if m := gopodProfilePointRE.FindStringSubmatch(t); m != nil {
			c.Body = append(c.Body, m[1])
		} else if t != "" && !strings.HasPrefix(t, "#") && !strings.HasPrefix(t, "|") && !strings.HasPrefix(t, "<") && !strings.HasPrefix(t, "!") && t != "---" {
			c.Body = append(c.Body, strings.TrimLeft(t, "-*> "))
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
	// 2026-11-20 (operator: "Build 2 indexes"): index 1, the words people say -
	// each topic's words point to its repo file ("rig" -> MOBILE_GEAR.md); index 2,
	// the repo's own words - a file's name and headings. A card either index
	// picks counts at its best score.
	at := map[string]int{}
	for i, h := range hits {
		at[filepath.Clean(h.card.Path)] = i
	}
	byPath := map[string]gopodLookupCard{}
	for _, c := range gopodLookupCards() {
		byPath[filepath.Clean(c.Path)] = c
	}
	bump := func(path string, score int) {
		c, ok := byPath[filepath.Clean(path)]
		if !ok || score <= 0 {
			return
		}
		if i, seen := at[filepath.Clean(path)]; seen {
			if score > hits[i].score {
				hits[i].score = score
			}
			return
		}
		at[filepath.Clean(path)] = len(hits)
		hits = append(hits, hit{c, score})
	}
	for _, v := range gopodLookupVocab() {
		if !strings.Contains(q, " "+strings.Join(v.tokens, " ")+" ") {
			continue
		}
		score := 0
		for _, t := range v.tokens {
			if !skip[t] && !gopodLookupLittle[t] {
				score += len(t)
			}
		}
		if score >= 3 {
			bump(v.path, score)
		}
	}
	// Index 2 only when the words people say found nothing.
	meaningful := map[string]bool{}
	if len(hits) == 0 {
		meaningful = gopodLookupMeaningful(question)
	}
	for _, c := range gopodLookupCards() {
		score := 0
		heads := " " + strings.Join(gopodLookupTokens(strings.TrimSuffix(filepath.Base(c.Path), ".md")+" "+strings.Join(c.Headings, " ")), " ") + " "
		for w := range meaningful {
			if strings.Contains(heads, " "+w+" ") {
				score += len(w)
			}
		}
		bump(c.Path, score)
	}
	// 2026-11-17 (operator: "Aim for full coverage"): no card's words said - the
	// cards whose lines say the most of the question's meaningful words (two or
	// more of them), by the letters matched.
	if len(hits) == 0 {
		words := gopodLookupMeaningful(question)
		for _, c := range gopodLookupCards() {
			seen := map[string]bool{}
			for _, l := range c.Body {
				for _, w := range gopodLookupTokens(l) {
					if words[w] {
						seen[w] = true
					}
				}
			}
			if len(seen) >= 2 {
				score := 0
				for w := range seen {
					score += len(w)
				}
				hits = append(hits, hit{c, score})
			}
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

// gopodLookupFact: what the brain's FACT says from the cards - each card's first
// lines, then (2026-11-17, operator: "Gap 2, depth? Add sufficient depth without
// repeating (fix brains, the selves within the group)") its other lines that say
// the question's words, no line twice, at most brain_chain.lookup_chars a card;
// and their names, for LOOKED UP. self's own card and the company's card are
// already in his prompt - never looked up again.
func gopodLookupFact(self, question string) (fact string, sources []string) {
	chars := gopodBrobotsSettings.BrainChain.LookupChars
	skipCard := map[string]bool{}
	if self != "" {
		if f := gopodPersonaProfileFile(self); f != "" {
			skipCard[filepath.Clean(f)] = true
		}
	}
	if f := gopodTopicTreeFind("math-aftermath-news-channel.md"); f != "" {
		skipCard[filepath.Clean(f)] = true
	}
	words := gopodLookupMeaningful(question)
	var facts []string
	said := map[string]bool{}
	byName := map[string]gopodLookupCard{}
	for _, c := range gopodLookupCards() {
		byName[strings.ToLower(filepath.Base(c.Path))] = c
	}
	listed := map[string]bool{}
	// one card's lines: its first lines, then those that say the question's words;
	// at most chars, no line said twice.
	cardText := func(c gopodLookupCard, chars int, swept bool) string {
		var lines []string
		add := func(l string) {
			l = strings.TrimSpace(strings.NewReplacer("**", "", "`", "").Replace(l))
			key := strings.ToLower(l)
			// A card label with nothing filled in ("My informal name (nickname): .").
			l = strings.TrimRight(l, ". ")
			// "My job title: X" - the value, not the label; an empty value ("{job}"
			// unfilled) is dropped.
			if v, ok := strings.CutPrefix(l, "My job title:"); ok {
				l = strings.TrimLeft(strings.TrimSpace(v), "- ")
			}
			if gopodLookupEmptyLabelRE.MatchString(l) || gopodLookupNameLabelRE.MatchString(l) {
				return
			}
			// A setting ("label: ANY PERSONA") or a dated build note is the
			// machine's, not a fact to say.
			if l == "" || said[key] || gopodLookupMachineRE.MatchString(l) {
				return
			}
			said[key] = true
			lines = append(lines, l)
		}
		if !swept {
			for _, l := range c.Lines {
				add(l)
			}
			// 2026-11-20: the wide look - the repo file's own topics, its headings.
			if len(c.Headings) > 1 {
				h := c.Headings
				if len(h) > 8 {
					h = h[:8]
				}
				add("Its topics: " + strings.Join(h, ", "))
			}
		}
		for _, l := range c.Body {
			for _, w := range gopodLookupTokens(l) {
				if words[w] {
					add(l)
					break
				}
			}
		}
		if len(lines) == 0 && !swept {
			// Nothing here to say (a settings card): its title and head, then.
			for _, l := range c.Body {
				if len(lines) < 2 {
					add(l)
				}
			}
		}
		if len(lines) == 0 {
			return ""
		}
		t := strings.NewReplacer("**", "", "`", "").Replace(strings.Join(lines, ". ") + ".")
		if chars > 0 && len(t) > chars {
			t = gopodCut(t, chars)
		}
		return strings.TrimSpace(gopodPersonaFill("", t))
	}
	take := func(c gopodLookupCard, chars int, swept bool) {
		name := filepath.Clean(c.Path)
		if listed[name] {
			return
		}
		listed[name] = true
		// Every card looked up is named (operator: "all files looked up are
		// listed"); his own card and the company's add no words - they're in his prompt.
		sources = append(sources, gopodLookupName(c.Path))
		if skipCard[name] || chars < 0 { // listed only: the link budget is spent
			return
		}
		if t := cardText(c, chars, swept); t != "" {
			facts = append(facts, t)
		}
	}
	most := gopodBrobotsSettings.BrainChain.LookupChainMax
	// 2026-11-18 (operator: "look ups load to the end of the chain ... then sweep
	// for information required and how to answer the question"): the picked
	// cards, then their "see also" cards, theirs in turn, to the chain's end - each
	// card once, at most lookup_chain_max of them; a linked card gives only its
	// lines that say the question's words.
	var queue []gopodLookupCard
	picked := gopodLookup(question)
	// 2026-11-18 (operator: "Since GOPOD game is essentially a git repo tour, have
	// 'duh' responses, default to look up the README.md file ... gravitate to GOPOD
	// Yourself * I don't know. GOPOD Yourself!"): a question no card matches gets
	// the repo's README (brain_chain.lookup_default, a path from the repo's top),
	// and its line (lookup_default_line).
	if len(picked) == 0 {
		if d := strings.TrimSpace(gopodBrobotsSettings.BrainChain.LookupDefault); d != "" {
			want := filepath.Clean(filepath.Join(gopodRepoRoot(), d))
			for _, c := range gopodLookupCards() {
				if filepath.Clean(c.Path) == want {
					take(c, chars, false)
					if line := strings.TrimSpace(gopodBrobotsSettings.BrainChain.LookupDefaultLine); line != "" {
						facts = append(facts, line)
					}
					break
				}
			}
		}
	}
	// Two (or more) cards tie for the question: both briefly, his job's part only.
	if len(picked) >= 2 {
		if line := strings.TrimSpace(gopodBrobotsSettings.BrainChain.TwoMatchesLine); line != "" {
			facts = append(facts, line)
		}
	}
	for _, c := range picked {
		take(c, chars, false)
		queue = append(queue, c)
	}
	for len(queue) > 0 && (most <= 0 || len(listed) < most) {
		c := queue[0]
		queue = queue[1:]
		for _, l := range c.SeeAlso {
			if most > 0 && len(listed) >= most {
				break
			}
			if lc, ok := byName[strings.ToLower(l)]; ok && !listed[filepath.Clean(lc.Path)] {
				// 2026-11-20 (operator: "CAMEO was a wide/deep probe"; live: fed nine
				// linked cards' "I" lines, she said "Goverlord (aka ME)"): the wide
				// answer is the card asked about - its "see also" cards are listed
				// (LOOKED UP), not said; asking about one of them is the deep look.
				take(lc, -1, true)
				queue = append(queue, lc)
			}
		}
	}
	return strings.Join(facts, " "), sources
}

// gopodLookupMachineRE: a line that's a setting (a key: value), a build note
// (dated, or quoting the operator) or a path - not a fact to say.
var gopodLookupMachineRE = regexp.MustCompile(`^_?[a-z][a-z0-9_.\[\]]*:\s|\b20\d\d-\d\d-\d\d\b|operator:|[\w-]+/[\w./-]+\.(md|go|js|json|py)\b`)

// gopodLookupEmptyLabelRE: a label line whose value is empty.
var gopodLookupEmptyLabelRE = regexp.MustCompile(`^[^:]{1,40}:\s*\.?\s*$`)

// gopodLookupNameLabelRE: a card's own name labels - said for him, not a fact.
var gopodLookupNameLabelRE = regexp.MustCompile(`^My (informal name|name)\b`)

// gopodLookupMeaningful: a question's words worth matching a line on (stemmed;
// not the little words, not lookup_skip's).
func gopodLookupMeaningful(question string) map[string]bool {
	skip := map[string]bool{}
	for _, w := range gopodBrobotsSettings.BrainChain.LookupSkip {
		skip[gopodLookupStem(strings.ToLower(strings.TrimSpace(w)))] = true
	}
	out := map[string]bool{}
	for _, t := range gopodLookupTokens(question) {
		if len(t) >= 4 && !skip[t] && !gopodLookupLittle[t] {
			out[t] = true
		}
	}
	return out
}

// gopodLookupName: a card as the bubble names it - its file, or its folder for a README.
func gopodLookupName(path string) string {
	if strings.EqualFold(filepath.Base(path), "README.md") && filepath.Dir(path) != gopodRepoRoot() {
		return filepath.Base(filepath.Dir(path)) + "/README.md"
	}
	return filepath.Base(path)
}

// gopodNonsenseLine (2026-11-19, operator: "I input for [persona-name]:
// rgvthrtgh$^$%^gfdgdrfg * the reaction should match the non-sense * like; what is
// 'rgvthrtgh$^$%^gfdgdrfg'? I don't know, let's start at README.md; rather than
// engage the nonsense query"): a line with none of brain_chain.real_words and no
// card its words pick gets brain_chain.nonsense_line, said as is; "" otherwise.
func gopodNonsenseLine(text string) string {
	bc := gopodBrobotsSettings.BrainChain
	t := strings.TrimSpace(text)
	if bc.NonsenseLine == "" || t == "" || gopodIsIntro(t) || len(gopodLookup(t)) > 0 {
		return ""
	}
	words := strings.FieldsFunc(strings.ToLower(t), func(r rune) bool { return !unicode.IsLetter(r) && r != '\'' })
	for _, w := range words {
		for _, real := range strings.Split(bc.RealWords, ",") {
			if w == strings.ToLower(strings.TrimSpace(real)) {
				return ""
			}
		}
	}
	return strings.ReplaceAll(bc.NonsenseLine, "{asked}", t)
}

// gopodLookupVocab: index 1 - each topic's words (the words people say) and
// the repo file its topic tells (its source, else its own card), read with the cards.
type gopodVocabEntry struct {
	tokens []string
	path   string
}

var gopodVocabCache = struct {
	sync.Mutex
	at      time.Time
	entries []gopodVocabEntry
}{}

func gopodLookupVocab() []gopodVocabEntry {
	gopodVocabCache.Lock()
	defer gopodVocabCache.Unlock()
	if gopodVocabCache.entries != nil && time.Since(gopodVocabCache.at) < time.Minute {
		return gopodVocabCache.entries
	}
	// 2026-11-20 (operator: "from repo-index 2, generate Index 1, the words people
	// say"): index 1 is its own file, written from index 2 (gopod_repo_index.go).
	var out []gopodVocabEntry
	index := gopodReadWordsIndex()
	var rels []string
	for rel := range index {
		rels = append(rels, rel)
	}
	sort.Strings(rels) // a fixed order: ties land the same way every time
	for _, rel := range rels {
		words := index[rel]
		path := filepath.Join(gopodRepoRoot(), rel)
		for _, w := range words {
			if toks := gopodLookupTokens(w); len(toks) > 0 {
				out = append(out, gopodVocabEntry{toks, path})
			}
		}
	}
	gopodVocabCache.entries, gopodVocabCache.at = out, time.Now()
	return out
}

// ## EOF
