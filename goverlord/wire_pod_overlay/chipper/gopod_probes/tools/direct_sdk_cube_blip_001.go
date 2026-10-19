// direct_sdk_cube_blip_001.go - standalone, decoupled from Wire-Pod's own
// running process. Net-new: the first GOPOD code to touch the cube
// instrument at all. Live-fire confirmed 2026-08-14 (mode "blip", the
// default): connect to the cube, set all four corner LEDs red (control
// established, standby), brief hold, set them all green (control confirmed,
// go), release. Mirrors direct_sdk_bingo_rattle_001.go's own structure
// exactly - same package, same connect/grant/release shape, same
// serial-as-CLI-arg convention, same status-line print shape - over the
// vendored fforchino/vector-go-sdk chipper/go.mod already requires, no new
// dependency. Two more modes added 2026-08-15, operator request: "flash"
// (a single on-then-off beat, 1s hold, for demonstrating raw control) and
// "off" (immediate, no hold - for clearing a cube left lit from a prior
// blip/flash rather than waiting on its own firmware timeout). Optional
// second CLI arg, defaults to "blip" - the bare `<serial>` call stays
// byte-identical to the live-proven 2026-08-14 run. Survey:
// gopod_notes/CUBE_DOOR_SURVEY_001.md. Build reports:
// gopod_notes/CUBE_BLIP_TOOL_BUILT_001.md,
// gopod_notes/CUBE_BLIP_MODES_ADDED_001.md.
//
// Connect/grant/release comes from the shared gopoddirect package
// (gopoddirect.ConnectAndGrant/Release) - the same primitive
// direct_sdk_bingo_rattle_001.go itself now uses (refactored there
// 2026-08-08, PHCAL_SLEEP_HOLD_BARRIER_FIX_001.md), not a hand-copied
// skeleton - this tool reuses that shared package directly rather than
// re-copying connect/grant/release logic a fourth time.
//
// Color encoding derived, not guessed: Vector's cube lights pack a color as
// a 32-bit 0xRRGGBBAA uint32, alpha byte always 0xff for a lit color, 0x0
// for off - confirmed against the authoritative Anki anki_vector/color.py
// (gopod_notes/SDK_SOURCES_BACKUP_001/wirepod-vector-python-sdk/anki_vector/color.py),
// the same packing SetCubeLightsRequest's on_color/off_color fields (both
// []uint32, per proto/cube.proto) expect. RED = 0xff0000ff, GREEN =
// 0x00ff00ff - matching that source's own red/green Color constants
// exactly.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/gopod_probes/tools/gopoddirect"
)

// Packed 0xRRGGBBAA, alpha forced to 0xff for a lit color - derived from
// anki_vector/color.py's own Color constants (red/green), not guessed.
// cubeColorOff is the SDK's own "off" value (Color() with no args stays
// 0x00000000 in anki_vector/color.py) - not guessed either.
const cubeColorRed uint32 = 0xff0000ff
const cubeColorGreen uint32 = 0x00ff00ff
const cubeColorOff uint32 = 0x00000000

// Brief hold between the red and green blip, so both are actually visible
// as two distinct beats rather than one instantaneous flip - a chosen
// duration for this proof, not derived from any spec.
const blipHoldDuration = 2 * time.Second

// flash mode's own hold - 2026-08-15, operator request ("a timed on for 1s,
// then off... I'd like to show the control feature") - a single deliberate
// on-then-off beat, distinct from blip's two-color proof.
const flashHoldDuration = 1 * time.Second

// Steady (non-blinking) display: on_period_ms long enough to never cycle
// into the off phase during the hold above, off_period_ms=0, no transition
// - same shape anki_vector.lights.Light's own defaults use (on_period_ms:
// 250, off_period_ms: 0) scaled up so a brief visual hold never flickers.
const cubeLightOnPeriodMs uint32 = 3000

func setAllCorners(ctx context.Context, robot *vector.Vector, objectID uint32, onColor uint32) error {
	corners := []uint32{onColor, onColor, onColor, onColor}
	periods := []uint32{cubeLightOnPeriodMs, cubeLightOnPeriodMs, cubeLightOnPeriodMs, cubeLightOnPeriodMs}
	offPeriods := []uint32{0, 0, 0, 0}
	transitions := []uint32{0, 0, 0, 0}

	_, err := robot.Conn.SetCubeLights(ctx, &vectorpb.SetCubeLightsRequest{
		ObjectId:              objectID,
		OnColor:               corners,
		OffColor:              corners,
		OnPeriodMs:            periods,
		OffPeriodMs:           offPeriods,
		TransitionOnPeriodMs:  transitions,
		TransitionOffPeriodMs: transitions,
	})
	return err
}

