// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_say_heard.go
package wirepod_ttr

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// reporter_voices.heard_marker: a voice-queue line "<marker><token>"; the page answers it on
// /api-sdk/gopod_voice_heard instead of playing it.
var gopodSayHeard = struct {
	sync.Mutex
	waiting map[string]chan struct{}
}{waiting: map[string]chan struct{}{}}

// gopodWaitHeard waits up to most for the page to have played every voice line queued before now
// (the heard marker); true once heard.
func gopodWaitHeard(most time.Duration) bool {
	rv := gopodBrobotsSettings.ReporterVoices
	if most <= 0 || rv.HeardMarker == "" {
		return false
	}
	if gopodRunCheckSilent() {
		return true
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

// GopodSystemSayHeard says text in Goverlord's voice with its bubble and, when most > 0, waits up
// to most for the page to have played it.
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
