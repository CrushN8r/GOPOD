package wirepod_ttr

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The settings file loads, and the values reach the code and the prompt from it.
func TestBrobotsSettingsReachTheCode(t *testing.T) {
	s := gopodBrobotsSettings
	if gopodAnimationMarker != s.AnimationMarker || gopodSharedBrobotChatExchanges != s.SharedChat.MaxQuestions || gopodBingoSnapshotNote != s.Bingo.SnapshotNote {
		t.Fatal("a setting did not reach the code")
	}
	if gopodBrobotPersonaLines["Brobot 1"] != s.Personas["Brobot 1"]+" "+gopodHostPersonaText("Brobot 1", s.Personas["Brobot 1"]) || !strings.Contains(gopodBrobotPersonaLines["Brobot 2"], "I go by Pip") || gopodBrobotBingoTakes["Brobot 2"] != s.Bingo.Takes["Brobot 2"] {
		t.Fatal("personas or Bingo takes did not reach the code")
	}
	got := gopodFillPromptTokens("$PLACEHOLDER_WORD$ANIMATION_MARKER spoken thought, marker: $ANIMATION_MARKER")
	if strings.Contains(got, "$") || !strings.HasPrefix(got, s.PlaceholderWord+s.AnimationMarker+" spoken thought") {
		t.Fatalf("prompt tokens not filled: %q", got)
	}
	if gopodFinalOutputGuard != strings.ReplaceAll(s.FinalOutputCheck, "$ANIMATION_MARKER", s.AnimationMarker) || strings.Contains(gopodFinalOutputGuard, s.PlaceholderWord) {
		t.Fatalf("final output check is not final_output_check, or shows the placeholder template: %q", gopodFinalOutputGuard)
	}
	for _, brobot := range []string{"Brobot 1", "Brobot 2"} {
		if shape := gopodReplyShapeSection(brobot); !strings.Contains(shape, s.ReplyExamples[brobot]) || strings.Contains(shape, "$") {
			t.Fatalf("%s's reply shape is missing his examples: %q", brobot, shape)
		}
	}
}

// ## EOF
// Scripted snippets (scripted_snippets in gopod_brobots_settings.json): the
// exact line is chosen by the question's topic pack and stitched in after the
// reply's cue word, so the cue still plays first.
func TestScriptedSnippetChoiceAndStitch(t *testing.T) {
	gopodBingoPlayed.Store(false)
	defer gopodBingoPlayed.Store(false)
	// No Bingo run in the chat memory either (as TestBingoPitchPackAndNoRunNote).
	savedStart := gopodSharedBrobotChatStart
	defer func() { gopodSharedBrobotChatStart = savedStart }()
	gopodSharedBrobotChatStart = time.Now().Add(-time.Hour)
	writeBrobotChatSession(t, nil)
	gopodChatFeed.Lock()
	savedEvents := gopodChatFeed.events
	gopodChatFeed.events = nil
	gopodChatFeed.Unlock()
	defer func() { gopodChatFeed.Lock(); gopodChatFeed.events = savedEvents; gopodChatFeed.Unlock() }()
	for q, want := range map[string]string{
		"are you on TikTok":         "No - I'm on GOPOD!",
		"can I book you":            "Join the CRUSHN8R CREW'd newsletter at crushn8r.ca!",
		"do you want to play bingo": "Sure, let's play! Say start Bingo test!",
		"tell me all about GOPOD":   "",
	} {
		if got := gopodScriptedSnippet("0dd1b9e9", q, "start", ""); got != want {
			t.Errorf("snippet for %q = %q, want %q", q, got, want)
		}
	}
	if got := gopodScriptedSnippet("0dd1d8bf", "what's your favourite food", "end", "happy___ Pizza and beer, bro!"); got != "Chocolate Bingo is fun for kids and adult kids too!" {
		t.Errorf("end line for a pizza reply = %q", got)
	}
	if got := gopodScriptedSnippet("0dd1d8bf", "what's your favourite food", "end", "happy___ Pizza, bro!"); got != "Chocolate Bingo is fun for kids of all ages!" {
		t.Errorf("end line for a pizza-only reply = %q", got)
	}
	if got := gopodStitchEndSnippet("Bingo is fun for kids and adult kids too!", "happy___ Pizza, bro!"); got != "happy___ Pizza, bro!"+gopodThoughtGap()+"Bingo is fun for kids and adult kids too!" {
		t.Errorf("end stitch = %q", got)
	}
	if got := gopodScriptedSnippet("0dd1b9e9", "thanks brobots", "end", "celebrate___ Party on!"); got != "Remember to GOPOD Yourself!" && got != "Join the CRUSHN8R CREW'd newsletter at crushn8r.ca!" {
		t.Errorf("thank-you end line = %q", got)
	}
	gopodBingoPlayed.Store(true)
	if got := gopodScriptedSnippet("0dd1b9e9", "do you want to play bingo", "start", ""); got != "" {
		t.Errorf("bingo snippet after a played run = %q, want none", got)
	}
	for reply, want := range map[string]string{
		"confused___ Ugh, TikTok.": "confused___ No - I'm on GOPOD! Ugh, TikTok.",
		"happy Ugh, TikTok.":       "happy___ No - I'm on GOPOD! Ugh, TikTok.",
		"Ugh, TikTok.":             "No - I'm on GOPOD! Ugh, TikTok.",
		"happy___ No, I'm on GOPOD! no - i'm on gopod Ugh.": "happy___ No - I'm on GOPOD! No, I'm on GOPOD! Ugh.",
		"thinking__ Ugh, TikTok.":                           "thinking___ No - I'm on GOPOD! Ugh, TikTok.",
	} {
		// 2026-10-15: the snippet is its own thought - its own line (thought_lines).
		want = strings.Replace(want, "GOPOD! ", "GOPOD!"+gopodThoughtGap(), 1)
		if got := gopodStitchSnippet("No - I'm on GOPOD!", reply); got != want {
			t.Errorf("stitch %q = %q, want %q", reply, got, want)
		}
	}
	if got := gopodStitchSnippet("", "happy___ hi"); got != "happy___ hi" {
		t.Errorf("no snippet changed the reply: %q", got)
	}
}

// Merged words are spoken split (merged_words, gopod_pronunciation.go).
func TestSplitMergedWords(t *testing.T) {
	if gopodDictionaryWords == nil {
		t.Skip("no dictionary on this machine")
	}
	for in, want := range map[string]string{
		"what a snoozefest, bro":  "what a snooze-fest, bro",
		"Snoozefest!":             "Snooze-fest!",
		"a wonderful bromance":    "a wonderful bromance",
		"the Brobots and GOPOD":   "the Brobots and GOPOD",
		"Brotobot vibes":          "Brotobot vibes",
		"dartingEyes partytime":   "dartingEyes party-time",
		"kickstart rockstars":     "kickstart rockstars",
		"gigglefest storytelling": "giggle-fest storytelling",
		"robotically somethin":    "robotically somethin",
	} {
		if got := gopodSplitMergedWords(in); got != want {
			t.Errorf("%q -> %q, want %q", in, got, want)
		}
	}
}

// A stuck wait gives up after chat_hold.stuck_seconds (gopod_stuck_reply.go).
func TestStuckReplyGivesUp(t *testing.T) {
	saved := gopodBrobotsSettings.ChatHold.StuckSeconds
	defer func() { gopodBrobotsSettings.ChatHold.StuckSeconds = saved }()
	gopodBrobotsSettings.ChatHold.StuckSeconds = 0.05
	start := time.Now()
	gopodSendOrGiveUp(make(chan bool), "0dd1d8bf", "test")
	gopodReceiveOrGiveUp(make(chan bool), "0dd1d8bf", "test")
	gopodWaitOrGiveUp(make(chan struct{}), "0dd1d8bf", "test")
	if took := time.Since(start); took > time.Second {
		t.Fatalf("stuck waits took %v", took)
	}
}

// GOLDEN highlight (gopod_highlight_pieces.go): for real-shaped replies, the
// bubble is cut once, here, and every spoken piece names a bubble piece - in
// order, from the first to the last, none skipped - and that piece holds most
// of the piece's shown words. Live replies from 2026-09-29 to 2026-10-06 that
// lit the wrong piece are in this list.
func TestHighlightPiecesGolden(t *testing.T) {
	marker := gopodBrobotsSettings.ScriptedSplit.SegmentMarker
	// 2026-10-08: the same rules with short pieces joined (live_split.join_up_to_words).
	savedJoin := gopodBrobotsSettings.LiveSplit.JoinUpToWords
	defer func() { gopodBrobotsSettings.LiveSplit.JoinUpToWords = savedJoin }()
	for _, join := range []int{0, 14} {
		gopodBrobotsSettings.LiveSplit.JoinUpToWords = join
		for _, raw := range []string{
			"confused___ I'm hearing you loud and clear, bro! *blink blink* *whispers* Sure, I hear ya, bro. 🙄🎤",
			"thinking___ Hey now, GOPOD is kinda like a teacher from outer space. You gotta be smartin' to get inside, but if you’re keen on learning the dance of AI, GOPOD is the way to go! robot gives thumbs-up It teaches ya words.",
			"dartingEyes___ We've got disco balls - a big ball, a small ball... and lights—so many lights, bro! 😎 You know what's even wilder? GOPOD runs on a local brain, no cloud.",
			"happy___ Bro-man? Napped on the job? Classic! What a snoozefest, dude. In each sentence 😉 You know we keep it fresh, right?",
			"celebrate___ No - I'm on GOPOD! Ugh, TikTok. What's the point, bro... 🙄 Join the CRUSHN8R CREW'd newsletter at crushn8r.ca!",
			"veryHappy___ Morning: forty people playing bingo at a seniors' residence, no wifi - doesn't matter! Afternoon: a classroom full of kids, shouting spelling words. Evening: a bar top, under stage lights. 🎉\n\nLet's gooo!",
			"thinking___ Short. Very short. Tiny bits, bro.",
			// live 2026-10-06 14:43: the curly apostrophe in "It’s".
			"thinking___ Now we're in HQ, the big stage awaits! 🚀 It\u2019s all about party bro beats - how your audience connects through our own wild robot vibes. 🤘 Whatcha say to that?!",
		} {
			rawInternal := gopodAnimationMarkerToInternal(gopodFixSelfName("0dd1b9e9", gopodCorrectNames(gopodStripLeadingAnimationWords(raw))))
			display := gopodDisplayReply(raw)
			pieces := gopodSpokenPieceTexts(buildGOPODStreamResult(rawInternal).AudibleChunks)
			marked, segOf := gopodSegmentDisplay(display, pieces)
			shown := strings.Split(marked, marker)
			for _, piece := range shown {
				if join == 0 && strings.Contains(piece, "all about party") && !strings.HasPrefix(strings.TrimSpace(piece), "It\u2019s") {
					t.Errorf("the piece with \"It\u2019s\" starts late: %q", piece)
				}
			}
			if len(pieces) == 0 {
				t.Fatalf("%q: no pieces", raw)
			}
			if strings.Join(strings.Fields(strings.ReplaceAll(marked, marker, "")), " ") != strings.Join(strings.Fields(display), " ") {
				t.Errorf("%q: the bubble's words changed: %q", raw, marked)
			}
			if segOf[0] != 0 || segOf[len(segOf)-1] != len(shown)-1 {
				t.Errorf("%q: pieces %v don't run from the first bubble piece to the last (%d)\nspoken %q\nshown %q", raw, segOf, len(shown), pieces, shown)
			}
			for i := 1; i < len(segOf); i++ {
				if segOf[i] < segOf[i-1] || segOf[i] > segOf[i-1]+1 {
					t.Errorf("%q: piece numbers %v go back or skip", raw, segOf)
				}
			}
			for i, p := range pieces {
				words := gopodHighlightWords(p)
				in := 0
				for _, w := range words {
					if strings.Contains(strings.ToLower(shown[segOf[i]]), w.word) {
						in++
					}
				}
				if len(words) >= 3 && in*2 < len(words) {
					t.Errorf("%q: spoken %q lights %q", raw, p, shown[segOf[i]])
				}
			}
		}
	}
}

// No made-up web addresses; crushn8r addresses keep only the bare name
// (web_links); "Chocolate Bingo" is not the Bingo test (bingo.not_a_test_phrases).
func TestWebLinksAndChocolateBingo(t *testing.T) {
	for in, want := range map[string]string{
		"Check us out at https://gc.crrshn8r.com/ bro!":        "Check us out at bro!",
		"Crushn8r on Tiktok: https://www.tiktok.com/@crushn8r": "Crushn8r on Tiktok:",
		"Join us at https://www.crushn8r.ca/news!":             "Join us at crushn8r.ca!",
		"foodmath.crushn8r.net has the recipes.":               "foodmath.crushn8r.net has the recipes.",
		"Visit www.partybots.com, it's wild.":                  "Visit , it's wild.",
		"Wait... I-9? That's 9 bro.":                           "Wait... I-9? That's 9 bro.",
	} {
		if got := strings.TrimSpace(gopodDropWebLinks(in)); got != want {
			t.Errorf("%q -> %q, want %q", in, got, want)
		}
	}
	if gopodIsBingoQuestion("what about the foodmath website and Chocolate Bingo?") {
		t.Error("Chocolate Bingo counted as the Bingo test")
	}
	if !gopodIsBingoQuestion("do you want to play bingo") {
		t.Error("a Bingo question not seen")
	}
}

// Stretched words are said stretched (stretched_words), and "Bro..." is a piece
// of its own, said with a pause (operator 2026-10-06).
func TestStretchedWordsAndBro(t *testing.T) {
	for in, want := range map[string]string{
		"brooooo, let's gooooo!": "bro-o-o, let's go-o-o!",
		"Brooo":                  "Bro-o-o",
		"too good":               "too good",
		"zzz yesss":              "zzz yesss",
	} {
		if got := gopodSayStretchedWords(in); got != want {
			t.Errorf("%q -> %q, want %q", in, got, want)
		}
	}
	for in, want := range map[string][]string{
		"Bro... that's wild.": {"Bro...", "that's wild."},
		"Bro...":              {"Bro..."},
		"Bro... We go!":       {"Bro...", "We go!"},
	} {
		if got := gopodSentences(in); strings.Join(got, "|") != strings.Join(want, "|") {
			t.Errorf("%q -> %q, want %q", in, got, want)
		}
	}
}

