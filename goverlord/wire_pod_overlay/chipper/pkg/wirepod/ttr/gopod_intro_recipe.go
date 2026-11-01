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
	"strings"
	"sync"
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
	for _, label := range []string{"My job title", "My role", "Who helps me"} {
		if v := strings.TrimSuffix(strings.TrimSpace(gopodPersonaProfileField(who, label)), "."); v != "" {
			facts = append(facts, v)
		}
	}
	return strings.TrimSpace(strings.NewReplacer("{opening}", gopodIntroOpening(who), "{facts}", strings.Join(facts, "; "),
		"{self}", strings.Join(facts, "; "), "{cast}", gopodIntroCast(who)).Replace(ir.Text))
}

// gopodIntroOpening: "I'm {name}{nickname_part}." from his profile - his name
// (formal), his nickname (informal) and why; each part left out when blank.
func gopodIntroOpening(who string) string {
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
		if why := strings.TrimSpace(gopodPersonaFill(who, head["why"])); why != "" {
			part += strings.ReplaceAll(ir.WhyPart, "{why}", why)
		}
	}
	return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(ir.Opening, "{name}", name), "{nickname_part}", part))
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
func gopodIntroWrap(who, middle string) string {
	opening := gopodIntroOpening(who)
	middle = strings.TrimSpace(gopodStripEcho(middle, opening))
	if opening != "" {
		middle = strings.TrimSpace(strings.Replace(middle, opening, "", 1))
	}
	names := strings.Split(gopodIntroCast(who), ", ")
	var keep []string
	for _, sentence := range gopodIntroSentences(middle) {
		n := 0
		for _, name := range names {
			if name != "" && strings.Contains(strings.ToLower(sentence), strings.ToLower(name)) {
				n++
			}
		}
		if n < 3 {
			keep = append(keep, sentence)
		}
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

// gopodIntroFor: a Brobot's answer now is an intro (its chain is brobots_intro).
var gopodIntroFor sync.Map // esn -> bool

// ## EOF
