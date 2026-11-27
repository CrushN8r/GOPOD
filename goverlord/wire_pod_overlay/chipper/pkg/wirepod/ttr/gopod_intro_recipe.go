// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_intro_recipe.go
package wirepod_ttr

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// gopodIsIntro: the line asks for an intro.
func gopodIsIntro(text string) bool {
	t := strings.ToLower(text)
	for _, w := range gopodBrobotsSettings.IntroRecipe.Words {
		if w = strings.ToLower(strings.TrimSpace(w)); w != "" && strings.Contains(t, w) {
			return true
		}
	}
	return false
}

// gopodIntroWeather: the line asks for an intro and the weather - his intro first, then his weather
// (group-chat-intro.md, then weather_run.md).
func gopodIntroWeather(text string) bool {
	return gopodIsIntro(text) && gopodHasWholeWords(strings.ToLower(text), "weather")
}

// gopodIntroWeatherFor: a Brobot's answer now is an intro + weather (brobots_intro_weather).
var gopodIntroWeatherFor sync.Map // esn -> bool

// gopodIntroRecipe: the recipe for who, "" when the line isn't an intro.
func gopodIntroRecipe(who, text string) string {
	ir := gopodBrobotsSettings.IntroRecipe
	if strings.TrimSpace(ir.Text) == "" || !gopodIsIntro(text) {
		return ""
	}
	var facts []string
	for _, label := range []string{"My job title"} {
		if v := strings.TrimSuffix(strings.TrimSpace(gopodPersonaProfileField(who, label)), "."); v != "" {
			facts = append(facts, v)
		}
	}
	sections := gopodPersonaProfileSections(who)
	for _, part := range ir.Relations {
		head, label, ok := strings.Cut(part, "|")
		if !ok {
			continue
		}
		var names []string
		for _, line := range sections[strings.TrimSpace(head)] {
			if line = strings.TrimSpace(gopodPersonaFill(who, line)); line != "" && !strings.Contains(line, "(to come)") {
				names = append(names, line)
			}
		}
		if len(names) > 0 {
			facts = append(facts, strings.TrimSpace(strings.TrimSpace(label)+" "+strings.Join(names, " ")))
		}
	}
	if o := strings.TrimSpace(gopodBrobotsSettings.BrainChain.IntroOpener); o != "" && strings.EqualFold(gopodCallName(who), o) {
		if outline := sections["Our outline"]; len(outline) > 0 {
			facts = nil
			for _, p := range outline {
				facts = append(facts, strings.TrimSuffix(strings.TrimSpace(gopodPersonaFill(who, p)), "."))
			}
		}
	}
	text2 := ir.Text
	if (who == "Brobot 1" || who == "Brobot 2") && gopodIntroWeather(text) {
		esn := gopodBrobot1ESN
		if who == "Brobot 2" {
			esn = gopodBrobot2ESN
		}
		take := ir.WeatherText
		if gopodIsFollowingUp(esn) {
			take = ir.WeatherTextNext
		}
		text2 = strings.TrimSpace(text2 + " " + take)
	}
	return strings.TrimSpace(strings.NewReplacer("{opening}", gopodIntroOpening(who), "{facts}", strings.Join(facts, "; "),
		"{self}", strings.Join(facts, "; "), "{cast}", gopodIntroCast(who)).Replace(text2))
}

// gopodIntroOpening: "I'm {name}{nickname_part}." from his profile - his name (formal), his
// nickname (informal) and why; each part left out when blank.
func gopodIntroOpening(who string) string { return gopodIntroOpeningWith(who, true) }

