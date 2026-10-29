// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_answer_chain.go
package wirepod_ttr

import (
	"strings"
	"sync"
	"time"

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
	// 2026-10-23: a point with notes - the point bubbles' lines (point_bubbles.prompt),
	// after one weather line when it's a weather question.
	if gopodChainHas(esn, "point_bubbles") {
		if pp := gopodPointBubblesPrompt(esn); pp != "" {
			if weather {
				lead := gopodBrobotsSettings.ChainPrompts["weather_lead"]
				if gopodIsFollowingUp(esn) {
					lead = gopodBrobotsSettings.ChainPrompts["weather_lead_next"]
				}
				return strings.TrimSpace(lead + " " + pp)
			}
			return pp
		}
	}
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

func gopodContains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// gopodGoldenAnswer (2026-10-23, operator: "What is AI wordplay? * the fun
// challenges humans input to the brobots * What is Explain the MAth? * the
// hilarity that ensues afterwards"): the exact line for a golden question
// (golden_answers), "" for none.
func gopodGoldenAnswer(esn, question string) string {
	// 2026-10-24: a "good <part of day>" opens with its greeting, checked
	// against the real time (gopod_greetings.go).
	if line := gopodGreetingLine(esn, question, time.Now()); line != "" {
		return line
	}
	lower := strings.ToLower(strings.Join(strings.Fields(question), " "))
	for _, g := range gopodBrobotsSettings.GoldenAnswers {
		for _, q := range g.Queries {
			if q = strings.ToLower(strings.TrimSpace(q)); q != "" && strings.Contains(lower, q) {
				return strings.TrimSpace(g.Line)
			}
		}
	}
	return ""
}

// gopodTopicGroupFor: the answer chain group a value point's topic asks for
// (topic_groups: a group's topics), "" for none.
func gopodTopicGroupFor(esn string) string {
	for _, g := range gopodGivenPoints(esn) {
		topic := strings.Fields(g)
		if len(topic) == 0 {
			continue
		}
		for group, tg := range gopodBrobotsSettings.TopicGroups {
			for _, t := range tg.Topics {
				if t == topic[0] {
					if _, ok := gopodBrobotsSettings.AnswerChains[group]; ok {
						return group
					}
				}
			}
		}
	}
	return ""
}

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
// duo A (duo.a.chain), the next duo B (duo.b.chain), a single answer brobots.
func gopodBrobotChainGroup(esn string) string {
	// 2026-10-22 (operator: "fractions is another example of catering the
	// answer-chain to a particular value points reveal target * Yes build it"):
	// an answer whose value point is a math topic's is brobots_math.
	if g := gopodTopicGroupFor(esn); g != "" {
		return g
	}
	// 2026-10-23 (operator: "a 'duo' answer-type templated mechanism, is a good
	// starter for a 'multi' answer-type * Brobot A 'duo-answer-type-A' > Brobot B
	// 'duo-answer-type-B'"): in a line to both, A answers first, B next (duo).
	if r, ok := gopodDuoRule(esn); ok && r.Chain != "" {
		if _, ok := gopodBrobotsSettings.AnswerChains[r.Chain]; ok {
			return r.Chain
		}
	}
	if len(gopodSummaryItemsFor(esn)) > 0 {
		if _, ok := gopodBrobotsSettings.AnswerChains["brobots_summary"]; ok {
			return "brobots_summary"
		}
	}
	// 2026-10-25: a single answer's chain from the response-chain template.
	if a, ok := gopodChainAnswerFor(esn); ok && a.Chain != "" {
		if _, ok := gopodBrobotsSettings.AnswerChains[a.Chain]; ok {
			return a.Chain
		}
	}
	return "brobots"
}

// gopodShapeBrobotAnswer runs a Brobot group's stitch parts around the model's
// answer (raw) - the parts listed before "llm" go before it, the rest after.
func gopodShapeBrobotAnswer(esn, question, raw string) string {
	// 2026-10-23: checked against its type's golden shape once made (gopod_answer_shapes.go).
	_, first := gopodAllFirstAnswer.Load(esn)
	shaped := gopodShapeBrobotAnswerParts(esn, question, raw)
	gopodAnswerShapeCheck(esn, question, shaped, first)
	return shaped
}

// gopodShapeBrobotAnswerParts: the group's parts around the model's answer.
func gopodShapeBrobotAnswerParts(esn, question, raw string) string {
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
	case "docbs":
		// 2026-10-29: Doc's brain wiring - the equation roast before his reply (gopod_docbs.go).
		// The docbs part is the answer's last stitch: its decision is cleared here, on or off.
		on := gopodDocBSOn(esn, question)
		gopodDocBSOnFor.Delete(esn + "\x00" + question)
		if on {
			if frame := gopodDocBSFrame(question); frame != "" {
				logger.Println("BROBOT_DOCBS frame " + gopodBrobotName(esn))
				return gopodStitchSnippet(frame, raw)
			}
		}
	case "docbs_close":
		if gopodDocBSOn(esn, question) {
			if line := gopodDocBSClose(); line != "" {
				return gopodStitchEndSnippet(line, raw)
			}
		}
	case "golden_line":
		// 2026-10-23: a golden question's exact answer, word for word, first.
		if line := gopodGoldenAnswer(esn, question); line != "" {
			logger.Println("BROBOT_GOLDEN_LINE " + gopodBrobotName(esn) + ": " + line)
			return gopodStitchSnippet(line, raw)
		}
	case "summary_bubbles":
		// 2026-10-23: a summary's bubbles, one per core topic (gopod_summary_answer.go).
		return gopodSummaryBubbles(esn, raw)
	case "point_bubbles":
		// 2026-10-23: a value point with notes becomes his thought bubbles
		// (gopod_point_bubbles.go) - after the model, before the other parts.
		return gopodPointBubbles(esn, question, raw)
	case "rule":
		// 2026-10-22: the group's rule line (topic_groups.<group>.rule_line, "Doc's
		// Rule: no credit for close."), its own thought - not when the point given
		// already is the rule.
		if g := gopodTopicGroupFor(esn); g != "" {
			ruleWords := strings.ToLower(strings.TrimRight(strings.TrimSpace(gopodBrobotsSettings.TopicGroups[g].RuleLine), ".!"))
			if line := strings.TrimSpace(gopodBrobotsSettings.TopicGroups[g].RuleLine); line != "" && !strings.Contains(strings.ToLower(raw), ruleWords) && !strings.Contains(strings.ToLower(gopodSayLine(esn)), ruleWords) {
				logger.Println("BROBOT_RULE " + gopodBrobotName(esn) + ": " + line)
				return strings.TrimSpace(raw) + gopodThoughtGap() + line
			}
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
		// Duo A hands over instead (duo.a.no_reminder), unless its type keeps one.
		if r, ok := gopodDuoRule(esn); ok && r.NoReminder && !gopodContains(r.KeepReminderTypes, gopodAnswerType(esn, question)) {
			return raw
		}
		if !gopodIsThanks(question) {
			if line := gopodReminderLine(question); line != "" && !gopodBrotherSaid(esn, line) {
				logger.Println("BROBOT_REMINDER " + gopodBrobotName(esn) + ": " + line)
				gopodReminderSaid.Store(esn, true)
				return gopodStitchEndSnippet(line, raw)
			}
		}
		// 2026-10-24 (link 5, live: "free: llm - DRIFT" after every point and egg
		// was found): nothing left to point at - no reminder is the right shape.
		gopodReminderNone.Store(esn, true)
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
		// 2026-10-23 (answer_shapes: a thank-you is his words only): no hand-over.
		if _, first := gopodAllFirstAnswer.LoadAndDelete(esn); first && !gopodIsThanks(question) {
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
	case "closing_remark":
		// 2026-10-25 (operator: "post-bingo, brobot 1, \"I Won!\", brobot 2, \"Oh!
		// Zero!\" * session chat > no bingo yet > scripted run > post experiences +
		// ending jokes as scripted closing remarks"; "No, before the 'Hey' reminder"):
		// "joke is closing remark of answer to how was bingo, not the response
		// answer-chain"): a Bingo question after a Bingo run - his answer's closing
		// sentence is his closing remark (bingo.closing), joined to it, right after
		// the model's words (before point bubbles and the reminder); his own copy of
		// it dropped first.
		if line := gopodExperienceClosing(esn, question); line != "" {
			logger.Println("BROBOT_CLOSING " + gopodBrobotName(esn) + ": " + line)
			return strings.TrimSpace(gopodDropSnippetCopy(line, raw)) + " " + line
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

// The exchange (2026-10-24, root-to-leaf link 5; operator: "the 2nd brobot
// following up not being accounted for in the run * so next question starts"):
// an answer with a follow-up coming (All's second Brobot, Explain the Math's
// follow-up) holds its line's turn until the follow-up is done - a new line's
// answer waits for it (addressing.all_next_wait_seconds at most). The Operator's
// own bubble still shows at once.
// 2026-10-24 (live: the hold never let go - a Brobot stays held after his answer,
// so his reply's code doesn't return until the next line): the line is let go the
// moment the follow-up Brobot has said his answer (gopodExchangeDone, kgsim.go).
var gopodExchange = struct {
	sync.Mutex
	text    string
	pending int
	waitFor map[string]int // esn -> answers still to come from him
}{waitFor: map[string]int{}}

// gopodExchangeOpen: the line's turn is held until follower has answered it.
func gopodExchangeOpen(text, follower string) {
	gopodExchange.Lock()
	if gopodExchange.pending == 0 {
		gopodExchange.text = strings.TrimSpace(text)
	}
	gopodExchange.pending++
	gopodExchange.waitFor[follower]++
	gopodExchange.Unlock()
}

// gopodExchangeDone: esn has said his answer - a held line waiting on him goes on.
func gopodExchangeDone(esn string) {
	gopodExchange.Lock()
	if gopodExchange.waitFor[esn] > 0 {
		gopodExchange.waitFor[esn]--
		gopodExchange.pending--
	}
	gopodExchange.Unlock()
}

// gopodExchangeWait: a new line waits for the line before's follow-up.
func gopodExchangeWait(esn, text string) {
	busy := func() (bool, string) {
		gopodExchange.Lock()
		defer gopodExchange.Unlock()
		return gopodExchange.pending > 0 && gopodExchange.text != strings.TrimSpace(text), gopodExchange.text
	}
	waiting, before := busy()
	if !waiting {
		return
	}
	logger.Println("GOPOD_EXCHANGE " + gopodBrobotName(esn) + " waits for the follow-up to \"" + gopodCut(before, 60) + "\"")
	limit := gopodSecondsToDuration(gopodBrobotsSettings.Addressing.AllNextWaitSeconds)
	started := time.Now()
	for waiting && time.Since(started) < limit {
		time.Sleep(200 * time.Millisecond)
		waiting, _ = busy()
	}
	if waiting {
		logger.Println("GOPOD_EXCHANGE gave up waiting after " + time.Since(started).Round(time.Second).String())
		gopodExchange.Lock()
		gopodExchange.pending = 0
		gopodExchange.waitFor = map[string]int{}
		gopodExchange.Unlock()
	}
}

// ## EOF
