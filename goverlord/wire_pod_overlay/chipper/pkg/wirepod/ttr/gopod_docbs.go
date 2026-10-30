// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_docbs.go
package wirepod_ttr

// DocBS - Doc's brain wiring (2026-10-29, operator: "Please build it as Doc's
// answer shape for Explain the Math lines * This is Doc's personality, not just
// for the livestream Explain the Math lines * This is 100% Doc's brain wiring *
// This is 100% Doc BS (LMAO! You get it now?) * from brobot 1 > dial > 100%
// Doc's brain wiring"; framework from his own takes,
// gopod_notes/DOCBS_EQUATION_ROAST_FRAMEWORK_001.md).
//
// Brobot 1's answer, when Doc's share of his stereo dial says so (the Doc & Pip
// pair's dial, or docbs.share): the answer chain's docbs part puts the equation
// roast in front of his reply - opener, lead-in with his nickname for the asker,
// the equation, the plug-in, the error, the hidden multiplier, the corrected
// formula, each its own thought bubble - and docbs_close ends it: "The math has
// spoken. And it's ...". His reply (the value point, his colour) is the kill shot
// in the middle, so the game goes on as before. The code builds the frame from
// the pool files (docbs.folder, one per beat, shuffled, never repeated until the
// deck runs out; _clean / _18 by the roast dial); gemma fills only the small
// slots - the equation's name, acronym, its three things, a nickname - in one
// short call. An "Explain the Math" line the math engine can't work out always
// gets it; a computable one keeps its exact steps.

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

var gopodDocBSDecks = struct {
	sync.Mutex
	left map[string][]string
}{left: map[string][]string{}}

