package wirepod_ttr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A song line is said one sentence at a time, and each sentence is written to
// the speaking-now file just before it is said; the file ends inactive.
func TestGopodSpeakSentencesWritesEachSentence(t *testing.T) {
	saved := gopodSpeakingNowPath
	defer func() { gopodSpeakingNowPath = saved }()
	gopodSpeakingNowPath = filepath.Join(t.TempDir(), "speaking.json")
	read := func() gopodSpeakingNow {
		var now gopodSpeakingNow
		data, _ := os.ReadFile(gopodSpeakingNowPath)
		_ = json.Unmarshal(data, &now)
		return now
	}
	var said, seen []string
	GopodSpeakSentences("ESN01", "Excuse Me? Bingo Test Continues. Round 3! Juggling big shiny bingo balls. Please wait.", -1, func(sentence string) {
		now := read()
		if !now.Active || now.ESN != "ESN01" {
			t.Fatalf("not marked active before saying %q: %+v", sentence, now)
		}
		seen = append(seen, now.Text)
		said = append(said, sentence)
	})
	if len(said) < 2 {
		t.Fatalf("want the line split into sentences, got %q", said)
	}
	for i := range said {
		if said[i] != seen[i] {
			t.Fatalf("sentence %d said %q but file had %q", i, said[i], seen[i])
		}
	}
	if read().Active {
		t.Fatal("file still active after the line")
	}
}

// A scripted song line arrives split by the song engine (segment marker); the
// robot says exactly those pieces, through the say_text path's own cleanup.
func TestScriptedPiecesAreSaidAsSent(t *testing.T) {
	m := gopodBrobotsSettings.ScriptedSplit.SegmentMarker
	acts := GetActionsFromString("{{playAnimationWI||thinking}}" + gopodAnimationLeadIn + " Bingo Test Begins." + m + "Round 1!" + m + "Ah... I like to party. Lol!")
	if len(acts) != 2 {
		t.Fatalf("actions = %+v", acts)
	}
	got := gopodSpeakPieces(acts[1].Parameter)
	want := []string{gopodAnimationLeadIn + " Bingo Test Begins.", "Round 1!", "Ah, I like to party. Lol!"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("pieces = %q, want %q", got, want)
	}
	if got := gopodSpeakPieces("Plain reply. With no marker in it at all, so the sentence rule applies."); len(got) != 3 {
		t.Fatalf("unmarked text = %q", got)
	}
}

// The split (operator 2026-10-06: "aim for a hard split between sentences,
// even if punctuation is missing ... within each sentence, split by comma ... a
// sentence can be one word ... forget minimum word rules"): sentences first,
// then every comma inside each one; no length rules.
func TestGopodSentencesSimpleSplit(t *testing.T) {
	for in, want := range map[string][]string{
		"... Dude, pizza! No, really, NO. bro I'm all about flavor, and pizza is the KING! What's not to love?": {
			"... Dude,", "pizza!", "No,", "really,", "NO.", "bro I'm all about flavor,", "and pizza is the KING!", "What's not to love?"},
		"Wow! That's it.":                                           {"Wow!", "That's it."},
		"Wait... I-9? Hold on, bro.":                                {"Wait...", "I-9?", "Hold on,", "bro."},
		"first line no stop\nsecond line, with a comma":             {"first line no stop", "second line,", "with a comma"},
		"Your move. Listen up! We're one robot crew, like a party.": {"Your move.", "Listen up!", "We're one robot crew,", "like a party."},
	} {
		if got := gopodSentences(in); strings.Join(got, "|") != strings.Join(want, "|") {
			t.Errorf("%q -> %q, want %q", in, got, want)
		}
	}
	// A comma piece keeps its comma (said as a pause); a hanging dash is a stop.
	if end := gopodEndWithPunctuation("Okay,"); end != "Okay," {
		t.Fatalf("comma piece said as %q", end)
	}
	if end := gopodEndWithPunctuation("with the right tools-"); end != "with the right tools." {
		t.Fatalf("dash end = %q", end)
	}
	// A missing stop before a capital is put in during clean-up.
	if fixed := gopodPrepareReply("spread good vibes   This ain't it."); fixed != "spread good vibes. This ain't it." {
		t.Fatalf("missing stop = %q", fixed)
	}
}

// An asterisk between two words leaves a space when speech drops it
// ("fun!*Giggles merrily*" was said "fun!Giggles", live 2026-10-04 20:41).
func TestRemoveSpecialCharactersAsteriskKeepsWordGap(t *testing.T) {
	got := removeSpecialCharacters("Who needs a deep explanation, just a little fun!*Giggles merrily*")
	if !strings.Contains(got, "fun! Giggles merrily") {
		t.Fatalf("got %q", got)
	}
	if got := removeSpecialCharacters("this is *really* fun"); got != "this is really fun" {
		t.Fatalf("emphasis: got %q", got)
	}
}