// pairCube connects the cube (ConnectCube) and reads its ObjectId - the one
// piece every mode below needs before it can call SetCubeLights. Extracted
// 2026-08-15 when flash/off modes were added, so the connect_cube error text
// has exactly one source instead of being duplicated per mode - the
// already-live-proven blip path's own behavior (including exactly where
// `start` gets stamped, before this call, same as it always was) is
// unchanged by this extraction.
func pairCube(ctx context.Context, robot *vector.Vector) (objectID uint32, err error) {
	from := time.Now()
	defer func() { connectSeconds = time.Since(from).Seconds() }()
	connResp, err := robot.Conn.ConnectCube(ctx, &vectorpb.ConnectCubeRequest{})
	if err != nil {
		return 0, fmt.Errorf("connect_cube: %v", err)
	}
	if !connResp.GetSuccess() {
		return 0, fmt.Errorf("connect_cube: not successful (no cube found/paired)")
	}
	return connResp.GetObjectId(), nil
}

// connectSeconds: how long ConnectCube took this run (2026-10-14, operator: the
// nap's cube "takes forever ... I need sync/async, not sequential based on time
// estimates" - the blip's 20 s was this, not the lights).
var connectSeconds float64

// assumeControlAndConnect (2026-10-14, cube option 2): ConnectCube only - the
// robot connects his cube over Bluetooth ahead of time, so a later blip finds it
// connected. No lights.
func assumeControlAndConnect(serial string) (ok bool, errMsg string, objectID uint32, start, end time.Time) {
	ctx := context.Background()
	robot, stream, err := gopoddirect.ConnectAndGrant(ctx, serial)
	if err != nil {
		return false, err.Error(), 0, start, end
	}
	defer gopoddirect.Release(stream)
	start = time.Now()
	objectID, err = pairCube(ctx, robot)
	end = time.Now()
	if err != nil {
		return false, err.Error(), objectID, start, end
	}
	return true, "", objectID, start, end
}

