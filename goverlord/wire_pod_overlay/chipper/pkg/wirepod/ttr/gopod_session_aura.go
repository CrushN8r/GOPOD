// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_session_aura.go
package wirepod_ttr

// The SESSION aura (2026-11-24, operator: "Make Goverlord a session aura brain
// bubble that each persona can then think with"; the scratch pad:
// gopod_notes/SESSION_AURA_SCRATCH_PAD_001.md - "SESSION Aura is bigger than
// Persona Aura ... the contextual identity of the entire experience"). Goverlord
// is the group-think persona, the silent background aura: his brain bubble's
// AURA line is the session's, the same for every persona that thinks this turn.
//
// Code-made from what the session already knows, never invented - its dimensions
// as the scratch pad names them: WHEN and WHERE (environment), WHO is here
// (human), what's happening NOW (activity), the TOPICS in play (object/interest),
// the VIBE (the activity's, then the pace of the chat) and, with both on stage,
// Doc and Pip's tension. Each persona keeps its own identity; the aura is the
// variable context ("Identity Kernel + ... + Variable Context = Output Behaviour").
// Its words are Goverlord's card's (brain_chain.aura).

import (
	"strconv"
	"strings"
	"time"
)

// gopodSessionAura: the session's AURA line, "" with no aura line on the card.
func gopodSessionAura() string {
	a := gopodBrobotsSettings.BrainChain.Aura
	if strings.TrimSpace(a["line"]) == "" {
		return ""
	}
	recs := gopodBrainSnapshot()
	now := time.Now()
	// WHO: the people (the operator, the talking visitors), then the voices heard.
	// 2026-11-24 (operator: "People here: Master CrushN8r * redundant ... and
	// confusing"): the operator is always here - only visitors are news.
	people := gopodVisitorsTalking()
	peoplePart := ""
	if len(people) > 0 {
		peoplePart = strings.ReplaceAll(a["visitors"], "{people}", gopodBrainList(people))
	}
	var voices []string
	heard := map[string]bool{}
	var questions []string
	justPlayed := ""
	for i := len(recs) - 1; i >= 0; i-- {
		r := recs[i]
		if r.Kind == gopodBrainPerformedKind {
			// 2026-11-24 (operator: "is there an aura ... keeping tabs on the chat to
			// help the personas?"): the latest run the session played.
			if justPlayed == "" {
				justPlayed = r.Question
			}
			continue
		}
		if !heard[r.To] {
			heard[r.To] = true
			voices = append(voices, r.To)
		}
		if r.Kind != "intro" && len(questions) < 6 {
			questions = append(questions, r.Question)
		}
	}
	// NOW: the activity - the comm check, a run, the session's first line, or chat.
	doing := a["chat"]
	switch run, running := gopodRunNow(); {
	case gopodCommCheckRunning.Load():
		doing = a["comm_check"]
	case running:
		doing = strings.ReplaceAll(a["run"], "{run}", run.Name)
	case len(recs) == 0:
		doing = a["first"]
	}
	// VIBE: the activity's own, then the pace - lines in the last few minutes.
	vibe := a["vibe_chat"]
	switch {
	case gopodCommCheckRunning.Load():
		vibe = a["vibe_comm_check"]
	case gopodRunActive():
		vibe = a["vibe_run"]
	}
	paceMin, _ := strconv.ParseFloat(a["pace_minutes"], 64)
	lively, _ := strconv.Atoi(a["lively_lines"])
	if paceMin > 0 && lively > 0 {
		if n := len(gopodBrainWithin(recs, paceMin)); n >= lively {
			vibe += a["pace_lively"]
		} else if n <= 1 && len(recs) > 0 {
			vibe += a["pace_quiet"]
		}
	}
	// TOPICS: the cards the recent questions asked about, newest first, each once.
	var topics []string
	seen := map[string]bool{}
	for _, q := range questions {
		for _, c := range gopodLookup(q) {
			if len(c.Words) == 0 {
				continue
			}
			if t := strings.TrimSpace(c.Words[0]); t != "" && !seen[strings.ToLower(t)] {
				seen[strings.ToLower(t)] = true
				topics = append(topics, t)
			}
			break
		}
		if len(topics) >= 3 {
			break
		}
	}
	// Doc and Pip both on stage: their tension is the session's too.
	tension := ""
	if heard[gopodCallName("DOC")] && heard[gopodCallName("PIP")] {
		tension = a["tension_doc_pip"]
	}
	none := gopodBrobotsSettings.BrainChain.None
	list := func(l []string) string {
		if len(l) == 0 {
			return none
		}
		return gopodBrainList(l)
	}
	when := now.Weekday().String() + " " + gopodDayPart(now.Hour())
	return strings.TrimSpace(strings.NewReplacer(
		"{when}", when, "{where}", a["where"], "{people_part}", peoplePart, "{voices}", list(voices),
		"{played_part}", gopodAuraPart(a["played"], "{run}", justPlayed),
		"{doing}", doing, "{topics}", strings.TrimSpace(gopodBrainListOr(topics, a["no_topics"], none)), "{vibe}", vibe, "{tension}", tension,
	).Replace(a["line"]))
}

// gopodAuraVoices: the voices heard this session, latest first - background
// thinking (the brain's "In the chat:" line, hidden from the bubble).
func gopodAuraVoices() string {
	recs := gopodBrainSnapshot()
	var voices []string
	heard := map[string]bool{}
	for i := len(recs) - 1; i >= 0; i-- {
		if r := recs[i]; r.Kind != gopodBrainPerformedKind && !heard[r.To] {
			heard[r.To] = true
			voices = append(voices, r.To)
		}
	}
	if len(voices) == 0 {
		return ""
	}
	return gopodBrainList(voices)
}

// gopodAuraPart: an aura piece with its value filled, "" when there's no value.
func gopodAuraPart(tmpl, slot, value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return strings.ReplaceAll(tmpl, slot, value)
}

// gopodBrainListOr: the list, or its own "none" word, else the brain's.
func gopodBrainListOr(l []string, empty, none string) string {
	if len(l) > 0 {
		return gopodBrainList(l)
	}
	if strings.TrimSpace(empty) != "" {
		return empty
	}
	return none
}

// ## EOF
