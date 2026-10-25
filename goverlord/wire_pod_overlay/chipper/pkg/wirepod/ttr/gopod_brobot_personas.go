// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_brobot_personas.go
package wirepod_ttr

import (
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/sashabaranov/go-openai"
)

// Each brother's own persona line, added after "Current robot ESN: ..." in the
// Wire-Pod chat request. Both keep the shared bratty party-bro traits
// (wire-pod_brobots_prompt.txt), with a slight twist each - the seeds of the
// GOPOD layer's Doc and Pip. Operator 2026-10-01: "Brobot 1 slightly on the
// grumpy/impatient side", "Brobot 2 happily gets something a little wrong", and
// Doc's sarcastic take "at least attempted". "Up to 10" sentences, per the
// operator 2026-09-29: "Maybe down to 10 sentences? I like giving the brobots
// room to talk". 2026-10-04 (operator: "Party-bro stays the base, with one mild
// flavour each, held for the whole answer"): Brobot 1 party bro + grumpy +
// sarcastic, Brobot 2 party bro + clueless + positive outlook.
//
// The brother is not named here, or in the question tag. In a gemma2:2b test
// (2026-09-29), naming "Brobot 2" in Brobot 1's own request made Brobot 1 call
// himself Brobot 2 in 9 of 18 replies; without it, 0 of 18. gemma2 has no system
// slot, so the prompt sits at the top of the chat, far from the question - the
// last name it reads near the question is the one it takes.
// The lines are personas in gopod_brobots_settings.json, each followed by his
// host persona (host_personas: Doc & Pip, operator 2026-10-21).
var gopodBrobotPersonaLines = gopodHostPersonaLines()

// gopodHostPersona: one GOPOD host's persona - the same shape for Doc, Pip,
// CHALK and PLAYHEAD (host_personas in gopod_brobots_settings.json).
type gopodHostPersona struct {
	Name      string   `json:"name"`
	Role      string   `json:"role"`
	Steps     []string `json:"steps"`
	Signature string   `json:"signature"`
	Handover  struct {
		Line      string   `json:"line"`
		Nicknames []string `json:"nicknames"`
		Titles    []string `json:"titles"`
	} `json:"handover"`
}

// gopodHostPair: two hosts - their banter and stereo dial (host_pairs). Dial is
// how far right, 0-100: 0 all left (the base), 50 pointing up, 100 all right
// (the host persona).
type gopodHostPair struct {
	Hosts     []string `json:"hosts"`
	Left      string   `json:"left"`
	Dial      int      `json:"dial"`
	Voice     string   `json:"voice"`
	PartnerAs string   `json:"partner_as"`
	Banter    string   `json:"banter"`
}

// gopodHostPairOf: the host's pair and his partner's key.
func gopodHostPairOf(host string) (gopodHostPair, string, bool) {
	for _, pair := range gopodBrobotsSettings.HostPairs {
		if len(pair.Hosts) == 2 {
			if pair.Hosts[0] == host {
				return pair, pair.Hosts[1], true
			}
			if pair.Hosts[1] == host {
				return pair, pair.Hosts[0], true
			}
		}
	}
	return gopodHostPair{}, "", false
}

