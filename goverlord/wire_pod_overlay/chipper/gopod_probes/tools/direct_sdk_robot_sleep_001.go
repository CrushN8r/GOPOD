// direct_sdk_robot_sleep_001.go - fires the real GoToSleep* animation
// TRIGGERS (not raw clips) directly over the vendored fforchino/vector-go-sdk,
// bypassing Wire-Pod's own dispatch entirely. Built 2026-07-25
// (VECTOR_SLEEP_PETTING_ANIM_PINNED_REFERENCE_001.md /
// ROBOT_SLEEP_DIRECT_SDK_BUILT_001.md) because Wire-Pod's own
// DoPlayAnimation/DoPlayAnimationWI (chipper/pkg/wirepod/ttr/kgsim_cmds.go)
// only ever call PlayAnimation (a one-shot raw clip), never
// PlayAnimationTrigger - confirmed zero hits for PlayAnimationTrigger
// anywhere in the live chipper tree. A raw clip play does not enter Vector's
// actual persistent sleep behavior/state; only the real trigger does. There
// is also no animation_vocab.json vocab entry for any gotosleep-family name
// (confirmed: zero "sleep" hits in the live file) - even the raw-clip path
// through Wire-Pod's own {{playAnimationWI||token}} mechanism would silently
// no-op (DoPlayAnimationWI logs "doesn't exist" and returns nil, no error) -
// so this could not have been built as a vocab-token addition even if that
// were desired.
//
// Source-verified trigger names, not guessed: GoToSleepGetIn/GoToSleepOff/
// GoToSleepSleeping are confirmed in the real robot-engine CLAD source
// (goverlord/SDK/sources/Viccyware/clad/src/clad/types/animationTrigger.clad:
// 201-203). No GoToSleepWakeup trigger exists anywhere in that source - only
// a wakeup_01 *clip*, which (per the reasoning above) cannot be fired as a
// real behavior through Wire-Pod either. GoToSleepOff is the nearest
// confirmed exit trigger - CONFIRMED LIVE 2026-07-25 (fifth live run,
// ROBOT_SLEEP_DIRECT_SDK_BUILT_001.md) to genuinely rouse a robot actually
// held in GoToSleepSleeping, both robots together, operator-witnessed.
//
// vector.NewWP(serial) reads Wire-Pod's own chipper/jdocs/botSdkInfo.json
// (WIREPOD_HOME env var, defaults ".") - same helper, same credential
// source, same reasoning as direct_sdk_brobots_ready_001.go, which this file
// otherwise mirrors line-for-line on the connect/BehaviorControl-grant/
// release shape. No new dependency, no new auth path.
//
// Usage: direct_sdk_robot_sleep_001 <sleep|wake> [--hold N | --wait-signal PATH [--max-wait N]] <label1>:<serial1> [<label2>:<serial2> ...]
//
//	       direct_sdk_robot_sleep_001 play [--hold N] [--clip] <name> <label1>:<serial1> [<label2>:<serial2> ...]
//
//		play: added for the sleep-lane test bench (SLEEP_BENCH_ALIASES_001.md) -
//		      fires exactly one named trigger (default, PlayAnimationTrigger) or,
//		      with --clip, one named raw clip (PlayAnimation), holds
//		      BehaviorControl open for --hold seconds (default 5.0, same as
//		      sleep/wake), then releases. Same connect/grant/release/timeout
//		      shape as sleep/wake, verbatim - no --wait-signal support, no
//		      change to the sleep/wake actions themselves.
//
//		sleep: fires GoToSleepGetIn, waits getInSettleSeconds, fires
//		       GoToSleepSleeping, then EITHER holds for --hold seconds (legacy,
//		       manual/testing mode) OR waits event-driven for --wait-signal's
//		       file to appear (see below) - exactly one of the two, --hold is
//		       the default (5.0s) when neither flag is given.
//		wake:  fires GoToSleepOff on its own, standalone (unaffected by --hold/
//		       --wait-signal - both are sleep-only).
//
//		--hold N: seconds to hold behavior control open after firing
//		          GoToSleepSleeping, before releasing it. Default 5.0. Added
//		          2026-07-25, live-observed need: a first live test with no hold
//		          at all had both robots wake right back up immediately -
//		          releasing BehaviorControl right after firing the trigger
//		          appears to hand control straight back to the robot's own
//		          onboard behavior, which does not stay asleep on its own.
//
//		--wait-signal PATH [--max-wait N]: event-driven alternative to --hold,
//		          added 2026-07-25 (operator request - "the golden sequence":
//		          wake -> synced sleep -> warm-up runs concurrently -> synced
//		          wake -> song starts). Instead of a fixed hold, every robot's
//		          goroutine blocks on a single shared in-process channel after
//		          firing GoToSleepSleeping. One dedicated watcher goroutine
//		          polls for PATH's existence (every 100ms) and closes that
//		          channel the instant it appears (or after --max-wait seconds
//		          elapse with no signal, default 60.0 - a safety net so a
//		          crashed/hung external warm-up process can never leave this
//		          tool holding behavior control forever). Channel close is a
//		          genuine Go broadcast - every blocked goroutine unblocks at the
//		          same instant, no polling jitter between robots (tighter even
//		          than the already-confirmed 104-microsecond start-barrier sync
//		          below). The instant a robot's goroutine unblocks, it fires
//		          GoToSleepOff immediately, in the SAME still-open
//		          BehaviorControl connection that has held sleep the whole
//		          time - never releasing control and re-granting it for a
//		          separate wake call, which would reopen exactly the gap that
//		          caused the first live run's "woke back up on its own" bug (the
//		          robot's own onboard behavior can act the instant control is
//		          released, however briefly). Sleep and wake are one continuous
//		          held session in this mode, not two.
//
//		Caller contract for --wait-signal: create PATH's parent dir if needed,
//		ensure PATH does NOT already exist before invoking (a stale leftover
//		file from a prior run would make this tool see "signal already given"
//		instantly and skip the wait entirely) - a fresh/unique path per
//		invocation (e.g. mktemp) is the caller's responsibility, not this
//		tool's. This tool never creates or deletes PATH itself, only polls for
//		its existence.
//
// Any number of robots, fired concurrently. Loops=1 on every trigger fire -
// GoToSleepSleeping's own loop behavior lives in the robot's own behavior
// tree once triggered, not in a re-fire loop from this tool.
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