// 2026-10-06: the interview's chapters, "reminds me" lines and topic pictures
// all point at real topic packs and files; chapter titles, pictures, reporter
// voice files and host lines stay out of the Brobots' memory.
func TestInterviewChaptersPicturesAndReporter(t *testing.T) {
	packs := map[string]bool{}
	for _, p := range gopodTopicRegistry() {
		if packs[p.Name] {
			t.Errorf("topic %s is in the chapters folder twice", p.Name)
		}
		packs[p.Name] = true
		if facts, err := gopodReadTopicFacts(p.File); err != nil || len(strings.TrimSpace(string(facts))) == 0 {
			t.Errorf("topic %s: no facts (%v)", p.Name, err)
		}
		if len(p.Triggers) == 0 {
			t.Errorf("topic %s: no words", p.Name)
		}
	}
	if len(packs) == 0 {
		t.Fatal("no topics in the chapters folder")
	}
	inChapter := map[string]bool{}
	for _, c := range gopodBrobotsSettings.InterviewChapters.Chapters {
		for _, topic := range c.Topics {
			if !packs[topic] {
				t.Errorf("chapter %q: no pack %q", c.Title, topic)
			}
			if inChapter[topic] {
				t.Errorf("topic %q is in two chapters", topic)
			}
			inChapter[topic] = true
		}
	}
	for name := range packs {
		// 2026-10-16: an egg-only topic (a card egg, the roast battle) sits outside
		// the interview's chapters, like what_else.
		_, eggOnly := gopodBrobotsSettings.TopicPictures.Cards[name]
		for _, n := range gopodBrobotsSettings.TopicPictures.NotListed {
			eggOnly = eggOnly || n == name // 2026-10-20: a surprise egg (camera feeds)
		}
		if !inChapter[name] && name != gopodBrobotsSettings.ValuePoints.WhatElseTopic && !eggOnly {
			t.Errorf("pack %q is in no chapter", name)
		}
	}
	for _, s := range gopodBrobotsSettings.ScriptedSnippets.Lines {
		special := map[string]bool{"any": true, "thanks": true, "bingo_not_played": true, "reply_mentions": true, "reminder": true}
		if !special[s.When] && !packs[s.When] {
			t.Errorf("scripted snippet for unknown pack %q", s.When)
		}
	}
	for topic := range gopodBrobotsSettings.ValuePoints.Reminders {
		if !packs[topic] {
			t.Errorf("reminder for unknown pack %q", topic)
		}
	}
	for topic, files := range gopodBrobotsSettings.TopicPictures.Pictures {
		if !packs[topic] {
			t.Errorf("pictures for unknown pack %q", topic)
		}
		for _, f := range files {
			if strings.HasPrefix(f, "/") {
				continue // 2026-10-20: a live camera stream, not a file
			}
			if _, err := os.Stat(filepath.Join(gopodChipperRoot(), "webroot", gopodBrobotsSettings.TopicPictures.Folder, f)); err != nil {
				t.Errorf("picture %s: %v", f, err)
			}
		}
	}
	if got := gopodChapterOf("doc_and_pip"); got != "Warm-up" {
		t.Errorf("doc_and_pip chapter = %q", got)
	}
	rv := gopodBrobotsSettings.ReporterVoices
	if rv.Hosts[rv.OperatorHost] != "af_heart" { // operator 2026-10-09: af_heart & am_michael
		t.Errorf("operator lines are read by %q (%q)", rv.OperatorHost, rv.Hosts[rv.OperatorHost])
	}
	for _, who := range []string{"Chapter", "Picture", "Reporter Voice", "Brobot 3", "Brobot 4", "Countdown"} {
		if !gopodPageOnlyRole(who) {
			t.Errorf("%q should be page-only", who)
		}
	}
	for _, who := range []string{"Operator", "Brobot 1", "Brobot 2"} {
		if gopodPageOnlyRole(who) {
			t.Errorf("%q should be remembered", who)
		}
	}
}

// 2026-10-06: the reporter's word replacements (reporter_voices.say_words).
func TestReporterSayWords(t *testing.T) {
	for in, want := range map[string]string{
		"thx bros":      "thanks bros",
		"Thx, Brobots!": "thanks, bro-bots!",
		"thx thx":       "thanks thanks",
		"thxx bros":     "thxx bros",
	} {
		if got := gopodReporterSays(in); got != want {
			t.Errorf("%q -> %q, want %q", in, got, want)
		}
	}
}

// 2026-10-06: LED / LEDs / LED's are said "ellee-dee(z)"; "led" (he led the way) is not.
func TestLEDPronunciation(t *testing.T) {
	got := gopodReplaceWholeWords("The LED, the LEDs and LED's - he led the way.", loadedGOPODPronunciations)
	if got != "The ellee-dee, the ellee-deez and ellee-deez - he led the way." {
		t.Errorf("got %q", got)
	}
	got = gopodReplaceWholeWords("Doctrine Barfallonyou Squawkadoodle, or Squawkadoodle.", loadedGOPODPronunciations)
	if got != "Doctrine Bar-fahlonyoo, Squawka-doodle, or Squawka-doodle." {
		t.Errorf("Doc's name: %q", got)
	}
}

// 2026-10-06: before the green light, questions to the Brobots are dropped.
func TestWarmupGate(t *testing.T) {
	defer gopodWarmGateClosed.Store(false)
	gopodWarmGateClosed.Store(true)
	for _, esn := range []string{gopodBrobot1ESN, gopodBrobot2ESN, gopodTypedAll} {
		if !gopodWarmGateDrop(esn, "Typed text", "hello") {
			t.Errorf("%s: not dropped while closed", esn)
		}
	}
	if gopodWarmGateDrop("someotherbot", "voice", "hello") {
		t.Error("a robot that isn't a Brobot is never gated")
	}
	// 2026-10-13 (operator: "LLM warming = disabled/pause on all input"): a
	// typed line while closed waits in the warm-up queue, played once it opens.
	savedReplay := gopodWarmupReplayLine
	defer func() { gopodWarmupReplayLine = savedReplay }()
	replayed := make(chan string, 1)
	gopodWarmupReplayLine = func(esn, text, how string) error { replayed <- text; return nil }
	if err := GopodTypedInput(gopodBrobot1ESN, "who are you"); err != nil {
		t.Fatal(err)
	}
	gopodWarmGateClosed.Store(false)
	select {
	case got := <-replayed:
		if got != "who are you" {
			t.Errorf("replayed %q", got)
		}
	case <-time.After(3 * time.Second):
		t.Error("the queued line wasn't played once the gate opened")
	}
	if gopodWarmGateDrop(gopodBrobot1ESN, "Typed text", "hello") {
		t.Error("dropped while open")
	}
}

// 2026-10-06: the reporters' closing plays once per ending - a second thanks
// waits on the first and doesn't queue it again; a new question resets it.
func TestReporterClosingOncePerEnding(t *testing.T) {
	saved := gopodBrobotsSettings.ReporterVoices.Closing
	defer func() { gopodBrobotsSettings.ReporterVoices.Closing = saved; gopodReporterNewQuestion() }()
	gopodBrobotsSettings.ReporterVoices.Closing = nil
	gopodReporterNewQuestion()
	gopodReporterClosing()
	gopodReporter.Lock()
	first := gopodReporter.closing
	gopodReporter.Unlock()
	gopodReporterClosing()
	gopodReporter.Lock()
	second := gopodReporter.closing
	gopodReporter.Unlock()
	if first == nil || first != second {
		t.Fatal("a second thanks started its own closing")
	}
	gopodReporterNewQuestion()
	gopodReporter.Lock()
	reset := gopodReporter.closing
	gopodReporter.Unlock()
	if reset != nil {
		t.Fatal("a new question should reset the closing")
	}
}

// 2026-10-06 (live 16:12): a Brobot calling himself Brobot 0 is fixed to his
// own name; talk about Brobot 0 is left alone. The reporters: bro-bot, and each
// Brobot's line read by his own reporter.
func TestBrobotZeroSelfNameAndReporters(t *testing.T) {
	if got := gopodFixSelfName(gopodBrobot2ESN, "Brobot 0 here again! You know us."); got != "Brobot 2 here again! You know us." {
		t.Errorf("got %q", got)
	}
	if got := gopodFixSelfName(gopodBrobot1ESN, "I'm Brobot zero, bro."); got != "I'm Brobot 1, bro." {
		t.Errorf("got %q", got)
	}
	if got := gopodFixSelfName(gopodBrobot1ESN, "Brobot 0 is the one holding the mic."); got != "Brobot 0 is the one holding the mic." {
		t.Errorf("talk about Brobot 0 changed: %q", got)
	}
	if got := gopodReporterSays("Brobot 1, thx brobots"); got != "bro-bot 1, thanks bro-bots" {
		t.Errorf("got %q", got)
	}
	rv := gopodBrobotsSettings.ReporterVoices
	if rv.ReadBy["Brobot 1"] != "Brobot 3" || rv.ReadBy["Brobot 2"] != "Brobot 4" || rv.Hosts["Brobot 4"] != "am_michael" {
		t.Errorf("read_by %v hosts %v", rv.ReadBy, rv.Hosts)
	}
}

// 2026-10-06: the value-point list - one point at a time, then covered; the
// reminder points to the next uncovered topic; the score counts it all.
func TestValuePointCoverage(t *testing.T) {
	// 2026-10-23 (A3): the point mechanics, over every point (core_points off here).
	defer gopodAllPointsForTest()()
	gopodCoverage.Lock()
	saved := gopodCoverage.revealed
	gopodCoverage.revealed = map[string]int{}
	gopodCoverage.reminded = map[string]bool{}
	gopodCoverage.Unlock()
	defer func() {
		gopodCoverage.Lock()
		gopodCoverage.revealed = saved
		gopodCoverage.scored = false
		gopodCoverage.Unlock()
	}()

	pack := "Who's behind it. Reveal these, in order:\n1. First point.\n2. Second point.\n"
	one := string(gopodCoverageFacts("zz_test", []byte(pack)))
	two := string(gopodCoverageFacts("zz_test", []byte(pack)))
	all := string(gopodCoverageFacts("zz_test", []byte(pack)))
	if !strings.Contains(one, "1. First") || strings.Contains(one, "2. Second") || !strings.HasPrefix(one, "Who's behind it.") {
		t.Errorf("first: %q", one)
	}
	if !strings.Contains(two, "2. Second") || strings.Contains(two, "1. First") {
		t.Errorf("second: %q", two)
	}
	if all != pack || !gopodTopicCovered("zz_test", 2) {
		t.Errorf("covered: %q", all)
	}
	// 2026-10-07: reminders never say a value point. Mid-chapter (who_we_are),
	// three topics left, chained by likeness - not the next two (brothers,
	// doc_and_pip); after a chapter's last topic (doc_and_pip), the next two
	// chapters with their summaries.
	ic := gopodBrobotsSettings.InterviewChapters
	got := gopodReminderLine("Brobot 1, who are you?")
	if !strings.HasPrefix(got, "Hey, that reminds me") || strings.Contains(got, ic.TopicLabels["brothers"]) || strings.Contains(got, ic.TopicLabels["doc_and_pip"]) || strings.Count(got, ",")+strings.Count(got, " and ") < 2 {
		t.Errorf("topics reminder %q", got)
	}
	for _, line := range gopodBrobotsSettings.ValuePoints.SayLines {
		if strings.Contains(got, line) {
			t.Errorf("reminder says a value point: %q", got)
		}
	}
	if again := gopodReminderLine("Brobot 1, who are you?"); again == got {
		t.Errorf("the same topics were named again: %q", again)
	}
	got = gopodReminderLine("Brobot 2, what's Doc's full name?")
	if !strings.Contains(got, ic.Chapters[1].Title) || !strings.Contains(got, strings.TrimRight(ic.Chapters[1].Summary, ".")) || !strings.Contains(got, ic.Chapters[2].Title) {
		t.Errorf("chapters reminder %q", got)
	}
	text, detail := gopodCoverageScoreText()
	// 2026-10-23 (A3): the one score template - every token filled.
	if text == "" || !strings.Contains(detail, "doc_and_pip=") || strings.Contains(text, "{") {
		t.Errorf("score %q / %q", text, detail)
	}
}

// 2026-10-08: the 420 topic is out of the interview (operator: "100% remove
// weed/420 topic"), and the trimmed cut's words are said his way.
func TestNo420TopicAndTrimmedCutWords(t *testing.T) {
	for _, p := range loadScenarioPacketsForMessage("Brobot 1, what's 420 all about in Windsor?") {
		if p.Name == "cannabis_420" {
			t.Fatal("cannabis_420 still loads")
		}
	}
	for in, want := range map[string]string{
		"Go to crushn8r.ca now.":          "Go to Crush-Ehnaydir-dotseeay now.",
		"DocBS and PipTA.":                "Doc-B-S and Pip-T-A.",
		"An Anker Solix C1000 powers it.": "An Anker Solix see-one-thousand powers it.",
	} {
		if got := gopodReplaceWholeWords(in, loadedGOPODPronunciations); got != want {
			t.Errorf("%q said %q, want %q", in, got, want)
		}
	}
}

