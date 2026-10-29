// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_page_template_test.go
package wirepod_ttr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestPageRunsFromSettings (2026-10-29, operator: "decouple dumb code from flexible
// data & bridge the two by loading external already templated files ... then code
// blindly runs templates from loaded files"): the chat page keeps no value of its
// own for what the settings file holds - every one is in the settings, and none is
// written into the page as well.
func TestPageRunsFromSettings(t *testing.T) {
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("no overlay here")
	}
	data, _ := os.ReadFile(filepath.Join(root, "webroot/gopod_brobots_settings.json"))
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		t.Fatal(err)
	}
	s := map[string]map[string]any{}
	for k, raw := range top {
		var block map[string]any
		if json.Unmarshal(raw, &block) == nil {
			s[k] = block
		}
	}
	// block.key: may it be empty?
	need := map[string]bool{
		"chat_window.max_bubbles": false, "chat_window.max_seen": false, "chat_window.baton_max_ms": false,
		"chat_window.speaking_max_ms": false, "chat_window.bubble_ready_max_ms": false, "chat_window.reply_delay_ms": false,
		"chat_window.top_gap_px": false, "chat_window.scroll_ms": false, "chat_window.focus_dead_zone": false,
		"chat_window.countdown_hold_ms": false, "chat_window.countdown_hold_most_ms": false,
		"chat_window.marquee_spacing_px": false, "chat_window.marquee_inset_px": false, "chat_window.marquee_pattern_ms": false,
		"chat_window.narrator_fade_in_s": false, "chat_window.narrator_fade_out_max_s": false, "chat_window.narrator_fade_out_end_s": false,
		"chat_window.voice_ready_max_ms": false, "chat_window.reveal_rest_ms": false, "chat_window.typed_text_min_px": false,
		"chat_window.hourglass_sprites": false, "page_sync.display_bubbles": false, "page_sync.display_ms": false,
		"countdown.role": false, "bingo_ball.role": false, "system_bubble.label": false, "interview_chapters.role": false,
		"interview_chapters.label": false, "value_points.score_role": true, "sign_off.system_role": false, "system_set.role": false,
		"topic_pictures.role": false, "topic_pictures.egg_zoom_text": true, "topic_pictures.popup_text": false,
		"math.popup_title": false, "reporter_voices.role": false, "reporter_voices.song_audio_role": false,
		"reporter_voices.narrator_role": false, "live_weather.military_label": false,
		"page_text.typed_placeholder": false, "page_text.submit": false, "page_text.all_first_title": false,
		"page_text.drag_title": false, "page_text.voices_title": false, "page_text.voices_unassigned": false,
		"page_text.usb_title": false, "page_text.no_answer": false, "page_text.visitors_title": false,
		"page_text.visitors_count": false, "page_text.visitors_none": false, "page_text.visitor_asks": false,
		"page_text.visitor_nameless": false, "system_menu.copy": false,
	}
	for path, emptyOK := range need {
		block, key, _ := strings.Cut(path, ".")
		v, ok := s[block][key]
		if !ok || v == nil || (!emptyOK && v == "") {
			t.Errorf("the settings have no %s - the page takes it from there", path)
		}
	}
	page, _ := os.ReadFile(filepath.Join(root, "webroot/js/gopod_brobots_chat.js"))
	for _, name := range []string{"CHAT_MAX_BUBBLES", "CHAT_MAX_SEEN", "TYPED_TEXT_MIN_PX", "CHAT_BATON_MAX_MS", "SPEAKING_MAX_MS",
		"BUBBLE_READY_MAX_MS", "CHAT_REPLY_DELAY_MS", "CHAT_TOP_GAP_PX", "CHAT_SCROLL_MS", "FOCUS_DEAD_ZONE", "COUNTDOWN_HOLD_MS",
		"COUNTDOWN_HOLD_MOST_MS", "MARQUEE_SPACING_PX", "MARQUEE_INSET_PX", "MARQUEE_PATTERN_MS", "NARRATOR_FADE_IN_S",
		"NARRATOR_FADE_OUT_MAX_S", "NARRATOR_FADE_OUT_END_S", "DISPLAY_BUBBLES", "DISPLAY_MS", "VOICE_READY_MAX_MS",
		"REVEAL_REST_MS", "HOURGLASS_SPRITES", "countdownRole", "ballRole", "systemLabel", "chapterRole", "chapterLabel",
		"scoreRole", "systemRole", "terminalRole", "pictureRole", "eggZoomText", "popupText", "mathPopupTitle", "reporterRole",
		"songAudioRole", "narratorRole", "militaryLabel"} {
		if regexp.MustCompile(`\bvar [^;]*\b` + name + ` = ("|[0-9])`).Match(page) {
			t.Errorf("the page writes its own %s - it takes it from the settings", name)
		}
	}
}

