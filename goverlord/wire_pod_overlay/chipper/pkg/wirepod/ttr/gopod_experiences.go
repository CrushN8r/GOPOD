// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_experiences.go
package wirepod_ttr

import (
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
)

// gopodExperienceList: Bingo, then the settings' experiences, in order.
func gopodExperienceList() []*gopodExperience {
	list := []*gopodExperience{&gopodBrobotsSettings.Bingo}
	for i := range gopodBrobotsSettings.Experiences {
		list = append(list, &gopodBrobotsSettings.Experiences[i])
	}
	return list
}

var gopodExperiencePlayedFlags sync.Map // name -> *atomic.Bool

// gopodExperiencePlayed: the experience has been played since this Wire-Pod start (Bingo's is
// gopodBingoPlayed, gopod_brobot_personas.go).
func gopodExperiencePlayed(e *gopodExperience) *atomic.Bool {
	if e == &gopodBrobotsSettings.Bingo {
		return &gopodBingoPlayed
	}
	v, _ := gopodExperiencePlayedFlags.LoadOrStore(e.Name, new(atomic.Bool))
	return v.(*atomic.Bool)
}

// gopodExperienceInMemory: its run is in the shared chat memory (its snapshot note).
func gopodExperienceInMemory(e *gopodExperience) bool {
	if strings.TrimSpace(e.SnapshotNote) == "" {
		return false
	}
	for _, msg := range gopodSharedBrobotChat() {
		if strings.HasPrefix(msg.Content, e.SnapshotNote) {
			return true
		}
	}
	return false
}

// gopodExperienceRan: played since this start, or still in the chat memory.
func gopodExperienceRan(e *gopodExperience) bool {
	return gopodExperiencePlayed(e).Load() || gopodExperienceInMemory(e)
}

// gopodIsExperienceQuestion: the question is about it (question_word), not one of its
// not_a_test_phrases ("Chocolate Bingo" is the party game, not the test).
func gopodIsExperienceQuestion(e *gopodExperience, userMessage string) bool {
	word := strings.ToLower(strings.TrimSpace(e.QuestionWord))
	if word == "" {
		return false
	}
	lower := strings.ToLower(userMessage)
	for _, phrase := range e.NotATest {
		if gopodHasWholeWords(lower, strings.ToLower(phrase)) {
			return false
		}
	}
	return strings.Contains(lower, word)
}

// gopodIsBingoQuestion: the question is about the Bingo test.
func gopodIsBingoQuestion(userMessage string) bool {
	return gopodIsExperienceQuestion(&gopodBrobotsSettings.Bingo, userMessage)
}

// gopodExperienceMarkPlayed: a performed line holding an experience's run_marker means it has been
// played (gopod_chat_feed.go).
func gopodExperienceMarkPlayed(who, text string) {
	if who == "Operator" {
		return
	}
	for _, e := range gopodExperienceList() {
		if e.RunMarker != "" && strings.Contains(text, e.RunMarker) {
			gopodExperiencePlayed(e).Store(true)
		}
	}
}

// gopodExperienceClosing: the end of an answer about an experience - played, his closing remark
// word for word (his take); not played, the move to start it (the redirect).
func gopodExperienceClosing(esn, question string) string {
	if e := gopodExperienceAsked(question); e != nil {
		if gopodExperienceRan(e) {
			return strings.TrimSpace(e.Closing[gopodBrobotName(esn)])
		}
		if _, first := gopodAllFirstAnswer.Load(esn); !first && e.RunName != "" {
			return gopodExperienceMove(e.RunName)
		}
		return ""
	}
	return gopodTopicClosing(question)
}

// gopodExperienceMove: one of experience_moves, {run} filled.
func gopodExperienceMove(run string) string {
	moves := gopodBrobotsSettings.ExperienceMoves
	if len(moves) == 0 {
		return ""
	}
	return strings.TrimSpace(strings.ReplaceAll(moves[rand.Intn(len(moves))], "{run}", run))
}

// gopodExperienceAsked: the experience a question asks about having had - its word and an
// experience ask (brain_chain.experience_asks: "how was", "did you"...); "do you do bingo?" is the
// gig, not the run.
func gopodExperienceAsked(question string) *gopodExperience {
	q := " " + strings.Join(gopodLookupTokens(question), " ") + " "
	asked := false
	for _, w := range strings.Split(gopodBrobotsSettings.BrainChain.ExperienceAsks, ",") {
		if w = strings.Join(gopodLookupTokens(w), " "); w != "" && strings.Contains(q, " "+w+" ") {
			asked = true
			break
		}
	}
	if !asked {
		return nil
	}
	for _, e := range gopodExperienceList() {
		if gopodIsExperienceQuestion(e, question) {
			return e
		}
	}
	return nil
}

// gopodExperienceFacts: for a question about an experience, its one brain line - played or not
// (brain_chain.experience_yes / experience_no, {run} filled).
func gopodExperienceFacts(esn, question string) []string {
	e := gopodExperienceAsked(question)
	if e == nil {
		return nil
	}
	bc := gopodBrobotsSettings.BrainChain
	line := bc.ExperienceNo
	if gopodExperienceRan(e) {
		line = bc.ExperienceYes
	}
	run := e.RunName
	if run == "" {
		run = e.Name
	}
	if line = strings.TrimSpace(strings.ReplaceAll(line, "{run}", run)); line == "" {
		return nil
	}
	return []string{line}
}

// gopodExperiencePacks: the topic packs a question gets.
func gopodExperiencePacks(question string, packets []loadedScenarioPacket) []loadedScenarioPacket {
	if e := gopodExperienceAsked(question); e != nil {
		var out []loadedScenarioPacket
		if gopodExperienceRan(e) {
			for _, p := range packets {
				if p.Name == e.RunPack {
					out = append(out, p)
				}
			}
		}
		return out
	}
	var out []loadedScenarioPacket
	for _, p := range packets {
		if !gopodIsRunPack(p.Name) {
			out = append(out, p)
		}
	}
	return out
}

// gopodIsRunPack: a run's own pack (an experience's run_pack).
func gopodIsRunPack(pack string) bool {
	for _, e := range gopodExperienceList() {
		if e.RunPack != "" && pack == e.RunPack {
			return true
		}
	}
	return false
}

// ## EOF
