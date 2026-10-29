// direct_sdk_robot_info_001.go - read-only robot info snapshot, directly
// over the vendored fforchino/vector-go-sdk, bypassing Wire-Pod's own HTTP
// SDK app entirely. Built 2026-08-10 for the Brobot 2 (Pip) intermittent-
// unresponsiveness investigation (BROBOT_2_INSTABILITY_EXTERNAL_AI_BRIEF_001.md) -
// the concrete next step that report's own external-AI research confirmed
// is safe: call VersionState/ProtocolVersion to settle the long-open
// "is this a Vector 1.0 or 2.0 firmware unit" question
// (VECTORX_2_0_SDK_DIVERGENCE_001.md, 2026-07-05, concluded "not established
// by anything on disk" - botSdkInfo.json/jdocs.json carry no version field
// for either robot's ESN), plus BatteryState (volts/level/charging) captured
// under the same conditions, per that investigation's own recommended
// diagnostic order.
//
// Confirmed via the vendored proto source, not guessed:
//   - VersionStateResponse: os_version (string), engine_build_id (string) -
//     messages.proto:951-957. This is the direct answer to the firmware
//     question; nothing else on disk in this project carries it.
//   - ProtocolVersionResponse: result (UNSUPPORTED/SUCCESS enum),
//     host_version (int64) - shared.proto:29-39. Request takes
//     client_version/min_host_version int64 fields; this tool sends 5 for
//     both (vectorpb.ProtocolVersion_PROTOCOL_VERSION_CURRENT, per
//     external_interface.proto:36-39 - "Current version (incremented on
//     2019.03.12...)"), the same value a normal SDK connection negotiates.
//   - BatteryStateResponse: status, battery_level, battery_volts,
//     is_charging, is_on_charger_platform, suggested_charger_sec,
//     cube_battery - messages.proto:930-938.
//
// None of these three RPCs require BehaviorControl - confirmed by reading
// Wire-Pod's own handler (chipper/pkg/wirepod/sdkapp/server.go's
// get_battery case calls robot.Conn.BatteryState directly, no assume/
// release around it) and by the RPCs' own nature (read-only info queries,
// not behavior/motor actions). So this tool is deliberately simpler than
// direct_sdk_robot_sleep_001.go/direct_sdk_brobots_ready_001.go - just
// vector.NewWP(serial), fire the three calls, print, done. No
// gopoddirect.ConnectAndGrant/Release, no Barrier (nothing here needs
// cross-robot sync - these are independent read-only queries, concurrency
// below is only for speed, not a shared timed event).
//
// Usage: direct_sdk_robot_info_001 <label1>:<serial1> [<label2>:<serial2> ...]
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/gopod_probes/tools/gopoddirect"
)

const rpcTimeoutSeconds = 20.0

type robotResult struct {
	Label  string
	Serial string
	Start  time.Time
	End    time.Time

	ConnectErr string

	OSVersion     string
	EngineBuildID string
	VersionErr    string

	ProtocolResult      string
	ProtocolHostVersion int64
	ProtocolErr         string

	BatteryStatus       string
	BatteryLevel        string
	BatteryVolts        float32
	IsCharging          bool
	IsOnChargerPlatform bool
	BatteryErr          string
}

