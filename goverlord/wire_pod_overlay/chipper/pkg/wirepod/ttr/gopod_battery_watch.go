// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_battery_watch.go
package wirepod_ttr

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// gopodBatteryLow: whether each Brobot has had his say for this drop, and his last reading (the one
// before this one, for a dip).
var gopodBatteryLow = struct {
	sync.Mutex
	said   map[string]bool
	prev   map[string]float64
	seen   map[string]bool // read at least once this Wire-Pod session
	misses map[string]int  // readings in a row that didn't work
	full   map[string]bool // on his charger, the robot's own level FULL (gopodBatteryFull)
}{said: map[string]bool{}, prev: map[string]float64{}, seen: map[string]bool{}, misses: map[string]int{}, full: map[string]bool{}}

// gopodBatteryFull: on his charger at FULL he's full, whatever the volts.
func gopodBatteryFull(level string, onCharger bool) bool {
	return onCharger && level == "FULL"
}

// gopodBatteryBack: a reading -> whether he's just back (a freeze, a reboot, a drained battery) -
// read before this session, then back_after_misses readings in a row that didn't work, and now one
// that does.
func gopodBatteryBack(esn string, ok bool) bool {
	n := gopodBrobotsSettings.BatteryWatch.BackAfterMisses
	gopodBatteryLow.Lock()
	defer gopodBatteryLow.Unlock()
	if !ok {
		if gopodBatteryLow.seen[esn] {
			gopodBatteryLow.misses[esn]++
		}
		return false
	}
	back := n > 0 && gopodBatteryLow.seen[esn] && gopodBatteryLow.misses[esn] >= n
	gopodBatteryLow.seen[esn], gopodBatteryLow.misses[esn] = true, 0
	return back
}

// gopodBatteryDecide: one reading -> whether to say the line now, and the new said state.
func gopodBatteryDecide(said bool, prev, volts float64, charging, onCharger, full bool) (sayNow, saidAfter bool) {
	bw := gopodBrobotsSettings.BatteryWatch
	if charging || full {
		return false, false
	}
	dip := bw.DipVolts > 0 && prev > 0 && volts > 0 && prev-volts >= bw.DipVolts
	if dip && !said {
		return true, true
	}
	if volts >= bw.RecoverVolts {
		return false, false
	}
	if volts > 0 && volts < bw.LowVolts && !said {
		return true, true
	}
	return false, said
}

func init() {
	go gopodBatteryWatch()
}

func gopodBatteryWatch() {
	for {
		bw := gopodBrobotsSettings.BatteryWatch
		every := gopodSecondsToDuration(bw.IntervalSeconds)
		if every <= 0 {
			every = time.Minute
		}
		time.Sleep(every)
		if !bw.Enabled {
			continue
		}
		for _, esn := range []string{gopodBrobot1ESN, gopodBrobot2ESN} {
			gopodBatteryCheck(esn)
		}
	}
}

type gopodBatteryLogged struct {
	volts              float64
	level              string
	charging, onCharge bool
}

// gopodBatteryLogLast: each robot's last logged battery line (esn -> gopodBatteryLogged).
var gopodBatteryLogLast sync.Map

// gopodBatteryLogChanged: true when this check is worth a log line (the first check, a new level,
// charging or charger state, or 0.05 V or more of change).
func gopodBatteryLogChanged(esn string, volts float64, level string, charging, onCharger bool) bool {
	now := gopodBatteryLogged{volts, level, charging, onCharger}
	if v, ok := gopodBatteryLogLast.Load(esn); ok {
		last := v.(gopodBatteryLogged)
		diff := volts - last.volts
		if diff < 0 {
			diff = -diff
		}
		if last.level == level && last.charging == charging && last.onCharge == onCharger && diff < 0.05 {
			return false
		}
	}
	gopodBatteryLogLast.Store(esn, now)
	return true
}