// gopodIntroOpeningWith: the opening, with or without its why.
func gopodIntroOpeningWith(who string, withWhy bool) string {
	ir := gopodBrobotsSettings.IntroRecipe
	field := func(label string) string { return gopodPersonaProfileField(who, label) }
	first := func(v string) string {
		if at := strings.Index(v, " - "); at > 0 {
			v = v[:at]
		}
		if at := strings.Index(v, ". "); at > 0 {
			v = v[:at]
		}
		return strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(v), "."))
	}
	head, _ := gopodPersonaProfileRead(gopodPersonaProfileFile(who))
	name := first(field("My name"))
	if name == "" {
		name = head["label"]
	}
	nick := first(field("My informal name (nickname)"))
	part := ""
	if nick != "" && !strings.EqualFold(nick, name) {
		part = strings.ReplaceAll(ir.NicknamePart, "{nickname}", nick)
		if why := strings.TrimSpace(gopodPersonaFill(who, head["why"])); why != "" && withWhy {
			part += strings.ReplaceAll(ir.WhyPart, "{why}", why)
		}
	}
	out := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(ir.Opening, "{name}", name), "{nickname_part}", part))
	if body := strings.TrimRight(strings.TrimSuffix(out, "."), "'\"’”"); strings.HasSuffix(out, ".") && body != "" && strings.ContainsAny(body[len(body)-1:], ".!?") {
		out = strings.TrimSuffix(out, ".")
	}
	return out
}

// gopodIntroCastLine: intro_recipe.cast_line with everyone else here.
func gopodIntroCastLine(who string) string {
	cast := gopodIntroCast(who)
	if cast == "" {
		return ""
	}
	return strings.TrimSpace(strings.ReplaceAll(gopodBrobotsSettings.IntroRecipe.CastLine, "{cast}", cast))
}

// gopodIntroWrap: the model's middle part with the code's opening before it and the cast line after
// it.
func gopodIntroFacts(who string) string {
	facts := gopodBrainIntroFact(gopodBrainFacts(who, gopodIntroFactAsk))
	if who != gopodBrobotsSettings.SystemBubble.VoicePersona {
		return facts
	}
	var company []string
	for _, l := range gopodBrobotsSettings.CompanyIntro.Lines {
		if l = strings.TrimSpace(gopodPersonaFill(who, l)); l != "" {
			company = append(company, l)
		}
	}
	return strings.TrimSpace(strings.Join(company, " ") + " " + facts)
}

func gopodIntroWrap(who, middle string, brobot ...bool) string {
	if gopodIntroGuide(who) != "" {
		selfNames := gopodIntroSelfNames(who)
		var keep []string
		var said []map[string]bool
		others := gopodIntroHeardBy(who)
		middle = gopodVoiceNarrationRE.ReplaceAllString(middle, "")
		for i, sentence := range gopodIntroSentences(strings.ReplaceAll(middle, "*", "")) {
			if i > 0 {
				sentence = gopodDropSelfIntro(sentence, selfNames)
			}
			bare := strings.TrimSpace(sentence)
			if bare == "" || strings.Contains(strings.ToLower(bare), "in the chat") || gopodSaysAgain(bare, said) || gopodSaysAgain(bare, others) {
				continue
			}
			said = append(said, gopodContentWords(bare))
			keep = append(keep, sentence)
		}
		gopodIntroHeardAdd(who, said)
		return gopodIntroGuideWithColour(who, keep)
	}
	opening := gopodIntroOpening(who)
	middle = strings.ReplaceAll(middle, "*", "")
	middle = strings.TrimSpace(gopodStripEcho(middle, opening))
	if opening != "" {
		middle = strings.TrimSpace(strings.Replace(middle, opening, "", 1))
	}
	names := strings.Split(gopodIntroCast(who), ", ")
	selfNames := gopodIntroSelfNames(who)
	core := gopodIntroOpeningWith(who, false)
	var keep []string
	// What the code says in this intro - the opening and the facts - a sentence each.
	var said []map[string]bool
	for _, s := range gopodIntroSentences(opening + " " + gopodIntroFacts(who)) {
		said = append(said, gopodContentWords(s))
	}
	carry := "" // a dropped sentence's animation cue, kept for the next one
	for _, sentence := range gopodIntroSentences(middle) {
		if carry != "" {
			sentence, carry = carry+" "+strings.TrimSpace(sentence), ""
		}
		cue := ""
		if loc := gopodAnimationMarkerPattern.FindStringIndex(sentence); loc != nil && strings.TrimSpace(sentence[:loc[0]]) == "" {
			cue = strings.TrimSpace(sentence[:loc[1]])
		}
		bare := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(sentence), cue))
		if opening != "" && (gopodIsOpeningAgain(bare, opening, core) || gopodSaysNameAndNickname(who, bare)) {
			carry = cue
			continue
		}
		sentence = gopodDropSelfIntro(sentence, selfNames)
		if strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(sentence), cue)) == "" {
			carry = cue
			continue
		}
		if strings.Contains(strings.ToLower(sentence), "in the chat") {
			carry = cue
			continue
		}
		if gopodSaysAgain(bare, said) {
			carry = cue
			continue
		}
		said = append(said, gopodContentWords(bare))
		n := 0
		for _, name := range names {
			if name != "" && strings.Contains(strings.ToLower(sentence), strings.ToLower(name)) {
				n++
			}
		}
		if n < 3 {
			keep = append(keep, sentence)
		} else {
			carry = cue
		}
	}
	facts := gopodIntroFacts(who)
	if len(brobot) > 0 && brobot[0] {
		out := strings.TrimSpace(facts + " " + strings.Join(keep, " "))
		if opening != "" {
			out = gopodStitchSnippet(opening, out)
		}
		if cast := gopodIntroCastLine(who); cast != "" {
			out = gopodStitchEndSnippet(cast, out)
		}
		return strings.TrimSpace(out)
	}
	out := []string{opening, facts, strings.Join(keep, " "), gopodIntroCastLine(who)}
	var parts []string
	for _, p := range out {
		if p = strings.TrimSpace(p); p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, " ")
}

