// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_play.go
package wirepod_ttr

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// pha0b and phcal in the chat (2026-10-19, operator: "What else from the
// survey, can be absorbed over from PLAYHEAD aliases to new menu?"; "Build it
// all in please"; gopod_notes/PLAYHEAD_ALIASES_ABSORB_SURVEY_001.md):
//   system-play <song> [from <step>] [to <step>] [robot 1|2] [dry]
//       pha0b: a song, or part of it (step id or number), one Brobot or both,
//       live or dry (no robots, nothing said). Golden-engine songs only.
//   system-gap <seconds>|off|default    pha0b's reporter gap? for system-play
//   system-tweaks on|off                 pha0b's apply phcal tweaks? (live plays)
//   system-calibrate <arm|nod|rattle|danger|cube> <1|2> [reps] [--hold S] [--speed N] [--cycles N] [--volume N] [dry]
//       phcal: one move alone on one Brobot (phcal_isolate_001.py)
// All start gopod_probes/tools/system_play_001.sh, which hands the run to the
// operator's user like the song-start intents; the engine checks in with
// Goverlord, so one run at a time still holds.

var gopodPlaySwitches = struct {
	sync.Mutex
	gap    string // "" = the song's own
	tweaks bool
}{}

func gopodPlayScript() string {
	return filepath.Join(gopodChipperRoot(), "gopod_probes", "tools", "system_play_001.sh")
}

