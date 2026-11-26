// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_job_route.go
package wirepod_ttr

// Is it my job? (2026-11-23, operator: "a lost visitor in a company ... whoever
// the lost visitor approaches for information, that GOPOD employee persona must
// know how to respond * that persona knows & respects its employee limits * that
// persona knows where to direct next"; the loop: "if I ask an employee a
// question outside that employee's company boundaries once, suggest/redirect
// accordingly * ... twice, suggest/redirect to brobot 1 & README.md * ... 3
// times, auto-fallback to brobot 1 & README.md lookup for brief GOPOD
// welcome/intro").
//
// Before a persona answers a line, the code checks whose job its topic is - the
// look-up's owner, the same one the "." key hands a line to (gopodAnyOwner). Not
// his: a strike, counted per persona and topic owner until thanks resets the
// session. Strike 1: his ROUTE line names the owner (brain_chain.route_other).
// Strike 2: it names the GOPOD rep and README.md (route_again). Strike 3: he says
// route_handoff, and the rep answers instead, with README.md looked up for a
// short GOPOD welcome (route_welcome). The reps themselves (the Brobots) answer
// everything - they're the fallback. One gemma call per answer, as before.

import (
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

var gopodJobRouteState = struct {
	sync.Mutex
	strikes map[string]int    // persona|owner -> lines asked outside his job
	note    map[string]string // persona -> his ROUTE line for the answer he's making
	welcome map[string]bool   // the rep -> his next look-up is README.md's welcome
	welcomeLine map[string]bool // the rep -> his next answer opens with route_welcome_line
}{strikes: map[string]int{}, note: map[string]string{}, welcome: map[string]bool{}, welcomeLine: map[string]bool{}}

// gopodJobRoute: who answers a line sent to host - host himself (his ROUTE line
// set for it), or the GOPOD rep on the third strike.
func gopodJobRoute(host, text string) string {
	bc := gopodBrobotsSettings.BrainChain
	gopodJobRouteState.Lock()
	delete(gopodJobRouteState.note, gopodCallName(host))
	gopodJobRouteState.Unlock()
	rep := gopodAnyNamed(bc.RouteRep)
	if rep == "" || host == rep || gopodIsBrobotBody(host) || host == gopodGoverlordHost ||
		gopodCommCheckRunning.Load() || gopodIsIntro(text) || gopodIsThanks(text) {
		return host
	}
	// 2026-11-24 (operator: "how was your [bingo] experience ... applies to all
	// runs"): a run's experience is the Brobots' - every run's handler is theirs.
	if gopodExperienceAsked(text) != nil {
		logger.Println("GOPOD_JOB_ROUTE " + host + ": a run's experience - over to " + rep + " for: " + text)
		if line := strings.TrimSpace(strings.ReplaceAll(gopodPersonaFill(host, bc.RouteHandoff), "{rep}", gopodRoleName(rep))); line != "" {
			gopodAnySay(host, line)
		}
		return rep
	}
	// Contact info - booking too, which is "email us" - is every staffer's to give
	// (2026-11-24); the contact line goes with the answer, and its QR goes up.
	if gopodAsksContact(text) {
		return host
	}
	// 2026-11-24 (operator: "Operator > Goverlord > Keys > Gomads/Interns >
	// Gominions ... Gomads/Interns/Reporters/Hosts DOC & PIP, are company staff,
	// not really info providers * Gominions ... are the info providers ... via
	// game lookups"): a topic the look-up finds but no card owns is the GOPOD
	// rep's - staff send info asks to the Gominions.
	// The company itself (its people, groups, chain) is every staffer's to talk about.
	owner := gopodAnyOwner(text)
	if cards := gopodLookup(text); owner == "" && len(cards) > 0 {
		company := false
		for _, c := range cards {
			if strings.Contains(filepath.ToSlash(c.Path), "/personas/") {
				company = true
				break
			}
		}
		if !company {
			owner = rep
		}
	}
	if owner == "" || owner == host {
		return host
	}

	fill := func(t string) string {
		return strings.TrimSpace(strings.NewReplacer("{owner}", gopodRoleName(owner), "{rep}", gopodRoleName(rep)).Replace(gopodPersonaFill(host, t)))
	}
	key := host + "|" + owner
	gopodJobRouteState.Lock()
	gopodJobRouteState.strikes[key]++
	n := gopodJobRouteState.strikes[key]
	switch n {
	case 1:
		gopodJobRouteState.note[gopodCallName(host)] = fill(bc.RouteOther)
	case 2:
		gopodJobRouteState.note[gopodCallName(host)] = fill(bc.RouteAgain)
	default:
		delete(gopodJobRouteState.strikes, key)
		gopodJobRouteState.welcome[gopodCallName(rep)] = true
		gopodJobRouteState.welcomeLine[gopodCallName(rep)] = true
	}
	gopodJobRouteState.Unlock()
	logger.Println("GOPOD_JOB_ROUTE " + host + ": owner=" + owner + " strike=" + strconv.Itoa(n) + " for: " + text)
	if n >= 3 {
		if line := fill(bc.RouteHandoff); line != "" {
			gopodAnySay(host, line)
		}
		return rep
	}
	return host
}

// gopodGoverlordHandsDown (2026-11-24, operator: "Goverlord redirects most
// things to Keys, as Goverlord is busy orchestrating * Keys is the key master";
// live: "what's the weather?" to Goverlord set off the weather hand-off with no
// Brobot to finish it): a line to Goverlord that isn't about him - he says his
// one cold line (brain_chain.route_goverlord) and KEYS takes it as the "." key
// does: he says whose job it is, and its owner answers. Goverlord takes no strikes.
func gopodGoverlordHandsDown(text, how string) bool {
	keys := gopodBrobotsSettings.SystemBubble.VoicePersona
	line := strings.TrimSpace(gopodBrobotsSettings.BrainChain.RouteGoverlord)
	if keys == "" || line == "" || gopodCommCheckRunning.Load() || gopodIsIntro(text) || gopodIsThanks(text) {
		return false
	}
	if gopodAnyOwner(text) == gopodGoverlordHost {
		return false
	}
	logger.Println("GOPOD_JOB_ROUTE GOVERLORD hands down to " + keys + " for: " + text)
	gopodAnySay(gopodGoverlordHost, strings.ReplaceAll(line, "{keys}", gopodCallName(keys)))
	if err := gopodAnyPersonaLineNamed(keys, text, how); err != nil {
		logger.Println("GOPOD_JOB_ROUTE " + keys + ": " + err.Error())
	}
	return true
}

// gopodRoleName (2026-11-24, operator: "why is there still overlap between
// brobots & DOC & PIP"): who to name when sending someone on - a Brobot as
// himself (the GOPOD rep), never the persona his dial puts on his body.
func gopodRoleName(who string) string {
	if gopodIsBrobotBody(who) {
		return who
	}
	return gopodCallName(who)
}

// gopodJobRouteNote: self's ROUTE line for the answer he's making ("" when it's his job).
func gopodJobRouteNote(self string) string {
	gopodJobRouteState.Lock()
	defer gopodJobRouteState.Unlock()
	return gopodJobRouteState.note[gopodCallName(self)]
}

// gopodJobRouteWelcome: self's look-up this time is README.md's welcome (taken once).
func gopodJobRouteWelcome(self string) bool {
	gopodJobRouteState.Lock()
	defer gopodJobRouteState.Unlock()
	w := gopodJobRouteState.welcome[gopodCallName(self)]
	delete(gopodJobRouteState.welcome, gopodCallName(self))
	return w
}

// gopodJobRouteWelcomeDue: self's look-up this time will be README.md's welcome (not taken).
func gopodJobRouteWelcomeDue(self string) bool {
	gopodJobRouteState.Lock()
	defer gopodJobRouteState.Unlock()
	return gopodJobRouteState.welcome[gopodCallName(self)]
}

// gopodJobRouteWelcomeLine (2026-11-24, live: "Brobot 0 here" - README's comeback
// story taken for himself): the third strike's welcome, word for word, once.
func gopodJobRouteWelcomeLine(self string) string {
	gopodJobRouteState.Lock()
	due := gopodJobRouteState.welcomeLine[gopodCallName(self)]
	delete(gopodJobRouteState.welcomeLine, gopodCallName(self))
	gopodJobRouteState.Unlock()
	if !due {
		return ""
	}
	return strings.TrimSpace(gopodBrobotsSettings.BrainChain.RouteWelcomeLine)
}

// gopodAsksContact (2026-11-24, operator: "contact info * make this its own
// golden process to be called by any persona"): a line asking for contact info,
// or about booking (booking is "email us") - route_staff_words.
func gopodAsksContact(text string) bool {
	return gopodAsksAny(text, gopodBrobotsSettings.BrainChain.RouteStaffWords)
}

// gopodAsksAny: text says one of words (a comma list), whole words.
func gopodAsksAny(text, words string) bool {
	q := " " + strings.Join(gopodLookupTokens(text), " ") + " "
	for _, w := range strings.Split(words, ",") {
		if w = strings.Join(gopodLookupTokens(w), " "); w != "" && strings.Contains(q, " "+w+" ") {
			return true
		}
	}
	return false
}

// gopodYourselfStitch (2026-11-25, the GOPOD Yourself QR card, gopod_yourself_qr.md):
// his own line, then the card's golden lines word for word - the keep-up line for a
// keeping-up or subscribing ask, the contact line for a contact or booking ask.
// Their words put the QRs on the board. The contact line not when his answer
// already gives the address.
func gopodYourselfStitch(question, raw string) string {
	if gopodIsThanks(question) || gopodIsIntro(question) {
		return raw
	}
	gy := gopodBrobotsSettings.GopodYourselfQr
	out := strings.TrimSpace(raw)
	if line := strings.TrimSpace(gy.KeepUpLine); line != "" && gopodAsksAny(question, gy.KeepUpWords) && !strings.Contains(out, line) {
		logger.Println("GOPOD_YOURSELF keep-up line given for: " + question)
		out += gopodThoughtGap() + line
	}
	if line := strings.TrimSpace(gy.ContactLine); line != "" && gopodAsksContact(question) {
		if addr := gopodContactAddress(line); addr == "" || !strings.Contains(strings.ToLower(out), strings.ToLower(addr)) {
			logger.Println("GOPOD_YOURSELF contact line given for: " + question)
			out += gopodThoughtGap() + line
		}
	}
	if out == strings.TrimSpace(raw) {
		return raw
	}
	return out
}

// gopodContactAddress: the email address in the contact line.
func gopodContactAddress(line string) string {
	for _, f := range strings.Fields(line) {
		if strings.Contains(f, "@") {
			return strings.Trim(f, ".,;:!?()")
		}
	}
	return ""
}

// gopodJobRouteReset: a new session - no strikes.
func gopodJobRouteReset() {
	gopodJobRouteState.Lock()
	gopodJobRouteState.strikes = map[string]int{}
	gopodJobRouteState.note = map[string]string{}
	gopodJobRouteState.welcome = map[string]bool{}
	gopodJobRouteState.welcomeLine = map[string]bool{}
	gopodJobRouteState.Unlock()
}

// ## EOF
