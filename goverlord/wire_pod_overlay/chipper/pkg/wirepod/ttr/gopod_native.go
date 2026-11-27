// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_native.go
package wirepod_ttr

import (
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

var gopodNativeOn atomic.Bool

func init() { gopodNativeOn.Store(gopodBrobotsSettings.Native.OnAtStart) }

// GopodNativeSet: the page's Native? box.
func GopodNativeSet(on bool) {
	if gopodNativeOn.Swap(on) != on {
		state := "off"
		if on {
			state = "on"
		}
		logger.Println("GOPOD_NATIVE Wire-Pod's own answers " + state)
	}
}

// GopodNativeOn: the box's state.
func GopodNativeOn() bool { return gopodNativeOn.Load() }

// gopodNativeBlocked: true (and the quiet ending sent) when this native intent must not answer now.
func gopodNativeBlocked(req interface{}, esn, intent, text string) bool {
	for _, keep := range gopodBrobotsSettings.Native.AlwaysOn {
		if keep == intent {
			return false
		}
	}
	why := ""
	if r, ok := gopodRunNow(); ok {
		why = "run " + r.Name + " is going"
	} else if !gopodNativeOn.Load() {
		why = "Native? is off"
	}
	if why == "" {
		return false
	}
	logger.Println("GOPOD_NATIVE " + intent + " not answered for " + esn + " (" + why + "): " + text)
	if req != nil {
		IntentPass(req, "intent_system_noaudio", text, map[string]string{}, false)
	}
	return true
}

// gopodNativeLetGo: an intent the robot acts on himself lets go of his chat hold first; the hold
// comes back with the next chat answer.
var gopodNativeBubbleAt = map[string]time.Time{}

// gopodNativeToManager: a native intent in native.to_manager goes to the chat instead (true:
// Wire-Pod's own answer is skipped), and the Brobot's chat answer gets that manager take
// (gopodNativeTakeFor) - spoken weather is PLAYHEAD's, the same as a typed one, even when Vosk
// heard "whether".
var gopodNativeTake sync.Map // Brobot name -> gopodNativeAsk

// gopodNativeAsk: the take kind a native intent left, and the line it came with.
type gopodNativeAsk struct{ kind, text string }

func gopodNativeToManager(esn, intent, text string) bool {
	if gopodGreetingSaid(text) != "" && !gopodScriptPlaying.Load() && !gopodRunActive() {
		logger.Println("GOPOD_NATIVE " + intent + " to the chat: a greeting, GOPOD's (" + text + ")")
		return true
	}
	kind, ok := gopodBrobotsSettings.Native.ToManager[intent]
	if !ok || kind == "" || gopodScriptPlaying.Load() || gopodRunActive() {
		return false
	}
	if !gopodLineIsKind(text, kind) {
		logger.Println("GOPOD_NATIVE " + intent + " to the chat as a plain line - no " + kind + " card for: " + text)
		return true
	}
	gopodNativeTake.Store(gopodBrobotName(esn), gopodNativeAsk{kind, strings.TrimSpace(strings.ToLower(text))})
	logger.Println("GOPOD_NATIVE " + intent + " to the chat for its manager (" + kind + "): " + text)
	return true
}

// gopodLineIsKind: the line looks up a card named for the kind (weather.md for "weather") - the
// card's own words decide, mishears ("whether") included.
func gopodLineIsKind(text, kind string) bool {
	for _, c := range gopodLookup(text) {
		if strings.Contains(strings.ToLower(filepath.Base(c.Path)), strings.ToLower(kind)) {
			return true
		}
	}
	return false
}

// gopodNativeTakeIs: a native intent left this take kind for this line of his.
func gopodNativeTakeIs(brobot, kind, question string) bool {
	v, ok := gopodNativeTake.Load(brobot)
	return ok && v.(gopodNativeAsk).kind == kind && v.(gopodNativeAsk).text == strings.TrimSpace(strings.ToLower(question))
}

// gopodNativeTakeFor: the manager take a native intent left for this Brobot's answer.
func gopodNativeTakeFor(brobot, question string) (gopodManagerTake, bool) {
	v, ok := gopodNativeTake.Load(brobot)
	if !ok || v.(gopodNativeAsk).text != strings.TrimSpace(strings.ToLower(question)) {
		return gopodManagerTake{}, false
	}
	for _, t := range gopodBrobotsSettings.Managers.Takes {
		if t.Kind == v.(gopodNativeAsk).kind && t.Host != "" {
			return t, true
		}
	}
	return gopodManagerTake{}, false
}

// gopodLLMClosePass: Wire-Pod closes a chat answer's voice request with intent_greeting_hello -
// that pass is no native intent, so no bubble.
var gopodLLMClosePass sync.Map

func gopodNativeBubble(esn, intent string) {
	if _, closing := gopodLLMClosePass.LoadAndDelete(esn); closing {
		return
	}
	ni := gopodBrobotsSettings.NativeIntent
	if esn == "" || strings.HasPrefix(intent, "intent_system_") {
		return
	}
	gopodChatFeed.Lock()
	last := gopodNativeBubbleAt[esn]
	recent := time.Since(last) < gopodSecondsToDuration(ni.OnceSeconds)
	if !recent {
		gopodNativeBubbleAt[esn] = time.Now()
	}
	gopodChatFeed.Unlock()
	if recent {
		return
	}
	logger.Println("GOPOD_NATIVE " + gopodBrobotName(esn) + ": " + intent)
	gopodNoticeShow(gopodBrobotName(esn), gopodBrobotsSettings.BubbleNotices.NativeIntent, map[string]string{"intent": intent})
}

func gopodNativeLetGo(esn, intent string) {
	if esn == "" || intent == "intent_system_noaudio" {
		return
	}
	GopodChatHoldRelease(esn)
}

// ## EOF