// 2026-10-06: "what else can we talk about?" gets the next chapters not covered.
func TestWhatElseCanWeTalkAbout(t *testing.T) {
	gopodCoverage.Lock()
	saved := gopodCoverage.revealed
	gopodCoverage.revealed = map[string]int{}
	gopodCoverage.reminded = map[string]bool{}
	gopodCoverage.Unlock()
	defer func() { gopodCoverage.Lock(); gopodCoverage.revealed = saved; gopodCoverage.Unlock() }()
	facts := gopodQuestionFacts("Brobot 2, what else can we talk about?")
	if !strings.Contains(facts, "- Warm-up") || !strings.Contains(facts, "- Brobot 0 and the comeback") || !strings.Contains(facts, "so far") {
		t.Fatalf("facts %q", facts)
	}
	_, counts := gopodPackPointCounts()
	gopodCoverage.Lock()
	for _, topic := range gopodBrobotsSettings.InterviewChapters.Chapters[0].Topics {
		gopodCoverage.revealed[topic] = counts[topic]
	}
	gopodCoverage.Unlock()
	if facts := gopodQuestionFacts("what else can we talk about?"); strings.Contains(facts, "- Warm-up") {
		t.Errorf("a covered chapter is still offered: %q", facts)
	}
}

// 2026-10-06: through the real prompt path, a topic gives one point, then the next.
func TestValuePointsReachThePrompt(t *testing.T) {
	// 2026-10-23 (A3): the point mechanics, over every point (core_points off here).
	defer gopodAllPointsForTest()()
	gopodCoverage.Lock()
	saved := gopodCoverage.revealed
	gopodCoverage.revealed = map[string]int{}
	gopodCoverage.reminded = map[string]bool{}
	gopodCoverage.Unlock()
	defer func() { gopodCoverage.Lock(); gopodCoverage.revealed = saved; gopodCoverage.Unlock() }()
	first := gopodQuestionFacts("Brobot 1, who's Brobot 0?")
	second := gopodQuestionFacts("Brobot 1, who's Brobot 0?")
	if !strings.Contains(first, "1. ") || strings.Contains(first, "2. ") {
		t.Errorf("first: %q", first)
	}
	if !strings.Contains(second, "2. ") || strings.Contains(second, "1. ") {
		t.Errorf("second: %q", second)
	}
}

// 2026-10-06: the scripted lines never reach the chat memory, and the LLM's own
// "Hey, that reminds me..." is cut - the code says it.
func TestScriptedLinesKeptFromMemory(t *testing.T) {
	remembered := "Yo, we rep GOPOD, bro! Hey, that reminds me, Doc explains too hard. Pip understands too late. The audience laughs, then learns."
	if got := gopodDropScriptedLines(remembered); got != "Yo, we rep GOPOD, bro!" {
		t.Errorf("reminder kept: %q", got)
	}
	if got := gopodDropScriptedLines("No - I'm on GOPOD! Ugh, TikTok. Join the CRUSHN8R CREW'd newsletter at crushn8r.ca!"); got != "Ugh, TikTok." {
		t.Errorf("flavour lines kept: %q", got)
	}
	if got := gopodDropLLMReminder("Party time! Hey, that reminds me... GOPOD your birthday!"); got != "Party time!" {
		t.Errorf("LLM reminder kept: %q", got)
	}
	if got := gopodDropLLMReminder("That reminds me of a party."); got != "That reminds me of a party." {
		t.Errorf("plain words cut: %q", got)
	}
}

// 2026-10-06: a hard value point's exact line (value_points.say_lines) is said
// first when that point is given out - to the Brobot it was given to - and each
// line is the point's own words.
func TestSayLinesForHardPoints(t *testing.T) {
	// 2026-10-23 (A3): the point mechanics, over every point (core_points off here).
	defer gopodAllPointsForTest()()
	gopodCoverage.Lock()
	saved := gopodCoverage.revealed
	gopodCoverage.revealed = map[string]int{}
	gopodCoverage.reminded = map[string]bool{}
	gopodCoverage.Unlock()
	defer func() { gopodCoverage.Lock(); gopodCoverage.revealed = saved; gopodCoverage.Unlock() }()
	say := gopodBrobotsSettings.ValuePoints.SayLines
	// 2026-10-22 (operator: "only the model's flavoured version of the notes"): a
	// point with notes has no exact line - Doc & Pip's points have notes now.
	ask := "Brobot 1, what's Doc's full name?"
	gopodQuestionFactsFor(gopodBrobot1ESN, ask)
	if got := gopodSayLine(gopodBrobot1ESN); got != "" {
		t.Errorf("a point with notes got an exact line: %q", got)
	}
	if !gopodGivenHasNotes(gopodBrobot1ESN) {
		t.Errorf("doc_and_pip 1 has no notes")
	}
	// Exact math keeps its exact lines (no notes).
	ask = "Brobot 1, convert a fraction to a decimal"
	gopodQuestionFactsFor(gopodBrobot1ESN, ask)
	if got := gopodSayLine(gopodBrobot1ESN); got != say["fractions_and_decimals 1"] {
		t.Errorf("point 1: %q", got)
	}
	// Said by the answer chain's colour:value_point part.
	if got := gopodChainStitch("colour:value_point", gopodBrobot1ESN, ask, "his colour.", true); !strings.Contains(got, say["fractions_and_decimals 1"]) {
		t.Errorf("colour:value_point didn't say point 1: %q", got)
	}
	if got := gopodSayLine(gopodBrobot2ESN); got != "" {
		t.Errorf("the other Brobot got a say line: %q", got)
	}
	gopodQuestionFactsFor(gopodBrobot1ESN, ask)
	if got := gopodSayLine(gopodBrobot1ESN); got != say["fractions_and_decimals 2"] {
		t.Errorf("point 2: %q", got)
	}
	reply := gopodStitchSnippet(say["doc_and_pip 1"], "happy___ "+say["doc_and_pip 1"]+" Yo bro!")
	if strings.Count(reply, "Squawkadoodle") != 1 {
		t.Errorf("say line twice: %q", reply)
	}

	_, counts := gopodPackPointCounts()
	words := regexp.MustCompile(`[\pL\pN']+`)
	norm := func(s string) string { return strings.ToLower(strings.Join(words.FindAllString(s, -1), " ")) }
	packs := map[string]string{}
	for _, p := range loadScenarioPacketsForMessage("") {
		packs[p.Name] = string(p.Content)
	}
	for key, line := range say {
		i := strings.LastIndex(key, " ")
		topic, n := key[:i], key[i+1:]
		// 2026-10-22: a background topic's lines aren't value points (CORE_TOPICS_002.md).
		if gopodBackgroundTopic(topic) {
			continue
		}
		if counts[topic] == 0 || n < "1" || len(n) != 1 || int(n[0]-'0') > counts[topic] {
			t.Errorf("%s: no such point", key)
			continue
		}
		content := packs[topic]
		if content == "" {
			for _, p := range gopodTopicRegistry() {
				if p.Name == topic {
					data, _ := gopodReadTopicFacts(p.File)
					content = string(data)
				}
			}
		}
		for _, part := range regexp.MustCompile(`\s+-\s+|[.:]\s+`).Split(line, -1) {
			if p := norm(part); p != "" && !strings.Contains(norm(content), p) {
				t.Errorf("%s: %q is not in the pack's words", key, part)
			}
		}
	}
}

// 2026-10-06: tighter replies - no lists, no stage directions, the first
// max_sentences sentences; a say line stitched in after is never cut.
func TestTightenReply(t *testing.T) {
	saved := gopodBrobotsSettings.ReplyTighten
	defer func() { gopodBrobotsSettings.ReplyTighten = saved }()
	gopodBrobotsSettings.ReplyTighten.BudgetOff = false
	gopodBrobotsSettings.ReplyTighten.BaseSentences = 3
	gopodBrobotsSettings.ReplyTighten.DropLists = true
	gopodBrobotsSettings.ReplyTighten.DropStageDirections = true
	in := "happy___ Dude, the **rig** is lit! *spins around in circles* Ah... so Dr. Doc runs it at crushn8r.ca all night.\nWe got:\n1. **Food:** pizza\n2. Money\n```\ncode\n```\nIt's *so* local, bro! 🚀 Fourth one. Fifth one?"
	got := gopodTightenReply(in, "can you hear me", "")
	if strings.Contains(gopodTightenReply("happy___ Hi [Name], we did it. [end]", "", ""), "[") {
		t.Error("a [placeholder] was kept")
	}
	if got := gopodTightenReply("happy___ We're in sync. *** It must be good.", "", ""); strings.Contains(got, "*") {
		t.Errorf("stray stars kept: %q", got)
	}
	if got := gopodTightenReply("happy___ Nah, bro, we test them. The response reflects each part of a great answer.", "", ""); strings.Contains(got, "response") {
		t.Errorf("the LLM's talk about its answer was kept: %q", got)
	}
	want := "happy___ Dude, the rig is lit! Ah... so Dr. Doc runs it at crushn8r.ca all night. We got: It's *so* local, bro! 🚀"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
	line := gopodBrobotsSettings.ValuePoints.SayLines["legacy 3"]
	if out := gopodStitchSnippet(line, got); !strings.Contains(out, line) {
		t.Errorf("say line lost: %q", out)
	}
	if got := gopodTightenReply("thinking___ 1. one\n2. two", "", ""); !strings.Contains(got, "one") {
		t.Errorf("an all-list reply keeps its words: %q", got)
	}
	// The response budget: a check-in gets the base, a "tell me all about" on
	// a topic gets more.
	gopodBrobotsSettings.ReplyTighten = saved
	short, long := gopodResponseBudget("Brobot 1, can you hear me?"), gopodResponseBudget("Brobot 1, tell me all about GOPOD")
	if short != saved.BaseSentences || long <= short || long > saved.MaxSentences {
		t.Errorf("budget: can you hear me=%d, tell me all about GOPOD=%d", short, long)
	}
	if got := gopodDropScriptedLines("Yo. " + line); got != "Yo." {
		t.Errorf("say line kept in memory: %q", got)
	}
}

// 2026-10-07: a draft performance plays back word for word - each Brobot gets
// the draft's answer to the draft's question, once; any other question is the
// LLM's.
func TestDraftReplayAnswers(t *testing.T) {
	draft := gopodDraft{Run: "test", Questions: []gopodDraftQuestion{
		{N: 1, To: "Brobot 1", Ask: "Brobot 1, what's the rig?", Answers: []gopodDraftAnswer{{Brobot: "Brobot 1", Cue: "thinking", Say: "Two crates on one dolly."}}},
		{N: 2, To: "Brobots", Ask: "Thanks Brobots!", Answers: []gopodDraftAnswer{{Brobot: "Brobot 1", Say: "Later."}, {Brobot: "Brobot 2", Cue: "happy", Say: "Bye!"}}},
	}}
	path := filepath.Join(t.TempDir(), "draft_performance.json")
	data, _ := json.Marshal(draft)
	os.WriteFile(path, data, 0o644)
	t.Setenv("GOPOD_DRAFT", path)
	got, loaded, err := gopodLoadDraft()
	if err != nil || got != path || len(loaded.Questions) != 2 {
		t.Fatalf("load: %v %s", err, got)
	}
	gopodScript.Lock()
	gopodScript.pending[gopodBrobot1ESN], gopodScript.ask[gopodBrobot1ESN] = draft.Questions[0].Answers[0], draft.Questions[0].Ask
	gopodScript.Unlock()
	if _, ok := gopodScriptAnswer(gopodBrobot1ESN, "tell me a joke"); ok {
		t.Error("another question got the draft's answer")
	}
	if a, ok := gopodScriptAnswer(gopodBrobot1ESN, "Brobot 1, what's the rig?"); !ok || a != "thinking"+gopodAnimationMarker+" Two crates on one dolly." {
		t.Errorf("answer: %q %v", a, ok)
	}
	if _, ok := gopodScriptAnswer(gopodBrobot1ESN, "Brobot 1, what's the rig?"); ok {
		t.Error("the draft's answer was given twice")
	}
	// 2026-10-09: "start interview" plays it; "stop run" stops it (gopodStopRunCommand).
	if gopodScriptCommand("Stop draft") || gopodScriptCommand("what's the draft beer?") || gopodScriptCommand("start the interview soon") {
		t.Error("play/stop words")
	}
}

// 2026-10-07 dry run: a word in the cue's place was shown and said ("happy_ My
// peeps", "loving... Two crates"). It is read as the cue; mid-sentence words
// and plain openers stay.
func TestCueSlotLeaks(t *testing.T) {
	cases := map[string]string{
		"happy_ My peeps call me Brobot 2.":     "My peeps call me Brobot 2.",
		"loving... Two crates on one dolly.":    "Two crates on one dolly.",
		"loving___ Two crates on one dolly.":    "Two crates on one dolly.",
		"laughing... Pick an idea.":             "Pick an idea.",
		"blorp___ Two crates on one dolly.":     "Two crates on one dolly.",
		"party... time!! Let's go.":             "party... time!! Let's go.",
		"Okay... So here's the deal.":           "Okay... So here's the deal.",
		"thinking___ Two crates on one dolly.":  "Two crates on one dolly.",
		"happy--- Awaken: a brobot wakes up.":   "Awaken: a brobot wakes up.",
		"darlingeyes---- Is that you?":          "Is that you?",
		"dar*ingEyes... You've got a good ear.": "You've got a good ear.",
		"Well-- listen up, bro.":                "Well-- listen up, bro.",
	}
	for in, want := range cases {
		if got := strings.TrimSpace(gopodDisplayReply(in)); got != want {
			t.Errorf("%q: bubble %q, want %q", in, got, want)
		}
	}
	if got := gopodFixOpeningCueSlot("loving... Two crates."); !strings.HasPrefix(got, "love") {
		t.Errorf("loving -> %q", got)
	}
}

