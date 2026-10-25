// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_egg_voice.go
package wirepod_ttr

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// The Easter egg's voice in pieces (2026-10-24, operator: '"Congratulations!",
// "Easter Egg \'", "\' found.", & "MouseOver to Zoom In", need to be generated
// once and reused * only [easter egg title(s)] need live generation * generating
// + hourglass toggles upon file generated rule applies'): Goverlord's
// congratulations is said from its fixed pieces - each made once and kept, like
// every voice - and the egg's title, the only piece made live. The pieces are
// joined into one voice (their word times with them), so the page lights it as
// one line; "Generating..." and the hourglass show until the joined voice is
// ready.

// gopodEggVoiceParts: what's said, piece by piece - each line of
// topic_pictures.egg_text its own piece (the quotes not said): the fixed lines made
// once and kept, the line with {eggs} (or the older {name}) filled with eggs and
// made live; then the zoom line.
// 2026-10-25 (operator: "NEW Template * [Congratulations! Easter Eggs Revealed Thus
// Far!] * '[easter eggs]' * Line 1 generated once/used many * Line 2 generated
// many/used few").
func gopodEggVoiceParts(eggs, zoom string) []string {
	tmpl := gopodBrobotsSettings.TopicPictures.EggText
	if !strings.Contains(tmpl, "{eggs}") && !strings.Contains(tmpl, "{name}") {
		return nil
	}
	if strings.TrimSpace(eggs) == "" {
		return nil
	}
	// Only the quotes around a line go; an apostrophe inside it stays ("Granny's").
	unquote := func(s string) string { return strings.TrimSpace(strings.Trim(strings.TrimSpace(s), "'\"")) }
	var parts []string
	for _, line := range strings.Split(tmpl, "\n") {
		line = strings.NewReplacer("{eggs}", eggs, "{name}", eggs).Replace(line)
		if p := unquote(line); p != "" {
			parts = append(parts, p)
		}
	}
	if z := strings.TrimSpace(zoom); z != "" {
		parts = append(parts, z)
	}
	return parts
}

// gopodEggSay: the egg's congratulations (text, as shown) said from its pieces.
// False when it can't be (no system voice, no {name} in egg_text) - the caller
// says it whole.
func gopodEggSay(text, name, zoom string) bool {
	host := gopodBrobotsSettings.SignOff.SystemRole
	rv := gopodBrobotsSettings.ReporterVoices
	parts := gopodEggVoiceParts(name, zoom)
	if host == "" || len(parts) == 0 || gopodDryVoice != nil {
		return false
	}
	file, page := gopodReporterPaths(host, "egg voice\n"+strings.Join(parts, "\n"))
	if file == "" {
		return false
	}
	shown := strings.TrimSpace(text)
	seconds, made := gopodWavSeconds(file)
	if rv.GeneratingMarker != "" && shown != "" {
		if made {
			GopodChatFeedAdd(rv.Role, rv.GeneratingMarker+"forlit:"+page+"|"+shown)
		} else {
			GopodChatFeedAdd(rv.Role, rv.GeneratingMarker+"forwait:"+page+"|"+shown)
		}
	}
	if !made {
		var err error
		if seconds, err = gopodEggVoiceMake(host, file, parts); err != nil {
			logger.Println("GOPOD_EGG_VOICE " + err.Error())
			if rv.GeneratingMarker != "" {
				GopodChatFeedAdd(rv.Role, rv.GeneratingMarker+"off")
			}
			return true
		}
	}
	gopodVoicePlay(host, text, page, seconds, true)
	return true
}

// gopodEggVoiceMake: each piece's kept voice (made the first time), joined into file.
func gopodEggVoiceMake(host, file string, parts []string) (float64, error) {
	var files []string
	for _, p := range parts {
		pg, _, ok := gopodReporterFile(host, p)
		if !ok {
			return 0, errors.New("piece not made: " + p)
		}
		files = append(files, filepath.Join(gopodReporterAudioDir(), filepath.Base(pg)))
	}
	seconds, err := gopodWavJoin(file, files, gopodBrobotsSettings.TopicPictures.EggVoiceGapMS)
	if err != nil {
		return 0, errors.New("join failed: " + err.Error())
	}
	logger.Println("GOPOD_EGG_VOICE made from " + strings.Join(parts, " | "))
	return seconds, nil
}