// TestLookIsTokens (2026-10-29, pass 2 of FLEXIBLE_DATA_SURVEY_001.md): every colour
// the GOPOD pages use is a look token, set once in a :root block; no rule types a
// colour of its own.
func TestLookIsTokens(t *testing.T) {
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("no overlay here")
	}
	strip := regexp.MustCompile(`(?s)/\*.*?\*/|:root\s*\{[^}]*\}`)
	colour := regexp.MustCompile(`#[0-9a-fA-F]{3,8}\b|rgba?\(`)
	for _, f := range []string{"gopod_chat_bubbles.css", "gopod_responsive.css", "gopod_weather_kit.css", "gopod_session_page.css"} {
		data, err := os.ReadFile(filepath.Join(root, "webroot/css", f))
		if err != nil {
			t.Fatal(err)
		}
		if m := colour.FindString(strip.ReplaceAllString(string(data), "")); m != "" {
			t.Errorf("%s types a colour in a rule (%s) - make it a look token in :root", f, m)
		}
	}
}

// TestSystemTextKeys (2026-10-29, pass 3 of FLEXIBLE_DATA_SURVEY_001.md): every
// system reply the code asks for (gopodSystemText) is in system_text, and its
// {names} are the ones the code fills - none missing, none left over.
func TestSystemTextKeys(t *testing.T) {
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("no overlay here")
	}
	call := regexp.MustCompile(`gopodSystemText\("([^"]+)"(.*)$`)
	names := regexp.MustCompile(`, "([a-z0-9_]+)", `)
	holes := regexp.MustCompile(`\{([a-z0-9_]+)\}`)
	files, _ := filepath.Glob(filepath.Join(root, "pkg/wirepod/ttr/*.go"))
	n := 0
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") || strings.HasSuffix(f, "gopod_system_text.go") {
			continue
		}
		data, _ := os.ReadFile(f)
		for _, line := range strings.Split(string(data), "\n") {
			m := call.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			n++
			tpl, ok := gopodBrobotsSettings.SystemText[m[1]]
			if !ok {
				t.Errorf("%s: system_text has no %s", filepath.Base(f), m[1])
				continue
			}
			given := map[string]bool{}
			for _, g := range names.FindAllStringSubmatch(m[2], -1) {
				given[g[1]] = true
			}
			used := map[string]bool{}
			for _, h := range holes.FindAllStringSubmatch(tpl, -1) {
				if !given[h[1]] {
					t.Errorf("system_text.%s has {%s} - the code doesn't fill it", m[1], h[1])
				}
				used[h[1]] = true
			}
			for g := range given {
				if !used[g] {
					t.Errorf("system_text.%s has no {%s} - the code fills it", m[1], g)
				}
			}
		}
	}
	if n < 100 {
		t.Errorf("only %d system replies read from system_text - the test no longer finds them", n)
	}
}

// TestTimingsInSettings (2026-10-29, pass 4 of FLEXIBLE_DATA_SURVEY_001.md): every
// show timing the code asks for (gopodTiming) is in the settings' timings block,
// more than 0.
func TestTimingsInSettings(t *testing.T) {
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("no overlay here")
	}
	call := regexp.MustCompile(`gopodTiming\("([a-z0-9_]+)"\)`)
	files, _ := filepath.Glob(filepath.Join(root, "pkg/wirepod/ttr/*.go"))
	n := 0
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, _ := os.ReadFile(f)
		for _, m := range call.FindAllStringSubmatch(string(data), -1) {
			n++
			if gopodBrobotsSettings.Timings[m[1]] <= 0 {
				t.Errorf("%s: timings.%s is missing or not more than 0", filepath.Base(f), m[1])
			}
		}
	}
	if n < 20 {
		t.Errorf("only %d timings read from the settings - the test no longer finds them", n)
	}
}

// ## EOF