// 2026-10-07: the Brobots' creativity - lower for a question whose point has a
// scripted line, as-is otherwise.
func TestSayLineTemperature(t *testing.T) {
	gopodCoverage.Lock()
	saved := gopodCoverage.revealed
	gopodCoverage.revealed = map[string]int{}
	gopodCoverage.Unlock()
	defer func() { gopodCoverage.Lock(); gopodCoverage.revealed = saved; gopodCoverage.Unlock() }()
	t.Setenv("GOPOD_SESSION_DIR", t.TempDir())
	llm := gopodBrobotsSettings.LLM
	if req := CreateAIReq("Brobot 1, what's Doc's full name?", gopodBrobot1ESN, false, false); req.Temperature != llm.SayLineTemperature {
		t.Errorf("scripted-line question: %v, want %v", req.Temperature, llm.SayLineTemperature)
	}
	if req := CreateAIReq("Brobot 1, tell me a joke", gopodBrobot1ESN, false, false); req.Temperature != llm.Temperature {
		t.Errorf("plain question: %v, want %v", req.Temperature, llm.Temperature)
	}
}

// 2026-10-07: speech rules - pumpin'/jumpin', Hmm, "@ me", a colon pause, '20.
func TestSpeechRules20261007(t *testing.T) {
	got := gopodReplaceWholeWords("Let's get pumpin', jumpin' and feelin's.", loadedGOPODPronunciations)
	for _, want := range []string{"pum-pin", "jum-pin", "feelin's"} {
		if !strings.Contains(got, want) {
			t.Errorf("%q missing %q", got, want)
		}
	}
	if got := gopodAtSignRe.ReplaceAllString("Don't even @ me!", "at "); got != "Don't even at me!" {
		t.Errorf("@: %q", got)
	}
	if got := gopodColonPauseRe.ReplaceAllString("Doc's Take: 1k subs at 4:20.", "$1, "); got != "Doc's Take, 1k subs at 4:20." {
		t.Errorf("colon: %q", got)
	}
	if got := gopodShortYearRe.ReplaceAllString("the summer of '20 man", "$1$2"); got != "the summer of 20 man" {
		t.Errorf("year: %q", got)
	}
	for in, want := range map[string]string{"Bars, 18+.": "Bars, 18 plus.", "2+2 is 4": "2 plus 2 is 4", "Ages 4 + up": "Ages 4 plus up"} {
		if got := gopodSayPlus(in); got != want {
			t.Errorf("+: %q said %q, want %q", in, got, want)
		}
	}
}

// 2026-10-07: a long dash splits the bubble's highlight where the robot pauses.
func TestHighlightLongDash(t *testing.T) {
	marked, seg := gopodSegmentDisplay("Brobot 0 ran the cameras—literally.", []string{"Brobot 0 ran the cameras,", "literally."})
	if seg[0] == seg[1] || !strings.Contains(marked, "—"+gopodBrobotsSettings.ScriptedSplit.SegmentMarker+"literally") {
		t.Errorf("not split at the dash: %q %v", marked, seg)
	}
}

// 2026-10-08: math symbols, "$", "#" and LaTeX math are said in words.
func TestSpokenSymbols(t *testing.T) {
	for in, want := range map[string]string{
		`Half is \(\frac{1}{2}\) or 0.5.`: "Half is 1/2 or 0.5.",
		`One third is 0.\overline{3}.`:    "One third is 0.3 repeating.",
		"Ten ninths is 1.1̅.":             "Ten ninths is 1.1 repeating.",
		"6 × 7 ÷ 2 ≈ 21":                  "6 times 7 divided by 2 almost equals 21",
		"It costs $5, or $2.50 on sale.":  "It costs 5 dollars, or 2.50 dollars on sale.",
		"Bring cash $ and use #GOPOD":     "Bring cash dollars and use hashtag GOPOD",
		"It's 20° and π is near 3.14":     "It's 20 degrees and pi is near 3.14",
		"Hmm... Hm, HMMMM, hmmm.":         "hmm-m-m... hmm-m-m, hmm-m-m, hmm-m-m.",
		`\(3 \times 4 = 12\)`:             "3 times 4 equals 12", // 2026-10-14: "=" said "equals" (math)
	} {
		if got := gopodSpeakSymbols(in); got != want {
			t.Errorf("%q said %q, want %q", in, got, want)
		}
	}
}

// 2026-10-08: a piece whose words are all said differently still gets its own
// highlight ("PipTA" said "Pip-T-A", "crushn8r.ca" said "Crush-Ehnaydir-dotseeay").
func TestHighlightRenamedWords(t *testing.T) {
	for text, spoken := range map[string][]string{
		"Pip is Pip N Stall — pip install, PipTA, the teacher's assistant.": {"Pip is Pip N Stall,", "pip install,", "Pip-T-A,", "the teacher's assistant."},
		"crushn8r.ca. The CREW'd newsletter — you're joining the crew.":     {"Crush-Ehnaydir-dotseeay.", "The crewd newsletter,", "you're joining the crew."},
	} {
		_, seg := gopodSegmentDisplay(text, spoken)
		for i, s := range seg {
			if s != i {
				t.Errorf("%q: pieces %v, want each its own", text, seg)
				break
			}
		}
	}
}

// 2026-10-08: the live weather's clock words (golden Awaken weather), and an
// offline {weather} sentence left out.
func TestLiveWeatherWords(t *testing.T) {
	for _, c := range []struct {
		h, m        int
		mil, twelve string
	}{
		{20, 25, "twenty twenty-five", "eight twenty-five pm"},
		{8, 0, "zero eight hundred", "eight o'clock 'a' m"},
		{0, 5, "zero zero five", "twelve oh five 'a' m"},
		{12, 30, "twelve thirty", "twelve thirty pm"},
	} {
		if got := gopodMilitaryTimeSpoken(c.h, c.m); got != c.mil {
			t.Errorf("%d:%02d 24h %q, want %q", c.h, c.m, got, c.mil)
		}
		if got := gopodTwelveHourTimeSpoken(c.h, c.m); got != c.twelve {
			t.Errorf("%d:%02d 12h %q, want %q", c.h, c.m, got, c.twelve)
		}
	}
	gopodWeatherCache.Lock()
	gopodWeatherCache.places = map[string]gopodWeatherCached{}
	gopodWeatherCache.Unlock()
	// Offline: offline_text stands in for {weather}.
	off := gopodBrobotsSettings.LiveWeather.OfflineText
	if off == "" {
		t.Fatal("live_weather.offline_text is empty")
	}
	if got := gopodScriptFill(gopodBrobot1ESN, "I wake up. The weather: {weather} Ready to party!"); got != "I wake up. The weather: "+off+" Ready to party!" {
		t.Errorf("offline: %q", got)
	}
	// With no offline_text, the weather's sentence is left out.
	gopodBrobotsSettings.LiveWeather.OfflineText = ""
	defer func() { gopodBrobotsSettings.LiveWeather.OfflineText = off }()
	if got := gopodScriptFill(gopodBrobot1ESN, "I wake up. The weather: {weather} Ready to party!"); got != "I wake up. Ready to party!" {
		t.Errorf("offline, no offline_text: %q", got)
	}
}

// 2026-10-13 (operator: "toronto weather was troublesome"): a weather question
// that names a place gets that place's weather.
func TestWeatherPlace(t *testing.T) {
	for q, want := range map[string]string{
		"what's the weather in toronto ontario, right now?": "toronto ontario",
		"Current weather for Toronto, ON, CA, please":       "Toronto, ON, CA",
		"Current weather please":                            "",
		"Did brobot 1 tell me the weather in Toronto?":      "Toronto",
		"what's the weather like here?":                     "",
		"Hey Brobot 1, how's the weather in Windsor today":  "Windsor",
		// 2026-10-16: the place named before the weather.
		"I want to visit my granny in Honolulu Hawaii. How is the weather there today?": "Honolulu Hawaii",
		"my granny lives in jerusalem what's the weather there today":                   "jerusalem",
		"what's the weather in new york today":                                          "new york",
		"I want to go to the party, how is the weather?":                                "",
		"Say hi to Brobot 2. How is the weather?":                                       "",
	} {
		if got := gopodWeatherPlace(q); got != want {
			t.Errorf("%q: got %q, want %q", q, got, want)
		}
	}
}

// TestWeatherPickPlace (2026-10-14, live 13:14: "Is it the same party weather in
// Windsor Ontario?" was answered with Windsor England): the named province wins.
func TestWeatherPickPlace(t *testing.T) {
	got := []gopodGeoPlace{
		{Name: "Windsor", State: "England", Country: "GB"},
		{Name: "Windsor", State: "Ontario", Country: "CA"},
		{Name: "Windsor", State: "Connecticut", Country: "US"},
	}
	cases := map[string]int{
		"Windsor Ontario":     1,
		"windsor, ontario":    1,
		"Windsor Connecticut": 2,
		"Windsor England":     0,
		"Windsor":             0,
		"Windsor CA":          1,
		"Windsor Narnia":      -1,
	}
	for place, want := range cases {
		if i := gopodPickPlace(got, place); i != want {
			t.Errorf("gopodPickPlace(%q) = %d, want %d", place, i, want)
		}
	}
}

