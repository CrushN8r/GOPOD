package wirepod_ttr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// SOFT-LOCKED 2026-10-25 (operator: "soft lock this, specifically so I can further
// tweak guiding the llm for better thought bubble output, without affecting
// colouring and/or flavouring * some of the random stuff that comes out of the
// brobots is unexpected comedic gold, so that's why soft lock the persona, just
// tweak the answering-chains, etc..."). Locked: each Brobot's persona as the LLM
// gets it (personas + host_personas name/role/steps/signature + host_pairs banter
// and dial + host_persona_frames), the colour room (colour_slots.max_sentences),
// the flavour lines (scripted_snippets, all but the reminder's), and the
// temperatures. Open: answer_chains, response_chains, openers, nicknames,
// chain_prompts, answer_shapes, point_bubbles - everything that guides the thought
// bubbles. Changing a locked part is the operator's word only; then
// GOPOD_RELOCK=1 saves the new lock (gopod_probes/locks/persona_soft_lock.json).

type gopodPersonaLock struct {
	Locked             string              `json:"locked"`
	PersonaLines       map[string]string   `json:"persona_lines"`
	ColourMaxSentences int                 `json:"colour_max_sentences"`
	FlavourLines       []json.RawMessage   `json:"flavour_lines"`
	Temperature        float32             `json:"temperature"`
	SayLineTemperature float32             `json:"say_line_temperature"`
}

func gopodPersonaLockNow(t *testing.T) gopodPersonaLock {
	raw, _ := gopodRawSettingsForTest(t)
	l := gopodPersonaLock{
		Locked:             "2026-10-25",
		PersonaLines:       gopodHostPersonaLines(),
		ColourMaxSentences: gopodBrobotsSettings.ColourSlots.MaxSentences,
		Temperature:        gopodBrobotsSettings.LLM.Temperature,
		SayLineTemperature: gopodBrobotsSettings.LLM.SayLineTemperature,
	}
	ss, _ := raw["scripted_snippets"].(map[string]any)
	lines, _ := ss["lines"].([]any)
	for _, e := range lines {
		if m, ok := e.(map[string]any); ok && m["when"] == "reminder" {
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
	if want.Temperature != got.Temperature || want.SayLineTemperature != got.SayLineTemperature {
		t.Errorf("SOFT-LOCKED: llm temperature changed")
	}
}
