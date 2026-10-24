// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_native.go
package wirepod_ttr

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// Wire-Pod's own (native) voice answers - "Hey Vector, what time is it", the
// weather, volume, and the rest of Wire-Pod's intent list - on or off
// (operator 2026-10-10: "Native “Hey Vector” answers during a run ... Keyboard
// Line: Brobots ()1 ()2 ()All | Native?[] Numpad?[] ... Where 'Native?[]' makes
// this optional? * Change Wire-Pod's code if need be"):
//   - the chat page's "Native?" box turns them on or off for the session
//     (native.on_at_start is the state at Wire-Pod's start);
//   - during a run they are always off (gopod_session.go, Rule C: a run owns
//     the Brobots).
// Off: a phrase that matches a native intent gets Wire-Pod's quiet ending
// instead, so the robot goes back to what it was doing. GOPOD's own voice
// commands (custom intents) and chat questions (the LLM) are not native and
// always get through. Native intents in native.always_on still answer.

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

// gopodNativeBlocked: true (and the quiet ending sent) when this native intent
// must not answer now.
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

// gopodNativeLetGo (operator 2026-10-14: "Face recognition should be working as
// native wire-pod ... my GOPOD stuff should be working on top of existing native
// working mechanisms"; live 20:22: "who am i" and "remember my face" were sent
// while the chat hold held Brobot 2, so his own face behaviors never ran): an
// intent the robot acts on himself lets go of his chat hold first; the hold comes
// back with the next chat answer. The quiet ending (a blocked one) keeps it.
// gopodNativeBubble (2026-10-18, operator: "\"Who am I\" (Backpack to Brobot 1)
// show a generic silent system bubble: wire-pod intent activated"): a native
// intent sent to a Brobot (Wire-Pod's own answer - his words aren't known here)
// shows native_intent.text in a silent system bubble (native_intent.role), once
// per native_intent.once_seconds. Wire-Pod's system intents (no audio,
// unmatched) don't.
var gopodNativeBubbleAt = map[string]time.Time{}

// gopodNativeToManager (2026-10-21, operator: "weather goes to PLAYHEAD
// instead"): a native intent in native.to_manager goes to the chat instead
// (true: Wire-Pod's own answer is skipped), and the Brobot's chat answer gets
// that manager take (gopodNativeTakeFor) - spoken weather is PLAYHEAD's, the
// same as a typed one, even when Vosk heard "whether".
var gopodNativeTake sync.Map // Brobot name -> gopodNativeAsk

// gopodNativeAsk: the take kind a native intent left, and the line it came with.
type gopodNativeAsk struct{ kind, text string }

func gopodNativeToManager(esn, intent, text string) bool {
	// 2026-10-24: every "good <part of day>" is GOPOD's, whatever Wire-Pod made
	// of it (good morning, praise...) - it goes to the chat (gopod_greetings.go).
	if gopodGreetingSaid(text) != "" && !gopodScriptPlaying.Load() && !gopodRunActive() {
		logger.Println("GOPOD_NATIVE " + intent + " to the chat: a greeting, GOPOD's (" + text + ")")
		return true
	}
	kind, ok := gopodBrobotsSettings.Native.ToManager[intent]
	// A run owns the Brobots (gopodManagerTakes is live only).
	if !ok || kind == "" || gopodScriptPlaying.Load() || gopodRunActive() {
		return false
	}
	gopodNativeTake.Store(gopodBrobotName(esn), gopodNativeAsk{kind, strings.TrimSpace(strings.ToLower(text))})
	logger.Println("GOPOD_NATIVE " + intent + " to the chat for its manager (" + kind + "): " + text)
	return true
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

// gopodLLMClosePass (2026-10-21, live 21:25: "Wire-Pod intent activated." after
// a chat answer): Wire-Pod closes a chat answer's voice request with
// intent_greeting_hello - that pass is no native intent, so no bubble.
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
	// 2026-10-21: a notice inside his bubble (the bubble notice template).
	gopodNoticeShow(gopodBrobotName(esn), gopodBrobotsSettings.BubbleNotices.NativeIntent, map[string]string{"intent": intent})
}

func gopodNativeLetGo(esn, intent string) {
	if esn == "" || intent == "intent_system_noaudio" {
		return
	}
	GopodChatHoldRelease(esn)
}

// ## EOF
