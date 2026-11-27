// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_brobot_personas.go
package wirepod_ttr

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
)

// Each brother's own persona line, added after "Current robot ESN: ..." in the Wire-Pod chat
// request.
var gopodBrobotPersonaLines = gopodHostPersonaLines()

// gopodHostPersona: one GOPOD host's persona - the same shape for Doc, Pip, CHALK and PLAYHEAD
// (host_personas in gopod_brobots_settings.json).
type gopodHostPersona struct {
	Name      string   `json:"name"`
	Nickname  string   `json:"nickname"`
	JobName   string   `json:"job_name"`
	Body      string   `json:"body,omitempty"`
	CallName  string   `json:"call_name"`
	Role      string   `json:"role"`
	Steps     []string `json:"steps"`
	Signature string   `json:"signature"`
	Handover  struct {
		Line      string   `json:"line"`
		Nicknames []string `json:"nicknames"`
		Titles    []string `json:"titles"`
	} `json:"handover"`
}

// gopodHostPair: two hosts - their banter and stereo dial (host_pairs).
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

// gopodHostPersonaText renders a host's persona and his pair's banter and stereo dial as a few
// plain sentences, added after base (his own text).
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
		if t := strings.TrimSpace(pair.Banter); t != "" && !(gopodSolo() && (host == "Brobot 1" || host == "Brobot 2")) {
			parts = append(parts, strings.ReplaceAll(t, "{partner}", as))
		}
		if n := strings.TrimSpace(h.Name); n != "" && strings.TrimSpace(pair.Left) != "" && pair.Dial >= 0 && pair.Dial <= 100 {
			parts = append(parts, strings.NewReplacer("{left_pct}", strconv.Itoa(100-pair.Dial), "{left}", strings.TrimSpace(pair.Left), "{right_pct}", strconv.Itoa(pair.Dial), "{name}", n).Replace(gopodPersonaFrame(gopodBrobotsSettings.HostPersonaFrames.Dial, i, my, "", "")))
		}
	}
	return strings.Join(parts, " ")
}

// gopodPersonaFrame: one host persona sentence from its frame (host_persona_frames): {i}, {my} and
// the one value filled in.
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
		lines[name] = gopodPersonaFillAs(name, line)
	}
	return lines
}

