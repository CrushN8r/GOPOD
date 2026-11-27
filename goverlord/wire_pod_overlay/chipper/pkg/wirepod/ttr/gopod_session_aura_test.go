package wirepod_ttr

import (
	"strings"
	"testing"
)

// TestSessionAura: Goverlord's AURA line - from the card's words, filled by the code; the pace and
// tension pieces keep their leading space/comma.
func TestSessionAura(t *testing.T) {
	a := gopodBrobotsSettings.BrainChain.Aura
	if a["line"] == "" || !strings.HasPrefix(a["pace_lively"], ", ") || !strings.HasPrefix(a["tension_doc_pip"], " ") {
		t.Fatalf("brain_chain.aura from Goverlord's card: %q", a)
	}
	gopodBrainReset()
	defer gopodBrainReset()
	line := gopodSessionAura()
	if strings.Contains(line, "Visitors here") {
		t.Fatalf("no visitors, yet the aura names some: %q", line)
	}
	for _, want := range []string{a["first"], "Vibe: " + a["vibe_chat"]} {
		if !strings.Contains(line, want) {
			t.Fatalf("fresh session's aura %q lacks %q", line, want)
		}
	}
	if strings.Contains(line, "{") {
		t.Fatalf("aura left a slot unfilled: %q", line)
	}
	gopodBrainPerformed("Bingo Test Live", "done")
	if line = gopodSessionAura(); !strings.Contains(line, "Just played: ") || !strings.Contains(line, "Bingo Test Live") {
		t.Fatalf("after a run, the aura %q", line)
	}
	gopodBrainAdd(gopodCallName("CAMEO"), "who do I talk to about booking you?", "answer", "")
	if line = gopodSessionAura(); !strings.Contains(line, a["chat"]) || !strings.Contains(line, "book") {
		t.Fatalf("after a booking question, the aura %q", line)
	}
}

// TestNeverIn: a line naming a never word is caught on the way in; a clean one isn't.
func TestNeverIn(t *testing.T) {
	words := gopodBrobotsSettings.CleanPolitics.NeverWords
	if len(words) == 0 || gopodBrobotsSettings.CleanPolitics.NeverInLine == "" {
		t.Fatal("clean_politics.never_words and never_in_line are needed")
	}
	if !gopodNeverIn("tell me about "+words[0]+" today", "a test") {
		t.Fatal("a never word got in")
	}
	if gopodNeverIn("what's the weather in Windsor?", "a test") {
		t.Fatal("a clean line was dropped")
	}
}

// TestSpeechTidy: note separators and made-up links never get said; GOPOD's own links stay.
func TestSpeechTidy(t *testing.T) {
	got := gopodStripLeaks("A granny | she helped. See (https://www.google.com/url?q=x) or https://crushn8r.ca/subscribe today.")
	if strings.Contains(got, "|") || strings.Contains(got, "google") || !strings.Contains(got, "https://crushn8r.ca/subscribe") {
		t.Fatalf("tidied: %q", got)
	}
}

// TestCommCheckCrewLive: the crew bubble from the company folders - whatever they're called: every
// persona in one crew line, KEYS marked, the company named, no slot left unfilled.
func TestCommCheckCrewLive(t *testing.T) {
	keys := gopodBrobotsSettings.SystemBubble.VoicePersona
	got := gopodCommCheckCrewLines(keys)
	all := strings.Join(got, "\n")
	if strings.Contains(all, "{") || !strings.Contains(all, "(Myself)") || !strings.Contains(all, "Company: MATH AFTERMATH NEWS") {
		t.Fatalf("crew bubble:\n%s", all)
	}
	if !strings.Contains(all, "--- Admin Crew --- OPERATOR, GOVERLORD, KEYS (Myself)") {
		t.Fatalf("admin crew:\n%s", all)
	}
	for _, who := range []string{"DOC", "PIP", "Brobot 1", "Brobot 2", "CHALK", "GEMMA", "INTEL 404 PNFH"} {
		n := 0
		for _, l := range got {
			if strings.HasPrefix(l, "--- ") && strings.Contains(l, gopodPersonaNick(who)) {
				n++
			}
		}
		if n != 1 {
			t.Fatalf("%s is in %d crew lines:\n%s", who, n, all)
		}
	}
	t.Log("\n" + all)
}

// TestReminderLead: where the reminders' ride starts.
func TestReminderLead(t *testing.T) {
	labels := gopodBrobotsSettings.InterviewChapters.TopicLabels
	seat := func(topic string) bool { return labels[topic] != "" }
	gopodUnknownAsks.Store(0)
	defer gopodUnknownAsks.Store(0)
	if got := gopodReminderLead("what's the weather like in Windsor?", seat); got != "booking" && got != "bookings" {
		t.Fatalf("weather (a location) leads to %q, want booking", got)
	}
	pull := strings.Split(gopodBrobotsSettings.BrainChain.ReminderPull, ",")[0]
	if got := gopodReminderLead("zzqx blorp", seat); got != strings.TrimSpace(pull) {
		t.Fatalf("nothing to link on: %q, want the sales pull %q", got, pull)
	}
	gopodUnknownAsks.Store(5)
	unknown := strings.TrimSpace(strings.Split(gopodBrobotsSettings.BrainChain.UnknownPull, ",")[0])
	if got := gopodReminderLead("what's the weather like?", seat); got != unknown {
		t.Fatalf("after many I-don't-knows: %q, want %q", got, unknown)
	}
}

// TestNativeWeatherNeedsWeather: a weather take only for a weather line.
func TestNativeWeatherNeedsWeather(t *testing.T) {
	if gopodLineIsKind("there's number the water to go that's how you get the way that had to drop", "weather") {
		t.Fatal("stray words took the weather")
	}
	for _, q := range []string{"what's the weather like?", "what's the whether today", "give me the forecast"} {
		if !gopodLineIsKind(q, "weather") {
			t.Fatalf("%q is weather", q)
		}
	}
}
