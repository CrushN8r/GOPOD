// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_answer_chain.go
package wirepod_ttr

import (
	"strings"
	"sync"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// GOLDEN thought bubbles template - the answer chain (2026-10-21, operator: "build
// the thought bubbles boilerplate template * start with a base scaffolding for
// answer chain * copy scaffolding template & cater to different response groups *
// you can now apply to gomad responses, and some system responses"; "All this too
// should be templated").
//
// answer_chains in the settings: "base" is the scaffold (an ordered list of
// parts); every group ("brobots", "brobots_follow_up", "gomad_persona" ...) is a
// copy of it ({"from": "base"}) or its own list. The parts:
//   stitch parts - add to the answer: start_snippet (before the answer), weather_kit,
//                  end_snippet, redirect (after it), in the list's order
//   answer parts - ARE the answer and end the chain when they apply: who_line,
//                  manager_take (gomads), llm (the language model's answer)
// A part that doesn't apply this time is skipped. Each part keeps its own rule
// for when it applies; the chain only decides the order.

// gopodAnswerChain: a group's parts, following "from" copies (base by default).
func gopodAnswerChain(group string) []string {
	chains := gopodBrobotsSettings.AnswerChains
	for i := 0; i < 5; i++ {
		c, ok := chains[group]
		if !ok {
			group = "base"
			c, ok = chains["base"]
			if !ok {
				return nil
			}
		}
		if len(c.Parts) > 0 || c.From == "" {
			return c.Parts
		}
		group = c.From
	}
	return nil
}

// gopodChainPrompt (2026-10-22, the multiple responses template: "2) Weather
// comment 3) [a day of parties] #1"): what the model writes, in its parts - a
// weather question with a value point: one line on the weather (first) or
// following his brother's (next), then a new paragraph on the point
// (chain_prompts). "" when nothing applies.
func gopodChainPrompt(esn, question string) string {
	weather := gopodHasWholeWords(strings.ToLower(question), "weather") || gopodNativeTakeIs(gopodBrobotName(esn), "weather", question)
	if !weather || len(gopodGivenPoints(esn)) == 0 {
		return ""
	}
	key := "weather_point_first"
	if gopodIsFollowingUp(esn) {
		key = "weather_point_next"
	}
	return strings.TrimSpace(gopodBrobotsSettings.ChainPrompts[key])
}

// gopodReminderSaid: the reminder part ended this answer (esn -> true).
var gopodReminderSaid sync.Map

// gopodChainHas: this Brobot's group for this answer has the part.
func gopodChainHas(esn, part string) bool {
	for _, p := range gopodAnswerChain(gopodBrobotChainGroup(esn)) {
		if p == part {
			return true
		}
	}
	return false
}

// gopodBrobotChainGroup: the Brobots' group for this answer.
// 2026-10-22 (operator: "Multiple Responses Template * no reminders in 1st
// brobot response * cater answer-chains to handle known/expected/desired prompts
// to reveal value points"): the first answer in a line to both is
// brobots_all_first, the next brobots_follow_up, a single answer brobots.
func gopodBrobotChainGroup(esn string) string {
	if gopodIsFollowingUp(esn) {
		return "brobots_follow_up"
	}
	if _, first := gopodAllFirstAnswer.Load(esn); first {
		if _, ok := gopodBrobotsSettings.AnswerChains["brobots_all_first"]; ok {
			return "brobots_all_first"
		}
	}
	return "brobots"
}

// gopodShapeBrobotAnswer runs a Brobot group's stitch parts around the model's
// answer (raw) - the parts listed before "llm" go before it, the rest after.
func gopodShapeBrobotAnswer(esn, question, raw string) string {
	parts := gopodAnswerChain(gopodBrobotChainGroup(esn))
	if len(parts) == 0 {
		parts = []string{"start_snippet", "llm", "weather_kit", "end_snippet", "redirect"}
	}
	at := len(parts)
	for i, p := range parts {
		if p == "llm" {
			at = i
		}
	}
	// After the answer, in order.
	from := at + 1
	if from > len(parts) {
		from = len(parts)
	}
	for _, p := range parts[from:] {
		raw = gopodChainStitch(p, esn, question, raw, false)
	}
	// Before it: nearest first, so the list's order holds.
	for i := at - 1; i >= 0; i-- {
		raw = gopodChainStitch(parts[i], esn, question, raw, true)
	}
	return raw
}

// gopodChainStitch: one stitch part; unchanged when it doesn't apply.
func gopodChainStitch(part, esn, question, raw string, before bool) string {
	switch part {
	case "start_snippet":
		if snippet := gopodScriptedSnippet(esn, question, "start", ""); snippet != "" {
			logger.Println("BROBOT_SCRIPTED_SNIPPET=" + snippet)
			return gopodStitchSnippet(snippet, raw)
		}
	case "colour:value_point":
		// 2026-10-22 (COLOUR_SLOT_STRATEGY_002.md step 3; operator: "set cap from 5
		// sentences to 4 sentences"): the point's exact line (its say line), said
		// word for word first; the model's part after it is its colour, cut to
		// colour_slots.max_sentences (gopodTightenReply).
		if line := gopodSayLine(esn); line != "" {
			logger.Println("BROBOT_COLOUR_SLOT value_point: " + line)
			return gopodStitchSnippet(line, raw)
		}
	case "weather_kit":
		// live_weather.follow_up_kit_line: the follow-up's weather kit link, its own thought.
		if line := strings.TrimSpace(gopodBrobotsSettings.LiveWeather.FollowUpKitLine); line != "" && gopodIsFollowingUp(esn) &&
			gopodHasWholeWords(strings.ToLower(question), "weather") {
			return strings.TrimSpace(raw) + gopodThoughtGap() + line
		}
	case "reminder":
		// 2026-10-22 (COLOUR_SLOT_STRATEGY_002.md step 4; operator: "Every answer
		// steers back to the money"): a free answer (no value point given) ends
		// always; then no other end line (end_snippet). 2026-10-22 (the multiple
		// responses template): wherever a group lists it - the group decides.
		if !gopodIsThanks(question) {
			if line := gopodReminderLine(question); line != "" && !gopodBrotherSaid(esn, line) {
				logger.Println("BROBOT_REMINDER " + gopodBrobotName(esn) + ": " + line)
				gopodReminderSaid.Store(esn, true)
				return gopodStitchEndSnippet(line, raw)
			}
		}
	case "end_snippet":
		// One end line: none after the reminder part's.
		if _, said := gopodReminderSaid.LoadAndDelete(esn); said {
			return raw
		}
		// A follow-up doesn't end with his brother's end line.
		if snippet := gopodScriptedSnippet(esn, question, "end", raw); snippet != "" && !gopodBrotherSaid(esn, snippet) {
			logger.Println("BROBOT_SCRIPTED_SNIPPET_END=" + snippet)
			return gopodStitchEndSnippet(snippet, raw)
		}
	case "redirect":
		// A line to both: the first answer hands over (addressing.redirect_text).
		if _, first := gopodAllFirstAnswer.LoadAndDelete(esn); first {
			// 2026-10-21: his own hand-over line (host_personas.handover) first.
			brother := gopodBrobotName(gopodBrotherOf(esn))
			t := gopodHandoverLine(gopodBrobotName(esn), brother)
			if t == "" {
				t = strings.ReplaceAll(strings.TrimSpace(gopodBrobotsSettings.Addressing.RedirectText), "{brother}", brother)
			}
			if t != "" {
				line := t
				logger.Println("BROBOT_REDIRECT " + gopodBrobotName(esn) + ": " + line)
				return strings.TrimSpace(raw) + gopodThoughtGap() + line
			}
		}
	case "llm":
	default:
		logger.Println("GOPOD_CHECK answer_chains: no part named " + part)
	}
	return raw
}

// gopodGomadAnswer runs the gomad_persona chain for a line to a manager: its
// answer parts in order, the first that applies answers. True when one did.
func gopodGomadAnswer(host, text string) bool {
	parts := gopodAnswerChain("gomad_persona")
	if len(parts) == 0 {
		parts = []string{"who_line", "manager_take", "llm"}
	}
	for _, p := range parts {
		switch p {
		case "who_line":
			// The operator's own line, word for word (managers.who_words).
			if who := strings.TrimSpace(gopodBrobotsSettings.Managers.Personas[host].WhoLine); who != "" {
				if at, _ := gopodEggWord(strings.ToLower(text), gopodBrobotsSettings.Managers.WhoWords); at >= 0 {
					logger.Println("GOPOD_PERSONA " + host + " (who): " + who)
					gopodReporterSay(host, who, true, true)
					return true
				}
			}
		case "manager_take":
			// A take's words (the weather, the time and place) - its manager does the job.
			if take, _, ok := gopodManagerTakes(strings.ToLower(text)); ok {
				gopodManagerDo(take, text)
				return true
			}
		case "llm":
			reply, err := gopodPersonaAsk(host, text)
			if err != nil {
				logger.Println("GOPOD_PERSONA " + host + ": " + err.Error())
				return false
			}
			logger.Println("GOPOD_PERSONA " + host + ": " + reply)
			gopodReporterSay(host, reply, true, true)
			return true
		default:
			logger.Println("GOPOD_CHECK answer_chains: no gomad part named " + p)
		}
	}
	return false
}

// ## EOF