// assumeControlAndLightOnSignal (2026-10-14, the nap, operator: "Cube always
// took way too long to light up"; "soft wake ... & then cube lights up"): one
// held session - ConnectCube now (Bluetooth, about 12 s), then wait for the
// signal file, then the blip's lights at once, then release. A later light
// never waits on a connect. maxWait (seconds) is the safety net: no signal,
// it lights anyway.
//
// delay (seconds) comes first: the nap's live runs (2026-10-15) - a ConnectCube
// started about 1 s after the robot woke failed ("no cube found", 2 of 2), one
// started 8 s after worked (13 of 13, and the bench: 3.7 s connect).
func assumeControlAndLightOnSignal(serial, signal string, maxWait, delay float64) (ok bool, errMsg string, objectID uint32, start, end time.Time) {
	time.Sleep(time.Duration(delay * float64(time.Second)))
	ctx := context.Background()
	robot, stream, err := gopoddirect.ConnectAndGrant(ctx, serial)
	if err != nil {
		return false, err.Error(), 0, start, end
	}
	defer gopoddirect.Release(stream)
	objectID, err = pairCube(ctx, robot)
	if err != nil {
		return false, err.Error(), objectID, start, end
	}
	deadline := time.Now().Add(time.Duration(maxWait * float64(time.Second)))
	for time.Now().Before(deadline) {
		if _, err := os.Stat(signal); err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	start = time.Now()
	if err := setAllCorners(ctx, robot, objectID, cubeColorRed); err != nil {
		end = time.Now()
		return false, fmt.Sprintf("set_cube_lights_red: %v", err), objectID, start, end
	}
	time.Sleep(blipHoldDuration)
	if err := setAllCorners(ctx, robot, objectID, cubeColorGreen); err != nil {
		end = time.Now()
		return false, fmt.Sprintf("set_cube_lights_green: %v", err), objectID, start, end
	}
	end = time.Now()
	return true, "", objectID, start, end
}

func assumeControlAndBlip(serial string) (ok bool, errMsg string, objectID uint32, start, end time.Time) {
	ctx := context.Background()

	robot, stream, err := gopoddirect.ConnectAndGrant(ctx, serial)
	if err != nil {
		return false, err.Error(), 0, start, end
	}
	defer gopoddirect.Release(stream)

	start = time.Now()

	objectID, err = pairCube(ctx, robot)
	if err != nil {
		end = time.Now()
		return false, err.Error(), objectID, start, end
	}

	if err := setAllCorners(ctx, robot, objectID, cubeColorRed); err != nil {
		end = time.Now()
		return false, fmt.Sprintf("set_cube_lights_red: %v", err), objectID, start, end
	}

	select {
	case <-ctx.Done():
		end = time.Now()
		return false, ctx.Err().Error(), objectID, start, end
	case <-time.After(blipHoldDuration):
	}

	if err := setAllCorners(ctx, robot, objectID, cubeColorGreen); err != nil {
		end = time.Now()
		return false, fmt.Sprintf("set_cube_lights_green: %v", err), objectID, start, end
	}

	end = time.Now()
	return true, "", objectID, start, end
}

// assumeControlAndOff - 2026-08-15, operator request: an explicit "turn it
// off" mode, for whenever the cube is sitting lit from a prior blip/flash
// and the operator wants it off now rather than waiting on the cube's own
// firmware timeout (confirmed live 2026-08-14: it does self-expire, but not
// instantly).
func assumeControlAndOff(serial string) (ok bool, errMsg string, objectID uint32, start, end time.Time) {
	ctx := context.Background()

	robot, stream, err := gopoddirect.ConnectAndGrant(ctx, serial)
	if err != nil {
		return false, err.Error(), 0, start, end
	}
	defer gopoddirect.Release(stream)

	start = time.Now()

	objectID, err = pairCube(ctx, robot)
	if err != nil {
		end = time.Now()
		return false, err.Error(), objectID, start, end
	}

	if err := setAllCorners(ctx, robot, objectID, cubeColorOff); err != nil {
		end = time.Now()
		return false, fmt.Sprintf("set_cube_lights_off: %v", err), objectID, start, end
	}

	end = time.Now()
	return true, "", objectID, start, end
}

// assumeControlAndFlash - 2026-08-15, operator request: a single deliberate
// on-then-off beat ("a timed on for 1s, then off... I'd like to show the
// control feature") - simpler than blip's two-color proof, meant for
// demonstrating raw on/off control rather than the red-then-green sequence.
func assumeControlAndFlash(serial string) (ok bool, errMsg string, objectID uint32, start, end time.Time) {
	ctx := context.Background()

	robot, stream, err := gopoddirect.ConnectAndGrant(ctx, serial)
	if err != nil {
		return false, err.Error(), 0, start, end
	}
	defer gopoddirect.Release(stream)

	start = time.Now()

	objectID, err = pairCube(ctx, robot)
	if err != nil {
		end = time.Now()
		return false, err.Error(), objectID, start, end
	}

	if err := setAllCorners(ctx, robot, objectID, cubeColorGreen); err != nil {
		end = time.Now()
		return false, fmt.Sprintf("set_cube_lights_on: %v", err), objectID, start, end
	}

	select {
	case <-ctx.Done():
		end = time.Now()
		return false, ctx.Err().Error(), objectID, start, end
	case <-time.After(flashHoldDuration):
	}

	if err := setAllCorners(ctx, robot, objectID, cubeColorOff); err != nil {
		end = time.Now()
		return false, fmt.Sprintf("set_cube_lights_off: %v", err), objectID, start, end
	}

	end = time.Now()
	return true, "", objectID, start, end
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: direct_sdk_cube_blip_001 <serial> [mode]")
		fmt.Fprintln(os.Stderr, "  mode: blip (default) - connect -> red -> 2s -> green -> release")
		fmt.Fprintln(os.Stderr, "        flash - connect -> on -> 1s -> off -> release")
		fmt.Fprintln(os.Stderr, "        off - connect -> off -> release, immediate")
		fmt.Fprintln(os.Stderr, "        connect - connect only (no lights) -> release")
		fmt.Fprintln(os.Stderr, "        light-on-signal PATH [MAX_WAIT] [DELAY] - DELAY s, connect, wait for PATH, then the blip's lights -> release")
		fmt.Fprintln(os.Stderr, "  Brobot 2 (cube keeper) ESN: 0dd1d8bf")
		os.Exit(2)
	}
	serial := os.Args[1]
	mode := "blip"
	if len(os.Args) >= 3 {
		mode = os.Args[2]
	}

	var ok bool
	var errMsg string
	var objectID uint32
	var start, end time.Time
	switch mode {
	case "blip":
		ok, errMsg, objectID, start, end = assumeControlAndBlip(serial)
	case "flash":
		ok, errMsg, objectID, start, end = assumeControlAndFlash(serial)
	case "off":
		ok, errMsg, objectID, start, end = assumeControlAndOff(serial)
	case "connect":
		ok, errMsg, objectID, start, end = assumeControlAndConnect(serial)
	case "light-on-signal":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "light-on-signal needs a signal file path")
			os.Exit(2)
		}
		maxWait, delay := 60.0, 0.0
		if len(os.Args) >= 5 {
			fmt.Sscanf(os.Args[4], "%f", &maxWait)
		}
		if len(os.Args) >= 6 {
			fmt.Sscanf(os.Args[5], "%f", &delay)
		}
		ok, errMsg, objectID, start, end = assumeControlAndLightOnSignal(serial, os.Args[3], maxWait, delay)
	default:
		fmt.Fprintf(os.Stderr, "usage: unknown mode '%s' - expected blip, flash, off, connect or light-on-signal\n", mode)
		os.Exit(2)
	}

	status := "OK"
	if !ok {
		status = "FAIL:" + errMsg
	}
	fmt.Printf(
		"CUBE_BLIP_DIRECT mode=%s serial=%s object_id=%d start=%.6f end=%.6f elapsed=%.3fs connect=%.3fs status=%s\n",
		mode, serial, objectID,
		float64(start.UnixNano())/1e9, float64(end.UnixNano())/1e9,
		end.Sub(start).Seconds(), connectSeconds, status,
	)
	if !ok {
		os.Exit(1)
	}
}