// Bumped 2.0 -> 2.7 2026-07-25 (operator request, "golden sequence" spec) -
// mandatory settle window between GoToSleepGetIn and GoToSleepSleeping.
const getInSettleSeconds = 2.7

const defaultMaxWaitSeconds = 60.0

// wakeClip (--wake-clip NAME, sleep only; 2026-10-14, operator: "Try slower
// anim_gotosleep_wakeup_01 animation"): the sleeping robot is woken by this raw
// clip, in the same held session, instead of the GoToSleepOff trigger - with
// --hold, at the end of the hold (before release); with --wait-signal, when
// the signal comes.
var wakeClip string

// fireWake wakes a sleeping robot: wakeClip when given, else GoToSleepOff.
func fireWake(ctx context.Context, robot *vector.Vector, res *robotResult) bool {
	// 2026-10-15 (bench: Pip's slow sleep fires used up the shared deadline and
	// the wake clip timed out, twice): the wake gets its own 20 s.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if wakeClip != "" {
		if err := fireClip(ctx, robot, wakeClip); err != nil {
			res.Err = fmt.Sprintf("%s: %v", wakeClip, err)
			return false
		}
		res.Fired = append(res.Fired, wakeClip)
		return true
	}
	if err := fireTrigger(ctx, robot, "GoToSleepOff"); err != nil {
		res.Err = fmt.Sprintf("GoToSleepOff: %v", err)
		return false
	}
	res.Fired = append(res.Fired, "GoToSleepOff")
	return true
}

type robotResult struct {
	Label               string
	Serial              string
	Action              string
	Start               time.Time
	End                 time.Time
	OK                  bool
	Err                 string
	Fired               []string
	GetInFireSeconds    float64
	SleepingFireSeconds float64
	PlayFireSeconds     float64
	WaitedForSignal     bool
}

