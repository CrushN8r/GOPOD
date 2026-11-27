// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_personas.go
package wirepod_ttr

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
)

// gopodPersonaFor: the host a voice-only persona talks through, or "".
func gopodPersonaFor(persona string) string {
	p, ok := gopodBrobotsSettings.Managers.Personas[strings.ToUpper(strings.TrimSpace(persona))]
	if !ok {
		return ""
	}
	return p.Host
}

// gopodPersonaPrompt: the persona's identity and its reporter's lines to talk like.
func gopodPersonaPrompt(host string) string { return gopodPersonaPromptFor(host, false) }

// gopodPersonaPromptFor: the prompt; for an intro, the profile's intro sections only
// (persona_profiles.intro_sections) - the opening says the name, nickname and why.
func gopodPersonaPromptFor(host string, intro bool) string {
	p := gopodBrobotsSettings.Managers.Personas[host]
	var b strings.Builder
	b.WriteString(gopodPersonaFill(host, p.Prompt))
	if t := gopodHostPersonaText(host, p.Prompt); t != "" && !intro {
		b.WriteString(" " + t)
	}
	profile := gopodPersonaProfileText(host)
	if is := gopodBrobotsSettings.PersonaProfiles.IntroSections; intro && len(is) > 0 {
		profile = gopodPersonaProfileTextOf(host, is)
	}
	if profile != "" {
		b.WriteString(" " + profile)
	}
	if n := gopodHereNow(host); n != "" && !intro {
		b.WriteString(" " + n)
	}
	if w := gopodPersonaWIP(host); w != "" && !intro {
		b.WriteString(" " + w)
	}
	var lines []string
	for _, l := range gopodReporterAllLines() {
		if l.Host == p.Reporter && len(lines) < 14 {
			lines = append(lines, "- "+l.Text)
		}
	}
	if len(lines) > 0 && !intro {
		b.WriteString("\n" + p.LinesNote + "\n" + strings.Join(lines, "\n"))
	}
	return b.String()
}

var gopodPersonaEmojiRE = regexp.MustCompile(`[\x{1F000}-\x{1FAFF}\x{2600}-\x{27BF}\x{FE0F}\x{200D}]`)

var gopodSpeakerLabelRE = regexp.MustCompile(`(^|[.!?]\s+)([A-Z][A-Za-z0-9 ]{1,30}):\s+`)

// gopodDropSpeakerLabels: "KEYNOTE: " at a sentence's start goes when it names a persona.
func gopodDropSpeakerLabels(reply string) string {
	return gopodSpeakerLabelRE.ReplaceAllStringFunc(reply, func(m string) string {
		sub := gopodSpeakerLabelRE.FindStringSubmatch(m)
		if (gopodAnyNamed(sub[2]) != "" || strings.EqualFold(strings.TrimSpace(sub[2]), gopodBrobotsSettings.BrainChain.Asked)) && len(strings.Fields(sub[2])) <= 4 {
			return sub[1]
		}
		return m
	})
}

