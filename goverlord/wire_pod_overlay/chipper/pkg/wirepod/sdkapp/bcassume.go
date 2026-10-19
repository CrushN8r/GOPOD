// goverlord/wire_pod_overlay/chipper/pkg/wirepod/sdkapp/bcassume.go
// Native Wire-Pod file (upstream 11e7b22) with one GOPOD guard, marked below.
package sdkapp

import (
	"log"
	"strings"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
)

func assumeBehaviorControl(robot Robot, robotIndex int, priority string) {
	var controlRequest *vectorpb.BehaviorControlRequest
	if priority == "high" {
		controlRequest = &vectorpb.BehaviorControlRequest{
			RequestType: &vectorpb.BehaviorControlRequest_ControlRequest{
				ControlRequest: &vectorpb.ControlRequest{
					Priority: vectorpb.ControlRequest_OVERRIDE_BEHAVIORS,
				},
			},
		}
	} else {
		controlRequest = &vectorpb.BehaviorControlRequest{
			RequestType: &vectorpb.BehaviorControlRequest_ControlRequest{
				ControlRequest: &vectorpb.ControlRequest{
					Priority: vectorpb.ControlRequest_DEFAULT,
				},
			},
		}
	}
	go func() {
		start := make(chan bool)
		stop := make(chan bool)
		robots[robotIndex].BcAssumption = true
		go func() {
			// * begin - modified from official vector-go-sdk
			r, err := robot.Vector.Conn.BehaviorControl(
				robot.Ctx,
			)
			if err != nil {
				log.Println(err)
				return
			}

			if err := r.Send(controlRequest); err != nil {
				log.Println(err)
				return
			}

			for {
				ctrlresp, err := r.Recv()
				if err != nil {
					log.Println(err)
					return
				}
				if ctrlresp.GetControlGrantedResponse() != nil {
					start <- true
					break
				}
			}

			for {
				select {
				case <-stop:
					if err := r.Send(
						&vectorpb.BehaviorControlRequest{
							RequestType: &vectorpb.BehaviorControlRequest_ControlRelease{
								ControlRelease: &vectorpb.ControlRelease{},
							},
						},
					); err != nil {
						log.Println(err)
						return
					}
					return
				default:
					continue
				}
			}
			// * end - modified from official vector-go-sdk
		}()
		for range start {
			// GOPOD 2026-10-08: the robot is found by its serial each time, not by
			// its place in the list. When the SDK connection timed out (connTimer),
			// the list shrank and robots[robotIndex] crashed Wire-Pod ("index out of
			// range", live 2026-10-08 11:16:53); a robot no longer listed is let go.
			for gopodBcAssumed(robot.ESN) {
				time.Sleep(time.Millisecond * 500)
			}
			stop <- true
		}
	}()
}

// gopodBcAssumed: whether the listed robot with this serial still holds behavior
// control; false once it isn't listed.
func gopodBcAssumed(esn string) bool {
	for _, r := range robots {
		if strings.EqualFold(r.ESN, esn) {
			return r.BcAssumption
		}
	}
	return false
}
