// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_answer_chain.go
package wirepod_ttr

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

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

// gopodChainPrompt: what the model writes, in its parts - a weather question with a value point:
// one line on the weather (first) or following his brother's (next), then a new paragraph on the
// point (chain_prompts).
func gopodChainPrompt(esn, question string) string {
	if gopodIsIntro(question) {
		return ""
	}
	weather := gopodHasWholeWords(strings.ToLower(question), "weather") || gopodNativeTakeIs(gopodBrobotName(esn), "weather", question)
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
	if !weather {
		return ""
	}
	if len(gopodGivenPoints(esn)) == 0 {
		key := "weather_take"
		if gopodIsFollowingUp(esn) {
			key = "weather_take_next"
		}
		return strings.TrimSpace(gopodBrobotsSettings.ChainPrompts[key])
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

// gopodGoldenAnswer: the exact line for a golden question (golden_answers), "" for none.
func gopodGoldenAnswer(esn, question string) string {
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

// gopodTopicGroupFor: the answer chain group a value point's topic asks for (topic_groups: a
// group's topics), "" for none.
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
func gopodBrobotChainGroup(esn string) string {
	if v, ok := gopodIntroFor.Load(esn); ok && v.(bool) {
		if w, ok := gopodIntroWeatherFor.Load(esn); ok && w.(bool) {
			if _, ok := gopodBrobotsSettings.AnswerChains["brobots_intro_weather"]; ok {
				return "brobots_intro_weather"
			}
		}
		if _, ok := gopodBrobotsSettings.AnswerChains["brobots_intro"]; ok {
			return "brobots_intro"
		}
	}
	if g := gopodTopicGroupFor(esn); g != "" {
		return g
	}
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
	if a, ok := gopodChainAnswerFor(esn); ok && a.Chain != "" {
		if _, ok := gopodBrobotsSettings.AnswerChains[a.Chain]; ok {
			return a.Chain
		}
	}
	return "brobots"
}

// gopodShapeBrobotAnswer runs a Brobot group's stitch parts around the model's answer (raw) - the
// parts listed before "llm" go before it, the rest after.
func gopodShapeBrobotAnswer(esn, question, raw string) string {
	_, first := gopodAllFirstAnswer.Load(esn)
	shaped := gopodShapeBrobotAnswerParts(esn, question, raw)
	gopodAnswerShapeCheck(esn, question, shaped, first)
	kind := "answer"
	if v, ok := gopodIntroFor.Load(esn); ok && v.(bool) {
		kind = "intro"
	}
	gopodBrainAdd(gopodCallName(gopodBrobotName(esn)), question, kind, raw)
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
	from := at + 1
	if from > len(parts) {
		from = len(parts)
	}
	for _, p := range parts[from:] {
		raw = gopodChainStitch(p, esn, question, raw, false)
	}
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
		if line := gopodSayLine(esn); line != "" {
			logger.Println("BROBOT_COLOUR_SLOT value_point: " + line)
			return gopodStitchSnippet(line, raw)
		}
	case "docbs":
		on := gopodDocBSOn(esn, question)
		gopodDocBSOnFor.Delete(esn + "\x00" + question)
		if on {
			if frame := gopodDocBSFrame(question); frame != "" {
				logger.Println("BROBOT_DOCBS frame " + gopodBrobotName(esn))
				return gopodStitchSnippet(frame, raw)
			}
		}
	case "persona_chain":
		line := gopodPersonaChainFor2(esn, question)[0]
		gopodPersonaChainFor.Delete(esn + "\x00" + question)
		if line != "" {
			return gopodStitchSnippet(line, raw)
		}
	case "persona_chain_close":
		if line := gopodPersonaChainFor2(esn, question)[1]; line != "" {
			return gopodStitchEndSnippet(line, raw)
		}
	case "docbs_close":
		if gopodDocBSOn(esn, question) {
			if line := gopodDocBSClose(); line != "" {
				return gopodStitchEndSnippet(line, raw)
			}
		}
	case "greeting":
		return gopodGreetingStitch(esn, question, raw)
	case "golden_line":
		if line := gopodJobRouteWelcomeLine(gopodBrobotName(esn)); line != "" {
			logger.Println("BROBOT_GOLDEN_LINE " + gopodBrobotName(esn) + " (welcome): " + line)
			return gopodStitchSnippet(line, raw)
		}
		if line := gopodGoldenAnswer(esn, question); line != "" {
			logger.Println("BROBOT_GOLDEN_LINE " + gopodBrobotName(esn) + ": " + line)
			return gopodStitchSnippet(line, raw)
		}
	case "summary_bubbles":
		return gopodSummaryBubbles(esn, raw)
	case "point_bubbles":
		return gopodPointBubbles(esn, question, raw)
	case "rule":
		if g := gopodTopicGroupFor(esn); g != "" {
			ruleWords := strings.ToLower(strings.TrimRight(strings.TrimSpace(gopodBrobotsSettings.TopicGroups[g].RuleLine), ".!"))
			if line := strings.TrimSpace(gopodBrobotsSettings.TopicGroups[g].RuleLine); line != "" && !strings.Contains(strings.ToLower(raw), ruleWords) && !strings.Contains(strings.ToLower(gopodSayLine(esn)), ruleWords) {
				logger.Println("BROBOT_RULE " + gopodBrobotName(esn) + ": " + line)
				return strings.TrimSpace(raw) + gopodThoughtGap() + line
			}
		}
	case "contact":
		return gopodYourselfStitch(question, raw)
	case "weather_kit":
		if line := strings.TrimSpace(gopodBrobotsSettings.LiveWeather.FollowUpKitLine); line != "" && gopodIsFollowingUp(esn) &&
			gopodHasWholeWords(strings.ToLower(question), "weather") {
			return strings.TrimSpace(raw) + gopodThoughtGap() + line
		}
	case "reminder":
		if r, ok := gopodDuoRule(esn); ok && r.NoReminder && !gopodContains(r.KeepReminderTypes, gopodAnswerType(esn, question)) {
			return raw
		}
		if line := gopodGameWelcomeLine(esn); line != "" {
			logger.Println("BROBOT_REMINDER " + gopodBrobotName(esn) + " (game welcome): " + line)
			gopodReminderSaid.Store(esn, true)
			return gopodStitchEndSnippet(line, raw)
		}
		if !gopodIsThanks(question) {
			if line := gopodReminderLine(question); line != "" && !gopodBrotherSaid(esn, line) {
				logger.Println("BROBOT_REMINDER " + gopodBrobotName(esn) + ": " + line)
				gopodReminderSaid.Store(esn, true)
				return gopodStitchEndSnippet(line, raw)
			}
		}
		gopodReminderNone.Store(esn, true)
	case "end_snippet":
		if _, said := gopodReminderSaid.LoadAndDelete(esn); said {
			return raw
		}
		if snippet := gopodScriptedSnippet(esn, question, "end", raw); snippet != "" && !gopodBrotherSaid(esn, snippet) {
			logger.Println("BROBOT_SCRIPTED_SNIPPET_END=" + snippet)
			return gopodStitchEndSnippet(snippet, raw)
		}
	case "redirect":
		if _, first := gopodAllFirstAnswer.LoadAndDelete(esn); first && !gopodIsThanks(question) {
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
		if line := gopodExperienceClosing(esn, question); line != "" {
			logger.Println("BROBOT_CLOSING " + gopodBrobotName(esn) + ": " + line)
			return strings.TrimSpace(gopodDropSnippetCopy(line, raw)) + " " + line
		}
	case "intro_wrap":
		return gopodIntroWrap(gopodBrobotName(esn), raw, true)
	case "intro_weather":
		return gopodIntroThenWeather(esn, question, raw)
	case "llm":
	default:
		logger.Println("GOPOD_CHECK answer_chains: no part named " + part)
	}
	return raw
}

// gopodGomadAnswer runs the gomad_persona chain for a line to a manager: its answer parts in order,
// the first that applies answers.
func gopodGomadAnswer(host, text string) bool {
	return gopodGomadAnswerAs(host, host, text)
}

// gopodGomadAnswers: the Gomads whose answer is being made now (host -> count).
var gopodGomadAnswers sync.Map

// gopodGomadAnswering: host's answer is being made (the line queue's idle wait).
func gopodGomadAnswering(host string) bool {
	n, ok := gopodGomadAnswers.Load(strings.ToUpper(strings.TrimSpace(host)))
	return ok && n.(*atomic.Int32).Load() > 0
}

// gopodGomadAnswerAs: the gomad_persona chain for host, its lines said as sayAs.
func gopodGomadAnswerAs(host, sayAs, text string) bool {
	return gopodGomadAnswerEarly(host, sayAs, text, nil)
}

// The early answer: a Gomad's answer is asked for, wrapped, and its first spoken chunk's voice made
// while the Operator's line is read out; once it's heard, the answer shows and plays as before (its
// brain first).
type gopodGomadEarlyAnswer struct {
	done  chan struct{}
	reply string
	err   error
	seq   int64
}

// gopodGomadPartFor: the gomad_persona chain's part that answers text ("" none).
func gopodGomadPartFor(host, text string) string {
	parts := gopodAnswerChain("gomad_persona")
	if len(parts) == 0 {
		parts = []string{"who_line", "manager_take", "llm"}
	}
	for _, p := range parts {
		switch p {
		case "who_line":
			if who := strings.TrimSpace(gopodPersonaFill(host, gopodBrobotsSettings.Managers.Personas[host].WhoLine)); who != "" {
				if at, _ := gopodEggWord(strings.ToLower(text), gopodBrobotsSettings.Managers.WhoWords); at >= 0 {
					return p
				}
			}
		case "manager_take":
			if _, _, ok := gopodManagerTakes(strings.ToLower(text)); ok {
				return p
			}
		case "llm":
			return p
		case "greeting":
		default:
			logger.Println("GOPOD_CHECK answer_chains: no gomad part named " + p)
		}
	}
	return ""
}

// gopodGomadAnswerStart starts host's answer to text now, when the model makes it (nil otherwise):
// asked, wrapped (intro, chain), its first chunk voiced.
func gopodGomadAnswerStart(host, sayAs, text string) *gopodGomadEarlyAnswer {
	if gopodGomadPartFor(host, text) != "llm" {
		return nil
	}
	e := &gopodGomadEarlyAnswer{done: make(chan struct{}), seq: gopodStopSeq.Load()}
	go func() {
		defer close(e.done)
		e.reply, e.err = gopodGomadReplyMade(host, text)
		if e.err != nil || gopodStopSeq.Load() != e.seq {
			return
		}
		first := e.reply
		if chunks := gopodSpokenChunks(e.reply); len(chunks) > 0 {
			first = chunks[0]
		}
		if first = strings.TrimSpace(first); first != "" {
			gopodReporterFile(sayAs, gopodSaysFor(sayAs, first))
		}
	}()
	return e
}

// gopodGomadReplyMade: host's answer to text from the model, wrapped - the intro template, then its
// chain.
func gopodGomadReplyMade(host, text string) (string, error) {
	if gopodGemmaEyesAsked(host, text) {
		return gopodGemmaEyesLook(), nil
	}
	if gopodCommCheckHandover(host, text) {
		return "", nil
	}
	if line := gopodNonsenseLine(text); line != "" {
		return line, nil
	}
	if gopodNeverIn(text, host+"'s line") {
		return gopodBrobotsSettings.CleanPolitics.NeverInLine, nil
	}
	reply, err := gopodPersonaAsk(host, text)
	if err != nil {
		return "", err
	}
	if gopodIsIntro(text) {
		reply = gopodIntroWrap(host, reply)
	}
	logger.Println("GOPOD_PERSONA " + host + ": " + reply)
	reply = gopodPersonaChainWrap(host, reply)
	for _, p := range gopodAnswerChain("gomad_persona") {
		switch p {
		case "greeting":
			reply = gopodGreetingStitch("", text, reply)
		case "contact":
			reply = gopodYourselfStitch(text, reply)
		}
	}
	return reply, nil
}

// gopodGomadAnswerEarly: host's answer, said and shown as sayAs; early, when given, is its answer
// already started (gopodGomadAnswerStart).
func gopodGomadAnswerEarly(host, sayAs, text string, early *gopodGomadEarlyAnswer) bool {
	n, _ := gopodGomadAnswers.LoadOrStore(strings.ToUpper(strings.TrimSpace(sayAs)), &atomic.Int32{})
	n.(*atomic.Int32).Add(1)
	defer n.(*atomic.Int32).Add(-1)
	switch gopodGomadPartFor(host, text) {
	case "who_line":
		who := strings.TrimSpace(gopodPersonaFill(host, gopodBrobotsSettings.Managers.Personas[host].WhoLine))
		logger.Println("GOPOD_PERSONA " + host + " (who): " + who)
		gopodReporterSay(sayAs, who, true, true)
		return true
	case "manager_take":
		take, _, _ := gopodManagerTakes(strings.ToLower(text))
		gopodManagerDo(take, text)
		return true
	case "llm":
		if early != nil && gopodStopSeq.Load() != early.seq {
			logger.Println("GOPOD_PERSONA " + host + ": answer dropped - stop run before it was said")
			return true
		}
		if got := strings.TrimSpace(gopodGotItNamed(host).Text); got != "" && !gopodCommCheckSelf.Load() {
			gopodReporterSay(sayAs, got, true, true)
		}
		gopodGomadBrain(host, sayAs, text)
		if gopodCommCheckHandover(host, text) {
			gopodCommCheckCrew(host, sayAs)
			gopodBrainAdd(gopodCallName(host), text, "answer", "")
			return true
		}
		var reply string
		var err error
		seq := gopodStopSeq.Load()
		if early != nil {
			<-early.done
			reply, err, seq = early.reply, early.err, early.seq
		} else {
			reply, err = gopodGomadReplyMade(host, text)
		}
		if gopodStopSeq.Load() != seq {
			logger.Println("GOPOD_PERSONA " + host + ": answer dropped - stop run while it was made")
			return true
		}
		if err != nil {
			logger.Println("GOPOD_PERSONA " + host + ": " + err.Error())
			return false
		}
		gopodReporterSay(sayAs, reply, true, true)
		kind := "answer"
		if gopodIsIntro(text) {
			kind = "intro"
		}
		gopodBrainAdd(gopodCallName(host), text, kind, reply)
		return true
	}
	return false
}

// The exchange: an answer with a follow-up coming (All's second Brobot, Explain the Math's
// follow-up) holds its line's turn until the follow-up is done - a new line's answer waits for it
// (addressing.all_next_wait_seconds at most).
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
