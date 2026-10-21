// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_robot_cmd_log.go
package wirepod_ttr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

// Every animation, speech line, arm/head move and behavior-control request sent to
// a Brobot is logged when it goes out ("send") and when the robot answers ("done",
// with how long it took and any error). Operator 2026-10-07, after Brobot 2 froze
// mid-answer: "please update to log animations and other robot control calls".
// A call that never gets a "done" line is the one the robot was stuck on.

func gopodRobotCmdSend(robot *vector.Vector, what string) time.Time {
	logger.Println("GOPOD_ROBOT_CMD " + gopodBrobotName(robot.Cfg.SerialNo) + " send " + what + gopodBatteryNote(robot.Cfg.SerialNo))
	return time.Now()
}

// gopodBatteryNote (2026-10-21, operator: "with every robot log entry include
// battery power level"): the Brobot's last battery reading (battery watch) and how
// old it is, e.g. " [4.00 V, FULL, 2m ago]".
func gopodBatteryNote(esn string) string {
	gopodBatteryLast.Lock()
	defer gopodBatteryLast.Unlock()
	text, at := gopodBatteryLast.text[esn], gopodBatteryLast.at[esn]
	if text == "" {
		return " [battery: no reading yet]"
	}
	return fmt.Sprintf(" [%s, %s ago]", text, time.Since(at).Round(time.Minute))
}

func gopodRobotCmdDone(robot *vector.Vector, what string, sent time.Time, err error) {
	line := fmt.Sprintf("GOPOD_ROBOT_CMD %s done %s took=%dms", gopodBrobotName(robot.Cfg.SerialNo), what, time.Since(sent).Milliseconds())
	if err != nil {
		line += " err=" + err.Error()
	}
	line += gopodBatteryNote(robot.Cfg.SerialNo)
	logger.Println(line)
}

func gopodLogPlayAnimation(robot *vector.Vector, ctx context.Context, req *vectorpb.PlayAnimationRequest) (*vectorpb.PlayAnimationResponse, error) {
	what := "anim " + req.GetAnimation().GetName()
	esn := robot.Cfg.SerialNo
	gopodAnimationPlaying(esn, 1)
	defer gopodAnimationPlaying(esn, -1)
	sent := gopodRobotCmdSend(robot, what)
	resp, err := robot.Conn.PlayAnimation(ctx, req)
	gopodRobotCmdDone(robot, what, sent, err)
	return resp, err
}

// The animations playing on each robot (gopodAnimationWaitDone).
var gopodAnimations = struct {
	sync.Mutex
	playing map[string]int
}{playing: map[string]int{}}

func gopodAnimationPlaying(esn string, delta int) {
	gopodAnimations.Lock()
	defer gopodAnimations.Unlock()
	if gopodAnimations.playing[esn] += delta; gopodAnimations.playing[esn] < 0 {
		gopodAnimations.playing[esn] = 0 // a late done after the hang guard cleared it
	}
}

