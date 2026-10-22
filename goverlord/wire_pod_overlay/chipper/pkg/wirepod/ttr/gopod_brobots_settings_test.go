package wirepod_ttr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
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
	if !strings.Contains(text, "Covered ") || !strings.Contains(detail, "doc_and_pip=") || strings.Contains(text, "{") {
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
	gopodCoverage.Lock()
	saved := gopodCoverage.revealed
	gopodCoverage.revealed = map[string]int{}
	gopodCoverage.reminded = map[string]bool{}
	gopodCoverage.Unlock()
	defer func() { gopodCoverage.Lock(); gopodCoverage.revealed = saved; gopodCoverage.Unlock() }()
	say := gopodBrobotsSettings.ValuePoints.SayLines
	ask := "Brobot 1, what's Doc's full name?"
	gopodQuestionFactsFor(gopodBrobot1ESN, ask)
	if got := gopodScriptedSnippet(gopodBrobot1ESN, ask, "start", ""); got != say["doc_and_pip 1"] {
		t.Errorf("point 1: %q", got)
	}
	if got := gopodSayLine(gopodBrobot2ESN); got != "" {
		t.Errorf("the other Brobot got a say line: %q", got)
	}
	gopodQuestionFactsFor(gopodBrobot1ESN, ask)
	if got := gopodSayLine(gopodBrobot1ESN); got != say["doc_and_pip 2"] {
		t.Errorf("point 2: %q", got)
	}
	gopodQuestionFactsFor(gopodBrobot1ESN, ask)
	if got := gopodSayLine(gopodBrobot1ESN); got != say["doc_and_pip 3"] {
		t.Errorf("point 3: %q", got)
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
	line := gopodBrobotsSettings.ValuePoints.SayLines["legacy 4"]
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
