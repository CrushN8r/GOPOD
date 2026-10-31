// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_source_check.go
package wirepod_ttr

// The source check (2026-10-31, operator: "Please ensure all my system checks,
// check from the root of the single source of truth & check the builds up from
// there concurrently, so lower level bugs are caught to reduce this wasteful
// direction we're going in"). Part of system-check, at every start too. For every
// chat participant at once, from his master source up, layer by layer:
//
//	0 source   his profile (the master format) reads; its head has his label;
//	           his Self has every line of the intro (nickname, name, job title,
//	           role, who helps me) - a line still "(to come)" is named
//	1 labels   everything filled from his labels has nothing left unfilled:
//	           his profile, prompt, who-line, cast line, intro recipe, brain
//	2 voice    a Gomad has a voice (his own or voice_of's)
//	3 the chat he's in the cast - everyone else knows he's here
//	4 the keys every live numpad key reaches someone
//	5 headers  the system label fills from its voice persona's labels

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
)

var gopodUnfilledRE = regexp.MustCompile(`\{(name|nickname|job|self|cast|first|middle|last|fullname|task|asked)\}`)

// gopodSelfLines: the Self lines every profile has, for the intro.
var gopodSelfLines = []string{"My informal name (nickname)", "My name", "My job title", "My role", "Who helps me"}

// gopodSourceCheck: the problems, and how many things were checked.
func gopodSourceCheck() ([]string, int) {
	whos := map[string]bool{"Brobot 1": true, "Brobot 2": true}
	for k := range gopodBrobotsSettings.Managers.Personas {
		whos[k] = true
	}
	var mu sync.Mutex
	var problems []string
	checked := 0
	add := func(p ...string) {
		mu.Lock()
		problems = append(problems, p...)
		mu.Unlock()
	}
	count := func(n int) {
		mu.Lock()
		checked += n
		mu.Unlock()
	}
	var wg sync.WaitGroup
	for who := range whos {
		wg.Add(1)
		go func(who string) {
			defer wg.Done()
			p, n := gopodSourceCheckPersona(who)
			add(p...)
			count(n)
		}(who)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		p, n := gopodSourceCheckKeys()
		add(p...)
		count(n)
	}()
	wg.Wait()
	// 5 headers: the system label from its voice persona's labels.
	count(1)
	if sb := gopodBrobotsSettings.SystemBubble; strings.TrimSpace(sb.VoicePersona) != "" {
		hp := gopodBrobotsSettings.HostPersonas[sb.VoicePersona]
		if strings.TrimSpace(hp.Name) == "" || (strings.Contains(sb.Label, "{job}") && strings.TrimSpace(hp.JobName) == "") {
			add("headers: system_bubble.label " + sb.Label + " can't fill from " + sb.VoicePersona + "'s labels (its profile head)")
		}
	}
	sort.Strings(problems)
	return problems, checked
}

// gopodSourceCheckPersona: one participant, from his source up.
func gopodSourceCheckPersona(who string) ([]string, int) {
	var out []string
	n := 0
	bad := func(layer, what string) { out = append(out, "source "+who+" ("+layer+"): "+what) }
	// 0 source
	n++
	head, sections := gopodPersonaProfileRead(gopodPersonaProfileFile(who))
	if head == nil {
		bad("0 source", "no profile ("+gopodPersonaProfileFile(who)+")")
		return out, n
	}
	if strings.TrimSpace(head["label"]) == "" {
		bad("0 source", "his profile has no \"# <full name>\" label")
	}
	raw := gopodPersonaProfileRawSelf(who)
	for _, label := range gopodSelfLines {
		n++
		v, ok := raw[label]
		switch {
		case !ok:
			bad("0 source", "his Self has no \""+label+"\" line (the intro needs it)")
		case strings.HasSuffix(v, "(to come)"):
			bad("0 source", "his Self \""+label+"\" is still to come")
		}
	}
	// 1 labels: nothing left unfilled, wherever his labels go.
	_, persona := gopodBrobotsSettings.Managers.Personas[who]
	filled := map[string]string{
		"profile":      gopodPersonaProfileText(who),
		"intro recipe": gopodIntroRecipe(who, "introduce yourself"),
	}
	if persona {
		filled["prompt"] = gopodPersonaPrompt(who)
		filled["who-line"] = gopodPersonaFill(who, gopodBrobotsSettings.Managers.Personas[who].WhoLine)
	}
	for _, c := range gopodBrobotsSettings.HereNow.Cast {
		if c.Who == who {
			filled["cast line"] = gopodPersonaFill(who, c.Line)
		}
	}
	for what, text := range filled {
		n++
		if m := gopodUnfilledRE.FindString(text); m != "" {
			bad("1 labels", "his "+what+" has "+m+" unfilled")
		}
	}
	_ = sections
	if persona {
		// 2 voice
		n++
		if gopodReporterVoice(who) == "" {
			bad("2 voice", "no voice (reporter_voices.hosts, or voice_of)")
		}
	}
	// 3 the chat
	n++
	inCast := false
	for _, c := range gopodBrobotsSettings.HereNow.Cast {
		inCast = inCast || c.Who == who
	}
	if !inCast {
		bad("3 the chat", "not in here_now.cast - nobody else knows he's here")
	}
	return out, n
}

// gopodPersonaProfileRawSelf: his Self lines by label, as written (to come too).
func gopodPersonaProfileRawSelf(who string) map[string]string {
	out := map[string]string{}
	path := gopodPersonaProfileFile(who)
	data := gopodReadFileString(path)
	inSelf := false
	for _, line := range strings.Split(data, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "## ") {
			inSelf = strings.TrimSpace(t[3:]) == "Self"
			continue
		}
		if !inSelf {
			continue
		}
		if m := gopodProfilePointRE.FindStringSubmatch(t); m != nil {
			if k, v, ok := strings.Cut(m[1], ":"); ok {
				out[strings.TrimSpace(k)] = strings.TrimSpace(v)
			}
		}
	}
	return out
}

// gopodSourceCheckKeys: 4 the keys - every live numpad key reaches someone.
func gopodSourceCheckKeys() ([]string, int) {
	m, _, err := gopodNumpadMap()
	if err != nil {
		return []string{"source numpad (4 the keys): the map can't be read: " + err.Error()}, 1
	}
	var out []string
	n := 0
	try := func(key string) {
		n++
		if _, _, err := gopodNumpadBrobot(key); err != nil {
			out = append(out, "source numpad (4 the keys): "+key+" is live in the map but reaches nobody: "+err.Error())
		}
	}
	for key, e := range m.NumlockOn {
		if e.Status == "live" {
			try(key)
		}
	}
	for key, e := range m.NumlockOff {
		if e.Status == "live" {
			try(key + gopodNumlockOffSuffix)
		}
	}
	try(gopodNumpadAllKey)
	return out, n
}

// gopodReadFileString: a file's text, "" when it can't be read.
func gopodReadFileString(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// ## EOF