// gopodDocBSPick: the next line of a pool file - a shuffled deck, never the same
// line twice until it's used up. The roast level picks _18 over _clean.
func gopodDocBSPick(pool string) string {
	name := pool
	if _, err := os.Stat(gopodDocBSFile(pool + "_clean")); err == nil {
		name = pool + "_clean"
		if n, ok := gopodRoastLevel(); ok && gopodBrobotsSettings.RoastDial.Levels[n].HumanSwears {
			name = pool + "_18"
		}
	}
	gopodDocBSDecks.Lock()
	defer gopodDocBSDecks.Unlock()
	deck := gopodDocBSDecks.left[name]
	if len(deck) == 0 {
		data, err := os.ReadFile(gopodDocBSFile(name))
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
	line := deck[0]
	gopodDocBSDecks.left[name] = deck[1:]
	return line
}

// gopodDocBSFile: a pool file - its private, uncensored z-copy when it's there
// (operator 2026-10-29: "Public GOPOD: star '*' out expletives ... like a corner
// store used to handle adult magazines"; the zKnobs.json pattern), else the
// public one, expletives starred.
func gopodDocBSFile(name string) string {
	dir := filepath.Join(gopodChipperRoot(), gopodBrobotsSettings.DocBS.Folder)
	if z := filepath.Join(dir, "z"+name+".txt"); gopodFileExists(z) {
		return z
	}
	return filepath.Join(dir, name+".txt")
}


// gopodDocBSShare: Doc's share of Brobot 1's dial, 0-100 (docbs.share when set).
func gopodDocBSShare() int {
	if s := gopodBrobotsSettings.DocBS.Share; s >= 0 {
		return s
	}
	if pair, _, ok := gopodHostPairOf("Brobot 1"); ok {
		return pair.Dial
	}
	return 0
}

var gopodDocBSOnFor sync.Map // esn + question -> bool, for this answer

// gopodDocBSOn: this answer of Brobot 1's is DocBS - decided once per answer.
func gopodDocBSOn(esn, question string) bool {
	if gopodBrobotName(esn) != "Brobot 1" || strings.TrimSpace(gopodBrobotsSettings.DocBS.Folder) == "" || gopodIsThanks(question) {
		return false
	}
	key := esn + "\x00" + question
	if v, ok := gopodDocBSOnFor.Load(key); ok {
		return v.(bool)
	}
	on := rand.Intn(100) < gopodDocBSShare()
	// An "Explain the Math" line the math engine can't work out: always.
	if _, explain := gopodMathExplain(question); explain {
		if _, _, computable := gopodMathParts(question); !computable {
			on = true
		}
	}
	v, loaded := gopodDocBSOnFor.LoadOrStore(key, on)
	if !loaded {
		// One answer's decision: cleared when the answer is shaped (the docbs
		// part), and in any case a little later - never kept for the next time
		// the same words are asked.
		time.AfterFunc(2*time.Minute, func() { gopodDocBSOnFor.Delete(key) })
	}
	return v.(bool)
}

// gopodDocBSParseSlots: "Name | ACR | a | b | c | nickname" from gemma's reply
// (its markdown and end marks out), nil when unreadable.
func gopodDocBSParseSlots(reply string) []string {
	for _, line := range strings.Split(reply, "\n") {
		f := strings.Split(line, "|")
		if len(f) < 6 {
			continue
		}
		ok := true
		for i := range f {
			// gemma's markdown and end marks out ("**EUD**", "Seaworthiness.").
			f[i] = strings.Trim(strings.TrimSpace(strings.NewReplacer("*", "", "`", "", "_", " ").Replace(f[i])), "\"' .;:,")
			if f[i] == "" || len(f[i]) > 60 {
				ok = false
			}
		}
		if ok {
			return f[:6]
		}
	}
	return nil
}

// gopodDocBSSlots: the equation's name, acronym, three things and a nickname -
// gemma's, asked ahead when the question came (gopodDocBSAskAhead), else asked
// now, else docbs.fallback.
func gopodDocBSSlots(question string) []string {
	fallback := gopodDocBSParseSlots(gopodBrobotsSettings.DocBS.Fallback)
	tpl := strings.TrimSpace(gopodBrobotsSettings.DocBS.SlotPrompt)
	if tpl == "" {
		return fallback
	}
	reply := gopodAskAheadGet("docbs-slots|"+question, 20*time.Second)
	if reply == "" {
		reply = gopodOllamaAsk(strings.ReplaceAll(tpl, "{question}", question), gopodBrobotsSettings.DocBS.NumPredict, 0.9)
	}
	if s := gopodDocBSParseSlots(reply); s != nil {
		logger.Println("BROBOT_DOCBS slots: " + strings.Join(s, " | "))
		return s
	}
	if reply != "" {
		logger.Println("BROBOT_DOCBS slots unreadable, fallback: " + gopodCut(reply, 120))
	}
	return fallback
}

// gopodDocBSAskAhead (2026-10-29, ask ahead): Brobot 1's answer going DocBS -
// its equation ask and its colour ask start now, while the line is read out and
// the bubbles before his play.
func gopodDocBSAskAhead(esn, question string) {
	if !gopodDocBSOn(esn, question) {
		return
	}
	db := gopodBrobotsSettings.DocBS
	if t := strings.TrimSpace(db.SlotPrompt); t != "" {
		gopodAskAhead("docbs-slots|"+question, strings.ReplaceAll(t, "{question}", question), db.NumPredict)
	}
	if t := gopodDocBSColourPrompt(); t != "" {
		gopodAskAhead("docbs-colour|"+question, strings.ReplaceAll(t, "{question}", question), db.NumPredict)
	}
}

// gopodDocBSColourPrompt: the colour ask for the roast level (18+ levels: colour_prompt_18).
func gopodDocBSColourPrompt() string {
	db := gopodBrobotsSettings.DocBS
	if n, ok := gopodRoastLevel(); ok && gopodBrobotsSettings.RoastDial.Levels[n].HumanSwears && strings.TrimSpace(db.ColourPrompt18) != "" {
		return strings.TrimSpace(db.ColourPrompt18)
	}
	return strings.TrimSpace(db.ColourPrompt)
}

// gopodDocBSColour: gemma's one escalated-metaphor line for this question ("how
// would you colour this part?"), asked ahead; "" when there's none.
func gopodDocBSColour(question string) string {
	line := gopodAskAheadGet("docbs-colour|"+question, 15*time.Second)
	line = strings.TrimSpace(strings.Split(line, "\n")[0])
	line = strings.Trim(strings.NewReplacer("*", "", "`", "").Replace(line), "\" ")
	if len(line) > 220 || gopodAnimationMarkerPattern.MatchString(line) {
		return ""
	}
	return line
}

var gopodDocBSTokenRE = regexp.MustCompile(`\{[a-z]+\}`)

// gopodDocBSFrame: the beats before his reply, each its own thought bubble.
func gopodDocBSFrame(question string) string {
	s := gopodDocBSSlots(question)
	if s == nil {
		return ""
	}
	mult := strings.SplitN(gopodDocBSPick("multipliers"), "|", 2)
	if len(mult) < 2 {
		mult = []string{"Delusion Coefficient", "DC"}
	}
	vals := map[string]string{
		"name": s[0], "acr": s[1], "a": s[2], "b": s[3], "c": s[4], "nickname": s[5],
		"va": gopodDocBSPick("constants_big"), "vb": gopodDocBSPick("constants_big"), "vc": gopodDocBSPick("constants_tiny"),
		"multiplier": mult[0], "ma": mult[1], "number": gopodDocBSPick("numbers"), "result": gopodDocBSPick("results"),
		"error": gopodDocBSPick("errors"),
	}
	fill := func(line string) string {
		return gopodDocBSTokenRE.ReplaceAllStringFunc(line, func(t string) string {
			if v, ok := vals[strings.Trim(t, "{}")]; ok {
				return v
			}
			return t
		})
	}
	chain := gopodDocBSChainPick()
	logger.Println("BROBOT_DOCBS chain: " + chain.Name)
	var beats []string
	for _, beat := range chain.Beats {
		pools := strings.Split(beat, "|")
		pool := strings.TrimSpace(pools[rand.Intn(len(pools))])
		var line string
		if pool == "colour" {
			// gemma's colour for this one (asked ahead): the escalated metaphor.
			line = gopodDocBSColour(question)
		} else {
			line = strings.TrimSpace(fill(gopodDocBSPick(pool)))
		}
		if line != "" {
			beats = append(beats, line)
		}
	}
	return strings.Join(beats, gopodThoughtGap())
}

// gopodDocBSChain: one DocBS chain - its beats before his reply, and the dial it needs.
type gopodDocBSChain struct {
	Name     string   `json:"name"`
	MinShare int      `json:"min_share"`
	Beats    []string `json:"beats"`
}

// gopodDocBSChainDefault: the equation roast, the chain DocBS had before the chains.
var gopodDocBSChainDefault = gopodDocBSChain{Name: "equation roast", Beats: []string{"openers", "leadins", "equations", "plugins", "breaks", "hiddens", "corrected", "colour"}}

// gopodDocBSChainPick (2026-10-30, the Master Library): the next chain from a
// shuffled deck of the chains the dial allows (min_share <= share), never the
// same one twice until the deck runs out.
func gopodDocBSChainPick() gopodDocBSChain {
	share := gopodDocBSShare()
	var allowed []gopodDocBSChain
	for _, c := range gopodBrobotsSettings.DocBS.Chains {
		if c.MinShare <= share && len(c.Beats) > 0 {
			allowed = append(allowed, c)
		}
	}
	if len(allowed) == 0 {
		return gopodDocBSChainDefault
	}
	gopodDocBSChains.Lock()
	defer gopodDocBSChains.Unlock()
	if gopodDocBSChains.share != share || len(gopodDocBSChains.left) == 0 {
		gopodDocBSChains.share = share
		gopodDocBSChains.left = append([]gopodDocBSChain(nil), allowed...)
		rand.Shuffle(len(gopodDocBSChains.left), func(i, j int) {
			gopodDocBSChains.left[i], gopodDocBSChains.left[j] = gopodDocBSChains.left[j], gopodDocBSChains.left[i]
		})
	}
	c := gopodDocBSChains.left[0]
	gopodDocBSChains.left = gopodDocBSChains.left[1:]
	return c
}

var gopodDocBSChains = struct {
	sync.Mutex
	share int
	left  []gopodDocBSChain
}{}

// gopodDocBSClose: the ending line ("The math has spoken. And it's ...").
func gopodDocBSClose() string {
	return strings.TrimSpace(strings.ReplaceAll(gopodDocBSPick("endings"), "{closer}", gopodDocBSPick("closers")))
}

// ## EOF
