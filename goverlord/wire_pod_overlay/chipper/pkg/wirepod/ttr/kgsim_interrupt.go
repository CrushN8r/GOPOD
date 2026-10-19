package wirepod_ttr

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
)

func InterruptKGSimWhenTouchedOrWaked(rob *vector.Vector, stop chan bool, stopStop chan bool, kbTouchInterruptWrapper ...func(string) bool) bool {
	strm, err := rob.Conn.EventStream(
		context.Background(),
		&vectorpb.EventRequest{
			ListType: &vectorpb.EventRequest_WhiteList{
				WhiteList: &vectorpb.FilterList{
					List: []string{"robot_state", "wake_word"},
				},
			},
		},
	)
	if err != nil {
		logger.Println("Couldn't make an event stream: " + err.Error())
		return false
	}
	var stopFunc bool
	go func() {
		for range stopStop {
			logger.Println("KG Interrupter has been stopped")
			stopFunc = true
			break
		}
	}()
	var origTouchValue uint32
	var origValueGotten bool
	var valsAboveValue int
	var valsAboveValueMax int = 5
	var stopResponse bool
	var stopSource string
	for {
		var resp *vectorpb.EventResponse
		resp, err = strm.Recv()
		if err != nil {
			break
		}
		switch resp.Event.EventType.(type) {
		case *vectorpb.Event_RobotState:
			origTouchValue = resp.Event.GetRobotState().TouchData.GetRawTouchValue()
			origValueGotten = true
		default:
		}
		if origValueGotten {
			break
		}
	}
	// GOPOD 2026-09-29: the robot sometimes reports a wake word about a second
	// after this listener starts, with nobody speaking (live 2026-09-29 00:49:29,
	// 01:33:23, 01:38:17 - each cut a reply before its first word). Wake-word
	// events in this first window are ignored; touch is never ignored.
	wakeWordIgnoreUntil := time.Now().Add(gopodWakeWordSettle)
	if origValueGotten {
		for {
			var resp *vectorpb.EventResponse
			resp, err = strm.Recv()
			if err != nil {
				logger.Println("Event stream error: " + err.Error())
				return false
			}
			switch resp.Event.EventType.(type) {
			case *vectorpb.Event_RobotState:
				if resp.Event.GetRobotState().TouchData.GetRawTouchValue() > origTouchValue+50 {
					valsAboveValue++
				} else {
					valsAboveValue = 0
				}
			case *vectorpb.Event_WakeWord:
				if time.Now().Before(wakeWordIgnoreUntil) {
					logger.Println("Ignoring wake word event right after the reply started (no real wake word expected yet)")
					break
				}
				logger.Println("Interrupting LLM response (source: wake word)")
				stopSource = "wake word"
				stopResponse = true
			default:
			}
			if valsAboveValue > valsAboveValueMax && gopodTouchInterruptOn() {
				logger.Println("Interrupting LLM response (source: touch sensor)")
				stopSource = "touch sensor"
				stopResponse = true
			}
			if stopResponse {
				if stopSource == "" {
					stopSource = "unknown"
				}
				if len(kbTouchInterruptWrapper) > 0 && kbTouchInterruptWrapper[0] != nil {
					if !kbTouchInterruptWrapper[0](stopSource) {
						return true
					}
				}
				stop <- true
				time.Sleep(time.Second / 4)
				return true
			}
			if stopFunc {
				strm.CloseSend()
				return false
			}
		}
	}
	return false
}

// gopodWakeWordSettle is how long after the reply listener starts a wake-word
// event is treated as stale (see InterruptKGSimWhenTouchedOrWaked).
const gopodWakeWordSettle = 2 * time.Second

// GOPOD 2026-10-01 (operator: "Touch interrupt ON/OFF to default ON again"):
// touching a Brobot cuts its chat reply short, as in native Wire-Pod - except
// while a scripted song's live run has turned it off (the Bingo test), so a
// Brobot can't be stopped by accident mid-run. The song engine creates the
// off file (touch_interrupt.off_file in gopod_brobots_settings.json, in the
// overlay chipper folder) when the run starts and deletes it when the run ends. The
// wake-word interrupt is unchanged.
func gopodTouchInterruptOn() bool {
	_, err := os.Stat(filepath.Join(gopodChipperRoot(), gopodBrobotsSettings.TouchInterrupt.OffFile))
	return err != nil
}
