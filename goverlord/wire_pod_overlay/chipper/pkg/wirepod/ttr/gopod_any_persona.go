// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_any_persona.go
package wirepod_ttr

// The "." key - any persona (2026-11-17, operator: "can I numpad-map the '.' to
// 'any-persona', where I call the persona by name, and the group knows the
// answer, but the group also know who should answer * the known persona responds
// adding to the group chat"; "ensure '.' from both PC keyboard & USB numpad, are
// mapped correctly"): a line on "." (held, or the keyboard's "." in the picker)
// goes to the persona it names - by name, nickname or call name. With no one
// named, it goes to whoever owns the topic the look-up finds (its card's owner:
// a card's "- **owner:**", else a persona's own card, else the system's for
// the scrolls and system cards, else numpad_ptt.any.default_owner - the GOPOD
// reps). Named but not the owner: the named one says whose job it is
// (numpad_ptt.any.redirect_text), and the owner answers. Every answer is the
// persona's usual one, in the group chat.

import (
	"path/filepath"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"sort"
	"strings"
)

// gopodNumpadAnyKey: the "." key (both layouts' NumpadDecimal).
const gopodNumpadAnyKey = "KPdot"

// gopodTypedAny: the "." key's target - any persona.
const gopodTypedAny = "any"

// gopodAnyNamed: the persona a line names (its key in host_personas), "" for none -
// the longest name said first, whole words.
func gopodAnyNamed(text string) string {
	q := " " + strings.Join(gopodLookupTokens(text), " ") + " "
	type name struct{ who, words string }
	var names []name
	for who, hp := range gopodBrobotsSettings.HostPersonas {
		for _, n := range []string{who, hp.Name, hp.Nickname, hp.CallName} {
			if n = strings.TrimSpace(n); n != "" {
				names = append(names, name{who, strings.Join(gopodLookupTokens(n), " ")})
			}
		}
	}
	sort.SliceStable(names, func(i, j int) bool { return len(names[i].words) > len(names[j].words) })
	for _, n := range names {
		if n.words != "" && strings.Contains(q, " "+n.words+" ") {
			return n.who
		}
	}
	return ""
}

// gopodAnyOwner: who owns the topic a line asks about - the first of the look-up's
// cards that has an owner (2026-11-23, live: "booking" found the repo's BOOKING.md
// first, ownerless, then the booking card that names the GOPOD rep).
func gopodAnyOwner(text string) string {
	for _, c := range gopodLookup(text) {
		if o := gopodCardOwner(c); o != "" {
			return o
		}
	}
	return ""
}

// gopodCardOwner: a card's owner - its "owner" head line, else the persona whose
// card it is, else the system's (the scrolls, the system cards), else "".
func gopodCardOwner(c gopodLookupCard) string {
	if c.Owner != "" {
		if o := gopodAnyNamed(c.Owner); o != "" {
			return o
		}
	}
	folder := gopodBrobotsSettings.PersonaProfiles.Folder
	for who := range gopodBrobotsSettings.HostPersonas {
		if f := gopodPersonaProfileFileIn(folder, who); f != "" && filepath.Clean(f) == filepath.Clean(c.Path) {
			return who
		}
	}
	if strings.Contains(c.Path, "/"+gopodTorahDir+"/") || strings.Contains(c.Path, "/99_system/") {
		return gopodBrobotsSettings.SystemBubble.VoicePersona
	}
	return ""
}

// gopodAnyTarget: a persona's line target - a Brobot himself (his body, the
// dial's 0%), DOC or PIP (their body), any other persona its own voice.
func gopodAnyTarget(who string) string {
	if gopodIsBrobotBody(who) {
		return gopodDialZeroMark + gopodBrobotESN(who)
	}
	if hp := gopodBrobotsSettings.HostPersonas[who]; hp.Body != "" {
		return gopodBrobotESN(hp.Body)
	}
	return gopodPersonaMark + who
}

// gopodAnyPersonaLine: a line on "." - to whoever it names, or to the topic's owner.
func gopodAnyPersonaLine(text, how string) error {
	return gopodAnyPersonaLineNamed(gopodAnyNamed(text), text, how)
}

// gopodAnyPersonaLineNamed: a line to named (the one it names, or KEYS when
// Goverlord hands it down) - he says whose job it is, and the owner answers.
func gopodAnyPersonaLineNamed(named, text, how string) error {
	ap := gopodBrobotsSettings.NumpadPTT.Any
	// 2026-11-17 (live: "keys where do i start" - KEYS handed it to the default
	// owner): only a topic the look-up finds has an owner to hand over to; the
	// default owner is for a line naming no one.
	owner := gopodAnyOwner(text)
	logger.Println("GOPOD_PICK any persona: named=" + named + " owner=" + owner + " for: " + text)
	who := named
	if who == "" {
		who = owner
		if who == "" {
			who = strings.TrimSpace(ap.DefaultOwner)
		}
	} else if owner != "" && owner != named && strings.TrimSpace(ap.RedirectText) != "" {
		// Not the named one's job: he says whose it is, then the owner answers.
		line := strings.ReplaceAll(ap.RedirectText, "{owner}", gopodRoleName(owner))
		gopodAnySay(named, line)
		who = owner
	}
	if who == "" {
		who = gopodBrobotsSettings.SystemBubble.VoicePersona
	}
	return gopodKeyLine(gopodAnyTarget(who), text, how)
}

// gopodAnySay: who says one line - a Brobot's or DOC's/PIP's on his body, a voice persona's in its voice.
func gopodAnySay(who, line string) {
	if gopodIsBrobotBody(who) {
		gopodBodySay(who, line)
		return
	}
	if hp := gopodBrobotsSettings.HostPersonas[who]; hp.Body != "" {
		gopodRecordSay(hp.Body, line)
		return
	}
	gopodReporterSay(who, line, true, true)
}

// ## EOF
