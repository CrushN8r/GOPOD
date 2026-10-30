// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_robots_off.go
package wirepod_ttr

// Robots off (2026-10-29, operator: "I want to show what the visitor's runs
// will look like * plus all audio a visitor would hear in the same room * the
// only things missing would be my voice & the vector robot speaker output *
// there are voices assigned in place of the brobots speaker speech * if I later
// replay with the brobots I want all the full replay there"; "go robots off";
// "Just capture the brobot safe say + playanimations"; "follow numpad for
// voices"). With robots_off.on, nothing reaches a Brobot: every robot Wire-Pod
// makes (kgsim, the SDK page the songs use, the chat hold, the battery watch)
// gets a stand-in connection. A Brobot's say is played on the chat page in the
// voice the numpad's voices panel gives his key (KP1/KP2, reporter_voices.hosts),
// and its words and every animation are written to the capture file
// (robots_off.capture_file, one JSON line each) for a replay with the Brobots
// later. Anything else a robot is asked fails as an unplugged robot would.

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// gopodRobotsOffOn: robots off, from robots_off.on at the start, then
// system-robots-off on|off.
var gopodRobotsOffOn atomic.Bool

func init() { gopodRobotsOffOn.Store(gopodBrobotsSettings.RobotsOff.On) }

// GopodRobotsOff: robots off now.
func GopodRobotsOff() bool { return gopodRobotsOffOn.Load() }

// GopodRobotsOffConn: every robot Wire-Pod makes gets the stand-in around its
// own connection; while robots are on, every call goes straight through to the
// robot, so switching takes effect at once, without a restart.
func GopodRobotsOffConn(robot *vector.Vector) *vector.Vector {
	if robot == nil {
		return robot
	}
	if _, ok := robot.Conn.(*gopodOffConn); ok {
		return robot
	}
	robot.Conn = &gopodOffConn{ExternalInterfaceClient: robot.Conn, esn: robot.Cfg.SerialNo}
	return robot
}

// gopodOffConn: the stand-in connection - the robot's own, with robots off
// standing in for its say, animations, moves, control and battery.
type gopodOffConn struct {
	vectorpb.ExternalInterfaceClient
	esn string
}

var gopodOffOK = &vectorpb.ResponseStatus{Code: vectorpb.ResponseStatus_RESPONSE_RECEIVED}

// SayText: the Brobot's say - his temp voice on the page (robots_off.voices,
// the numpad's KP1/KP2), waited for until heard, as the robot's own say waits.
func (c *gopodOffConn) SayText(ctx context.Context, in *vectorpb.SayTextRequest, opts ...grpc.CallOption) (*vectorpb.SayTextResponse, error) {
	if !GopodRobotsOff() {
		return c.ExternalInterfaceClient.SayText(ctx, in, opts...)
	}
	host := gopodBrobotName(c.esn)
	gopodRobotsOffCapture(c.esn, "say", in.GetText())
	// 2026-10-29 (operator: "wrong bubbles * there's templates!! * nothing
	// should change, only placeholder voices"): only the voice goes to the
	// page - the Brobot's bubble is his own, as with the robots; no line,
	// no marker of its own.
	page, seconds, ok := gopodReporterFile(host, gopodReporterSays(in.GetText()))
	if !ok {
		logger.Println("GOPOD_ROBOTS_OFF " + host + ": no voice for the say - check his voice in the numpad's voices panel: " + gopodHeardShort(in.GetText()))
	} else {
		gopodVoicePlay(host, in.GetText(), page, seconds, true)
	}
	return &vectorpb.SayTextResponse{Status: gopodOffOK, State: vectorpb.SayTextResponse_FINISHED}, nil
}

// PlayAnimation: captured, done at once.
func (c *gopodOffConn) PlayAnimation(ctx context.Context, in *vectorpb.PlayAnimationRequest, opts ...grpc.CallOption) (*vectorpb.PlayAnimationResponse, error) {
	if !GopodRobotsOff() {
		return c.ExternalInterfaceClient.PlayAnimation(ctx, in, opts...)
	}
	gopodRobotsOffCapture(c.esn, "anim", in.GetAnimation().GetName())
	return &vectorpb.PlayAnimationResponse{Status: gopodOffOK, Animation: in.GetAnimation()}, nil
}

