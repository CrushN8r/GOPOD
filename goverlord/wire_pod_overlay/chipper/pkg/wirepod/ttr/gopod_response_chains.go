// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_response_chains.go
package wirepod_ttr

import (
	"math/rand"
	"strconv"
	"strings"
	"sync"
)

// The response-chain template (2026-10-25, operator: "Do response-chain template
// as proposed last time to make the rest inherent * this to me falls under,
// flexible data sets"). A response is its answers in turn - single (one Brobot)
// or duo (A, then B; multi later). Each answer is one Brobot's answer-chain, his
// thought bubbles (answer_chains.<chain>). Everything about an answer's place in
// the response lives in response_chains in the settings:
//   chain     - its answer-chain group
//   opener    - what he says first, while thinking (openers.<name>)
//   ends_with, no_reminder, keep_reminder_types, weather_type, no_manager - the
//              duo rules every answer type follows in that place (was duo)
//   plan      - the notice bubble's plan line for the response ("" = none)
// One nickname list (nicknames) serves every opener and hand-over.

type gopodChainAnswer struct {
	Chain             string   `json:"chain"`
	Opener            string   `json:"opener"`
	EndsWith          string   `json:"ends_with"`
	NoReminder        bool     `json:"no_reminder"`
	KeepReminderTypes []string `json:"keep_reminder_types"`
	WeatherType       string   `json:"weather_type"`
	NoManager         bool     `json:"no_manager"`
}

type gopodResponseChain struct {
	Plan    string             `json:"plan"`
	Answers []gopodChainAnswer `json:"answers"`
}

// gopodResponseSlot: this Brobot's response-chain and his place in it - duo 0
// (first in a line to both), duo 1 (following up), else single 0.
func gopodResponseSlot(esn string) (string, int) {
	if gopodIsFollowingUp(esn) {
		return "duo", 1
	}
	if _, first := gopodAllFirstAnswer.Load(esn); first {
		return "duo", 0
	}
	return "single", 0
}

// gopodChainAnswerFor: his answer's template in its response-chain.
func gopodChainAnswerFor(esn string) (gopodChainAnswer, bool) {
	kind, n := gopodResponseSlot(esn)
	rc, ok := gopodBrobotsSettings.ResponseChains[kind]
	if !ok || n >= len(rc.Answers) {
		return gopodChainAnswer{}, false
	}
	return rc.Answers[n], true
}

// gopodDuoRule: his duo answer's rules; false for a single answer.
func gopodDuoRule(esn string) (gopodChainAnswer, bool) {
	if kind, _ := gopodResponseSlot(esn); kind != "duo" {
		return gopodChainAnswer{}, false
	}
	return gopodChainAnswerFor(esn)
}

// gopodNicknameAt: each host's last nickname (never twice running).
var gopodNicknameAt = struct {
	sync.Mutex
	last map[string]int
}{last: map[string]int{}}

// gopodNicknameFill fills {title} and {nickname} in line for host: a random
// title (nicknames.titles) and a random nickname, never his last one
// (host_personas.<host>.handover.nicknames when he has his own, else
// nicknames.list).
func gopodNicknameFill(host, line string) string {
	nn := gopodBrobotsSettings.Nicknames
	list := gopodBrobotsSettings.HostPersonas[host].Handover.Nicknames
	if len(list) == 0 {
		list = nn.List
	}
	titles := gopodBrobotsSettings.HostPersonas[host].Handover.Titles
	if len(titles) == 0 {
		titles = nn.Titles
	}
	nick, title := "", ""
	if len(list) > 0 {
		gopodNicknameAt.Lock()
		last, seen := gopodNicknameAt.last[host]
		n := rand.Intn(len(list))
		for seen && len(list) > 1 && n == last {
			n = rand.Intn(len(list))
		}
		gopodNicknameAt.last[host] = n
		gopodNicknameAt.Unlock()
		nick = strings.TrimSpace(list[n])
	}
	if len(titles) > 0 {
		title = strings.TrimSpace(titles[rand.Intn(len(titles))])
	}
	line = strings.ReplaceAll(line, "{title}", title)
	line = strings.ReplaceAll(line, "{nickname}", nick)
	return strings.Join(strings.Fields(line), " ")
}

// gopodChainOpenerSaid: the opener each Brobot said for this answer (his notice
// shows the same words).
var gopodChainOpenerSaid sync.Map

// gopodOpenerPick: his opener for this answer (openers.<his answer's opener>),
// {got_it} his own Got it words; his Got it when there's none.
func gopodOpenerPick(esn string) string {
	text := gopodGotItFor(esn).Text
	if a, ok := gopodChainAnswerFor(esn); ok {
		if tpl := strings.TrimSpace(gopodBrobotsSettings.Openers[a.Opener]); tpl != "" {
			if t := gopodNicknameFill(gopodBrobotName(esn), strings.ReplaceAll(tpl, "{got_it}", text)); t != "" {
				text = t
			}
		}
	}
	gopodChainOpenerSaid.Store(esn, text)
	return text
}

// gopodOpenerText: the opener he said for this answer, else his Got it words.
func gopodOpenerText(esn string) string {
	if t, ok := gopodChainOpenerSaid.Load(esn); ok {
		return t.(string)
	}
	return gopodGotItFor(esn).Text
}

// gopodResponsePlan: the notice bubble's plan line for his response - who
// answers in turn ({order}), his place ({n} of {total}); "" when the response's
// plan is empty.
func gopodResponsePlan(esn string) string {
	kind, n := gopodResponseSlot(esn)
	rc := gopodBrobotsSettings.ResponseChains[kind]
	if strings.TrimSpace(rc.Plan) == "" {
		return ""
	}
	order := []string{gopodBrobotName(esn)}
	if kind == "duo" {
		brother := gopodBrobotName(gopodBrotherOf(esn))
		if n == 0 {
			order = append(order, brother)
		} else {
			order = []string{brother, gopodBrobotName(esn)}
		}
	}
	return gopodNoticeFill(rc.Plan, map[string]string{
		"order": strings.Join(order, ", then "),
		"n":     strconv.Itoa(n + 1),
		"total": strconv.Itoa(len(order)),
	})
}