// TestSpeakMilitaryTime (2026-10-14, operator: "18:14 (Military Time) ...
// generate a 'Military Time' statement for robot safe say"): the bubble keeps
// "18:14 (Military Time)", the robot says it in words.
func TestSpeakMilitaryTime(t *testing.T) {
	lw := gopodBrobotsSettings.LiveWeather
	if lw.MilitaryLabel == "" {
		t.Skip("no live_weather.military_label")
	}
	label, said := lw.MilitaryLabel, strings.ToLower(lw.MilitaryLabel)
	cases := map[string]string{
		"as of 6:14 PM or 1814 " + label + ".": "as of 6:14 PM or eighteen fourteen " + said + ".",
		"as of 0800 " + label + " or 8:00 AM.": "as of zero eight hundred " + said + " or 8:00 AM.",
		"17° Celsius or 62° Fahrenheit":        "17 degrees Celsius or 62 degrees Fahrenheit",
		"Meet at 1814 bro":                     "Meet at 1814 bro",
	}
	// The operator's table (2026-10-14, "Military Time Rules").
	table := map[[2]int]string{
		{0, 0}: "zero hundred", {1, 0}: "zero one hundred", {9, 0}: "zero nine hundred",
		{10, 0}: "ten hundred", {12, 0}: "twelve hundred", {18, 0}: "eighteen hundred",
		{18, 14}: "eighteen fourteen", {18, 30}: "eighteen thirty", {23, 59}: "twenty-three fifty-nine",
		{6, 30}: "zero six thirty", {7, 9}: "zero seven zero nine",
	}
	for hm, want := range table {
		if got := gopodMilitaryTimeSpoken(hm[0], hm[1]); got != want {
			t.Errorf("gopodMilitaryTimeSpoken(%02d%02d) = %q, want %q", hm[0], hm[1], got, want)
		}
	}
	for in, want := range cases {
		if got := gopodSpeakSymbols(in); got != want {
			t.Errorf("gopodSpeakSymbols(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestPopupsOnlyLive (2026-10-14, operator: "treat these as found easter eggs
// throughout a live chat * ensure live pop-ups are not triggered by a song
// run"): a topic question finds its pop-up in the live chat, never while a run
// is going.
func TestPopupsOnlyLive(t *testing.T) {
	tp := gopodBrobotsSettings.TopicPictures
	topic := ""
	for name, files := range tp.Pictures {
		// 2026-10-19: an egg taken out (topic_pictures.not_eggs) pops nothing.
		if key, _ := gopodTopicChapter(name); len(files) > 0 && !gopodNotEgg(key) && !gopodNotEgg(name) {
			topic = name
			break
		}
	}
	if tp.Role == "" || topic == "" {
		t.Skip("no topic pictures in the settings")
	}
	question := ""
	for _, msg := range []string{"tell me about " + topic, "what about " + strings.ReplaceAll(topic, "_", " "), "bingo", "granny"} {
		for _, p := range loadScenarioPacketsForMessage(msg) {
			if len(tp.Pictures[p.Name]) > 0 {
				question = msg
			}
		}
		if question != "" {
			break
		}
	}
	if question == "" {
		t.Skip("no question loads a topic with pictures")
	}
	savedFeed := gopodChatFeedPath
	savedMax, savedLag := gopodBrobotsSettings.ReporterVoices.MaxWaitSeconds, gopodBrobotsSettings.ReporterVoices.PageLagSeconds
	defer func() {
		gopodChatFeedPath = savedFeed
		gopodBrobotsSettings.ReporterVoices.MaxWaitSeconds, gopodBrobotsSettings.ReporterVoices.PageLagSeconds = savedMax, savedLag
		gopodSession.Lock()
		gopodSession.run = nil
		gopodSession.Unlock()
	}()
	gopodBrobotsSettings.ReporterVoices.MaxWaitSeconds, gopodBrobotsSettings.ReporterVoices.PageLagSeconds = 0.01, 0
	gopodChatFeedPath = filepath.Join(t.TempDir(), "feed.json")
	popups := func() int {
		gopodChatFeed.Lock()
		defer gopodChatFeed.Unlock()
		n := 0
		for _, e := range gopodChatFeed.events {
			if strings.HasPrefix(e.Text, gopodPopupMark) {
				n++
			}
		}
		gopodChatFeed.next, gopodChatFeed.events = 0, nil
		return n
	}
	reset := func() {
		gopodInterviewShown.Lock()
		gopodInterviewShown.pictures = map[string]int{}
		gopodInterviewShown.Unlock()
		popups()
	}
	reset()
	gopodShowTopicPictures(gopodBrobot2ESN, question)
	if popups() == 0 {
		t.Fatalf("%q found no pop-up in the live chat", question)
	}
	reset()
	gopodSession.Lock()
	gopodSession.run = &gopodRun{Name: "Fart Joke", Kind: "song"}
	gopodSession.Unlock()
	gopodShowTopicPictures(gopodBrobot2ESN, question)
	if n := popups(); n != 0 {
		t.Fatalf("%q found %d pop-ups during a run", question, n)
	}
}

// 2026-10-23 (operator: "either the machine (game-mechanism) fucking works or it
// doesn't, regardless of the editable data"): a point with notes always comes out
// as its bubbles - a title line, then one line per note - whatever the model said.
func TestPointBubblesShapeAlwaysHolds(t *testing.T) {
	gopodCoverage.Lock()
	saved := gopodCoverage.revealed
	gopodCoverage.revealed = map[string]int{}
	gopodCoverage.Unlock()
	defer func() { gopodCoverage.Lock(); gopodCoverage.revealed = saved; gopodCoverage.Unlock() }()
	gopodQuestionFactsFor(gopodBrobot1ESN, "Brobot 1, tell me about Pip")
	title, notes := gopodPointBubblesNotes(gopodBrobot1ESN)
	if title == "" || len(notes) == 0 {
		t.Fatalf("no notes given: %v", gopodGivenPoints(gopodBrobot1ESN))
	}
	for _, raw := range []string{"", "happy___ Toronto, eh? I love dinosaurs.", "happy___ " + title + " rocks! " + notes[0] + " for sure."} {
		got := gopodPointBubbles(gopodBrobot1ESN, "Brobot 1, tell me about Pip", raw)
		lines := strings.Split(got, gopodThoughtGap())
		if len(lines) < len(notes)+1 || !strings.Contains(got, title) {
			t.Errorf("%q -> %q", raw, got)
		}
	}
}

// 2026-10-23 (FINISH_PLAN_001.md step A2): every answer type, run through the real
// answer chain with fixed model words, has its golden shape (answer_shapes).
func TestAnswerShapesHold(t *testing.T) {
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
		facts: gopodWeatherFacts{Name: "Windsor", State: "Ontario", Country: "CA", Description: "clear sky", TempC: 20, Local: time.Date(2026, 10, 23, 15, 49, 0, 0, time.UTC)},
		at:    time.Now(),
	}}
	gopodWeatherCache.Unlock()
	defer func() { gopodWeatherCache.Lock(); gopodWeatherCache.places = savedWx; gopodWeatherCache.Unlock() }()
	raw := "happy___ Nice one, bro! That's the vibe."
	// 2026-10-23 (operator: "input to a single brobot 1 or 2 * input to 'All'
	// brobots starting with 1 or 2"): every type, at every place an answer can be.
	asks := []struct{ name, ask string }{
		{"weather", "what's the weather?"},
		{"topic", "tell me about Pip"},
		{"free", "what's your favourite colour?"},
		{"math", "convert a fraction to a decimal"},
		{"thanks", "thanks brobots"},
		{"summary", "what is AI wordplay?"},
	}
	places := []struct {
		name        string
		esn         string
		first, next bool
	}{
		{"single to Brobot 1", gopodBrobot1ESN, false, false},
		{"single to Brobot 2", gopodBrobot2ESN, false, false},
		{"All starting with 1, Brobot 1 first", gopodBrobot1ESN, true, false},
		{"All starting with 1, Brobot 2 next", gopodBrobot2ESN, false, true},
		{"All starting with 2, Brobot 2 first", gopodBrobot2ESN, true, false},
		{"All starting with 2, Brobot 1 next", gopodBrobot1ESN, false, true},
	}
	type shapeCase struct {
		name, ask, esn string
		first, next    bool
	}
	var cases []shapeCase
	for _, a := range asks {
		for _, p := range places {
			cases = append(cases, shapeCase{a.name + ", " + p.name, a.ask, p.esn, p.first, p.next})
		}
	}
	for _, c := range cases {
		if c.first {
			gopodAllFirstAnswer.Store(c.esn, true)
		}
		if c.next {
			gopodFollowingUp.Store(c.esn, true)
		}
		gopodQuestionFactsFor(c.esn, c.ask)
		got := gopodShapeBrobotAnswerParts(c.esn, c.ask, raw)
		shape := gopodAnswerShape(c.esn, c.ask, c.first)
		labels := gopodShapeLabels(c.esn, got)
		if len(shape) == 0 || !gopodShapeMatches(shape, labels) {
			t.Errorf("%s (%s): %v, want %v\n%s", c.name, gopodAnswerType(c.esn, c.ask), labels, shape, got)
		}
		gopodAllFirstAnswer.Delete(c.esn)
		gopodFollowingUp.Delete(c.esn)
		// A fresh list for the next case (the pull names what's left).
		gopodCoverage.Lock()
		gopodCoverage.revealed, gopodCoverage.reminded = map[string]int{}, map[string]bool{}
		gopodCoverage.Unlock()
	}
}

// gopodAllPointsForTest: every point a value point while a test runs (the
// mechanics tests); returns the undo.
func gopodAllPointsForTest() func() {
	saved := gopodBrobotsSettings.ValuePoints.CorePoints
	savedTopics := gopodBrobotsSettings.ValuePoints.CoreTopics
	gopodBrobotsSettings.ValuePoints.CorePoints = nil
	gopodBrobotsSettings.ValuePoints.CoreTopics = nil
	return func() {
		gopodBrobotsSettings.ValuePoints.CorePoints = saved
		gopodBrobotsSettings.ValuePoints.CoreTopics = savedTopics
	}
}

// 2026-10-23 (operator: "for brobot 2 first, the mechanics appear to work
// correctly, when my intent is to call brobot 2 first, so please soft lock this
// templated mechanism"): the All menu's 2, or the words, make Brobot 2 duo A and
// Brobot 1 duo B; the menu's 1 (the default) the other way round.
func TestDuoAnswersHold(t *testing.T) {
	line := "what's the weather?"
	GopodAllFirstPick(line, "2")
	if gopodAllFirst(line) != gopodBrobot2ESN {
		t.Errorf("menu 2: Brobot 2 isn't first")
	}
	GopodAllFirstPick(line, "1")
	if gopodAllFirst(line) != gopodBrobot1ESN {
		t.Errorf("menu 1: Brobot 1 isn't first")
	}
	if gopodAllFirst("Starting with Brobot 2, "+line) != gopodBrobot2ESN {
		t.Errorf("the words: Brobot 2 isn't first")
	}
	// 2026-10-23: the Brobot named first anywhere in the line answers first.
	if gopodAllFirst("for both of you bro but to what's the weather like") != gopodBrobot2ESN {
		t.Errorf("named first (Brobot 2, mid-line): not first")
	}
	if gopodAllFirst("hey you two, brobot one then brobot two, what's up?") != gopodBrobot1ESN {
		t.Errorf("named first (Brobot 1): not first")
	}
	GopodAllFirstPick("", "")
	gopodAllFirstAnswer.Store(gopodBrobot2ESN, true)
	if r, ok := gopodDuoRule(gopodBrobot2ESN); !ok || r.Chain != gopodBrobotsSettings.Duo["a"].Chain || gopodBrobotChainGroup(gopodBrobot2ESN) != r.Chain {
		t.Errorf("Brobot 2 first isn't duo A")
	}
	gopodAllFirstAnswer.Delete(gopodBrobot2ESN)
	gopodFollowingUp.Store(gopodBrobot1ESN, true)
	if r, ok := gopodDuoRule(gopodBrobot1ESN); !ok || r.Chain != gopodBrobotsSettings.Duo["b"].Chain || gopodBrobotChainGroup(gopodBrobot1ESN) != r.Chain {
		t.Errorf("Brobot 1 next isn't duo B")
	}
	gopodFollowingUp.Delete(gopodBrobot1ESN)
}

// 2026-10-23 (FINISH_PLAN_001.md A4): every value point revealed gets the grand
// prize once; the reset starts the next session at 0.
func TestGameWinAndReset(t *testing.T) {
	gopodCoverage.Lock()
	saved, savedRem := gopodCoverage.revealed, gopodCoverage.reminded
	gopodCoverage.Unlock()
	defer func() {
		gopodCoverage.Lock()
		gopodCoverage.revealed, gopodCoverage.reminded = saved, savedRem
		gopodCoverage.Unlock()
	}()
	gopodGameReset("test")
	if gopodGameAllRevealed() || gopodGameWinCheck() {
		t.Fatalf("won at 0")
	}
	names, counts := gopodPackPointCounts()
	gopodCoverage.Lock()
	for _, n := range names {
		gopodCoverage.revealed[n] = counts[n]
	}
	gopodCoverage.Unlock()
	if !gopodGameAllRevealed() {
		t.Fatalf("not all revealed")
	}
	text := gopodGameWinText()
	score, _ := gopodCoverageScoreText()
	if text == "" || !strings.Contains(text, score) || strings.Contains(text, "{") {
		t.Errorf("win text %q", text)
	}
	gopodGame.Lock()
	gopodGame.won = true // the line itself plays on the page; once is the rule
	gopodGame.Unlock()
	if gopodGameWinCheck() {
		t.Errorf("won twice")
	}
	gopodGameReset("test")
	if gopodGameAllRevealed() {
		t.Errorf("the reset didn't start at 0")
	}
	if s, _ := gopodCoverageScoreText(); !strings.Contains(s, "Value Points 0 of") {
		t.Errorf("score after reset %q", s)
	}
}

// 2026-10-23 (live: "a day of parties #1 of 3" with only #1 a value point;
// operator: "Please template these bugs out of GOPOD"): every value point's notice
// line says its place among its topic's value points, "#N of N" - never the topic
// file's line count.
func TestNoticeCountsValuePoints(t *testing.T) {
	core := gopodBrobotsSettings.ValuePoints.CorePoints
	if len(core) == 0 {
		t.Skip("no core_points")
	}
	perTopic := map[string][]int{}
	for _, id := range core {
		f := strings.SplitN(id, "#", 2)
		n, _ := strconv.Atoi(f[1])
		perTopic[f[0]] = append(perTopic[f[0]], n)
	}
	for topic, nums := range perTopic {
		sort.Ints(nums)
		var content []byte
		for _, p := range gopodTopicRegistry() {
			if p.Name == topic {
				content, _ = gopodReadTopicFacts(p.File)
			}
		}
		_, points := gopodSplitPoints(string(content))
		total := len(points)
		if total == 0 {
			total = 1
		}
		for rank, n := range nums {
			line := gopodNoticePoint(topic, content, fmt.Sprintf("%s %d/%d", topic, n, total))
			want := fmt.Sprintf("#%d of %d", rank+1, len(nums))
			if !strings.Contains(line, want) {
				t.Errorf("%s#%d: %q, want %q", topic, n, line, want)
			}
		}
	}
}

// 2026-10-23 (operator: "Please template these bugs out of GOPOD"): the duo run's
// bugs stay out - duo B calls no manager (no second "Thanks for your help"); a
// country alone with its continent is that country ("zimbabwe africa"), its
// capital the place; "bro but to" is Brobot 2.
func TestDuoRunBugsStayOut(t *testing.T) {
	if b := gopodBrobotsSettings.Duo["b"]; !b.NoManager {
		t.Errorf("duo B may call a manager again")
	}
	city, country := gopodWeatherCountry("zimbabwe africa")
	if country != "ZW" {
		t.Errorf("zimbabwe africa: country %q", country)
	}
	for _, c := range gopodBrobotsSettings.CleanPolitics.Continents {
		city = strings.TrimSpace(strings.ReplaceAll(" "+strings.ToLower(city)+" ", " "+c+" ", " "))
	}
	if city == "" {
		city = gopodBrobotsSettings.CleanPolitics.Capitals[country]
	}
	if city != "Harare" {
		t.Errorf("zimbabwe africa: place %q, want Harare", city)
	}
	// (A name counts in a line's first words: addressing.start_within_words.)
	if got := gopodAddressedTo("bro but to, what's the weather like"); got != "Brobot 2" {
		t.Errorf("bro but to: %q", got)
	}
}

// 2026-10-23 (FINISH_PLAN_001.md A5): the interview song built from the game
// reveals every value point exactly once, chapter by core topic, and plays (its
// story and knobs read back as the same draft).
func TestInterviewFromGame(t *testing.T) {
	d, err := gopodInterviewDraftFromGame([]gopodDraftLine{{Who: "Brobot 2", Text: "Press my backpack."}})
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]int{}
	for _, q := range d.Questions {
		if q.Ask == "" || q.Chapter == "" || len(q.Answers) != 1 || strings.TrimSpace(q.Answers[0].Say) == "" {
			t.Errorf("question %d incomplete: %+v", q.N, q)
		}
		for _, p := range q.Answers[0].Points {
			seen[p]++
		}
	}
	for _, id := range gopodBrobotsSettings.ValuePoints.CorePoints {
		if seen[id] != 1 {
			t.Errorf("%s revealed %d times", id, seen[id])
		}
	}
	story, knobs, err := gopodInterviewSongFiles(d)
	if err != nil {
		t.Fatal(err)
	}
	texts, err := gopodInterviewStory(story)
	if err != nil || len(texts) == 0 || len(knobs) == 0 {
		t.Fatalf("story didn't read back: %v", err)
	}
}

// 2026-10-23 (operator: "a template-oriented machine is supposed to snuff out your
// repeated 'oops, sorry' * is there still any lazy patch work in my clean GOPOD?"):
// every voiced line goes through the one template - gopodReporterSay(..., show
// true, ...): its bubble, Generating and the hourglass, then its words lit. The
// voice-only path (gopodSystemSay, gopodSystemSayNoSting, gopodReporterSay with
// show false) is allowed only where another template already shows that bubble,
// each named here with its reason. New code on the voice-only path fails the deploy.
func TestEveryVoicedLineUsesTheTemplate(t *testing.T) {
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("GOPOD_OVERLAY_CHIPPER not set")
	}
	allowed := map[string]int{
		"gopod_chat_feed.go":          1, // a song's system_say_role line: its warm-up bubble shows it
		"gopod_interview_chapters.go": 1, // an egg: its pop-up bubble waits as one line (lineWaits)
		"gopod_llm_warmup.go":         3, // the LLM warm-up bubble's own lines (llm_warmup)
		"gopod_opener_song.go":        3, // the opener's warm-up bubble lines (golden opener)
		"gopod_reporter_voices.go":    4, // the template itself, and the Operator's own line (generating:op)
	}
	re := regexp.MustCompile(`gopodSystemSay\(|gopodSystemSayNoSting\(|gopodReporterSay\([^()]*,[^()]*,\s*false\s*,`)
	files, _ := filepath.Glob(filepath.Join(root, "pkg/wirepod/ttr/*.go"))
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, _ := os.ReadFile(f)
		n := 0
		for _, line := range strings.Split(string(data), "\n") {
			l := strings.TrimSpace(line)
			if strings.HasPrefix(l, "//") || strings.HasPrefix(l, "func ") {
				continue
			}
			if re.MatchString(l) {
				n++
			}
		}
		if n > allowed[filepath.Base(f)] {
			t.Errorf("%s: %d voice-only calls (allowed %d) - use gopodReporterSay(role, text, true, wait), the one voiced-line template", filepath.Base(f), n, allowed[filepath.Base(f)])
		}
	}
}

