// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_system_robots.go
package wirepod_ttr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// The Brobots group (2026-10-19, operator: "Bring all of it in a new Brobots
// group in the menu"; gopod_notes/WIREPOD_PAGES_ABSORB_SURVEY_001.md): Wire-Pod's
// own bot-page controls as system commands, for Brobot 1, 2 or both (the last
// word; both when left out). Each calls Wire-Pod's own route on this machine
// (/api-sdk/..., the same the bot page calls), so nothing new talks to a robot.
//   system-robot stats | volume <0-5> | eyes <colour> | locale <xx-XX> |
//     units c|f | location <place> | timezone <zone> | say <text> | wake |
//     head|lift <speed> | wheels <left> <right> | camera   [1|2|both]
//   system-faces [list | rename <old> <new> | delete <name> | add <name>] [1|2]
//   system-photos [1|2]
//   system-version   Wire-Pod's commit, the newest, and which of GOPOD's nine files changed

func gopodRobotWho(f []string) ([]string, []string) {
	if len(f) > 0 {
		switch strings.ToLower(f[len(f)-1]) {
		case "1":
			return []string{gopodBrobot1ESN}, f[:len(f)-1]
		case "2":
			return []string{gopodBrobot2ESN}, f[:len(f)-1]
		case "both":
			return []string{gopodBrobot1ESN, gopodBrobot2ESN}, f[:len(f)-1]
		}
	}
	return []string{gopodBrobot1ESN, gopodBrobot2ESN}, f
}

// gopodWirePodCall: one of Wire-Pod's own /api-sdk routes, for a robot.
func gopodWirePodCall(route, esn string, params url.Values) (string, error) {
	if params == nil {
		params = url.Values{}
	}
	params.Set("serial", esn)
	resp, err := (&http.Client{Timeout: 15 * time.Second}).PostForm("http://127.0.0.1:"+vars.WebPort+"/api-sdk/"+route, params)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%d %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return strings.TrimSpace(string(body)), nil
}

// The robot's eye colours (Vector's own presets, in its order).
var gopodRobotEyeColors = []string{"teal", "orange", "yellow", "lime", "sapphire", "purple", "green"}

