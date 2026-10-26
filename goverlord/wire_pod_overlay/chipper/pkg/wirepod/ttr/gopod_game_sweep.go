// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_game_sweep.go
package wirepod_ttr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// The game sweep (2026-10-24, operator: "Test every: Topic, Value Point, Easter
// Egg * fullest run without live brobots"): every topic is found by its own
// words; every value point is picked by its own question (the interview's
// golden questions, interview_harvest/harvest_questions.json); every Easter egg
// on the list is reachable (its chapter's topics, or its topic). Each topic's
// questions are asked in order, the points already revealed set aside. Nothing is
// said, no model asked, the game's score untouched. system-test-game, and a
// test on every deploy (TestGameSweep).

type gopodSweepItem struct {
	Kind, Item string
	OK         bool
	Why        string
}

func gopodGameSweep() []gopodSweepItem {
	var out []gopodSweepItem
	loads := func(q, name string) bool {
		for _, p := range loadScenarioPacketsForMessage(q) {
			if p.Name == name {
				return true
			}
		}
		return false
	}
	// Topics: found by their own first word.
	topicOK := map[string]bool{}
	contents := map[string]string{}
	for _, p := range gopodTopicRegistry() {
		data, _ := gopodReadTopicFacts(p.File)
		contents[p.Name] = string(data)
		if len(p.Triggers) == 0 {
			topicOK[p.Name] = true // always loaded
			continue
		}
		q := "tell me about " + p.Triggers[0]
		ok := loads(q, p.Name)
		topicOK[p.Name] = ok
		why := ""
		if !ok {
			why = "\"" + q + "\" doesn't find it"
		}
		out = append(out, gopodSweepItem{"topic", p.Name, ok, why})
	}
	// Value points: each one's golden question through the real pipeline
	// (gopodQuestionFactsFor - the point a live question is given), each topic's
	// questions in order, the way the game plays them. The live score is put back
	// as it was after.
	var qfile struct {
		Questions []struct {
			Topic string `json:"topic"`
			To    string `json:"to"`
			Ask   string `json:"ask"`
		} `json:"questions"`
	}
	data, err := os.ReadFile(filepath.Join(gopodChipperRoot(), "gopod_probes/interview_harvest/harvest_questions.json"))
	if err != nil || json.Unmarshal(data, &qfile) != nil {
		out = append(out, gopodSweepItem{"value point", "harvest_questions.json", false, "can't read it"})
	}
	// As at a session's start: no experience played yet (it changes its pitch).
	for _, e := range gopodExperienceList() {
		flag := gopodExperiencePlayed(e)
		played := flag.Load()
		flag.Store(false)
		defer flag.Store(played)
	}
	gopodCoverage.Lock()
	savedRevealed, savedReminded, savedScored, savedSaid := gopodCoverage.revealed, gopodCoverage.reminded, gopodCoverage.scored, gopodCoverageSaid
	gopodCoverage.revealed, gopodCoverage.reminded, gopodCoverageSaid = map[string]int{}, map[string]bool{}, map[string]map[int]bool{}
	gopodCoverage.Unlock()
	defer func() {
		gopodCoverage.Lock()
		gopodCoverage.revealed, gopodCoverage.reminded, gopodCoverage.scored, gopodCoverageSaid = savedRevealed, savedReminded, savedScored, savedSaid
		gopodCoverage.Unlock()
	}()
	esnOf := map[string]string{"Brobot 1": gopodBrobot1ESN, "Brobot 2": gopodBrobot2ESN}
	nth := map[string]int{}
	for _, q := range qfile.Questions {
		nth[q.Topic]++
		n := nth[q.Topic]
		_, points := gopodSplitPoints(contents[q.Topic])
		count := len(points)
		if count == 0 {
			count = 1
		}
		if gopodBackgroundTopic(q.Topic) || !gopodCorePoint(q.Topic, n) {
			continue
		}
		// A Bingo run in the chat memory answers bingo from the run, not its pitch.
		if gopodExperiencePitchSkipped(q.Topic) {
			continue
		}
		esn := esnOf[q.To]
		if esn == "" {
			esn = gopodBrobot1ESN
		}
		want := fmt.Sprintf("%s %d/%d", q.Topic, n, count)
		gopodQuestionFactsFor(esn, gopodSpellNames(q.Ask))
		got := gopodGivenPoints(esn)
		item := fmt.Sprintf("%s #%d", q.Topic, n)
		if len(got) == 1 && got[0] == want {
			out = append(out, gopodSweepItem{"value point", item, true, ""})
		} else {
			out = append(out, gopodSweepItem{"value point", item, false, fmt.Sprintf("\"%s\" is given %v", q.Ask, got)})
		}
	}
	// Easter eggs: reachable.
	chapterTopics := map[string][]string{}
	for _, c := range gopodBrobotsSettings.InterviewChapters.Chapters {
		chapterTopics["chapter:"+strings.TrimSpace(c.Title)] = c.Topics
	}
	for _, e := range gopodEggList() {
		ok, why := true, ""
		if ts, isChapter := chapterTopics[e.Key]; isChapter {
			for _, t := range ts {
				if !topicOK[t] {
					ok, why = false, "its topic "+t+" isn't found"
				}
			}
		} else if found, known := topicOK[e.Key]; known && !found {
			ok, why = false, "its topic isn't found"
		}
		out = append(out, gopodSweepItem{"easter egg", e.Label, ok, why})
	}
	// The surprise eggs (topic_pictures.not_listed): off the list, still in the
	// score - counted the score's way (gopodCoverageScoreText).
	listed := map[string]bool{}
	for _, e := range gopodEggList() {
		listed[e.Key] = true
	}
	for _, n := range gopodBrobotsSettings.TopicPictures.NotListed {
		if listed[n] || gopodNotEgg(n) {
			continue
		}
		found, known := topicOK[n]
		ok, why := known && found, ""
		if !ok {
			why = "its topic isn't found"
		}
		out = append(out, gopodSweepItem{"easter egg", n + " (surprise)", ok, why})
	}
	return out
}

// gopodGameSweepLine: "Topics 31/31 · Value points 91/91 · Easter eggs 4/4", and
// the first misses.
func gopodGameSweepLine(items []gopodSweepItem) (bool, string) {
	total, good := map[string]int{}, map[string]int{}
	var misses []string
	for _, it := range items {
		total[it.Kind]++
		if it.OK {
			good[it.Kind]++
		} else if len(misses) < 6 {
			misses = append(misses, it.Kind+" "+it.Item+": "+it.Why)
		}
	}
	line := fmt.Sprintf("Topics %d/%d · Value points %d/%d · Easter eggs %d/%d", good["topic"], total["topic"], good["value point"], total["value point"], good["easter egg"], total["easter egg"])
	if len(misses) > 0 {
		line += " - " + strings.Join(misses, "; ")
	}
	return len(misses) == 0, line
}

// ## EOF
