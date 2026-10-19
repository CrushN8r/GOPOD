// direct_sdk_brobots_ready_001.go - standalone, decoupled from Wire-Pod's own
// running process. Connects to both Brobots directly over the vendored
// fforchino/vector-go-sdk (the same module chipper/go.mod already requires -
// no new dependency), fires "Brobots ready!" (or whatever text is passed) on
// both concurrently via goroutines in one process, and reports real
// dispatch timing so the caller can measure actual overlap.
//
// vector.NewWP(serial) is a helper already vendored in the SDK itself
// (pkg/vector/vector.go) specifically for reading Wire-Pod's own
// chipper/jdocs/botSdkInfo.json (WIREPOD_HOME env var, defaults to ".") -
// no new credential source, no new machinery, exactly the file Wire-Pod's
// own newRobot() reads.
//
// This file lives under gopod_probes/tools/, a directory of independent
// tools alongside the Python runners - it is its own `package main`, never
// swept into the live wire-pod binary (start.sh's build only ever targets
// specific cmd/*/main.go files, confirmed by reading start.sh directly).
// No wire-pod binary rebuild needed for this file to exist or change.
//
// Behavior-control request/release below mirrors the exact ControlRequest/
// priority shape chipper/pkg/wirepod/sdkapp/bcassume.go uses for "high"
// priority (OVERRIDE_BEHAVIORS) - reused shape, not new protocol. Unlike
// bcassume.go's fire-and-forget async version (needed there because it's a
// long-lived server), this is a short-lived one-shot tool, so the control
// grant is waited on synchronously before speaking - a real confirmed grant,
// not just an HTTP ack.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/gopod_probes/tools/gopoddirect"
)

type robotResult struct {
	Label  string
	Serial string
	Start  time.Time
	End    time.Time
	OK     bool
	Err    string
}

// assumeControlAndSay gets this robot's own control grant, then blocks on
// grantBarrier - every robot's grant (success or failure) must land before
// ANY robot is allowed to proceed to SayText, so a fast grant never fires
// ahead of a slow one. Confirmed live 2026-07-26: without this barrier,
// TOGETHER start-gap varied run to run (0.001s to 0.216s) purely on which
// robot's independent grant round-trip happened to land first.
//
// 2026-08-08: connect/grant/release now come from the shared gopoddirect
// package (PHCAL_SLEEP_HOLD_BARRIER_FIX_001.md's own follow-up - this
// file's own hand-copied skeleton was direct_sdk_robot_sleep_001.go's own
// original template, per this file's own header comment; the duplication
// is now gone, not just documented) and the barrier is gopoddirect.Barrier
// instead of a private *sync.WaitGroup - same shape, shared name.
func assumeControlAndSay(serial, text string, grantBarrier *gopoddirect.Barrier) robotResult {
	res := robotResult{Serial: serial}
	ctx := context.Background()

	robot, stream, err := gopoddirect.ConnectAndGrant(ctx, serial)
	grantBarrier.Done()
	if err != nil {
		res.Err = err.Error()
		return res
	}

	grantBarrier.Wait()

	res.Start = time.Now()
	_, sayErr := robot.Conn.SayText(ctx, &vectorpb.SayTextRequest{
		DurationScalar: 1,
		UseVectorVoice: true,
		Text:           text,
	})
	res.End = time.Now()
	if sayErr != nil {
		res.Err = fmt.Sprintf("say_text: %v", sayErr)
	} else {
		res.OK = true
	}

	gopoddirect.Release(stream)

	return res
}

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: direct_sdk_brobots_ready_001 <text> <label1>:<serial1> <label2>:<serial2>")
		os.Exit(2)
	}
	text := os.Args[1]
	specs := os.Args[2:4]

	results := make([]robotResult, len(specs))
	var wg sync.WaitGroup
	grantBarrier := gopoddirect.NewBarrier(len(specs))
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
			r := assumeControlAndSay(serial, text, grantBarrier)
			r.Label = label
			results[i] = r
		}(i, label, serial)
	}
	wg.Wait()

	allOK := true
	for _, r := range results {
		status := "OK"
		if !r.OK {
			status = "FAIL:" + r.Err
			allOK = false
		}
		fmt.Printf(
			"TOGETHER_DIRECT label=%s serial=%s start=%.6f end=%.6f elapsed=%.3fs status=%s\n",
			r.Label, r.Serial,
			float64(r.Start.UnixNano())/1e9, float64(r.End.UnixNano())/1e9,
			r.End.Sub(r.Start).Seconds(), status,
		)
	}
	if !allOK {
		os.Exit(1)
	}
}