func gopodSystemRobot(text string) string {
	f := strings.Fields(text)
	if len(f) < 2 {
		return gopodSystemText("system_robots.system_brobot_stats_anims_word", "eye_colors", strings.Join(gopodRobotEyeColors, "|"))
	}
	who, rest := gopodRobotWho(f[1:])
	what := strings.ToLower(rest[0])
	arg := strings.TrimSpace(strings.Join(rest[1:], " "))
	if gopodScriptPlaying.Load() || gopodRunActive() {
		if what != "stats" && what != "camera" && what != "anims" && what != "beats" {
			return gopodBrobotsSettings.Session.BusyText
		}
	}
	var out []string
	// Each Brobot's call through gopodRobotCall: woken first, one call at a
	// time, run to its end; both Brobots at once.
	eachHow := func(keep bool, do func(esn string, robot *vector.Vector, ctx context.Context) string) {
		res := make([]string, len(who))
		var wg sync.WaitGroup
		for i, esn := range who {
			wg.Add(1)
			go func(i int, esn string) {
				defer wg.Done()
				res[i] = gopodBrobotName(esn) + ": " + gopodRobotCall(esn, keep, func(robot *vector.Vector, ctx context.Context) string { return do(esn, robot, ctx) })
			}(i, esn)
		}
		wg.Wait()
		out = append(out, res...)
	}
	each := func(do func(esn string) string) {
		eachHow(true, func(esn string, _ *vector.Vector, _ context.Context) string { return do(esn) })
	}
	call := func(route string, p url.Values) func(string) string {
		return func(esn string) string {
			got, err := gopodWirePodCall(route, esn, p)
			if err != nil {
				return "not done - " + err.Error()
			}
			// 2026-10-20 (operator: "in Brobot N: done, 'done' doesn't tell me
			// what was done"): what was set.
			if got == "" || got == "done" || route == "custom_eye_color" {
				if what == "wake" {
					return "woken"
				}
				return gopodSystemText("system_robots.set_to", "what", what, "arg", arg)
			}
			return got
		}
	}
	switch what {
	case "stats":
		each(func(esn string) string {
			got, err := gopodWirePodCall("get_robot_stats", esn, nil)
			if err != nil {
				return "no stats - " + err.Error()
			}
			var m map[string]any
			if json.Unmarshal([]byte(got), &m) != nil || len(m) == 0 {
				return gopodSystemText("system_robots.no_stats_yet_the_robot")
			}
			var keys []string
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			var parts []string
			for _, k := range keys {
				parts = append(parts, fmt.Sprintf("%s %v", k, m[k]))
			}
			return strings.Join(parts, ", ")
		})
	case "volume":
		v, err := strconv.Atoi(arg)
		if err != nil || v < 0 || v > 5 {
			return gopodSystemText("system_robots.system_brobot_volume_0_5")
		}
		each(call("volume", url.Values{"volume": {arg}}))
	case "eyes":
		n := -1
		for i, c := range gopodRobotEyeColors {
			if strings.EqualFold(c, arg) {
				n = i
			}
		}
		if n < 0 {
			return gopodSystemText("system_robots.system_brobot_eyes_1_2", "eye_colors", strings.Join(gopodRobotEyeColors, "|"))
		}
		each(call("eye_color", url.Values{"color": {strconv.Itoa(n)}}))
	case "locale":
		each(call("locale", url.Values{"locale": {arg}}))
	case "units":
		route := map[string]string{"c": "temp_c", "f": "temp_f"}[strings.ToLower(arg)]
		if route == "" {
			return gopodSystemText("system_robots.system_brobot_units_c_f")
		}
		each(call(route, nil))
	case "location":
		each(call("location", url.Values{"location": {arg}}))
	case "timezone":
		each(call("timezone", url.Values{"timezone": {arg}}))
	case "say":
		if arg == "" {
			return gopodSystemText("system_robots.system_brobot_say_text_1")
		}
		eachHow(true, func(esn string, robot *vector.Vector, ctx context.Context) string {
			resp, err := gopodLogSayText(robot, ctx, &vectorpb.SayTextRequest{
				Text: gopodApplyPronunciations(removeSpecialCharacters(arg)), UseVectorVoice: true, DurationScalar: 1})
			if err != nil {
				return "not said - " + err.Error()
			}
			return gopodSystemText("system_robots.said", "value", resp.GetState().String())
		})
	case "anim":
		if arg == "" {
			return gopodSystemText("system_robots.system_brobot_anim_name_1")
		}
		eachHow(true, func(esn string, robot *vector.Vector, ctx context.Context) string {
			if _, err := gopodLogPlayAnimation(robot, ctx, &vectorpb.PlayAnimationRequest{Animation: &vectorpb.Animation{Name: arg}, Loops: 1}); err != nil {
				return "not played - " + err.Error()
			}
			return gopodSystemText("system_robots.played", "arg", arg)
		})
	case "anims":
		// The robot's own animation list (Vector keeps them on the robot) - e.g.
		// "anims sleep" for every sleep one; any plays with "anim <name>".
		eachHow(true, func(esn string, robot *vector.Vector, ctx context.Context) string {
			resp, err := robot.Conn.ListAnimations(ctx, &vectorpb.ListAnimationsRequest{})
			if err != nil {
				return err.Error()
			}
			var names []string
			for _, a := range resp.GetAnimationNames() {
				if arg == "" || strings.Contains(strings.ToLower(a.GetName()), strings.ToLower(arg)) {
					names = append(names, a.GetName())
				}
			}
			sort.Strings(names)
			return fmt.Sprintf("%d animations: %s", len(names), strings.Join(names, ", "))
		})
	case "beats":
		var out2 []string
		fams := gopodBrobotsSettings.SleepBeats.Families
		var keys []string
		for k := range fams {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if arg != "" && !strings.EqualFold(arg, k) {
				continue
			}
			var names []string
			for _, b := range fams[k] {
				names = append(names, b.Alias)
			}
			out2 = append(out2, fmt.Sprintf("%s (%d): %s", k, len(names), strings.Join(names, ", ")))
		}
		return gopodSystemText("system_robots.system_brobot_beat_alias_1", "out2", strings.Join(out2, "\n"))
	case "beat":
		// 2026-10-19 (operator: "all the 100+ sleep commands"): one sleep beat -
		// a trigger or a clip (sleep_beats, from SLEEP_BENCH_ALIASES_001.md).
		name, kind := "", ""
		for _, list := range gopodBrobotsSettings.SleepBeats.Families {
			for _, b := range list {
				if strings.EqualFold(arg, b.Alias) || arg == b.Name {
					name, kind = b.Name, b.Kind
				}
			}
		}
		if name == "" {
			return gopodSystemText("system_robots.system_brobot_beat_alias_1_2")
		}
		eachHow(true, func(esn string, robot *vector.Vector, ctx context.Context) string {
			var err error
			if kind == "trigger" {
				_, err = robot.Conn.PlayAnimationTrigger(ctx, &vectorpb.PlayAnimationTriggerRequest{AnimationTrigger: &vectorpb.AnimationTrigger{Name: name}, Loops: 1})
			} else {
				_, err = gopodLogPlayAnimation(robot, ctx, &vectorpb.PlayAnimationRequest{Animation: &vectorpb.Animation{Name: name}, Loops: 1})
			}
			if err != nil {
				return "not played - " + err.Error()
			}
			return gopodSystemText("system_robots.played_2", "name", name)
		})
	case "cube":
		// 2026-10-19 (operator: "cube calls pop a wheelie, where's your cube,
		// etc..."): the robot's own cube behaviours (native_cube.intents), woken
		// first, control given back so his behaviour runs.
		intent := ""
		var names []string
		for _, c := range gopodBrobotsSettings.NativeCube.Intents {
			name := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(c.Intent, "intent_play_"), "intent_imperative_"), "intent_")
			names = append(names, name)
			if strings.EqualFold(arg, name) {
				intent = c.Intent
			}
			for _, ph := range c.Phrases {
				if strings.EqualFold(arg, ph) {
					intent = c.Intent
				}
			}
		}
		if intent == "" {
			return gopodSystemText("system_robots.system_brobot_cube_1_2", "names", strings.Join(names, "|"))
		}
		eachHow(false, func(esn string, robot *vector.Vector, ctx context.Context) string {
			if _, err := robot.Conn.AppIntent(ctx, &vectorpb.AppIntentRequest{Intent: intent}); err != nil {
				return "not done - " + err.Error()
			}
			return gopodSystemText("system_robots.doing", "arg", arg)
		})
	case "wake":
		each(call("trigger_wake_word", nil))
	case "head", "lift", "wheels":
		// A short nudge, woken first: the move, a moment, stop.
		g := strings.Fields(arg)
		if (what == "wheels" && len(g) != 2) || (what != "wheels" && len(g) != 1) {
			return gopodSystemText("system_robots.system_brobot_head_lift_speed")
		}
		var v []float64
		for _, x := range g {
			n, err := strconv.ParseFloat(x, 64)
			if err != nil {
				return gopodSystemText("system_robots.system_brobot_numbers_only", "what", what)
			}
			v = append(v, n)
		}
		eachHow(true, func(esn string, robot *vector.Vector, ctx context.Context) string {
			var err error
			switch what {
			case "head":
				_, err = robot.Conn.MoveHead(ctx, &vectorpb.MoveHeadRequest{SpeedRadPerSec: float32(v[0])})
				time.Sleep(600 * time.Millisecond)
				robot.Conn.MoveHead(ctx, &vectorpb.MoveHeadRequest{SpeedRadPerSec: 0})
			case "lift":
				_, err = robot.Conn.MoveLift(ctx, &vectorpb.MoveLiftRequest{SpeedRadPerSec: float32(v[0])})
				time.Sleep(600 * time.Millisecond)
				robot.Conn.MoveLift(ctx, &vectorpb.MoveLiftRequest{SpeedRadPerSec: 0})
			default:
				_, err = robot.Conn.DriveWheels(ctx, &vectorpb.DriveWheelsRequest{LeftWheelMmps: float32(v[0]), RightWheelMmps: float32(v[1])})
				time.Sleep(600 * time.Millisecond)
				robot.Conn.DriveWheels(ctx, &vectorpb.DriveWheelsRequest{})
			}
			if err != nil {
				return "not moved - " + err.Error()
			}
			return "moved"
		})
	case "camera":
		// 2026-10-19 (operator: "how to add in robot video?"): Wire-Pod's own
		// live camera stream (/cam-stream), shown as a picture in the chat (the
		// menu's results while it's open); it stops when its bubble is gone.
		var paths []string
		for _, esn := range who {
			paths = append(paths, "/cam-stream?serial="+esn)
			out = append(out, gopodBrobotName(esn)+": camera on")
		}
		GopodChatFeedAdd(gopodBrobotsSettings.TopicPictures.Role, gopodMenuPictureMark+strings.Join(paths, "|"))
	default:
		return gopodSystemText("system_robots.system_robot_system_brobot_alone", "what", what)
	}
	return gopodJoinLines(out)
}

