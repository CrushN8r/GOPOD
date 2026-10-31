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
// intro answer is the brobots_intro chain (no value point, no reminder); a
// Gomad's brain bubble says intro_recipe.task.

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
func gopodIntroRecipe(who, text string) string {
	ir := gopodBrobotsSettings.IntroRecipe
	if strings.TrimSpace(ir.Text) == "" || !gopodIsIntro(text) {
		return ""
	}
	self := gopodPersonaFill(who, strings.Join(gopodPersonaProfileSections(who)["Self"], " "))
	return strings.TrimSpace(strings.NewReplacer("{self}", self, "{cast}", gopodIntroCast(who)).Replace(ir.Text))
}

// gopodIntroCast: everyone else in the chat, by the names they're called.
func gopodIntroCast(who string) string {
	var cast []string
	for _, c := range gopodBrobotsSettings.HereNow.Cast {
		if c.Who != who {
			cast = append(cast, gopodCallName(c.Who))
		}
	}
	return strings.Join(cast, ", ")
}

// gopodIntroFor: a Brobot's answer now is an intro (its chain is brobots_intro).
var gopodIntroFor sync.Map // esn -> bool

// ## EOF
