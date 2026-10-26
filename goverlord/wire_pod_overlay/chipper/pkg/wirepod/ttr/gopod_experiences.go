// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_experiences.go
package wirepod_ttr

// the duo answer-chain template (2026-10-25, operator:
// "Now, like pre-[bingo experience] > [bingo test] > post-[bingo experience] *
// duo answer-chain template * Make pre-[sleep experience] > [nap song] >
// post-[sleep experience]"; "nearly cookie-cutter copy/paste, tweaks").
//
// An experience is a scripted song the Brobots talk about (settings bingo, then
// experiences[] - the same keys). Before it has played, a question about it
// (question_word) gets each Brobot's <name>_not_played line (scripted_snippets)
// and no_run_note. Once a performed line holds its run_marker, the run is
// remembered as one snapshot note (gopodSharedBrobotChat), the asking Brobot's
// take goes with it (gopodAddExperienceTakes), the pitch pack stays out, and a
// question about it ends with his closing remark (gopodExperienceClosing).

import (
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/sashabaranov/go-openai"
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

// gopodExperienceClosing: the closing remark for a question about an
// experience - once it has played, his own (closing, by Brobot); then, in the
// second answer of a line to both (operator 2026-10-26: "'start Bingo test' in
// 2nd response regardless (A/B or B/A), before reminders"), before and after
// alike, experience_move with its run_name. "" for none.
func gopodExperienceClosing(esn, question string) string {
	for _, e := range gopodExperienceList() {
		if !gopodIsExperienceQuestion(e, question) {
			continue
		}
		var parts []string
		if gopodExperienceRan(e) {
			parts = append(parts, strings.TrimSpace(e.Closing[gopodBrobotName(esn)]))
		}
		if kind, n := gopodResponseSlot(esn); kind == "duo" && n == 1 && e.RunName != "" {
			parts = append(parts, gopodExperienceMove(e.RunName))
		}
		if line := strings.TrimSpace(strings.Join(parts, " ")); line != "" {
			return line
		}
	}
	return ""
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

// gopodExperienceFacts: for a question about an experience, before it has
// played its no_run_note and his pre_attitude, after it his post_attitude.
func gopodExperienceFacts(esn, question string) []string {
	var out []string
	for _, e := range gopodExperienceList() {
		if !gopodIsExperienceQuestion(e, question) {
			continue
		}
		name := gopodBrobotName(esn)
		note := ""
		if !gopodExperienceRan(e) {
			note = strings.TrimSpace(e.NoRunNote + " " + e.PreAttitude[name])
		} else {
			note = strings.TrimSpace(e.PostAttitude[name])
		}
		if note != "" {
			out = append(out, note)
		}
	}
	return out
}

// gopodExperienceNotPlayed: the question is about an experience not played yet,
// matching a scripted_snippets entry's when "<name>_not_played".
func gopodExperienceNotPlayed(when, userMessage string) bool {
	for _, e := range gopodExperienceList() {
		if e.Name != "" && when == e.Name+"_not_played" {
			return gopodIsExperienceQuestion(e, userMessage) && !gopodExperienceRan(e)
		}
		// 2026-10-26 (operator: "b2 framing, \"Oh! Zero!\""): <name>_played,
		// a question about it once it has played.
		if e.Name != "" && when == e.Name+"_played" {
			return gopodIsExperienceQuestion(e, userMessage) && gopodExperienceRan(e)
		}
	}
	return false
}

// gopodAddExperienceTakes adds this brother's take to each experience's
// snapshot note in the shared chat; played but no longer remembered, the
// snapshot and the take go first. Nothing changes for one not played.
func gopodAddExperienceTakes(esn string, chat []openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
	for _, e := range gopodExperienceList() {
		take, ok := e.Takes[gopodBrobotName(esn)]
		if !ok || e.SnapshotNote == "" {
			continue
		}
		added := false
		for i := len(chat) - 1; i >= 0; i-- {
			if strings.HasPrefix(chat[i].Content, e.SnapshotNote) {
				chat[i].Content += "\n[" + take + "]"
				added = true
				break
			}
		}
		if !added && gopodExperiencePlayed(e).Load() {
			note := openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: e.SnapshotNote + "\n[" + take + "]"}
			chat = append([]openai.ChatCompletionMessage{note}, chat...)
		}
	}
	return chat
}

// gopodExperiencePitchSkipped: a pack left out - a pitch pack once its run is
// in memory (the answer comes from the run), a run_pack until it has played.
func gopodExperiencePitchSkipped(pack string) bool {
	for _, e := range gopodExperienceList() {
		if e.PitchPack != "" && pack == e.PitchPack && gopodExperienceRan(e) {
			return true
		}
		if e.RunPack != "" && pack == e.RunPack && !gopodExperienceRan(e) {
			return true
		}
	}
	return false
}

// ## EOF
