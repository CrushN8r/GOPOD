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
}{strikes: map[string]int{}, note: map[string]string{}, welcome: map[string]bool{}}

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
	owner := gopodAnyOwner(text)
	if owner == "" || owner == host {
		return host
	}
	// A Brobot is named as himself, the GOPOD rep - not the persona his dial puts on his body.
	name := func(who string) string {
		if gopodIsBrobotBody(who) {
			return who
		}
		return gopodCallName(who)
	}
	fill := func(t string) string {
		return strings.TrimSpace(strings.NewReplacer("{owner}", name(owner), "{rep}", name(rep)).Replace(gopodPersonaFill(host, t)))
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

// gopodJobRouteReset: a new session - no strikes.
func gopodJobRouteReset() {
	gopodJobRouteState.Lock()
	gopodJobRouteState.strikes = map[string]int{}
	gopodJobRouteState.note = map[string]string{}
	gopodJobRouteState.welcome = map[string]bool{}
	gopodJobRouteState.Unlock()
}

// ## EOF