// TestWordsLiveInTheSettings (2026-10-23, operator: "TEMPLATE IT ALL OUT UNTIL
// ALL THAT's LEFT IS TO MOVE FLEXIBLE DATA INTO LOADABLE/EDITABLE FILES"): the
// words the Brobots are told and the page's waits live in the settings file,
// never back in code.
func TestWordsLiveInTheSettings(t *testing.T) {
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("GOPOD_OVERLAY_CHIPPER not set")
	}
	s := gopodBrobotsSettings
	f := s.HostPersonaFrames
	sm := s.SessionMemory
	for name, v := range map[string]string{
		"chain_prompts.weather_lead": s.ChainPrompts["weather_lead"], "chain_prompts.weather_lead_next": s.ChainPrompts["weather_lead_next"],
		"interview_build.default_ask": s.InterviewBuild.DefaultAsk,
		"host_persona_frames.name":    f.Name, "host_persona_frames.steps": f.Steps, "host_persona_frames.signature": f.Signature, "host_persona_frames.dial": f.Dial,
		"session_memory.line": sm.Line, "session_memory.scenario": sm.Scenario, "session_memory.empty_text": sm.EmptyText,
	} {
		if strings.TrimSpace(v) == "" {
			t.Errorf("%s is empty", name)
		}
	}
	if len(s.InterviewBuild.Reporters) == 0 || len(s.BrotherRepeat.PlainWords) == 0 {
		t.Error("interview_build.reporters or brother_repeat.plain_words is empty")
	}
	gone := []string{"First one short line about this weather", "As a GOPOD host ", " way through an answer: ", "When it fits, ", " stereo dial: ",
		"* User asked ", "Prior matched scenario: ", "a short GOPOD point", "Tell us about ", `"Reporter 1", "Reporter 2"`, "gonna wanna"}
	files, _ := filepath.Glob(filepath.Join(root, "pkg/wirepod/ttr/*.go"))
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		data, _ := os.ReadFile(file)
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue
			}
			for _, g := range gone {
				if strings.Contains(line, g) {
					t.Errorf("%s: %q is back in code - it lives in the settings", filepath.Base(file), g)
				}
			}
		}
	}
	page, _ := os.ReadFile(filepath.Join(root, "webroot/js/gopod_brobots_chat.js"))
	for _, k := range []string{"speaking_max_ms", "voice_ready_max_ms", "reveal_rest_ms", "marquee_pattern_ms", "baton_max_ms"} {
		if !strings.Contains(string(page), "cw."+k) {
			t.Errorf("the page doesn't read chat_window.%s", k)
		}
	}
	// The persona sentences read the same as before the move.
	if got := gopodPersonaFrame(f.Name, "I", "My", "{name}", "Doc"); got != "As a GOPOD host I go by Doc." {
		t.Errorf("persona name line: %q", got)
	}
}

// TestInterviewEndAndPauseHold (2026-10-23, live: the interview's end held 4.5
// min; a tab flicker said paused/continues twice; the question bubble had no
// Generating or highlight): a later bubble never holds an earlier voice; the
// voice-only line (the Operator's, a warm-up line) gets the voice template too
// ("forwait:"/"forlit:"); a page
// hidden shorter than session.pause_after_ms pauses nothing.
func TestInterviewEndAndPauseHold(t *testing.T) {
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("GOPOD_OVERLAY_CHIPPER not set")
	}
	page, _ := os.ReadFile(filepath.Join(root, "webroot/js/gopod_brobots_chat.js"))
	for _, want := range []string{"p._feedId > voiceFeedId", `genKind.indexOf("forwait:") === 0`, "bubble._feedId = feedIdNow", "voiceForText[path]", "Promise.resolve().then(voiceTieAhead)", "          voiceTieAhead();\n        }"} {
		if !strings.Contains(string(page), want) {
			t.Errorf("the page lost %q", want)
		}
	}
	src, _ := os.ReadFile(filepath.Join(root, "pkg/wirepod/ttr/gopod_reporter_voices.go"))
	if !strings.Contains(string(src), "page, seconds, ok = gopodSayForShownLine(host, text)") {
		t.Error("a voice-only line skips the voice template (gopodSayForShownLine)")
	}
	// RULE (operator 2026-10-23: "when any voiced line is created, the fucking
	// template applies!!"): a voice reaches the page only from the template.
	files, _ := filepath.Glob(filepath.Join(root, "pkg/wirepod/ttr/*.go"))
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		data, _ := os.ReadFile(file)
		n := strings.Count(string(data), "gopodWithVolumeShare(page)")
		base := filepath.Base(file)
		if n > 0 && base != "gopod_reporter_voices.go" && base != "gopod_say_heard.go" {
			t.Errorf("%s puts a voice on the page outside the voice template", base)
		}
	}
	old := gopodBrobotsSettings.Session.PauseAfterMS
	defer func() { gopodBrobotsSettings.Session.PauseAfterMS = old; GopodPageShown(true) }()
	gopodBrobotsSettings.Session.PauseAfterMS = 60
	GopodPageShown(true)
	GopodPageShown(false)
	GopodPageShown(true)
	time.Sleep(150 * time.Millisecond)
	gopodPagePauseState.Lock()
	flicker := gopodPagePauseState.paused
	gopodPagePauseState.Unlock()
	if flicker {
		t.Error("a short flicker paused the run")
	}
	GopodPageShown(false)
	time.Sleep(150 * time.Millisecond)
	gopodPagePauseState.Lock()
	held := gopodPagePauseState.paused
	gopodPagePauseState.Unlock()
	if !held {
		t.Error("a page hidden past pause_after_ms didn't pause")
	}
}

// TestOnePageOwnsTheShow (2026-10-23, FINISH_PLAN_001.md C1): one main page -
// it plays the sound and its reports time the run; every other page is a
// silent viewer whose reports are ignored; gone past the lease, the next page
// that can play takes over; a viewer by choice never does.
func TestOnePageOwnsTheShow(t *testing.T) {
	old := gopodBrobotsSettings.PageSync.MainLeaseSeconds
	defer func() {
		gopodBrobotsSettings.PageSync.MainLeaseSeconds = old
		gopodMainPage.Lock()
		gopodMainPage.id = ""
		gopodMainPage.Unlock()
	}()
	gopodBrobotsSettings.PageSync.MainLeaseSeconds = 0.1
	gopodMainPage.Lock()
	gopodMainPage.id = ""
	gopodMainPage.Unlock()
	if GopodMainPageClaim("b", false) {
		t.Error("a page that can't play claimed the show")
	}
	if !GopodMainPageClaim("a", true) || GopodMainPageClaim("b", true) {
		t.Error("the first page that can play is main, the second a viewer")
	}
	if GopodIsMainPage("b") || !GopodIsMainPage("a") || !GopodIsMainPage("") {
		t.Error("only the main page's reports (or an older page's, no id) count")
	}
	time.Sleep(150 * time.Millisecond)
	if !GopodMainPageClaim("b", true) || GopodIsMainPage("a") {
		t.Error("the main page gone past the lease: the next page takes over")
	}
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		return
	}
	page, _ := os.ReadFile(filepath.Join(root, "webroot/js/gopod_brobots_chat.js"))
	for _, want := range []string{`if (!isMainPage) { return 0; }`, `!pageRoleKnown || !isMainPage) { return; }`, `heardBody.set("page", pageId)`, `openingBody.set("page", pageId)`, `body.set("page", pageId)`} {
		if !strings.Contains(string(page), want) {
			t.Errorf("the page lost %q", want)
		}
	}
	h, _ := os.ReadFile(filepath.Join(root, "pkg/wirepod/sdkapp/gopod_typed_input.go"))
	if strings.Count(string(h), "GopodIsMainPage(r.FormValue(\"page\"))") != 3 || !strings.Contains(string(h), "GopodMainPageClaim(") {
		t.Error("Wire-Pod no longer takes the run's timing from the main page only")
	}
}

// TestStrikeAndPlacingHold (2026-10-23): listed expletives show struck through
// (said as their substitute); a picture's load wait never blocks bubble placing
// (live: every bubble after the granny photo waited until F5); every voice
// loads the moment it reaches the page.
func TestStrikeAndPlacingHold(t *testing.T) {
	if got := gopodStrikeExpletives("Comedic bullshit. Bullshitting! Bullseye."); got != "Comedic b̶u̶l̶l̶s̶h̶i̶t̶. B̶u̶l̶l̶s̶h̶i̶t̶t̶i̶n̶g̶! Bullseye." {
		t.Errorf("strike: %q", got)
	}
	if got := gopodStrikeExpletives(gopodStrikeExpletives("bullshit")); strings.Count(got, "̶") != 8 {
		t.Errorf("struck twice: %q", got)
	}
	if got := gopodApplyPronunciations("bullshitting"); got != "bull bitting" {
		t.Errorf("said: %q", got)
	}
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		return
	}
	page, _ := os.ReadFile(filepath.Join(root, "webroot/js/gopod_brobots_chat.js"))
	if strings.Contains(string(page), "revealTimer = setTimeout(check") {
		t.Error("whenBubblesReady sets revealTimer again - a picture's wait blocks every bubble after it")
	}
	if !strings.Contains(string(page), "isVoiceEvent(item.e) && !restoringChat") {
		t.Error("voices no longer load the moment they reach the page")
	}
}

// TestPhoneLayoutHolds (2026-10-23, C4): the typing row is one line or three,
// the system window is the admin's only and fills the room above the row, and
// Bot Settings reads "Vector Config".
func TestPhoneLayoutHolds(t *testing.T) {
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("GOPOD_OVERLAY_CHIPPER not set")
	}
	page, _ := os.ReadFile(filepath.Join(root, "webroot/js/gopod_brobots_chat.js"))
	for _, want := range []string{"if (!isAdminPage) { return; }", `typedRow.classList.toggle("stacked"`, `outer.style.height = room + "px"`} {
		if !strings.Contains(string(page), want) {
			t.Errorf("the page lost %q", want)
		}
	}
	css, _ := os.ReadFile(filepath.Join(root, "webroot/css/gopod_responsive.css"))
	for _, want := range []string{".gopod-typed-input.stacked .gopod-typed-text { flex: 1 1 100%; }", "#botStats {", ".brobot-bubble.system-menu-host { flex: none; }"} {
		if !strings.Contains(string(css), want) {
			t.Errorf("gopod_responsive.css lost %q", want)
		}
	}
	index, _ := os.ReadFile(filepath.Join(root, "webroot/index.html"))
	if !strings.Contains(string(index), `href="css/gopod_responsive.css"`) {
		t.Error("the page no longer loads gopod_responsive.css")
	}
	for _, f := range []string{"index.html", "settings.html"} {
		b, _ := os.ReadFile(filepath.Join(root, "webroot/sdkapp", f))
		if !strings.Contains(string(b), `<h1 class="center">Vector Config</h1>`) {
			t.Errorf("sdkapp/%s: not \"Vector Config\"", f)
		}
	}
}

// TestDisplayMirrorHolds (2026-10-24, operator: "the visitor page does no
// thinking * visitor page zombie follows GOPOD admin feed"): only the main page
// runs the chat and sends what it shows; every other page only paints it; only
// the admin sends; a visitor fetches it.
func TestDisplayMirrorHolds(t *testing.T) {
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("GOPOD_OVERLAY_CHIPPER not set")
	}
	page, _ := os.ReadFile(filepath.Join(root, "webroot/js/gopod_brobots_chat.js"))
	for _, want := range []string{"!pageRoleKnown || !isMainPage) { return; }", "function displaySend() {", "function displayPaint(d) {", `fetch("/api-sdk/gopod_display?since="`} {
		if !strings.Contains(string(page), want) {
			t.Errorf("the page lost %q", want)
		}
	}
	// The admin's template never reaches a visitor (live: the system window showed on the phone).
	for _, want := range []string{`!el.classList.contains("gopod-admin-only")`, `system-menu-host gopod-admin-only"`, `terminal gopod-admin-only"`} {
		if !strings.Contains(string(page), want) {
			t.Errorf("the page lost %q", want)
		}
	}
	for _, gone := range []string{"viewerBehind", "mainPlayed", "gopod_heard_lines", `fetch("/api/get_logs")`} {
		if strings.Contains(string(page), gone) {
			t.Errorf("the old follow code is back: %s", gone)
		}
	}
	post := func(addr string) int {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api-sdk/gopod_display", strings.NewReader(`{"html":"<div>hi</div>","anchor":0,"offset":0,"at_end":true}`))
		req.RemoteAddr = addr
		http.DefaultServeMux.ServeHTTP(rec, req)
		return rec.Code
	}
	if post("203.0.113.58:1") != http.StatusForbidden || post("127.0.0.1:1") != http.StatusOK {
		t.Error("only the admin may send the display")
	}
	rec := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rec, httptest.NewRequest("GET", "/api-sdk/gopod_display?since=0", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "<div>hi</div>") {
		t.Errorf("a visitor gets the display: %d %s", rec.Code, rec.Body.String())
	}
}