// gopodWavJoin: the voices in files, one after another with gapMS of silence
// between, written to out (same format as the first), and their word times
// (<name>.words.json) moved to where each now starts. Returns out's seconds.
func gopodWavJoin(out string, files []string, gapMS int) (float64, error) {
	var head []byte
	var pcm []byte
	type word struct {
		T string  `json:"t"`
		S float64 `json:"s"`
		E float64 `json:"e"`
	}
	var words []word
	aligner, at := "", 0.0
	byteRate, align := 0, 2
	for i, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return 0, err
		}
		if len(b) < 44 || string(b[0:4]) != "RIFF" {
			return 0, os.ErrInvalid
		}
		if head == nil {
			head = append([]byte{}, b[:44]...)
			byteRate = int(binary.LittleEndian.Uint32(b[28:32]))
			align = int(binary.LittleEndian.Uint16(b[32:34]))
			if byteRate <= 0 || align <= 0 {
				return 0, os.ErrInvalid
			}
		} else if string(b[20:36]) != string(head[20:36]) {
			return 0, os.ErrInvalid // a different format
		}
		if i > 0 && gapMS > 0 {
			gap := byteRate * gapMS / 1000
			gap -= gap % align
			pcm = append(pcm, make([]byte, gap)...)
			at += float64(gap) / float64(byteRate)
		}
		var times struct {
			Aligner string `json:"aligner"`
			Words   []word `json:"words"`
		}
		data := b[44:]
		cut := 0.0
		if wb, err := os.ReadFile(strings.TrimSuffix(f, ".wav") + ".words.json"); err == nil && json.Unmarshal(wb, &times) == nil {
			if aligner == "" {
				aligner = times.Aligner
			}
			// 2026-10-24 (live: 9.8 s joined, each piece's quiet ends between the
			// words): each piece trimmed to its words - egg_voice_edge_ms of air
			// kept - except the first piece's start and the last one's end.
			if n := len(times.Words); n > 0 {
				edge := float64(gopodBrobotsSettings.TopicPictures.EggVoiceEdgeMS) / 1000
				dur := float64(len(data)) / float64(byteRate)
				from, to := 0.0, dur
				if i > 0 {
					from = times.Words[0].S - edge
				}
				if i < len(files)-1 {
					to = times.Words[n-1].E + edge
				}
				if from < 0 {
					from = 0
				}
				if to > dur || to <= from {
					to = dur
				}
				a := int(from * float64(byteRate))
				z := int(to * float64(byteRate))
				a -= a % align
				z -= z % align
				data = data[a:z]
				cut = float64(a) / float64(byteRate)
			}
			for _, w := range times.Words {
				words = append(words, word{w.T, w.S - cut + at, w.E - cut + at})
			}
		}
		pcm = append(pcm, data...)
		at += float64(len(data)) / float64(byteRate)
	}
	binary.LittleEndian.PutUint32(head[4:8], uint32(36+len(pcm)))
	binary.LittleEndian.PutUint32(head[40:44], uint32(len(pcm)))
	tmp := out + ".part"
	if err := os.WriteFile(tmp, append(head, pcm...), 0o644); err != nil {
		return 0, err
	}
	if aligner == "" {
		aligner = "joined"
	}
	// The voice worker's own format ("aligner": "vosk" - the checks read it so).
	wj, _ := json.Marshal(map[string]any{"aligner": aligner, "words": words})
	wj = []byte(strings.Replace(string(wj), `"aligner":"`, `"aligner": "`, 1))
	_ = os.WriteFile(strings.TrimSuffix(out, ".wav")+".words.json", wj, 0o644)
	if err := os.Rename(tmp, out); err != nil {
		return 0, err
	}
	return at, nil
}

// gopodEggVoiceCheck (system-test-eggvoice): an egg voice made for real - the kept
// pieces and a sample title - joined, not played. Returns what was made.
func gopodEggVoiceCheck() (bool, string) {
	host := gopodBrobotsSettings.SignOff.SystemRole
	parts := gopodEggVoiceParts("Get Photo", gopodBrobotsSettings.TopicPictures.EggZoomText)
	if host == "" || len(parts) == 0 {
		return false, "no system voice, or no {name} in topic_pictures.egg_text"
	}
	file, _ := gopodReporterPaths(host, "egg voice\n"+strings.Join(parts, "\n"))
	seconds, err := gopodEggVoiceMake(host, file, parts)
	if err != nil {
		return false, err.Error()
	}
	return true, fmt.Sprintf("%d pieces (%s) joined: %.1f s, word times kept", len(parts), strings.Join(parts, " | "), seconds)
}

// ## EOF