func gopodBatteryCheck(esn string) {
	var robot *vector.Vector
	for _, bot := range vars.BotInfo.Robots {
		if bot.Esn == esn {
			robot, _ = vector.New(vector.WithSerialNo(esn), vector.WithToken(bot.GUID), vector.WithTarget(bot.IPAddress+":443"))
			robot = GopodRobotsOffConn(robot) // robots off (gopod_robots_off.go)
			break
		}
	}
	if robot == nil {
		return
	}
	name := gopodBrobotName(esn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	b, err := robot.Conn.BatteryState(ctx, &vectorpb.BatteryStateRequest{})
	cancel()
	if err != nil {
		logger.Println("GOPOD_BATTERY_WATCH " + name + " unknown: " + err.Error())
		gopodBatteryRemember(esn, "unknown: "+err.Error())
		gopodNetworkMiss(esn, err)
		gopodBatteryBack(esn, false)
		return
	}
	gopodNetworkBack(esn)
	if gopodBatteryBack(esn, true) {
		if line := strings.TrimSpace(gopodBrobotsSettings.BatteryWatch.BackLines[name]); line != "" {
			logger.Println("GOPOD_BATTERY_WATCH " + name + " back (freeze, reboot or drained battery): " + line)
			go gopodBatterySafeSay(name, line)
		}
	}
	volts := float64(b.GetBatteryVolts())
	gopodBatteryRemember(esn, fmt.Sprintf("%.2f V, %s%s", volts, strings.TrimPrefix(b.GetBatteryLevel().String(), "BATTERY_LEVEL_"),
		map[bool]string{true: ", on the charger", false: ""}[b.GetIsOnChargerPlatform()]+map[bool]string{true: ", charging", false: ""}[b.GetIsCharging()]))
	level := strings.TrimPrefix(b.GetBatteryLevel().String(), "BATTERY_LEVEL_")
	if gopodBatteryLogChanged(esn, volts, level, b.GetIsCharging(), b.GetIsOnChargerPlatform()) {
		logger.Println(fmt.Sprintf("GOPOD_BATTERY_WATCH %s volts=%.2f level=%s charging=%t on_charger=%t",
			name, volts, level, b.GetIsCharging(), b.GetIsOnChargerPlatform()))
	}
	gopodBatteryLow.Lock()
	prev := gopodBatteryLow.prev[esn]
	full := gopodBatteryFull(level, b.GetIsOnChargerPlatform())
	sayNow, said := gopodBatteryDecide(gopodBatteryLow.said[esn], prev, volts, b.GetIsCharging(), b.GetIsOnChargerPlatform(), full)
	gopodBatteryLow.said[esn] = said
	gopodBatteryLow.full[esn] = full
	if volts > 0 {
		gopodBatteryLow.prev[esn] = volts
	}
	if prev > 0 && volts > 0 && prev-volts >= gopodBrobotsSettings.BatteryWatch.DipVolts && gopodBrobotsSettings.BatteryWatch.DipVolts > 0 {
		logger.Println(fmt.Sprintf("GOPOD_BATTERY_WATCH %s dip: %.2f V -> %.2f V", name, prev, volts))
	}
	gopodBatteryLow.Unlock()
	if sayNow {
		go gopodBatterySay(name, volts, b.GetIsOnChargerPlatform())
	}
}

// gopodBatteryLast: each Brobot's last reading.
var gopodBatteryLast = struct {
	sync.Mutex
	text map[string]string
	at   map[string]time.Time
}{text: map[string]string{}, at: map[string]time.Time{}}

func gopodBatteryRemember(esn, text string) {
	gopodBatteryLast.Lock()
	gopodBatteryLast.text[esn], gopodBatteryLast.at[esn] = text, time.Now()
	gopodBatteryLast.Unlock()
}

// gopodBatterySay: once no reply and no run is going, the Brobot says one of his lines
// (gopodRecordSay: his bubble and his voice).
func gopodBatterySay(name string, volts float64, onCharger bool) {
	lines := gopodBrobotsSettings.BatteryWatch.Lines[name]
	if len(lines) == 0 {
		return
	}
	defer gopodBatterySafeSpot()()
	said := lines[rand.Intn(len(lines))]
	if home := strings.TrimSpace(gopodBrobotsSettings.BatteryWatch.HomeLines[name]); home != "" && !onCharger {
		said = strings.TrimSpace(said) + " " + home
	}
	guard, warning, reply, line := gopodBatteryLines(name, volts, said)
	logger.Println(fmt.Sprintf("GOPOD_BATTERY_WATCH %s low (%.2f V): %s", name, volts, line))
	if guard != "" && gopodReporterSayTo(guard, "", warning, true, true) {
		line = reply
	}
	gopodBodySay(name, line)
}

// gopodBatterySafeSpot: live chat only, between replies - never while a reply, a run or a draft is
// going.
func gopodBatterySafeSpot() func() {
	for gopodReplyBusy() || gopodRunActive() || gopodScriptPlaying.Load() {
		time.Sleep(2 * time.Second)
	}
	return gopodTurnBeginOwn("battery", false)
}

// gopodBatterySafeSay: his line alone, his bubble and voice, at the next safe spot.
func gopodBatterySafeSay(name, line string) {
	defer gopodBatterySafeSpot()()
	gopodBodySay(name, line)
}

// gopodBatteryLines: what a low battery says - his guard (none: ""), the guard's warning and
// handover, his reply after it, and his line alone (with no guard).
func gopodBatteryLines(name string, volts float64, line string) (guard, warning, reply, alone string) {
	bw := gopodBrobotsSettings.BatteryWatch
	guard = bw.Guards[name]
	if guard == "" {
		return "", "", line, line
	}
	warn := strings.NewReplacer("{brobot}", name, "{volts}", fmt.Sprintf("%.2f", volts)).Replace(bw.GuardWarning)
	hand := gopodNicknameFill(guard, strings.ReplaceAll(bw.GuardHandover, "{brobot}", name))
	come := gopodNicknameFill(name, gopodBrobotsSettings.Openers["comeback"])
	reply = strings.Join(strings.Fields(strings.NewReplacer("{comeback}", come, "{line}", line).Replace(bw.Reply)), " ")
	return guard, strings.TrimSpace(warn + " " + hand), reply, line
}

// gopodSecuritySweep: before a run checks in, each Brobot's guard (battery_watch.guards) sweeps him
// - found, and above low_volts.
func gopodSecuritySweep() bool {
	bw := gopodBrobotsSettings.BatteryWatch
	if GopodRobotsOff() || bw.SweepClear == "" {
		return true
	}
	gopodBatteryLast.Lock()
	read := len(gopodBatteryLast.text) > 0
	gopodBatteryLast.Unlock()
	if len(vars.BotInfo.Robots) == 0 && !read {
		return true
	}
	fresh := 2 * gopodSecondsToDuration(bw.IntervalSeconds)
	var wg sync.WaitGroup
	for _, esn := range []string{gopodBrobot1ESN, gopodBrobot2ESN} {
		gopodBatteryLast.Lock()
		old := time.Since(gopodBatteryLast.at[esn]) > fresh
		gopodBatteryLast.Unlock()
		if old {
			wg.Add(1)
			go func(esn string) { defer wg.Done(); gopodBatteryCheck(esn) }(esn)
		}
	}
	wg.Wait()
	ok := true
	var notReady []string
	for n, esn := range []string{gopodBrobot1ESN, gopodBrobot2ESN} {
		name := gopodBrobotName(esn)
		gopodBatteryLast.Lock()
		last := gopodBatteryLast.text[esn]
		gopodBatteryLast.Unlock()
		gopodBatteryLow.Lock()
		volts := gopodBatteryLow.prev[esn]
		full := gopodBatteryLow.full[esn]
		gopodBatteryLow.Unlock()
		line := bw.SweepClear
		switch {
		case last == "" || strings.HasPrefix(last, "unknown"):
			line, ok = bw.SweepMissing, false
			notReady = append(notReady, name)
		case full && bw.SweepFull != "":
			line = bw.SweepFull
		case volts > 0 && volts < bw.LowVolts:
			line, ok = bw.SweepLow, false
			notReady = append(notReady, name)
		}
		line = strings.NewReplacer("{n}", fmt.Sprint(n+1), "{brobot}", name, "{volts}", fmt.Sprintf("%.2f", volts)).Replace(line)
		logger.Println("GOPOD_SECURITY_SWEEP " + name + ": " + line)
		if guard := bw.Guards[name]; guard != "" {
			gopodReporterSay(guard, line, true, false)
		}
	}
	if !ok && bw.SweepStop != "" {
		if role := gopodBrobotsSettings.SystemBubble.VoicePersona; role != "" {
			gopodReporterSay(role, strings.ReplaceAll(bw.SweepStop, "{who}", strings.Join(notReady, " and ")), true, false)
		}
	}
	return ok
}

// ## EOF