func gopodSystemFaces(text string) string {
	f := strings.Fields(text)
	who, rest := gopodRobotWho(f[1:])
	if len(who) > 1 {
		who = who[:1]
	}
	esn := who[0]
	what := "list"
	if len(rest) > 0 {
		what = strings.ToLower(rest[0])
	}
	var got string
	var err error
	gopodRobotCall(esn, true, func(*vector.Vector, context.Context) string {
		got, err = gopodWirePodCall("get_faces", esn, nil)
		return ""
	})
	if err != nil {
		return "system-faces: " + err.Error()
	}
	var faces []struct {
		Name   string `json:"name"`
		FaceID int    `json:"face_id"`
	}
	json.Unmarshal([]byte(got), &faces)
	find := func(name string) int {
		for _, fc := range faces {
			if strings.EqualFold(fc.Name, name) {
				return fc.FaceID
			}
		}
		return -1
	}
	switch what {
	case "list":
		var names []string
		for _, fc := range faces {
			names = append(names, fc.Name)
		}
		if len(names) == 0 {
			return gopodSystemText("system_robots.knows_no_faces_yet_system", "brobot", gopodBrobotName(esn), "value", map[string]string{gopodBrobot1ESN: "1", gopodBrobot2ESN: "2"}[esn])
		}
		return gopodSystemText("system_robots.knows", "brobot", gopodBrobotName(esn), "names", strings.Join(names, ", "))
	case "rename":
		if len(rest) != 3 {
			return gopodSystemText("system_robots.system_faces_rename_old_new")
		}
		id := find(rest[1])
		if id < 0 {
			return gopodSystemText("system_robots.no_face", "arg1", rest[1])
		}
		_, err = gopodWirePodCall("rename_face", esn, url.Values{"id": {strconv.Itoa(id)}, "oldname": {rest[1]}, "newname": {rest[2]}})
	case "delete":
		if len(rest) != 2 {
			return gopodSystemText("system_robots.system_faces_delete_name_1")
		}
		id := find(rest[1])
		if id < 0 {
			return gopodSystemText("system_robots.no_face", "arg1", rest[1])
		}
		_, err = gopodWirePodCall("delete_face", esn, url.Values{"id": {strconv.Itoa(id)}})
	case "add":
		if len(rest) != 2 {
			return gopodSystemText("system_robots.system_faces_add_name_1", "brobot", gopodBrobotName(esn))
		}
		_, err = gopodWirePodCall("add_face", esn, url.Values{"name": {rest[1]}})
		if err == nil {
			return gopodSystemText("system_robots.is_learning_s_face_look", "brobot", gopodBrobotName(esn), "arg1", rest[1])
		}
	default:
		return gopodSystemText("system_robots.system_faces_list_rename_old")
	}
	if err != nil {
		return "system-faces: not done - " + err.Error()
	}
	return gopodSystemText("system_robots.system_faces_done", "what", what)
}

