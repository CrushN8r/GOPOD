// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_mic_voice.go
package wirepod_ttr

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// The mic's own voice (2026-10-24, operator: "When audio is voice from mic,
// process raw mic voice same as raw kokoro voice file * do the volume level
// smoothing part * same vosk for highlight timings * play processed mic-voice
// instead of operator-voice"): a numpad push-to-talk line keeps its 48 kHz
// recording; it's evened the way every made voice is (gopodLoudnessFile), timed
// by the voice worker's Vosk (an "align" job: the words as heard, for the word-
// by-word highlight), and played through the voice template in place of the
// Operator's made voice - "Generating..." and the hourglass while it's
// processed, then lit word by word as it plays.

var gopodMicVoiceFor sync.Map // the line's words (lower case) -> its recording

func gopodMicVoiceKey(text string) string {
	return strings.ToLower(strings.Join(strings.Fields(text), " "))
}

// gopodMicVoiceRemember: this line was said into the mic - its recording.
func gopodMicVoiceRemember(text, master string) {
	if strings.TrimSpace(master) != "" && gopodFileExists(master) {
		gopodMicVoiceFor.Store(gopodMicVoiceKey(text), master)
	}
}

// gopodMicVoicePaths: the evened copy's file and page path, named like a made
// voice ("op_mic_<16 hex>.wav") so the page reads its word times beside it.
func gopodMicVoicePaths(master, text string) (string, string) {
	sum := sha256.Sum256([]byte(master + "\n" + text))
	name := "op_mic_" + hex.EncodeToString(sum[:8]) + ".wav"
	return filepath.Join(gopodReporterAudioDir(), name), "/gopod_reporter_audio/" + name
}

// gopodMicVoiceFile: the recording evened and timed; its page path and length.
func gopodMicVoiceFile(master, text string) (string, float64, bool) {
	file, page := gopodMicVoicePaths(master, text)
	_ = gopodMkdirAll(filepath.Dir(file), 0o755)
	if err := gopodLoudnessFile(master, file); err != nil || !gopodFileExists(file) {
		// Evening off (or no ffmpeg): the recording as it is.
		data, rerr := os.ReadFile(master)
		if rerr != nil || gopodWriteFile(file, data, 0o644) != nil {
			return "", 0, false
		}
	}
	seconds, ok := gopodReporterAlign(file)
	return page, seconds, ok
}

// gopodReporterAlign: the voice worker times a recording's words (Vosk, as
// heard) beside it, <name>.words.json; its length in seconds.
func gopodReporterAlign(file string) (float64, bool) {
	gopodReporterWaitReady()
	gopodReporter.Lock()
	defer gopodReporter.Unlock()
	if !gopodReporter.ready {
		return 0, false
	}
	job, _ := json.Marshal(map[string]any{"align": file})
	if _, err := gopodReporter.in.Write(append(job, '\n')); err != nil || gopodReporter.in.Flush() != nil || !gopodReporter.out.Scan() {
		gopodReporter.ready = false
		gopodReporterStarting.Store(false)
		logger.Println("GOPOD_REPORTER_VOICES worker gone")
		return 0, false
	}
	var reply struct {
		OK      bool    `json:"ok"`
		Seconds float64 `json:"seconds"`
		Error   string  `json:"error"`
	}
	if json.Unmarshal(gopodReporter.out.Bytes(), &reply) != nil || !reply.OK {
		logger.Println("GOPOD_MIC_VOICE unable to time it: " + reply.Error)
		return 0, false
	}
	return reply.Seconds, true
}

// gopodMicVoiceSay: the line's own mic recording, through the voice template.
// False when it has none (the Operator's made voice then).
func gopodMicVoiceSay(host, text string) bool {
	v, ok := gopodMicVoiceFor.LoadAndDelete(gopodMicVoiceKey(text))
	if !ok {
		return false
	}
	master := v.(string)
	rv := gopodBrobotsSettings.ReporterVoices
	shown := strings.TrimSpace(text)
	_, page := gopodMicVoicePaths(master, text)
	if rv.GeneratingMarker != "" {
		GopodChatFeedAdd(rv.Role, rv.GeneratingMarker+"forwait:"+page+"|"+shown)
	}
	page, seconds, ok := gopodMicVoiceFile(master, text)
	if !ok {
		if rv.GeneratingMarker != "" {
			GopodChatFeedAdd(rv.Role, rv.GeneratingMarker+"off")
		}
		logger.Println("GOPOD_MIC_VOICE not ready, the made voice instead: " + gopodHeardShort(text))
		return false
	}
	logger.Println("GOPOD_MIC_VOICE the mic's own voice: " + filepath.Base(master) + " -> " + page)
	// 2026-10-26: its profile; the same words recorded before - archived, their
	// profile taken over (gopodVoiceProfileTakeOver).
	gopodVoiceProfileMade(filepath.Base(page), "mic", host, "", text)
	gopodVoiceProfileTakeOver(filepath.Base(page), text)
	gopodVoicePlay(host, text, page, seconds, true)
	return true
}

// ## EOF