func fetchInfo(serial, label string) (res robotResult) {
	// Named return value - a bare `return res` on a non-named return
	// copies res's value BEFORE deferred functions run, so a deferred
	// `res.End = time.Now()` would silently mutate a copy nobody sees,
	// leaving End at its zero value forever (confirmed live: this exact
	// bug produced `elapsed=-9223372036.855s`, the gap between Go's zero
	// time and Start). Naming the return value fixes it - the defer below
	// now mutates the actual value that gets returned.
	res = robotResult{Serial: serial, Label: label, Start: time.Now()}
	defer func() { res.End = time.Now() }()

	// Same "a hang here must not hang forever" reasoning every other tool
	// in this family already documents (direct_sdk_robot_sleep_001.go's own
	// comment on this exact gap) - vector.NewWP itself has no internal
	// deadline, and Robot 2's own observed failure mode (this session,
	// BROBOT_2_INSTABILITY_EXTERNAL_AI_BRIEF_001.md) includes multi-second-
	// to-tens-of-seconds stalls, so each RPC call below gets its own
	// bounded context rather than sharing one that could be exhausted by
	// an earlier slow/hung call.
	robot, err := vector.NewWP(gopoddirect.RobotSerial(serial))
	if err != nil {
		res.ConnectErr = err.Error()
		return res
	}

	versionCtx, versionCancel := context.WithTimeout(context.Background(), time.Duration(rpcTimeoutSeconds*float64(time.Second)))
	defer versionCancel()
	if resp, err := robot.Conn.VersionState(versionCtx, &vectorpb.VersionStateRequest{}); err != nil {
		res.VersionErr = err.Error()
	} else {
		res.OSVersion = resp.GetOsVersion()
		res.EngineBuildID = resp.GetEngineBuildId()
	}

	protoCtx, protoCancel := context.WithTimeout(context.Background(), time.Duration(rpcTimeoutSeconds*float64(time.Second)))
	defer protoCancel()
	if resp, err := robot.Conn.ProtocolVersion(protoCtx, &vectorpb.ProtocolVersionRequest{
		ClientVersion:  int64(vectorpb.ProtocolVersion_PROTOCOL_VERSION_CURRENT),
		MinHostVersion: int64(vectorpb.ProtocolVersion_PROTOCOL_VERSION_CURRENT),
	}); err != nil {
		res.ProtocolErr = err.Error()
	} else {
		res.ProtocolResult = resp.GetResult().String()
		res.ProtocolHostVersion = resp.GetHostVersion()
	}

	battCtx, battCancel := context.WithTimeout(context.Background(), time.Duration(rpcTimeoutSeconds*float64(time.Second)))
	defer battCancel()
	if resp, err := robot.Conn.BatteryState(battCtx, &vectorpb.BatteryStateRequest{}); err != nil {
		res.BatteryErr = err.Error()
	} else {
		res.BatteryStatus = resp.GetStatus().String()
		res.BatteryLevel = resp.GetBatteryLevel().String()
		res.BatteryVolts = resp.GetBatteryVolts()
		res.IsCharging = resp.GetIsCharging()
		res.IsOnChargerPlatform = resp.GetIsOnChargerPlatform()
	}

	return res
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: direct_sdk_robot_info_001 <label1>:<serial1> [<label2>:<serial2> ...]")
		os.Exit(2)
	}
	specs := os.Args[1:]

	results := make([]robotResult, len(specs))
	var wg sync.WaitGroup
	for i, spec := range specs {
		parts := strings.SplitN(spec, ":", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "bad spec %q, want label:serial\n", spec)
			os.Exit(2)
		}
		label, serial := parts[0], parts[1]
		wg.Add(1)
		go func(i int, label, serial string) {
			defer wg.Done()
			results[i] = fetchInfo(serial, label)
		}(i, label, serial)
	}
	wg.Wait()

	allOK := true
	for _, r := range results {
		status := "OK"
		if r.ConnectErr != "" {
			status = "FAIL:connect:" + r.ConnectErr
			allOK = false
		} else if r.VersionErr != "" || r.ProtocolErr != "" || r.BatteryErr != "" {
			status = "PARTIAL"
			allOK = false
		}
		fmt.Printf(
			"ROBOT_INFO_DIRECT label=%s serial=%s os_version=%q engine_build_id=%q "+
				"version_err=%q protocol_result=%s protocol_host_version=%d protocol_err=%q "+
				"battery_status=%s battery_level=%s battery_volts=%.3f is_charging=%t "+
				"is_on_charger_platform=%t battery_err=%q elapsed=%.3fs status=%s\n",
			r.Label, r.Serial, r.OSVersion, r.EngineBuildID,
			r.VersionErr, r.ProtocolResult, r.ProtocolHostVersion, r.ProtocolErr,
			r.BatteryStatus, r.BatteryLevel, r.BatteryVolts, r.IsCharging,
			r.IsOnChargerPlatform, r.BatteryErr, r.End.Sub(r.Start).Seconds(), status,
		)
	}
	if !allOK {
		os.Exit(1)
	}
}
