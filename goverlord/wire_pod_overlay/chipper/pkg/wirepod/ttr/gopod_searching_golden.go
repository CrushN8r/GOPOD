// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_searching_golden.go
package wirepod_ttr

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

type gopodSearching struct {
	Clip, Getout             string
	MinSeconds, FloorSeconds float64
	OK                       bool
}

func gopodSearchingGoldenPath() string {
	return filepath.Join(gopodChipperRoot(), "gopod_probes", "locks", "searching_golden.md")
}

// gopodSearchingGolden: the note's values.
func gopodSearchingGolden() gopodSearching {
	_, sections := gopodPersonaProfileRead(gopodSearchingGoldenPath())
	val := func(section, key string) string {
		for _, p := range sections[section] {
			if k, v, ok := strings.Cut(p, ": "); ok && strings.TrimSpace(k) == key {
				return strings.TrimSpace(v)
			}
		}
		return ""
	}
	g := gopodSearching{Clip: val("The clips", "searching"), Getout: val("The clips", "getout")}
	min, e1 := strconv.ParseFloat(val("The pace", "min_seconds"), 64)
	floor, e2 := strconv.ParseFloat(val("The pace", "floor_seconds"), 64)
	g.MinSeconds, g.FloorSeconds = min, floor
	g.OK = g.Clip != "" && g.Getout != "" && e1 == nil && e2 == nil && floor > 0
	if !g.OK {
		logger.Println("GOPOD_SEARCHING_GOLDEN missing or broken: " + gopodSearchingGoldenPath() + " - no searching clips sent")
	}
	return g
}

// gopodSearchingPlay: one of the note's clips, once - the only place the searching clips are sent
// from.
func gopodSearchingPlay(robot *vector.Vector, ctx context.Context, sg gopodSearching, clip string) {
	if !sg.OK || clip == "" {
		return
	}
	gopodLogPlayAnimation(robot, ctx, &vectorpb.PlayAnimationRequest{
		Animation: &vectorpb.Animation{Name: clip},
		Loops:     1,
	})
}

// ## EOF
