// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_yourself.go
package wirepod_ttr

import (
	"strings"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// "GOPOD Yourself" in the Brobots layer (operator 2026-10-09: "GOPOD Yourself > GOPOD
// layer (isn't built yet) ... tie in to brobot 1 CTA > newsletter ... brobot 2 CTA >
// crushn8r.ca"). Until the GOPOD layer is built, the wake phrase gets each Brobot's own
// call to action (gopod_yourself.lines): typed, on the numpad, or spoken (the
// GOPOD_YOURSELF voice command hands it to the chat). Said to both, Brobot 1 goes first.
// The line must be the phrase itself ("Brobot 2, GOPOD Yourself!"), so a question about
// it ("what is GOPOD Yourself?") still goes to the chat.

// gopodYourselfLeadWords may come before the phrase: the wake word and a Brobot's name.
var gopodYourselfLeadWords = map[string]bool{"hey": true, "vector": true, "brobot": true,
	"brobots": true, "bro": true, "bot": true, "bots": true, "1": true, "2": true,
	"one": true, "two": true, "won": true, "to": true, "too": true}

var gopodYourselfLast = struct {
	sync.Mutex
	at time.Time
}{}

// gopodIsYourself: the line is gopod_yourself.phrases, maybe after the lead words.
func gopodIsYourself(text string) bool {
	words := gopodAddressWords(gopodSpellNames(text))
	for len(words) > 0 && gopodYourselfLeadWords[words[0]] {
		words = words[1:]
	}
	said := strings.Join(words, " ")
	for _, p := range gopodBrobotsSettings.GOPODYourself.Phrases {
		if said != "" && said == strings.Join(gopodAddressWords(gopodSpellNames(p)), " ") {
			return true
		}
	}
	return false
}

// gopodYourselfCommand: a GOPOD Yourself line for one Brobot or "all". True when it was
// one. The same phrase again within repeat_seconds (both robots heard it) is dropped.
func gopodYourselfCommand(esn, text, how string) bool {
	if !gopodIsYourself(text) {
		return false
	}
	g := gopodBrobotsSettings.GOPODYourself
	gopodYourselfLast.Lock()
	again := time.Since(gopodYourselfLast.at) < gopodSecondsToDuration(g.RepeatSeconds)
	if !again {
		gopodYourselfLast.at = time.Now()
	}
	gopodYourselfLast.Unlock()
	if again {
		logger.Println("GOPOD_YOURSELF again within repeat_seconds - dropped")
		return true
	}
	who := []string{gopodBrobotName(esn)}
	to := who[0]
	if esn == gopodTypedAll {
		who, to = []string{"Brobot 1", "Brobot 2"}, "Brobots"
	}
	// Spoken, the page already shows what the robot heard (live 15:21: two
	// operator lines, Brobot 1's bubble between them) - gopodOperatorShow.
	gopodOperatorShow(to, text, how)
	if how == gopodVoiceCommandHow {
		gopodSessionLogHeard(to, text)
	}
	logger.Println("GOPOD_YOURSELF to " + to)
	go func() {
		// 2026-10-10 (live 22:17: Brobot 1's newsletter line showed above
		// "Backpack to Brobot 1"): spoken, the page shows what the robot heard
		// from Wire-Pod's log, placed at the end of that log line's second - the
		// Brobots' lines start after it.
		if how == gopodVoiceCommandHow {
			time.Sleep(time.Until(time.Now().Truncate(time.Second).Add(1200 * time.Millisecond)))
		}
		for _, b := range who {
			if line := g.Lines[b]; line != "" {
				gopodRecordSay(b, line)
			}
		}
	}()
	return true
}

// ## EOF
