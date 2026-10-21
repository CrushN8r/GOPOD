// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_personas.go
package wirepod_ttr

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// The reporters' live personas (2026-10-18, operator: "Shared reporters persona =
// scripted + voices * reporter 1 & 2 lines = scaffolding for CHALK & PLAYHEAD
// live chat personas"; "KP7/8, Who are you? * I'm CHALK BORED, display manager
// at the CRUSHN8R CREW'd news station, CREW'd on location news reporter with my
// long time partner PLAYHEAD ZUMINTU * I'm PLAYHEAD ZUMINTU, session manager at
// the CRUSHN8R CREW'd news station, CREW'd on location news reporter with my long
// time partner CHALK BORED"). A voice-only key (KP7, KP8 in the persona map)
// talks to its persona's host (managers.personas): the operator's line, then the
// persona's answer from the language model - its identity, and its reporter's
// scripted lines as the style to talk in - said in its reporter voice, in its
// own bubble.

// gopodPersonaFor: the host a voice-only persona talks through, or "".
func gopodPersonaFor(persona string) string {
	p, ok := gopodBrobotsSettings.Managers.Personas[strings.ToUpper(strings.TrimSpace(persona))]
	if !ok {
		return ""
	}
	return p.Host
}

// gopodPersonaPrompt: the persona's identity and its reporter's lines to talk like.
func gopodPersonaPrompt(host string) string {
	p := gopodBrobotsSettings.Managers.Personas[host]
	var b strings.Builder
	b.WriteString(p.Prompt)
	var lines []string
	for _, l := range gopodReporterAllLines() {
		if l.Host == p.Reporter && len(lines) < 14 {
			lines = append(lines, "- "+l.Text)
		}
	}
	if len(lines) > 0 {
		b.WriteString("\n" + p.LinesNote + "\n" + strings.Join(lines, "\n"))
	}
	return b.String()
}

var gopodPersonaEmojiRE = regexp.MustCompile(`[\x{1F000}-\x{1FAFF}\x{2600}-\x{27BF}\x{FE0F}\x{200D}]`)

// gopodPersonaAsk: the persona's answer from the language model (short, no emojis).
var gopodPersonaAsk = func(host, text string) (string, error) {
	base, model, ok := gopodWarmupOllama()
	if !ok || model == "" {
		return "", errors.New("no language model")
	}
	body, _ := json.Marshal(map[string]any{
		"model": model, "stream": false,
		"messages": gopodPersonaMessages(host, text),
		"options": map[string]any{"num_predict": 120},
	})
	client := &http.Client{Timeout: 60 * time.Second}
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
	reply := strings.TrimSpace(gopodPersonaEmojiRE.ReplaceAllString(out.Message.Content, ""))
	reply = strings.Join(strings.Fields(strings.ReplaceAll(reply, "*", "")), " ")
	if reply == "" {
		return "", errors.New("the language model said nothing")
	}
	return reply, nil
}

// gopodPersonaMessages (2026-10-21, operator: "a manager is still a participant
// in the chat, so yes to context is required, so yes to chat history access"):
// the persona's prompt, the Brobots' shared chat (the same recent exchanges the
// Brobots answer from), then the line.
func gopodPersonaMessages(host, text string) []map[string]string {
	msgs := []map[string]string{{"role": "system", "content": gopodPersonaPrompt(host)}}
	for _, m := range gopodSharedBrobotChat() {
		if c := strings.TrimSpace(m.Content); c != "" && (m.Role == "user" || m.Role == "assistant") {
			msgs = append(msgs, map[string]string{"role": m.Role, "content": c})
		}
	}
	return append(msgs, map[string]string{"role": "user", "content": text})
}

// gopodPersonaReply: the operator's line to a persona, and its answer.
func gopodPersonaReply(host, text, how string) {
	gopodChatFeedAddOperator(host, text, gopodOperatorVia(how))
	// 2026-10-18 (operator: "KP7 to CHALK ... ["Generating..." + hourglass]/
	// [transcription]"): read in the Operator voice, like a line to the Brobots.
	gopodReporterReadsOperator(text, how, host)
	// "Who are you?" gets the operator's own line, word for word (managers.who_words).
	if who := strings.TrimSpace(gopodBrobotsSettings.Managers.Personas[host].WhoLine); who != "" {
		if at, _ := gopodEggWord(strings.ToLower(text), gopodBrobotsSettings.Managers.WhoWords); at >= 0 {
			logger.Println("GOPOD_PERSONA " + host + " (who): " + who)
			gopodReporterSay(host, who, true, true)
			return
		}
	}
	// 2026-10-18 (operator: "If I ask CHALK, PLAYHEAD ... = take over trigger"):
	// a take's words (the weather, the time and place) - its manager does the job.
	if take, _, ok := gopodManagerTakes(strings.ToLower(text)); ok {
		gopodManagerDo(take, text)
		return
	}
	reply, err := gopodPersonaAsk(host, text)
	if err != nil {
		logger.Println("GOPOD_PERSONA " + host + ": " + err.Error())
		return
	}
	logger.Println("GOPOD_PERSONA " + host + ": " + reply)
	gopodReporterSay(host, reply, true, true)
}

// ## EOF
