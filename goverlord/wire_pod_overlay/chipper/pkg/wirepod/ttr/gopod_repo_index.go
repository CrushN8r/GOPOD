// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_repo_index.go
package wirepod_ttr

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func gopodRepoIndexPath() string {
	return filepath.Join(gopodChipperRoot(), "gopod_probes", "locks", "repo_index.md")
}

func gopodWordsIndexPath() string {
	return filepath.Join(gopodChipperRoot(), "gopod_probes", "locks", "words_index.md")
}

// gopodCardRel: a card's path from the repo's top.
func gopodCardRel(path string) string {
	rel, err := filepath.Rel(gopodRepoRoot(), path)
	if err != nil {
		return path
	}
	return rel
}

// gopodRepoIndexText: index 2 - each card and its headings, in path order.
func gopodRepoIndexText() string {
	cards := gopodLookupCards()
	sort.Slice(cards, func(i, j int) bool { return gopodCardRel(cards[i].Path) < gopodCardRel(cards[j].Path) })
	var b strings.Builder
	b.WriteString("# Repo index 2 - the repo's own words (written from the repo; GOPOD_WRITE_INDEX=1 relocks it)\n")
	for _, c := range cards {
		b.WriteString("\n## " + gopodCardRel(c.Path) + "\n")
		for _, h := range c.Headings {
			b.WriteString("- " + h + "\n")
		}
	}
	return b.String()
}

// gopodHeadingPhrase: a heading as words people say - small letters, no numbering, no brackets; ""
// when too long to be said (over five words).
func gopodHeadingPhrase(h string) string {
	if i := strings.Index(h, "("); i >= 0 && strings.ContainsAny(h[i:], "._") {
		return ""
	}
	h = strings.ToLower(h)
	if i := strings.IndexAny(h, "(["); i > 0 {
		h = h[:i]
	}
	h = strings.TrimSpace(gopodLookupWordRE.ReplaceAllString(h, " "))
	words := strings.Fields(h)
	for len(words) > 0 && (strings.Trim(words[0], "0123456789") == "" || gopodLookupLittle[words[0]] || words[0] == "a" || words[0] == "an") {
		words = words[1:]
	}
	if len(words) == 0 || len(words) > 5 {
		return ""
	}
	return strings.Join(words, " ")
}

// gopodWordsIndexText: index 1, written from index 2 and the gold topic words.
func gopodWordsIndexText() string {
	gold := map[string][]string{}
	chapters, always, _ := gopodReadChaptersFolder()
	topics := append([]gopodTopicFile{}, always...)
	for _, ch := range chapters {
		topics = append(topics, ch.Topics...)
	}
	for _, t := range topics {
		var paths []string
		for _, s := range t.Sources {
			if s = strings.TrimSpace(strings.SplitN(s, "#", 2)[0]); s != "" {
				paths = append(paths, s)
			}
		}
		if len(paths) == 0 {
			paths = []string{gopodCardRel(strings.SplitN(t.Path, "#", 2)[0])}
		}
		for _, p := range paths {
			gold[p] = append(gold[p], t.Words...)
		}
	}
	cards := gopodLookupCards()
	sort.Slice(cards, func(i, j int) bool { return gopodCardRel(cards[i].Path) < gopodCardRel(cards[j].Path) })
	shared := map[string]int{}
	for _, c := range cards {
		once := map[string]bool{}
		for _, h := range c.Headings {
			if p := gopodHeadingPhrase(h); p != "" && !once[p] {
				once[p] = true
				shared[p]++
			}
		}
	}
	var b strings.Builder
	b.WriteString("# Index 1 - the words people say (written from repo index 2 and the gold topic words; GOPOD_WRITE_INDEX=1 writes it again)\n")
	for _, c := range cards {
		rel := gopodCardRel(c.Path)
		seen := map[string]bool{}
		var words []string
		add := func(w string) {
			w = strings.TrimSpace(strings.ToLower(w))
			if w != "" && !seen[w] && !strings.ContainsAny(w, ",|") {
				seen[w] = true
				words = append(words, w)
			}
		}
		name := strings.TrimSuffix(filepath.Base(rel), ".md")
		if strings.EqualFold(name, "README") || strings.EqualFold(name, "story") || strings.EqualFold(name, "chapter") {
			name = filepath.Base(filepath.Dir(rel))
		}
		add(gopodHeadingPhrase(strings.NewReplacer("_", " ", "-", " ").Replace(name)))
		for _, h := range c.Headings {
			if p := gopodHeadingPhrase(h); shared[p] < 3 {
				add(p)
			}
		}
		for _, w := range gold[rel] {
			add(w)
		}
		for _, w := range gopodSameWordsExpand(words) {
			add(w)
		}
		if len(words) == 0 {
			continue
		}
		b.WriteString("\n## " + rel + "\n- words: " + strings.Join(words, ", ") + "\n")
	}
	return b.String()
}