func fireTrigger(ctx context.Context, robot *vector.Vector, name string) error {
	_, err := robot.Conn.PlayAnimationTrigger(ctx, &vectorpb.PlayAnimationTriggerRequest{
		AnimationTrigger: &vectorpb.AnimationTrigger{Name: name},
		Loops:            1,
	})
	return err
}

// fireClip plays a raw one-shot clip by exact name (vectorpb.PlayAnimationRequest),
// the same call shape Wire-Pod's own DoPlayAnimation uses - added for the sleep-lane
// test bench (SLEEP_BENCH_ALIASES_001.md) so a candidate clip name can be fired the
// same isolated way a trigger can, via the new "play" action below.
func fireClip(ctx context.Context, robot *vector.Vector, name string) error {
	_, err := robot.Conn.PlayAnimation(ctx, &vectorpb.PlayAnimationRequest{
		Animation: &vectorpb.Animation{Name: name},
		Loops:     1,
	})
	return err
}

// watchSignalFile polls path for existence and closes releaseSignal the
// instant it appears, or after maxWaitSeconds elapse with no signal
// (printing a timeout notice first) - either way, releaseSignal always gets
// closed exactly once, so callers blocked on it can never hang forever.
func watchSignalFile(path string, releaseSignal chan struct{}, maxWaitSeconds float64) {
	deadline := time.Now().Add(time.Duration(maxWaitSeconds * float64(time.Second)))
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			close(releaseSignal)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Fprintf(os.Stderr, "ROBOT_SLEEP_WAIT_SIGNAL_TIMEOUT path=%s after %.1fs - releasing anyway\n", path, maxWaitSeconds)
	close(releaseSignal)
}

