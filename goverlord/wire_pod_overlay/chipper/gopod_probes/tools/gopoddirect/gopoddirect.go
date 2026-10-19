// Package gopoddirect is the shared low-level per-robot connect/grant/
// release primitive every direct_sdk_*.go tool in this family (sleep,
// rattle, brobots_ready) independently hand-rolled until 2026-08-08
// (PHCAL_SLEEP_HOLD_BARRIER_FIX_001.md's own follow-up: the sleep hold bug
// was fixed in one file, but the SAME connect/grant/release scaffolding -
// and two independently-invented sync.WaitGroup barrier patterns - existed
// three separate times, confirmed by direct_sdk_bingo_rattle_001.go's own
// comment admitting it was "built by copying" this exact skeleton from
// direct_sdk_brobots_ready_001.go). Golden per-robot timing correctness now
// lives in ONE place; every current and future tool in this family imports
// it instead of re-copying it, so a future new dual-robot mechanism
// inherits response-time-aware coordination by construction rather than
// needing its own audit-and-patch pass later (GOPOD_WIDGET_TARGET_SHAPE_001.md
// / GOLDEN_BROBOTS_CONTROL_CATALOG_SURVEY_001.md's own "eliminate the
// patches" ask).
package gopoddirect

import (
	"context"
	"fmt"
	"sync"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
)

// ConnectAndGrant connects to serial and requests+waits for a real,
// confirmed BehaviorControl grant (OVERRIDE_BEHAVIORS priority) - the exact
// connect -> control_request -> wait-for-ControlGrantedResponse sequence
// every tool in this family already needed, reproduced here verbatim (same
// SDK calls, same priority, same blocking-receive-loop shape, not
// reinvented). Returns the connected robot and the open stream (for the
// caller's own action calls, then Release), or an error identifying which
// stage failed - same "stage: <err>" message shape every existing tool's
// own hand-rolled error strings already used (connect/behavior_control_open/
// control_request_send/control_grant_recv), so nothing downstream that logs
// or inspects that text sees a different shape than before this refactor.
func ConnectAndGrant(ctx context.Context, serial string) (*vector.Vector, vectorpb.ExternalInterface_BehaviorControlClient, error) {
	robot, err := vector.NewWP(serial)
	if err != nil {
		return nil, nil, fmt.Errorf("connect: %v", err)
	}

	stream, err := robot.Conn.BehaviorControl(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("behavior_control_open: %v", err)
	}

	if err := stream.Send(&vectorpb.BehaviorControlRequest{
		RequestType: &vectorpb.BehaviorControlRequest_ControlRequest{
			ControlRequest: &vectorpb.ControlRequest{
				Priority: vectorpb.ControlRequest_OVERRIDE_BEHAVIORS,
			},
		},
	}); err != nil {
		return nil, nil, fmt.Errorf("control_request_send: %v", err)
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			return nil, nil, fmt.Errorf("control_grant_recv: %v", err)
		}
		if resp.GetControlGrantedResponse() != nil {
			return robot, stream, nil
		}
	}
}

// Release hands BehaviorControl back - the same ControlRelease shape every
// tool in this family already sent (mirrors chipper/pkg/wirepod/sdkapp/
// bcassume.go's own release shape). Errors are deliberately not returned -
// every existing caller already discarded this same send's own error
// (`_ = stream.Send(...)`), matching that established, live-proven
// precedent exactly, not a new decision made in this refactor.
func Release(stream vectorpb.ExternalInterface_BehaviorControlClient) {
	_ = stream.Send(&vectorpb.BehaviorControlRequest{
		RequestType: &vectorpb.BehaviorControlRequest_ControlRelease{
			ControlRelease: &vectorpb.ControlRelease{},
		},
	})
	_ = stream.CloseSend()
}

// Barrier is a thin, named wrapper over sync.WaitGroup for the "wait until
// every robot reaches checkpoint X before any of them proceeds" pattern -
// used at least three times independently before this refactor (the sleep
// tool's own start barrier AND all-asleep barrier, the brobots_ready tool's
// own grant barrier), each a separately hand-rolled WaitGroup with its own
// name. One shared name/shape now, so a future dual-robot tool reaches for
// this instead of re-inventing it - and so a bug fix to the pattern itself
// (like 2026-08-08's barrier-corrected sleep hold) only ever needs making
// once, here, not once per tool that happens to have copied the old shape.
// Single-robot tools (rattle) have no barrier need at all and don't use
// this type - nothing about it is mandatory plumbing.
type Barrier struct {
	wg *sync.WaitGroup
}

// NewBarrier creates a barrier for n participants.
func NewBarrier(n int) *Barrier {
	wg := &sync.WaitGroup{}
	wg.Add(n)
	return &Barrier{wg: wg}
}

// Done marks this participant as having reached the checkpoint. Same
// exactly-once-per-participant contract sync.WaitGroup.Done() itself
// already has - not widened or changed here.
func (b *Barrier) Done() {
	b.wg.Done()
}

// Wait blocks until every participant has called Done().
func (b *Barrier) Wait() {
	b.wg.Wait()
}