// The seven templates the master list had no lock for (2026-10-23, operator:
// "Lock tests"): each one's rules in the settings, and the page's one builder.
func gopodRawSettingsForTest(t *testing.T) (map[string]any, string) {
	t.Helper()
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("GOPOD_OVERLAY_CHIPPER not set")
	}
	data, err := os.ReadFile(filepath.Join(root, "webroot/gopod_brobots_settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	page, _ := os.ReadFile(filepath.Join(root, "webroot/js/gopod_brobots_chat.js"))
	return raw, string(page)
}

func gopodBlockForTest(t *testing.T, raw map[string]any, key string) map[string]any {
	t.Helper()
	b, ok := raw[key].(map[string]any)
	if !ok {
		t.Fatalf("settings: no %s block", key)
	}
	return b
}

// TestHeaderTemplateHolds: every bubble kind rule names one of the five kinds,
// the display names have their fallback, and the page builds a header one way.
func TestHeaderTemplateHolds(t *testing.T) {
	raw, page := gopodRawSettingsForTest(t)
	kinds := map[string]bool{"operator": true, "goverlord": true, "gomad": true, "gominion": true, "other": true}
	rules, _ := gopodBlockForTest(t, raw, "bubble_kinds")["rules"].([]any)
	if len(rules) == 0 {
		t.Error("bubble_kinds.rules is empty")
	}
	for _, r := range rules {
		m, _ := r.(map[string]any)
		has, _ := m["has"].([]any)
		if k, _ := m["kind"].(string); !kinds[k] || len(has) == 0 {
			t.Errorf("bubble_kinds rule %v: a kind of the five, and what it has", m)
		}
	}
	if fb, _ := gopodBlockForTest(t, raw, "display_names")["fallback"].(string); !strings.Contains(fb, "{n}") {
		t.Errorf("display_names.fallback %q: no {n}", fb)
	}
	if n := strings.Count(page, `head.className = "brobot-bubble-head"`); n != 1 {
		t.Errorf("the page builds a bubble header %d ways, want 1", n)
	}
}

// TestNarratorTemplateHolds: the narrator's fades are set and in order.
func TestNarratorTemplateHolds(t *testing.T) {
	raw, page := gopodRawSettingsForTest(t)
	cw := gopodBlockForTest(t, raw, "chat_window")
	in, _ := cw["narrator_fade_in_s"].(float64)
	outMax, _ := cw["narrator_fade_out_max_s"].(float64)
	outEnd, _ := cw["narrator_fade_out_end_s"].(float64)
	if in <= 0 || outMax <= 0 || outEnd < outMax {
		t.Errorf("narrator fades: in %v, out max %v, out end %v", in, outMax, outEnd)
	}
	if strings.TrimSpace(gopodBrobotsSettings.ReporterVoices.NarratorRole) == "" || !strings.Contains(page, "function showNarratorLine(") {
		t.Error("the narrator lines have no role or no builder")
	}
}

// TestCountdownAndBingoBallTemplatesHold: their roles are set, and the page
// takes them from the settings.
func TestCountdownAndBingoBallTemplatesHold(t *testing.T) {
	_, page := gopodRawSettingsForTest(t)
	if strings.TrimSpace(gopodBrobotsSettings.Countdown.Role) == "" || strings.TrimSpace(gopodBrobotsSettings.BingoBall.Role) == "" {
		t.Error("countdown.role or bingo_ball.role is empty")
	}
	for _, want := range []string{"countdownRole = settings.countdown.role", "ballRole = settings.bingo_ball.role", "function showCountdown(", "function showBingoBall("} {
		if !strings.Contains(page, want) {
			t.Errorf("the page lost %q", want)
		}
	}
}

// TestSearchResultsTemplateHolds: the results have their role, count and lines.
func TestSearchResultsTemplateHolds(t *testing.T) {
	_, page := gopodRawSettingsForTest(t)
	s := gopodBrobotsSettings.SystemSearch
	if s.ResultsRole == "" || s.Results <= 0 || s.TimeoutSeconds <= 0 || strings.TrimSpace(s.NoneNote) == "" {
		t.Errorf("system_search: role %q, results %d, timeout %v, none_note %q", s.ResultsRole, s.Results, s.TimeoutSeconds, s.NoneNote)
	}
	if !strings.Contains(page, "searchSettings = settings.system_search") || !strings.Contains(page, "function showSearchResults(") {
		t.Error("the page no longer shows search results from system_search")
	}
}

// TestSystemMenuTemplateHolds: every system- command in the menu is one
// Wire-Pod knows (system-help lists it), every item has words.
func TestSystemMenuTemplateHolds(t *testing.T) {
	raw, page := gopodRawSettingsForTest(t)
	help := gopodSystemHelp()
	groups, _ := gopodBlockForTest(t, raw, "system_menu")["groups"].([]any)
	if len(groups) == 0 {
		t.Fatal("system_menu.groups is empty")
	}
	for _, g := range groups {
		gm, _ := g.(map[string]any)
		items, _ := gm["items"].([]any)
		for _, it := range items {
			im, _ := it.(map[string]any)
			cmd, _ := im["cmd"].(string)
			if strings.TrimSpace(cmd) == "" {
				t.Errorf("system_menu %v: an item with no command", gm["title"])
				continue
			}
			word := strings.Fields(cmd)[0]
			if strings.HasPrefix(word, "system-") && !strings.Contains(help, strings.TrimRight(word, "-")) {
				t.Errorf("system_menu: %q isn't a system command (system-help)", word)
			}
		}
	}
	if !strings.Contains(page, "function systemMenuOpen(") || !strings.Contains(page, "if (!isAdminPage) { return; }") {
		t.Error("the system menu lost its builder or its admin-only rule")
	}
}

// TestQRCodeTemplateHolds: every QR code has its words, address, title and
// picture file.
func TestQRCodeTemplateHolds(t *testing.T) {
	raw, page := gopodRawSettingsForTest(t)
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	codes, _ := gopodBlockForTest(t, raw, "qr_codes")["codes"].([]any)
	if len(codes) == 0 {
		t.Fatal("qr_codes.codes is empty")
	}
	for _, c := range codes {
		m, _ := c.(map[string]any)
		words, _ := m["words"].([]any)
		url, _ := m["url"].(string)
		img, _ := m["image"].(string)
		title, _ := m["title"].(string)
		if len(words) == 0 || !strings.HasPrefix(url, "http") || title == "" || img == "" {
			t.Errorf("qr code %v: words, url, title and image", m)
			continue
		}
		if _, err := os.Stat(filepath.Join(root, "webroot", img)); err != nil {
			t.Errorf("qr code %q: no picture %s", title, img)
		}
	}
	if !strings.Contains(page, "qrCodes = settings.qr_codes.codes") {
		t.Error("the page no longer takes its QR codes from qr_codes")
	}
}

// TestChatOnlyViewHolds (2026-10-23, C2: "from the wire-pod page 'Log' button
// tap"): a visitor's Log opens the chat alone; the admin's Log is unchanged.
func TestChatOnlyViewHolds(t *testing.T) {
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("GOPOD_OVERLAY_CHIPPER not set")
	}
	page, _ := os.ReadFile(filepath.Join(root, "webroot/js/gopod_brobots_chat.js"))
	for _, want := range []string{"if (!pageRoleKnown) { chatOnlyPending = true; return; }", "if (isAdminPage) { document.body.classList.remove(\"gopod-chat-only\"); return; }", "window.showLog = function () {", `document.body.classList.add("gopod-chat-only")`} {
		if !strings.Contains(string(page), want) {
			t.Errorf("the page lost %q", want)
		}
	}
	css, _ := os.ReadFile(filepath.Join(root, "webroot/css/gopod_responsive.css"))
	if !strings.Contains(string(css), "body.gopod-chat-only .main-nav-parent") {
		t.Error("the chat-only view no longer hides Wire-Pod's settings")
	}
}

// TestOnlyGOPODsMachineIsMain (2026-10-23, live: the phone took the main page
// after a restart and the PC went silent): only a page on GOPOD's own machine is
// the admin and can be the main page.
func TestOnlyGOPODsMachineIsMain(t *testing.T) {
	for addr, want := range map[string]bool{"127.0.0.1:51000": true, "[::1]:51000": true, "203.0.113.9:51000": false, "nonsense": false} {
		if got := GopodIsLocalMachine(addr); got != want {
			t.Errorf("GopodIsLocalMachine(%q) = %v, want %v", addr, got, want)
		}
	}
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		return
	}
	h, _ := os.ReadFile(filepath.Join(root, "pkg/wirepod/sdkapp/gopod_typed_input.go"))
	if !strings.Contains(string(h), "GopodMainPageClaim(r.FormValue(\"page\"), shown && admin)") {
		t.Error("a page off GOPOD's machine can claim the main page again")
	}
	page, _ := os.ReadFile(filepath.Join(root, "webroot/js/gopod_brobots_chat.js"))
	if !strings.Contains(string(page), "isAdminPage = d.admin === true;") {
		t.Error("the page no longer takes its admin role from Wire-Pod")
	}
}

// TestVisitorSignUp (2026-10-24, operator: "allow a chat participant gathering *
// initially visitor chat participation is OFF ... Save the visitor profile in
// case of a disconnect reconnect"): a visitor's first line asks for a username
// (first come, first served; taken: name1), lines after it are dropped until
// approved; ON and ticked passes under their name; commands never; clear asks
// again; profiles come back after a restart; the admin list sets the admin.
func TestVisitorSignUp(t *testing.T) {
	dir := t.TempDir()
	oldVS := gopodBrobotsSettings.Visitors
	defer func() {
		gopodBrobotsSettings.Visitors = oldVS
		gopodVisitors.Lock()
		gopodVisitors.byAddr, gopodVisitors.on, gopodVisitors.set, gopodVisitors.loaded = map[string]*gopodVisitor{}, false, false, false
		gopodVisitors.Unlock()
	}()
	vs := &gopodBrobotsSettings.Visitors
	vs.Participation, vs.ProfilesFile, vs.AdminAddressesFile = false, filepath.Join(dir, "visitors.json"), filepath.Join(dir, "admins")
	vs.SeenSeconds, vs.NameMax, vs.WaitingText, vs.MutedText, vs.ApprovedText = 30, 20, "Waiting for approval", "Muted", "Username Approved"
	os.WriteFile(vs.AdminAddressesFile, []byte("# laptop\n203.0.113.7\n"), 0o600)
	gopodVisitors.Lock()
	gopodVisitors.byAddr, gopodVisitors.on, gopodVisitors.set, gopodVisitors.loaded = map[string]*gopodVisitor{}, false, false, false
	gopodVisitors.Unlock()
	if !GopodIsAdmin("203.0.113.7:5000") || GopodIsAdmin("203.0.113.5:5000") {
		t.Error("the admin list: the laptop is admin, another device a visitor")
	}
	phone := "203.0.113.5:5000"
	line := func(addr, id, text string) (bool, string, string) {
		return GopodVisitorLine(addr, id, "Mozilla/5.0 (Linux; Android 14) Mobile Chrome/130", text)
	}
	if pass, _, note := line(phone, "idA", "Pip"); pass || note != "Waiting for approval" {
		t.Errorf("name request: pass %v note %q", pass, note)
	}
	if pass, _, note := line(phone, "idA", "hello"); pass || note != "Waiting for approval" {
		t.Errorf("before approval: pass %v note %q", pass, note)
	}
	line("203.0.113.9:1", "idB", "pip")
	line("203.0.113.10:1", "idC", "Operator")
	list := gopodVisitorList()
	names := map[string]string{}
	for _, v := range list {
		names[v.Key] = v.Name
	}
	if names["id:idA"] != "Pip" || names["id:idB"] != "pip1" || names["id:idC"] != "Operator1" {
		t.Errorf("names first come, first served: %v", names)
	}
	gopodVisitorAct("approve", "id:idA", true)
	if pass, _, note := line(phone, "idA", "hello"); pass || note != "Username Approved" {
		t.Errorf("approved, participation OFF: pass %v note %q", pass, note)
	}
	if st := GopodVisitorStatus(phone, "idA", ""); st["state"] != "approved" {
		t.Errorf("status after approval, OFF: %v", st["state"])
	}
	gopodVisitorAct("participation", "", true)
	if pass, how, _ := line(phone, "idA", "hello Brobots"); !pass || how != "Visitor: Pip" {
		t.Errorf("approved, ON, ticked: pass %v how %q", pass, how)
	}
	if pass, _, _ := line(phone, "idA", "stop run"); pass {
		t.Error("a visitor ran a command")
	}
	gopodVisitorAct("allow", "id:idA", false)
	if pass, _, note := line(phone, "idA", "hello"); pass || note != "Muted" {
		t.Errorf("unticked: pass %v note %q", pass, note)
	}
	gopodVisitorAct("allow", "id:idA", true)
	// A reconnect on a new address is the same visitor (its browser ID).
	if pass, how, _ := line("203.0.113.66:1", "idA", "still me"); !pass || how != "Visitor: Pip" {
		t.Errorf("reconnected: pass %v how %q", pass, how)
	}
	gopodVisitorAct("clear", "id:idB", false)
	line("203.0.113.9:1", "idB", "Doc Fan")
	// A restart: the profiles come back, participation starts OFF.
	gopodVisitors.Lock()
	gopodVisitors.byAddr, gopodVisitors.on, gopodVisitors.set, gopodVisitors.loaded = map[string]*gopodVisitor{}, false, false, false
	gopodVisitors.Unlock()
	if pass, _, note := line(phone, "idA", "hello"); pass || note != "Username Approved" {
		t.Errorf("after a restart: approved Pip, participation OFF - pass %v note %q", pass, note)
	}
	back := map[string]gopodVisitor{}
	for _, v := range gopodVisitorList() {
		back[v.Key] = v
	}
	if !back["id:idA"].Approved || back["id:idB"].Name != "Doc Fan" {
		t.Errorf("profiles after a restart: %+v", back)
	}
	// Wire-Pod runs as root: "~/" must be the operator's home, never root's.
	if root := os.Getenv("GOPOD_OVERLAY_CHIPPER"); root != "" {
		src, _ := os.ReadFile(filepath.Join(root, "pkg/wirepod/ttr/gopod_visitors.go"))
		if strings.Contains(string(src), "os.UserHomeDir()") || strings.Count(string(src), "gopodHomePath(") < 2 {
			t.Error("the visitor files must use the operator's home (gopodHomePath), not root's")
		}
	}
	// Wire-Pod's own intents yes; GOPOD's commands never (2026-10-24, operator:
	// "visitors can trigger native wire-pod intents * not GOPOD admin or other
	// system level commands, like the 'thx bros'").
	for text, cmd := range map[string]bool{"what time is it": false, "what's the weather": false, "thx bros": true, "thanks Brobots": true, "system-status": true, "stop run": true} {
		if got := gopodVisitorCommand(text); got != cmd {
			t.Errorf("gopodVisitorCommand(%q) = %v, want %v", text, got, cmd)
		}
	}
	if got := gopodDeviceKind("Mozilla/5.0 (Linux; Android 14) Mobile Chrome/130"); got != "Android phone · Chrome" {
		t.Errorf("device: %q", got)
	}
}