func assumeControlAndAct(
	serial, label, action string,
	holdSeconds float64,
	readyBarrier *gopoddirect.Barrier,
	releaseSignal <-chan struct{},
	allAsleepBarrier *gopoddirect.Barrier,
	playName string,
	playIsClip bool,
) robotResult {
	res := robotResult{Serial: serial, Label: label, Action: action}

	// Deadline. Baseline 20s + whichever release mechanism is in play -
	// holdSeconds for the legacy fixed-hold path, or maxWaitSeconds (passed
	// in via holdSeconds's own slot when --wait-signal is active - see
	// main()) for the event-driven path, so a long wait isn't itself cut
	// off by this same safety net. Added 2026-07-25 after a live run hung
	// indefinitely (Ctrl-C needed) - context.Background() has no deadline
	// at all, so stream.Recv() below blocks forever if a robot never
	// answers the control-grant request (e.g. unpaired from Wire-Pod). Same
	// gap direct_sdk_brobots_ready_001.go has; its only caller (core.sh)
	// wraps it in an external `timeout 30` for the same reason - this file
	// gets its own internal deadline too (the bash aliases below also add a
	// `timeout` wrapper, belt and suspenders).
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20+holdSeconds)*time.Second)
	defer cancel()

	// 2026-08-08: connect/grant/release now come from the shared gopoddirect
	// package (PHCAL_SLEEP_HOLD_BARRIER_FIX_001.md's own follow-up) instead
	// of this file's own hand-rolled skeleton - see that package's own doc
	// comment for why.
	robot, stream, err := gopoddirect.ConnectAndGrant(ctx, serial)
	if err != nil {
		res.Err = err.Error()
		readyBarrier.Done()
		if allAsleepBarrier != nil {
			allAsleepBarrier.Done()
		}
		return res
	}

	// Start barrier, added 2026-07-25 (operator-observed live gap): a real
	// run showed the two robots' fire calls landing ~365ms apart - each
	// robot's own connect+BehaviorControl-grant handshake simply takes a
	// different real amount of time, and res.Start used to be stamped
	// independently right after THIS robot's own handshake finished,
	// inheriting that skew. Every robot's goroutine calls Done() the instant
	// it reaches this point (on the connect/grant failure path above too, so
	// one robot's failure can never leave the others hung waiting on the
	// barrier), then Wait()s here - gopoddirect.Barrier (a sync.WaitGroup
	// underneath) is safe for multiple concurrent callers and unblocks all
	// of them together the instant the LAST one calls Done(). Only after
	// every robot has reached this same real moment does any of them
	// proceed to fire its first trigger. Confirmed live 2026-07-25: 104
	// microseconds apart (fifth live run).
	readyBarrier.Done()
	readyBarrier.Wait()

	res.Start = time.Now()
	switch action {
	case "sleep":
		getInStart := time.Now()
		if err := fireTrigger(ctx, robot, "GoToSleepGetIn"); err != nil {
			res.Err = fmt.Sprintf("GoToSleepGetIn: %v", err)
			if allAsleepBarrier != nil {
				allAsleepBarrier.Done()
			}
			break
		}
		res.Fired = append(res.Fired, "GoToSleepGetIn")
		res.GetInFireSeconds = time.Since(getInStart).Seconds()
		time.Sleep(time.Duration(getInSettleSeconds * float64(time.Second)))
		sleepingStart := time.Now()
		if err := fireTrigger(ctx, robot, "GoToSleepSleeping"); err != nil {
			res.Err = fmt.Sprintf("GoToSleepSleeping: %v", err)
			if allAsleepBarrier != nil {
				allAsleepBarrier.Done()
			}
			break
		}
		res.Fired = append(res.Fired, "GoToSleepSleeping")
		res.SleepingFireSeconds = time.Since(sleepingStart).Seconds()

		if releaseSignal != nil {
			// All-asleep barrier, added 2026-07-25 (operator-observed live
			// gap): the start barrier above syncs when robots BEGIN firing,
			// but each robot's own GoToSleepGetIn/GoToSleepSleeping fire
			// calls are real network round-trips whose duration varies
			// wildly robot to robot (a live run showed one robot's
			// GoToSleepSleeping fire taking 10-15s longer than the other's).
			// Without this barrier, a slow robot could still be mid-fire
			// when the external warm-up finishes and the release channel
			// closes - it would then just finish on its own schedule and
			// wake late, unsynced, even though the channel close itself is
			// a genuine simultaneous broadcast. Same sync.WaitGroup
			// Done()/Wait() pattern as the start barrier: every robot calls
			// Done() the instant it finishes firing GoToSleepSleeping (on
			// every earlier failure path too, so one robot's connect/grant/
			// fire failure can never leave the others hung here), then all
			// Wait() together before any of them looks at releaseSignal -
			// so if the channel is already closed, every robot sees it at
			// effectively the same instant, and if it isn't yet, they all
			// start waiting on it from the same real moment.
			if allAsleepBarrier != nil {
				allAsleepBarrier.Done()
				allAsleepBarrier.Wait()
			}
			// Event-driven path ("golden sequence" mode). Block here until
			// watchSignalFile closes the channel (external warm-up done, or
			// the max-wait safety net elapsed) - see this file's own module
			// docstring for why a channel close is used instead of file-
			// polling per robot. The instant this unblocks, fire
			// GoToSleepOff immediately, in this SAME still-open connection
			// - transitioning straight from asleep to awake without ever
			// releasing control in between (see the docstring's own
			// paragraph on why that gap matters).
			<-releaseSignal
			res.WaitedForSignal = true
			if !fireWake(ctx, robot, &res) {
				break
			}
		} else {
			// Barrier-corrected fixed-hold path (2026-08-08, live-observed
			// bug - GOLDEN_BROBOTS_CONTROL_CATALOG_SURVEY_001.md /
			// PHCAL_SLEEP_HOLD_BARRIER_FIX_001.md). Previously slept until
			// the shared releaseAt timestamp computed once in main(),
			// BEFORE either robot's goroutine started firing - a robot
			// whose own GoToSleepGetIn/GoToSleepSleeping round trips ran
			// long (live-observed: 9.625s elapsed on one robot vs 2.988s on
			// the other, same call) could finish firing AFTER that blind
			// deadline had already passed, releasing control with ZERO
			// settle time - which live-confirmed wakes it immediately.
			// Fixed the same way the --wait-signal path already
			// coordinates multi-robot timing: every robot calls Done() the
			// instant IT actually finishes firing GoToSleepSleeping, then
			// all Wait() together, so the hold below starts from a real,
			// synchronized "everyone is now asleep" moment instead of a
			// deadline set blind to per-robot response time. Every robot
			// now gets its full requested hold AFTER it is genuinely
			// asleep, regardless of the other robot's own latency.
			allAsleepBarrier.Done()
			allAsleepBarrier.Wait()
			time.Sleep(time.Duration(holdSeconds * float64(time.Second)))
			if wakeClip != "" && !fireWake(ctx, robot, &res) {
				break
			}
		}
	case "wake":
		if err := fireTrigger(ctx, robot, "GoToSleepOff"); err != nil {
			res.Err = fmt.Sprintf("GoToSleepOff: %v", err)
			break
		}
		res.Fired = append(res.Fired, "GoToSleepOff")
	case "play":
		// Sleep-lane test bench (SLEEP_BENCH_ALIASES_001.md): fire exactly one
		// named trigger or clip, hold BehaviorControl open for holdSeconds,
		// then release - same connect/grant/release/timeout shape as
		// sleep/wake above, no barrier/signal machinery beyond the shared
		// start barrier every action already goes through.
		playStart := time.Now()
		var playErr error
		if playIsClip {
			playErr = fireClip(ctx, robot, playName)
		} else {
			playErr = fireTrigger(ctx, robot, playName)
		}
		if playErr != nil {
			res.Err = fmt.Sprintf("%s: %v", playName, playErr)
			break
		}
		res.Fired = append(res.Fired, playName)
		res.PlayFireSeconds = time.Since(playStart).Seconds()
		time.Sleep(time.Duration(holdSeconds * float64(time.Second)))
	default:
		res.Err = fmt.Sprintf("unknown action %q, must be sleep|wake|play", action)
	}
	res.End = time.Now()
	res.OK = res.Err == ""

	gopoddirect.Release(stream)

	return res
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: direct_sdk_robot_sleep_001 <sleep|wake> [--hold N | --wait-signal PATH [--max-wait N]] [--wake-clip NAME] <label1>:<serial1> [<label2>:<serial2> ...]")
		fmt.Fprintln(os.Stderr, "       direct_sdk_robot_sleep_001 play [--hold N] [--clip] <name> <label1>:<serial1> [<label2>:<serial2> ...]")
		os.Exit(2)
	}
	action := os.Args[1]
	rest := os.Args[2:]

	holdSeconds := 5.0
	holdExplicit := false
	waitSignalPath := ""
	maxWaitSeconds := defaultMaxWaitSeconds
	playIsClip := false

	for len(rest) > 0 {
		switch rest[0] {
		case "--hold":
			if len(rest) < 2 {
				fmt.Fprintln(os.Stderr, "--hold needs a number of seconds")
				os.Exit(2)
			}
			parsed, err := fmt.Sscanf(rest[1], "%f", &holdSeconds)
			if err != nil || parsed != 1 {
				fmt.Fprintf(os.Stderr, "bad --hold value %q\n", rest[1])
				os.Exit(2)
			}
			holdExplicit = true
			rest = rest[2:]
		case "--wait-signal":
			if len(rest) < 2 {
				fmt.Fprintln(os.Stderr, "--wait-signal needs a file path")
				os.Exit(2)
			}
			waitSignalPath = rest[1]
			rest = rest[2:]
		case "--max-wait":
			if len(rest) < 2 {
				fmt.Fprintln(os.Stderr, "--max-wait needs a number of seconds")
				os.Exit(2)
			}
			parsed, err := fmt.Sscanf(rest[1], "%f", &maxWaitSeconds)
			if err != nil || parsed != 1 {
				fmt.Fprintf(os.Stderr, "bad --max-wait value %q\n", rest[1])
				os.Exit(2)
			}
			rest = rest[2:]
		case "--clip":
			playIsClip = true
			rest = rest[1:]
		case "--wake-clip":
			if len(rest) < 2 {
				fmt.Fprintln(os.Stderr, "--wake-clip needs a clip name")
				os.Exit(2)
			}
			wakeClip = rest[1]
			rest = rest[2:]
		default:
			goto specsDone
		}
	}
