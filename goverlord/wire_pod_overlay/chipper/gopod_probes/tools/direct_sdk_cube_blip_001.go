// PRESENTATION COPY — not the live build location. This file compiles as
// part of Wire-Pod's own chipper Go module (it imports the sibling
// gopoddirect package and the vendored fforchino/vector-go-sdk, both
// anchored under chipper/gopod_probes/tools/) and can't build standalone
// here — this directory has no go.mod, no gopoddirect package, no chipper
// module context. Canonical source, the actual build/edit location, and
// the only copy that ever runs: ~/wire-pod/chipper/gopod_probes/tools/
// direct_sdk_cube_blip_001.go — tracked in ~/wire-pod's own local git
// history (per tech/WIRED-POD.md's "The interview/demo probe tree"
// section), never this repo's git history, per "keep ~/wire-pod as native
// as possible." This copy exists purely so the cube work is visible in
// GOPOD's public-facing repo — kept in sync by hand; port any real change
// to the live file first, then re-copy it here. Survey:
// gopod_notes/CUBE_DOOR_SURVEY_001.md. Build report:
// gopod_notes/CUBE_BLIP_TOOL_BUILT_001.md.
//
// direct_sdk_cube_blip_001.go - standalone, decoupled from Wire-Pod's own
// running process. Net-new: the first GOPOD code to touch the cube
// instrument at all. Proves the smallest real thing before any bingo
// integration - connect to the cube, set all four corner LEDs red (control
// established, standby), brief hold, set them all green (control confirmed,
// go), release. Mirrors direct_sdk_bingo_rattle_001.go's own structure
// exactly - same package, same connect/grant/release shape, same
// serial-as-CLI-arg convention, same status-line print shape - over the
// vendored fforchino/vector-go-sdk chipper/go.mod already requires, no new
// dependency.
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
const cubeColorRed uint32 = 0xff0000ff
const cubeColorGreen uint32 = 0x00ff00ff

// Brief hold between the red and green blip, so both are actually visible
// as two distinct beats rather than one instantaneous flip - a chosen
// duration for this proof, not derived from any spec.
const blipHoldDuration = 2 * time.Second

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

func assumeControlAndBlip(serial string) (ok bool, errMsg string, objectID uint32, start, end time.Time) {
	ctx := context.Background()

	robot, stream, err := gopoddirect.ConnectAndGrant(ctx, serial)
	if err != nil {
		return false, err.Error(), 0, start, end
	}
	defer gopoddirect.Release(stream)

	start = time.Now()

	connResp, err := robot.Conn.ConnectCube(ctx, &vectorpb.ConnectCubeRequest{})
	if err != nil {
		end = time.Now()
		return false, fmt.Sprintf("connect_cube: %v", err), 0, start, end
	}
	if !connResp.GetSuccess() {
		end = time.Now()
		return false, "connect_cube: not successful (no cube found/paired)", 0, start, end
	}
	objectID = connResp.GetObjectId()

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

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: direct_sdk_cube_blip_001 <serial>")
		fmt.Fprintln(os.Stderr, "  Pip / Brobot 2 (cube keeper) ESN: 0dd1d8bf")
		os.Exit(2)
	}
	serial := os.Args[1]

	ok, errMsg, objectID, start, end := assumeControlAndBlip(serial)
	status := "OK"
	if !ok {
		status = "FAIL:" + errMsg
	}
	fmt.Printf(
		"CUBE_BLIP_DIRECT serial=%s object_id=%d start=%.6f end=%.6f elapsed=%.3fs status=%s\n",
		serial, objectID,
		float64(start.UnixNano())/1e9, float64(end.UnixNano())/1e9,
		end.Sub(start).Seconds(), status,
	)
	if !ok {
		os.Exit(1)
	}
}
