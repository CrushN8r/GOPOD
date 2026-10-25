// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_say_heard.go
package wirepod_ttr

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// GOLDEN rule: a bubble is not its spoken words (operator 2026-10-09: "please
// understand/make a fucking rule: a bubble is not the same as the fucking
// expected contents of the bubble!!"; "Do not go forward if an expected sleep is
// not fucking spoken!!"). A song's system line ("Sleep.") is made into its
// voice file first; then its bubble, its voice and a heard marker go into the
// page's queue together, so the bubble shows with the voice. The page tells
// Wire-Pod when it reaches the marker - the voice has played - and only then
// does the song go on. Not heard (no page playing, no voice file) is a "no".

// reporter_voices.heard_marker: a voice-queue line "<marker><token>"; the page
// answers it on /api-sdk/gopod_voice_heard instead of playing it.
var gopodSayHeard = struct {
	sync.Mutex
	waiting map[string]chan struct{}
}{waiting: map[string]chan struct{}{}}

// gopodWaitHeard waits up to most for the page to have played every voice line
// queued before now (the heard marker); true once heard.
func gopodWaitHeard(most time.Duration) bool {
	rv := gopodBrobotsSettings.ReporterVoices
	if most <= 0 || rv.HeardMarker == "" {
		return false
	}
	token := fmt.Sprintf("%d", time.Now().UnixNano())
	heard := make(chan struct{})
	gopodSayHeard.Lock()
	gopodSayHeard.waiting[token] = heard
	gopodSayHeard.Unlock()
	defer func() {
		gopodSayHeard.Lock()
		delete(gopodSayHeard.waiting, token)
		gopodSayHeard.Unlock()
	}()
	GopodChatFeedAdd(rv.Role, rv.HeardMarker+token)
	select {
	case <-heard:
		return true
	case <-time.After(most):
		return false
	}
}

// GopodSystemSayHeard says text in Goverlord's voice with its bubble and, when
// most > 0, waits up to most for the page to have played it. It returns true
// once heard (or at once when not waiting and the line was queued).
func GopodSystemSayHeard(text string, most time.Duration) bool {
	rv := gopodBrobotsSettings.ReporterVoices
	role := gopodBrobotsSettings.SignOff.SystemRole
	text = strings.TrimSpace(text)
	if role == "" || text == "" {
		return false
	}
	page, seconds, ok := gopodSayShown(role, "", text)
	if !ok {
		if gopodDryVoice != nil {
			gopodDryVoice(role, text, true)
			return true
		}
		return false
	}
	token := fmt.Sprintf("%d", time.Now().UnixNano())
	heard := make(chan struct{})
	if most > 0 && rv.HeardMarker != "" {
		gopodSayHeard.Lock()
		gopodSayHeard.waiting[token] = heard
		gopodSayHeard.Unlock()
	}
	// 2026-10-13: the system sting, then the speech (gopodSystemSting).
	// 2026-10-17 (operator: "if \"(music)\" is in the story.md ... play djent"):
	// a line ending with (music) has it after its words instead.
	musicFirst, musicLast := gopodMusicMarks(text)
	sting, stingSeconds := gopodSystemStingNow(text)
	if musicFirst {
		sting, stingSeconds = gopodMusicSting()
	}
	if sting != "" && !musicLast {
		GopodChatFeedAdd(rv.Role, gopodWithVolumeShare(sting))
		seconds += stingSeconds + rv.GapSeconds
	}
	GopodChatFeedAdd(rv.Role, gopodWithVolumeShare(page))
	if musicLast {
		if sting, stingSeconds := gopodMusicSting(); sting != "" {
			GopodChatFeedAdd(rv.Role, gopodWithVolumeShare(sting))
			seconds += stingSeconds + rv.GapSeconds
		}
	}
	ends := gopodVoiceQueueAdd(seconds + rv.GapSeconds)
	if most <= 0 || rv.HeardMarker == "" {
		return true
	}
	// 2026-10-14: no longer than the lines' own length plus heard_margin_seconds.
	if limit := gopodHeardLimit(ends); limit < most {
		most = limit
	}
	GopodChatFeedAdd(rv.Role, rv.HeardMarker+token)
	defer func() {
		gopodSayHeard.Lock()
		delete(gopodSayHeard.waiting, token)
		gopodSayHeard.Unlock()
	}()
	select {
	case <-heard:
		return true
	case <-time.After(most):
		return false
	}
}

// GopodVoiceHeard: the page has played every voice line up to this marker.
// GopodVoiceHeardSkipped: the page passed a heard marker; skipped voices before
// it didn't play, and the journal says so.
func GopodVoiceHeardSkipped(token, skipped string) {
	if n := strings.TrimSpace(skipped); n != "" && n != "0" {
		logger.Println("GOPOD_HEARD " + n + " voice(s) before it didn't play on the page (missing or blocked)")
	}
	GopodVoiceHeard(token)
}

func GopodVoiceHeard(token string) {
	gopodSayHeard.Lock()
	defer gopodSayHeard.Unlock()
	if ch, ok := gopodSayHeard.waiting[token]; ok {
		close(ch)
		delete(gopodSayHeard.waiting, token)
	}
}

// ## EOF