// gopodSystemPhotos: a Brobot's photos as thumbnails (a picture bubble, the magnifier on each).
func gopodSystemPhotos(text string) string {
	who, _ := gopodRobotWho(strings.Fields(text)[1:])
	var out []string
	for _, esn := range who {
		var got string
		var err error
		gopodRobotCall(esn, true, func(*vector.Vector, context.Context) string {
			got, err = gopodWirePodCall("get_image_ids", esn, nil)
			return ""
		})
		if err != nil {
			out = append(out, gopodBrobotName(esn)+": "+err.Error())
			continue
		}
		var ids []uint32
		json.Unmarshal([]byte(got), &ids)
		var paths []string
		for i := len(ids) - 1; i >= 0 && len(paths) < 12; i-- {
			paths = append(paths, fmt.Sprintf("/api-sdk/get_image_thumb?serial=%s&id=%d", esn, ids[i]))
		}
		if len(paths) == 0 {
			out = append(out, gopodBrobotName(esn)+": no photos")
			continue
		}
		GopodChatFeedAdd(gopodBrobotsSettings.TopicPictures.Role, gopodMenuPictureMark+strings.Join(paths, "|"))
		out = append(out, fmt.Sprintf("%s: %d photos (newest %d shown)", gopodBrobotName(esn), len(ids), len(paths)))
	}
	return gopodJoinLines(out)
}

