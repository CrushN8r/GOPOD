// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_robots_off_test.go
package wirepod_ttr

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
)

// The tests run with the robots on, whatever the live switch says.
func init() { gopodRobotsOffOn.Store(false) }

// Robots off: nothing reaches the robot - control is granted, the battery reads
// full, moves are done, a say and an animation are captured, and the numpad's
// KP1/KP2 pick a voice.
func TestRobotsOff(t *testing.T) {
	old, oldSettings := GopodRobotsOff(), gopodBrobotsSettings
	defer func() { gopodRobotsOffOn.Store(old); gopodBrobotsSettings = oldSettings }()
	capture := filepath.Join(t.TempDir(), "capture.jsonl")
	gopodBrobotsSettings.RobotsOff.CaptureFile = capture

	real, _ := vector.New(vector.WithSerialNo(gopodBrobot1ESN), vector.WithToken("x"), vector.WithTarget("127.0.0.1:9"))
	robot := GopodRobotsOffConn(real)
	ctx0, cancel0 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel0()
	gopodRobotsOffOn.Store(false)
	if _, err := robot.Conn.MoveLift(ctx0, &vectorpb.MoveLiftRequest{}); err == nil {
		t.Fatal("robots on: the call goes to the robot (none here, so it fails)")
	}
	gopodRobotsOffOn.Store(true)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if b, err := robot.Conn.BatteryState(ctx, &vectorpb.BatteryStateRequest{}); err != nil || b.GetBatteryLevel() != vectorpb.BatteryLevel_BATTERY_LEVEL_FULL {
		t.Fatalf("battery: %v %v", b, err)
	}
	start, stop := make(chan bool), make(chan bool)
	go BControl(robot, ctx, start, stop)
	<-start
	if _, err := robot.Conn.MoveLift(ctx, &vectorpb.MoveLiftRequest{SpeedRadPerSec: 1}); err != nil {
		t.Fatal("move:", err)
	}
	if _, err := robot.Conn.PlayAnimation(ctx, &vectorpb.PlayAnimationRequest{Animation: &vectorpb.Animation{Name: "anim_test_01"}}); err != nil {
		t.Fatal("anim:", err)
	}
	var dry []string
	oldDry := gopodDryVoice
	gopodDryVoice = func(host, text string, show bool) { dry = append(dry, host+": "+text) }
	defer func() { gopodDryVoice = oldDry }()
	if r, err := robot.Conn.SayText(ctx, &vectorpb.SayTextRequest{Text: "robots off test line"}); err != nil || r.GetState() != vectorpb.SayTextResponse_FINISHED {
		t.Fatalf("say: %v %v", r, err)
	}
	data, _ := os.ReadFile(capture)
	for _, want := range []string{`"kind":"anim","what":"anim_test_01"`, `"kind":"say","what":"robots off test line"`, `"brobot":"Brobot 1"`, `"esn":"ESN01"`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("capture has no %s:\n%s", want, data)
		}
	}
	for _, row := range gopodVoiceRows() {
		if (row.Key == "KP1" || row.Key == "KP2") && (row.Bodied || !strings.HasPrefix(row.Host, "Brobot ") || row.Voice != gopodBrobotsSettings.RobotsOff.Voices[row.Host]) {
			t.Errorf("robots off: %s picks a voice, got %+v", row.Key, row)
		}
	}
	if gopodReporterVoice("Brobot 1") == "" {
		t.Error("robots off: Brobot 1 has a temp voice")
	}
	gopodRobotsOffOn.Store(false)
	if gopodReporterVoice("Brobot 1") != "" {
		t.Error("robots on: Brobot 1 has no page voice")
	}
	for _, row := range gopodVoiceRows() {
		if (row.Key == "KP1" || row.Key == "KP2") && !row.Bodied {
			t.Errorf("robots on: %s needs no voice, got %+v", row.Key, row)
		}
	}
}
