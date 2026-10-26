// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_audio_levels.go
package wirepod_ttr

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// Overall speaker volume normalization (2026-10-17, operator: "not actually
// edit the files, just a quick adjustment in the code * evaluate audio files
// to normalize all volumes to -3db? * tag & apply a volume offset to the raw
// audio, before \"default\" knob setting * this will give me a consistent
// volume range to adjust within"). Every WAV the chat page plays (the voice
// folder: voices, stings, song sounds and soundtracks) is measured once - its
// loudest sample - and given the offset that brings that peak to
// chat_audio.normalize_db. The offsets go to webroot/gopod_audio_levels.json;
// the page multiplies each sound's volume by its offset, before the master
// volume and the sound's own knob. The files are never changed. 0 = off.

var gopodAudioLevelsPath = gopodPathWithEnvOverride("GOPOD_AUDIO_LEVELS_PATH", "webroot/gopod_audio_levels.json")

// gopodAudioLevelsStart measures the voice folder now and every 30 s (new
// voice files are made all the time).
func gopodAudioLevelsStart() {
	go func() {
		for {
			gopodAudioLevelsUpdate()
			time.Sleep(30 * time.Second)
		}
	}()
}

// 2026-10-17 (operator: "Measure by average loudness instead * include .mp3
// weather sounds * include all sounds GOPOD uses"; "What I do in my audio
// editor is * I aim to adjustment a bit more than average loudness, to as close
// to but within the -3db range * then loudness over -3db I compress back to
// -3db * Can you do what I do"): each sound is measured for its average
// loudness (ffmpeg's EBU R128 meter), turned up or down to
// chat_audio.loudness_lufs, and anything that then goes over
// chat_audio.normalize_db is held back under it (ffmpeg's limiter). A made
// voice is done as it's made (gopodLoudnessFile, in place - it's GOPOD's own
// file); every other sound - the voice folder's stings, song sounds and
// soundtracks, and the weather sounds - gets a done copy in
// webroot/gopod_audio_norm/ (the originals are never changed), and
// gopod_audio_levels.json tells the page to play the copy ("made").

var gopodAudioNormDir = gopodPathWithEnvOverride("GOPOD_AUDIO_NORM_DIR", "webroot/gopod_audio_norm")

// gopodAudioSources: the folders of sounds the chat page plays, with their page paths.
func gopodAudioSources() [][2]string {
	return [][2]string{
		{gopodReporterAudioDir(), "/gopod_reporter_audio/"},
		{filepath.Join(gopodChipperRoot(), "webroot", "weather", "sounds"), "/weather/sounds/"},
	}
}

// gopodMadeVoiceRE: a voice the voice worker made (done as it's made).
var gopodMadeVoiceRE = regexp.MustCompile(`^[a-z]{2}_[a-z]+_(s\d+_)?[0-9a-f]{16}\.wav$`)

var gopodAudioMade = struct {
	sync.Mutex
	seen map[string]string // page path -> its source's size|mtime|settings, done
	made map[string]string // page path -> the done copy's page path
}{seen: map[string]string{}, made: map[string]string{}}

