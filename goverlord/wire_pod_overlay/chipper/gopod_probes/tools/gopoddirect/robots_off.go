// Robots off for the direct tools (2026-10-29, operator: "do all scripted runs";
// Wire-Pod's own side is ttr/gopod_robots_off.go). With robots_off.on in the
// settings folder, ConnectAndGrant reaches no robot: it hands back a stand-in.
// A say or an animation goes to Wire-Pod's /api-sdk/say_text, where robots off
// voices the say on the chat page and captures both, as for every song line;
// everything else (wheels, the robot's own speaker, the cube) is done at once
// and nothing moves.
package gopoddirect

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// RobotsOff: robots_off.on in the settings folder.
func RobotsOff() bool {
	home, _ := os.UserHomeDir()
	files, _ := filepath.Glob(filepath.Join(home, "crushn8r_git/GOPOD/goverlord/wire_pod_overlay/chipper/webroot/gopod_settings/*.json"))
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil || !strings.Contains(string(data), `"robots_off"`) {
			continue
		}
		var m struct {
			RobotsOff struct {
				On bool `json:"on"`
			} `json:"robots_off"`
		}
		if json.Unmarshal(data, &m) == nil {
			return m.RobotsOff.On
		}
	}
	return false
}

// robotsOffRobot: the stand-in robot for serial.
func robotsOffRobot(serial string) (*vector.Vector, error) {
	nowhere, err := vector.New(vector.WithSerialNo(serial), vector.WithToken("robots-off"), vector.WithTarget("127.0.0.1:9"))
	if err != nil {
		return nil, err
	}
	nowhere.Conn = &offConn{ExternalInterfaceClient: nowhere.Conn, serial: serial}
	return nowhere, nil
}

type offConn struct {
	vectorpb.ExternalInterfaceClient
	serial string
}

var offOK = &vectorpb.ResponseStatus{Code: vectorpb.ResponseStatus_RESPONSE_RECEIVED}

// wirePodSay: one say_text to Wire-Pod for this robot, waited for.
func (c *offConn) wirePodSay(ctx context.Context, text string) error {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("GOPOD_WIREPOD_BASE_URL")), "/")
	if base == "" {
		base = "http://localhost:8080"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/api-sdk/say_text", strings.NewReader(url.Values{"serial": {c.serial}, "text": {text}}.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := (&http.Client{Timeout: 3 * time.Minute}).Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *offConn) SayText(ctx context.Context, in *vectorpb.SayTextRequest, _ ...grpc.CallOption) (*vectorpb.SayTextResponse, error) {
	if err := c.wirePodSay(ctx, in.GetText()); err != nil {
		return nil, err
	}
	return &vectorpb.SayTextResponse{Status: offOK, State: vectorpb.SayTextResponse_FINISHED}, nil
}

func (c *offConn) PlayAnimation(ctx context.Context, in *vectorpb.PlayAnimationRequest, _ ...grpc.CallOption) (*vectorpb.PlayAnimationResponse, error) {
	if err := c.wirePodSay(ctx, "{{playAnimationWI||"+in.GetAnimation().GetName()+"}}"); err != nil {
		return nil, err
	}
	return &vectorpb.PlayAnimationResponse{Status: offOK, Animation: in.GetAnimation()}, nil
}

func (c *offConn) PlayAnimationTrigger(ctx context.Context, in *vectorpb.PlayAnimationTriggerRequest, _ ...grpc.CallOption) (*vectorpb.PlayAnimationResponse, error) {
	if err := c.wirePodSay(ctx, "{{playAnimationWI||"+in.GetAnimationTrigger().GetName()+"}}"); err != nil {
		return nil, err
	}
	return &vectorpb.PlayAnimationResponse{Status: offOK}, nil
}

func (c *offConn) DriveWheels(context.Context, *vectorpb.DriveWheelsRequest, ...grpc.CallOption) (*vectorpb.DriveWheelsResponse, error) {
	return &vectorpb.DriveWheelsResponse{Status: offOK}, nil
}

func (c *offConn) ConnectCube(context.Context, *vectorpb.ConnectCubeRequest, ...grpc.CallOption) (*vectorpb.ConnectCubeResponse, error) {
	return &vectorpb.ConnectCubeResponse{Status: offOK, Success: true}, nil
}

func (c *offConn) SetCubeLights(context.Context, *vectorpb.SetCubeLightsRequest, ...grpc.CallOption) (*vectorpb.SetCubeLightsResponse, error) {
	return &vectorpb.SetCubeLightsResponse{Status: offOK}, nil
}

func (c *offConn) BatteryState(context.Context, *vectorpb.BatteryStateRequest, ...grpc.CallOption) (*vectorpb.BatteryStateResponse, error) {
	return &vectorpb.BatteryStateResponse{Status: offOK, BatteryLevel: vectorpb.BatteryLevel_BATTERY_LEVEL_FULL, BatteryVolts: 4.1}, nil
}

func (c *offConn) BehaviorControl(ctx context.Context, _ ...grpc.CallOption) (vectorpb.ExternalInterface_BehaviorControlClient, error) {
	return &offControl{offStream: offStream{ctx: ctx}}, nil
}

// ExternalAudioStreamPlayback: the robot's own speaker - nothing plays.
func (c *offConn) ExternalAudioStreamPlayback(ctx context.Context, _ ...grpc.CallOption) (vectorpb.ExternalInterface_ExternalAudioStreamPlaybackClient, error) {
	return &offAudio{offStream{ctx: ctx}}, nil
}

type offStream struct{ ctx context.Context }

func (s *offStream) Header() (metadata.MD, error) { return metadata.MD{}, nil }
func (s *offStream) Trailer() metadata.MD         { return metadata.MD{} }
func (s *offStream) CloseSend() error             { return nil }
func (s *offStream) Context() context.Context     { return s.ctx }
func (s *offStream) SendMsg(any) error            { return nil }
func (s *offStream) RecvMsg(any) error            { return s.wait() }
func (s *offStream) wait() error {
	if s.ctx == nil {
		return errors.New("robots off")
	}
	<-s.ctx.Done()
	return s.ctx.Err()
}

type offControl struct {
	offStream
	granted bool
}

func (s *offControl) Send(*vectorpb.BehaviorControlRequest) error { return nil }
func (s *offControl) Recv() (*vectorpb.BehaviorControlResponse, error) {
	if !s.granted {
		s.granted = true
		return &vectorpb.BehaviorControlResponse{ResponseType: &vectorpb.BehaviorControlResponse_ControlGrantedResponse{ControlGrantedResponse: &vectorpb.ControlGrantedResponse{}}}, nil
	}
	return nil, s.wait()
}

type offAudio struct{ offStream }

func (s *offAudio) Send(*vectorpb.ExternalAudioStreamRequest) error { return nil }
func (s *offAudio) Recv() (*vectorpb.ExternalAudioStreamResponse, error) {
	return &vectorpb.ExternalAudioStreamResponse{}, nil
}
