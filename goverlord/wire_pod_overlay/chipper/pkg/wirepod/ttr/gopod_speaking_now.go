// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_speaking_now.go
package wirepod_ttr

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// The chat bubble page highlights the sentence a robot is saying. Wire-Pod
// speaks a chat reply one sentence at a time (gopod_speech_sentences.go); before
// each sentence it writes that sentence here, and when the reply ends it writes
// active=false. A touch interrupt instead writes interrupted=true with the
// sentence the robot stopped on, so the page can show that outline solid. The page (webroot/index.html,
// pollSpeakingNow) reads this file and outlines the matching sentence in that
// robot's newest bubble. A plain file next to the page, the same way
// gopod_brobot_eye_colors.json works, so nothing is added to the Logs page.
var gopodSpeakingNowPath = gopodPathWithEnvOverride(
	"GOPOD_SPEAKING_NOW_PATH", "webroot/gopod_speaking_now.json",
)

var gopodSpeakingNowMu sync.Mutex
var gopodSpeakingNowSeq int64
var gopodSpeakingNowLast = map[string]string{}
var gopodSpeakingNowLastPiece = map[string]int{}

// gopodLastBrobotFeed: each Brobot's newest bubble's chat event id.
var gopodLastBrobotFeed sync.Map

type gopodSpeakingNow struct {
	ESN         string `json:"esn"`
	Text        string `json:"text"`
	Active      bool   `json:"active"`
	Interrupted bool   `json:"interrupted"`
	// Piece: the number of the bubble piece being said (gopod_highlight_pieces.go);
	// -1 when there is none and the page matches by words instead.
	Piece  int   `json:"piece"`
	Seq    int64 `json:"seq"`
	// Feed: the chat event of the bubble being said (2026-10-21) - the page lights
	// that bubble, no guessing which.
	Feed int `json:"feed,omitempty"`
	TimeMs int64 `json:"time_ms"`
}

func gopodWriteSpeakingNow(esn, text string, active, interrupted bool) {
	gopodWriteSpeakingPiece(esn, text, -1, active, interrupted)
}

// gopodWriteSpeakingPiece is gopodWriteSpeakingNow with the bubble piece's number.
func gopodWriteSpeakingPiece(esn, text string, piece int, active, interrupted bool) {
	// 2026-10-11 (operator: "can you log highlighted splits within a bubble?"):
	// each lit piece, by its number in the bubble (-1: matched by its words).
	if active {
		logger.LogDebugUI(fmt.Sprintf("GOPOD_HIGHLIGHT %s piece=%d: %s", gopodBrobotName(esn), piece, text))
	}
	gopodSpeakingNowMu.Lock()
	defer gopodSpeakingNowMu.Unlock()
	gopodSpeakingNowSeq++
	if active {
		gopodSpeakingNowLast[esn] = text
		gopodSpeakingNowLastPiece[esn] = piece
	}
	feed := 0
	if v, ok := gopodLastBrobotFeed.Load(gopodBrobotName(esn)); ok {
		feed, _ = v.(int)
	}
	data, _ := json.Marshal(gopodSpeakingNow{ESN: esn, Text: text, Active: active, Interrupted: interrupted, Piece: piece, Seq: gopodSpeakingNowSeq, TimeMs: time.Now().UnixMilli(), Feed: feed})
	tmp := gopodSpeakingNowPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		logger.Println("GOPOD speaking-now write failed: " + err.Error())
		return
	}
	if err := os.Rename(tmp, gopodSpeakingNowPath); err != nil {
		logger.Println("GOPOD speaking-now write failed: " + err.Error())
	}
}

// gopodMarkSpeakingInterrupted: a touch (or wake word) interrupt was accepted -
// the sentence this robot is saying is written again with interrupted=true, so
// the page shows it solid right away.
func gopodMarkSpeakingInterrupted(esn string) {
	gopodSpeakingNowMu.Lock()
	text, piece := gopodSpeakingNowLast[esn], gopodSpeakingNowLastPiece[esn]
	gopodSpeakingNowMu.Unlock()
	gopodWriteSpeakingPiece(esn, text, piece, false, true)
}

// GopodSpeakSentences (2026-09-30): a song line (sdkapp say_text with
// sentence_highlight=1) said one sentence at a time, the same split and the
// same file as a chat reply, so the chat page lights each sentence of a Bingo
// bubble while it is said (operator: "make the bingo text act the same as the
// bubbles ... per sentence"). say speaks one sentence and returns when done.
// offset >= 0 (2026-10-10, a combo step: one bubble, its parts said apart):
// these pieces start at that piece of the bubble.
func GopodSpeakSentences(esn, text string, offset int, say func(sentence string)) {
	marked := strings.Contains(text, gopodBrobotsSettings.ScriptedSplit.SegmentMarker)
	pieces := gopodSpeakPieces(text)
	if offset >= 0 && !marked {
		pieces = []string{strings.TrimSpace(text)} // one piece of the bubble ("... Lol!")
	}
	for i, sentence := range pieces {
		piece := -1
		if marked {
			piece = i // the page shows the song's own marked pieces, in this order
		}
		if offset >= 0 {
			piece = offset + i
		}
		gopodWriteSpeakingPiece(esn, sentence, piece, true, false)
		say(sentence)
	}
	gopodWriteSpeakingNow(esn, "", false, false)
}

// gopodSpeakPieces: a scripted song line arrives already split by the song
// engine, its pieces joined with scripted_split.segment_marker
// (gopod_brobots_settings.json) - the robot says exactly those pieces, the same
// ones the chat bubble lights. Any other text keeps the sentence rule
// (gopodSentences).
func gopodSpeakPieces(text string) []string {
	marker := gopodBrobotsSettings.ScriptedSplit.SegmentMarker
	if marker == "" || !strings.Contains(text, marker) {
		return gopodSentences(text)
	}
	var pieces []string
	for _, piece := range strings.Split(text, marker) {
		if piece = strings.TrimSpace(piece); piece != "" {
			pieces = append(pieces, piece)
		}
	}
	return pieces
}

// ## EOF