// gopodAudioLevelsUpdate does every sound not done yet and writes the list.
func gopodAudioLevelsUpdate() {
	ca := gopodBrobotsSettings.ChatAudio
	if ca.LoudnessLUFS == 0 {
		return
	}
	settings := jsonNumber(ca.LoudnessLUFS) + "|" + jsonNumber(ca.NormalizeDB)
	changed := false
	for _, src := range gopodAudioSources() {
		files, _ := filepath.Glob(filepath.Join(src[0], "*"))
		for _, f := range files {
			name := filepath.Base(f)
			ext := strings.ToLower(filepath.Ext(name))
			if (ext != ".wav" && ext != ".mp3") || gopodMadeVoiceRE.MatchString(name) {
				continue
			}
			info, err := os.Stat(f)
			if err != nil || info.IsDir() {
				continue
			}
			page := src[1] + name
			stamp := info.ModTime().String() + "|" + jsonNumber(float64(info.Size())) + "|" + settings
			sub := strings.Trim(src[1], "/")
			out := filepath.Join(gopodAudioNormDir, sub, strings.TrimSuffix(name, filepath.Ext(name))+".wav")
			gopodAudioMade.Lock()
			done := gopodAudioMade.seen[page] == stamp
			gopodAudioMade.Unlock()
			if done {
				continue
			}
			// A copy made before this Wire-Pod start, newer than its source, is kept.
			if o, err := os.Stat(out); err != nil || o.ModTime().Before(info.ModTime()) || !gopodAudioNormKept(out, settings) {
				if err := gopodLoudnessFile(f, out); err != nil {
					logger.Println("GOPOD_AUDIO_LEVELS " + page + ": " + err.Error())
					continue
				}
				gopodWriteFile(out+".settings", []byte(settings), 0o644)
			}
			gopodAudioMade.Lock()
			gopodAudioMade.seen[page] = stamp
			gopodAudioMade.made[page] = "/" + filepath.Base(gopodAudioNormDir) + "/" + sub + "/" + filepath.Base(out)
			gopodAudioMade.Unlock()
			changed = true
		}
	}
	if !changed && gopodFileExists(gopodAudioLevelsPath) {
		return
	}
	gopodAudioMade.Lock()
	out := struct {
		LoudnessLUFS float64           `json:"loudness_lufs"`
		CeilingDB    float64           `json:"ceiling_db"`
		Files        map[string]string `json:"files"`
		Made         map[string]string `json:"made"`
	}{ca.LoudnessLUFS, ca.NormalizeDB, map[string]string{}, map[string]string{}}
	for k, v := range gopodAudioMade.made {
		out.Made[k] = v
	}
	gopodAudioMade.Unlock()
	data, _ := json.Marshal(out)
	tmp := gopodAudioLevelsPath + ".tmp"
	if gopodWriteFile(tmp, data, 0o644) == nil {
		os.Rename(tmp, gopodAudioLevelsPath)
		logger.Println(fmt.Sprintf("GOPOD_AUDIO_LEVELS %d sounds evened to %s LUFS, held under %s dB", len(out.Made), jsonNumber(ca.LoudnessLUFS), jsonNumber(ca.NormalizeDB)))
	}
}

// gopodAudioNormKept: the copy was made with these settings.
func gopodAudioNormKept(out, settings string) bool {
	b, err := os.ReadFile(out + ".settings")
	return err == nil && string(b) == settings
}

// gopodLoudnessFile: the operator's method on one sound - measured, turned to
// chat_audio.loudness_lufs, then held under chat_audio.normalize_db (half a dB
// lower for the limiter, so the true peak stays within it) - written to out
// (out may be in: it's replaced when done). Not done when ffmpeg isn't there.
var gopodLoudnessFile = func(in, out string) error {
	ca := gopodBrobotsSettings.ChatAudio
	if ca.LoudnessLUFS == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	meter, _ := gopodAsOperator(exec.CommandContext(ctx, "ffmpeg", "-hide_banner", "-nostats", "-i", in, "-af", "ebur128", "-f", "null", "-")).CombinedOutput()
	gain := 0.0
	if m := gopodLoudnessIRE.FindAllSubmatch(meter, -1); len(m) > 0 {
		if lufs, err := strconv.ParseFloat(string(m[len(m)-1][1]), 64); err == nil && lufs > -60 {
			gain = ca.LoudnessLUFS - lufs
		}
	}
	if gain > 24 {
		gain = 24
	}
	limit := math.Pow(10, (ca.NormalizeDB-0.5)/20)
	if limit > 1 {
		limit = 1
	}
	if limit < 0.0625 {
		limit = 0.0625
	}
	if err := gopodMkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	tmp := out + ".tmp.wav"
	filter := fmt.Sprintf("volume=%.2fdB,alimiter=limit=%.4f:level=disabled:attack=5:release=50", gain, limit)
	if msg, err := gopodAsOperator(exec.CommandContext(ctx, "ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-i", in, "-af", filter, "-map_metadata", "-1", "-fflags", "+bitexact", "-flags:a", "+bitexact", "-c:a", "pcm_s16le", tmp)).CombinedOutput(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("ffmpeg: %v %s", err, strings.TrimSpace(string(msg)))
	}
	return os.Rename(tmp, out)
}

// gopodLoudnessIRE: the meter's integrated loudness line ("I: -18.6 LUFS").
var gopodLoudnessIRE = regexp.MustCompile(`I:\s+(-?[0-9.]+) LUFS`)

func gopodFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ## EOF
