// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_opener_song_test.go
package wirepod_ttr

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestOpenerSongPlaysLikeBuiltIn (2026-10-16, operator: "make the opener LLM
// Warming song from it, and prove it plays exactly like today's opener"): the
// opener song (00_llm_warming) plays the built-in opener's chat lines, voices
// and stings, in the same order, given the same words.
func TestOpenerSongPlaysLikeBuiltIn(t *testing.T) {
	steps, err := gopodOpenerSong()
	if err != nil || len(steps) == 0 {
		t.Fatalf("the opener song: %v (%d steps)", err, len(steps))
	}
	savedFeed, savedState := gopodChatFeedPath, gopodLLMWarmupStatePath
	savedMade, savedWait, savedOpening := gopodWarmupVoicesMade, gopodWarmupVoicesWait, gopodWarmupOpening
	savedLLM := gopodBrobotsSettings.LLM
	savedMax, savedLag := gopodBrobotsSettings.ReporterVoices.MaxWaitSeconds, gopodBrobotsSettings.ReporterVoices.PageLagSeconds
	defer func() {
		gopodChatFeedPath, gopodLLMWarmupStatePath = savedFeed, savedState
		gopodWarmupVoicesMade, gopodWarmupVoicesWait, gopodWarmupOpening = savedMade, savedWait, savedOpening
		gopodBrobotsSettings.LLM = savedLLM
		gopodBrobotsSettings.ReporterVoices.MaxWaitSeconds, gopodBrobotsSettings.ReporterVoices.PageLagSeconds = savedMax, savedLag
	}()
	gopodBrobotsSettings.ReporterVoices.MaxWaitSeconds, gopodBrobotsSettings.ReporterVoices.PageLagSeconds = 0.01, 0
	dir := t.TempDir()
	gopodChatFeedPath = filepath.Join(dir, "feed.json")
	gopodLLMWarmupStatePath = filepath.Join(dir, "warmup.json")
	gopodWarmupVoicesMade = func() bool { return true }
	gopodWarmupVoicesWait = func() {}
	gopodWarmupOpening = func() { GopodChatFeedAdd("Brobot 3", "opening line") }
	play := func(steps []gopodOpenerStep) []string {
		gopodChatFeed.Lock()
		gopodChatFeed.next, gopodChatFeed.events = 0, nil
		gopodChatFeed.Unlock()
		gopodWarmingEnd()
		warmed := make(chan struct{})
		close(warmed)
		gopodPlayOpener(steps, warmed)
		gopodWarmingEnd()
		gopodChatFeed.Lock()
		defer gopodChatFeed.Unlock()
		var out []string
		for _, e := range gopodChatFeed.events {
			text := e.Text
			// "warmup_text|<ms>|<text>" -> "warmup_text|<text>"
			if parts := strings.SplitN(text, "|", 3); len(parts) == 3 {
				text = parts[0] + "|" + parts[2]
			}
			// The page's heard check carries its own number.
			if hm := gopodBrobotsSettings.ReporterVoices.HeardMarker; hm != "" && strings.HasPrefix(text, hm) {
				text = hm
			}
			out = append(out, e.Who+": "+text)
		}
		return out
	}
	song := play(steps)
	// The built-in opener with the song's words.
	var texts []string
	for _, s := range steps {
		if s.Note == "warm_message" || s.Note == "warm_ready" {
			texts = append(texts, s.Text)
		}
	}
	if len(texts) != 3 {
		// The song has grown past the built-in opener (more steps): nothing to compare.
		t.Skipf("the opener song has %d messages, the built-in opener 3", len(texts))
	}
	gopodBrobotsSettings.LLM.VoicesText, gopodBrobotsSettings.LLM.WarmupText, gopodBrobotsSettings.LLM.ReadyText = texts[0], texts[1], texts[2]
	gopodBrobotsSettings.LLM.WarmupDoneText = ""
	builtIn := play(gopodBuiltInOpener())
	if strings.Join(song, "\n") != strings.Join(builtIn, "\n") {
		t.Errorf("the song:\n  %s\nthe built-in opener:\n  %s", strings.Join(song, "\n  "), strings.Join(builtIn, "\n  "))
	}
	if len(song) < 4 {
		t.Errorf("the song played only %d lines", len(song))
	}
}

// TestOpenerSkipsEmptyLines (2026-10-17, operator: "a single in-between
// message = \"\" = the indication to skip"): reporter_lines with no words plays nothing.
func TestOpenerSkipsEmptyLines(t *testing.T) {
	saved := gopodWarmupOpening
	defer func() { gopodWarmupOpening = saved }()
	played := 0
	gopodWarmupOpening = func() { played++ }
	gopodPlayOpener([]gopodOpenerStep{{Note: "reporter_lines"}, {Note: "reporter_lines", Text: "opening"}}, nil)
	if played != 1 {
		t.Errorf("played the opening %d times, want 1", played)
	}
}

// TestLLMWarmTemplate: the template reads as an LLM warm song, and a "_" folder
// is never a song.
func TestLLMWarmTemplate(t *testing.T) {
	steps, err := gopodReadOpenerSong(filepath.Join(gopodSongsDir(), "_new_song", "_llm_warm"))
	if err != nil || len(steps) != 6 {
		t.Fatalf("the template: %v (%d steps)", err, len(steps))
	}
	songs, err := gopodAllSongs()
	if err != nil {
		t.Fatal(err)
	}
	for id, dir := range songs {
		if strings.HasPrefix(filepath.Base(dir), "_") {
			t.Errorf("%s (%s) is a template, not a song", id, filepath.Base(dir))
		}
	}
}

// TestOpenerStepVolume (2026-10-16, operator: "add volume, default or %age"):
// a step's volume rides on its voices and stings as the page's share.
func TestOpenerStepVolume(t *testing.T) {
	for raw, want := range map[string]float64{``: 1, `""`: 1, `"default"`: 1, `"50%"`: 0.5, `"150%"`: 1.5, `"0%"`: 0, `0.4`: 0.4} {
		if got, err := gopodVolumePercent([]byte(raw)); err != nil || got != want {
			t.Errorf("%s: %v %v, want %v", raw, got, err, want)
		}
	}
	if _, err := gopodVolumePercent([]byte(`"loud"`)); err == nil {
		t.Error("\"loud\" read as a volume")
	}
	defer gopodSetVolumeShare(1)
	gopodSetVolumeShare(0.5)
	if got := gopodWithVolumeShare("/a/x.wav"); got != "/a/x.wav#vol=0.5" {
		t.Errorf("got %q", got)
	}
	if got := gopodWithVolumeShare("/a/grunge.wav#vol=0.4"); got != "/a/grunge.wav#vol=0.2" {
		t.Errorf("got %q", got)
	}
	gopodSetVolumeShare(1)
	if got := gopodWithVolumeShare("/a/x.wav"); got != "/a/x.wav" {
		t.Errorf("default: got %q", got)
	}
}

// ## EOF