// gopodHostPersonaText renders a host's persona and his pair's banter and
// stereo dial as a few plain sentences, added after base (his own text).
func gopodHostPersonaText(host, base string) string {
	h := gopodBrobotsSettings.HostPersonas[host]
	pair, partner, paired := gopodHostPairOf(host)
	i, my := "I", "My"
	if strings.EqualFold(pair.Voice, "you") {
		i, my = "you", "Your"
	}
	var parts []string
	if n := strings.TrimSpace(h.Name); n != "" && !strings.Contains(base, n) {
		parts = append(parts, gopodPersonaFrame(gopodBrobotsSettings.HostPersonaFrames.Name, i, my, "{name}", n))
	}
	if r := strings.TrimSpace(h.Role); r != "" {
		parts = append(parts, r)
	}
	if len(h.Steps) > 0 {
		parts = append(parts, gopodPersonaFrame(gopodBrobotsSettings.HostPersonaFrames.Steps, i, my, "{steps}", strings.Join(h.Steps, ", then ")))
	}
	if sig := strings.TrimSpace(h.Signature); sig != "" {
		parts = append(parts, gopodPersonaFrame(gopodBrobotsSettings.HostPersonaFrames.Signature, i, my, "{signature}", sig))
	}
	if paired {
		as := strings.TrimSpace(pair.PartnerAs)
		if as == "" {
			as = gopodBrobotsSettings.HostPersonas[partner].Name
		}
		if t := strings.TrimSpace(pair.Banter); t != "" {
			parts = append(parts, strings.ReplaceAll(t, "{partner}", as))
		}
		if n := strings.TrimSpace(h.Name); n != "" && strings.TrimSpace(pair.Left) != "" && pair.Dial >= 0 && pair.Dial <= 100 {
			parts = append(parts, strings.NewReplacer("{left_pct}", strconv.Itoa(100-pair.Dial), "{left}", strings.TrimSpace(pair.Left), "{right_pct}", strconv.Itoa(pair.Dial), "{name}", n).Replace(gopodPersonaFrame(gopodBrobotsSettings.HostPersonaFrames.Dial, i, my, "", "")))
		}
	}
	return strings.Join(parts, " ")
}

// gopodPersonaFrame: one host persona sentence from its frame
// (host_persona_frames): {i}, {my} and the one value filled in.
func gopodPersonaFrame(frame, i, my, key, value string) string {
	r := []string{"{i}", i, "{my}", my}
	if key != "" {
		r = append(r, key, value)
	}
	return strings.NewReplacer(r...).Replace(frame)
}

// gopodHostPersonaLines: each brother's personas line, then his host persona.
func gopodHostPersonaLines() map[string]string {
	lines := map[string]string{}
	for name, line := range gopodBrobotsSettings.Personas {
		if t := gopodHostPersonaText(name, line); t != "" {
			line = strings.TrimSpace(line) + " " + t
		}
		lines[name] = line
	}
	return lines
}

// gopodPrebuildContext (2026-10-22, operator: "brobot gets an input *
// pre-build context: I am brobt N. My role is a GOPOD rep. I have a
// reminders-list to go through. What was asked of me? How to best answer?"):
// the section for his prompt, from prebuild_context; "" when off. A line whose
// token is empty is left out.
func gopodPrebuildContext(esn, question string) string {
	pc := gopodBrobotsSettings.PrebuildContext
	brobot := gopodBrobotName(esn)
	if strings.TrimSpace(pc.Header) == "" || (brobot != "Brobot 1" && brobot != "Brobot 2") {
		return ""
	}
	h := gopodBrobotsSettings.HostPersonas[brobot]
	fill := map[string]string{
		"brobot":    brobot,
		"host":      strings.TrimSpace(h.Name),
		"reminders": gopodReminderTopics(question),
		"question":  strings.TrimSpace(question),
		"steps":     strings.Join(h.Steps, ", then "),
	}
	// 2026-10-23 (A3): the score, so he answers from it too.
	score, _ := gopodCoverageScoreText()
	fill["score"] = score
	fill["game"] = strings.TrimSpace(gopodBrobotsSettings.Game.Rules)
	out := []string{strings.TrimSpace(pc.Header)}
	for _, line := range pc.Lines {
		empty := false
		for k, v := range fill {
			if strings.Contains(line, "{"+k+"}") && v == "" {
				empty = true
			}
			line = strings.ReplaceAll(line, "{"+k+"}", v)
		}
		if !empty && strings.TrimSpace(line) != "" {
			out = append(out, strings.TrimSpace(strings.ReplaceAll(line, " ()", "")))
		}
	}
	return strings.Join(out, "\n")
}

