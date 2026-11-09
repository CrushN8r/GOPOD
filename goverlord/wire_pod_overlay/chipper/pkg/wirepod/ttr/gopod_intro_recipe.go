// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_intro_recipe.go
package wirepod_ttr

// The intro recipe (2026-10-31, operator: "Please build the intro recipe in one
// batch"; his tester: "Hey [name]! You party animal! Introduce yourself fully to
// the chat and name who all else is here in the chat with us!!" - "Introduce
// yourself fully to the chat (my informal name (nickname), my name, my job title,
// my role, who helps me) * Name who all else is here in the chat with us (list
// other chat participants)"). One recipe for every chat participant: a line that
// asks for an intro (intro_recipe.words) gets intro_recipe.text with his
// profile's Self points ({self}) and everyone else here ({cast}). A Brobot's
// intro answer is the brobots_intro chain (no value point, no reminder); every
// intro, a Brobot's or a Gomad's, is wrapped the same way (gopodIntroWrap).

import (
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

// gopodIntroWeather (2026-11-03, operator: "Hey [name], intro & weather > duo
// intro run"): the line asks for an intro and the weather - his intro first,
// then his weather (group-chat-intro.md, then weather_run.md).
func gopodIntroWeather(text string) bool {
	return gopodIsIntro(text) && gopodHasWholeWords(strings.ToLower(text), "weather")
}

// gopodIntroWeatherFor: a Brobot's answer now is an intro + weather (brobots_intro_weather).
var gopodIntroWeatherFor sync.Map // esn -> bool

// gopodIntroRecipe: the recipe for who, "" when the line isn't an intro.
// 2026-10-31 (operator: "names vs nicknames = formal vs informal * I'm KEYNOTE
// VALE, my nickname is Keys because I hold all the Keys * [name][nickname]
// [nickname-history/why-that-nickname] * skipped if blank * > interns"; then
// "FIX IT CORRECT & PROPER"): an intro is a template - code says the opening
// (gopodIntroOpening) and the cast line (gopodIntroCastLine); the model writes
// only the middle: his job, role and who helps him ({facts}), in his own words.
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
	// 2026-11-04 (operator: "relationships aren't being talked about"): who he
	// answers to and who answers to him (his profile's own sections).
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
	// 2026-11-09 (operator: "KEYNOTE's introduction should establish the chain of
	// command and the core definition of GOPOD in a single, reusable outline that
	// every subsequent persona can assume is already known"): the one who opens
	// the session says the company's outline (## Our outline) as his facts.
	if o := strings.TrimSpace(gopodBrobotsSettings.BrainChain.IntroOpener); o != "" && strings.EqualFold(gopodCallName(who), o) {
		if outline := sections["Our outline"]; len(outline) > 0 {
			facts = nil
			for _, p := range outline {
				facts = append(facts, strings.TrimSuffix(strings.TrimSpace(gopodPersonaFill(who, p)), "."))
			}
		}
	}
	text2 := ir.Text
	// A Brobot's intro + weather: his take on the report after his one line.
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

// gopodIntroOpening: "I'm {name}{nickname_part}." from his profile - his name
// (formal), his nickname (informal) and why; each part left out when blank.
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
	// 2026-11-04 (Git Repo's why ends "... 'Git yer butt in gear!!'"): the
	// opening's own full stop is dropped when the why already ends its sentence,
	// a closing quote aside - "!!'" not "!!'.", "BRUH!" not "BRUH!.".
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

// gopodIntroWrap: the model's middle part with the code's opening before it and
// the cast line after it. A copy of the opening the model said anyway is dropped,
// and so is any sentence of his that lists the cast (three or more of its names) -
// the cast line says them once.
// 2026-11-02 (live: Pip's KP2 intro said only its opening; Doc's and Pip's All
// intros skipped their first thought): a Brobot's answer is cut into spoken
// chunks at its line breaks, one animation cue each - the opening joined to his
// answer with a space put two cues in one chunk, and everything after the second
// was never said. A Brobot's opening and cast line go in the way every code line
// does (gopodStitchSnippet / gopodStitchEndSnippet: after his cue, on its own
// line); a Gomad's (one voice, chunked by sentence) stay joined by spaces.
func gopodIntroWrap(who, middle string, brobot ...bool) string {
	opening := gopodIntroOpening(who)
	// 2026-11-05 (live 08:34: Doc's line came back wrapped in "*...*" and the
	// code's opening landed inside it - "*I'm Brobot 1, my nickname is Doc. *"):
	// an intro is a template, no stage directions - its asterisks go.
	middle = strings.ReplaceAll(middle, "*", "")
	middle = strings.TrimSpace(gopodStripEcho(middle, opening))
	if opening != "" {
		middle = strings.TrimSpace(strings.Replace(middle, opening, "", 1))
	}
	names := strings.Split(gopodIntroCast(who), ", ")
	selfNames := gopodIntroSelfNames(who)
	core := gopodIntroOpeningWith(who, false)
	var keep []string
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
		// 2026-11-02 (live: CHALK's opening twice, "-" then "–"; "I'm Cache PYC.
		// I'm Cache PYC, and I'm hustling"; "I'm Brobot 2, and hey!"; "I'm Doc,
		// Brobot 1, which means..."): his own "I'm <name>" again is dropped -
		// the whole sentence when it's the opening again (dashes, punctuation and
		// case aside), else just its "I'm <name>," start.
		if opening != "" && (gopodIsOpeningAgain(bare, opening, core) || gopodSaysNameAndNickname(who, bare)) {
			carry = cue
			continue
		}
		sentence = gopodDropSelfIntro(sentence, selfNames)
		if strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(sentence), cue)) == "" {
			carry = cue
			continue
		}
		// 2026-11-02 (live: Doc's "Here in the chat: all our brothers are rocking
		// it, including Pip" before the code's cast line): his own "here in the
		// chat" sentence goes too - the cast line says it.
		if strings.Contains(strings.ToLower(sentence), "in the chat") {
			carry = cue
			continue
		}
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
	if len(brobot) > 0 && brobot[0] {
		out := strings.Join(keep, " ")
		if opening != "" {
			out = gopodStitchSnippet(opening, out)
		}
		if cast := gopodIntroCastLine(who); cast != "" {
			out = gopodStitchEndSnippet(cast, out)
		}
		return strings.TrimSpace(out)
	}
	out := []string{opening, strings.Join(keep, " "), gopodIntroCastLine(who)}
	var parts []string
	for _, p := range out {
		if p = strings.TrimSpace(p); p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, " ")
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

