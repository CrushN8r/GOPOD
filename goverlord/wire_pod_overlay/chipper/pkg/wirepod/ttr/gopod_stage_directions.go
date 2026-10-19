// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_stage_directions.go
package wirepod_ttr

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// Text the model wraps in asterisks is either a stage direction (*laugh*) or
// emphasis (*crazy*, **AI Wordplay**). Only stage directions whose first word is
// on the list in stage_directions.json are dropped from speech; everything else
// keeps its words and loses only the asterisks. The list is data, read at
// startup: edit the JSON and restart wire-pod.service, no rebuild.

// gopodStageDirectionsPath resolves relative to gopodChipperRoot() (see
// gopod_paths.go), overridable via GOPOD_STAGE_DIRECTIONS_PATH.
var gopodStageDirectionsPath = gopodPathWithEnvOverride(
	"GOPOD_STAGE_DIRECTIONS_PATH", "gopod_probes/stage_directions.json",
)

// Floor list, used only if stage_directions.json cannot be read or parsed.
var gopodStageDirectionsFloor = []string{"laugh", "laughs", "giggle", "giggles", "chuckle", "chuckles", "sigh", "sighs", "wink", "winks", "grin", "grins"}

type gopodStageDirectionsFile struct {
	Description     string   `json:"description"`
	StageDirections []string `json:"stage_directions"`
}

func loadGOPODStageDirections() map[string]bool {
	words := gopodStageDirectionsFloor
	data, err := os.ReadFile(gopodStageDirectionsPath)
	if err != nil {
		logger.Println("Unable to load GOPOD stage directions at " + gopodStageDirectionsPath + ", using built-in floor list: " + err.Error())
	} else {
		var file gopodStageDirectionsFile
		if err := json.Unmarshal(data, &file); err != nil {
			logger.Println("Unable to parse GOPOD stage directions at " + gopodStageDirectionsPath + ", using built-in floor list: " + err.Error())
		} else if len(file.StageDirections) > 0 {
			words = file.StageDirections
		}
	}
	set := make(map[string]bool, len(words))
	for _, w := range words {
		if w = strings.ToLower(strings.TrimSpace(w)); w != "" {
			set[w] = true
		}
	}
	return set
}

var loadedGOPODStageDirections = loadGOPODStageDirections()

// One to three asterisks either side: *x*, **x**, ***x***.
var gopodAsteriskSpanPattern = regexp.MustCompile(`\*{1,3}([^*\n]+?)\*{1,3}`)
var gopodFirstWordPattern = regexp.MustCompile(`[A-Za-z']+`)

func gopodIsStageDirection(inner string) bool {
	first := gopodFirstWordPattern.FindString(inner)
	return first != "" && loadedGOPODStageDirections[strings.ToLower(first)]
}

// gopodStripStageDirections drops listed stage directions and keeps every other
// asterisk-wrapped span as plain spoken words. Stray asterisks left over are
// removed.
func gopodStripStageDirections(line string) string {
	line = gopodAsteriskSpanPattern.ReplaceAllStringFunc(line, func(span string) string {
		inner := gopodAsteriskSpanPattern.FindStringSubmatch(span)[1]
		if gopodIsStageDirection(inner) {
			return ""
		}
		return inner
	})
	line = strings.ReplaceAll(line, "*", "")
	return strings.Join(strings.Fields(line), " ")
}

// gopodCueWordNotePattern: a note the model adds about its own cue word,
// "(Cue word: thinking)".
var gopodCueWordNotePattern = regexp.MustCompile(`(?i)\(\s*cue\s*words?\s*:[^)]*\)`)

// gopodStripCueWordNotes (2026-10-11, operator: "(Cue word: thinking), (Cue
// word: darts) ... messes up highlighting splits"): the note is dropped before
// the reply is split, so neither the bubble nor the speech has it.
func gopodStripCueWordNotes(text string) string {
	if !gopodCueWordNotePattern.MatchString(text) {
		return text
	}
	return strings.TrimSpace(gopodCueWordNotePattern.ReplaceAllString(text, ""))
}

// ## EOF