// gopodIntroGuide: the operator's own intro on who's card ("## My intro guide"), "" for none.
func gopodIntroGuide(who string) string {
	return strings.TrimSpace(strings.Join(gopodPersonaProfileSections(who)["My intro guide"], " "))
}

// gopodContentWords: a sentence's words of 3+ letters, small letters.
func gopodContentWords(s string) map[string]bool {
	out := map[string]bool{}
	for _, w := range strings.Fields(gopodLooseWords(s)) {
		if len(w) >= 3 {
			out[w] = true
		}
	}
	return out
}

// gopodSaysAgain: a sentence whose words are mostly (70%) in one sentence already said.
func gopodSaysAgain(sentence string, said []map[string]bool) bool {
	words := gopodContentWords(sentence)
	if len(words) == 0 {
		return false
	}
	for _, s := range said {
		in := 0
		for w := range words {
			if s[w] {
				in++
			}
		}
		if float64(in) >= 0.7*float64(len(words)) {
			return true
		}
	}
	return false
}

// gopodLooseWords: text as its words only - small letters, no dashes or marks.
func gopodLooseWords(text string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(text) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		case b.Len() > 0 && !strings.HasSuffix(b.String(), " "):
			b.WriteByte(' ')
		}
	}
	return strings.TrimSpace(b.String())
}

// gopodIsOpeningAgain: the sentence is the opening said again - it starts with the opening's core
// ("I'm <name>, my nickname is <nick>"), and what follows is nothing or more of the opening's own
// words ("- because it's my name").
func gopodIsOpeningAgain(sentence, opening, core string) bool {
	s, c := gopodLooseWords(sentence), gopodLooseWords(core)
	if c == "" || (s != c && !strings.HasPrefix(s, c+" ")) {
		return false
	}
	rest := strings.TrimSpace(strings.TrimPrefix(s, c))
	return rest == "" || strings.Contains(gopodLooseWords(opening), rest)
}

