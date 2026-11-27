package wirepod_ttr

import (
	"encoding/json"
	"os"
	"regexp"
	"sort"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

var gopodNameSpellingsPath = gopodPathWithEnvOverride("GOPOD_NAME_SPELLINGS_PATH", "webroot/gopod_name_spellings.json")

func loadGOPODNameSpellings() []gopodPronunciationEntry {
	data, err := os.ReadFile(gopodNameSpellingsPath)
	if err != nil {
		logger.Println("Unable to load GOPOD name spellings at " + gopodNameSpellingsPath + ": " + err.Error())
		return nil
	}
	var list struct {
		Entries []gopodPronunciationEntry `json:"name_spellings"`
	}
	if err := json.Unmarshal(data, &list); err != nil {
		logger.Println("Unable to parse GOPOD name spellings: " + err.Error())
		return nil
	}
	var entries []gopodPronunciationEntry
	for _, e := range list.Entries {
		if e.From == "" || e.To == "" {
			continue
		}
		e.pattern = regexp.MustCompile("(?i)" + regexp.QuoteMeta(e.From))
		entries = append(entries, e)
	}
	sort.SliceStable(entries, func(i, j int) bool { return len(entries[i].From) > len(entries[j].From) })
	return entries
}

var loadedGOPODNameSpellings = loadGOPODNameSpellings()

// gopodSpellNames writes GOPOD's own names the way the repo spells them.
func gopodSpellNames(text string) string {
	return gopodReplaceWholeWords(text, loadedGOPODNameSpellings)
}

// ## EOF