// PlayAnimationTrigger: captured, done at once.
func (c *gopodOffConn) PlayAnimationTrigger(ctx context.Context, in *vectorpb.PlayAnimationTriggerRequest, opts ...grpc.CallOption) (*vectorpb.PlayAnimationResponse, error) {
	if !GopodRobotsOff() {
		return c.ExternalInterfaceClient.PlayAnimationTrigger(ctx, in, opts...)
	}
	gopodRobotsOffCapture(c.esn, "anim_trigger", in.GetAnimationTrigger().GetName())
	return &vectorpb.PlayAnimationResponse{Status: gopodOffOK}, nil
}

// BatteryState: a full battery, so nothing waits on a robot that isn't there.
func (c *gopodOffConn) BatteryState(ctx context.Context, in *vectorpb.BatteryStateRequest, opts ...grpc.CallOption) (*vectorpb.BatteryStateResponse, error) {
	if !GopodRobotsOff() {
		return c.ExternalInterfaceClient.BatteryState(ctx, in, opts...)
	}
	return &vectorpb.BatteryStateResponse{Status: gopodOffOK, BatteryLevel: vectorpb.BatteryLevel_BATTERY_LEVEL_FULL, BatteryVolts: 4.1}, nil
}

// Moves: done at once, nothing moves.
func (c *gopodOffConn) MoveHead(ctx context.Context, in *vectorpb.MoveHeadRequest, opts ...grpc.CallOption) (*vectorpb.MoveHeadResponse, error) {
	if !GopodRobotsOff() {
		return c.ExternalInterfaceClient.MoveHead(ctx, in, opts...)
	}
	return &vectorpb.MoveHeadResponse{Status: gopodOffOK}, nil
}

func (c *gopodOffConn) MoveLift(ctx context.Context, in *vectorpb.MoveLiftRequest, opts ...grpc.CallOption) (*vectorpb.MoveLiftResponse, error) {
	if !GopodRobotsOff() {
		return c.ExternalInterfaceClient.MoveLift(ctx, in, opts...)
	}
	return &vectorpb.MoveLiftResponse{Status: gopodOffOK}, nil
}

func (c *gopodOffConn) DriveWheels(ctx context.Context, in *vectorpb.DriveWheelsRequest, opts ...grpc.CallOption) (*vectorpb.DriveWheelsResponse, error) {
	if !GopodRobotsOff() {
		return c.ExternalInterfaceClient.DriveWheels(ctx, in, opts...)
	}
	return &vectorpb.DriveWheelsResponse{Status: gopodOffOK}, nil
}

// BehaviorControl: granted at once, let go on release.
func (c *gopodOffConn) BehaviorControl(ctx context.Context, opts ...grpc.CallOption) (vectorpb.ExternalInterface_BehaviorControlClient, error) {
	if !GopodRobotsOff() {
		return c.ExternalInterfaceClient.BehaviorControl(ctx, opts...)
	}
	return &gopodOffControl{gopodOffStream: gopodOffStream{ctx: ctx}}, nil
}

// EventStream: no events - it waits until it's let go.
func (c *gopodOffConn) EventStream(ctx context.Context, in *vectorpb.EventRequest, opts ...grpc.CallOption) (vectorpb.ExternalInterface_EventStreamClient, error) {
	if !GopodRobotsOff() {
		return c.ExternalInterfaceClient.EventStream(ctx, in, opts...)
	}
	return &gopodOffEvents{gopodOffStream{ctx: ctx}}, nil
}

// gopodOffStream: a stream with nothing behind it.
type gopodOffStream struct{ ctx context.Context }