// gopodIntroSelfNames: the names who goes by - his call name, his name, his nickname.
func gopodIntroSelfNames(who string) []string {
	first := func(v string) string {
		if at := strings.Index(v, " - "); at > 0 {
			v = v[:at]
		}
		return strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(v), "."))
	}
	var out []string
	for _, n := range []string{who, gopodCallName(who), first(gopodPersonaProfileField(who, "My name")), first(gopodPersonaProfileField(who, "My informal name (nickname)"))} {
		if n != "" {
			out = append(out, n)
		}
	}
	sort.Slice(out, func(a, b int) bool { return len(out[a]) > len(out[b]) })
	return out
}

// gopodDropSelfIntro: the sentence without a starting "I'm <his name>[, <his name>]," - what
// follows starts the sentence ("and hey!" -> "And hey!").
func gopodDropSelfIntro(sentence string, names []string) string {
	lead := ""
	if loc := gopodAnimationMarkerPattern.FindStringIndex(sentence); loc != nil && loc[0] == 0 {
		lead, sentence = sentence[:loc[1]]+" ", strings.TrimSpace(sentence[loc[1]:])
	}
	low := strings.ToLower(strings.TrimSpace(sentence))
	rest := ""
	for _, prefix := range []string{"i'm ", "i\u2019m ", "i am "} {
		if strings.HasPrefix(low, prefix) {
			rest = strings.TrimSpace(sentence)[len(prefix):]
			break
		}
	}
	if rest == "" {
		return lead + sentence
	}
	matched := false
	for again := true; again; {
		again = false
		for _, n := range names {
			if len(rest) >= len(n) && strings.EqualFold(rest[:len(n)], n) && (len(rest) == len(n) || strings.ContainsRune(" ,.!;", rune(rest[len(n)])) || (rest[len(n)] == '-' && (len(rest) == len(n)+1 || rest[len(n)+1] == ' '))) {
				rest = strings.TrimLeft(rest[len(n):], " ,;-")
				matched, again = true, true
			}
		}
	}
	if !matched {
		return lead + sentence
	}
	rest = strings.TrimSpace(strings.TrimPrefix(strings.TrimLeft(rest, ".!"), " "))
	low = strings.ToLower(rest)
	if strings.HasPrefix(low, "and ") {
		rest = rest[4:]
	}
	if rest == "" || strings.Trim(rest, ".!?, ") == "" {
		return ""
	}
	r := []rune(rest)
	r[0] = unicode.ToUpper(r[0])
	return lead + string(r)
}

// gopodIntroSentences: text cut after each .
func gopodIntroSentences(text string) []string {
	var out []string
	start := 0
	for i, r := range text {
		if (r == '.' || r == '!' || r == '?') && (i+1 == len(text) || text[i+1] == ' ') {
			out = append(out, strings.TrimSpace(text[start:i+1]))
			start = i + 1
		}
	}
	if rest := strings.TrimSpace(text[start:]); rest != "" {
		out = append(out, rest)
	}
	return out
}

// gopodIntroCast: everyone else in the chat, by the names they're called.
func gopodIntroCast(who string) string {
	var cast []string
	for _, c := range gopodBrobotsSettings.HereNow.Cast {
		if c.Who != who && c.Who != gopodBrobotsSettings.BrainChain.OperatorWho {
			cast = append(cast, gopodCallName(c.Who))
		}
	}
	return strings.Join(cast, ", ")
}

// gopodIntroThenWeather: a Brobot's intro + weather answer - his intro (the opening, his first
// line, the cast line), then the weather report or "Confirmed." (gopodScriptedSnippet), then his
// take (the rest of his words).
func gopodIntroThenWeather(esn, question, raw string) string {
	who := gopodBrobotName(esn)
	intro, take := strings.TrimSpace(raw), ""
	if at := strings.Index(intro, "\n"); at > 0 {
		intro, take = strings.TrimSpace(intro[:at]), strings.TrimSpace(intro[at+1:])
	}
	out := gopodIntroWrap(who, intro, true)
	if line := gopodScriptedSnippet(esn, question, "start", ""); line != "" {
		logger.Println("BROBOT_SCRIPTED_SNIPPET=" + line)
		out = gopodStitchEndSnippet(line, out)
	}
	if take != "" {
		out = strings.TrimSpace(out) + gopodThoughtGap() + take
	}
	return strings.TrimSpace(out)
}

