// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_feed_dialogue_test.go
package wirepod_ttr

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestFeedDialogue (2026-10-17, operator: "one event list, single source of
// truth"; "the web page and terminal must have some sort of coordinated
// dialogue"): a live reply is a feed line (via "llm"), and the page's "played up
// to" moves forward only, for this boot.
func TestFeedDialogue(t *testing.T) {
	saved := gopodChatFeedPath
	defer func() { gopodChatFeedPath = saved }()
	gopodChatFeedPath = filepath.Join(t.TempDir(), "feed.json")
	gopodChatFeed.Lock()
	gopodChatFeed.next, gopodChatFeed.events = 0, nil
	boot := gopodChatFeed.boot
	gopodChatFeed.Unlock()
	gopodPagePlayedTo = 0
	GopodChatFeedAddLive(gopodBrobot1ESN, "Hello there.")
	gopodChatFeed.Lock()
	e := gopodChatFeed.events[len(gopodChatFeed.events)-1]
	gopodChatFeed.Unlock()
	if e.Via != gopodFeedLiveVia || e.Who != "Brobot 1" || e.Text != "Hello there." {
		t.Errorf("live line: %+v", e)
	}
	GopodPagePlayed(boot+1, 5) // another boot's page: ignored
	if gopodPagePlayedTo != 0 {
		t.Errorf("another boot moved it to %d", gopodPagePlayedTo)
	}
	GopodPagePlayed(boot, e.ID)
	GopodPagePlayed(boot, e.ID-1) // never back
	if gopodPagePlayedTo != e.ID {
		t.Errorf("played to %d, want %d", gopodPagePlayedTo, e.ID)
	}
}

// ## EOF

// TestSpokenCueWordKept (2026-10-17, operator: "celebrate skipped? not heard"):
// a piece that opens with an animation's name and just goes on keeps the word.
func TestSpokenCueWordKept(t *testing.T) {
	if got := stripSpokenActionLead("celebrate small wins,"); got != "celebrate small wins," {
		t.Errorf("got %q", got)
	}
	if got := stripSpokenActionLead("happy Hey there!"); got != "Hey there!" {
		t.Errorf("a cue before a new sentence: got %q", got)
	}
}

// TestBrobotLingoChoco (2026-10-17, operator: "brobot lingo: choco[text] =
// chalk[new text] * brobot lingo: chocolateric = chalk-o-latteric"; then "Oops.
// choco[text] = chalk-o-[new text]"; "Remove \"chocolate\" * chalk-o- is intended
// for brobot made up lingo"; "math = madth").
func TestBrobotLingoChoco(t *testing.T) {
	for in, want := range map[string]string{
		"The Chocolateric Tango?": "The chalk-o-latteric Tango?",
		"a chocoyum treat":        "a chalk-o-yum treat",
		"Chocolate Bingo":         "Chocolate Bingo", // "chalk-o- is intended for brobot made up lingo"
		"do the math":             "do the madth",
		"mathematics":             "mathematics",
	} {
		if got := gopodApplyPronunciations(in); got != want {
			t.Errorf("%q -> %q, want %q", in, got, want)
		}
	}
}

// TestWeatherCountry (2026-10-17, operator: "[City], [state/province], [Country]
// = Toronto, \"\", USA"): a country named in the place is split off, as its code.
func TestWeatherCountry(t *testing.T) {
	for in, want := range map[string][2]string{
		"Toronto USA":           {"Toronto", "US"},
		"Toronto":               {"Toronto", ""},
		"Paris, France":         {"Paris", "FR"},
		"Windsor United States": {"Windsor", "US"},
	} {
		city, code := gopodWeatherCountry(in)
		if city != want[0] || code != want[1] {
			t.Errorf("%q -> %q %q, want %q %q", in, city, code, want[0], want[1])
		}
	}
}

// TestEggList (2026-10-17, operator: "whatever easter eggs not discovered, are
// left on the list to mention"): a found egg leaves the list; the rest come in turn.
func TestEggList(t *testing.T) {
	list := gopodEggList()
	if len(list) < 3 {
		t.Fatalf("only %d eggs: %+v", len(list), list)
	}
	gopodInterviewShown.Lock()
	saved := gopodInterviewShown.pictures
	gopodInterviewShown.pictures = map[string]int{list[0].Key: 1}
	gopodInterviewShown.Unlock()
	gopodCoverage.Lock()
	savedR := gopodCoverage.reminded
	gopodCoverage.reminded = map[string]bool{}
	gopodCoverage.Unlock()
	defer func() {
		gopodInterviewShown.Lock()
		gopodInterviewShown.pictures = saved
		gopodInterviewShown.Unlock()
		gopodCoverage.Lock()
		gopodCoverage.reminded = savedR
		gopodCoverage.Unlock()
	}()
	seen := map[string]bool{}
	for i := 0; i < len(list); i++ {
		for _, e := range gopodEggsToMention(1, nil) {
			if e.Key == list[0].Key {
				t.Errorf("a found egg was named: %s", e.Label)
			}
			seen[e.Key] = true
		}
	}
	if len(seen) != len(list)-1 {
		t.Errorf("named %d of the %d eggs still to find", len(seen), len(list)-1)
	}
}

// TestAllStartsWith (2026-10-17, operator: "Starting with Brobot 2 ... this
// reverses"), and TestWaitNoteDropped ("--- (waiting for Brobot 0 before replying)").
func TestAllStartsWith(t *testing.T) {
	for in, want := range map[string]string{
		"This question is for the both of you. Starting with Brobot 2. Did you play bingo yet?": "Brobot 2",
		"Pip first, what's your favourite snack?":                                               "Brobot 2",
		"This question is for both of you. How was the test?":                                   "",
		"this question is for the both of you starting with bro bought to what's the weather":   "Brobot 2",
		"starting with brobot one, hi":                                                          "Brobot 1",
	} {
		if got := gopodAllStartsWith(in); got != want {
			t.Errorf("%q -> %q, want %q", in, got, want)
		}
	}
	if gopodAllFirst("starting with brobot 2, hi") != gopodBrobot2ESN || gopodAllFirst("hi both") != gopodBrobot1ESN {
		t.Error("first")
	}
}

func TestWaitNoteDropped(t *testing.T) {
	for _, in := range []string{
		"They obviously can't handle a little complexity. 🍫🤖 --- (Wait for Brobot 0 before replying)",
		"They obviously can't handle a little complexity. 🍫🤖 (waiting for Brobot 0 before replying)",
	} {
		if got := gopodPrepareReply(in); strings.Contains(strings.ToLower(got), "wait") || strings.Contains(got, "---") {
			t.Errorf("%q -> %q", in, got)
		}
	}
}

// 2026-10-17: the year said in halves; a flag said as its country.
func TestSpeakYearsAndFlags(t *testing.T) {
	saved := gopodBrobotsSettings.CleanPolitics.Countries
	defer func() { gopodBrobotsSettings.CleanPolitics.Countries = saved }()
	gopodBrobotsSettings.CleanPolitics.Countries = map[string]string{"canada": "CA", "united states": "US", "usa": "US"}
	if got := gopodSpeakYears("October 17, 2026, 11"); got != "October 17, 20 26, 11" {
		t.Fatalf("year: %q", got)
	}
	if got := gopodSpeakFlags("Windsor Ontario 🇨🇦, October"); got != "Windsor Ontario Canada, October" {
		t.Fatalf("flag: %q", got)
	}
	if got := gopodSpeakFlags("Toronto, Canada 🇨🇦 rocks"); got != "Toronto, Canada  rocks" {
		t.Fatalf("named flag: %q", got)
	}
	if got := gopodSpeakFlags("Ohio 🇺🇸"); got != "Ohio United States" {
		t.Fatalf("us flag: %q", got)
	}
}

// 2026-10-17 (operator: "easter eggs are live run chapters"): every chapter is
// an egg, titled by it; the roast battle (in no chapter) stays its own.
func TestChapterEggs(t *testing.T) {
	list := gopodEggList()
	chapters := 0
	roast := false
	photo := false
	for _, e := range list {
		if strings.HasPrefix(e.Key, "chapter:") {
			chapters++
		}
		if e.Key == "roast_battle" {
			roast = true // a surprise egg: never listed
		}
		// 2026-10-24 (operator: "isn't the photo thing an easter egg also? * if not,
		// make it an easter egg please"): the photo is an egg again.
		if e.Key == "photo" {
			photo = true
		}
		if e.Key == "weather" || e.Key == "halloween" || gopodNotEgg(e.Key) {
			t.Errorf("%s is no longer an egg", e.Key)
		}
	}
	if !photo {
		t.Error("the photo is an Easter egg (Get Photo)")
	}
	// 2026-10-22 (operator: "fewer easter eggs"): with eggs_keep, its chapters only.
	wantChapters := len(gopodBrobotsSettings.InterviewChapters.Chapters) - len(gopodBrobotsSettings.TopicPictures.NotEggs)
	if keep := gopodBrobotsSettings.TopicPictures.EggsKeep; len(keep) > 0 {
		wantChapters = 0
		for _, k := range keep {
			if strings.HasPrefix(k, "chapter:") {
				wantChapters++
			}
		}
	}
	if chapters != wantChapters || roast {
		t.Fatalf("chapters %d of %d, roast %v: %+v", chapters, len(gopodBrobotsSettings.InterviewChapters.Chapters), roast, list)
	}
	if key, title := gopodTopicChapter("birthday_and_merch"); key != "chapter:Booking" || title != "Booking" {
		t.Errorf("birthday chapter: %q %q", key, title)
	}
	t.Logf("EGGS %+v", list)
}

// 2026-10-18 (step 4): every file the settings, chapters and songs name is there.
func TestSystemCheck(t *testing.T) {
	problems, checked := gopodSystemCheckProblems()
	for _, p := range problems {
		t.Log(p)
	}
	t.Logf("SYSTEM_CHECK %d checked, %d to fix", checked, len(problems))
	if gopodCheckFind("no_such_picture_zz.png") || !gopodCheckFind("dJent3b.wav") || !gopodCheckFind("/weather/sounds/weather_rain.mp3") {
		t.Error("the file finder is wrong")
	}
}

// 2026-10-18: KP7/KP8's voice-only personas talk through their hosts, with their
// identity (the operator's words) and their reporter's lines in the prompt.
func TestPersonas(t *testing.T) {
	if gopodPersonaFor("chalk") != "CHALK" || gopodPersonaFor("PLAYHEAD") != "PLAYHEAD" || gopodPersonaFor("Doc") != "" {
		t.Fatal("persona hosts")
	}
	p := gopodPersonaPrompt("CHALK")
	if !strings.Contains(p, "I'm CHALK BORED, display manager at the CRUSHN8R CREW'd news station") {
		t.Errorf("CHALK's identity: %q", p)
	}
	if !strings.Contains(gopodPersonaPrompt("PLAYHEAD"), "PLAYHEAD ZUMINTU, session manager") {
		t.Error("PLAYHEAD's identity")
	}
	// PLAYHEAD's location QR: a web address in the line.
	if got := gopodManagerURL("Brobot 1, show me flare4.com please"); got != "https://flare4.com" {
		t.Errorf("url: %q", got)
	}
	if got := gopodManagerURL("what's the weather like?"); got != "" {
		t.Errorf("no url: %q", got)
	}
}
