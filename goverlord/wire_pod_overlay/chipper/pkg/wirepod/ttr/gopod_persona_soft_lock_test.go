package wirepod_ttr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type gopodPersonaLock struct {
	Locked             string            `json:"locked"`
	PersonaLines       map[string]string `json:"persona_lines"`
	ColourMaxSentences int               `json:"colour_max_sentences"`
	FlavourLines       []json.RawMessage `json:"flavour_lines"`
	Temperature        float32           `json:"temperature"`
	SayLineTemperature float32           `json:"say_line_temperature"`
	GomadPersonas      map[string]string `json:"gomad_personas"`
}

func gopodPersonaLockNow(t *testing.T) gopodPersonaLock {
	_, _ = gopodRawSettingsForTest(t)
	gopodDialZero.Range(func(k, _ any) bool { gopodDialZero.Delete(k); return true })
	l := gopodPersonaLock{
		Locked:             "2026-10-26",
		PersonaLines:       gopodHostPersonaLines(),
		ColourMaxSentences: gopodBrobotsSettings.ColourSlots.MaxSentences,
		Temperature:        gopodBrobotsSettings.LLM.Temperature,
		SayLineTemperature: gopodBrobotsSettings.LLM.SayLineTemperature,
	}
	l.GomadPersonas = map[string]string{}
	for host, p := range gopodBrobotsSettings.Managers.Personas {
		l.GomadPersonas[host] = strings.TrimSpace(strings.ReplaceAll(gopodPersonaPrompt(host), gopodHereNow(host), "")) + "\n" + p.WhoLine
	}
	for _, e := range gopodBrobotsSettings.ScriptedSnippets.Lines {
		if e.When == "reminder" {
			continue
		}
		b, _ := json.Marshal(e)
		l.FlavourLines = append(l.FlavourLines, b)
	}
	return l
}

func TestPersonaSoftLock(t *testing.T) {
	root := os.Getenv("GOPOD_OVERLAY_CHIPPER")
	if root == "" {
		t.Skip("GOPOD_OVERLAY_CHIPPER not set")
	}
	path := filepath.Join(root, "gopod_probes", "locks", "persona_soft_lock.json")
	now := gopodPersonaLockNow(t)
	if os.Getenv("GOPOD_RELOCK") == "1" {
		b, _ := json.MarshalIndent(now, "", "  ")
		if err := os.WriteFile(path, append(b, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("the persona lock is missing: %v", err)
	}
	var want gopodPersonaLock
	if err := json.Unmarshal(b, &want); err != nil {
		t.Fatal(err)
	}
	nb, _ := json.Marshal(now)
	var got gopodPersonaLock
	json.Unmarshal(nb, &got)
	want.Locked, got.Locked = "", ""
	if !reflect.DeepEqual(want.PersonaLines, got.PersonaLines) {
		t.Errorf("SOFT-LOCKED: a Brobot's persona changed (operator's word only, then GOPOD_RELOCK=1)")
	}
	if want.ColourMaxSentences != got.ColourMaxSentences {
		t.Errorf("SOFT-LOCKED: colour_slots.max_sentences changed")
	}
	wf, _ := json.Marshal(want.FlavourLines)
	gf, _ := json.Marshal(got.FlavourLines)
	if string(wf) != string(gf) {
		t.Errorf("SOFT-LOCKED: the flavour lines (scripted_snippets) changed")
	}
	if !reflect.DeepEqual(want.GomadPersonas, got.GomadPersonas) {
		t.Errorf("SOFT-LOCKED: a Gomad's persona changed (managers.personas; operator's word only, then GOPOD_RELOCK=1)")
	}
	if want.Temperature != got.Temperature || want.SayLineTemperature != got.SayLineTemperature {
		t.Errorf("SOFT-LOCKED: llm temperature changed")
	}
}
