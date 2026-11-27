// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_lingo_improver.go
package wirepod_ttr

import (
	"context"
	"sort"
	"strings"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
)

// GopodLingoVoice: one voice the improver can try.
type GopodLingoVoice struct {
	Voice string `json:"voice"` // the speaker: "Brobot 1", "CHALK", "Goverlord" (KEYNOTE's voice)
	Label string `json:"label"` // as the chat names it
	Lingo string `json:"lingo"` // brobots | butler | reporters
	Robot bool   `json:"robot"` // said by the robot itself
}

// GopodLingoVoices: the two robots, the system voice, then every persona voice.
func GopodLingoVoices() []GopodLingoVoice {
	out := []GopodLingoVoice{
		{Voice: "Brobot 1", Label: "Brobot 1", Lingo: "brobots", Robot: true},
		{Voice: "Brobot 2", Label: "Brobot 2", Lingo: "brobots", Robot: true},
	}
	if sr := gopodBrobotsSettings.SignOff.SystemRole; sr != "" {
		out = append(out, GopodLingoVoice{Voice: sr, Label: gopodLingoLabel(gopodBrobotsSettings.SystemBubble.VoicePersona, sr), Lingo: "butler"})
	}
	var hosts []string
	for h := range gopodBrobotsSettings.ReporterVoices.Hosts {
		hosts = append(hosts, h)
	}
	sort.Strings(hosts)
	for _, h := range hosts {
		lingo := "reporters"
		if gopodSpeaksAsSystem(h) {
			lingo = "butler"
		}
		out = append(out, GopodLingoVoice{Voice: h, Label: h, Lingo: lingo})
	}
	return out
}

// gopodLingoLabel: a persona's nickname as the chat shows it, else fallback.
func gopodLingoLabel(persona, fallback string) string {
	if n := strings.TrimSpace(gopodBrobotsSettings.HostPersonas[persona].Nickname); n != "" {
		return n
	}
	return fallback
}

func gopodLingoVoiceOf(voice string) (GopodLingoVoice, bool) {
	for _, v := range GopodLingoVoices() {
		if v.Voice == voice {
			return v, true
		}
	}
	return GopodLingoVoice{}, false
}

// GopodLingoTry: text as voice says it (its lingo applied), and heard - the robot says it, or its
// voice file is made for the page to play (audio).
func GopodLingoTry(voice, text string) (said, audio, note string) {
	v, ok := gopodLingoVoiceOf(voice)
	text = strings.TrimSpace(text)
	if !ok || text == "" {
		return "", "", gopodSystemText("lingo_improver.pick_and_type")
	}
	if v.Robot {
		said = gopodApplyPronunciations(removeSpecialCharacters(text))
		esn := gopodBrobotESN(v.Voice)
		note = gopodRobotCall(esn, true, func(robot *vector.Vector, ctx context.Context) string {
			gopodLogSayText(robot, ctx, &vectorpb.SayTextRequest{Text: said, UseVectorVoice: true, DurationScalar: 1})
			return ""
		})
		return said, "", note
	}
	said = gopodSaysFor(v.Voice, text)
	page, _, made := gopodReporterFile(v.Voice, said)
	if !made {
		return said, "", gopodSystemText("lingo_improver.no_voice_file")
	}
	return said, page, ""
}

// GopodLingoKeep: word said as said, kept in voice's lingo (system-lingo).
func GopodLingoKeep(voice, word, said string) string {
	v, ok := gopodLingoVoiceOf(voice)
	word, said = strings.TrimSpace(word), strings.TrimSpace(said)
	if !ok || word == "" || said == "" {
		return gopodSystemText("lingo_improver.pick_and_type")
	}
	return gopodSystemLingo("system-lingo " + v.Lingo + " " + word + " = " + said)
}

// GopodSystemTexts: every system text whose key starts with prefix, the prefix off
// ("lingo_improver.title" -> "title") - a page's words, one source.
func GopodSystemTexts(prefix string) map[string]string {
	out := map[string]string{}
	for k, v := range gopodBrobotsSettings.SystemText {
		if strings.HasPrefix(k, prefix) {
			out[strings.TrimPrefix(k, prefix)] = v
		}
	}
	return out
}

// ## EOF