// TestPhoneAppHolds (2026-10-24, C3): GOPOD's own certificates - an authority
// made once, a server certificate for this machine's names that chains to it and
// isn't made again while they stay the same - and the app's manifest, icon,
// worker and chat start.
func TestPhoneAppHolds(t *testing.T) {
	old := gopodBrobotsSettings.App
	defer func() { gopodBrobotsSettings.App = old }()
	gopodBrobotsSettings.App.CertDir = t.TempDir()
	gopodBrobotsSettings.App.CAName = "GOPOD test"
	crt, _, err := gopodEnsureCerts()
	if err != nil {
		t.Fatal(err)
	}
	ca, err := x509.ParseCertificate(gopodReadPEM(filepath.Join(gopodBrobotsSettings.App.CertDir, "ca.crt")))
	if err != nil || !ca.IsCA {
		t.Fatalf("the authority: %v", err)
	}
	srv, err := x509.ParseCertificate(gopodReadPEM(crt))
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(ca)
	if _, err := srv.Verify(x509.VerifyOptions{Roots: pool, DNSName: "localhost"}); err != nil {
		t.Errorf("the server certificate doesn't chain to GOPOD's authority for localhost: %v", err)
	}
	if info, _ := os.Stat(filepath.Join(gopodBrobotsSettings.App.CertDir, "ca.key")); info == nil || info.Mode().Perm() != 0o600 {
		t.Error("the authority's key isn't private (0600)")
	}
	crt2, _, _ := gopodEnsureCerts()
	srv2, _ := x509.ParseCertificate(gopodReadPEM(crt2))
	if srv2 == nil || srv2.SerialNumber.Cmp(srv.SerialNumber) != 0 {
		t.Error("the server certificate was made again with the same names")
	}
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		return
	}
	var m struct {
		StartURL string `json:"start_url"`
		Display  string `json:"display"`
		Icons    []struct {
			Src string `json:"src"`
		} `json:"icons"`
	}
	data, _ := os.ReadFile(filepath.Join(root, "webroot/app/gopod_manifest.json"))
	if err := json.Unmarshal(data, &m); err != nil || m.StartURL != "/?app" || m.Display != "standalone" || len(m.Icons) < 2 {
		t.Errorf("the app manifest: %+v %v", m, err)
	}
	for _, ic := range m.Icons {
		if _, err := os.Stat(filepath.Join(root, "webroot", ic.Src)); err != nil {
			t.Errorf("no icon %s", ic.Src)
		}
	}
	index, _ := os.ReadFile(filepath.Join(root, "webroot/index.html"))
	page, _ := os.ReadFile(filepath.Join(root, "webroot/js/gopod_brobots_chat.js"))
	if !strings.Contains(string(index), `href="/app/gopod_manifest.json"`) || !strings.Contains(string(page), `navigator.serviceWorker.register("/gopod_sw.js")`) {
		t.Error("the page lost its app manifest or its worker")
	}
	if _, err := os.Stat(filepath.Join(root, "webroot/gopod_sw.js")); err != nil {
		t.Error("no gopod_sw.js")
	}
}

// TestVisitorGateAndGreetings (2026-10-24): a visitor reaches the Brobots chat
// and nothing of Wire-Pod's ("redo visitor page with no other wire-pod
// access"); every "good <part of day>" is answered against the real time.
func TestVisitorGateAndGreetings(t *testing.T) {
	gate := func(addr, path string) int {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", path, nil)
		req.RemoteAddr = addr
		if gopodVisitorGate(rec, req) {
			return 200
		}
		return rec.Code
	}
	visitor := "203.0.113.58:5000"
	for path, want := range map[string]int{"/": 200, "/index.html": 200, "/js/gopod_brobots_chat.js": 200, "/gopod_chat_feed.json": 200,
		"/api-sdk/gopod_typed_input": 200, "/api-sdk/gopod_display": 200, "/app/gopod_manifest.json": 200,
		"/setup.html": 403, "/sdkapp/": 403, "/api/get_config": 403, "/api/get_logs": 403, "/api/set_kg_api": 403,
		"/api-sdk/gopod_record_session": 403, "/api-sdk/gopod_visitors": 403, "/session-certs/x": 403, "/css/../setup.html": 403} {
		if got := gate(visitor, path); got != want {
			t.Errorf("visitor %s: %d, want %d", path, got, want)
		}
	}
	if gate("127.0.0.1:5000", "/setup.html") != 200 || gate("127.0.0.1:5000", "/api/get_config") != 200 {
		t.Error("the admin is refused")
	}
	old := gopodBrobotsSettings.Greetings
	defer func() { gopodBrobotsSettings.Greetings = old }()
	g := &gopodBrobotsSettings.Greetings
	g.RightText, g.WrongText = "Good {part} to you too!", "Good {said}? It's {time} here - good {part}!"
	g.Parts = nil
	for _, p := range []struct {
		n    string
		f, t int
	}{{"morning", 5, 12}, {"afternoon", 12, 17}, {"evening", 17, 22}, {"night", 22, 5}} {
		g.Parts = append(g.Parts, struct {
			Name string `json:"name"`
			From int    `json:"from"`
			To   int    `json:"to"`
		}{p.n, p.f, p.t})
	}
	morning := time.Date(2026, 10, 24, 9, 25, 0, 0, time.Local)
	if got := gopodGreetingLine("", "good morning Brobots", morning); got != "Good morning to you too!" {
		t.Errorf("good morning at 9:25: %q", got)
	}
	if got := gopodGreetingLine("", "it's a good evening", morning); got != "Good evening? It's "+gopodClockFor("", morning)+" here - good morning!" {
		t.Errorf("good evening at 9:25: %q", got)
	}
	if got := gopodGreetingLine("", "Good night bros", time.Date(2026, 10, 24, 23, 10, 0, 0, time.Local)); got != "Good night to you too!" {
		t.Errorf("good night at 23:10: %q", got)
	}
	if gopodGreetingLine("", "good job", morning) != "" || gopodGreetingSaid("evening greeting") != "" {
		t.Error("not a greeting, yet answered as one")
	}
}

// TestHeardLineAndFreeCap (2026-10-24): a line a Brobot heard himself goes in the
// chat at once, the Operator's "[Backpack]" bubble, to the Brobot it names first
// (live: it showed under his answer); a free answer has its own cap.
func TestHeardLineAndFreeCap(t *testing.T) {
	gopodChatFeed.Lock()
	before := len(gopodChatFeed.events)
	gopodChatFeed.Unlock()
	gopodHeardLine(gopodBrobot1ESN, "Brobot 2, good afternoon")
	gopodChatFeed.Lock()
	events := append([]gopodChatFeedEvent(nil), gopodChatFeed.events[before:]...)
	gopodChatFeed.Unlock()
	if len(events) != 1 || events[0].Who != "Operator" || events[0].To != "Brobot 2" || events[0].Via != "Backpack" {
		t.Errorf("heard line: %+v", events)
	}
	if gopodBrobotsSettings.ReplyTighten.FreeMaxSentences <= 0 {
		t.Error("reply_tighten.free_max_sentences is not set")
	}
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		return
	}
	src, _ := os.ReadFile(filepath.Join(root, "pkg/wirepod/ttr/matchIntentSend.go"))
	if !strings.Contains(string(src), "gopodHeardLine(botSerial, voiceText)") {
		t.Error("heard lines no longer go in the chat at once")
	}
}

// TestSpeakerAndSystemWindowHold (2026-10-24, operator: "speaker icon ... starts
// OFF * When speaker icon is OFF, no volume bar * In system window, all slots
// start as closed * resize system menu bubble to show all slots"): the speaker
// starts as background_loop.starts_on says, never from the browser; no volume
// bar while it's off; every slot closed; the menu as tall as its slots.
func TestSpeakerAndSystemWindowHold(t *testing.T) {
	raw, page := gopodRawSettingsForTest(t)
	if on, ok := gopodBlockForTest(t, raw, "background_loop")["starts_on"].(bool); !ok || on {
		t.Error("background_loop.starts_on must be false (the speaker starts off)")
	}
	for _, want := range []string{`speakerOn = cfg.starts_on === true;`, `if (!backgroundLoop.speaker()) { volumeWrap.classList.remove("open"); return; }`} {
		if !strings.Contains(page, want) {
			t.Errorf("the page lost %q", want)
		}
	}
	for _, gone := range []string{`localStorage.getItem("gopodLoopSpeaker")`, `if (gi === 0) { group.open = true; }`, `menu.style.flex = "0 0 70%"`} {
		if strings.Contains(page, gone) {
			t.Errorf("back again: %s", gone)
		}
	}
	css, _ := os.ReadFile(filepath.Join(os.Getenv("GOPOD_OVERLAY_CHIPPER"), "webroot/css/gopod_chat_bubbles.css"))
	if !strings.Contains(string(css), "flex: 0 1 auto; /* 2026-10-24") {
		t.Error("the system menu is no longer as tall as its slots")
	}
}

// TestMicVoiceAndVisitorRowHold (2026-10-24): a numpad line plays its own mic
// recording (evened, timed by the worker's Vosk "align" job) through the voice
// template; the speech's fallback is what the bubble shows; a typed answer to
// Goverlord is read out like every Operator line; a visitor's chat opens without
// Wire-Pod's log; the marquee is each page's own; doin' is said dew-in.
func TestMicVoiceAndVisitorRowHold(t *testing.T) {
	file, page := gopodMicVoicePaths("/tmp/kp0_1.wav", "good morning")
	if !strings.HasPrefix(filepath.Base(file), "op_mic_") || !regexp.MustCompile(`^/gopod_reporter_audio/[a-z]{2}_[a-z]+_[0-9a-f]{16}\.wav$`).MatchString(page) {
		t.Errorf("mic voice name %q must look like a made voice (the page reads its word times)", page)
	}
	gopodMicVoiceRemember("Good morning", "/nonexistent/kp0.wav")
	if _, ok := gopodMicVoiceFor.Load("good morning"); ok {
		t.Error("a missing recording was remembered")
	}
	if got := gopodApplyPronunciations("doin' fine"); !strings.Contains(got, "dew") {
		t.Errorf("doin': %q", got)
	}
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		return
	}
	read := func(rel string) string { b, _ := os.ReadFile(filepath.Join(root, rel)); return string(b) }
	checks := map[string][]string{
		"pkg/wirepod/ttr/gopod_numpad_ptt.go":               {"gopodMicVoiceRemember(text, master)"},
		"pkg/wirepod/ttr/gopod_reporter_voices.go":          {"if gopodMicVoiceSay(host, text) {", "func gopodVoicePlay("},
		"gopod_probes/tools/gopod_reporter_voice_worker.py": {`if job.get("align"):`},
		"pkg/wirepod/ttr/kgsim.go":                          {"if streamResult.Fallback {"},
		"pkg/wirepod/ttr/gopod_record_session.go":           {"gopodReporterReadsOperator(text, how, to)"},
		"webroot/js/gopod_brobots_chat.js":                  {"function visitorOpenChat() {", "function marqueeTrack(bubble, layer) {", `l.textContent = "";`, `typedSubmit.type = "submit";`},
		"webroot/css/gopod_responsive.css":                  {"body.gopod-visitor .gopod-typed-input .gopod-typed-submit {", "body.gopod-visitor .brobot-bubble { font-size: 15px; }", ".narrator-single .narrator-stage { aspect-ratio: auto;"},
	}
	for rel, wants := range checks {
		src := read(rel)
		for _, w := range wants {
			if !strings.Contains(src, w) {
				t.Errorf("%s lost %q", rel, w)
			}
		}
	}
}

// TestGameSweep (2026-10-24, operator: "Test every: Topic, Value Point, Easter
// Egg"): every topic found by its words, every value point picked by its own
// question, every Easter egg reachable - on every deploy (system-test-game live).
func TestGameSweep(t *testing.T) {
	if os.Getenv("GOPOD_OVERLAY_CHIPPER") == "" {
		t.Skip("GOPOD_OVERLAY_CHIPPER not set")
	}
	defer gopodAllPointsForTest()()
	t.Setenv("GOPOD_SESSION_DIR", t.TempDir()) // a fresh chat memory, as at a start
	ok, line := gopodGameSweepLine(gopodGameSweep())
	t.Log("GAME_SWEEP " + line)
	if !ok {
		t.Error(line)
	}
	src, _ := os.ReadFile(filepath.Join(os.Getenv("GOPOD_OVERLAY_CHIPPER"), "pkg/wirepod/ttr/gopod_record_session.go"))
	if !strings.Contains(string(src), "GopodChatFeedAddLive(esn, text)") || !strings.Contains(string(src), "gopodWaitReplyShown(esn)") {
		t.Error("the ending's Brobot lines no longer wait for their bubble (not lit)")
	}
}