specsDone:
	if waitSignalPath != "" && holdExplicit {
		fmt.Fprintln(os.Stderr, "--hold and --wait-signal are mutually exclusive")
		os.Exit(2)
	}
	if action != "sleep" && wakeClip != "" {
		fmt.Fprintln(os.Stderr, "--wake-clip only applies to the sleep action")
		os.Exit(2)
	}
	if action != "play" && playIsClip {
		fmt.Fprintln(os.Stderr, "--clip only applies to the play action")
		os.Exit(2)
	}
	playName := ""
	if action == "play" {
		if waitSignalPath != "" {
			fmt.Fprintln(os.Stderr, "play does not support --wait-signal")
			os.Exit(2)
		}
		if len(rest) == 0 {
			fmt.Fprintln(os.Stderr, "play needs a trigger or clip name before the robot specs")
			os.Exit(2)
		}
		playName = rest[0]
		rest = rest[1:]
	}
	specs := rest
	if len(specs) == 0 {
		fmt.Fprintln(os.Stderr, "no robots given, want at least one label:serial")
		os.Exit(2)
	}

	// 2026-08-08 (gopoddirect refactor, PHCAL_SLEEP_HOLD_BARRIER_FIX_001.md's
	// own follow-up): releaseAt (the old blind shared deadline) is gone
	// entirely now, not just unread - this pass touches every call site
	// anyway, so the vestigial parameter from the barrier fix is cleaned up
	// here rather than left in place.
	var releaseSignal chan struct{}
	// All-asleep barrier - see assumeControlAndAct's own comment on it for
	// why this exists. Allocated whenever action=="sleep" (widened
	// 2026-08-08 from --wait-signal-only): the legacy fixed-hold branch uses
	// this same barrier too, not just the event-driven one. Stays nil for
	// "wake"/"play" (neither references it) - same nil-safe shape as
	// releaseSignal. gopoddirect.Barrier, not a private *sync.WaitGroup -
	// same shared piece assumeControlAndAct's own start barrier and
	// direct_sdk_brobots_ready_001.go's grant barrier now use too.
	var allAsleepBarrier *gopoddirect.Barrier
	if waitSignalPath != "" || action == "sleep" {
		allAsleepBarrier = gopoddirect.NewBarrier(len(specs))
	}
	if waitSignalPath != "" {
		releaseSignal = make(chan struct{})
		go watchSignalFile(waitSignalPath, releaseSignal, maxWaitSeconds)
		// Context deadline above uses holdSeconds as its "+N" term; widen it
		// to cover the real max wait instead so a legitimate long warm-up
		// isn't killed by the connection-level timeout.
		holdSeconds = maxWaitSeconds
	}

	results := make([]robotResult, len(specs))
	var wg sync.WaitGroup
	// Start barrier - see assumeControlAndAct's own comment on it for why
	// this exists (syncs the FIRE moment across robots, not just the
	// release moment allAsleepBarrier/releaseSignal handle).
	readyBarrier := gopoddirect.NewBarrier(len(specs))
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
			results[i] = assumeControlAndAct(serial, label, action, holdSeconds, readyBarrier, releaseSignal, allAsleepBarrier, playName, playIsClip)
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
			"ROBOT_SLEEP_DIRECT label=%s serial=%s action=%s fired=%s getin_fire=%.3fs sleeping_fire=%.3fs play_fire=%.3fs waited_for_signal=%t start=%.6f end=%.6f elapsed=%.3fs status=%s\n",
			r.Label, r.Serial, r.Action, strings.Join(r.Fired, ","),
			r.GetInFireSeconds, r.SleepingFireSeconds, r.PlayFireSeconds, r.WaitedForSignal,
			float64(r.Start.UnixNano())/1e9, float64(r.End.UnixNano())/1e9,
			r.End.Sub(r.Start).Seconds(), status,
		)
	}
	if !allOK {
		os.Exit(1)
	}
}