// gopodAnimationWaitDone (operator 2026-10-09: "no new animations until enough
// time has lapsed for playAnimationWI run completion, or a 'done' message is
// received"): before a reply's next animation, wait until the robot has reported
// the one before it done (its PlayAnimation call has returned, failed or not) -
// never past chat_moves.animation_done_most_seconds, the hang guard.
func gopodAnimationWaitDone(esn string) {
	deadline := time.Now().Add(gopodSecondsToDuration(gopodBrobotsSettings.ChatMoves.AnimationDoneMostSeconds))
	waited := time.Now()
	for {
		gopodAnimations.Lock()
		playing := gopodAnimations.playing[esn]
		if playing > 0 && !time.Now().Before(deadline) {
			// The hang guard: clear the flag so the next animation isn't held
			// too (the RPC itself is left to Wire-Pod).
			gopodAnimations.playing[esn] = 0
			gopodAnimations.Unlock()
			logger.Println("GOPOD_ANIMATION_WAIT " + gopodBrobotName(esn) + " no done within animation_done_most_seconds - flag cleared")
			break
		}
		gopodAnimations.Unlock()
		if playing <= 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if w := time.Since(waited); w >= 100*time.Millisecond {
		logger.Println(fmt.Sprintf("GOPOD_ANIMATION_WAIT %s waited %dms for the last animation's done", gopodBrobotName(esn), w.Milliseconds()))
	}
}

func gopodLogSayText(robot *vector.Vector, ctx context.Context, req *vectorpb.SayTextRequest) (*vectorpb.SayTextResponse, error) {
	what := fmt.Sprintf("say %q", req.GetText())
	sent := gopodRobotCmdSend(robot, what)
	resp, err := robot.Conn.SayText(ctx, req)
	if err == nil && resp != nil {
		what += " state=" + resp.GetState().String()
	}
	gopodRobotCmdDone(robot, what, sent, err)
	return resp, err
}

func gopodLogMoveLift(robot *vector.Vector, ctx context.Context, req *vectorpb.MoveLiftRequest) (*vectorpb.MoveLiftResponse, error) {
	what := fmt.Sprintf("lift speed=%.2f", req.GetSpeedRadPerSec())
	sent := gopodRobotCmdSend(robot, what)
	resp, err := robot.Conn.MoveLift(ctx, req)
	gopodRobotCmdDone(robot, what, sent, err)
	return resp, err
}

func gopodLogMoveHead(robot *vector.Vector, ctx context.Context, req *vectorpb.MoveHeadRequest) (*vectorpb.MoveHeadResponse, error) {
	what := fmt.Sprintf("head speed=%.2f", req.GetSpeedRadPerSec())
	sent := gopodRobotCmdSend(robot, what)
	resp, err := robot.Conn.MoveHead(ctx, req)
	gopodRobotCmdDone(robot, what, sent, err)
	return resp, err
}

// gopodLogBControl is BControl with the request and the grant logged. The grant
// is passed on to start as BControl would send it.
func gopodLogBControl(robot *vector.Vector, ctx context.Context, start, stop chan bool) {
	sent := gopodRobotCmdSend(robot, "control")
	granted := make(chan bool)
	go func() {
		<-granted
		gopodRobotCmdDone(robot, "control granted", sent, nil)
		start <- true
	}()
	BControl(robot, ctx, granted, stop)
}

// gopodAnimationContext (operator 2026-10-09: "can 'Wire-Pod's queue can still
// block that robot's next animation until the call fails' be addressed?"): the
// context for the native PlayAnimation calls (DoPlayAnimation/WI, kgsim_cmds.go).
// The stock call has none, so a robot that drops mid-clip (live 14:15, 58 s) held
// Wire-Pod's own animation queue until the connection failed. With
// chat_moves.animation_done_most_seconds as its deadline, the call gives up, the
// native StopAnim_Queue runs and the queue is released. The robot's clip itself
// isn't stopped; only the wait for it ends.
func gopodAnimationContext() (context.Context, context.CancelFunc) {
	most := gopodSecondsToDuration(gopodBrobotsSettings.ChatMoves.AnimationDoneMostSeconds)
	if most <= 0 {
		return context.WithCancel(context.Background())
	}
	return context.WithTimeout(context.Background(), most)
}

// gopodLogBattery (operator 2026-10-09: "add in the logs * battery power level per
// brobot response * I have a hunch it may be a power drawing/draining issue"): one
// GOPOD_BATTERY line after each Brobot reply - volts, level, charging, on the
// charger. In the background, so the reply never waits on it.
func gopodLogBattery(robot *vector.Vector) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		name := gopodBrobotName(robot.Cfg.SerialNo)
		b, err := robot.Conn.BatteryState(ctx, &vectorpb.BatteryStateRequest{})
		if err != nil {
			logger.Println("GOPOD_BATTERY " + name + " unknown: " + err.Error())
			return
		}
		logger.Println(fmt.Sprintf("GOPOD_BATTERY %s volts=%.2f level=%s charging=%t on_charger=%t",
			name, b.GetBatteryVolts(), strings.TrimPrefix(b.GetBatteryLevel().String(), "BATTERY_LEVEL_"),
			b.GetIsCharging(), b.GetIsOnChargerPlatform()))
	}()
}

// gopodSoundOnPage (operator 2026-10-09: "keep but pause playing through the
// brobot, too much trouble now * instead make it play through the web page like
// the reporter voices" - "same with danger wav"): with sounds_on "page", a sound
// file is copied once into the page's voice folder and put in the chat feed as a
// reporter-voice line, so the page plays it in its queue. True when it did.
// "robot" keeps the robot's own playback.
func gopodSoundOnPage(path string) bool {
	if gopodBrobotsSettings.SoundsOn != "page" {
		return false
	}
	GopodChatFeedAdd(gopodBrobotsSettings.ReporterVoices.Role, "/"+gopodBrobotsSettings.ReporterVoices.AudioFolder+"/"+filepath.Base(path))
	logger.Println("GOPOD_SOUND_PAGE " + filepath.Base(path))
	return true
}

// gopodSoundsToPage: at Wire-Pod's start, every WAV in the overlay's sounds/ is
// copied into the page's voice folder (webroot/<reporter_voices.audio_folder>/, Wire-Pod's
// own, not tracked), so the page can play the rattle and the danger sound - a
// song running as the desktop user can't write there (live 16:23).
func gopodSoundsToPage() {
	folder := gopodReporterAudioDir()
	files, _ := filepath.Glob(filepath.Join(gopodChipperRoot(), "sounds", "*.wav"))
	// 2026-10-10: every song's own soundtrack (its knobs.json "audio"), found by
	// the song, wherever its folder is (gopod_song_folders.go).
	if found := gopodSongAudioFiles(); len(found) > 0 {
		files = append(files, found...)
	}
	// page_sound_files: any other sound, relative to the overlay's chipper folder.
	for _, f := range gopodBrobotsSettings.PageSoundFiles {
		if !filepath.IsAbs(f) {
			f = filepath.Join(gopodChipperRoot(), f)
		}
		if _, err := os.Stat(f); err == nil {
			files = append(files, f)
		} else {
			logger.Println("GOPOD_SOUND_PAGE not found (private media not on this machine?): " + f)
		}
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err == nil {
			_ = os.MkdirAll(folder, 0o755)
			err = os.WriteFile(filepath.Join(folder, filepath.Base(f)), data, 0o644)
		}
		if err != nil {
			logger.Println("GOPOD_SOUND_PAGE copy failed: " + err.Error())
		}
	}
}