// gopodIsOpeningAgain: the sentence is the opening said again - it starts with
// the opening's core ("I'm <name>, my nickname is <nick>"), and what follows is
// nothing or more of the opening's own words ("- because it's my name").
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
	// The longest first ("CHALK BORED" before "CHALK").
	sort.Slice(out, func(a, b int) bool { return len(out[a]) > len(out[b]) })
	return out
}

// gopodDropSelfIntro: the sentence without a starting "I'm <his name>[, <his
// name>]," - what follows starts the sentence ("and hey!" -> "And hey!").
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
			if len(rest) >= len(n) && strings.EqualFold(rest[:len(n)], n) && (len(rest) == len(n) || strings.ContainsRune(" ,.!;-", rune(rest[len(n)]))) {
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

// gopodIntroSentences: text cut after each . ! or ? that ends a word.
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
		// 2026-10-31: the operator is who's asking (brain_chain.asked), not one
		// more name in the list (live: "Brobot 0" and "Master CrushN8r" both named).
		if c.Who != who && c.Who != gopodBrobotsSettings.BrainChain.OperatorWho {
			cast = append(cast, gopodCallName(c.Who))
		}
	}
	return strings.Join(cast, ", ")
}

// gopodIntroThenWeather: a Brobot's intro + weather answer - his intro (the
// opening, his first line, the cast line), then the weather report or
// "Confirmed." (gopodScriptedSnippet), then his take (the rest of his words).
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

// gopodSaysNameAndNickname (2026-11-06, live: "My name is KEYNOTE VALE, my
// nickname is KEYS - because ..." after the code's opening): the sentence says his
// card's name and nickname again - the opening said them; it goes, however worded.
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

// ## EOF