// An emoji used as a sentence ending splits there (the page's splitSentences
// does the same).
func TestGopodSentencesSplitAtEmojiEnding(t *testing.T) {
	said := removeSpecialCharacters("We keep the party going in each and every sentence we say 😉 You know the vibe is real, bro, all night long!")
	got := gopodSentences(said)
	if len(got) < 2 || !strings.HasSuffix(got[0], "sentence we say.") || !strings.HasPrefix(got[1], "You know") {
		t.Fatalf("got %q", got)
	}
}

// 2026-10-08: joined runs break at a sentence's end when the sentence fits
// (live 20:24: "Squawka-doodle. Doc," / "or Doc-B-S if you're in a hurry.").
func TestJoinedPiecesBreakAtSentenceEnds(t *testing.T) {
	saved := gopodBrobotsSettings.LiveSplit.JoinUpToWords
	defer func() { gopodBrobotsSettings.LiveSplit.JoinUpToWords = saved }()
	gopodBrobotsSettings.LiveSplit.JoinUpToWords = 14
	in := "Doc is Doctrine Barfallonyou, Squawkadoodle. Doc, or DocBS if you're in a hurry. Gruff, sure of himself, and he swings at delay, committee brain, and fake caution. Never at people."
	want := []string{
		"Doc is Doctrine Barfallonyou, Squawkadoodle. Doc, or DocBS if you're in a hurry.",
		"Gruff, sure of himself, and he swings at delay, committee brain, and fake caution.",
		"Never at people.",
	}
	var got []string
	for _, a := range gopodSplitSpeechIntoSentences([]RobotAction{{Action: ActionSayText, Parameter: in}}) {
		got = append(got, a.Parameter)
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

// 2026-10-08: the record-cut draft loads with reporter lines among question 7's
// answers (a reporter's label as "brobot").
func TestDraftLoadsWithReporterLines(t *testing.T) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, "gopod_tts/sessions/harvest/2026-10-08/1300_record_cut_best_takes/draft_performance.json")
	if _, err := os.Stat(path); err != nil {
		t.Skip("no record-cut draft on this machine")
	}
	t.Setenv("GOPOD_DRAFT", path)
	got, draft, err := gopodLoadDraft()
	if err != nil {
		t.Fatal(err)
	}
	// 2026-10-09: the script beside it is what plays, when it's there.
	if _, err := os.Stat(filepath.Join(filepath.Dir(path), gopodDraftScriptName)); err == nil && filepath.Base(got) != gopodDraftScriptName {
		t.Errorf("played %s, not the script beside it", got)
	}
	if len(draft.Questions) != 10 {
		t.Errorf("%d questions, want 10", len(draft.Questions))
	}
}

// A reply's "(Cue word: thinking)" note is dropped before the split (operator
// 2026-10-11: "messes up highlighting splits").
func TestCueWordNotesDropped(t *testing.T) {
	got := gopodStripCueWordNotes("It's overcast. (Cue word: thinking) Like, who cares? (cue words: love)")
	if got != "It's overcast.  Like, who cares?" && strings.Contains(strings.ToLower(got), "cue word") {
		t.Fatalf("note left in: %q", got)
	}
	if strings.Contains(strings.ToLower(got), "cue") {
		t.Fatalf("note left in: %q", got)
	}
}

// "verySad, frustrated, angry__ Haha" opens with one cue, not the list as words.
func TestCueListKeepsLastCue(t *testing.T) {
	if got := gopodDisplayReply("verySad, frustrated, angry__ Haha, that's a good one!"); strings.Contains(got, "verySad") || strings.Contains(got, "angry") || !strings.HasPrefix(got, "Haha") {
		t.Fatalf("display = %q", got)
	}
}

// GOLDEN (2026-10-10, a combo step, operator: "Should be combined!!"): one
// bubble, its parts said apart. A later part ("... Lol!") lights its own piece
// of the bubble (piece_offset), and the highlight is released after it.
func TestComboPartLightsItsPieceThenReleases(t *testing.T) {
	saved := gopodSpeakingNowPath
	defer func() { gopodSpeakingNowPath = saved }()
	gopodSpeakingNowPath = filepath.Join(t.TempDir(), "speaking.json")
	read := func() gopodSpeakingNow {
		var now gopodSpeakingNow
		data, _ := os.ReadFile(gopodSpeakingNowPath)
		_ = json.Unmarshal(data, &now)
		return now
	}
	var pieces []int
	GopodSpeakSentences("ESN02", gopodAnimationLeadIn+" Lol!", 5, func(string) { pieces = append(pieces, read().Piece) })
	if len(pieces) != 1 || pieces[0] != 5 {
		t.Fatalf("pieces lit = %v, want [5]", pieces)
	}
	if read().Active {
		t.Fatal("the highlight wasn't released after the part")
	}
}

// 2026-10-14 (live: "‍ I'll tell you" was said "yirl"): a joined emoji
// leaves no joiner or style mark behind in what the robot says.
func TestEmojiJoinersNotSpoken(t *testing.T) {
	got := removeSpecialCharacters("We're just gonna give it up on that front, bro. 🙅‍♂️ I'll tell you the place.")
	if strings.ContainsAny(got, "‍️︎") {
		t.Fatalf("joiner left in %q", got)
	}
}