// gopodReadWordsIndex: index 1 from its file - each card's words.
func gopodReadWordsIndex() map[string][]string {
	data, err := os.ReadFile(gopodWordsIndexPath())
	if err != nil {
		return nil
	}
	out := map[string][]string{}
	cur := ""
	for _, line := range strings.Split(string(data), "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "## ") {
			cur = strings.TrimSpace(strings.TrimPrefix(t, "## "))
			continue
		}
		if v, ok := strings.CutPrefix(t, "- words:"); ok && cur != "" {
			for _, w := range strings.Split(v, ",") {
				if w = strings.TrimSpace(w); w != "" {
					out[cur] = append(out[cur], w)
				}
			}
		}
	}
	return out
}

// The index questions: one question per card, from its first words in index 1 - "Tell me about the
// rig" -> the card it should find.
func gopodIndexQuestionsPath() string {
	return filepath.Join(gopodChipperRoot(), "gopod_probes", "locks", "index_questions.md")
}

func gopodIndexQuestionsText() string {
	words := gopodReadWordsIndex()
	var rels []string
	for rel := range words {
		rels = append(rels, rel)
	}
	sort.Strings(rels)
	var b strings.Builder
	b.WriteString("# Index questions - one per card, from index 1 (GOPOD_WRITE_INDEX=1 writes them again)\n# question | the card it should find\n")
	for _, rel := range rels {
		ws := words[rel]
		if len(ws) == 0 {
			continue
		}
		b.WriteString("Tell me about " + ws[0] + " | " + rel + "\n")
	}
	return b.String()
}

// gopodIndexQuestionsScore: how many index questions find their own card.
func gopodIndexQuestionsScore() (right, total int, misses []string) {
	data, err := os.ReadFile(gopodIndexQuestionsPath())
	if err != nil {
		return 0, 0, []string{err.Error()}
	}
	for _, line := range strings.Split(string(data), "\n") {
		q, want, ok := strings.Cut(strings.TrimSpace(line), " | ")
		if !ok || strings.HasPrefix(line, "#") {
			continue
		}
		total++
		found := false
		for _, c := range gopodLookup(q) {
			if gopodCardRel(c.Path) == strings.TrimSpace(want) {
				found = true
			}
		}
		if found {
			right++
		} else {
			misses = append(misses, q)
		}
	}
	return right, total, misses
}

// gopodSameWords: brain_chain.same_words, set as the settings load (the topics are read while they
// load - gopodApplyChaptersFolder).
var gopodSameWords string

// gopodSameWordsExpand: words plus every word that means the same (brain_chain.same_words) - "rig"
// brings "gear", "equipment", "junk".
func gopodSameWordsExpand(words []string) []string {
	out := append([]string{}, words...)
	seen := map[string]bool{}
	for _, w := range words {
		seen[strings.ToLower(strings.TrimSpace(w))] = true
	}
	for _, line := range strings.Split(gopodSameWords, "/") {
		var group []string
		hit := false
		for _, g := range strings.Split(line, ",") {
			if g = strings.ToLower(strings.TrimSpace(g)); g != "" {
				group = append(group, g)
				hit = hit || seen[g]
			}
		}
		if !hit {
			continue
		}
		for _, g := range group {
			if !seen[g] {
				seen[g] = true
				out = append(out, g)
			}
		}
	}
	return out
}

// ## EOF
