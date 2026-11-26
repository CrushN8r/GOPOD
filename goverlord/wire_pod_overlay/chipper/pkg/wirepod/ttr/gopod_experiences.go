// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_experiences.go
package wirepod_ttr

// the duo answer-chain template (2026-10-25, operator:
// "Now, like pre-[bingo experience] > [bingo test] > post-[bingo experience] *
// duo answer-chain template * Make pre-[sleep experience] > [nap song] >
// post-[sleep experience]"; "nearly cookie-cutter copy/paste, tweaks").
//
// An experience is a scripted song the Brobots talk about (each run card's
// experience in settings game.runs, filled into bingo and experiences at load -
// gopod_game_engine.go). 2026-11-24 (operator: "SIMPLIFY!! NO PATCH SCAR TISSUE
// ... Q: did you have [this] experience? y/n * if no, say so and move on or
// redirect * if yes, give your take"): a question about it (question_word) gets
// one brain line - played or not (brain_chain.experience_yes / experience_no) -
// and no topic points with it. Played: his closing remark, word for word, is his
// take. Not played: the move to start it (experience_moves) is the redirect.
// Once a performed line holds its run_marker, the run is remembered as one
// snapshot note (gopodSharedBrobotChat).

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

// gopodExperiencePlayed: the experience has been played since this Wire-Pod
// start (Bingo's is gopodBingoPlayed, gopod_brobot_personas.go).
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

// gopodIsExperienceQuestion: the question is about it (question_word), not one
// of its not_a_test_phrases ("Chocolate Bingo" is the party game, not the test).
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

// gopodExperienceMarkPlayed: a performed line (not the operator's) holding an
// experience's run_marker means it has been played (gopod_chat_feed.go).
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

// gopodExperienceClosing: the end of an answer about an experience - played, his
// closing remark word for word (his take); not played, the move to start it (the
// redirect). Otherwise a covered topic's own closing remark (gopod_topic_asks.go).
func gopodExperienceClosing(esn, question string) string {
	if e := gopodExperienceAsked(question); e != nil {
		if gopodExperienceRan(e) {
			return strings.TrimSpace(e.Closing[gopodBrobotName(esn)])
		}
		// A line to both: only the second answer redirects (one "start ..." is enough).
		if _, first := gopodAllFirstAnswer.Load(esn); !first && e.RunName != "" {
			return gopodExperienceMove(e.RunName)
		}
		return ""
	}
	return gopodTopicClosing(question)
}

// gopodExperienceMove (2026-10-26, operator: "recover & add a few more to a
// list of 5, to randomly choose from"): one of experience_moves, {run} filled.
func gopodExperienceMove(run string) string {
	moves := gopodBrobotsSettings.ExperienceMoves
	if len(moves) == 0 {
		return ""
	}
	return strings.TrimSpace(strings.ReplaceAll(moves[rand.Intn(len(moves))], "{run}", run))
}

// gopodExperienceAsked: the experience a question asks about having had - its
// word and an experience ask (brain_chain.experience_asks: "how was", "did you"...);
// "do you do bingo?" is the gig, not the run. nil for none.
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

// gopodExperienceFacts: for a question about an experience, its one brain line -
// played or not (brain_chain.experience_yes / experience_no, {run} filled).
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

// gopodExperiencePacks (2026-11-24, operator: "Q: did you have [this] experience?
// y/n"): the topic packs a question gets. A run's own pack (run_pack) goes only
// with an experience ask once the run has played - his take has the run's facts;
// not played, none. Every other question keeps its packs but the run packs -
// "bingo for seniors" is the gig, run or no run.
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