// gopodReminderTopics: the names in his "Hey, that reminds me" line, "" for none.
func gopodReminderTopics(question string) string {
	line := gopodReminderLine(question)
	tpl := gopodBrobotsSettings.ValuePoints.ReminderTopicsText
	at := strings.Index(tpl, "{topics}")
	if line == "" || at < 0 {
		return ""
	}
	before, after := tpl[:at], tpl[at+len("{topics}"):]
	if !strings.HasPrefix(line, before) || !strings.HasSuffix(line, after) || len(line) < len(before)+len(after) {
		return ""
	}
	return strings.TrimSpace(line[len(before) : len(line)-len(after)])
}

// gopodHandoverLine (2026-10-21, operator: "Each host's hand-over line * can we
// use variations of gear-head?"): the host's hand-over line to partner, {title}
// and {nickname} filled (gopodNicknameFill, gopod_response_chains.go); "" when
// he has none.
func gopodHandoverLine(host, partner string) string {
	line := strings.TrimSpace(gopodBrobotsSettings.HostPersonas[host].Handover.Line)
	if line == "" {
		return ""
	}
	return gopodNicknameFill(host, strings.ReplaceAll(line, "{partner}", partner))
}

// Each brother's own take on the Bingo test, added to the shared chat's Bingo
// snapshot note when a Bingo run is in the chat memory - so "how was Bingo?"
// gets his side of the story, in his own words.
// Operator 2026-10-01: Brobot 1 "the bingo tester being bothered by an
// over-enthusiastic brobot 2 brat, and getting brobot 2 back at the end"; the
// balls go "Real bingo number, then a test number, then an intentional
// non-existent number"; Brobot 2 "happy to be involved, but cluelessly" messed up.
// The real moments (operator 2026-10-01: the first takes were "too out there")
// are the Bingo test song's own lines (01a_brobots_bingo_test/story.md), in
// order - a change to those lines needs the same change in the settings file. The other brother
// is "your brother", not named (see above).
// The takes are bingo.takes in gopod_brobots_settings.json.
var gopodBrobotBingoTakes = gopodBrobotsSettings.Bingo.Takes

// gopodAddBingoTake adds this brother's Bingo take to the Bingo snapshot note in
// the shared chat (gopodBingoSnapshotNote). Nothing changes when there is no
// Bingo run in the chat memory.
// gopodBingoRunInMemory: a Bingo test run is in the shared chat memory (its
// snapshot note, gopod_session_memory.go).
func gopodBingoRunInMemory() bool {
	for _, msg := range gopodSharedBrobotChat() {
		if strings.HasPrefix(msg.Content, gopodBingoSnapshotNote) {
			return true
		}
	}
	return false
}

// gopodBingoPlayed: the Bingo test has been played since this Wire-Pod start
// (operator 2026-10-05: "a 'bingo-test-played' flag? Not played yet = 0, played
// already = 1"). Set by the song's own lines reaching the chat feed
// (gopod_chat_feed.go); a Wire-Pod restart sets it back to 0. Unlike the chat
// memory, it never fades: after the 120-minute memory has let the run go, a
// Bingo question still gets the Brobot's own take on it. With 0, a Bingo
// question gets bingo.no_run_note ("sure, let's play").
var gopodBingoPlayed atomic.Bool

func gopodAddBingoTake(esn string, chat []openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
	take, ok := gopodBrobotBingoTakes[gopodBrobotName(esn)]
	if !ok {
		return chat
	}
	for i := len(chat) - 1; i >= 0; i-- {
		if strings.HasPrefix(chat[i].Content, gopodBingoSnapshotNote) {
			chat[i].Content += "\n[" + take + "]"
			return chat
		}
	}
	// Played, but no longer in the chat memory: the snapshot and the take go first.
	if gopodBingoPlayed.Load() {
		note := openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: gopodBingoSnapshotNote + "\n[" + take + "]"}
		chat = append([]openai.ChatCompletionMessage{note}, chat...)
	}
	return chat
}

// ## EOF