// gopodPrebuildContext: the section for his prompt, from prebuild_context; "" when off.
func gopodPrebuildContext(esn, question string) string {
	pc := gopodBrobotsSettings.PrebuildContext
	brobot := gopodBrobotName(esn)
	if strings.TrimSpace(pc.Header) == "" || (brobot != "Brobot 1" && brobot != "Brobot 2") {
		return ""
	}
	h := gopodPersona(brobot)
	fill := map[string]string{
		"brobot":    brobot,
		"host":      strings.TrimSpace(h.Name),
		"reminders": gopodReminderTopics(question),
		"question":  strings.TrimSpace(question),
		"steps":     strings.Join(h.Steps, ", then "),
		"role":      strings.TrimSpace(h.JobName),
	}
	if gopodAtDialZero(brobot) || fill["role"] == "" {
		fill["role"] = strings.TrimSpace(pc.RepRole)
	}
	score, _ := gopodCoverageScoreText()
	fill["score"] = score
	fill["game"] = strings.TrimSpace(gopodBrobotsSettings.Game.Rules)
	fill["place"], fill["time"] = gopodHereNowPlaceTime()
	fill["gender"] = gopodHereNowGender(brobot)
	fill["cast"] = gopodHereNowCast(brobot)
	fill["plan"] = gopodResponsePlan(esn)
	fill["notice"] = gopodNoticeGuide(esn)
	if gopodSolo() {
		fill["solo"] = strings.TrimSpace(gopodBrobotsSettings.BrobotsSetup.SoloNote)
	} else {
		fill["solo"] = ""
	}
	gopodDocBSAskAhead(esn, question)
	fill["roast"] = gopodRoastLine(question)
	fill["spice"] = gopodSpiceLine(question)
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

// gopodCallName: the name others call a host by - in a thank-you (helper_from call_name) and a
// hand-over ({intern}): host_personas.<host>.
func gopodPersonaShown(body string) string {
	return gopodAnswerer(body)
}

// gopodPersona: the labels of the persona answering as name (gopodAnswerer) - a body's are the
// persona's on it now, never the body's own when it hosts one.
func gopodPersona(name string) gopodHostPersona {
	return gopodBrobotsSettings.HostPersonas[gopodAnswerer(name)]
}

func gopodCallName(host string) string {
	host = gopodAnswerer(host)
	if gopodIsBrobotBody(host) {
		return host
	}
	p := gopodBrobotsSettings.HostPersonas[host]
	for _, n := range []string{p.CallName, p.Nickname, p.Name, host} {
		if n = strings.TrimSpace(n); n != "" {
			return n
		}
	}
	return ""
}

// gopodHandoverLine: the host's hand-over line to partner, {title} and {nickname} filled
// (gopodNicknameFill, gopod_response_chains.go); "" when he has none.
func gopodHandoverLine(host, partner string) string {
	line := strings.TrimSpace(gopodBrobotsSettings.HostPersonas[host].Handover.Line)
	if line == "" {
		return ""
	}
	return gopodNicknameFill(host, strings.ReplaceAll(line, "{partner}", partner))
}

// gopodBingoPlayed: the Bingo test has been played since this Wire-Pod start.
var gopodBingoPlayed atomic.Bool

// gopodApplyNicknames: a host with a nickname is headed with reporter_voices.nickname_label -
// "Buzzer (Intel 404 PNFH)".
func gopodApplyNicknames(s *gopodBrobotsSettingsFile) {
	tpl := strings.TrimSpace(s.ReporterVoices.NicknameLabel)
	if tpl == "" {
		return
	}
	for host, hp := range s.HostPersonas {
		if strings.TrimSpace(hp.Nickname) == "" {
			continue
		}
		if s.ReporterVoices.Labels == nil {
			s.ReporterVoices.Labels = map[string]string{}
		}
		s.ReporterVoices.Labels[host] = strings.NewReplacer("{nickname}", hp.Nickname, "{name}", hp.Name).Replace(tpl)
	}
}

// gopodPersonaFill: a persona's words with its {name}, {nickname} and {job} filled from
// host_personas - so a rename is those labels alone.
func gopodPersonaFill(host, text string) string {
	if !strings.Contains(text, "{") {
		return text
	}
	return gopodPersonaFillAs(gopodAnswerer(host), text)
}

// gopodPersonaFillAs: gopodPersonaFill for a persona's own name (Brobot 1, DOC) - its own labels,
// never resolved through a body.
func gopodPersonaFillAs(persona, text string) string {
	if !strings.Contains(text, "{") {
		return text
	}
	text = gopodNameVarsFill(text, gopodBrobotsSettings.HostPersonas)
	p := gopodBrobotsSettings.HostPersonas[persona]
	return strings.NewReplacer("{name}", p.Name, "{nickname}", p.Nickname, "{job}", p.JobName).Replace(text)
}

var gopodPersonaVarRE = regexp.MustCompile(`\{([^{}.]+)\.(name|nickname|job|call)\}`)

// gopodNameVarsFill: text with every {<persona>.name|nickname|job|call} filled from hps (a
// persona's labels, from its profile head); an unknown one stays.
func gopodNameVarsFill(text string, hps map[string]gopodHostPersona) string {
	if !strings.Contains(text, ".") || !strings.Contains(text, "{") {
		return text
	}
	return gopodPersonaVarRE.ReplaceAllStringFunc(text, func(m string) string {
		parts := gopodPersonaVarRE.FindStringSubmatch(m)
		o, ok := hps[parts[1]]
		if !ok {
			return m
		}
		switch parts[2] {
		case "name":
			return o.Name
		case "nickname":
			return o.Nickname
		case "job":
			return o.JobName
		}
		if o.CallName != "" {
			return o.CallName
		}
		return o.Name
	})
}

// gopodApplyNameVars: every string in the settings (struct fields, map values, list items, at any
// depth) with its persona variables filled (gopodNameVarsFill).
func gopodApplyNameVars(s *gopodBrobotsSettingsFile) {
	hps := s.HostPersonas
	var walk func(v reflect.Value)
	walk = func(v reflect.Value) {
		switch v.Kind() {
		case reflect.Ptr, reflect.Interface:
			if !v.IsNil() {
				walk(v.Elem())
			}
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				if f := v.Field(i); f.CanSet() || f.Kind() == reflect.Struct || f.Kind() == reflect.Map || f.Kind() == reflect.Slice {
					walk(f)
				}
			}
		case reflect.Slice, reflect.Array:
			for i := 0; i < v.Len(); i++ {
				walk(v.Index(i))
			}
		case reflect.Map:
			for _, k := range v.MapKeys() {
				e := v.MapIndex(k)
				if e.Kind() == reflect.String {
					if s := gopodNameVarsFill(e.String(), hps); s != e.String() {
						v.SetMapIndex(k, reflect.ValueOf(s).Convert(e.Type()))
					}
					continue
				}
				c := reflect.New(e.Type()).Elem()
				c.Set(e)
				walk(c)
				v.SetMapIndex(k, c)
			}
		case reflect.String:
			if v.CanSet() {
				v.SetString(gopodNameVarsFill(v.String(), hps))
			}
		}
	}
	walk(reflect.ValueOf(s).Elem())
}

// ## EOF