// gopodPersonaAsk: the persona's answer from the language model (short; emojis welcome).
var gopodPersonaAsk = func(host, text string) (string, error) {
	base, model, ok := gopodWarmupOllama()
	if !ok || model == "" {
		return "", errors.New("no language model")
	}
	body, _ := json.Marshal(map[string]any{
		"model": model, "stream": false,
		"messages": gopodPersonaMessages(host, text),
		"options":  map[string]any{"num_predict": gopodPersonaNumPredict()},
	})
	client := &http.Client{Timeout: gopodTiming("persona_answer_seconds")}
	resp, err := client.Post(base+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	reply := strings.TrimSpace(out.Message.Content)
	reply = strings.Join(strings.Fields(strings.ReplaceAll(reply, "*", "")), " ")
	reply = strings.TrimSpace(gopodAnimationMarkerPattern.ReplaceAllString(reply, ""))
	reply = gopodPersonaStageRE.ReplaceAllString(reply, "")
	reply = gopodDropSpeakerLabels(reply)
	if q := strings.TrimSpace(text); len(strings.Fields(q)) >= 3 {
		reply = strings.Replace(reply, q, "", 1)
	}
	reply = gopodStripLeaks(reply)
	reply = gopodDropNeverWords(reply)
	reply = gopodPersonaSayTagRE.ReplaceAllString(reply, "")
	reply = strings.Join(strings.Fields(strings.NewReplacer("\u201c", "", "\u201d", "", "\"", "").Replace(reply)), " ")
	reply = gopodPersonaGapRE.ReplaceAllString(reply, "$1")
	reply = gopodStripEcho(reply, text)
	if reply == "" {
		return "", errors.New("the language model said nothing")
	}
	return gopodEndAtWholeSentence(reply), nil
}

// gopodPersonaNumPredict: managers.answer_num_predict, else 400.
func gopodPersonaNumPredict() int {
	if n := gopodBrobotsSettings.Managers.AnswerNumPredict; n > 0 {
		return n
	}
	return 400
}

// gopodEndAtWholeSentence: a reply that ran out mid-sentence ends at its last whole one.
func gopodEndAtWholeSentence(reply string) string {
	reply = strings.TrimSpace(reply)
	if reply == "" || strings.ContainsAny(reply[len(reply)-1:], ".!?\"')") {
		return reply
	}
	if at := strings.LastIndexAny(reply, ".!?"); at > 0 {
		return strings.TrimSpace(reply[:at+1])
	}
	return reply
}

var (
	gopodPersonaStageRE  = regexp.MustCompile(`\[[^\]]*\]|(?i)\((?:[^)]*\b(?:laugh|laughs|laughing|chuckle|chuckles|sigh|sighs|smile|smiles|grin|grins|pause|pauses|winks?|nods?|shrugs?|clears? (?:his|her|my) throat)\b[^)]*)\)`)
	gopodPersonaGapRE    = regexp.MustCompile(`\s+([,.!?;:])`)
	gopodPersonaSayTagRE = regexp.MustCompile(`(?i),?\s*\b(I|she|he) (say|says|said|chuckle|chuckles|laugh|laughs)\b[^.!?,"\x{201d}]*[,.]?`)
)

// gopodPersonaMessages: the persona's prompt, the Brobots' shared chat (the same recent exchanges
// the Brobots answer from), then the line.
func gopodPersonaMessages(host, text string) []map[string]string {
	msgs := []map[string]string{{"role": "system", "content": gopodPersonaPromptFor(host, gopodIsIntro(text))}}
	hn := gopodBrobotsSettings.HereNow
	if b := gopodBrainText(host, text); b != "" {
		msgs = append(msgs, map[string]string{"role": "user", "content": b})
	}
	var lines []string
	for _, m := range gopodSharedBrobotChat() {
		if !gopodBrobotsSettings.BrainChain.GomadsRawChat {
			break
		}
		c := strings.TrimSpace(m.Content)
		switch {
		case c == "":
		case m.Role == "assistant":
			lines = append(lines, strings.ReplaceAll(hn.ChatAnswer, "{line}", c))
		case m.Role == "user":
			lines = append(lines, c)
		}
	}
	if len(lines) > 0 {
		msgs = append(msgs, map[string]string{"role": "user", "content": hn.ChatNote + "\n" + strings.Join(lines, "\n")})
	}
	if intro := gopodIntroRecipe(host, text); intro != "" {
		text += "\n\n" + intro
	}
	return append(msgs, map[string]string{"role": "user", "content": text})
}

// gopodPersonaReply: the operator's line to a persona, and its answer.
func gopodPersonaReply(host, text, how string) {
	gopodPersonaReplyAs(host, host, text, how)
}

// gopodPersonaReplyAs: host's answer, said and shown as sayAs - its reporter, for the reporter
// suit.
func gopodPersonaReplyAs(host, sayAs, text, how string) {
	gopodOperatorShow(host, text, how)
	early := gopodGomadAnswerStart(host, sayAs, text)
	gopodOperatorRead(host, text, how)
	gopodGomadAnswerEarly(host, sayAs, text, early)
}

// gopodSpeaksAsSystem: host speaks with the butler's lingo (reporter_voices.
func gopodSpeaksAsSystem(host string) bool {
	sr := gopodBrobotsSettings.SignOff.SystemRole
	if sr == "" {
		return false
	}
	return host == sr || strings.EqualFold(host, sr) || strings.TrimSpace(gopodBrobotsSettings.Managers.Personas[host].VoiceOf) == sr
}

// gopodStripEcho: the reply without a leading copy of the question (its words, in order, at the
// reply's start - punctuation and case aside).
func gopodStripEcho(reply, question string) string {
	norm := func(s string) []string {
		var out []string
		for _, w := range strings.Fields(strings.ToLower(s)) {
			if w = strings.Trim(w, ".,!?;:'\"-"); w != "" {
				out = append(out, w)
			}
		}
		return out
	}
	q := norm(question)
	if len(q) < 4 {
		return reply
	}
	words := strings.Fields(reply)
	matched, wi := 0, 0
	for wi < len(words) && matched < len(q) {
		w := strings.ToLower(strings.Trim(words[wi], ".,!?;:'\"-"))
		if w != "" && w != q[matched] {
			break
		}
		wi++
		if w != "" {
			matched++
		}
	}
	if matched >= len(q) || (matched >= 4 && matched*2 >= len(q)) {
		if rest := strings.TrimSpace(strings.Join(words[wi:], " ")); rest != "" {
			return rest
		}
	}
	return reply
}

// ## EOF
