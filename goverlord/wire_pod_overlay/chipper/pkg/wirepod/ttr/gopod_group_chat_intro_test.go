package wirepod_ttr

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// LOCKED 2026-11-03 (operator: "company-persona-orientation-introduction-group-
// chat"; "Lock in a * single intro run * duo intro + weather run * brobot A:
// intro + weather * brobot B: intro + weather (follow up)"): the group chat
// intro is gopod_probes/topic_clusters/show/group-chat-intro.md and only that. The deploy
// fails when it changed without the operator's relock (GOPOD_RELOCK=1 saves its
// fingerprint, gopod_probes/locks/group_chat_intro.lock) or a settings file
// holds a copy.
func TestGroupChatIntroLock(t *testing.T) {
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("GOPOD_OVERLAY_CHIPPER not set (the deploy sets it)")
	}
	note, err := os.ReadFile(gopodTopicTreeFind("group-chat-intro.md"))
	if err != nil {
		t.Fatalf("LOCKED: the group chat intro is missing: %v", err)
	}
	sum := sha256.Sum256(note)
	got := hex.EncodeToString(sum[:])
	lock := filepath.Join(root, "gopod_probes", "locks", "group_chat_intro.lock")
	if os.Getenv("GOPOD_RELOCK") == "1" {
		if err := os.WriteFile(lock, []byte(got+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	} else if want, err := os.ReadFile(lock); err != nil {
		t.Errorf("LOCKED: the group chat intro lock is missing: %v", err)
	} else if strings.TrimSpace(string(want)) != got {
		t.Errorf("LOCKED: the group chat intro changed (operator's word only, then GOPOD_RELOCK=1)")
	}
	data, _ := gopodSettingsMerged()
	for _, copy := range []string{"\"cast_line\":", "\"nickname_part\":", "Introduce yourself. b) Name who"} {
		if strings.Contains(string(data), copy) {
			t.Errorf("LOCKED: a settings file holds a copy of the group chat intro: %s", copy)
		}
	}
	bc := gopodBrobotsSettings.BrainChain
	if !strings.Contains(bc.Flow["intro"], "Keys will be the first to speak") || !strings.Contains(bc.Flow["intro_first"], "You're the first to speak") || bc.IntroOpener != "Keys" {
		t.Errorf("LOCKED: the Guided Task Flow in use isn't the golden one: %q / %q / %q", bc.Flow["intro"], bc.Flow["intro_first"], bc.IntroOpener)
	}
}

// The four locked runs, through the real answer chain with fixed model words: a
// single intro; Brobot A's intro + weather (duo, first); Brobot B's (duo, next);
// one Brobot's intro + weather alone.
func TestIntroWeatherOrder(t *testing.T) {
	if os.Getenv("GOPOD_OVERLAY_CHIPPER") == "" {
		t.Skip("GOPOD_OVERLAY_CHIPPER not set (the deploy sets it)")
	}
	gopodCoverage.Lock()
	saved, savedRem := gopodCoverage.revealed, gopodCoverage.reminded
	gopodCoverage.revealed, gopodCoverage.reminded = map[string]int{}, map[string]bool{}
	gopodCoverage.Unlock()
	defer func() {
		gopodCoverage.Lock()
		gopodCoverage.revealed, gopodCoverage.reminded = saved, savedRem
		gopodCoverage.Unlock()
	}()
	gopodWeatherCache.Lock()
	savedWx := gopodWeatherCache.places
	gopodWeatherCache.places = map[string]gopodWeatherCached{strings.ToLower(strings.TrimSpace(gopodBrobotsSettings.LiveWeather.Location)): {
		facts: gopodWeatherFacts{Name: "Windsor", State: "Ontario", Country: "CA", Description: "clear sky", TempC: 20, Local: time.Date(2026, 11, 3, 15, 49, 0, 0, time.UTC)},
		at:    time.Now(),
	}}
	gopodWeatherCache.Unlock()
	defer func() { gopodWeatherCache.Lock(); gopodWeatherCache.places = savedWx; gopodWeatherCache.Unlock() }()
	raw := "happy___ I keep the party going with my brother.\nhappy___ Sweet sky out there, bro!"
	inOrder := func(name, got string, parts ...string) {
		at := 0
		for _, p := range parts {
			i := strings.Index(got[at:], p)
			if i < 0 {
				t.Errorf("%s: %q missing or out of order in:\n%s", name, p, got)
				return
			}
			at += i + len(p)
		}
	}
	run := func(esn, q string, first, next bool) string {
		gopodIntroFor.Store(esn, gopodIsIntro(q))
		gopodIntroWeatherFor.Store(esn, gopodIntroWeather(q))
		if first {
			gopodAllFirstAnswer.Store(esn, true)
		}
		if next {
			gopodFollowingUp.Store(esn, true)
		}
		defer func() {
			gopodIntroFor.Delete(esn)
			gopodIntroWeatherFor.Delete(esn)
			gopodAllFirstAnswer.Delete(esn)
			gopodFollowingUp.Delete(esn)
		}()
		return gopodShapeBrobotAnswerParts(esn, q, raw)
	}
	cast := "Here in the chat with us:"
	// A single intro: opening, his line, who's here - no weather.
	got := run(gopodBrobot1ESN, "Hey Doc, intro", false, false)
	inOrder("single intro", got, gopodIntroOpening("Brobot 1"), "I keep the party going", cast)
	if strings.Contains(got, "Weather Report") {
		t.Errorf("single intro has the weather: %s", got)
	}
	q := "Hey Brobots, intro & weather"
	// Brobot A: intro, weather report, his take, the hand-over.
	got = run(gopodBrobot2ESN, q, true, false)
	inOrder("Brobot A", got, gopodIntroOpening("Brobot 2"), "I keep the party going", cast, "Brobot Weather Report", "Sweet sky", "what say you on this")
	// Brobot B: intro, Confirmed., his take, the Weather Kit, the reminder.
	got = run(gopodBrobot1ESN, q, false, true)
	inOrder("Brobot B", got, gopodIntroOpening("Brobot 1"), "I keep the party going", cast, "Confirmed. Temperature", "Sweet sky", "Weather Kit", "that reminds me")
	// One Brobot alone: intro, weather report, his take, the reminder - no hand-over.
	got = run(gopodBrobot1ESN, "Hey Doc, intro & weather", false, false)
	inOrder("Brobot alone", got, gopodIntroOpening("Brobot 1"), "I keep the party going", cast, "Brobot Weather Report", "Sweet sky", "that reminds me")
	if strings.Contains(got, "what say you on this") {
		t.Errorf("one Brobot alone handed over: %s", got)
	}
}