func (s *gopodOffStream) Header() (metadata.MD, error) { return metadata.MD{}, nil }
func (s *gopodOffStream) Trailer() metadata.MD         { return metadata.MD{} }
func (s *gopodOffStream) CloseSend() error             { return nil }
func (s *gopodOffStream) Context() context.Context     { return s.ctx }
func (s *gopodOffStream) SendMsg(any) error            { return nil }
func (s *gopodOffStream) RecvMsg(any) error            { return s.wait() }
func (s *gopodOffStream) wait() error {
	if s.ctx == nil {
		return errors.New("robots off")
	}
	<-s.ctx.Done()
	return s.ctx.Err()
}

type gopodOffControl struct {
	gopodOffStream
	mu      sync.Mutex
	granted bool
}

func (s *gopodOffControl) Send(*vectorpb.BehaviorControlRequest) error { return nil }
func (s *gopodOffControl) Recv() (*vectorpb.BehaviorControlResponse, error) {
	s.mu.Lock()
	first := !s.granted
	s.granted = true
	s.mu.Unlock()
	if first {
		return &vectorpb.BehaviorControlResponse{ResponseType: &vectorpb.BehaviorControlResponse_ControlGrantedResponse{ControlGrantedResponse: &vectorpb.ControlGrantedResponse{}}}, nil
	}
	return nil, s.wait()
}

type gopodOffEvents struct{ gopodOffStream }

func (s *gopodOffEvents) Recv() (*vectorpb.EventResponse, error) { return nil, s.wait() }

// gopodRobotsOffCapture: one line in the capture file - when, which Brobot,
// what kind (say, anim, anim_trigger), what.
var gopodRobotsOffCaptureMu sync.Mutex

func gopodRobotsOffCapture(esn, kind, what string) {
	host := gopodBrobotName(esn)
	logger.Println("GOPOD_ROBOTS_OFF " + host + " " + kind + " " + gopodHeardShort(what))
	file := strings.TrimSpace(gopodBrobotsSettings.RobotsOff.CaptureFile)
	if file == "" {
		return
	}
	path := strings.ReplaceAll(file, "{date}", time.Now().Format("20060102"))
	if !filepath.IsAbs(path) {
		path = filepath.Join(gopodChipperRoot(), path)
	}
	line, _ := json.Marshal(map[string]string{"at": time.Now().Format(time.RFC3339Nano), "brobot": host, "esn": GopodESNLabel(esn), "kind": kind, "what": what})
	gopodRobotsOffCaptureMu.Lock()
	defer gopodRobotsOffCaptureMu.Unlock()
	_, uid, gid := gopodRecordRoot()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err == nil {
		gopodRecordChown(uid, gid, filepath.Dir(path))
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		logger.Println("GOPOD_ROBOTS_OFF capture not written: " + err.Error())
		return
	}
	defer f.Close()
	_, _ = f.Write(append(line, '\n'))
	gopodRecordChown(uid, gid, path)
}

// gopodSystemRobotsOff: system-robots-off [on|off] - robots off now, or switch
// it (saved, so it holds after a restart).
func gopodSystemRobotsOff(text string) string {
	f := strings.Fields(strings.ToLower(text))
	state := func() string {
		if GopodRobotsOff() {
			return "on"
		}
		return "off"
	}
	if len(f) < 2 {
		return gopodSystemText("robots_off.now", "state", state())
	}
	if f[1] != "on" && f[1] != "off" {
		return gopodSystemText("robots_off.how")
	}
	on := f[1] == "on"
	path := []string{"robots_off", "on"}
	if _, err := gopodJSONFileSet(gopodSettingsFileFor(path), path, map[bool]string{true: "true", false: "false"}[on]); err != nil {
		return gopodSystemText("robots_off.not_saved", "error", err.Error())
	}
	gopodBrobotsSettings.RobotsOff.On = on
	gopodRobotsOffOn.Store(on)
	return gopodSystemText("robots_off.switched", "state", state())
}

// gopodIsBodiedBrobot: host is a Brobot with a robot body (numpad_ptt.body_to_brobot).
func gopodIsBodiedBrobot(host string) bool {
	for _, brobot := range gopodBrobotsSettings.NumpadPTT.BodyToBrobot {
		if host == brobot {
			return true
		}
	}
	return false
}
