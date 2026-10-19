// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_audio_levels_test.go
package wirepod_ttr

import "testing"

// 2026-10-18 (the operator's loudness method replaced the peak offsets): the
// meter's loudness is read, and only the voice worker's own files count as
// made voices (done as they're made; every other sound gets a done copy).
func TestAudioLevels(t *testing.T) {
	m := gopodLoudnessIRE.FindAllSubmatch([]byte("Integrated loudness:\n    I:         -18.6 LUFS\n"), -1)
	if len(m) != 1 || string(m[0][1]) != "-18.6" {
		t.Fatalf("loudness: %q", m)
	}
	if !gopodMadeVoiceRE.MatchString("bm_george_s098_891c93ef8239daa1.wav") || gopodMadeVoiceRE.MatchString("dJent3b.wav") {
		t.Fatal("made voice names")
	}
}