// gopodIntroFor: a Brobot's answer now is an intro (its chain is brobots_intro).
var gopodIntroFor sync.Map // esn -> bool

// gopodSaysNameAndNickname: the sentence says his card's name and nickname again - the opening said
// them; it goes, however worded.
func gopodSaysNameAndNickname(who, sentence string) bool {
	first := func(v string) string {
		if at := strings.Index(v, " - "); at > 0 {
			v = v[:at]
		}
		return gopodLooseWords(v)
	}
	name, nick := first(gopodPersonaProfileField(who, "My name")), first(gopodPersonaProfileField(who, "My informal name (nickname)"))
	s := gopodLooseWords(sentence)
	return name != "" && nick != "" && strings.Contains(s, name) && strings.Contains(" "+s+" ", " "+nick+" ")
}

// gopodIntroHeard: each persona's intro sentences said this session (content words), so one persona
// never repeats another's (gopodIntroWrap).
var gopodIntroHeard = struct {
	sync.Mutex
	by map[string][]map[string]bool
}{by: map[string][]map[string]bool{}}

func gopodIntroHeardBy(who string) []map[string]bool {
	if !gopodCommCheckRunning.Load() {
		return nil
	}
	gopodIntroHeard.Lock()
	defer gopodIntroHeard.Unlock()
	var out []map[string]bool
	for host, s := range gopodIntroHeard.by {
		if host != who {
			out = append(out, s...)
		}
	}
	return out
}

func gopodIntroHeardAdd(who string, said []map[string]bool) {
	if !gopodCommCheckRunning.Load() {
		return
	}
	gopodIntroHeard.Lock()
	gopodIntroHeard.by[who] = said
	gopodIntroHeard.Unlock()
}

// gopodIntroHeardReset: a new comm check, a clean slate.
func gopodIntroHeardReset() {
	gopodIntroHeard.Lock()
	gopodIntroHeard.by = map[string][]map[string]bool{}
	gopodIntroHeard.Unlock()
}

// gopodVoiceNarrationRE: "GOVERLORD's voice, deep and resonant:" - the model describing a voice
// instead of using it.
var gopodVoiceNarrationRE = regexp.MustCompile(`(?i)\b[\p{L} ]{0,30}'s voice[^:.!?]{0,60}:\s*`)

// gopodIntroGuideWithColour: the model says the guide itself, coloured in its own voice; when its
// words wander off the guide (under intro_base_share of the guide's words kept), the guide is said
// word for word instead.
func gopodIntroGuideWithColour(who string, coloured []string) string {
	guide := strings.TrimSpace(gopodIntroGuide(who))
	said := strings.TrimSpace(strings.Join(coloured, " "))
	if guide == "" {
		return said
	}
	base := gopodContentWords(guide)
	got := gopodContentWords(said)
	kept := 0
	for w := range base {
		if got[w] {
			kept++
		}
	}
	nick := strings.ToLower(strings.TrimSpace(gopodPersonaNickname(who, false)))
	if nick != "" && strings.Contains(strings.ToLower(guide), nick) && !strings.Contains(strings.ToLower(said), nick) {
		logger.Println("GOPOD_INTRO " + who + ": the colour doesn't name him - the guide, word for word")
		return guide
	}
	if len(base) == 0 || float64(kept)/float64(len(base)) < gopodIntroBaseShare {
		logger.Println(fmt.Sprintf("GOPOD_INTRO %s: the colour kept %d of %d of the guide's words - the guide, word for word", who, kept, len(base)))
		return guide
	}
	return said
}

// gopodIntroBaseShare: how much of the guide a coloured intro keeps, at least.
const gopodIntroBaseShare = 0.6

// ## EOF