func gopodPlayStart(env []string, args ...string) error {
	cmd := exec.Command(gopodPlayScript(), args...)
	cmd.Env = append(os.Environ(), env...)
	gopodAsOperator(cmd)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%v %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// gopodPlayStep: a step id, or its number (1 = the first step), in the song.
func gopodPlayStep(dir, which string) (string, error) {
	steps, _, err := gopodSystemSetKnobsFile(dir)
	if err != nil {
		return "", err
	}
	if n, err := strconv.Atoi(which); err == nil {
		if n < 1 || n > len(steps) {
			return "", fmt.Errorf("%s has steps 1 to %d", filepath.Base(dir), len(steps))
		}
		return steps[n-1].StepID, nil
	}
	if gopodSystemSetStepIndex(steps, which) < 0 {
		return "", fmt.Errorf("no step %s in %s (system-score %s lists them)", which, filepath.Base(dir), filepath.Base(dir))
	}
	return which, nil
}

func gopodSystemPlay(text string) string {
	f := strings.Fields(text)
	if len(f) < 2 {
		return gopodSystemText("system_play.system_play_song_from_step")
	}
	if gopodScriptPlaying.Load() || gopodRunActive() {
		return gopodBrobotsSettings.Session.BusyText
	}
	dir, err := gopodSystemSetSong(f[1])
	if err != nil {
		return err.Error()
	}
	var knobs struct {
		SongType string `json:"song_type"`
	}
	if data, err := os.ReadFile(filepath.Join(dir, "knobs.json")); err == nil {
		json.Unmarshal(data, &knobs)
	}
	if knobs.SongType == "interview" || knobs.SongType == "llm_warm" {
		return gopodSystemText("system_play.isn_t_a_golden_engine", "dir", filepath.Base(dir))
	}
	from, to, robot, live := "", "", "both", true
	for i := 2; i < len(f); i++ {
		w := strings.ToLower(f[i])
		next := ""
		if i+1 < len(f) {
			next = f[i+1]
		}
		switch w {
		case "from", "to":
			if next == "" {
				return gopodSystemText("system_play.system_play_needs_a_step", "w", w)
			}
			step, err := gopodPlayStep(dir, next)
			if err != nil {
				return "system-play: " + err.Error()
			}
			if w == "from" {
				from = step
			} else {
				to = step
			}
			i++
		case "robot":
			if next != "1" && next != "2" {
				return gopodSystemText("system_play.system_play_robot_1_or")
			}
			robot = next
			i++
		case "dry":
			live = false
		default:
			return gopodSystemText("system_play.system_play_system_play_song", "value", f[i])
		}
	}
	gopodPlaySwitches.Lock()
	gap, tweaks := gopodPlaySwitches.gap, gopodPlaySwitches.tweaks
	gopodPlaySwitches.Unlock()
	env := []string{"GOPOD_GOLDEN_SONG_DIR=" + dir, "GOPOD_GOLDEN_ROBOT_FILTER=" + robot}
	if from != "" {
		env = append(env, "GOPOD_GOLDEN_PLAYHEAD_FROM="+from)
	}
	if to != "" {
		if from == "" {
			steps, _, _ := gopodSystemSetKnobsFile(dir)
			if len(steps) > 0 {
				env = append(env, "GOPOD_GOLDEN_PLAYHEAD_FROM="+steps[0].StepID)
			}
		}
		env = append(env, "GOPOD_GOLDEN_PLAYHEAD_TO="+to)
	} else if from != "" {
		steps, _, _ := gopodSystemSetKnobsFile(dir)
		env = append(env, "GOPOD_GOLDEN_PLAYHEAD_TO="+steps[len(steps)-1].StepID)
	}
	if robot != "both" {
		env = append(env, "GOPOD_GOLDEN_ROBOT="+robot)
	}
	if gap != "" {
		env = append(env, "GOPOD_GOLDEN_REPORTER_GAP_OVERRIDE="+gap)
	}
	if live {
		env = append(env, "GOPOD_PLAY_LIVE=1")
	}
	if tweaks {
		env = append(env, "GOPOD_PLAY_TWEAKS=1")
	}
	if err := gopodPlayStart(env, "play"); err != nil {
		return "system-play: not started - " + err.Error()
	}
	logger.Println("GOPOD_SYSTEM_PLAY " + strings.Join(env, " "))
	what := "the whole song"
	if from != "" || to != "" {
		what = "from " + map[bool]string{true: from, false: "the start"}[from != ""] + " to " + map[bool]string{true: to, false: "the end"}[to != ""]
	}
	mode := map[bool]string{true: "live", false: "dry - no robots, nothing said"}[live]
	return fmt.Sprintf("Playing %s, %s, robot %s, %s. stop run stops it. (log: /tmp/gopod_system_play.log)", filepath.Base(dir), what, robot, mode)
}

func gopodSystemGap(text string) string {
	f := strings.Fields(strings.ToLower(text))
	gopodPlaySwitches.Lock()
	defer gopodPlaySwitches.Unlock()
	if len(f) > 1 {
		switch {
		case f[1] == "default":
			gopodPlaySwitches.gap = ""
		case f[1] == "off":
			gopodPlaySwitches.gap = "0"
		default:
			if v, err := strconv.ParseFloat(f[1], 64); err != nil || v < 0 {
				return gopodSystemText("system_play.system_gap_seconds_off_default")
			}
			gopodPlaySwitches.gap = f[1]
		}
	}
	now := map[bool]string{true: "the song's own", false: gopodPlaySwitches.gap + " s"}[gopodPlaySwitches.gap == ""]
	return gopodSystemText("system_play.reporter_gap_for_system_play", "now", now)
}

func gopodSystemTweaks(text string) string {
	f := strings.Fields(strings.ToLower(text))
	gopodPlaySwitches.Lock()
	defer gopodPlaySwitches.Unlock()
	if len(f) > 1 && (f[1] == "on" || f[1] == "off") {
		gopodPlaySwitches.tweaks = f[1] == "on"
	}
	return gopodSystemText("system_play.apply_phcal_tweaks_before_a", "value", map[bool]string{true: "on", false: "off"}[gopodPlaySwitches.tweaks])
}

func gopodSystemCalibrate(text string) string {
	f := strings.Fields(text)
	if len(f) < 3 {
		return gopodSystemText("system_play.system_calibrate_arm_nod_rattle")
	}
	if gopodScriptPlaying.Load() || gopodRunActive() {
		return gopodBrobotsSettings.Session.BusyText
	}
	live := true
	var args []string
	for _, w := range f[1:] {
		if strings.EqualFold(w, "dry") {
			live = false
			continue
		}
		if strings.ContainsAny(w, ";&|`$<>\\\"'") {
			return gopodSystemText("system_play.system_calibrate_words_numbers_and", "w", w)
		}
		args = append(args, w)
	}
	env := []string{"GOPOD_CALIBRATE_ARGS=" + strings.Join(args, " ")}
	if live {
		env = append(env, "GOPOD_PLAY_LIVE=1")
	}
	if err := gopodPlayStart(env, "calibrate"); err != nil {
		return "system-calibrate: not started - " + err.Error()
	}
	logger.Println("GOPOD_SYSTEM_CALIBRATE " + strings.Join(env, " "))
	return gopodSystemText("system_play.calibrating_its_values_tmp_gopod", "args", strings.Join(args, " "), "value", map[bool]string{true: "live", false: "dry"}[live])
}

// ## EOF
