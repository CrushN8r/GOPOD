// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_persona_chains.go
package wirepod_ttr

import (
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

type gopodPersonaChains struct {
	Share  int               `json:"share"`
	Chains []gopodDocBSChain `json:"chains"`
}

var gopodPersonaChainSlugRE = regexp.MustCompile(`[^a-z0-9]+`)

// gopodPersonaChainDir: a persona's pool folder ("Brobot 1" -> brobot-1).
func gopodPersonaChainDir(who string) string {
	folder := strings.TrimSpace(gopodBrobotsSettings.PersonaChains.Folder)
	if folder == "" {
		return ""
	}
	slug := strings.Trim(gopodPersonaChainSlugRE.ReplaceAllString(strings.ToLower(who), "-"), "-")
	folder = gopodTopicPath(folder) // moved, found by its folder name
	return filepath.Join(folder, slug)
}

var gopodPoolDecks = struct {
	sync.Mutex
	left map[string][]string
}{left: map[string][]string{}}

// gopodPoolPick: the next line of dir's pool - _18 over _clean by the roast level, the private
// z-copy first; a shuffled deck, no repeat until used up.
func gopodPoolPick(dir, pool string) string {
	file := func(name string) string {
		if z := filepath.Join(dir, "z"+name+".txt"); gopodFileExists(z) {
			return z
		}
		return filepath.Join(dir, name+".txt")
	}
	name := pool
	if _, err := os.Stat(file(pool + "_clean")); err == nil {
		name = pool + "_clean"
		if n, ok := gopodRoastLevel(); ok && gopodBrobotsSettings.RoastDial.Levels[n].HumanSwears {
			name = pool + "_18"
		}
	}
	path := file(name)
	gopodPoolDecks.Lock()
	defer gopodPoolDecks.Unlock()
	deck := gopodPoolDecks.left[path]
	if len(deck) == 0 {
		data, err := os.ReadFile(path)
		if err != nil {
			return ""
		}
		for _, l := range strings.Split(string(data), "\n") {
			if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "#") {
				deck = append(deck, l)
			}
		}
		rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	}
	if len(deck) == 0 {
		return ""
	}
	gopodPoolDecks.left[path] = deck[1:]
	return deck[0]
}

var gopodPersonaChainDecks = struct {
	sync.Mutex
	left map[string][]gopodDocBSChain
}{left: map[string][]gopodDocBSChain{}}

// gopodPersonaChainPick: who's next chain the dial allows, from its deck.
func gopodPersonaChainPick(who string) (gopodDocBSChain, bool) {
	pc, ok := gopodBrobotsSettings.PersonaChains.Personas[who]
	if !ok {
		return gopodDocBSChain{}, false
	}
	var allowed []gopodDocBSChain
	for _, c := range pc.Chains {
		if c.MinShare <= pc.Share && len(c.Beats) > 0 {
			allowed = append(allowed, c)
		}
	}
	if len(allowed) == 0 {
		return gopodDocBSChain{}, false
	}
	gopodPersonaChainDecks.Lock()
	defer gopodPersonaChainDecks.Unlock()
	deck := gopodPersonaChainDecks.left[who]
	if len(deck) == 0 {
		deck = append([]gopodDocBSChain(nil), allowed...)
		rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	}
	gopodPersonaChainDecks.left[who] = deck[1:]
	return deck[0], true
}

// gopodPersonaChainLines: one answer's chain for who - the lines before its reply and after it ("",
// "" when it isn't one, by its dial or its pools).
func gopodPersonaChainLines(who string) (before, after string) {
	pc, ok := gopodBrobotsSettings.PersonaChains.Personas[who]
	if !ok || pc.Share <= 0 || rand.Intn(100) >= pc.Share {
		return "", ""
	}
	chain, ok := gopodPersonaChainPick(who)
	dir := gopodPersonaChainDir(who)
	if !ok || dir == "" {
		return "", ""
	}
	var pre, post []string
	side := &pre
	for _, beat := range chain.Beats {
		if strings.TrimSpace(beat) == "reply" {
			side = &post
			continue
		}
		pools := strings.Split(beat, "|")
		if line := strings.TrimSpace(gopodPoolPick(dir, strings.TrimSpace(pools[rand.Intn(len(pools))]))); line != "" {
			*side = append(*side, line)
		}
	}
	if len(pre)+len(post) > 0 {
		logger.Println("GOPOD_PERSONA_CHAIN " + who + ": " + chain.Name)
	}
	return strings.Join(pre, gopodThoughtGap()), strings.Join(post, gopodThoughtGap())
}

// One answer's chain, decided once (the before part picks it, the close part uses it), the same way
// DocBS's decision is kept.
var gopodPersonaChainFor sync.Map // esn + question -> [2]string

func gopodPersonaChainFor2(esn, question string) [2]string {
	key := esn + "\x00" + question
	if v, ok := gopodPersonaChainFor.Load(key); ok {
		return v.([2]string)
	}
	var lines [2]string
	if !gopodDocBSOn(esn, question) {
		lines[0], lines[1] = gopodPersonaChainLines(gopodBrobotName(esn))
	}
	v, loaded := gopodPersonaChainFor.LoadOrStore(key, lines)
	if !loaded {
		time.AfterFunc(2*time.Minute, func() { gopodPersonaChainFor.Delete(key) })
	}
	return v.([2]string)
}

// gopodPersonaChainWrap: a Gomad's reply inside its chain.
func gopodPersonaChainWrap(who, reply string) string {
	before, after := gopodPersonaChainLines(who)
	if before != "" {
		reply = before + gopodThoughtGap() + reply
	}
	if after != "" {
		reply = reply + gopodThoughtGap() + after
	}
	return reply
}

// ## EOF
