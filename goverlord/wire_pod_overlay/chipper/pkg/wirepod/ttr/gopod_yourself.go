// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_yourself.go
package wirepod_ttr

import (
	"strings"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

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

// gopodYourselfCommand: a GOPOD Yourself line for one Brobot or "all".
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
	gopodOperatorShow(to, text, how)
	if how == gopodVoiceCommandHow {
		gopodSessionLogHeard(to, text)
	}
	logger.Println("GOPOD_YOURSELF to " + to)
	go func() {
		if how == gopodVoiceCommandHow {
			time.Sleep(time.Until(time.Now().Truncate(time.Second).Add(1200 * time.Millisecond)))
		}
		for _, b := range who {
			if line := g.Lines[b]; line != "" {
				gopodBodySay(b, line) // a scripted Brobot line: said as the Brobot
			}
		}
	}()
	return true
}

// ## EOF