// gopodSystemVersion: Wire-Pod's commit, the newest on its remote, and which of
// the native files GOPOD's overlay replaces changed between them (the update guide:
// WIRED-POD.md, gopod_notes/WIREPOD_PAGES_ABSORB_SURVEY_001.md).
func gopodSystemVersion() string {
	git := func(args ...string) string {
		cmd := gopodAsOperator(exec.Command("git", append([]string{"-c", "safe.directory=*", "-C", gopodHomePath("~/wire-pod")}, args...)...))
		out, _ := cmd.Output()
		return strings.TrimSpace(string(out))
	}
	here := git("rev-parse", "--short", "HEAD")
	remote := git("ls-remote", "origin", "HEAD")
	newest := ""
	if f := strings.Fields(remote); len(f) > 0 && len(f[0]) >= 7 {
		newest = f[0][:7]
	}
	out := []string{"Wire-Pod installed: " + here}
	switch {
	case newest == "":
		out = append(out, "Newest: not reachable now.")
	case strings.HasPrefix(newest, here) || strings.HasPrefix(here, newest):
		out = append(out, "Newest: "+newest+" - up to date.")
	default:
		out = append(out, "Newest: "+newest+" - an update is waiting. Update with: git -C ~/wire-pod pull (never push), then each changed file below into GOPOD's copy, then deploy_overlay_build.sh, wpr 2, system-test-all.")
		git("fetch", "--quiet", "origin")
		var changed []string
		for _, p := range gopodOverlayNativeFiles {
			if git("diff", "--stat", here, newest, "--", "chipper/"+p) != "" {
				changed = append(changed, p)
			}
		}
		if len(changed) == 0 {
			out = append(out, "None of GOPOD's nine files changed upstream - pull, deploy, done.")
		} else {
			out = append(out, "Changed upstream in GOPOD's files: "+strings.Join(changed, ", "))
		}
	}
	return gopodJoinLines(out)
}

// gopodOverlayNativeFiles: the native files GOPOD's overlay replaces (WIRED-POD.md "Stay thin").
var gopodOverlayNativeFiles = []string{
	"pkg/wirepod/ttr/kgsim.go", "pkg/wirepod/ttr/kgsim_cmds.go", "pkg/wirepod/ttr/kgsim_interrupt.go",
	"pkg/wirepod/ttr/weather.go", "pkg/wirepod/ttr/matchIntentSend.go", "pkg/wirepod/sdkapp/server.go",
	"pkg/wirepod/config-ws/webserver.go", "pkg/wirepod/sdkapp/bcassume.go", "pkg/wirepod/sdkapp/robot.go",
}

// ## EOF

// gopodSystemAnim (2026-10-21, operator: "How doable to test playAnimationWI via
// GOPOD keyboard input * test text to enter: "{{playAnimationWI||angry}} wait 1,
// 2, 3 ..." * can I have the makings (WIP) of an animationWI tester, in system
// menu?"): the text through Wire-Pod's own say route - the path the song engine's
// emotion beats use - with control taken first (the 1.5 s settle), then given
// back. "system-anim list": the animation tokens. The clip played and its time
// are in the journal (GOPOD_ROBOT_CMD lines, with the battery).
func gopodSystemAnim(text string) string {
	rest := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(text), strings.Fields(text)[0]))
	if rest == "" {
		return gopodSystemText("system_robots.system_anim_text_with_1")
	}
	if strings.EqualFold(rest, "list") {
		return gopodSystemText("system_robots.animation_tokens", "value", strings.Join(verifiedTokenNames(loadedAnimationVocab), ", "))
	}
	if gopodScriptPlaying.Load() || gopodRunActive() {
		return gopodBrobotsSettings.Session.BusyText
	}
	who := []string{gopodBrobot1ESN}
	f := strings.Fields(rest)
	switch strings.ToLower(f[len(f)-1]) {
	case "1":
		rest = strings.TrimSpace(strings.TrimSuffix(rest, f[len(f)-1]))
	case "2":
		who, rest = []string{gopodBrobot2ESN}, strings.TrimSpace(strings.TrimSuffix(rest, f[len(f)-1]))
	case "both":
		who, rest = []string{gopodBrobot1ESN, gopodBrobot2ESN}, strings.TrimSpace(strings.TrimSuffix(rest, f[len(f)-1]))
	}
	var out []string
	for _, esn := range who {
		lock := gopodRobotCallLock(esn)
		lock.Lock()
		start := time.Now()
		_, aerr := gopodWirePodCall("assume_behavior_control", esn, url.Values{"priority": {"high"}})
		time.Sleep(1500 * time.Millisecond)
		got, err := gopodWirePodCall("say_text", esn, url.Values{"text": {rest}})
		took := time.Since(start).Round(100 * time.Millisecond)
		gopodWirePodCall("release_behavior_control", esn, nil)
		lock.Unlock()
		line := fmt.Sprintf("%s: %s in %s", gopodBrobotName(esn), map[bool]string{true: "done", false: "failed"}[err == nil], took)
		if aerr != nil {
			line += " (control: " + aerr.Error() + ")"
		}
		if err != nil {
			line += " - " + err.Error()
		} else if got != "" && got != "done" {
			line += " - " + gopodShort(got, 80)
		}
		out = append(out, line+gopodBatteryNote(esn))
	}
	return gopodJoinLines(append(out, "the clip and its time: system-find GOPOD_ROBOT_CMD"))
}
