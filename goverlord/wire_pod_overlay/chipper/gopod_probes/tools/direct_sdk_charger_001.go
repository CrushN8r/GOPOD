// direct_sdk_charger_001.go - a Brobot's own charger moves, over GOPOD's direct
// robot connection (gopoddirect - the same connect/grant/release every direct
// tool here uses: brobots ready, sleep, rattle, cube blip). Wire-Pod's web SDK
// page has no charger moves; the robot's own SDK does:
//
//	off - DriveOffCharger: he drives off his charger by himself.
//	on  - DriveOnCharger: he docks by himself - it works when his charger is in
//	      sight (point him at it first; air hockey, live 2026-11-23).
//
// Usage (WIREPOD_HOME=~/wire-pod, as every direct tool):
//
//	direct_sdk_charger_001 off ESN02
//	direct_sdk_charger_001 on ESN02
//
// Build (gopoddirect resolves inside Wire-Pod's chipper module): see the air
// hockey file's build note, or the alias that builds it.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/gopod_probes/tools/gopoddirect"
)

func main() {
	if len(os.Args) < 3 || (os.Args[1] != "off" && os.Args[1] != "on") {
		fmt.Fprintln(os.Stderr, "usage: direct_sdk_charger_001 <off|on> <ESN01|ESN02|serial>")
		os.Exit(2)
	}
	move, serial := os.Args[1], os.Args[2]
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	robot, stream, err := gopoddirect.ConnectAndGrant(ctx, serial)
	if err != nil {
		fmt.Printf("CHARGER move=%s status=FAIL:%v\n", move, err)
		os.Exit(1)
	}
	defer gopoddirect.Release(stream)

	start := time.Now()
	var result string
	if move == "off" {
		resp, err := robot.Conn.DriveOffCharger(ctx, &vectorpb.DriveOffChargerRequest{})
		result = fmt.Sprint(resp.GetResult(), err)
		if err != nil {
			fmt.Printf("CHARGER move=off status=FAIL:%v\n", err)
			os.Exit(1)
		}
	} else {
		resp, err := robot.Conn.DriveOnCharger(ctx, &vectorpb.DriveOnChargerRequest{})
		result = fmt.Sprint(resp.GetResult(), err)
		if err != nil {
			fmt.Printf("CHARGER move=on status=FAIL:%v\n", err)
			os.Exit(1)
		}
	}
	fmt.Printf("CHARGER move=%s result=%s elapsed=%.1fs\n", move, result, time.Since(start).Seconds())
}
